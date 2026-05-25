package mcp

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleWorkflowDeploy_Stub(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"template":  "full-deploy",
		"folder_id": float64(100),
	}

	result, err := srv.handleWorkflowDeploy(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "not yet available")
}

func TestHandleWorkflowMaster_Success(t *testing.T) {
	callCount := 0
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// First call: ListAll to check existing records
		// Subsequent calls: Create/Update records
		if callCount == 1 {
			fmt.Fprint(w, `{"StatusCode":200,"Response":{"Offset":0,"PageSize":200,"TotalCount":0,"Data":[]}}`)
		} else {
			fmt.Fprint(w, `{"StatusCode":200,"Id":1,"Message":"Created"}`)
		}
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"csv":     "Title,ClassA\nItem1,Cat1\nItem2,Cat2",
		"key":     "ClassA",
	}

	result, err := srv.handleWorkflowMaster(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleWorkflowMaster_MissingCSV(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"site_id": float64(100)}

	result, err := srv.handleWorkflowMaster(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleWorkflowExport_NoRecords(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Response":{"Offset":0,"PageSize":200,"TotalCount":0,"Data":[]}}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"header_site_id": float64(100),
		"detail_site_id": float64(200),
		"from":           "2026-01-01",
		"to":             "2026-03-31",
	}

	result, err := srv.handleWorkflowExport(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "No matching records")
}

func TestHandleWorkflowExport_MissingDates(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"header_site_id": float64(100),
		"detail_site_id": float64(200),
	}

	result, err := srv.handleWorkflowExport(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleMigrateExecute_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Id":1,"Message":"Created"}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"csv":     "Name,Category\nAlice,A\nBob,B",
		"mapping": "columns:\n  Name: Title\n  Category: ClassA",
	}

	result, err := srv.handleMigrateExecute(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, `"migrated": 2`)
	assert.Contains(t, text, `"mode": "create"`)
}

func TestHandleMigrateExecute_WithKeys(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Message":"Upserted"}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"csv":     "Name,Category\nAlice,A",
		"mapping": "columns:\n  Name: Title\n  Category: ClassA",
		"keys":    "ClassA",
	}

	result, err := srv.handleMigrateExecute(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, `"mode": "upsert"`)
}

func TestHandleMigrateExecute_MissingMapping(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"csv":     "A,B\n1,2",
	}

	result, err := srv.handleMigrateExecute(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleMigrateExecute_InvalidMapping(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"csv":     "A,B\n1,2",
		"mapping": "{{invalid yaml",
	}

	result, err := srv.handleMigrateExecute(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestPrompts_CreateApp(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.GetPromptRequest{}
	request.Params.Arguments = map[string]string{
		"app_name":  "勤怠管理",
		"folder_id": "32085",
	}

	result, err := srv.handleCreateAppPrompt(context.Background(), request)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Messages)
	text := result.Messages[0].Content.(mcp.TextContent).Text
	assert.Contains(t, text, "勤怠管理")
	assert.Contains(t, text, "32085")
}

func TestPrompts_MigrateCSV(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.GetPromptRequest{}
	request.Params.Arguments = map[string]string{
		"site_id": "12345",
	}

	result, err := srv.handleMigrateCSVPrompt(context.Background(), request)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Messages)
	text := result.Messages[0].Content.(mcp.TextContent).Text
	assert.Contains(t, text, "12345")
}
