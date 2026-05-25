package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

func (s *Server) registerSiteGetTool() {
	tool := mcp.NewTool("site_get",
		mcp.WithDescription("Get site metadata and settings using getsite API. Returns full SiteSettings."),
		mcp.WithNumber("site_id",
			mcp.Required(),
			mcp.Description("Site ID"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleSiteGet)
}

func (s *Server) handleSiteGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	siteID, err := requiredInt64(request, "site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	svc := pleasanter.NewSiteService(s.client)
	result, err := svc.Get(ctx, siteID)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return successResult(string(data)), nil
}

func (s *Server) registerSiteUpdateTool() {
	tool := mcp.NewTool("site_update",
		mcp.WithDescription("Update site settings. IMPORTANT: SiteSettings is fully overwritten — always include the complete SiteSettings object."),
		mcp.WithNumber("site_id",
			mcp.Required(),
			mcp.Description("Site ID to update"),
		),
		mcp.WithString("json",
			mcp.Required(),
			mcp.Description("JSON payload with fields to update (Title, Body, SiteSettings, etc.)"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleSiteUpdate)
}

func (s *Server) handleSiteUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	siteID, err := requiredInt64(request, "site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	jsonStr, ok := request.GetArguments()["json"].(string)
	if !ok || jsonStr == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError, "json parameter is required")), nil
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid json parameter")), nil
	}

	svc := pleasanter.NewSiteService(s.client)
	result, err := svc.Update(ctx, siteID, payload)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return successResult(string(data)), nil
}

func (s *Server) registerSiteCopyTool() {
	tool := mcp.NewTool("site_copy",
		mcp.WithDescription("Copy a site (table) to a new location. Copies settings, columns, and configuration."),
		mcp.WithNumber("source_site_id",
			mcp.Required(),
			mcp.Description("Site ID of the source to copy"),
		),
		mcp.WithNumber("parent_id",
			mcp.Required(),
			mcp.Description("Destination parent folder ID"),
		),
		mcp.WithString("title",
			mcp.Description("Override title for the copied site"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleSiteCopy)
}

func (s *Server) handleSiteCopy(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourceID, err := requiredInt64(request, "source_site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}
	parentID, err := requiredInt64(request, "parent_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	overrides := make(map[string]any)
	if title := optionalString(request, "title"); title != "" {
		overrides["Title"] = title
	}

	svc := pleasanter.NewSiteService(s.client)
	result, err := svc.Copy(ctx, sourceID, parentID, overrides)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return successResult(string(data)), nil
}

func (s *Server) registerSiteSearchTool() {
	tool := mcp.NewTool("site_search",
		mcp.WithDescription("Search child sites under a parent folder by title keyword."),
		mcp.WithNumber("parent_id",
			mcp.Required(),
			mcp.Description("Parent folder site ID to search under"),
		),
		mcp.WithString("keyword",
			mcp.Required(),
			mcp.Description("Title keyword to search"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleSiteSearch)
}

func (s *Server) handleSiteSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	parentID, err := requiredInt64(request, "parent_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	keyword, ok := request.GetArguments()["keyword"].(string)
	if !ok || keyword == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError, "keyword is required")), nil
	}

	svc := pleasanter.NewSiteService(s.client)
	result, err := svc.Search(ctx, parentID, keyword)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return successResult(string(data)), nil
}
