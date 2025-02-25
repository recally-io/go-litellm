package llms

import (
	"net/http"
	"time"
)

// Config is the configuration for the LLM client.
type Config struct {
	// BaseURL is the base URL of the LLM provider.
	BaseURL string `json:"base_url,omitempty"`
	// APIKey is the API key for the LLM provider.
	APIKey  string `json:"api_key"`
	// Prefix is the prefix to use for the model name.
	Prefix  string `json:"prefix,omitempty"`

	// Timeout is the timeout for the HTTP client.
	Timeout    time.Duration `json:"timeout,omitempty"`
	// HttpClient is the HTTP client to use.
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

// WithBaseURL sets the base URL for the LLM provider.
// baseURL: The base URL to use for API requests
func WithBaseURL(baseURL string) ConfigOptions {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithPrefix sets the prefix for the model name.
// prefix: The prefix to prepend to model names
func WithPrefix(prefix string) ConfigOptions {
	return func(c *Config) {
		c.Prefix = prefix
	}
}

// WithTimeout sets the timeout for HTTP requests.
// timeout: The duration after which requests will timeout
func WithTimeout(timeout time.Duration) ConfigOptions {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithHttpClient sets a custom HTTP client.
// client: The *http.Client to use for requests
func WithHttpClient(client *http.Client) ConfigOptions {
	return func(c *Config) {
		c.HttpClient = client
	}
}

// WithAPIKey sets the API key for authentication.
// apiKey: The API key to use for authentication
func WithAPIKey(apiKey string) ConfigOptions {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}
