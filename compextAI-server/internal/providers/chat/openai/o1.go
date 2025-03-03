package openai

import (
	"github.com/burnerlee/compextAI/internal/logger"
	"github.com/burnerlee/compextAI/models"
	"gorm.io/gorm"
)

const (
	O1_MODEL          = "o1"
	O1_OWNER          = "openai"
	O1_IDENTIFIER     = "o1"
	O1_EXECUTOR_ROUTE = "/chatcompletion/openai"

	O1_DEFAULT_TEMPERATURE           = 1
	O1_DEFAULT_MAX_COMPLETION_TOKENS = 32768
	O1_DEFAULT_TIMEOUT               = 600
)

type O1 struct {
	owner         string
	model         string
	allowedRoles  []string
	executorRoute string
}

func NewO1() *O1 {
	return &O1{
		owner:         O1_OWNER,
		model:         O1_MODEL,
		allowedRoles:  openaiAllowedRoles,
		executorRoute: O1_EXECUTOR_ROUTE,
	}
}

func (g *O1) GetProviderOwner() string {
	return g.owner
}

func (g *O1) GetProviderModel() string {
	return g.model
}

func (g *O1) GetProviderIdentifier() string {
	return O1_IDENTIFIER
}

func (g *O1) ValidateMessage(message *models.Message) error {
	return validateMessage(message)
}

func (g *O1) ConvertMessageToProviderFormat(message *models.Message) (interface{}, error) {
	return convertMessageToProviderFormat(message)
}

func (g *O1) ConvertExecutionResponseToMessage(response interface{}) (*models.Message, error) {
	return convertExecutionResponseToMessage(response)
}

func (g *O1) ExecuteThread(db *gorm.DB, user *models.User, messages []*models.Message, threadExecutionParamsTemplate *models.ThreadExecutionParamsTemplate, threadExecutionIdentifier string, tools []*models.ExecutionTool) (int, interface{}, error) {
	messages, err := handleSystemPromptForO1(messages, threadExecutionParamsTemplate)
	if err != nil {
		logger.GetLogger().Errorf("Error handling system prompt for o1: %v", err)
		return -1, nil, err
	}

	return BaseExecuteThread(db, user, messages, threadExecutionParamsTemplate, threadExecutionIdentifier, &ExecuteParamConfigs{
		Model:                      g.model,
		ExecutorRoute:              g.executorRoute,
		DefaultTemperature:         O1_PREVIEW_DEFAULT_TEMPERATURE,
		DefaultMaxCompletionTokens: O1_PREVIEW_DEFAULT_MAX_COMPLETION_TOKENS,
		DefaultTimeout:             O1_PREVIEW_DEFAULT_TIMEOUT,
	}, tools, map[string]interface{}{
		g.owner: user.OpenAIKey,
	})
}
