package mcps

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Provider struct {
	// BaseURl for sse mcp server
	BaseURL string            `json:"base_url,omitempty"`
	// Command for stdio mcp server
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func CreateMCPClients(
	config map[string]Provider,
) (map[string]mcpclient.MCPClient, error) {
	clients := make(map[string]mcpclient.MCPClient)
	var err error
	for name, server := range config {
		var client mcpclient.MCPClient
		if server.BaseURL != "" {
			client, err = mcpclient.NewSSEMCPClient(server.BaseURL)
		} else {
			var env []string
			for k, v := range server.Env {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}
			client, err = mcpclient.NewStdioMCPClient(
				server.Command,
				env,
				server.Args...)
		}
		if err != nil {
			for _, c := range clients {
				c.Close()
			}
			return nil, fmt.Errorf(
				"failed to create MCP client for %s: %w",
				name,
				err,
			)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		slog.Info("Initializing mcp server...", "name", name)
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "polyllm",
			Version: "0.1.0",
		}
		initRequest.Params.Capabilities = mcp.ClientCapabilities{}

		_, err = client.Initialize(ctx, initRequest)
		if err != nil {
			client.Close()
			for _, c := range clients {
				c.Close()
			}
			return nil, fmt.Errorf(
				"failed to initialize MCP client for %s: %w",
				name,
				err,
			)
		}

		clients[name] = client
		slog.Info("Initialized mcp server", "name", name)
	}

	return clients, nil
}
