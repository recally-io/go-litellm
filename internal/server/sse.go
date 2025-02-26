package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/recally-io/polyllm/llms"
)

func handleStreamingResponse(w http.ResponseWriter, ctx context.Context, llm llms.LLM, req llms.ChatCompletionRequest) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	streamingFunc := func(content llms.StreamingChatCompletionResponse) {
		if content.Err != nil {
			if content.Err == io.EOF {
				// Send the final [DONE] message
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				return
			}
			// Handle error
			errMsg := fmt.Sprintf("data: {\"error\":{\"message\":\"%s\"}}\n\n", content.Err.Error())
			fmt.Fprint(w, errMsg)
			flusher.Flush()
			return
		}

		if content.Response != nil {
			// Format the response as SSE
			jsonData, err := json.Marshal(content.Response)
			if err != nil {
				fmt.Fprintf(w, "data: {\"error\":{\"message\":\"Failed to marshal response\"}}\n\n")
				flusher.Flush()
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()
		}
	}

	llm.ChatCompletion(ctx, req, streamingFunc)
}

func handleNonStreamingResponse(w http.ResponseWriter, ctx context.Context, llm llms.LLM, req llms.ChatCompletionRequest) {
	w.Header().Set("Content-Type", "application/json")

	var fullResponse *llms.ChatCompletionResponse

	streamingFunc := func(content llms.StreamingChatCompletionResponse) {
		if content.Err != nil {
			if content.Err != io.EOF {
				// Only handle non-EOF errors here
				http.Error(w, fmt.Sprintf("Error: %v", content.Err), http.StatusInternalServerError)
			}
			return
		}

		// For non-streaming, we collect the full response
		if content.Response != nil {
			fullResponse = content.Response
		}
	}

	// Execute the chat completion
	llm.ChatCompletion(ctx, req, streamingFunc)

	// After completion, return the full response
	if fullResponse != nil {
		if err := json.NewEncoder(w).Encode(fullResponse); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "No response generated", http.StatusInternalServerError)
	}
}
