package openai

import (
	"github.com/burnerlee/compextAI/models"
	"gorm.io/gorm"
)

const (
	GPT4_MODEL          = "gpt-4"
	GPT4_OWNER          = "openai"
	GPT4_IDENTIFIER     = "gpt4"
	GPT4_EXECUTOR_ROUTE = "/chatcompletion/openai"

	GPT4_DEFAULT_TEMPERATURE           = 0.5
	GPT4_DEFAULT_MAX_COMPLETION_TOKENS = 8192
	GPT4_DEFAULT_TIMEOUT               = 600
)

type GPT4 struct {
	owner         string
	model         string
	allowedRoles  []string
	executorRoute string
}

func NewGPT4() *GPT4 {
	return &GPT4{
		owner:         GPT4_OWNER,
		model:         GPT4_MODEL,
		allowedRoles:  openaiAllowedRoles,
		executorRoute: GPT4_EXECUTOR_ROUTE,
	}
}

func (g *GPT4) GetProviderOwner() string {
	return g.owner
}

func (g *GPT4) GetProviderModel() string {
	return g.model
}

func (g *GPT4) GetProviderIdentifier() string {
	return GPT4_IDENTIFIER
}

func (g *GPT4) ValidateMessage(message *models.Message) error {
	return validateMessage(message)
}

func (g *GPT4) ConvertMessageToProviderFormat(message *models.Message) (interface{}, error) {
	return convertMessageToProviderFormat(message)
}

func (g *GPT4) ConvertExecutionResponseToMessage(response interface{}) (*models.Message, error) {
	return convertExecutionResponseToMessage(response)
}

func (g *GPT4) ExecuteThread(db *gorm.DB, user *models.User, messages []*models.Message, threadExecutionParamsTemplate *models.ThreadExecutionParamsTemplate, threadExecutionIdentifier string, tools []*models.ExecutionTool) (int, interface{}, error) {
	return BaseExecuteThread(db, user, messages, threadExecutionParamsTemplate, threadExecutionIdentifier, &ExecuteParamConfigs{
		Model:                      g.model,
		ExecutorRoute:              g.executorRoute,
		DefaultTemperature:         GPT4_DEFAULT_TEMPERATURE,
		DefaultMaxCompletionTokens: GPT4_DEFAULT_MAX_COMPLETION_TOKENS,
		DefaultTimeout:             GPT4_DEFAULT_TIMEOUT,
	}, tools, map[string]interface{}{
		g.owner: user.OpenAIKey,
	})
}
