package polyllm

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/recally-io/polyllm/llms"
)

func (p *PolyLLM) ChatCompletion(ctx context.Context, req llms.ChatCompletionRequest, streamingFunc func(resp llms.StreamingChatCompletionResponse), options ...llms.RequestOption) {
	finalToolCalls := make([]llms.ToolCall, 0)
	localStreamingFunc := func(resp llms.StreamingChatCompletionResponse) {
		p.streamingFunc(ctx, req, resp, streamingFunc, &finalToolCalls)
	}

	p.chatCompletion(ctx, req, localStreamingFunc, options...)
}

// preProcess preprocess the model and return the llm client, provider model name and llm tools from mcp servers
func (p *PolyLLM) preProcess(ctx context.Context, model string) (LLM, string, []llms.Tool, error) {
	info := strings.Split(model, "?")
	model = info[0]

	llm, ok := p.modelLLMMappings[model]
	if !ok {
		return nil, "", nil, ErrProviderNotFound
	}
	providerModel := llm.GetProvider().GetRealModel(model)

	tools := []llms.Tool{}
	if len(info) > 1 {
		tools = p.GetMCPTools(ctx, info[1])
	}

	return llm, providerModel, tools, nil
}

func (p *PolyLLM) chatCompletion(ctx context.Context, req llms.ChatCompletionRequest, streamingFunc func(resp llms.StreamingChatCompletionResponse), options ...llms.RequestOption) {
	client, model, tools, err := p.preProcess(ctx, req.Model)
	if err != nil {
		slog.Error("failed to get provider", "err", err, "model", req.Model)
		streamingFunc(llms.StreamingChatCompletionResponse{Err: err})
		return
	}
	if len(tools) > 0 {
		req.Tools = append(req.Tools, tools...)
	}
	req.Model = model
	client.ChatCompletion(ctx, req, streamingFunc, options...)
}

func (p *PolyLLM) streamingFunc(ctx context.Context, req llms.ChatCompletionRequest, resp llms.StreamingChatCompletionResponse, userStreamingFunc func(resp llms.StreamingChatCompletionResponse), finalToolCalls *[]llms.ToolCall) {
	// nonstreaming
	if !req.Stream {
		if resp.Err != nil && resp.Err != io.EOF {
			userStreamingFunc(resp)
			return
		}
		toolCalls := resp.Response.Choices[0].Message.ToolCalls
		if len(toolCalls) == 0 {
			userStreamingFunc(resp)
			return
		}

		// invoke mcp tools
		req.Messages = append(req.Messages, llms.ChatCompletionMessage{
			Role:      llms.ChatMessageRoleAssistant,
			ToolCalls: toolCalls,
		})
		messages := p.invokeMCPTools(ctx, toolCalls)
		req.Messages = append(req.Messages, messages...)
		// send tool result to user
		p.ChatCompletion(ctx, req, userStreamingFunc)
	}

	if resp.Err != nil {
		userStreamingFunc(resp)
		return
	}

	choice := resp.Response.Choices[0]

	if choice.FinishReason == llms.FinishReasonToolCalls {
		// invoke mcp tools
		req.Messages = append(req.Messages, llms.ChatCompletionMessage{
			Role:      llms.ChatMessageRoleAssistant,
			ToolCalls: *finalToolCalls,
		})
		messages := p.invokeMCPTools(ctx, *finalToolCalls)
		req.Messages = append(req.Messages, messages...)
		*finalToolCalls = nil
		// send tool result to user
		p.ChatCompletion(ctx, req, userStreamingFunc)
	}

	toolCalls := choice.Delta.ToolCalls
	if len(toolCalls) == 0 {
		userStreamingFunc(resp)
		return
	}

	if len(*finalToolCalls) == 0 {
		*finalToolCalls = append(*finalToolCalls, toolCalls...)
	} else {
		for _, tc := range toolCalls {
			idx := tc.Index
			(*finalToolCalls)[*idx].Function.Arguments += tc.Function.Arguments
		}
	}
}

func (p *PolyLLM) invokeMCPTools(ctx context.Context, tools []llms.ToolCall) []llms.ChatCompletionMessage {
	var messages []llms.ChatCompletionMessage

	for _, tool := range tools {
		mcpName, req, err := convertLLMToolToMCPToolRequest(tool.Function)
		if err != nil {
			slog.Error("failed to convert tool to mcp tool request", "tool", tool.Function.Name, "err", err)
			continue
		}
		resp, err := p.mcpClientMappings[mcpName].CallTool(ctx, req)
		if err != nil {
			slog.Error("failed to call tool", "tool", tool.Function.Name, "err", err, "args", req.Params.Arguments)
			continue
		}

		if resp.Content == nil {
			slog.Error("tool returned nil response", "tool", tool.Function.Name)
			continue
		}

		slog.Info("start invoking mcp tool", "tool", tool.Function.Name, "args", req.Params.Arguments)

		var resultText string

		for _, chunk := range resp.Content {
			if contentMap, ok := chunk.(map[string]any); ok {
				if text, ok := contentMap["text"].(string); ok {
					resultText += fmt.Sprintf("%v", text)
				}
			}
		}
		messages = append(messages, llms.ChatCompletionMessage{
			Role:       llms.ChatMessageRoleTool,
			ToolCallID: tool.ID,
			Content:    strings.TrimSpace(resultText),
		})
		slog.Info("finished invoking mcp tool", "tool", tool.Function.Name, "result", resultText[:min(100, len(resultText))])
	}

	return messages
}
