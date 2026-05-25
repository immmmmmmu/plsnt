package mcp

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"

	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

func (s *Server) registerRecordUpsertTool() {
	tool := mcp.NewTool("record_upsert",
		mcp.WithDescription("Bulk create or update records in a Pleasanter table. Uses bulkupsert API endpoint."),
		mcp.WithNumber("site_id",
			mcp.Required(),
			mcp.Description("Target site ID"),
		),
		mcp.WithString("json",
			mcp.Required(),
			mcp.Description("JSON array of record objects, or a single record object"),
		),
		mcp.WithString("key",
			mcp.Description("Column name used as upsert key (e.g., ClassA). If set, existing records matching the key value are updated instead of created."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleRecordUpsert)
}

func (s *Server) handleRecordUpsert(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	siteID, err := requiredInt64(request, "site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	jsonStr, ok := request.GetArguments()["json"].(string)
	if !ok || jsonStr == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError, "json parameter is required")), nil
	}

	key, _ := request.GetArguments()["key"].(string)

	// Parse records — accept array or single object.
	var records []map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &records); err != nil {
		var single map[string]any
		if err2 := json.Unmarshal([]byte(jsonStr), &single); err2 != nil {
			return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "json must be an array or object")), nil
		}
		records = []map[string]any{single}
	}

	svc := pleasanter.NewRecordService(s.client)

	var keys []string
	if key != "" {
		keys = []string{key}
	}

	results, err := svc.Import(ctx, siteID, records, keys)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(map[string]any{
		"upserted": len(results),
		"results":  results,
	}, "", "  ")
	return successResult(string(data)), nil
}

func (s *Server) registerRecordImportTool() {
	tool := mcp.NewTool("record_import",
		mcp.WithDescription("Import records from CSV data into a Pleasanter table."),
		mcp.WithNumber("site_id",
			mcp.Required(),
			mcp.Description("Target site ID"),
		),
		mcp.WithString("csv",
			mcp.Required(),
			mcp.Description("CSV data with header row"),
		),
		mcp.WithString("mapping",
			mcp.Description("YAML mapping definition (column name mappings). If omitted, header names are used as-is."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleRecordImport)
}

func (s *Server) handleRecordImport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	siteID, err := requiredInt64(request, "site_id")
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	csvData, ok := request.GetArguments()["csv"].(string)
	if !ok || csvData == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError, "csv parameter is required")), nil
	}

	// Build mapping: if mapping YAML is provided, parse it; otherwise use identity mapping.
	mappingYAML := optionalString(request, "mapping")
	var mapping *pleasanter.MappingConfig

	if mappingYAML != "" {
		var cfg pleasanter.MappingConfig
		if err := yaml.Unmarshal([]byte(mappingYAML), &cfg); err != nil {
			return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid mapping YAML")), nil
		}
		mapping = &cfg
	} else {
		// Auto-generate identity mapping from CSV header.
		mapping, err = autoMapping(csvData)
		if err != nil {
			return ErrorToToolResult(err), nil
		}
	}

	records, err := pleasanter.ParseCSV(strings.NewReader(csvData), mapping)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	svc := pleasanter.NewRecordService(s.client)
	results, err := svc.Import(ctx, siteID, records, nil)
	if err != nil {
		return ErrorToToolResult(err), nil
	}

	data, _ := json.MarshalIndent(map[string]any{
		"imported": len(results),
		"site_id":  fmt.Sprintf("%d", siteID),
	}, "", "  ")
	return successResult(string(data)), nil
}

// autoMapping creates an identity mapping from CSV header columns.
func autoMapping(csvData string) (*pleasanter.MappingConfig, error) {
	r := csv.NewReader(strings.NewReader(csvData))
	header, err := r.Read()
	if err != nil {
		return nil, errs.New(errs.CodeInvalidInput, "failed to read CSV header")
	}

	cols := make(map[string]string, len(header))
	for _, h := range header {
		h = strings.TrimSpace(h)
		if h != "" {
			cols[h] = h
		}
	}

	return &pleasanter.MappingConfig{Columns: cols}, nil
}
