package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"

	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

func (s *Server) registerMigrateExecuteTool() {
	tool := mcp.NewTool("migrate_execute",
		mcp.WithDescription("Execute a CSV data migration into a Pleasanter table using a mapping definition."),
		mcp.WithNumber("site_id",
			mcp.Required(),
			mcp.Description("Target site ID"),
		),
		mcp.WithString("csv",
			mcp.Required(),
			mcp.Description("CSV data with header row"),
		),
		mcp.WithString("mapping",
			mcp.Required(),
			mcp.Description("YAML mapping definition (columns section maps CSV headers to Pleasanter fields)"),
		),
		mcp.WithString("keys",
			mcp.Description("Comma-separated upsert key columns. If set, uses bulkupsert for update-or-insert. Without keys, creates records one by one."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleMigrateExecute)
}

func (s *Server) handleMigrateExecute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	siteID, err := requiredInt64(request, "site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	csvData, ok := request.GetArguments()["csv"].(string)
	if !ok || csvData == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError, "csv parameter is required")), nil
	}

	mappingYAML, ok := request.GetArguments()["mapping"].(string)
	if !ok || mappingYAML == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError, "mapping parameter is required")), nil
	}

	var mapping pleasanter.MappingConfig
	if err := yaml.Unmarshal([]byte(mappingYAML), &mapping); err != nil {
		return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid mapping YAML")), nil
	}
	if len(mapping.Columns) == 0 {
		return ErrorToToolResult(errs.New(errs.CodeInvalidInput,
			"mapping must define at least one column")), nil
	}

	records, err := pleasanter.ParseCSV(strings.NewReader(csvData), &mapping)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	if len(records) == 0 {
		return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "CSV contains no data rows")), nil
	}

	var keys []string
	if keysStr := optionalString(request, "keys"); keysStr != "" {
		keys = splitTrim(keysStr)
	}

	svc := pleasanter.NewRecordService(s.client)
	results, err := svc.Import(ctx, siteID, records, keys)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(map[string]any{
		"migrated": len(results),
		"site_id":  fmt.Sprintf("%d", siteID),
		"mode":     migrationMode(keys),
	}, "", "  ")
	return successResult(string(data)), nil
}

func migrationMode(keys []string) string {
	if len(keys) > 0 {
		return "upsert"
	}
	return "create"
}
