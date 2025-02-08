package openai

import (
	"github.com/burnerlee/compextAI/internal/logger"
	"github.com/burnerlee/compextAI/models"
	"gorm.io/gorm"
)

const (
	O3_MINI_MODEL          = "o3-mini"
	O3_MINI_OWNER          = "openai"
	O3_MINI_IDENTIFIER     = "o3-mini"
	O3_MINI_EXECUTOR_ROUTE = "/chatcompletion/openai"

	O3_MINI_DEFAULT_TEMPERATURE           = 1
	O3_MINI_DEFAULT_MAX_COMPLETION_TOKENS = 65536
	O3_MINI_DEFAULT_TIMEOUT               = 600
)

type O3Mini struct {
	owner         string
	model         string
	allowedRoles  []string
	executorRoute string
}

func NewO3Mini() *O3Mini {
	return &O3Mini{
		owner:         O3_MINI_OWNER,
		model:         O3_MINI_MODEL,
		allowedRoles:  openaiAllowedRoles,
		executorRoute: O3_MINI_EXECUTOR_ROUTE,
	}
}

func (g *O3Mini) GetProviderOwner() string {
	return g.owner
}

func (g *O3Mini) GetProviderModel() string {
	return g.model
}

func (g *O3Mini) GetProviderIdentifier() string {
	return O3_MINI_IDENTIFIER
}

func (g *O3Mini) ValidateMessage(message *models.Message) error {
	return validateMessage(message)
}

func (g *O3Mini) ConvertMessageToProviderFormat(message *models.Message) (interface{}, error) {
	return convertMessageToProviderFormat(message)
}

func (g *O3Mini) ConvertExecutionResponseToMessage(response interface{}) (*models.Message, error) {
	return convertExecutionResponseToMessage(response)
}

func (g *O3Mini) ExecuteThread(db *gorm.DB, user *models.User, messages []*models.Message, threadExecutionParamsTemplate *models.ThreadExecutionParamsTemplate, threadExecutionIdentifier string, tools []*models.ExecutionTool) (int, interface{}, error) {
	// o1 models don't support system prompts, so we need to handle it here
	messages, err := handleSystemPromptForO1(messages, threadExecutionParamsTemplate)
	if err != nil {
		logger.GetLogger().Errorf("Error handling system prompt for o1: %v", err)
		return -1, nil, err
	}

	return BaseExecuteThread(db, user, messages, threadExecutionParamsTemplate, threadExecutionIdentifier, &ExecuteParamConfigs{
		Model:                      g.model,
		ExecutorRoute:              g.executorRoute,
		DefaultTemperature:         O3_MINI_DEFAULT_TEMPERATURE,
		DefaultMaxCompletionTokens: O3_MINI_DEFAULT_MAX_COMPLETION_TOKENS,
		DefaultTimeout:             O3_MINI_DEFAULT_TIMEOUT,
	}, tools, map[string]interface{}{
		g.owner: user.OpenAIKey,
	})
}
