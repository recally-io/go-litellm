package polyllm

import (
	"context"
	"fmt"

	"github.com/recally-io/polyllm/llms"
	"github.com/recally-io/polyllm/llms/openai"
	"github.com/recally-io/polyllm/llms/openaicompatible"
)

type LLM interface {
	GetProvider() *llms.Provider
	ListModels(ctx context.Context) ([]llms.Model, error)

	ChatCompletion(ctx context.Context, req llms.ChatCompletionRequest, streamingFunc func(resp llms.StreamingChatCompletionResponse), options ...llms.RequestOption)
	// GenerateText(ctx context.Context, model, prompt string, options ...llms.RequestOption) (string, error)
	// StreamGenerateText(ctx context.Context, model, prompt string, streamingFunc func(resp llms.StreamingChatCompletionText), options ...llms.RequestOption)
}

// Factory function to create appropriate client for a provider
func NewLLM(provider *llms.Provider) (LLM, error) {
	opts := provider.ToOptions()
	switch provider.Type {
	case llms.ProviderTypeOpenAI:
		return openai.New(provider.APIKey, opts...)
	case llms.ProviderTypeOpenAICompatible:
		return openaicompatible.New(provider.BaseURL, provider.APIKey, opts...)
	case llms.ProviderTypeDeepSeek, llms.ProviderTypeGemini, llms.ProviderTypeQwen, llms.ProviderTypeOpenRouter, llms.ProviderTypeVolcengine, llms.ProviderTypeGroq, llms.ProviderTypeXai, llms.ProviderTypeSiliconflow, llms.ProviderTypeFireworks, llms.ProviderTypeTogether:
		return openaicompatible.New(provider.BaseURL, provider.APIKey, opts...)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider.Type)
	}
}
