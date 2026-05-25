package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

func (s *Server) registerSchemaGetTool() {
	tool := mcp.NewTool("schema_get",
		mcp.WithDescription("Get column definitions for a Pleasanter table (site). Returns structured column info including types, choices, and constraints."),
		mcp.WithNumber("site_id",
			mcp.Required(),
			mcp.Description("Site ID of the table to inspect"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleSchemaGet)
}

func (s *Server) handleSchemaGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	siteID, err := requiredInt64(request, "site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	svc := pleasanter.NewSchemaService(s.client)
	schema, err := svc.GetSchema(ctx, siteID)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(schema, "", "  ")
	return successResult(string(data)), nil
}
