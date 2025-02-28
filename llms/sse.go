package llms

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// StreamingSSEResponse handles streaming responses from OpenAI's API.
// It reads the response body line by line, processes each chunk of data,
// and calls the provided streaming function with the processed content.
// respBody: The response body from the HTTP request
// streamingFunc: The callback function to handle each chunk of streaming data
func StreamingSSEResponse(respBody io.ReadCloser, streamingFunc func(content StreamingChatCompletionResponse)) {
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
			streamingFunc(StreamingChatCompletionResponse{
				Response: nil,
				Err:      io.EOF,
			})
			break
		}

		var chunk ChatCompletionResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			streamingFunc(StreamingChatCompletionResponse{
				Response: nil,
				Err:      fmt.Errorf("error unmarshaling response: %v", err),
			})
			return
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		streamingFunc(StreamingChatCompletionResponse{
			Response: &chunk,
			Err:      nil,
		})
	}

	if err := scanner.Err(); err != nil {
		streamingFunc(StreamingChatCompletionResponse{
			Response: nil,
			Err:      fmt.Errorf("error reading response: %v", err),
		})
	}
}
