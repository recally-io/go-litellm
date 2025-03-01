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
	- [Configuration](#configuration)
		- [JSON Configuration File](#json-configuration-file)
		- [MCP Configuration](#mcp-configuration)
	- [Usage](#usage)
		- [API Usage](#api-usage)
			- [Basic Example](#basic-example)
			- [Using Configuration File](#using-configuration-file)
			- [Chat Completion Example](#chat-completion-example)
			- [Using MCP](#using-mcp)
		- [CLI Usage](#cli-usage)
			- [Installation](#installation-1)
			- [Examples](#examples)
		- [HTTP Server](#http-server-1)
			- [Installation](#installation-2)
			- [Starting the Server](#starting-the-server)
			- [API Endpoints](#api-endpoints)
			- [Example Request](#example-request)
	- [License](#license)

## Features

- **Single Interface**: Interact with multiple LLM providers through a unified API
- **Provider Agnostic**: Easily switch between providers without changing your code
- **Multiple Interfaces**: Access LLMs through a Go API, CLI tool, or HTTP server
- **MCP Support**: Builtin support for [Model Context Protocol](https://modelcontextprotocol.io/introduction)
- **Configuration File**: JSON-based configuration for providers and MCP tools

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

## Configuration

### JSON Configuration File

PolyLLM supports configuration via a JSON file. This allows you to define LLM providers and MCP tools.

Example configuration file:

```json
{
  "llms": [
    {
      "name": "gemini",
      "type": "gemini",
      "base_url": "https://generativelanguage.googleapis.com/v1beta/openai",
      "env_prefix": "GEMINI_",
      "api_key": "<GOOGLE_API_KEY>",
      "models": [
        {
          "id": "gemini-2.0-flash"
        },
        {
          "id": "gemini-2.0-flash-lite"
        },
        {
          "id": "gemini-1.5-flash"
        },
        {
          "id": "gemini-1.5-flash-8b"
        },
        {
          "id": "gemini-1.5-pro"
        },
        {
          "id": "text-embedding-004"
        }
      ]
    }
  ],
  "mcps": {
    "fetch": {
      "command": "uvx",
      "args": [
        "mcp-server-fetch"
      ]
    },
    "puppeteer": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "--init",
        "-e",
        "DOCKER_CONTAINER=true",
        "mcp/puppeteer"
      ]
    }
  }
}
```

### MCP Configuration

Model Context Protocol (MCP) tools can be defined in the configuration file under the `mcps` section. Each tool is specified with a command and arguments.

To use MCP tools with a model, append `?mcp=<tool1>,<tool2>` to the model name or use `?mcp=all` to enable all configured MCP tools.

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

#### Using Configuration File

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/recally-io/polyllm"
)

func main() {
	// Load configuration from file
	cfg, err := polyllm.LoadConfig("config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create a new PolyLLM instance with the configuration
	llm := polyllm.NewFromConfig(cfg)

	// Generate text using a configured model
	response, err := llm.GenerateText(
		context.Background(),
		"gemini/gemini-1.5-pro",
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

#### Using MCP

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
	// Load configuration from file
	cfg, err := polyllm.LoadConfig("config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create a new PolyLLM instance with the configuration
	llm := polyllm.NewFromConfig(cfg)

	// Create a chat completion request with MCP enabled
	// Use all configured MCP tools
	req := llms.ChatCompletionRequest{
		Model: "qwen/qwen-max?mcp=all", // Use all MCP tools
		Messages: []llms.Message{
			{
				Role:    llms.RoleUser,
				Content: "List the top 10 news from Hacker News",
			},
		},
	}

	// Or specify specific MCP tools
	req2 := llms.ChatCompletionRequest{
		Model: "qwen/qwen-max?mcp=fetch,puppeteer", // Use specific MCP tools
		Messages: []llms.Message{
			{
				Role:    llms.RoleUser,
				Content: "List the top 10 news from Hacker News",
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

# Using a config file
polyllm-cli -c "config.json" -m "gemini/gemini-1.5-pro" "What is quantum computing?"

# Using MCP with all tools
polyllm-cli -c "config.json" -m "qwen/qwen-max?mcp=all" "Top 10 news in hackernews"

# Using MCP with specific tools
polyllm-cli -c "config.json" -m "qwen/qwen-max?mcp=fetch,puppeteer" "Top 10 news in hackernews"
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

# Start with a configuration file
polyllm-server -c config.json

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

# Request with MCP enabled
curl -X POST http://localhost:8088/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen/qwen-max?mcp=all",
    "messages": [
      {"role": "user", "content": "Top 10 news in hackernews"}
    ]
  }'
```

## License

This project is licensed under the terms provided in the [LICENSE](LICENSE) file.