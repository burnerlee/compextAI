package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/burnerlee/compextAI/controllers"
	"github.com/burnerlee/compextAI/models"
	"github.com/burnerlee/compextAI/utils"
	"github.com/burnerlee/compextAI/utils/responses"
	"github.com/gorilla/mux"
)

func (s *Server) ListMessages(w http.ResponseWriter, r *http.Request) {
	threadID := mux.Vars(r)["thread_id"]

	if threadID == "" {
		responses.Error(w, http.StatusBadRequest, "thread_id parameter is required")
		return
	}

	includeExecutionMessagesFromThread := false

	includeExecution := r.URL.Query().Get("include_execution")
	if includeExecution == "true" {
		includeExecutionMessagesFromThread = true
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckThreadAccess(s.DB, threadID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to access this thread")
		return
	}

	var messages []*models.Message
	if includeExecutionMessagesFromThread {
		messages, err = models.GetAllMessagesWithExecution(s.DB, threadID)
	} else {
		messages, err = models.GetAllMessages(s.DB, threadID)
	}
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	messagesResponse := []*messageResponse{}
	for _, message := range messages {
		messageResponse, err := convertMessageModelToResponse(message)
		if err != nil {
			responses.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		messagesResponse = append(messagesResponse, messageResponse)
	}
	responses.JSON(w, http.StatusOK, messagesResponse)
}

func (s *Server) CreateMessage(w http.ResponseWriter, r *http.Request) {
	threadID := mux.Vars(r)["thread_id"]

	if threadID == "" {
		responses.Error(w, http.StatusBadRequest, "thread_id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckThreadAccess(s.DB, threadID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to access this thread")
		return
	}

	var messages CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&messages); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := messages.Validate(); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	messagesController := []*controllers.CreateMessage{}
	for _, message := range messages.Messages {
		messagesController = append(messagesController, &controllers.CreateMessage{
			Content:      message.Content,
			Role:         message.Role,
			ToolCallID:   message.ToolCallID,
			Metadata:     message.Metadata,
			ToolCalls:    message.ToolCalls,
			FunctionCall: message.FunctionCall,
		})
	}
	createdMessages, err := controllers.CreateMessages(s.DB, &controllers.CreateMessageRequest{
		ThreadID: threadID,
		Messages: messagesController,
	})
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, createdMessages)
}

func (s *Server) GetMessage(w http.ResponseWriter, r *http.Request) {
	messageID := mux.Vars(r)["id"]

	if messageID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckMessageAccess(s.DB, messageID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to access this message")
		return
	}

	message, err := models.GetMessage(s.DB, messageID)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	messageResponse, err := convertMessageModelToResponse(message)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, messageResponse)
}

func (s *Server) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	messageID := mux.Vars(r)["id"]

	if messageID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	var message UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := message.Validate(); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckMessageAccess(s.DB, messageID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to update this message")
		return
	}

	metadataJsonBlob, err := json.Marshal(message.Metadata)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	contentMap := map[string]interface{}{
		"content": message.Content,
	}
	contentJsonBlob, err := json.Marshal(contentMap)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	toolCallsJsonBlob, err := json.Marshal(message.ToolCalls)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	functionCallJsonBlob, err := json.Marshal(message.FunctionCall)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	updatedMessage, err := models.UpdateMessage(s.DB, &models.Message{
		Base: models.Base{
			Identifier: messageID,
		},
		ContentMap:   contentJsonBlob,
		Role:         message.Role,
		ToolCallID:   message.ToolCallID,
		Metadata:     metadataJsonBlob,
		ToolCalls:    toolCallsJsonBlob,
		FunctionCall: functionCallJsonBlob,
	})
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	updatedMessageResponse, err := convertMessageModelToResponse(updatedMessage)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, updatedMessageResponse)
}

func (s *Server) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	messageID := mux.Vars(r)["id"]

	if messageID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckMessageAccess(s.DB, messageID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to delete this message")
		return
	}

	if err := models.DeleteMessage(s.DB, messageID); err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusNoContent, "message deleted successfully")
}

func convertMessageModelToResponse(message *models.Message) (*messageResponse, error) {
	content := map[string]interface{}{}
	if err := json.Unmarshal(message.ContentMap, &content); err != nil {
		return nil, err
	}

	contentMsg, ok := content["content"]
	if !ok {
		return nil, errors.New("content is required")
	}
	messagesResponse := &messageResponse{
		Content:      contentMsg,
		ThreadID:     message.ThreadID,
		Identifier:   message.Identifier,
		Role:         message.Role,
		ToolCallID:   message.ToolCallID,
		ToolCalls:    message.ToolCalls,
		FunctionCall: message.FunctionCall,
		Metadata:     message.Metadata,
		CreatedAt:    message.CreatedAt,
		UpdatedAt:    message.UpdatedAt,
	}
	return messagesResponse, nil
}
