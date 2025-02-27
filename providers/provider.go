package providers

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/recally-io/polyllm/llms"
)

// ProviderType is a string type for the names of supported LLM providers.
type ProviderType string

// Provider constants for all supported providers
const (
	ProviderTypeOpenAICompatible ProviderType = "openai-compatible"
	ProviderTypeOpenAI           ProviderType = "openai"
	ProviderTypeDeepSeek         ProviderType = "deepseek"
	ProviderTypeQwen             ProviderType = "qwen"
	ProviderTypeGemini           ProviderType = "gemini"
	ProviderTypeOpenRouter       ProviderType = "openrouter"
	ProviderTypeVolcengine       ProviderType = "volcengine"
	ProviderTypeGroq             ProviderType = "groq"
	ProviderTypeXai              ProviderType = "xai"
	ProviderTypeSiliconflow      ProviderType = "siliconflow"
	ProviderTypeTogether         ProviderType = "together"
	ProviderTypeFireworks        ProviderType = "fireworks"
)

// Provider represents a provider of LLM services.
type Provider struct {
	// Type is the type of the provider.
	Type ProviderType `json:"type"`
	// Name is the name of the provider.
	Name string `json:"name" `
	// BaseURL is the base URL of the provider's API.
	BaseURL string `json:"base_url,omitempty"`
	// APIKey is the API key for authentication.
	APIKey string `json:"api_key,omitempty"`
	// EnvPrefix is the environment variable name prefix for the API key.
	EnvPrefix string `json:"env_prefix,omitempty"`

	// ModelPrefix is the prefix of the provider's model name.
	ModelPrefix string `json:"model_prefix,omitempty"`
	// Models is a list of model ids.
	// In env it should be set as a comma separated string: "model1,model2"
	Models []llms.Model `json:"models,omitempty" `
	// ModelAlias is a map of model aliases to model ids.
	// In env it should be set as a key value value: "alias1=model1,alias2=model2"
	ModelAlias map[string]string `json:"model_alias,omitempty"`

	// HttpTimeout is the timeout for the HTTP client.
	HttpTimeout time.Duration `json:"timeout,omitempty"`
	// HttpClient is the HTTP client to use.
	HttpClient *http.Client `json:"-"`
}

func (p *Provider) Load() {
	if p.EnvPrefix != "" {
		getEnvValue := func(key string) string {
			return os.Getenv(p.EnvPrefix + key)
		}

		baseUrl := getEnvValue("BASE_URL")
		if baseUrl != "" {
			p.BaseURL = baseUrl
		}
		apiKey := getEnvValue("API_KEY")
		if apiKey != "" {
			p.APIKey = apiKey
		}

		models := strings.Split(getEnvValue("MODELS"), ",")
		if len(models) > 0 {
			p.Models = make([]llms.Model, 0)
			for _, model := range models {
				if model != "" {
					p.Models = append(p.Models, llms.Model{
						ID:     model,
						Name:   model,
						Object: "model",
					})
				}
			}
		}

		modelAlias := getEnvValue("MODEL_ALIAS")
		if modelAlias != "" {
			alias := strings.Split(modelAlias, ",")
			p.ModelAlias = map[string]string{}
			for _, a := range alias {
				p.ModelAlias[strings.Split(a, ":")[0]] = strings.Split(a, ":")[1]
			}
		}

		timeout := getEnvValue("TIMEOUT")
		if timeout != "" {
			timeoutInt, err := strconv.Atoi(timeout)
			if err == nil {
				p.HttpTimeout = time.Duration(timeoutInt) * time.Second
			}
		}
	}

	if p.HttpTimeout == 0 {
		p.HttpTimeout = 60 * time.Second
	}

	if p.HttpClient == nil {
		p.HttpClient = http.DefaultClient
	}
}

// GetRealModel returns the real model name based on the provider's prefix and model alias.
func (p *Provider) GetRealModel(model string) string {
	model = strings.TrimPrefix(model, p.ModelPrefix)
	if realModel, ok := p.ModelAlias[model]; ok {
		return realModel
	}
	return model
}

func (p *Provider) GetModelList(ctx context.Context) []llms.Model {
	models := make([]llms.Model, 0)

	if len(p.Models) > 0 {
		models = append(models, p.Models...)
	}

	if p.ModelAlias != nil {
		for alias := range p.ModelAlias {
			if p.ModelPrefix != "" {
				alias = p.ModelPrefix + alias
			}
			models = append(models, llms.Model{
				ID:     alias,
				Object: "model",
			})
		}
	}

	return models
}

func (p *Provider) SetHttpHeaders(req *http.Request, stream bool, extraHeaders map[string]string) {
	headers := map[string]string{
		"Authorization": "Bearer " + p.APIKey,
		"Content-Type":  "application/json",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
		"Accept":        "application/json",
		"HTTP-Referer":  "https://github.com/recally-io/polyllm",
		"X-Title":       "polyllm",
	}
	if stream {
		headers["Accept"] = "text/event-stream"
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}
}

func (p *Provider) ToOptions() []Option {
	opts := make([]Option, 0)

	if p.Name != "" {
		opts = append(opts, WithName(p.Name))
	}
	if p.BaseURL != "" {
		opts = append(opts, WithBaseURL(p.BaseURL))
	}
	if p.APIKey != "" {
		opts = append(opts, WithAPIKey(p.APIKey))
	}
	if p.EnvPrefix != "" {
		opts = append(opts, WithEnvPrefix(p.EnvPrefix))
	}
	if p.ModelPrefix != "" {
		opts = append(opts, WithModelPrefix(p.ModelPrefix))
	}
	if len(p.Models) != 0 {
		opts = append(opts, WithModels(p.Models))
	}
	if len(p.ModelAlias) != 0 {
		opts = append(opts, WithModelAlias(p.ModelAlias))
	}
	if p.HttpTimeout != 0 {
		opts = append(opts, WithHttpTimeout(p.HttpTimeout))
	}
	if p.HttpClient != nil {
		opts = append(opts, WithHttpClient(p.HttpClient))
	}
	return opts
}
