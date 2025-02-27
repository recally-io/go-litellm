# PolyLLM

[![Go Reference](https://pkg.go.dev/badge/github.com/recally-io/polyllm.svg)](https://pkg.go.dev/github.com/recally-io/polyllm)
[![Go Report Card](https://goreportcard.com/badge/github.com/recally-io/polyllm)](https://goreportcard.com/report/github.com/recally-io/polyllm)
[![License](https://img.shields.io/github/license/recally-io/polyllm)](LICENSE)

PolyLLM is a unified Go interface for multiple Large Language Model (LLM) providers. It allows you to interact with various LLM APIs through a single, consistent interface, making it easy to switch between different providers or use multiple providers in the same application.

## Table of Contents

- [PolyLLM](#polyllm)
	- [Table of Contents](#table-of-contents)
	- [Features](#features)
	- [Supported Providers](#supported-providers)
	- [Installation](#installation)
		- [Library](#library)
		- [CLI Tool](#cli-tool)
		- [HTTP Server](#http-server)
	- [Usage](#usage)
		- [API Usage](#api-usage)
			- [Basic Example](#basic-example)
			- [Chat Completion Example](#chat-completion-example)
		- [CLI Usage](#cli-usage)
			- [Installation](#installation-1)
			- [Examples](#examples)
		- [HTTP Server](#http-server-1)
			- [Installation](#installation-2)
			- [Starting the Server](#starting-the-server)
			- [API Endpoints](#api-endpoints)
			- [Example Request](#example-request)
	- [Contributing](#contributing)
	- [License](#license)

## Features

- **Single Interface**: Interact with multiple LLM providers through a unified API
- **Provider Agnostic**: Easily switch between providers without changing your code
- **Streaming Support**: Full support for streaming responses from supported LLM providers
- **Extensible**: Simple to add support for new providers
- **Multiple Interfaces**: Access LLMs through a Go API, CLI tool, or HTTP server
- **OpenAI-Compatible**: HTTP server provides OpenAI-compatible endpoints
- **Docker Support**: Ready-to-use Docker images for both CLI and server components

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

### Library

```bash
go get github.com/recally-io/polyllm
```

### CLI Tool

```bash
go install github.com/recally-io/polyllm/cmd/polyllm-cli@latest

# use docker 
docker pull ghcr.io/recally-io/polyllm-cli:latest
```

### HTTP Server

```bash
go install github.com/recally-io/polyllm/cmd/polyllm-server@latest

# use docker 
docker pull ghcr.io/recally-io/polyllm-server:latest
```

## Usage

### API Usage

#### Basic Example

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

	// Generate text using OpenAI's GPT-3.5 Turbo
	// Make sure OPENAI_API_KEY environment variable is set
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

#### Chat Completion Example

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

#### Installation

```bash
# Install the CLI
go install github.com/recally-io/polyllm/cmd/polyllm-cli@latest
```

#### Examples

```bash
# Set your API key
export OPENAI_API_KEY=your_api_key
# or use docker
alias polyllm-cli="docker run --rm -e OPENAI_API_KEY=your_api_key ghcr.io/recally-io/polyllm-cli:latest"

# show help
polyllm-cli --help

# List available models
polyllm-cli models

# Generate text
polyllm-cli -m "openai/gpt-3.5-turbo" "Tell me a joke about programming"

# Use a different model
polyllm-cli -m "deepseek/deepseek-chat" "What is the meaning of life?"
```

### HTTP Server

#### Installation

```bash
# Install the server
go install github.com/recally-io/polyllm/cmd/polyllm-server@latest
```

#### Starting the Server

```bash
# Start the server on default port 8088
export OPENAI_API_KEY=your_api_key
polyllm-server
# or use docker 
docker run --rm -e OPENAI_API_KEY=your_api_key -p 8088:8088 ghcr.io/recally-io/polyllm-server:latest

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
# In a terminal, make a request
curl -X POST http://localhost:8088/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'

# Request to list models
curl http://localhost:8088/models

# Request with authentication (if enabled)
curl -H "Authorization: Bearer your_api_key" http://localhost:8088/models
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
