package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

func (s *Server) registerConfigTestTool() {
	tool := mcp.NewTool("config_test",
		mcp.WithDescription("Test connection to Pleasanter server. Returns server info on success."),
	)

	s.mcpServer.AddTool(tool, s.handleConfigTest)
}

func (s *Server) handleConfigTest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	svc := pleasanter.NewUserService(s.client)
	result, err := svc.List(ctx)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(map[string]any{
		"status":  "connected",
		"profile": s.profile,
		"response": result,
	}, "", "  ")
	return successResult(string(data)), nil
}
