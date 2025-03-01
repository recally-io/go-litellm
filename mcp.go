package polyllm

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/recally-io/polyllm/llms"
)

func (p *PolyLLM) GetMCPTools(ctx context.Context, modelInfo string) []llms.Tool {
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
		}
	}
	return llmTools
}

func (p *PolyLLM) GetMCPClientByToolName(ctx context.Context, toolName string) (mcpclient.MCPClient, string, error) {
	params := strings.Split(toolName, "_")
	if len(params) != 3 || params[0] != "mcp" {
		return nil, "", fmt.Errorf("tool name must be in format mcp_{mcp_server_name}_{tool_name}")
	}
	mcpName := params[1]
	mcpToolName := params[2]
	if client, ok := p.mcpClientMappings[mcpName]; ok {
		return client, mcpToolName, nil
	}
	return nil, "", fmt.Errorf("tool %s not found", toolName)
}
