package polyllm

import (
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/recally-io/polyllm/llms"
	"github.com/recally-io/polyllm/llms/openai"
	"github.com/recally-io/polyllm/pkg/providers"
)

//go:embed providers.json
var builtInProvidersBytes []byte
var builtInProviders []*providers.Provider

type PolyLLM struct {
	mu                    sync.RWMutex
	providers             []*providers.Provider
	modelProviderMappings map[string]*providers.Provider
}

func New() *PolyLLM {
	p := &PolyLLM{
		modelProviderMappings: make(map[string]*providers.Provider),
	}

	if err := json.Unmarshal(builtInProvidersBytes, &builtInProviders); err == nil {
		p.AddProviders(builtInProviders...)
	}

	return p
}

func (p *PolyLLM) AddProviders(providers ...*providers.Provider) {
	for _, provider := range providers {
		provider.Load()
		if provider.APIKey != "" {
			p.initProviderLLM(provider)
			p.initProviderModels(provider)
		}
	}
}

func (p *PolyLLM) initProviderLLM(provider *providers.Provider) {
	switch provider.Name {
	case providers.ProviderNameOpenAI, providers.ProviderNameDeepSeek, providers.ProviderNameQwen, providers.ProviderNameGemini, providers.ProviderNameOpenRouter, providers.ProviderNameVolcengine, providers.ProviderNameGroq, providers.ProviderNameXai, providers.ProviderNameSiliconflow:
		provider.LLM = &openai.OpenAI{Provider: provider}
	default:
		slog.Error("unsupported provider", "provider", provider.Name)
	}
}

func (p *PolyLLM) initProviderModels(provider *providers.Provider) {
	slog.Info("initializing provider", "provider", provider.Name)
	models, err := provider.GetModelList(context.Background())
	if err != nil {
		slog.Error("failed to get models", "err", err, "provider", provider.Name)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.providers = append(p.providers, provider)
	for _, model := range models {
		p.modelProviderMappings[model.Name] = provider
	}
}

func (p *PolyLLM) GetProviderForModel(model string) *providers.Provider {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.modelProviderMappings[model]
}

func (p *PolyLLM) ListModels(ctx context.Context) ([]llms.Model, error) {
	models := make([]llms.Model, 0)
	for _, provider := range p.providers {
		providerModels, err := provider.GetModelList(ctx)
		if err != nil {
			return nil, err
		}
		models = append(models, providerModels...)
	}
	return models, nil
}

func (p *PolyLLM) ChatCompletion(ctx context.Context, req llms.ChatCompletionRequest, streamingFunc func(resp llms.StreamingChatCompletionResponse), options ...llms.RequestOption) {
	provider := p.GetProviderForModel(req.Model)
	if provider == nil {
		slog.Error("provider not found", "model", req.Model)
		streamingFunc(llms.StreamingChatCompletionResponse{Err: ErrProviderNotFound})
		return
	}
	provider.ChatCompletion(ctx, req, streamingFunc, options...)
}

func (p *PolyLLM) GenerateText(ctx context.Context, model, prompt string, options ...llms.RequestOption) (string, error) {
	provider := p.GetProviderForModel(model)
	if provider == nil {
		return "", ErrProviderNotFound
	}
	return provider.GenerateText(ctx, model, prompt, options...)
}

func (p *PolyLLM) StreamGenerateText(ctx context.Context, model, prompt string, streamingTextFunc func(resp llms.StreamingChatCompletionText), options ...llms.RequestOption) {
	provider := p.GetProviderForModel(model)
	if provider == nil {
		slog.Error("provider not found", "model", model)
		streamingTextFunc(llms.StreamingChatCompletionText{Err: ErrProviderNotFound})
		return
	}
	provider.StreamGenerateText(ctx, model, prompt, streamingTextFunc, options...)
}
