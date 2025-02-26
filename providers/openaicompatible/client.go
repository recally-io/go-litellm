package openaicompatible

import (
	"fmt"

	"github.com/recally-io/polyllm/providers"
	"github.com/recally-io/polyllm/providers/openai"
)

func New(baseUrl, apiKey string, opts ...providers.Option) (*openai.Client, error) {
	provider := &providers.Provider{
		Type:    providers.ProviderTypeOpenAICompatible,
		BaseURL: baseUrl,
		APIKey:  apiKey,
	}
	for _, opt := range opts {
		opt(provider)
	}

	if provider.APIKey == "" || provider.BaseURL == "" {
		return nil, fmt.Errorf("API key, base URL, and name are required")
	}

	return &openai.Client{Provider: provider}, nil
}
