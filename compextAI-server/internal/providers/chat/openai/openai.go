package openai

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/burnerlee/compextAI/internal/logger"
	"github.com/burnerlee/compextAI/internal/providers/chat/base"
	"github.com/burnerlee/compextAI/models"
	"gorm.io/gorm"
)

var (
	openaiAllowedRoles = []string{"user", "assistant", "system"}
)

func validateMessage(message *models.Message) error {
	if message.ContentMap == nil {
		return fmt.Errorf("message content map is nil")
	}

	if !slices.Contains(openaiAllowedRoles, message.Role) {
		return fmt.Errorf("message role is invalid, only %v are allowed", openaiAllowedRoles)
	}
	return nil
}

type openaiMessage struct {
	Role     string                 `json:"role"`
	Content  interface{}            `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

func convertMessageToProviderFormat(message *models.Message) (interface{}, error) {
	var metadata map[string]interface{}
	if message.Metadata != nil {
		if err := json.Unmarshal(message.Metadata, &metadata); err != nil {
			return nil, err
		}
	}

	var contentMap map[string]interface{}
	if err := json.Unmarshal(message.ContentMap, &contentMap); err != nil {
		return nil, err
	}
	content, ok := contentMap["content"]
	if !ok {
		return nil, fmt.Errorf("content map does not contain 'content' key")
	}

	return openaiMessage{
		Role:     message.Role,
		Content:  content,
		Metadata: metadata,
	}, nil
}

func convertExecutionResponseToMessage(response interface{}) (*models.Message, error) {
	responseMap, ok := response.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("response is not a map")
	}

	messageChoices := responseMap["choices"].([]interface{})
	if len(messageChoices) == 0 {
		return nil, fmt.Errorf("no message choices found")
	}
	messageChoice := messageChoices[0].(map[string]interface{})
	message := messageChoice["message"].(map[string]interface{})

	role, ok := message["role"].(string)
	if !ok {
		return nil, fmt.Errorf("message role is not a string")
	}
	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("message content is not a string")
	}

	openAIChatCompletionID := responseMap["id"].(string)
	usage := responseMap["usage"].(map[string]interface{})

	metadata := map[string]interface{}{
		"openai_chat_completion_id": openAIChatCompletionID,
		"usage":                     usage,
	}

	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	contentMap := map[string]interface{}{
		"content": content,
	}
	contentMapJson, err := json.Marshal(contentMap)
	if err != nil {
		return nil, err
	}

	return &models.Message{
		Role:       role,
		ContentMap: contentMapJson,
		Metadata:   metadataJson,
	}, nil
}

type openaiExecutionData struct {
	APIKey              string          `json:"api_key"`
	Model               string          `json:"model"`
	Messages            []openaiMessage `json:"messages"`
	Temperature         float64         `json:"temperature"`
	MaxCompletionTokens int             `json:"max_completion_tokens"`
	Timeout             int             `json:"timeout"`
	ResponseFormat      interface{}     `json:"response_format"`
}

func (d *openaiExecutionData) Validate() error {
	return nil
}

type executeParamConfigs struct {
	Model                      string
	ExecutorRoute              string
	DefaultTemperature         float64
	DefaultMaxCompletionTokens int
	DefaultTimeout             int
}

func getSystemPrompt(messages []*models.Message, threadExecutionParamsTemplate *models.ThreadExecutionParamsTemplate) (string, error) {
	if threadExecutionParamsTemplate.SystemPrompt != "" {
		return threadExecutionParamsTemplate.SystemPrompt, nil
	}

	systemPrompt := ""

	// find the last system message and use it as the system prompt
	for _, message := range messages {
		if message.Role == "system" {
			var contentMap map[string]interface{}
			if err := json.Unmarshal(message.ContentMap, &contentMap); err != nil {
				return "", err
			}
			content, ok := contentMap["content"]
			if !ok {
				logger.GetLogger().Errorf("System message content map does not contain 'content' key")
				return "", fmt.Errorf("system message content map does not contain 'content' key")
			}
			contentStr, ok := content.(string)
			if !ok {
				logger.GetLogger().Errorf("System message content is not a string")
				return "", fmt.Errorf("system message content is not a string")
			}
			systemPrompt = contentStr
		}
	}
	return systemPrompt, nil
}

func filterNonSystemMessages(messages []*models.Message) []*models.Message {
	nonSystemMessages := make([]*models.Message, 0)
	for _, message := range messages {
		if message.Role != "system" {
			nonSystemMessages = append(nonSystemMessages, message)
		}
	}
	return nonSystemMessages
}

func executeThread(db *gorm.DB, user *models.User, messages []*models.Message, threadExecutionParamsTemplate *models.ThreadExecutionParamsTemplate, threadExecutionIdentifier string, configs *executeParamConfigs) (int, interface{}, error) {
	systemPrompt := ""

	modelMessages := make([]openaiMessage, 0)
	for _, message := range messages {
		modelMessage, err := convertMessageToProviderFormat(message)
		if err != nil {
			logger.GetLogger().Errorf("Error converting message to provider format: %v", err)
			return -1, nil, err
		}
		if message.Role == "system" {
			var contentMap map[string]interface{}
			if err := json.Unmarshal(message.ContentMap, &contentMap); err != nil {
				return -1, nil, err
			}
			content, ok := contentMap["content"]
			if !ok {
				logger.GetLogger().Errorf("System message content map does not contain 'content' key")
				return -1, nil, fmt.Errorf("system message content map does not contain 'content' key")
			}
			systemPromptStr, ok := content.(string)
			if !ok {
				logger.GetLogger().Errorf("System message content is not a string")
				return -1, nil, fmt.Errorf("system message content is not a string")
			}
			systemPrompt = systemPromptStr
			continue
		}
		modelMessages = append(modelMessages, modelMessage.(openaiMessage))
	}

	// override the system prompt if it is provided for execution
	if threadExecutionParamsTemplate.SystemPrompt != "" {
		systemPrompt = threadExecutionParamsTemplate.SystemPrompt
	}

	// add the system prompt to the beginning of the messages thread if it is provided
	if systemPrompt != "" {
		modelMessages = append([]openaiMessage{{
			Role:    "system",
			Content: systemPrompt,
		}}, modelMessages...)
	}

	if threadExecutionParamsTemplate.Temperature <= 0 {
		threadExecutionParamsTemplate.Temperature = configs.DefaultTemperature
	}
	if threadExecutionParamsTemplate.MaxCompletionTokens <= 0 {
		threadExecutionParamsTemplate.MaxCompletionTokens = configs.DefaultMaxCompletionTokens
	}
	if threadExecutionParamsTemplate.Timeout <= 0 {
		threadExecutionParamsTemplate.Timeout = configs.DefaultTimeout
	}

	executionData := openaiExecutionData{
		APIKey:              user.OpenAIKey,
		Model:               configs.Model,
		Messages:            modelMessages,
		Temperature:         threadExecutionParamsTemplate.Temperature,
		MaxCompletionTokens: threadExecutionParamsTemplate.MaxCompletionTokens,
		Timeout:             threadExecutionParamsTemplate.Timeout,
		ResponseFormat:      threadExecutionParamsTemplate.ResponseFormat,
	}

	if err := executionData.Validate(); err != nil {
		logger.GetLogger().Errorf("Error validating execution data: %v", err)
		return -1, nil, err
	}

	executionParams := &base.ExecuteParams{
		Timeout: time.Duration(executionData.Timeout) * time.Second,
	}

	return base.Execute(db, configs.ExecutorRoute, executionParams, executionData, threadExecutionIdentifier, modelMessages)
}
