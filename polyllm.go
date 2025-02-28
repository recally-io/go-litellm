package polyllm

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"maps"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/recally-io/polyllm/llms"
	"github.com/recally-io/polyllm/logger"
	"github.com/recally-io/polyllm/mcps"
	"github.com/recally-io/polyllm/providers"
)

//go:embed providers.json
var builtInProvidersBytes []byte
var builtInProviders []*providers.Provider

type PolyLLM struct {
	Options
	clients               []llms.LLM
	modelProviderMappings map[string]*providers.Provider
	modelClientMappings   map[string]llms.LLM

	mcpClientMappings map[string]mcpclient.MCPClient
}

type Options struct {
	providers  []*providers.Provider
	mcpServers map[string]mcps.ServerConfig
}

type Option func(*Options)

func WithMCPServerConfig(mcpServers map[string]mcps.ServerConfig) Option {
	return func(p *Options) {
		p.mcpServers = mcpServers
	}
}

func WithProviders(providers ...*providers.Provider) Option {
	return func(p *Options) {
		p.providers = providers
	}
}

func New(opts ...Option) *PolyLLM {
	p := &PolyLLM{
		clients:               make([]llms.LLM, 0),
		modelProviderMappings: make(map[string]*providers.Provider),
		modelClientMappings:   make(map[string]llms.LLM),
	}

	if err := json.Unmarshal(builtInProvidersBytes, &builtInProviders); err != nil {
		logger.DefaultLogger.Error("failed to unmarshal built-in providers", "err", err)
	}

	for _, opt := range opts {
		opt(&p.Options)
	}

	// add built-in providers and user-provided providers
	providers := append(builtInProviders, p.providers...)
	p.AddProviders(providers...)

	// start MCP clients
	if len(p.mcpServers) > 0 {
		mcpClients, err := mcps.CreateMCPClients(p.mcpServers)
		if err != nil {
			logger.DefaultLogger.Error("failed to create MCP clients", "err", err)
		}
		p.mcpClientMappings = mcpClients
	}

	return p
}

func (p *PolyLLM) GetProviderName() string {
	return "polyllm"
}

func (p *PolyLLM) AddProviders(providers ...*providers.Provider) {
	for _, provider := range providers {
		p.AddProvider(provider)
	}
}

func (p *PolyLLM) AddProvider(provider *providers.Provider) {
	provider.Load()
	if provider.APIKey != "" {
		client, err := NewClient(provider)
		if err != nil {
			logger.DefaultLogger.Error("failed to create client", "err", err, "provider", provider.Name)
			return
		}
		p.clients = append(p.clients, client)

		models, err := p.loadProviderModelsWithCache(context.Background(), client)
		if err != nil {
			return
		}
		for _, model := range models {
			p.modelClientMappings[model.ID] = client
			p.modelProviderMappings[model.ID] = provider
		}
	}
}

func (p *PolyLLM) ListModels(ctx context.Context) ([]llms.Model, error) {
	models := make([]llms.Model, 0)
	for _, client := range p.clients {
		clientModels, err := p.loadProviderModelsWithCache(ctx, client)
		if err != nil {
			continue
		}
		models = append(models, clientModels...)
	}
	return models, nil
}

func (p *PolyLLM) ChatCompletion(ctx context.Context, req llms.ChatCompletionRequest, streamingFunc func(resp llms.StreamingChatCompletionResponse), options ...llms.RequestOption) {
	client, model, tools, err := p.preProcess(ctx, req.Model)
	if err != nil {
		logger.DefaultLogger.Error("failed to get provider", "err", err, "model", req.Model)
		streamingFunc(llms.StreamingChatCompletionResponse{Err: err})
		return
	}
	if len(tools) > 0 {
		req.Tools = append(req.Tools, tools...)
	}
	req.Model = model
	client.ChatCompletion(ctx, req, streamingFunc, options...)
}

func (p *PolyLLM) GenerateText(ctx context.Context, model, prompt string, options ...llms.RequestOption) (string, error) {
	client, model, tools, err := p.preProcess(ctx, model)
	if err != nil {
		logger.DefaultLogger.Error("failed to get provider", "err", err, "model", model)
		return "", err
	}
	options = append(options, llms.WithTools(tools))
	return client.GenerateText(ctx, model, prompt, options...)
}

func (p *PolyLLM) StreamGenerateText(ctx context.Context, model, prompt string, streamingTextFunc func(resp llms.StreamingChatCompletionText), options ...llms.RequestOption) {
	client, model, tools, err := p.preProcess(ctx, model)
	if err != nil {
		logger.DefaultLogger.Error("failed to get provider", "err", err, "model", model)
		streamingTextFunc(llms.StreamingChatCompletionText{Err: err})
		return
	}
	options = append(options, llms.WithTools(tools))
	client.StreamGenerateText(ctx, model, prompt, streamingTextFunc, options...)
}

func (p *PolyLLM) GetClientForModel(model string) llms.LLM {
	return p.modelClientMappings[model]
}

func (p *PolyLLM) preProcess(ctx context.Context, model string) (llms.LLM, string, []llms.Tool, error) {
	info := strings.Split(model, "?")
	model = info[0]

	provider, ok := p.modelProviderMappings[model]
	if !ok {
		return nil, "", nil, ErrProviderNotFound
	}

	client, ok := p.modelClientMappings[model]
	if !ok {
		return nil, "", nil, ErrProviderNotFound
	}
	providerModel := provider.GetRealModel(model)

	tools := []llms.Tool{}
	if len(info) > 1 {
		tools = p.GetMCPTools(ctx, info[1])
	}

	return client, providerModel, tools, nil
}

func (p *PolyLLM) loadProviderModelsWithCache(ctx context.Context, client llms.LLM) ([]llms.Model, error) {
	// Try to load models from cache
	modelCache, err := loadModelCache(client.GetProviderName())
	if err == nil && isModelCacheValid(modelCache) {
		logger.DefaultLogger.Debug("using cached models", "timestamp", modelCache.Timestamp.Format(time.RFC1123))
		return modelCache.Models, nil
	}

	logger.DefaultLogger.Debug("loading models from providers")
	// load models using providers

	providerModels, err := client.ListModels(ctx)
	if err != nil {
		logger.DefaultLogger.Error("failed to list models", "provider", client.GetProviderName(), "err", err)
	}

	if len(providerModels) == 0 {
		return nil, errors.New("no models found")
	}

	modelCache.Models = providerModels
	modelCache.Timestamp = time.Now()
	if err := saveModelCache(client.GetProviderName(), modelCache); err != nil {
		logger.DefaultLogger.Error("failed to save model cache", "err", err)
	}
	return modelCache.Models, nil
}

func (p *PolyLLM) GetMCPTools(ctx context.Context, modelInfo string) []llms.Tool {
	// model=gpt-4o?mcp=fetch,everything&tools=fetch,everything
	// Extract MCP servers from model string
	llmTools := make([]llms.Tool, 0)
	params := strings.Split(modelInfo, "&")
	for _, param := range params {
		parts := strings.Split(param, "=")
		if len(parts) != 2 {
			logger.DefaultLogger.Error("invalid param", "param", param)
			continue
		}
		if parts[0] == "mcp" {
			mcpNames := strings.Split(parts[1], ",")
			if slices.Contains(mcpNames, "all") {
				mcpNames = slices.Sorted(maps.Keys(p.mcpClientMappings))
			}
			for _, name := range mcpNames {
				name = strings.TrimSpace(name)
				if client, ok := p.mcpClientMappings[name]; ok {
					mcpTools, err := client.ListTools(ctx, mcp.ListToolsRequest{})
					if err != nil {
						logger.DefaultLogger.Error("failed to list tools", "err", err, "mcp_server", name)
						continue
					}
					for _, tool := range mcpTools.Tools {
						llmTools = append(llmTools, mcpToolToLlmsTool(name, tool))
					}
				}
			}
		}
	}
	return llmTools
}

func (p *PolyLLM) GetMCPClientByToolName(ctx context.Context, toolName string) (mcpclient.MCPClient, string, error) {
	params := strings.Split(toolName, "_")
	if len(params) != 3 || params[0] != "mcp" {
		return nil, "", fmt.Errorf("tool name must be in format mcp_{mcp_server_name}_{tool_name}")
	}
	mcpName := params[1]
	mcpToolName := params[2]
	if client, ok := p.mcpClientMappings[mcpName]; ok {
		return client, mcpToolName, nil
	}
	return nil, "", fmt.Errorf("tool %s not found", toolName)
}

func mcpToolToLlmsTool(mcpName string, tool mcp.Tool) llms.Tool {
	return llms.Tool{
		Type: llms.ToolTypeFunction,
		Function: &llms.FunctionDefinition{
			Name:        fmt.Sprintf("mcp_%s_%s", mcpName, tool.Name),
			Description: tool.Description,
			Parameters:  tool.InputSchema,
		},
	}
}
