package polyllm

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/recally-io/polyllm/llms"
	"github.com/recally-io/polyllm/providers"
)

//go:embed providers.json
var builtInProvidersBytes []byte
var builtInProviders []*providers.Provider

type PolyLLM struct {
	clients               []llms.LLM
	modelProviderMappings map[string]*providers.Provider
	modelClientMappings   map[string]llms.LLM
}

func New() *PolyLLM {
	p := &PolyLLM{
		clients:               make([]llms.LLM, 0),
		modelProviderMappings: make(map[string]*providers.Provider),
		modelClientMappings:   make(map[string]llms.LLM),
	}

	if err := json.Unmarshal(builtInProvidersBytes, &builtInProviders); err == nil {
		p.AddProviders(builtInProviders...)
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
			slog.Error("failed to create client", "err", err, "provider", provider.Name)
			return
		}
		p.clients = append(p.clients, client)

		models, err := p.loadProviderModelsWithCache(context.Background(), client)
		if err != nil {
			slog.Error("failed to list models", "err", err, "provider", provider.Name)
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
			slog.Error("failed to list models", "err", err, "provider", client.GetProviderName())
			continue
		}
		models = append(models, clientModels...)
	}
	return models, nil
}

func (p *PolyLLM) ChatCompletion(ctx context.Context, req llms.ChatCompletionRequest, streamingFunc func(resp llms.StreamingChatCompletionResponse), options ...llms.RequestOption) {
	client, model, err := p.GetProvider(req.Model)
	if err != nil {
		slog.Error("failed to get provider", "err", err, "model", req.Model)
		streamingFunc(llms.StreamingChatCompletionResponse{Err: err})
		return
	}
	req.Model = model
	client.ChatCompletion(ctx, req, streamingFunc, options...)
}

func (p *PolyLLM) GenerateText(ctx context.Context, model, prompt string, options ...llms.RequestOption) (string, error) {
	client, model, err := p.GetProvider(model)
	if err != nil {
		slog.Error("failed to get provider", "err", err, "model", model)
		return "", err
	}
	return client.GenerateText(ctx, model, prompt, options...)
}

func (p *PolyLLM) StreamGenerateText(ctx context.Context, model, prompt string, streamingTextFunc func(resp llms.StreamingChatCompletionText), options ...llms.RequestOption) {
	client, model, err := p.GetProvider(model)
	if err != nil {
		slog.Error("failed to get provider", "err", err, "model", model)
		streamingTextFunc(llms.StreamingChatCompletionText{Err: err})
		return
	}
	client.StreamGenerateText(ctx, model, prompt, streamingTextFunc, options...)
}

func (p *PolyLLM) GetClientForModel(model string) llms.LLM {
	return p.modelClientMappings[model]
}

func (p *PolyLLM) GetProvider(model string) (llms.LLM, string, error) {
	provider, ok := p.modelProviderMappings[model]
	if !ok {
		return nil, "", ErrProviderNotFound
	}

	client, ok := p.modelClientMappings[model]
	if !ok {
		return nil, "", ErrProviderNotFound
	}

	return client, provider.GetRealModel(model), nil
}

func (p *PolyLLM) loadProviderModelsWithCache(ctx context.Context, client llms.LLM) ([]llms.Model, error) {
	// Try to load models from cache
	modelCache, err := loadModelCache(client.GetProviderName())
	if err == nil && isModelCacheValid(modelCache) {
		slog.Debug("using cached models", "timestamp", modelCache.Timestamp.Format(time.RFC1123))
		return modelCache.Models, nil
	}

	slog.Debug("loading models from providers")
	// load models using providers

	providerModels, err := client.ListModels(ctx)
	if err != nil {
		slog.Error("failed to list models", "err", err)
	}

	if len(providerModels) == 0 {
		return nil, errors.New("no models found")
	}

	modelCache.Models = providerModels
	modelCache.Timestamp = time.Now()
	if err := saveModelCache(client.GetProviderName(), modelCache); err != nil {
		slog.Error("failed to save model cache", "err", err)
	}
	return modelCache.Models, nil
}
