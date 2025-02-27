package llms

import (
	"context"
)

type LLM interface {
	GetProviderName() string
	ListModels(ctx context.Context) ([]Model, error)

	ChatCompletion(ctx context.Context, req ChatCompletionRequest, streamingFunc func(resp StreamingChatCompletionResponse), options ...RequestOption)
	GenerateText(ctx context.Context, model, prompt string, options ...RequestOption) (string, error)
	StreamGenerateText(ctx context.Context, model, prompt string, streamingFunc func(resp StreamingChatCompletionText), options ...RequestOption)
}

type StreamingChatCompletionText struct {
	Content string
	Err     error
}
