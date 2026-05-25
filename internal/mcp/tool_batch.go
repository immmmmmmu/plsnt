package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

func (s *Server) registerBatchRunTool() {
	tool := mcp.NewTool("batch_run",
		mcp.WithDescription("Execute a YAML batch definition. Runs multiple plsnt commands in sequence with variable interpolation."),
		mcp.WithString("yaml",
			mcp.Required(),
			mcp.Description("YAML batch definition content"),
		),
		mcp.WithString("template",
			mcp.Description("Template name from templates/ directory (alternative to yaml parameter)"),
		),
		mcp.WithString("variables",
			mcp.Description("JSON object of variables to pass to the batch (e.g., {\"folder_id\": \"12345\"})"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleBatchRun)
}

func (s *Server) handleBatchRun(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	yamlContent, _ := request.GetArguments()["yaml"].(string)
	templateName, _ := request.GetArguments()["template"].(string)

	if yamlContent == "" && templateName == "" {
		return ErrorToToolResult(errs.New(errs.CodeValidationError, "either yaml or template parameter is required")), nil
	}

	// Parse variables if provided.
	var vars map[string]string
	if varsJSON, ok := request.GetArguments()["variables"].(string); ok && varsJSON != "" {
		if err := json.Unmarshal([]byte(varsJSON), &vars); err != nil {
			return ErrorToToolResult(errs.New(errs.CodeInvalidInput, "invalid variables JSON")), nil
		}
	}

	// Batch execution delegates to the CLI subprocess for now.
	// Full in-process execution would require refactoring the batch engine
	// to accept an api.Client directly instead of going through cobra commands.
	// For Phase 1, we return an informative message.
	return ErrorToToolResult(errs.New(errs.CodeInternalError,
		"batch_run is not yet fully implemented in MCP mode. "+
			"Use 'plsnt batch run' via CLI instead.").
		WithSuggestion("Batch execution will be available in a future release")), nil
}
