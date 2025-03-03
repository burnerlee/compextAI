package openai

import (
	"github.com/burnerlee/compextAI/models"
	"gorm.io/gorm"
)

const (
	GPT4O_MODEL          = "gpt-4o"
	GPT4O_OWNER          = "openai"
	GPT4O_IDENTIFIER     = "gpt-4o"
	GPT4O_EXECUTOR_ROUTE = "/chatcompletion/openai"

	GPT4O_DEFAULT_TEMPERATURE           = 0.5
	GPT4O_DEFAULT_MAX_COMPLETION_TOKENS = 10000
	GPT4O_DEFAULT_TIMEOUT               = 600
)

type GPT4O struct {
	owner         string
	model         string
	allowedRoles  []string
	executorRoute string
}

func NewGPT4O() *GPT4O {
	return &GPT4O{
		owner:         GPT4O_OWNER,
		model:         GPT4O_MODEL,
		allowedRoles:  openaiAllowedRoles,
		executorRoute: GPT4O_EXECUTOR_ROUTE,
	}
}

func (g *GPT4O) GetProviderOwner() string {
	return g.owner
}

func (g *GPT4O) GetProviderModel() string {
	return g.model
}

func (g *GPT4O) GetProviderIdentifier() string {
	return GPT4O_IDENTIFIER
}

func (g *GPT4O) ValidateMessage(message *models.Message) error {
	return validateMessage(message)
}

func (g *GPT4O) ConvertMessageToProviderFormat(message *models.Message) (interface{}, error) {
	return convertMessageToProviderFormat(message)
}

func (g *GPT4O) ConvertExecutionResponseToMessage(response interface{}) (*models.Message, error) {
	return convertExecutionResponseToMessage(response)
}

func (g *GPT4O) ExecuteThread(db *gorm.DB, user *models.User, messages []*models.Message, threadExecutionParamsTemplate *models.ThreadExecutionParamsTemplate, threadExecutionIdentifier string, tools []*models.ExecutionTool) (int, interface{}, error) {
	return BaseExecuteThread(db, user, messages, threadExecutionParamsTemplate, threadExecutionIdentifier, &ExecuteParamConfigs{
		Model:                      g.model,
		ExecutorRoute:              g.executorRoute,
		DefaultTemperature:         GPT4O_DEFAULT_TEMPERATURE,
		DefaultMaxCompletionTokens: GPT4O_DEFAULT_MAX_COMPLETION_TOKENS,
		DefaultTimeout:             GPT4O_DEFAULT_TIMEOUT,
	}, tools, map[string]interface{}{
		g.owner: user.OpenAIKey,
	})
}
