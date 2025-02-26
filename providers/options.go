package providers

import (
	"net/http"
	"time"

	"github.com/recally-io/polyllm/llms"
)

type Option func(*Provider)

func New(opts ...Option) *Provider {
	p := &Provider{
		HttpTimeout: 60 * time.Second,
		HttpClient:  http.DefaultClient,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func WithName(name string) Option {
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

func WithHttpTimeout(timeout time.Duration) Option {
	return func(p *Provider) {
		p.HttpTimeout = timeout
	}
}

func WithHttpClient(client *http.Client) Option {
	return func(p *Provider) {
		p.HttpClient = client
	}
}
