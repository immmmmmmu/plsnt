package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

func (s *Server) registerRecordListTool() {
	tool := mcp.NewTool("record_list",
		mcp.WithDescription("List records from a Pleasanter table. Supports filtering with View, field selection, and automatic pagination."),
		mcp.WithNumber("site_id",
			mcp.Required(),
			mcp.Description("Site ID of the table"),
		),
		mcp.WithString("view",
			mcp.Description(`View filter JSON, e.g. {"ColumnFilterHash":{"ClassA":"Red"}}`),
		),
		mcp.WithString("fields",
			mcp.Description("Comma-separated field names to include in output"),
		),
		mcp.WithBoolean("all_pages",
			mcp.Description("Fetch all pages automatically (default: false, returns first page only)"),
		),
		mcp.WithString("json",
			mcp.Description("RAW JSON payload (overrides other parameters)"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleRecordList)
}

func (s *Server) handleRecordList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	siteID, err := requiredInt64(request, "site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	svc := pleasanter.NewRecordService(s.client)

	// RAW JSON bypass
	if jsonStr := optionalString(request, "json"); jsonStr != "" {
		var payload map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
			return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid json parameter")), nil
		}
		result, err := svc.ListRaw(ctx, siteID, payload)
		if err != nil {
			return ErrorToToolResult(err), nil
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		return successResult(string(data)), nil
	}

	// Parse View filter
	var view *pleasanter.View
	if viewStr := optionalString(request, "view"); viewStr != "" {
		view = &pleasanter.View{}
		if err := json.Unmarshal([]byte(viewStr), view); err != nil {
			return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid view JSON")), nil
		}
	}

	opts := pleasanter.ListOptions{SiteID: siteID, View: view}

	allPages, _ := request.GetArguments()["all_pages"].(bool)

	var resp *pleasanter.APIResponse
	if allPages {
		resp, err = svc.ListAll(ctx, opts)
	} else {
		resp, err = svc.List(ctx, opts)
	}
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	return successResult(string(data)), nil
}

func (s *Server) registerRecordGetTool() {
	tool := mcp.NewTool("record_get",
		mcp.WithDescription("Get a single record by ID."),
		mcp.WithNumber("record_id",
			mcp.Required(),
			mcp.Description("Record ID"),
		),
		mcp.WithString("json",
			mcp.Description("RAW JSON payload for advanced options"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleRecordGet)
}

func (s *Server) handleRecordGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	recordID, err := requiredInt64(request, "record_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	svc := pleasanter.NewRecordService(s.client)

	if jsonStr := optionalString(request, "json"); jsonStr != "" {
		var payload map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
			return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid json parameter")), nil
		}
		result, err := svc.GetRaw(ctx, recordID, payload)
		if err != nil {
			return ErrorToToolResult(err), nil
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		return successResult(string(data)), nil
	}

	resp, err := svc.Get(ctx, recordID)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	return successResult(string(data)), nil
}

func (s *Server) registerRecordCreateTool() {
	tool := mcp.NewTool("record_create",
		mcp.WithDescription("Create a new record in a Pleasanter table. Not available in Pleasanter MCP."),
		mcp.WithNumber("site_id",
			mcp.Required(),
			mcp.Description("Target site ID"),
		),
		mcp.WithString("title",
			mcp.Description("Record title"),
		),
		mcp.WithString("body",
			mcp.Description("Record body"),
		),
		mcp.WithNumber("status",
			mcp.Description("Record status code"),
		),
		mcp.WithString("json",
			mcp.Description("RAW JSON payload for custom fields (ClassHash, NumHash, etc.)"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleRecordCreate)
}

func (s *Server) handleRecordCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	siteID, err := requiredInt64(request, "site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	payload := make(map[string]any)

	if title := optionalString(request, "title"); title != "" {
		payload["Title"] = title
	}
	if body := optionalString(request, "body"); body != "" {
		payload["Body"] = body
	}
	if statusVal, ok := request.GetArguments()["status"]; ok {
		if s, ok := statusVal.(float64); ok {
			payload["Status"] = int(s)
		}
	}

	// Merge RAW JSON (takes precedence).
	if jsonStr := optionalString(request, "json"); jsonStr != "" {
		var extra map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &extra); err != nil {
			return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid json parameter")), nil
		}
		for k, v := range extra {
			payload[k] = v
		}
	}

	svc := pleasanter.NewRecordService(s.client)
	result, err := svc.Create(ctx, siteID, payload)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return successResult(string(data)), nil
}

func (s *Server) registerRecordUpdateTool() {
	tool := mcp.NewTool("record_update",
		mcp.WithDescription("Update an existing record. Single-step operation (unlike Pleasanter MCP's 2-step flow)."),
		mcp.WithNumber("record_id",
			mcp.Required(),
			mcp.Description("Record ID to update"),
		),
		mcp.WithString("title",
			mcp.Description("New title"),
		),
		mcp.WithString("body",
			mcp.Description("New body"),
		),
		mcp.WithNumber("status",
			mcp.Description("New status code"),
		),
		mcp.WithString("json",
			mcp.Description("RAW JSON payload for custom fields"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleRecordUpdate)
}

func (s *Server) handleRecordUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	recordID, err := requiredInt64(request, "record_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	payload := make(map[string]any)

	if title := optionalString(request, "title"); title != "" {
		payload["Title"] = title
	}
	if body := optionalString(request, "body"); body != "" {
		payload["Body"] = body
	}
	if statusVal, ok := request.GetArguments()["status"]; ok {
		if s, ok := statusVal.(float64); ok {
			payload["Status"] = int(s)
		}
	}

	if jsonStr := optionalString(request, "json"); jsonStr != "" {
		var extra map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &extra); err != nil {
			return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid json parameter")), nil
		}
		for k, v := range extra {
			payload[k] = v
		}
	}

	svc := pleasanter.NewRecordService(s.client)
	result, err := svc.Update(ctx, recordID, payload)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return successResult(string(data)), nil
}

func (s *Server) registerRecordDeleteTool() {
	tool := mcp.NewTool("record_delete",
		mcp.WithDescription("Delete a single record. Not available in Pleasanter MCP."),
		mcp.WithNumber("record_id",
			mcp.Required(),
			mcp.Description("Record ID to delete"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleRecordDelete)
}

func (s *Server) handleRecordDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	recordID, err := requiredInt64(request, "record_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	svc := pleasanter.NewRecordService(s.client)
	result, err := svc.Delete(ctx, recordID)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return successResult(string(data)), nil
}

func (s *Server) registerRecordBulkDeleteTool() {
	tool := mcp.NewTool("record_bulk_delete",
		mcp.WithDescription("Delete multiple records by IDs or by View filter."),
		mcp.WithNumber("site_id",
			mcp.Required(),
			mcp.Description("Site ID"),
		),
		mcp.WithString("record_ids",
			mcp.Description("Comma-separated record IDs to delete"),
		),
		mcp.WithString("view",
			mcp.Description("View filter JSON for condition-based deletion"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleRecordBulkDelete)
}

func (s *Server) handleRecordBulkDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	siteID, err := requiredInt64(request, "site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	svc := pleasanter.NewRecordService(s.client)

	// By View filter
	if viewStr := optionalString(request, "view"); viewStr != "" {
		var view map[string]any
		if err := json.Unmarshal([]byte(viewStr), &view); err != nil {
			return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid view JSON")), nil
		}
		result, err := svc.BulkDeleteByView(ctx, siteID, view)
		if err != nil {
			return ErrorToToolResult(err), nil
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		return successResult(string(data)), nil
	}

	// By record IDs
	idsStr := optionalString(request, "record_ids")
	if idsStr == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError,
			"either record_ids or view parameter is required")), nil
	}

	ids, err := parseIDList(idsStr)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	result, err := svc.BulkDelete(ctx, siteID, ids)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return successResult(string(data)), nil
}

// parseIDList parses a comma-separated list of integer IDs.
func parseIDList(s string) ([]int64, error) {
	var ids []int64
	for _, part := range splitTrim(s) {
		var id int64
		if _, err := fmt.Sscanf(part, "%d", &id); err != nil {
			return nil, errs.New(errs.CodeInvalidInput,
				fmt.Sprintf("invalid record ID: %s", part))
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errs.New(errs.CodeValidationError, "no valid record IDs provided")
	}
	return ids, nil
}
