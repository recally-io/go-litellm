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

type Provider struct {
	Name    string `json:"name"`
	BaseURL string `json:"baseUrl"`
	// Prefix is the prefix to use for listing models, the prefix will be added to the model name when calling the provider API. for example, if the model name is "gpt-4" and the prefix is "openai", the model name will be "openai/gpt-4".
	Prefix string `json:"prefix"`
}

type ProviderName string

const ProviderOpenAICompatible ProviderName = "openai-compatible"
const ProviderNameOpenAI ProviderName = "openai"
const ProviderNameDeepSeek ProviderName = "deepseek"
const ProviderNameQwen ProviderName = "qwen"
const ProviderNameGemini ProviderName = "gemini"
const ProviderNameOpenRouter ProviderName = "openrouter"
const ProviderNameVolcengine ProviderName = "volcengine"
const ProviderNameGroq ProviderName = "groq"
const ProviderNameXai ProviderName = "xai"

func init() {
	if err := json.Unmarshal([]byte(providersJSON), &providers); err != nil {
		panic("failed to unmarshal providers.json: " + err.Error())
	}
}

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
