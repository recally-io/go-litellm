# PolyLLM

PolyLLM is a unified Go interface for multiple Large Language Model (LLM) providers. It allows you to interact with various LLM APIs through a single, consistent interface, making it easy to switch between different providers or use multiple providers in the same application.

## Features

- **Single Interface**: Interact with multiple LLM providers through a unified API
- **Provider Agnostic**: Easily switch between providers without changing your code
- **Streaming Support**: Full support for streaming responses from supported LLM providers
- **Extensible**: Simple to add support for new providers
- **Multiple Interfaces**: Access LLMs through a Go API, CLI tool, or HTTP server

## Supported Providers

PolyLLM currently supports the following LLM providers:

- OpenAI
- DeepSeek
- Qwen (Alibaba Cloud)
- Gemini (Google)
- OpenRouter
- Volcengine
- Groq
- Xai
- Siliconflow

Additional providers can be easily added.

## Installation

```bash
go get github.com/recally-io/polyllm
```

## Usage

### API Usage

### Basic Example

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/recally-io/polyllm"
)

func main() {
	// Create a new PolyLLM instance
	llm := polyllm.New()

	// Generate text using the default provider (first available)
	response, err := llm.GenerateText(
		context.Background(),
		"openai/gpt-3.5-turbo",
		"Explain quantum computing in simple terms",
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(response)
}
```

### Chat Completion Example

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/recally-io/polyllm"
	"github.com/recally-io/polyllm/llms"
)

func main() {
	// Create a new PolyLLM instance
	llm := polyllm.New()

	// Create a chat completion request
	req := llms.ChatCompletionRequest{
		Model: "openai/gpt-4",
		Messages: []llms.Message{
			{
				Role:    llms.RoleSystem,
				Content: "You are a helpful assistant.",
			},
			{
				Role:    llms.RoleUser,
				Content: "What are the key features of Go programming language?",
			},
		},
	}

	// Stream the response
	llm.ChatCompletion(context.Background(), req, func(resp llms.StreamingChatCompletionResponse) {
		if resp.Err != nil {
			fmt.Printf("Error: %v\n", resp.Err)
			os.Exit(1)
		}
		
		if len(resp.Choices) > 0 && resp.Choices[0].Delta.Content != "" {
			fmt.Print(resp.Choices[0].Delta.Content)
		}
	})
}
```

### CLI Usage

PolyLLM comes with a command-line interface that allows you to interact with various LLM providers directly from your terminal.

#### Installation

```bash
# Install the CLI
go install github.com/recally-io/polyllm/cmd/polyllm-cli@latest
```

#### Examples

List all available models:
```bash
polyllm-cli models
```

Chat with a specific model:
```bash
polyllm-cli -m "openai/gpt-3.5-turbo" "Tell me a joke about programming"
```

```bash
polyllm-cli -m "deepseek/deepseek-chat" "What is the meaning of life?"
```

### HTTP Server

PolyLLM also provides an HTTP server that exposes OpenAI-compatible endpoints, making it easy to use with existing tools and libraries that support the OpenAI API.

#### Installation

```bash
# Install the server
go install github.com/recally-io/polyllm/cmd/polyllm-server@latest
```

#### Starting the Server

```bash
# Start the server on default port 8088
polyllm-server

# Or specify a custom port
PORT=3000 polyllm-server

# Add API key authentication
API_KEY=your_api_key polyllm-server
```

#### API Endpoints

The server provides OpenAI-compatible endpoints:

- `GET /models` or `GET /v1/models` - List all available models
- `POST /chat/completions` or `POST /v1/chat/completions` - Create a chat completion

#### Example Request

```bash
# Request to list models
curl http://localhost:8088/models

# Request with authentication (if enabled)
curl -H "Authorization: Bearer your_api_key" http://localhost:8088/models

# Chat completion request
curl -X POST http://localhost:8088/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'
```

### Configuration

PolyLLM uses environment variables for configuration. Each provider has its own prefix:

```
OPENAI_API_KEY=your_api_key
DEEPSEEK_API_KEY=your_api_key
QWEN_API_KEY=your_api_key
GEMINI_API_KEY=your_api_key
OPENROUTER_API_KEY=your_api_key
VOLCENGINE_API_KEY=your_api_key
GROQ_API_KEY=your_api_key
```

You can also add providers programmatically:

```go
llm := polyllm.New()
llm.AddProvider(&providers.Provider{
	Name:       "custom-provider",
	Type:       providers.ProviderTypeOpenAICompatible,
	BaseURL:    "https://api.custom-provider.com/v1",
	APIKey:     "your-api-key",
	ModelPrefix: "custom/",
})
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the terms provided in the [LICENSE](LICENSE) file.
