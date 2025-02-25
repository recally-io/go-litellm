package llms

import (
	"net/http"
	"time"
)

type Config struct {
	BaseURL string `json:"base_url,omitempty"`
	APIKey  string `json:"api_key"`
	Prefix  string `json:"prefix,omitempty"`

	Timeout    time.Duration `json:"timeout,omitempty"`
	HttpClient *http.Client  `json:"-"`
}

func (c *Config) SetHttpHeaders(req *http.Request, stream bool, extraHeaders map[string]string) {
	headers := map[string]string{
		"Authorization": "Bearer " + c.APIKey,
		"Content-Type":  "application/json",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
		"Accept":        "application/json",
		"HTTP-Referer":  "https://github.com/recally-io/go-litellm",
		"X-Title":       "go-litellm",
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

func NewConfig(opts ...ConfigOptions) Config {
	config := Config{
		Timeout:    60 * time.Second,
		HttpClient: http.DefaultClient,
		Prefix:     "",
	}
	for _, opt := range opts {
		opt(&config)
	}

	if config.Timeout != 0 {
		config.HttpClient.Timeout = config.Timeout
	}

	return config
}

type ConfigOptions func(*Config)

func WithBaseURL(baseURL string) ConfigOptions {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

func WithPrefix(prefix string) ConfigOptions {
	return func(c *Config) {
		c.Prefix = prefix
	}
}

func WithTimeout(timeout time.Duration) ConfigOptions {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

func WithHttpClient(client *http.Client) ConfigOptions {
	return func(c *Config) {
		c.HttpClient = client
	}
}

func WithAPIKey(apiKey string) ConfigOptions {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}
