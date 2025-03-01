package polyllm

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/recally-io/polyllm/llms"
	"github.com/recally-io/polyllm/mcps"
)

//go:embed builtin-llm-providers.json
var builtInLLMProvidersBytes []byte
var builtInLLMProviders []llms.Provider

type PolyLLM struct {
	Config
	llms              []LLM
	modelLLMMappings  map[string]LLM
	mcpClientMappings map[string]mcpclient.MCPClient
}

type Config struct {
	LLMProvides  []llms.Provider          `json:"llms"`
	MCPProviders map[string]mcps.Provider `json:"mcps"`
}

func init() {
	if err := json.Unmarshal(builtInLLMProvidersBytes, &builtInLLMProviders); err != nil {
		slog.Error("failed to unmarshal built-in llms", "err", err)
	}
}

func LoadConfig(configPath string) (Config, error) {
	var cfg Config
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := json.Unmarshal(configFile, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to load config file: %w", err)
	}
	return cfg, nil
}

type Option func(*Config)

func WithMCPProviders(mcpServers map[string]mcps.Provider) Option {
	return func(c *Config) {
		c.MCPProviders = mcpServers
	}
}

func WithLLMProviders(providers ...llms.Provider) Option {
	return func(c *Config) {
		c.LLMProvides = append(c.LLMProvides, providers...)
	}
}

func NewFromConfig(cfg Config) *PolyLLM {
	cfg.LLMProvides = append(builtInLLMProviders, cfg.LLMProvides...)
	p := &PolyLLM{
		llms:              make([]LLM, 0),
		modelLLMMappings:  make(map[string]LLM),
		mcpClientMappings: make(map[string]mcpclient.MCPClient),
		Config:            cfg,
	}
	// add llm providers
	p.addLLMProviders(p.Config.LLMProvides...)
	// add mcp providers
	p.addMCPProviders(p.Config.MCPProviders)

	return p
}

func New(opts ...Option) *PolyLLM {
	cfg := Config{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return NewFromConfig(cfg)
}

func (p *PolyLLM) addLLMProviders(providers ...llms.Provider) {
	for _, provider := range providers {
		provider.Load()
		if provider.APIKey != "" {
			llm, err := NewLLM(&provider)
			if err != nil {
				slog.Error("failed to create llm client", "provider", provider.Name, "err", err)
				continue
			}
			p.llms = append(p.llms, llm)

			models, err := p.loadProviderModelsWithCache(context.Background(), llm)
			if err != nil {
				slog.Error("failed to load llm models", "provider", provider.Name, "err", err)
				continue
			}
			for _, model := range models {
				p.modelLLMMappings[model.ID] = llm
			}
		}
	}
}

func (p *PolyLLM) addMCPProviders(providers map[string]mcps.Provider) {
	// start MCP clients
	if len(providers) > 0 {
		mcpClients, err := mcps.CreateMCPClients(providers)
		if err != nil {
			slog.Error("failed to create MCP clients", "err", err)
		}
		p.mcpClientMappings = mcpClients
	}
}

func (p *PolyLLM) GetLLMByModel(model string) (LLM, error) {
	llm, ok := p.modelLLMMappings[model]
	if !ok {
		return nil, fmt.Errorf("model %s not found", model)
	}
	return llm, nil
}
