package mcp

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

func TestCLIErrorToMCPResult(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantText string
		wantErr  bool
	}{
		{
			name:     "validation error",
			err:      errs.New(errs.CodeValidationError, "site ID is required"),
			wantText: "site ID is required",
			wantErr:  true,
		},
		{
			name:     "not found error",
			err:      errs.New(errs.CodeRecordNotFound, "record 123 not found"),
			wantText: "record 123 not found",
			wantErr:  true,
		},
		{
			name:     "auth error",
			err:      errs.New(errs.CodeAuthError, "authentication failed"),
			wantText: "authentication failed",
			wantErr:  true,
		},
		{
			name:     "error with suggestion",
			err:      errs.NewWithSuggestion(errs.CodeConnectionError, "connection failed", "check URL"),
			wantText: "connection failed\nSuggestion: check URL",
			wantErr:  true,
		},
		{
			name:     "generic error",
			err:      assert.AnError,
			wantText: "assert.AnError general error for testing",
			wantErr:  true,
		},
		{
			name:    "nil error",
			err:     nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrorToToolResult(tt.err)

			if !tt.wantErr {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Len(t, result.Content, 1)

			textContent, ok := result.Content[0].(mcp.TextContent)
			assert.True(t, ok, "content should be TextContent")
			assert.Equal(t, tt.wantText, textContent.Text)
		})
	}
}
