package providers

import (
	"strings"

	"github.com/recally-io/polyllm"
)

// ProviderInfo represents information about a provider
type ProviderInfo struct {
	Name         string
	ProviderName polyllm.ProviderName
}

// GetAllProviders returns a list of all supported providers
func GetAllProviders() []ProviderInfo {
	return []ProviderInfo{
		{"OpenAI", polyllm.ProviderNameOpenAI},
		{"DeepSeek", polyllm.ProviderNameDeepSeek},
		{"Qwen", polyllm.ProviderNameQwen},
		{"Gemini", polyllm.ProviderNameGemini},
		{"OpenRouter", polyllm.ProviderNameOpenRouter},
		{"Volcengine", polyllm.ProviderNameVolcengine},
		{"Groq", polyllm.ProviderNameGroq},
		{"Xai", polyllm.ProviderNameXai},
	}
}

// DetermineProvider determines the provider from the model name
func DetermineProvider(modelName string) (polyllm.ProviderName, string) {
	// Default to OpenAI if no prefix is provided
	if !strings.Contains(modelName, "/") {
		return polyllm.ProviderNameOpenAI, modelName
	}
	
	// Split the model name by "/"
	parts := strings.SplitN(modelName, "/", 2)
	prefix := parts[0]
	model := parts[1]
	
	// Map prefix to provider
	switch prefix {
	case "openai":
		return polyllm.ProviderNameOpenAI, model
	case "deepseek":
		return polyllm.ProviderNameDeepSeek, model
	case "qwen":
		return polyllm.ProviderNameQwen, model
	case "gemini":
		return polyllm.ProviderNameGemini, model
	case "openrouter":
		return polyllm.ProviderNameOpenRouter, model
	case "volcengine":
		return polyllm.ProviderNameVolcengine, model
	case "groq":
		return polyllm.ProviderNameGroq, model
	case "xai":
		return polyllm.ProviderNameXai, model
	default:
		// If the prefix is not recognized, use OpenAI-compatible with the full model name
		return polyllm.ProviderOpenAICompatible, modelName
	}
}
