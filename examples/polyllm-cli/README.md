# PolyLLM CLI

A command-line interface for interacting with various LLM providers through the PolyLLM library.

## Features

- List available models from different providers
- Chat with models in streaming mode

## Installation

```bash
# Clone the repository
git clone https://github.com/recally-io/polyllm.git
cd polyllm/examples/polyllm-cli

# Build the CLI
go build -o polyllm-cli ./cmd/polyllm-cli
```

## Usage

### List Available Models

```bash
./polyllm-cli models
```

This command will attempt to list models from all supported providers. For each provider, it will try to authenticate using the corresponding environment variable for the API key.

### Chat with a Model

```bash
./polyllm-cli -m "model-name" "Your prompt here"
```

Examples:
```bash
# Chat with OpenAI's GPT-4
./polyllm-cli -m "gpt-4" "Tell me a joke"

# Chat with a specific provider's model using the provider/model format
./polyllm-cli -m "deepseek/deepseek-chat" "What is the meaning of life?"
```

## Environment Variables

The CLI uses environment variables for API keys. Set the following variables for the providers you want to use:

- `OPENAI_API_KEY`: For OpenAI models
- `DEEPSEEK_API_KEY`: For DeepSeek models
- `QWEN_API_KEY`: For Qwen models
- `GEMINI_API_KEY`: For Gemini models
- `OPENROUTER_API_KEY`: For OpenRouter models
- `VOLCENGINE_API_KEY`: For Volcengine models
- `GROQ_API_KEY`: For Groq models
- `XAI_API_KEY`: For Xai models

## Model Name Format

You can specify models in two formats:

1. Simple format: `model-name` (defaults to OpenAI)
   - Example: `gpt-4`

2. Provider-specific format: `provider/model-name`
   - Example: `deepseek/deepseek-chat`

## Error Handling

If a provider's API key is not set or is invalid, the CLI will display an error message but continue with other providers when listing models.

When chatting with a model, if there's an error initializing the client or during the chat, the CLI will display the error and exit.
