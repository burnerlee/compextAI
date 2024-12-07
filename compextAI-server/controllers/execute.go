package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/burnerlee/compextAI/constants"
	"github.com/burnerlee/compextAI/internal/logger"
	"github.com/burnerlee/compextAI/internal/providers/chat"
	"github.com/burnerlee/compextAI/models"
	"gorm.io/gorm"
)

func ExecuteThread(db *gorm.DB, req *ExecuteThreadRequest) (interface{}, error) {
	threadExecutionParamsTemplate, err := models.GetThreadExecutionParamsTemplateByID(db, req.ThreadExecutionParamTemplateID)
	if err != nil {
		logger.GetLogger().Errorf("Error getting thread execution params template: %s: %v", req.ThreadExecutionParamTemplateID, err)
		return nil, err
	}

	if req.ThreadExecutionSystemPrompt != "" {
		logger.GetLogger().Infof("Setting thread execution system prompt: %s", req.ThreadExecutionSystemPrompt)
		threadExecutionParamsTemplate.SystemPrompt = req.ThreadExecutionSystemPrompt
	}

	chatProvider, err := chat.GetChatCompletionsProvider(threadExecutionParamsTemplate.Model)
	if err != nil {
		logger.GetLogger().Errorf("Error getting chat provider: %s: %v", threadExecutionParamsTemplate.Model, err)
		return nil, err
	}

	var messages []*models.Message
	if req.ThreadID != constants.THREAD_IDENTIFIER_FOR_NULL_THREAD && req.FetchMessagesFromThread {
		// get the thread
		threadMessages, err := models.GetAllMessages(db, req.ThreadID)
		if err != nil {
			logger.GetLogger().Errorf("Error getting thread: %s: %v", req.ThreadID, err)
			return nil, err
		}
		messages = threadMessages
	} else {
		messages = req.Messages
	}

	// validating that the project id matches the thread project id
	if req.ThreadID != constants.THREAD_IDENTIFIER_FOR_NULL_THREAD {
		thread, err := models.GetThread(db, req.ThreadID)
		if err != nil {
			logger.GetLogger().Errorf("Error getting thread: %s: %v", req.ThreadID, err)
			return nil, err
		}
		if thread.ProjectID != req.ProjectID {
			return nil, fmt.Errorf("thread %s does not belong to project %s", req.ThreadID, req.ProjectID)
		}
	}

	threadExecution := &models.ThreadExecution{
		UserID:                          req.UserID,
		ThreadID:                        req.ThreadID,
		ThreadExecutionParamsTemplateID: req.ThreadExecutionParamTemplateID,
		Status:                          models.ThreadExecutionStatus_IN_PROGRESS,
		ProjectID:                       req.ProjectID,
		Metadata:                        req.Metadata,
	}

	threadExecution, err = models.CreateThreadExecution(db, threadExecution)
	if err != nil {
		logger.GetLogger().Errorf("Error creating thread execution: %v", err)
		return nil, err
	}

	// add a execution message to the thread
	// this is used to identify the thread execution in the thread messages
	if req.ThreadID != constants.THREAD_IDENTIFIER_FOR_NULL_THREAD {
		contentMap := map[string]interface{}{
			"content": threadExecution.Identifier,
		}
		contentJsonBlob, err := json.Marshal(contentMap)
		if err != nil {
			logger.GetLogger().Errorf("Error marshalling execution message content: %v", err)
			return nil, err
		}
		if err := models.CreateMessage(db, &models.Message{
			ThreadID:   req.ThreadID,
			Role:       "execution",
			ContentMap: contentJsonBlob,
		}); err != nil {
			logger.GetLogger().Errorf("Error creating execution message: %v", err)
			return nil, err
		}
	}

	go func(p chat.ChatCompletionsProvider, messages []*models.Message, threadExecution models.ThreadExecution, threadExecutionParamsTemplate models.ThreadExecutionParamsTemplate, appendAssistantResponse bool) {
		// get the user
		user, err := models.GetUserByID(db, threadExecution.UserID)
		if err != nil {
			logger.GetLogger().Errorf("Error getting user: %d: %v", threadExecution.UserID, err)
			return
		}

		if threadExecutionParamsTemplate.ResponseFormat == nil {
			threadExecutionParamsTemplate.ResponseFormat = json.RawMessage("{}")
		}

		// execute the thread using the chat provider
		statusCode, threadExecutionResponse, err := chatProvider.ExecuteThread(db, user, messages, &threadExecutionParamsTemplate, threadExecution.Identifier)
		if err != nil {
			logger.GetLogger().Errorf("Error executing thread: %s: %v: %v", req.ThreadID, err, threadExecutionResponse)
			handleThreadExecutionError(db, &threadExecution, fmt.Errorf("error executing thread: %v: %v", err, threadExecutionResponse))
			return
		}

		if statusCode != http.StatusOK {
			logger.GetLogger().Errorf("Error executing thread: %s: status code: %d: %v", req.ThreadID, statusCode, threadExecutionResponse)
			handleThreadExecutionError(db, &threadExecution, fmt.Errorf("status code: %d: %v", statusCode, threadExecutionResponse))
			return
		}

		logger.GetLogger().Infof("Thread execution completed: %s", req.ThreadID)
		handleThreadExecutionSuccess(db, p, &threadExecution, threadExecutionResponse, appendAssistantResponse)
	}(chatProvider, messages, *threadExecution, *threadExecutionParamsTemplate, req.AppendAssistantResponse)

	return threadExecution, nil
}

func handleThreadExecutionError(db *gorm.DB, threadExecution *models.ThreadExecution, execErr error) {
	errJson, jsonErr := json.Marshal(struct {
		Error string `json:"error"`
	}{
		Error: execErr.Error(),
	})
	if jsonErr != nil {
		logger.GetLogger().Errorf("Error marshalling error: %v", jsonErr)
		return
	}

	updatedThreadExecution := models.ThreadExecution{
		Base: models.Base{
			ID:         threadExecution.ID,
			Identifier: threadExecution.Identifier,
		},
		Status: models.ThreadExecutionStatus_FAILED,
		Output: errJson,
	}

	models.UpdateThreadExecution(db, &updatedThreadExecution)
}

func handleThreadExecutionSuccess(db *gorm.DB, p chat.ChatCompletionsProvider, threadExecution *models.ThreadExecution, threadExecutionResponse interface{}, appendAssistantResponse bool) {
	responseJson, err := json.Marshal(threadExecutionResponse)
	if err != nil {
		logger.GetLogger().Errorf("Error marshalling thread execution response: %v", err)
		handleThreadExecutionError(db, threadExecution, fmt.Errorf("error marshalling thread execution response: %v", err))
		return
	}

	message, err := p.ConvertExecutionResponseToMessage(threadExecutionResponse)
	if err != nil {
		logger.GetLogger().Errorf("Error converting thread execution response to message: %v", err)
		handleThreadExecutionError(db, threadExecution, fmt.Errorf("error converting thread execution response to message: %v", err))
		return
	}

	if appendAssistantResponse {
		logger.GetLogger().Infof("Appending assistant response")

		if err := models.CreateMessage(db, &models.Message{
			ThreadID:   threadExecution.ThreadID,
			Role:       message.Role,
			ContentMap: message.ContentMap,
			Metadata:   message.Metadata,
		}); err != nil {
			logger.GetLogger().Errorf("Error creating assistant message: %v", err)
			handleThreadExecutionError(db, threadExecution, fmt.Errorf("error creating assistant message: %v", err))
			return
		} else {
			logger.GetLogger().Infof("assistant message created")
		}
	}

	var contentMap map[string]interface{}
	if err := json.Unmarshal(message.ContentMap, &contentMap); err != nil {
		logger.GetLogger().Errorf("Error unmarshalling content map: %v", err)
		return
	}
	outputContent, ok := contentMap["content"]
	if !ok {
		logger.GetLogger().Errorf("Content map does not contain 'content' key")
		return
	}

	outputContentString, ok := outputContent.(string)
	if !ok {
		logger.GetLogger().Errorf("Content is not a string")
		return
	}

	updatedThreadExecution := models.ThreadExecution{
		Base: models.Base{
			ID:         threadExecution.ID,
			Identifier: threadExecution.Identifier,
		},
		Status:                    models.ThreadExecutionStatus_COMPLETED,
		Output:                    responseJson,
		Content:                   outputContentString,
		Role:                      message.Role,
		ExecutionResponseMetadata: message.Metadata,
	}
	models.UpdateThreadExecution(db, &updatedThreadExecution)
}

func RerunThreadExecution(db *gorm.DB, req *RerunThreadExecutionRequest) (interface{}, error) {
	threadExecution, err := models.GetThreadExecutionByID(db, req.ExecutionID)
	if err != nil {
		logger.GetLogger().Errorf("Error getting thread execution: %s: %v", req.ExecutionID, err)
		return nil, err
	}

	if threadExecution.InputMessages == nil {
		return nil, fmt.Errorf("thread execution input messages are nil")
	}

	var messages []*models.Message
	if err := json.Unmarshal(threadExecution.InputMessages, &messages); err != nil {
		logger.GetLogger().Errorf("Error unmarshalling input messages: %v", err)
		return nil, err
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("thread execution input messages are empty")
	}

	return ExecuteThread(db, &ExecuteThreadRequest{
		UserID:                         threadExecution.UserID,
		ThreadID:                       threadExecution.ThreadID,
		ThreadExecutionParamTemplateID: req.ThreadExecutionParamTemplateID,
		ThreadExecutionSystemPrompt:    req.SystemPrompt,
		AppendAssistantResponse:        req.AppendAssistantResponse,
		Messages:                       messages,
		FetchMessagesFromThread:        false,
		ProjectID:                      threadExecution.ProjectID,
	})
}
