package openai

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/recally-io/go-litellm/llms"
)

// testHelper collects streaming responses
func testHelper(resp llms.StreamingChatCompletionResponse, responses *[]llms.StreamingChatCompletionResponse) {
	*responses = append(*responses, resp)
}

// TestStreamResponse_Normal tests the normal flow of the streamResponse function with valid JSON responses and a [DONE] marker.
func TestStreamResponse_Normal(t *testing.T) {
	// Prepare a mock response stream
	input := "data: {\"Choices\":[{\"Delta\":{\"Content\":\"Hello\"}}]}\n" +
		"data: {\"Choices\":[{\"Delta\":{\"Content\":\" World\"}}]}\n" +
		"data: [DONE]\n"
	reader := io.NopCloser(strings.NewReader(input))

	var responses []llms.StreamingChatCompletionResponse
	streamResponse(reader, func(resp llms.StreamingChatCompletionResponse) {
		testHelper(resp, &responses)
	})

	if len(responses) != 3 {
		t.Errorf("expected 3 responses, got %d", len(responses))
	}

	// Check first response
	if responses[0].Err != nil {
		t.Errorf("unexpected error in first response: %v", responses[0].Err)
	}
	if responses[0].Response == nil || len(responses[0].Response.Choices) == 0 || responses[0].Response.Choices[0].Delta.Content != "Hello" {
		t.Errorf("first response content mismatch")
	}

	// Check second response
	if responses[1].Err != nil {
		t.Errorf("unexpected error in second response: %v", responses[1].Err)
	}
	if responses[1].Response == nil || len(responses[1].Response.Choices) == 0 || responses[1].Response.Choices[0].Delta.Content != " World" {
		t.Errorf("second response content mismatch")
	}

	// Check third response which should signal end of stream with io.EOF
	if responses[2].Err != io.EOF {
		t.Errorf("expected io.EOF in third response, got %v", responses[2].Err)
	}
}

// TestStreamResponse_InvalidJSON tests the case when JSON unmarshaling fails.
func TestStreamResponse_InvalidJSON(t *testing.T) {
	// Prepare a mock response stream with invalid JSON
	input := "data: {invalid json}\n"
	reader := io.NopCloser(strings.NewReader(input))

	var responses []llms.StreamingChatCompletionResponse
	streamResponse(reader, func(resp llms.StreamingChatCompletionResponse) {
		testHelper(resp, &responses)
	})

	if len(responses) != 1 {
		t.Errorf("expected 1 response, got %d", len(responses))
	}

	if responses[0].Err == nil || !strings.Contains(responses[0].Err.Error(), "error unmarshaling response") {
		t.Errorf("expected unmarshaling error, got %v", responses[0].Err)
	}
}

// errorReader is an io.ReadCloser that returns an error when read is called
// It is used to simulate a scanner error

type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func (e *errorReader) Close() error {
	return nil
}

// TestStreamResponse_ScannerError tests the behavior when the scanner encounters a read error.
func TestStreamResponse_ScannerError(t *testing.T) {
	reader := &errorReader{}
	var responses []llms.StreamingChatCompletionResponse
	streamResponse(reader, func(resp llms.StreamingChatCompletionResponse) {
		testHelper(resp, &responses)
	})

	if len(responses) != 1 {
		t.Errorf("expected 1 response due to scanner error, got %d", len(responses))
	}

	if responses[0].Err == nil || !strings.Contains(responses[0].Err.Error(), "error reading response") {
		t.Errorf("expected scanner error, got %v", responses[0].Err)
	}
}

// TestStreamResponse_EmptyData tests that no responses are sent when input doesn't contain valid data lines.
func TestStreamResponse_EmptyData(t *testing.T) {
	input := "Not data line\n\nAnother line without prefix\n"
	reader := io.NopCloser(strings.NewReader(input))

	var responses []llms.StreamingChatCompletionResponse
	streamResponse(reader, func(resp llms.StreamingChatCompletionResponse) {
		responses = append(responses, resp)
	})

	if len(responses) != 0 {
		t.Errorf("expected no responses, got %d", len(responses))
	}
}

// TestStreamResponse_NoChoices tests that no response is sent when the JSON has an empty choices array.
func TestStreamResponse_NoChoices(t *testing.T) {
	input := "data: {\"Choices\": []}\n" +
		"data: [DONE]\n"
	reader := io.NopCloser(strings.NewReader(input))

	var responses []llms.StreamingChatCompletionResponse
	streamResponse(reader, func(resp llms.StreamingChatCompletionResponse) {
		responses = append(responses, resp)
	})

	if len(responses) != 1 {
		t.Errorf("expected 1 response (EOF), got %d", len(responses))
	}

	if responses[0].Err != io.EOF {
		t.Errorf("expected io.EOF error, got %v", responses[0].Err)
	}
}

// TestStreamResponse_NoContent tests that no response is sent for valid JSON data when the choices contain an empty content string.
func TestStreamResponse_NoContent(t *testing.T) {
	input := "data: {\"Choices\":[{\"Delta\":{\"Content\":\"\"}}]}\n" +
		"data: [DONE]\n"
	reader := io.NopCloser(strings.NewReader(input))

	var responses []llms.StreamingChatCompletionResponse
	streamResponse(reader, func(resp llms.StreamingChatCompletionResponse) {
		responses = append(responses, resp)
	})

	if len(responses) != 1 {
		t.Errorf("expected 1 response (EOF), got %d", len(responses))
	}

	if responses[0].Err != io.EOF {
		t.Errorf("expected io.EOF error, got %v", responses[0].Err)
	}
}
