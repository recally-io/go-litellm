package polyllm

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/recally-io/polyllm/llms"
)

func convertMCPToolToLLMTool(mcpName string, tool mcp.Tool) llms.Tool {
	return llms.Tool{
		Type: llms.ToolTypeFunction,
		Function: &llms.FunctionDefinition{
			Name:        fmt.Sprintf("mcp_%s_%s", mcpName, tool.Name),
			Description: tool.Description,
			Parameters:  tool.InputSchema,
		},
	}
}

func convertLLMToolToMCPToolRequest(tool llms.FunctionCall) (string, mcp.CallToolRequest, error) {
	req := mcp.CallToolRequest{}
	params := strings.Split(tool.Name, "_")
	if len(params) < 3 || params[0] != "mcp" {
		return "", req, fmt.Errorf("tool name must be in format mcp_{mcp_server_name}_{tool_name}")
	}
	mcpName := params[1]

	mcpToolName := strings.Join(params[2:], "_")
	req.Params.Name = mcpToolName
	var arguments map[string]any
	if err := json.Unmarshal([]byte(tool.Arguments), &arguments); err != nil {
		slog.Error("failed to unmarshal arguments", "err", err)
		return "", req, err
	}
	req.Params.Arguments = arguments

	return mcpName, req, nil
}
