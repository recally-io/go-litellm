package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/recally-io/polyllm/llms"
	"github.com/recally-io/polyllm/logger"
)

type LLMService struct {
	provider LLMProvider
}

type LLMProvider interface {
	ListModels(ctx context.Context) ([]llms.Model, error)
	ChatCompletion(ctx context.Context, req llms.ChatCompletionRequest, streamingFunc func(resp llms.StreamingChatCompletionResponse), options ...llms.RequestOption)
}

func NewLLMService(provider LLMProvider) *LLMService {
	return &LLMService{
		provider: provider,
	}
}

func (s *LLMService) ListModels() {
	ctx := context.Background()
	models, err := s.provider.ListModels(ctx)
	if err != nil {
		fmt.Printf("Failed to list models: %v\n", err)
		return
	}

	fmt.Println("Available models:")
	for _, model := range models {
		fmt.Printf(" %s - %s\n", model.Name, model.ID)
	}
}

func (s *LLMService) ChatCompletion(modelName, prompt string) {
	slog.Debug(fmt.Sprintf("%sChatting with model: %s%s\n", logger.ColorBlue, modelName, logger.ColorReset))
	slog.Debug(fmt.Sprintf("%sPrompt: %s%s\n\n", logger.ColorGreen, prompt, logger.ColorReset))

	// Create a context
	ctx := context.Background()

	// Create a request
	req := llms.ChatCompletionRequest{
		Model: modelName,
		Messages: []llms.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}

	// Stream the response
	s.provider.ChatCompletion(ctx, req, func(resp llms.StreamingChatCompletionResponse) {
		if resp.Err != nil && resp.Err != io.EOF {
			slog.Error("Error streaming response", "err", resp.Err)
			return
		}

		if resp.Response != nil && len(resp.Response.Choices) > 0 {
			if resp.Response.Choices[0].Delta != nil {
				fmt.Printf("%s%s%s", logger.ColorCyan, resp.Response.Choices[0].Delta.Content, logger.ColorReset)
			}
		}
	})

	fmt.Println() // Add a newline at the end
}
