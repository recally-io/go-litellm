package litellm

import (
	_ "embed"
	"encoding/json"
	"errors"
	"os"

	"github.com/recally-io/go-litellm/llms"
	"github.com/recally-io/go-litellm/llms/openai"
)

//go:embed llm-providers.json
var providersJSON []byte

var providers map[string]Provider

// Provider represents a provider of LLM services.
type Provider struct {
	// Name is the name of the provider.
	Name    string `json:"name"`
	// BaseURL is the base URL of the provider's API.
	BaseURL string `json:"baseUrl"`
	// Prefix is the prefix to use for listing models, the prefix will be added to the model name when calling the provider API.
	// For example, if the model name is "gpt-4" and the prefix is "openai", the model name will be "openai/gpt-4".
	Prefix string `json:"prefix"`
}

// ProviderName is a string type for the names of supported LLM providers.
type ProviderName string

// ProviderOpenAICompatible is the provider name for OpenAI-compatible services.
const ProviderOpenAICompatible ProviderName = "openai-compatible"
// ProviderNameOpenAI is the provider name for OpenAI.
const ProviderNameOpenAI ProviderName = "openai"
// ProviderNameDeepSeek is the provider name for DeepSeek.
const ProviderNameDeepSeek ProviderName = "deepseek"
// ProviderNameQwen is the provider name for Qwen.
const ProviderNameQwen ProviderName = "qwen"
// ProviderNameGemini is the provider name for Gemini.
const ProviderNameGemini ProviderName = "gemini"
// ProviderNameOpenRouter is the provider name for OpenRouter.
const ProviderNameOpenRouter ProviderName = "openrouter"
// ProviderNameVolcengine is the provider name for Volcengine.
const ProviderNameVolcengine ProviderName = "volcengine"
// ProviderNameGroq is the provider name for Groq.
const ProviderNameGroq ProviderName = "groq"
// ProviderNameXai is the provider name for Xai.
const ProviderNameXai ProviderName = "xai"

// init initializes the providers map by unmarshaling the providersJSON.
func init() {
	if err := json.Unmarshal([]byte(providersJSON), &providers); err != nil {
		panic("failed to unmarshal providers.json: " + err.Error())
	}
}

// New creates a new LLM client for the specified provider.
func New(providerName ProviderName, opts ...llms.ConfigOptions) (llms.LLM, error) {
	switch providerName {
	case ProviderNameOpenAI:
		client := openai.New(opts...)
		if err := setApiKeyFromEnv(&client.Config, "OPENAI_API_KEY"); err != nil {
			return nil, err
		}
		return client, nil
	case ProviderOpenAICompatible:
		oai := openai.New(opts...)
		if oai.BaseURL == "" {
			return nil, errors.New("openai baseUrl is empty")
		}
		if oai.APIKey == "" {
			return nil, errors.New("openai apiKey is empty")
		}
		return oai, nil
	case ProviderNameDeepSeek:
		return newOpenAICompatibleClient(providerName, "DEEPSEEK_API_KEY", opts...)
	case ProviderNameQwen:
		return newOpenAICompatibleClient(providerName, "QWEN_API_KEY", opts...)
	case ProviderNameGemini:
		return newOpenAICompatibleClient(providerName, "GEMINI_API_KEY", opts...)
	case ProviderNameOpenRouter:
		return newOpenAICompatibleClient(providerName, "OPENROUTER_API_KEY", opts...)
	case ProviderNameVolcengine:
		return newOpenAICompatibleClient(providerName, "VOLCENGINE_API_KEY", opts...)
	case ProviderNameGroq:
		return newOpenAICompatibleClient(providerName, "GROQ_API_KEY", opts...)
	case ProviderNameXai:
		return newOpenAICompatibleClient(providerName, "XAI_API_KEY", opts...)
	default:
		return nil, errors.New("provider " + string(providerName) + " not found")
	}
}

// setApiKeyFromEnv sets the API key from the environment variable if it is not already set in the config.
func setApiKeyFromEnv(cfg *llms.Config, apiKeyEnvName string) error {
	if cfg.APIKey != "" {
		return nil
	}

	cfg.APIKey = os.Getenv(apiKeyEnvName)
	if cfg.APIKey == "" {
		return errors.New("environment variable " + apiKeyEnvName + " is empty")
	}
	return nil
}

// newOpenAICompatibleClient creates a new OpenAI-compatible client with the specified provider name and API key environment variable name.
func newOpenAICompatibleClient(providerName ProviderName, apiKeyEnvName string, opts ...llms.ConfigOptions) (llms.LLM, error) {
	provider := providers[string(providerName)]
	defaultOpts := []llms.ConfigOptions{
		llms.WithBaseURL(provider.BaseURL),
		llms.WithPrefix(provider.Prefix),
	}
	opts = append(defaultOpts, opts...)
	client := openai.New(opts...)
	if err := setApiKeyFromEnv(&client.Config, apiKeyEnvName); err != nil {
		return nil, err
	}
	return client, nil
}
