package openai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/recally-io/polyllm/llms"
)

// streamResponse handles streaming responses from OpenAI's API.
// It reads the response body line by line, processes each chunk of data,
// and calls the provided streaming function with the processed content.
// respBody: The response body from the HTTP request
// streamingFunc: The callback function to handle each chunk of streaming data
func streamResponse(respBody io.ReadCloser, streamingFunc func(content llms.StreamingChatCompletionResponse)) {
	scanner := bufio.NewScanner(respBody)
	defer respBody.Close()

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			streamingFunc(llms.StreamingChatCompletionResponse{
				Response: nil,
				Err:      io.EOF,
			})
			break
		}

		var chunk llms.ChatCompletionResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			streamingFunc(llms.StreamingChatCompletionResponse{
				Response: nil,
				Err:      fmt.Errorf("error unmarshaling response: %v", err),
			})
			return
		}

		if len(chunk.Choices) == 0 || chunk.Choices[0].Delta.Content == "" {
			continue
		}

		streamingFunc(llms.StreamingChatCompletionResponse{
			Response: &chunk,
			Err:      nil,
		})
	}

	if err := scanner.Err(); err != nil {
		streamingFunc(llms.StreamingChatCompletionResponse{
			Response: nil,
			Err:      fmt.Errorf("error reading response: %v", err),
		})
	}
}
