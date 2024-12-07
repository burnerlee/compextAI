package openai

import (
	"github.com/burnerlee/compextAI/internal/logger"
	"github.com/burnerlee/compextAI/models"
	"gorm.io/gorm"
)

const (
	O1_MINI_MODEL          = "o1-mini"
	O1_MINI_OWNER          = "openai"
	O1_MINI_IDENTIFIER     = "o1-mini"
	O1_MINI_EXECUTOR_ROUTE = "/chatcompletion/openai"

	O1_MINI_DEFAULT_TEMPERATURE           = 1
	O1_MINI_DEFAULT_MAX_COMPLETION_TOKENS = 65536
	O1_MINI_DEFAULT_TIMEOUT               = 600
)

type O1Mini struct {
	owner         string
	model         string
	allowedRoles  []string
	executorRoute string
}

func NewO1Mini() *O1Mini {
	return &O1Mini{
		owner:         O1_MINI_OWNER,
		model:         O1_MINI_MODEL,
		allowedRoles:  openaiAllowedRoles,
		executorRoute: O1_MINI_EXECUTOR_ROUTE,
	}
}

func (g *O1Mini) GetProviderOwner() string {
	return g.owner
}

func (g *O1Mini) GetProviderModel() string {
	return g.model
}

func (g *O1Mini) GetProviderIdentifier() string {
	return O1_MINI_IDENTIFIER
}

func (g *O1Mini) ValidateMessage(message *models.Message) error {
	return validateMessage(message)
}

func (g *O1Mini) ConvertMessageToProviderFormat(message *models.Message) (interface{}, error) {
	return convertMessageToProviderFormat(message)
}

func (g *O1Mini) ConvertExecutionResponseToMessage(response interface{}) (*models.Message, error) {
	return convertExecutionResponseToMessage(response)
}

func (g *O1Mini) ExecuteThread(db *gorm.DB, user *models.User, messages []*models.Message, threadExecutionParamsTemplate *models.ThreadExecutionParamsTemplate, threadExecutionIdentifier string) (int, interface{}, error) {
	// o1 models don't support system prompts, so we need to handle it here
	messages, err := handleSystemPromptForO1(messages, threadExecutionParamsTemplate)
	if err != nil {
		logger.GetLogger().Errorf("Error handling system prompt for o1: %v", err)
		return -1, nil, err
	}

	return executeThread(db, user, messages, threadExecutionParamsTemplate, threadExecutionIdentifier, &executeParamConfigs{
		Model:                      g.model,
		ExecutorRoute:              g.executorRoute,
		DefaultTemperature:         O1_MINI_DEFAULT_TEMPERATURE,
		DefaultMaxCompletionTokens: O1_MINI_DEFAULT_MAX_COMPLETION_TOKENS,
		DefaultTimeout:             O1_MINI_DEFAULT_TIMEOUT,
	})
}
