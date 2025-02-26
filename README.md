# PolyLLM

A powerful Go SDK that serves as a unified gateway to interact with 100+ LLM APIs using OpenAI-compatible format. This library simplifies the integration of various LLM providers into your Go applications.

## Features

- 🚀 OpenAI-compatible API interface
- 🔌 Support for multiple LLM providers:
  - OpenAI
  - DeepSeek
  - Qwen
  - Gemini
  - OpenRouter
  - Volcengine
  - Groq
  - XAI
  - Custom OpenAI-compatible endpoints
- 💬 Chat completion support with both streaming and non-streaming options
- 📋 Model listing capability
- ⚙️ Configurable base URLs and API endpoints

## Installation

```bash
go get github.com/recally-io/polyllm
```

## Quick Start

Here's a simple example of how to use polyllm:

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/recally-io/polyllm"
    "github.com/recally-io/polyllm/llms"
)

func main() {
    // Initialize the LLM client, and it will read the API key from the environment variable OPENAI_API_KEY
    llm := polyllm.New(polyllm.ProviderNameOpenAI)

    // Initialize the LLM client with a specific API key and base URL
    llm := polyllm.New(polyllm.ProviderOpenAICompatible, llms.WithBaseURL("https://your-custom-endpoint"), llms.WithAPIKey("your-api-key"))
    
    // Create a chat completion request
    req := llms.ChatCompletionRequest{
        Model: "gpt-3.5-turbo",
        Messages: []llms.ChatCompletionMessage{
            {
                Role:    "user",
                Content: "Hello, how are you?",
            },
        },
    }
    
    // Make the request
    ctx := context.Background()
    llm.ChatCompletion(ctx, req, func(content llms.StreamingChatCompletionResponse) {
        fmt.Print(content.Response.Choices[0].Message.Content)
    })
}
```

## Advanced Usage

### Streaming Responses

```go
req := llms.ChatCompletionRequest{
    Model:  "gpt-4",
    Stream: true,
    Messages: []llms.ChatCompletionMessage{
        {
            Role:    "user",
            Content: "Write a story",
        },
    },
}

llm.ChatCompletion(ctx, req, func(content llms.StreamingChatCompletionResponse) {
    if content.Err != nil && content.Err != io.EOF {
        log.Printf("Stream error: %v", content.Err)
        return
    }
    fmt.Print(content.Response.Choices[0].Delta.Content)
})
```

### Custom Provider Configuration

```go
llm := polyllm.New(polyllm.ProviderOpenAICompatible, "your-api-key",
    llms.WithBaseURL("https://your-custom-endpoint"),
    llms.WithPrefix("custom-prefix"),
)
```

### Listing Available Models

```go
models, err := llm.ListModels(ctx)
if err != nil {
    panic(err)
}

for _, model := range models {
    fmt.Println(model)
}
```

## Supported Providers

The library supports various LLM providers through a unified interface. Each provider can be initialized using the corresponding provider name constant:

- `ProviderNameOpenAI`: OpenAI API
- `ProviderNameDeepSeek`: DeepSeek API
- `ProviderNameQwen`: Qwen API
- `ProviderNameGemini`: Google's Gemini API
- `ProviderNameOpenRouter`: OpenRouter API
- `ProviderNameVolcengine`: Volcengine API
- `ProviderNameGroq`: Groq API
- `ProviderNameXai`: XAI API
- `ProviderOpenAICompatible`: Any OpenAI-compatible endpoint

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the terms of the LICENSE file included in the repository.
