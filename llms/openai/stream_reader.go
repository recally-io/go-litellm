package openai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/recally-io/go-litellm/llms"
)

func streamResponse(respBody io.ReadCloser, streamingFunc func(content llms.StreamingChatCompletionResponse)) {
	scanner := bufio.NewScanner(respBody)
	defer respBody.Close()

	// Create a response object that will be updated with each chunk
	chunk := &llms.ChatCompletionResponse{}

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check if the line starts with "data: "
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract the data part
		data := strings.TrimPrefix(line, "data: ")

		// Check if it's the end of the stream
		if data == "[DONE]" {
			// Signal the end of the stream with io.EOF
			streamingFunc(llms.StreamingChatCompletionResponse{
				Response: chunk,
				Err:      io.EOF,
			})
			break
		}

		// Parse the JSON data
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

		// Send the updated response through the streaming function
		streamingFunc(llms.StreamingChatCompletionResponse{
			Response: chunk,
			Err:      nil,
		})
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		streamingFunc(llms.StreamingChatCompletionResponse{
			Response: nil,
			Err:      fmt.Errorf("error reading response: %v", err),
		})
	}
}
