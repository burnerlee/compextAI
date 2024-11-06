package chat

import (
	"github.com/burnerlee/compextAI/internal/providers/chat/anthropic"
	"github.com/burnerlee/compextAI/internal/providers/chat/openai"
)

// add all the provider enums here
const (
	GPT4O    ChatCompletionsProvider_Enum = openai.GPT4O_IDENTIFIER
	CLAUDE35 ChatCompletionsProvider_Enum = anthropic.ANTHROPIC_IDENTIFIER
)

func init() {
	chatCompletionsProviderRegistry = NewChatCompletionsProviderRegistry()

	// register all the providers
	chatCompletionsProviderRegistry.register(openai.NewGPT4O())
	chatCompletionsProviderRegistry.register(anthropic.NewClaude35())
}
