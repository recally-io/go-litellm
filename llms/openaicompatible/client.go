package openaicompatible

import (
	"fmt"

	"github.com/recally-io/polyllm/llms"
	"github.com/recally-io/polyllm/llms/openai"
)

func New(baseUrl, apiKey string, opts ...llms.Option) (*openai.Client, error) {
	provider := &llms.Provider{
		Type:    llms.ProviderTypeOpenAICompatible,
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
