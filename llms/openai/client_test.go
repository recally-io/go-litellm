package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/recally-io/go-litellm/llms"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	client := New()
	assert.Equal(t, baseURL, client.BaseURL)

	customURL := "https://custom.openai.com"
	client = New(llms.WithBaseURL(customURL))
	assert.Equal(t, customURL, client.BaseURL)
}

func TestListModels(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []byte
		prefix         string
		expectedModels []llms.Model
		expectError    bool
	}{
		{
			name: "successful response",
			serverResponse: []byte(`{
				"object": "list",
				"data": [
					{"id": "gpt-4", "object": "model"},
					{"id": "gpt-3.5-turbo", "object": "model"}
				]
			}`),
			expectedModels: []llms.Model{
				{ID: "gpt-4", Object: "model"},
				{ID: "gpt-3.5-turbo", Object: "model"},
			},
		},
		{
			name:   "with prefix",
			prefix: "openai",
			serverResponse: []byte(`{
				"object": "list",
				"data": [
					{"id": "gpt-4", "object": "model"}
				]
			}`),
			expectedModels: []llms.Model{
				{ID: "openai/gpt-4", Object: "model"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/models", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				w.Write(tt.serverResponse)
			}))
			defer server.Close()

			client := New(
				llms.WithBaseURL(server.URL),
				llms.WithPrefix(tt.prefix),
			)

			models, err := client.ListModels(context.Background())
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedModels, models)
		})
	}
}

func TestChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		request        llms.ChatCompletionRequest
		serverResponse []byte
		expectError    bool
		expectedResp   llms.StreamingChatCompletionResponse
	}{
		{
			name: "successful non-streaming response",
			request: llms.ChatCompletionRequest{
				Model: "gpt-4",
				Messages: []llms.ChatCompletionMessage{
					{Role: llms.ChatMessageRoleUser, Content: "Hello"},
				},
				Stream: false,
			},
			serverResponse: []byte(`{
				"id": "chatcmpl-123",
				"object": "chat.completion",
				"choices": [
					{
						"index": 0,
						"message": {
							"role": "assistant",
							"content": "Hello! How can I help you today?"
						},
						"finish_reason": "stop"
					}
				]
			}`),
			expectedResp: llms.StreamingChatCompletionResponse{
				Response: &llms.ChatCompletionResponse{
					ID: "chatcmpl-123",
					Choices: []llms.ChatCompletionChoice{
						{
							Index: 0,
							Message: &llms.ChatCompletionMessage{
								Role:    llms.ChatMessageRoleAssistant,
								Content: "Hello! How can I help you today?",
							},
							FinishReason: llms.FinishReasonStop,
						},
					},
					Object: "chat.completion",
				},
				Err: io.EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat/completions", r.URL.Path)

				var reqBody llms.ChatCompletionRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.request, reqBody)

				w.WriteHeader(http.StatusOK)
				w.Write(tt.serverResponse)
			}))
			defer server.Close()

			client := New(llms.WithBaseURL(server.URL))

			var response llms.StreamingChatCompletionResponse
			client.ChatCompletion(context.Background(), tt.request, func(resp llms.StreamingChatCompletionResponse) {
				response = resp
			})

			if tt.expectError {
				assert.Error(t, response.Err)
				assert.NotEqual(t, io.EOF, response.Err)
				return
			}

			assert.Equal(t, tt.expectedResp, response)
		})
	}
}

func TestGenerateText(t *testing.T) {
	tests := []struct {
		name           string
		model          string
		prompt         string
		serverResponse []byte
		expectedText   string
		expectError    bool
	}{
		{
			name:   "successful response",
			model:  "gpt-4",
			prompt: "Hello",
			serverResponse: []byte(`{
				"id": "chatcmpl-123",
				"object": "chat.completion",
				"choices": [
					{
						"index": 0,
						"message": {
							"role": "assistant",
							"content": "Hello! How can I help you today?"
						},
						"finish_reason": "stop"
					}
				]
			}`),
			expectedText: "Hello! How can I help you today?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat/completions", r.URL.Path)

				var reqBody llms.ChatCompletionRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, llms.ChatCompletionRequest{
					Model:  tt.model,
					Stream: false,
					Messages: []llms.ChatCompletionMessage{
						{Role: "user", Content: tt.prompt},
					},
				}, reqBody)

				w.WriteHeader(http.StatusOK)
				w.Write(tt.serverResponse)
			}))
			defer server.Close()

			client := New(llms.WithBaseURL(server.URL))

			text, err := client.GenerateText(context.Background(), tt.model, tt.prompt)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedText, text)
		})
	}
}
