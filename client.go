package polyllm

import (
	"fmt"

	"github.com/recally-io/polyllm/llms"
	"github.com/recally-io/polyllm/providers"
	"github.com/recally-io/polyllm/providers/openai"
	"github.com/recally-io/polyllm/providers/openaicompatible"
)

// Factory function to create appropriate client for a provider
func NewClient(provider *providers.Provider) (llms.LLM, error) {
	opts := provider.ToOptions()
	switch provider.Type {
	case providers.ProviderTypeOpenAI:
		return openai.New(provider.APIKey, opts...)
	case providers.ProviderTypeOpenAICompatible:
		return openaicompatible.New(provider.BaseURL, provider.APIKey, opts...)
	case providers.ProviderTypeDeepSeek, providers.ProviderTypeGemini, providers.ProviderTypeQwen, providers.ProviderTypeOpenRouter, providers.ProviderTypeVolcengine, providers.ProviderTypeGroq, providers.ProviderTypeXai, providers.ProviderTypeSiliconflow:
		return openaicompatible.New(provider.BaseURL, provider.APIKey, opts...)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider.Type)
	}
}
