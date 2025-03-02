package polyllm

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/recally-io/polyllm/llms"
)

func (p *PolyLLM) ListMCPTools(ctx context.Context) ([]llms.Tool, error) {
	mcpNames := slices.Sorted(maps.Keys(p.mcpClientMappings))
	return p.listMCPToolsByMCPNames(ctx, mcpNames)
}

func (p *PolyLLM) listMCPToolsByMCPNames(ctx context.Context, mcpNames []string) ([]llms.Tool, error) {
	llmTools := make([]llms.Tool, 0)
	for _, name := range mcpNames {
		name = strings.TrimSpace(name)
		if client, ok := p.mcpClientMappings[name]; ok {
			mcpTools, err := client.ListTools(ctx, mcp.ListToolsRequest{})
			if err != nil {
				slog.Error("failed to list tools", "err", err, "mcp_server", name)
				continue
			}
			for _, tool := range mcpTools.Tools {
				llmTools = append(llmTools, convertMCPToolToLLMTool(name, tool))
			}
		}
	}
	return llmTools, nil
}

func (p *PolyLLM) getMCPToolsByModel(ctx context.Context, modelInfo string) []llms.Tool {
	// model=gpt-4o?mcp=fetch,everything&tools=fetch,everything
	// Extract MCP servers from model string
	llmTools := make([]llms.Tool, 0)
	params := strings.Split(modelInfo, "&")
	for _, param := range params {
		parts := strings.Split(param, "=")
		if len(parts) != 2 {
			slog.Error("invalid param", "param", param)
			continue
		}
		if parts[0] == "mcp" {
			mcpNames := strings.Split(parts[1], ",")
			if slices.Contains(mcpNames, "all") {
				mcpNames = slices.Sorted(maps.Keys(p.mcpClientMappings))
			}
			tools, err := p.listMCPToolsByMCPNames(ctx, mcpNames)
			if err != nil {
				slog.Error("failed to list tools", "err", err)
				continue
			}
			llmTools = append(llmTools, tools...)
		}
	}
	return llmTools
}
