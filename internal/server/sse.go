package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/recally-io/polyllm/llms"
)

func handleStreamingResponse(w http.ResponseWriter, ctx context.Context, llm LLMProvider, req llms.ChatCompletionRequest) {
	slog.Info("Starting streaming response handler", "model", req.Model)
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		slog.Error("Streaming not supported by the response writer")
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	streamingFunc := func(content llms.StreamingChatCompletionResponse) {
		if content.Err != nil {
			if content.Err == io.EOF {
				// Send the final [DONE] message
				slog.Debug("Streaming complete, sending [DONE] message")
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				return
			}
			// Handle error
			slog.Error("Error streaming response", "err", content.Err)
			errMsg := fmt.Sprintf("data: {\"error\":{\"message\":\"%s\"}}\n\n", content.Err.Error())
			fmt.Fprint(w, errMsg)
			flusher.Flush()
			return
		}

		if content.Response != nil {
			// Format the response as SSE
			jsonData, err := json.Marshal(content.Response)
			if err != nil {
				slog.Error("Error marshalling response", "err", err)
				fmt.Fprintf(w, "data: {\"error\":{\"message\":\"Failed to marshal response\"}}\n\n")
				flusher.Flush()
				return
			}
			slog.Debug("Sending chunk of streaming response",
				"chunk_size", len(jsonData),
				"finish_reason", content.Response.Choices[0].FinishReason)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()
		}
	}

	slog.Debug("Initiating chat completion with streaming")
	llm.ChatCompletion(ctx, req, streamingFunc)
	slog.Debug("Completed streaming response handling")
}

func handleNonStreamingResponse(w http.ResponseWriter, ctx context.Context, llm LLMProvider, req llms.ChatCompletionRequest) {
	slog.Info("Starting non-streaming response handler", "model", req.Model)
	w.Header().Set("Content-Type", "application/json")

	var fullResponse *llms.ChatCompletionResponse

	streamingFunc := func(content llms.StreamingChatCompletionResponse) {
		if content.Err != nil {
			if content.Err != io.EOF {
				// Only handle non-EOF errors here
				slog.Error("Error during non-streaming response generation", "err", content.Err)
				http.Error(w, fmt.Sprintf("Error: %v", content.Err), http.StatusInternalServerError)
			}
			return
		}

		// For non-streaming, we collect the full response
		if content.Response != nil {
			slog.Debug("Collected part of non-streaming response")
			fullResponse = content.Response
		}
	}

	// Execute the chat completion
	slog.Debug("Initiating chat completion without streaming")
	llm.ChatCompletion(ctx, req, streamingFunc)

	// After completion, return the full response
	if fullResponse != nil {
		slog.Debug("Encoding and sending complete non-streaming response")
		if err := json.NewEncoder(w).Encode(fullResponse); err != nil {
			slog.Error("Failed to encode non-streaming response", "err", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	} else {
		slog.Error("No response generated in non-streaming mode")
		http.Error(w, "No response generated", http.StatusInternalServerError)
	}
	slog.Info("Completed non-streaming response handling")
}
