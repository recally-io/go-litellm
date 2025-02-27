package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/recally-io/polyllm/llms"
)

type LLMService struct {
	llm llms.LLM
}

func NewLLMService(llm llms.LLM) *LLMService {
	return &LLMService{
		llm: llm,
	}
}

func (s *LLMService) ListModels() {
	ctx := context.Background()
	models, err := s.llm.ListModels(ctx)
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
	slog.Debug(fmt.Sprintf("Chatting with model: %s\n", modelName))
	slog.Debug(fmt.Sprintf("Prompt: %s\n\n", prompt))

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
	s.llm.ChatCompletion(ctx, req, func(resp llms.StreamingChatCompletionResponse) {
		if resp.Err != nil && resp.Err != io.EOF {
			slog.Error("Error streaming response", "err", resp.Err)
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
