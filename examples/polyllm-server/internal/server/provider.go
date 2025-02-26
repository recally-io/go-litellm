package server

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/recally-io/polyllm"
	"github.com/recally-io/polyllm/llms"
)

var (
	modelProviderMappings = make(map[string]*Provider)
	providerMapMutex      = &sync.RWMutex{}
)

func initProviders() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 8) // Buffer for potential errors

	// Function to handle provider initialization in a goroutine
	initProviderAsync := func(providerName polyllm.ProviderName, provider Provider) {
		defer wg.Done()
		if err := initProvider(providerName, provider); err != nil {
			errChan <- fmt.Errorf("failed to init %s: %w", providerName, err)
		}
	}

	if Settings.OpenAI.APIKey != "" {
		wg.Add(1)
		go initProviderAsync(polyllm.ProviderNameOpenAI, Settings.OpenAI)
	}

	if Settings.DeepSeek.APIKey != "" {
		wg.Add(1)
		go initProviderAsync(polyllm.ProviderNameDeepSeek, Settings.DeepSeek)
	}

	if Settings.Qwen.APIKey != "" {
		wg.Add(1)
		go initProviderAsync(polyllm.ProviderNameQwen, Settings.Qwen)
	}

	if Settings.Gemini.APIKey != "" {
		wg.Add(1)
		go initProviderAsync(polyllm.ProviderNameGemini, Settings.Gemini)
	}

	if Settings.OpenRouter.APIKey != "" {
		wg.Add(1)
		go initProviderAsync(polyllm.ProviderNameOpenRouter, Settings.OpenRouter)
	}

	if Settings.Volcengine.APIKey != "" {
		wg.Add(1)
		go initProviderAsync(polyllm.ProviderNameVolcengine, Settings.Volcengine)
	}

	if Settings.Groq.APIKey != "" {
		wg.Add(1)
		go initProviderAsync(polyllm.ProviderNameGroq, Settings.Groq)
	}

	if Settings.Xai.APIKey != "" {
		wg.Add(1)
		go initProviderAsync(polyllm.ProviderNameXai, Settings.Xai)
	}

	if Settings.Siliconflow.APIKey != "" {
		wg.Add(1)
		go initProviderAsync(polyllm.ProviderNameSiliconflow, Settings.Siliconflow)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Check for any errors
	for err := range errChan {
		return err // Return the first error encountered
	}

	providerMapMutex.RLock()
	slog.Info("initialized providers", "models", modelProviderMappings)
	providerMapMutex.RUnlock()
	return nil
}

func initProvider(providerName polyllm.ProviderName, provider Provider) error {
	opts := make([]llms.ConfigOptions, 0)
	if provider.APIKey != "" {
		opts = append(opts, llms.WithAPIKey(provider.APIKey))
	}
	if provider.BaseURL != "" {
		opts = append(opts, llms.WithBaseURL(provider.BaseURL))
	}

	if provider.Prefix != "" {
		opts = append(opts, llms.WithPrefix(provider.Prefix))
	}

	llm, err := polyllm.New(providerName, opts...)
	if err != nil {
		return fmt.Errorf("failed to init %s client: %w", providerName, err)
	}
	provider.llm = llm

	if err := initProviderModels(&provider); err != nil {
		slog.Error("failed to initialize provider models", "err", err)
	}

	return nil
}

func initProviderModels(provider *Provider) error {
	ctx := context.Background()
	modelSet := make(map[string]struct{})

	if provider.Models != nil {
		for _, model := range provider.Models {
			modelSet[model] = struct{}{}
		}
	}

	if provider.ModelAlias != nil {
		for alias := range provider.ModelAlias {
			modelSet[alias] = struct{}{}
		}
	}

	if len(modelSet) == 0 {
		models, err := provider.llm.ListModels(ctx)
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}

		for _, model := range models {
			modelID := model.ID
			modelSet[modelID] = struct{}{}
		}
	}

	slog.Info("initialized providers", "providers", modelSet)

	// Safely write to the shared map
	providerMapMutex.Lock()
	for model := range modelSet {
		if provider.Prefix != "" {
			model = provider.Prefix + "/" + model
		}
		modelProviderMappings[model] = provider
	}
	providerMapMutex.Unlock()

	return nil
}

func getProviderByModelName(modelName string) (*Provider, error) {
	providerMapMutex.RLock()
	provider, ok := modelProviderMappings[modelName]
	providerMapMutex.RUnlock()

	if ok {
		return provider, nil
	}
	return nil, fmt.Errorf("model not found: %s", modelName)
}
