package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/recally-io/polyllm/llms"
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

type listModelsResponse struct {
	Data   []llms.Model `json:"data"`
	Object string       `json:"object"`
}

func (s *LLMService) listModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	models, err := s.provider.ListModels(ctx)
	if err != nil {
		http.Error(w, "Failed to list models", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(listModelsResponse{
		Data:   models,
		Object: "list",
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *LLMService) chatCompletion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var req llms.ChatCompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	if req.Stream {
		handleStreamingResponse(w, ctx, s.provider, req)
	} else {
		handleNonStreamingResponse(w, ctx, s.provider, req) // updated to use s.llmService
	}
}
