package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
	"github.com/immmmmmmu/plsnt/internal/workflow/export"
	"github.com/immmmmmmu/plsnt/internal/workflow/master"
)

func (s *Server) registerWorkflowDeployTool() {
	tool := mcp.NewTool("workflow_deploy",
		mcp.WithDescription("Deploy workflow tables from a template. Creates master + application tables under a folder."),
		mcp.WithString("template",
			mcp.Required(),
			mcp.Description("Template name from templates/workflow/ directory (without .yaml extension)"),
		),
		mcp.WithNumber("folder_id",
			mcp.Required(),
			mcp.Description("Parent folder site ID to create tables under"),
		),
		mcp.WithString("variables",
			mcp.Description("JSON object of template variables (e.g., {\"dept_site_id\": \"32100\"})"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleWorkflowDeploy)
}

func (s *Server) handleWorkflowDeploy(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// workflow deploy requires batch engine + subprocess execution.
	// Delegate to CLI for now.
	return ErrorToToolResult(errs.New(errs.CodeInternalError,
		"workflow_deploy requires batch engine which is not yet available in MCP mode. "+
			"Use 'plsnt workflow deploy --template <name> --folder-id <id>' via CLI.").
		WithSuggestion("Run: plsnt workflow deploy --template full-deploy --folder-id <folder_id>")), nil
}

func (s *Server) registerWorkflowMasterTool() {
	tool := mcp.NewTool("workflow_master",
		mcp.WithDescription("Import master data from CSV into a Pleasanter table using key-based upsert."),
		mcp.WithNumber("site_id",
			mcp.Required(),
			mcp.Description("Target site ID for master data"),
		),
		mcp.WithString("csv",
			mcp.Required(),
			mcp.Description("CSV data with header row (column names must match Pleasanter field names)"),
		),
		mcp.WithString("key",
			mcp.Description("Column name used as upsert key for matching existing records (e.g., ClassA)"),
		),
		mcp.WithBoolean("dry_run",
			mcp.Description("Preview changes without executing (default: false)"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleWorkflowMaster)
}

func (s *Server) handleWorkflowMaster(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	siteID, err := requiredInt64(request, "site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	csvData, ok := request.GetArguments()["csv"].(string)
	if !ok || csvData == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError, "csv parameter is required")), nil
	}

	key := optionalString(request, "key")
	dryRun, _ := request.GetArguments()["dry_run"].(bool)

	svc := pleasanter.NewRecordService(s.client)
	importer := master.NewImporter(svc, master.Options{
		SiteID: siteID,
		Key:    key,
		DryRun: dryRun,
	})

	result, err := importer.ImportCSV(ctx, strings.NewReader(csvData))
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return successResult(string(data)), nil
}

func (s *Server) registerWorkflowExportTool() {
	tool := mcp.NewTool("workflow_export",
		mcp.WithDescription("Export approved application detail records as CSV."),
		mcp.WithNumber("header_site_id",
			mcp.Required(),
			mcp.Description("Application header table site ID"),
		),
		mcp.WithNumber("detail_site_id",
			mcp.Required(),
			mcp.Description("Application detail table site ID"),
		),
		mcp.WithString("from",
			mcp.Required(),
			mcp.Description("Start date (YYYY-MM-DD)"),
		),
		mcp.WithString("to",
			mcp.Required(),
			mcp.Description("End date (YYYY-MM-DD)"),
		),
		mcp.WithString("statuses",
			mcp.Description("Comma-separated status codes to filter (default: 400,900 = approved + settled)"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleWorkflowExport)
}

func (s *Server) handleWorkflowExport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	headerSiteID, err := requiredInt64(request, "header_site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}
	detailSiteID, err := requiredInt64(request, "detail_site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	from := optionalString(request, "from")
	to := optionalString(request, "to")
	if from == "" || to == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError, "from and to parameters are required")), nil
	}

	// Parse statuses (default: 400,900)
	statusFilter := []int{400, 900}
	if statusStr := optionalString(request, "statuses"); statusStr != "" {
		statusFilter = nil
		for _, s := range splitTrim(statusStr) {
			var code int
			if _, err := fmt.Sscanf(s, "%d", &code); err == nil {
				statusFilter = append(statusFilter, code)
			}
		}
	}

	svc := pleasanter.NewRecordService(s.client)

	// Build date filter: "[from,to]" format for Pleasanter ColumnFilterHash
	dateRange := fmt.Sprintf("[%s,%s]", from, to)

	// Build status filter: comma-separated status codes
	var statusParts []string
	for _, s := range statusFilter {
		statusParts = append(statusParts, fmt.Sprintf("%d", s))
	}

	// Fetch headers with date + status filter
	headerView := &pleasanter.View{
		ColumnFilterHash: map[string]string{
			"DateA":  dateRange,
			"Status": strings.Join(statusParts, ","),
		},
	}

	headers, err := svc.ListAll(ctx, pleasanter.ListOptions{
		SiteID: headerSiteID,
		View:   headerView,
	})
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	if len(headers.Response.Data) == 0 {
		return successResult("No matching records found for the given date range and statuses."), nil
	}

	// Fetch all details
	details, err := svc.ListAll(ctx, pleasanter.ListOptions{SiteID: detailSiteID})
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	// Generate CSV
	var buf bytes.Buffer
	if err := export.GenerateCSV(headers.Response.Data, details.Response.Data, &buf, nil); err != nil {
		return ErrorToToolResult(err), nil
	}

	return successResult(buf.String()), nil
}
