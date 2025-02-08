package chat

import (
	"github.com/burnerlee/compextAI/internal/providers/chat/anthropic"
	"github.com/burnerlee/compextAI/internal/providers/chat/litellm"
	"github.com/burnerlee/compextAI/internal/providers/chat/openai"
)

// add all the provider enums here
const (
	GPT4O     ChatCompletionsProvider_Enum = openai.GPT4O_IDENTIFIER
	GPT4      ChatCompletionsProvider_Enum = openai.GPT4_IDENTIFIER
	CLAUDE35  ChatCompletionsProvider_Enum = anthropic.ANTHROPIC_IDENTIFIER
	O1PREVIEW ChatCompletionsProvider_Enum = openai.O1_PREVIEW_IDENTIFIER
	O1MINI    ChatCompletionsProvider_Enum = openai.O1_MINI_IDENTIFIER
	O1        ChatCompletionsProvider_Enum = openai.O1_IDENTIFIER
	O3MINI    ChatCompletionsProvider_Enum = openai.O3_MINI_IDENTIFIER
	LITELLM   ChatCompletionsProvider_Enum = litellm.LITELLM_IDENTIFIER
)

func init() {
	chatCompletionsProviderRegistry = NewChatCompletionsProviderRegistry()

	// register all the providers

	// openai providers
	chatCompletionsProviderRegistry.register(openai.NewGPT4O())
	chatCompletionsProviderRegistry.register(openai.NewO1Mini())
	chatCompletionsProviderRegistry.register(openai.NewGPT4())
	chatCompletionsProviderRegistry.register(openai.NewO1Preview())
	chatCompletionsProviderRegistry.register(openai.NewO1())
	chatCompletionsProviderRegistry.register(openai.NewO3Mini())

	// anthropic providers
	chatCompletionsProviderRegistry.register(anthropic.NewClaude35())

	// litellm provider
	chatCompletionsProviderRegistry.register(litellm.NewLitellm())
}
