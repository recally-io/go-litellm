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

// ProviderName is a string type for the names of supported LLM providers.
type ProviderName string

// Provider constants for all supported providers
const (
	ProviderNameOpenAICompatible ProviderName = "openai-compatible"
	ProviderNameOpenAI           ProviderName = "openai"
	ProviderNameDeepSeek         ProviderName = "deepseek"
	ProviderNameQwen             ProviderName = "qwen"
	ProviderNameGemini           ProviderName = "gemini"
	ProviderNameOpenRouter       ProviderName = "openrouter"
	ProviderNameVolcengine       ProviderName = "volcengine"
	ProviderNameGroq             ProviderName = "groq"
	ProviderNameXai              ProviderName = "xai"
	ProviderNameSiliconflow      ProviderName = "siliconflow"
)

// Provider represents a provider of LLM services.
type Provider struct {
	llms.LLM

	// Name is the name of the provider.
	Name ProviderName `json:"name" `
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

	// Timeout is the timeout for the HTTP client.
	Timeout time.Duration `json:"timeout,omitempty"`
	// HttpClient is the HTTP client to use.
	HttpClient *http.Client `json:"-"`
}

type Option func(*Provider)

func WithName(name ProviderName) Option {
	return func(p *Provider) {
		p.Name = name
	}
}

func WithBaseURL(url string) Option {
	return func(p *Provider) {
		p.BaseURL = url
	}
}

func WithAPIKey(apiKey string) Option {
	return func(p *Provider) {
		p.APIKey = apiKey
	}
}

func WithEnvPrefix(envPrefix string) Option {
	return func(p *Provider) {
		p.EnvPrefix = envPrefix
	}
}

func WithModelPrefix(prefix string) Option {
	return func(p *Provider) {
		p.ModelPrefix = prefix
	}
}

func WithModels(models []llms.Model) Option {
	return func(p *Provider) {
		p.Models = models
	}
}

func WithModelAlias(alias map[string]string) Option {
	return func(p *Provider) {
		p.ModelAlias = alias
	}
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
				p.Timeout = time.Duration(timeoutInt) * time.Second
			}
		}
	}

	if p.Timeout == 0 {
		p.Timeout = 60 * time.Second
	}

	if p.HttpClient == nil {
		p.HttpClient = http.DefaultClient
	}
}

// GetRealModel returns the real model name based on the provider's prefix and model alias.
func (p *Provider) GetRealModel(model string) string {
	if realModel, ok := p.ModelAlias[model]; ok {
		return realModel
	}
	return strings.TrimPrefix(model, p.ModelPrefix)
}

func (p *Provider) GetModelList(ctx context.Context) ([]llms.Model, error) {
	models := make([]llms.Model, 0)

	if len(p.Models) > 0 {
		models = append(models, p.Models...)
	}

	if p.ModelAlias != nil {
		for alias, model := range p.ModelAlias {
			models = append(models, llms.Model{
				ID:     model,
				Name:   alias,
				Object: "model",
			})
		}
	}

	if len(models) == 0 {
		fetchedModels, err := p.LLM.ListModels(ctx)
		if err != nil {
			return models, err
		}
		models = append(models, fetchedModels...)
	}

	return models, nil
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

func New(opts ...Option) *Provider {
	p := &Provider{
		Timeout:    60 * time.Second,
		HttpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}
