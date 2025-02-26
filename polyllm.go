package polyllm

import (
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"

	"github.com/recally-io/polyllm/llms"
	"github.com/recally-io/polyllm/providers"
)

//go:embed providers.json
var builtInProvidersBytes []byte
var builtInProviders []*providers.Provider

type PolyLLM struct {
	clients             []llms.LLM
	modelClientMappings map[string]llms.LLM
}

func New() *PolyLLM {
	p := &PolyLLM{
		clients:             make([]llms.LLM, 0),
		modelClientMappings: make(map[string]llms.LLM),
	}

	if err := json.Unmarshal(builtInProvidersBytes, &builtInProviders); err == nil {
		p.AddProviders(builtInProviders...)
	}

	return p
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

		models, err := client.ListModels(context.Background())
		if err != nil {
			slog.Error("failed to list models", "err", err, "provider", provider.Name)
			return
		}
		for _, model := range models {
			p.modelClientMappings[model.ID] = client
		}
	}
}

func (p *PolyLLM) ListModels(ctx context.Context) ([]llms.Model, error) {
	models := make([]llms.Model, 0)
	for _, client := range p.clients {
		providerModels, err := client.ListModels(ctx)
		if err != nil {
			return nil, err
		}
		models = append(models, providerModels...)
	}
	return models, nil
}

func (p *PolyLLM) ChatCompletion(ctx context.Context, req llms.ChatCompletionRequest, streamingFunc func(resp llms.StreamingChatCompletionResponse), options ...llms.RequestOption) {
	client := p.GetClientForModel(req.Model)
	if client == nil {
		slog.Error("client not found", "model", req.Model)
		streamingFunc(llms.StreamingChatCompletionResponse{Err: ErrProviderNotFound})
		return
	}
	client.ChatCompletion(ctx, req, streamingFunc, options...)
}

func (p *PolyLLM) GenerateText(ctx context.Context, model, prompt string, options ...llms.RequestOption) (string, error) {
	client := p.GetClientForModel(model)
	if client == nil {
		return "", ErrProviderNotFound
	}
	return client.GenerateText(ctx, model, prompt, options...)
}

func (p *PolyLLM) StreamGenerateText(ctx context.Context, model, prompt string, streamingTextFunc func(resp llms.StreamingChatCompletionText), options ...llms.RequestOption) {
	client := p.GetClientForModel(model)
	if client == nil {
		slog.Error("client not found", "model", model)
		streamingTextFunc(llms.StreamingChatCompletionText{Err: ErrProviderNotFound})
		return
	}
	client.StreamGenerateText(ctx, model, prompt, streamingTextFunc, options...)
}

func (p *PolyLLM) GetClientForModel(model string) llms.LLM {
	return p.modelClientMappings[model]
}
