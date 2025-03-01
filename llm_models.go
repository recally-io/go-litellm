package polyllm

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/recally-io/polyllm/llms"
)

func (p *PolyLLM) ListModels(ctx context.Context) ([]llms.Model, error) {
	models := make([]llms.Model, 0)
	for _, client := range p.llms {
		clientModels, err := p.loadProviderModelsWithCache(ctx, client)
		if err != nil {
			continue
		}
		models = append(models, clientModels...)
	}
	return models, nil
}

func (p *PolyLLM) loadProviderModelsWithCache(ctx context.Context, llm LLM) ([]llms.Model, error) {
	// Try to load models from cache
	modelCache, err := llms.LoadModelCache(llm.GetProvider().Name)
	if err == nil && llms.IsModelCacheValid(modelCache) {
		slog.Debug("using cached models", "timestamp", modelCache.Timestamp.Format(time.RFC1123))
		return modelCache.Models, nil
	}

	slog.Debug("loading models from llms")
	// load models using llms

	providerModels, err := llm.ListModels(ctx)
	if err != nil {
		slog.Error("failed to list models", "provider", llm.GetProvider().Name, "err", err)
	}

	if len(providerModels) == 0 {
		return nil, errors.New("no models found")
	}

	modelCache.Models = providerModels
	modelCache.Timestamp = time.Now()
	if err := llms.SaveModelCache(llm.GetProvider().Name, modelCache); err != nil {
		slog.Error("failed to save model cache", "err", err)
	}
	return modelCache.Models, nil
}
