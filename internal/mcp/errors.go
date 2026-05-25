package mcp

import (
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

// ErrorToToolResult converts an error to an MCP CallToolResult with IsError=true.
// Returns nil if err is nil.
func ErrorToToolResult(err error) *mcp.CallToolResult {
	if err == nil {
		return nil
	}

	msg := err.Error()

	var cliErr *errs.CLIError
	if errors.As(err, &cliErr) && cliErr.ErrorBody != nil {
		msg = cliErr.ErrorBody.Message
		if cliErr.ErrorBody.Suggestion != "" {
			msg = fmt.Sprintf("%s\nSuggestion: %s", msg, cliErr.ErrorBody.Suggestion)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(msg),
		},
		IsError: true,
	}
}
