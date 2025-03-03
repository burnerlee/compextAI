package litellm

import (
	"github.com/burnerlee/compextAI/internal/providers/chat/openai"
	"github.com/burnerlee/compextAI/models"
	"gorm.io/gorm"
)

const (
	LITELLM_IDENTIFIER     = "litellm"
	LITTELM_EXECUTOR_ROUTE = "/chatcompletion/litellm"
)

type Litellm struct {
	allowedRoles   []string
	executorRoute  string
	openaiProvider *openai.GPT4O
	owner          string
	model          string
}

func NewLitellm() *Litellm {
	return &Litellm{
		allowedRoles:  []string{"user", "assistant", "system", "tool"},
		executorRoute: LITTELM_EXECUTOR_ROUTE,
		// adding openai provider since litellm follows openai specs
		openaiProvider: openai.NewGPT4O(),
		owner:          "litellm",
	}
}

func (l *Litellm) GetProviderOwner() string {
	return l.owner
}

func (l *Litellm) GetProviderModel() string {
	return l.model
}

func (l *Litellm) GetProviderIdentifier() string {
	return LITELLM_IDENTIFIER
}

func (l *Litellm) ValidateMessage(message *models.Message) error {
	return l.openaiProvider.ValidateMessage(message)
}

func (l *Litellm) ConvertMessageToProviderFormat(message *models.Message) (interface{}, error) {
	return l.openaiProvider.ConvertMessageToProviderFormat(message)
}

func (l *Litellm) ConvertExecutionResponseToMessage(response interface{}) (*models.Message, error) {
	return l.openaiProvider.ConvertExecutionResponseToMessage(response)
}

func (l *Litellm) ExecuteThread(db *gorm.DB, user *models.User, messages []*models.Message, threadExecutionParamsTemplate *models.ThreadExecutionParamsTemplate, threadExecutionIdentifier string, tools []*models.ExecutionTool) (int, interface{}, error) {
	l.model = threadExecutionParamsTemplate.Model
	return openai.BaseExecuteThread(db, user, messages, threadExecutionParamsTemplate, threadExecutionIdentifier, &openai.ExecuteParamConfigs{
		Model:         l.model,
		ExecutorRoute: l.executorRoute,
	}, tools, map[string]interface{}{
		"openai":                       user.OpenAIKey,
		"anthropic":                    user.AnthropicKey,
		"azure":                        user.AzureKey,
		"azure_endpoint":               user.AzureEndpoint,
		"google_service_account_creds": user.GoogleServiceAccountCreds,
	})
}
