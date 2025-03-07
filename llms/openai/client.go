package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/recally-io/polyllm/llms"
)

const baseURL = "https://api.openai.com/v1"

// Client is the client for interacting with OpenAI's API.
// It implements various LLM operations using the OpenAI API.
type Client struct {
	*llms.Provider
}

// New creates a new OpenAI client with the provided configuration options.
// opts: Configuration options for the client
func New(apiKey string, opts ...llms.Option) (*Client, error) {
	provider := &llms.Provider{
		Type:    llms.ProviderTypeOpenAI,
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
	for _, opt := range opts {
		opt(provider)
	}

	if provider.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	return &Client{Provider: provider}, nil
}

func (c *Client) GetProvider() *llms.Provider {
	return c.Provider
}

// ListModels retrieves the list of available models from OpenAI.
// ctx: Context for the request
// Returns: List of available models or error if the request fails
func (c *Client) ListModels(ctx context.Context) ([]llms.Model, error) {
	// get models from provider config
	models := c.Provider.GetModelList(ctx)
	if len(models) > 0 {
		return models, nil
	}

	// get models from API
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	c.SetHttpHeaders(req, false, nil)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		message, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unexpected status code: %d: %s", res.StatusCode, message)
	}

	// Define a struct to match the OpenAI API response format
	var response struct {
		Data   []llms.Model `json:"data"`
		Object string       `json:"object"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	models = response.Data
	for i := range models {
		if c.ModelPrefix != "" {
			models[i].ID = c.ModelPrefix + models[i].ID
		}
		if models[i].Name == "" {
			models[i].Name = models[i].ID
		}
	}
	return models, nil
}

// ChatCompletion performs a chat completion request to OpenAI.
// ctx: Context for the request
// req: The chat completion request parameters
// streamingFunc: Callback function for handling streaming responses
// options: Additional request options
func (c *Client) ChatCompletion(ctx context.Context, req llms.ChatCompletionRequest, streamingFunc func(content llms.StreamingChatCompletionResponse), options ...llms.RequestOption) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", nil)
	if err != nil {
		streamingFunc(llms.StreamingChatCompletionResponse{Err: fmt.Errorf("failed to create request: %w", err)})
		return
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		streamingFunc(llms.StreamingChatCompletionResponse{Err: fmt.Errorf("failed to marshal request: %w", err)})
		return
	}
	c.SetHttpHeaders(httpReq, req.Stream, req.ExtraHeaders)
	httpReq.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		streamingFunc(llms.StreamingChatCompletionResponse{Err: fmt.Errorf("failed to send request: %w", err)})
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		message, err := io.ReadAll(resp.Body)
		if err != nil {
			streamingFunc(llms.StreamingChatCompletionResponse{Err: fmt.Errorf("failed to read response: %w", err)})
			return
		}
		streamingFunc(llms.StreamingChatCompletionResponse{Err: fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, message)})
		return
	}

	// Process Non-streaming request
	if !req.Stream {
		var response llms.ChatCompletionResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			streamingFunc(llms.StreamingChatCompletionResponse{Err: fmt.Errorf("failed to decode response: %w", err)})
		} else {
			streamingFunc(llms.StreamingChatCompletionResponse{Response: &response, Err: io.EOF})
		}
		return
	}

	// Process the streaming response
	llms.StreamingSSEResponse(resp.Body, streamingFunc)
}

// GenerateText generates text using the specified model and prompt.
// ctx: Context for the request
// model: The model to use for text generation
// prompt: The input prompt for text generation
// options: Additional request options
// Returns: Generated text or error if the request fails
func (c *Client) GenerateText(ctx context.Context, model, prompt string, options ...llms.RequestOption) (string, error) {
	req := llms.ChatCompletionRequest{
		Model:  model,
		Stream: false,
		Messages: []llms.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}
	resp := ""
	var err error

	streamingFunc := func(content llms.StreamingChatCompletionResponse) {
		if content.Err != nil && content.Err != io.EOF {
			err = fmt.Errorf("generate text error: %w", content.Err)
			return
		}
		if content.Response != nil && len(content.Response.Choices) > 0 && content.Response.Choices[0].Message != nil {
			resp = content.Response.Choices[0].Message.Content
		}
	}

	c.ChatCompletion(ctx, req, streamingFunc, options...)
	return resp, err
}

// StreamGenerateText generates text in a streaming fashion.
// ctx: Context for the request
// model: The model to use for text generation
// prompt: The input prompt for text generation
// streamingTextFunc: Callback function for handling streaming text chunks
// options: Additional request options
func (c *Client) StreamGenerateText(ctx context.Context, model, prompt string, streamingTextFunc func(resp llms.StreamingChatCompletionText), options ...llms.RequestOption) {
	req := llms.ChatCompletionRequest{
		Model:  model,
		Stream: true,
		Messages: []llms.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	streamingFunc := func(content llms.StreamingChatCompletionResponse) {
		if content.Err != nil {
			streamingTextFunc(llms.StreamingChatCompletionText{Err: content.Err})
			return
		}
		streamingTextFunc(llms.StreamingChatCompletionText{Content: content.Response.Choices[0].Delta.Content})
	}

	c.ChatCompletion(ctx, req, streamingFunc, options...)
}
