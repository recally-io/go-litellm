package server

import (
	"encoding/json"
	"io"
	"net/http"

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

type listModelsResponse struct {
	Data   []llms.Model `json:"data"`
	Object string       `json:"object"`
}

func (s *LLMService) listModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	models, err := s.llm.ListModels(ctx)
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
		handleStreamingResponse(w, ctx, s.llm, req)
	} else {
		handleNonStreamingResponse(w, ctx, s.llm, req)
	}
}
