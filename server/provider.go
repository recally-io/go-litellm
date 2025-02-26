package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/recally-io/polyllm"
	"github.com/recally-io/polyllm/llms"
)

var modelProviderMappings = make(map[string]*Provider)

func initProviders() error {
	if Settings.OpenAI.APIKey != "" {
		if err := initProvider(polyllm.ProviderNameOpenAI, Settings.OpenAI); err != nil {
			return err
		}
	}

	if Settings.DeepSeek.APIKey != "" {
		if err := initProvider(polyllm.ProviderNameDeepSeek, Settings.DeepSeek); err != nil {
			return err
		}
	}

	if Settings.Qwen.APIKey != "" {
		if err := initProvider(polyllm.ProviderNameQwen, Settings.Qwen); err != nil {
			return err
		}
	}

	if Settings.Gemini.APIKey != "" {
		if err := initProvider(polyllm.ProviderNameGemini, Settings.Gemini); err != nil {
			return err
		}
	}

	if Settings.OpenRouter.APIKey != "" {
		if err := initProvider(polyllm.ProviderNameOpenRouter, Settings.OpenRouter); err != nil {
			return err
		}
	}

	if Settings.Volcengine.APIKey != "" {
		if err := initProvider(polyllm.ProviderNameVolcengine, Settings.Volcengine); err != nil {
			return err
		}
	}

	if Settings.Groq.APIKey != "" {
		if err := initProvider(polyllm.ProviderNameGroq, Settings.Groq); err != nil {
			return err
		}
	}

	if Settings.Xai.APIKey != "" {
		if err := initProvider(polyllm.ProviderNameXai, Settings.Xai); err != nil {
			return err
		}
	}
	slog.Info("initialized providers", "models", modelProviderMappings)
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

	for model := range modelSet {
		if provider.Prefix != "" {
			model = provider.Prefix + "/" + model
		}
		modelProviderMappings[model] = provider
	}

	return nil
}

func getProviderByModelName(modelName string) (*Provider, error) {
	if provider, ok := modelProviderMappings[modelName]; ok {
		return provider, nil
	}
	return nil, fmt.Errorf("model not found: %s", modelName)
}
