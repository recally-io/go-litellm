package models

import (
	"context"
	"fmt"
	"time"

	"github.com/recally-io/polyllm"
	"github.com/recally-io/polyllm/examples/polyllm-cli/pkg/cache"
	"github.com/recally-io/polyllm/examples/polyllm-cli/pkg/providers"
	"github.com/recally-io/polyllm/llms"
)

// UpdateModelsInBackground updates the models cache in the background
func UpdateModelsInBackground() {
	go func() {
		// Create a context
		ctx := context.Background()
		
		// Get all providers
		providerList := providers.GetAllProviders()
		
		// Load existing cache
		modelCache, err := cache.Load()
		if err != nil {
			fmt.Printf("Error loading model cache: %v\n", err)
			modelCache = cache.ModelCache{
				Models:    make(map[string][]llms.Model),
				Timestamp: time.Now(),
			}
		}
		
		// For each provider, try to list models
		for _, provider := range providerList {
			// Create a client for the provider
			client, err := polyllm.New(provider.ProviderName)
			if err != nil {
				continue
			}
			
			// List models
			models, err := client.ListModels(ctx)
			if err != nil {
				continue
			}
			
			// Update cache
			modelCache.Models[provider.Name] = models
		}
		
		// Update timestamp
		modelCache.Timestamp = time.Now()
		
		// Save cache
		if err := cache.Save(modelCache); err != nil {
			fmt.Printf("Error saving model cache: %v\n", err)
		}
	}()
}

// ListModels lists all available models
func ListModels() {
	fmt.Println("Listing available models...")
	
	// Try to load models from cache
	modelCache, err := cache.Load()
	if err != nil {
		fmt.Printf("Error loading model cache: %v\n", err)
	}
	
	// Check if cache is valid
	cacheValid := cache.IsValid(modelCache)
	
	// If cache is valid, use it
	if cacheValid && len(modelCache.Models) > 0 {
		fmt.Printf("Using cached models (last updated: %s)\n", modelCache.Timestamp.Format(time.RFC1123))
		
		// Print models from cache
		for provider, models := range modelCache.Models {
			fmt.Printf("\n%s Models:\n", provider)
			
			if len(models) == 0 {
				fmt.Println("  No models available")
			} else {
				for _, model := range models {
					fmt.Printf("  - %s\n", model.ID)
				}
			}
		}
		
		// Update models in background
		fmt.Println("\nUpdating models in background...")
		UpdateModelsInBackground()
		return
	}
	
	// If cache is not valid, fetch models directly
	ctx := context.Background()
	
	// Get all providers
	providerList := providers.GetAllProviders()
	
	// Create a new cache
	newCache := cache.ModelCache{
		Models:    make(map[string][]llms.Model),
		Timestamp: time.Now(),
	}
	
	// For each provider, try to list models
	for _, provider := range providerList {
		fmt.Printf("\n%s Models:\n", provider.Name)
		
		// Create a client for the provider
		client, err := polyllm.New(provider.ProviderName)
		if err != nil {
			fmt.Printf("  Error initializing %s client: %v\n", provider.Name, err)
			continue
		}
		
		// List models
		models, err := client.ListModels(ctx)
		if err != nil {
			fmt.Printf("  Error listing models: %v\n", err)
			continue
		}
		
		// Update cache
		newCache.Models[provider.Name] = models
		
		// Print models
		if len(models) == 0 {
			fmt.Println("  No models available")
		} else {
			for _, model := range models {
				fmt.Printf("  - %s\n", model.ID)
			}
		}
	}
	
	// Save the new cache
	if err := cache.Save(newCache); err != nil {
		fmt.Printf("Error saving model cache: %v\n", err)
	}
}

// ChatWithModel chats with the specified model
func ChatWithModel(modelName, prompt string) {
	fmt.Printf("Chatting with model: %s\n", modelName)
	fmt.Printf("Prompt: %s\n\n", prompt)
	
	// Create a context
	ctx := context.Background()
	
	// Determine provider from model name prefix
	providerName, modelID := providers.DetermineProvider(modelName)
	
	// Create a client for the provider
	client, err := polyllm.New(providerName)
	if err != nil {
		fmt.Printf("Error initializing client: %v\n", err)
		return
	}
	
	// Create a request
	req := llms.ChatCompletionRequest{
		Model: modelID,
		Messages: []llms.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}
	
	// Stream the response
	client.ChatCompletion(ctx, req, func(resp llms.StreamingChatCompletionResponse) {
		if resp.Err != nil {
			fmt.Printf("Error: %v\n", resp.Err)
			return
		}
		
		if resp.Response != nil && len(resp.Response.Choices) > 0 {
			if resp.Response.Choices[0].Delta != nil {
				fmt.Print(resp.Response.Choices[0].Delta.Content)
			}
		}
	})
	
	fmt.Println() // Add a newline at the end
}
