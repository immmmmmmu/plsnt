package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

func (s *Server) registerSiteCreateTool() {
	tool := mcp.NewTool("site_create",
		mcp.WithDescription("Create a new Pleasanter site (table). Automatically includes SiteSettings to avoid 302 errors."),
		mcp.WithNumber("parent_id",
			mcp.Required(),
			mcp.Description("Parent folder site ID"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Site title"),
		),
		mcp.WithString("reference_type",
			mcp.Description("Table type: Results (default) or Issues"),
		),
		mcp.WithString("body",
			mcp.Description("Site description"),
		),
		mcp.WithString("json",
			mcp.Description("RAW JSON payload for advanced settings (SiteSettings, etc.)"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleSiteCreate)
}

func (s *Server) handleSiteCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	parentID, err := requiredInt64(request, "parent_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	title, _ := request.GetArguments()["title"].(string)
	if title == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError, "title is required")), nil
	}

	refType := "Results"
	if rt, ok := request.GetArguments()["reference_type"].(string); ok && rt != "" {
		refType = rt
	}

	payload := map[string]any{
		"Title":         title,
		"ReferenceType": refType,
		"SiteSettings": map[string]any{
			"ReferenceType": refType,
		},
	}

	if body, ok := request.GetArguments()["body"].(string); ok && body != "" {
		payload["Body"] = body
	}

	if jsonStr, ok := request.GetArguments()["json"].(string); ok && jsonStr != "" {
		var extra map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &extra); err != nil {
			return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid json parameter")), nil
		}
		for k, v := range extra {
			payload[k] = v
		}
	}

	svc := pleasanter.NewSiteService(s.client)
	result, err := svc.Create(ctx, parentID, payload)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return successResult(string(data)), nil
}
