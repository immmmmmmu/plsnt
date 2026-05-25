package mcp

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

// requiredInt64 extracts and validates a required integer parameter from a tool call request.
func requiredInt64(request mcp.CallToolRequest, name string) (int64, error) {
	args := request.GetArguments()
	val, ok := args[name]
	if !ok {
		return 0, errs.New(errs.CodeValidationError,
			fmt.Sprintf("%s is required", name))
	}

	switch v := val.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, errs.New(errs.CodeValidationError,
			fmt.Sprintf("%s must be a number, got %T", name, val))
	}
}

// optionalString extracts an optional string parameter.
func optionalString(request mcp.CallToolRequest, name string) string {
	args := request.GetArguments()
	val, ok := args[name].(string)
	if !ok {
		return ""
	}
	return val
}

// splitTrim splits a string by comma and trims whitespace from each part.
func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
