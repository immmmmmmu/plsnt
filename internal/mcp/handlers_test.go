package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/config"
)

// newTestServer creates a Server with a mock HTTP backend.
func newTestServer(t *testing.T, handler http.HandlerFunc) *Server {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	client := api.NewWithHTTPClient(ts.URL, "test-key", "1.1", ts.Client())
	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles: map[string]*config.Profile{
			"test": {URL: ts.URL, APIKey: "test-key"},
		},
	}
	return NewServer("test", client, cfg, "test")
}

func TestHandleConfigTest_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Response":{"Data":[],"TotalCount":5}}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{}

	result, err := srv.handleConfigTest(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "connected")
	assert.Contains(t, text, "test")
}

func TestHandleConfigTest_ConnectionError(t *testing.T) {
	// Create a server that's immediately closed to simulate connection error.
	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles: map[string]*config.Profile{
			"test": {URL: "http://localhost:1", APIKey: "test-key"},
		},
	}
	client := api.NewWithHTTPClient("http://localhost:1", "test-key", "1.1", &http.Client{})
	srv := NewServer("test", client, cfg, "test")

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{}

	result, err := srv.handleConfigTest(context.Background(), request)
	require.NoError(t, err) // MCP errors are returned as tool results, not Go errors
	assert.True(t, result.IsError)
}

func TestHandleSchemaGet_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"StatusCode": 200,
			"Response": {
				"Data": {
					"SiteId": 100,
					"Title": "TestTable",
					"SiteSettings": {
						"ReferenceType": "Results",
						"Columns": [
							{"ColumnName": "ClassA", "LabelText": "Category", "Required": true, "ChoicesText": "A\nB\nC"},
							{"ColumnName": "NumA", "LabelText": "Amount"}
						]
					}
				}
			}
		}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"site_id": float64(100)}

	result, err := srv.handleSchemaGet(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "TestTable")
	assert.Contains(t, text, "ClassA")
	assert.Contains(t, text, "Category")
}

func TestHandleSchemaGet_MissingSiteID(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{}

	result, err := srv.handleSchemaGet(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleSiteCreate_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		assert.Equal(t, "TestSite", body["Title"])
		assert.Equal(t, "Results", body["ReferenceType"])
		assert.NotNil(t, body["SiteSettings"])

		fmt.Fprint(w, `{"StatusCode":200,"Id":999,"Message":"Success"}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"parent_id": float64(100),
		"title":     "TestSite",
	}

	result, err := srv.handleSiteCreate(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleSiteCreate_MissingTitle(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"parent_id": float64(100),
	}

	result, err := srv.handleSiteCreate(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)

	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "title is required")
}

func TestHandleRecordUpsert_SingleRecord(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Id":1,"Message":"Created"}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"json":    `{"Title":"Test Record"}`,
	}

	result, err := srv.handleRecordUpsert(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, `"upserted": 1`)
}

func TestHandleRecordUpsert_MissingJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
	}

	result, err := srv.handleRecordUpsert(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleBatchRun_NotImplemented(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"yaml": "steps:\n  - name: test",
	}

	result, err := srv.handleBatchRun(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "not yet fully implemented")
}

func TestHandleBatchRun_MissingParams(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{}

	result, err := srv.handleBatchRun(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "yaml or template")
}

func TestHandleRecordUpsert_WithKey(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Message":"Upserted"}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"json":    `[{"Title":"A","ClassA":"cat1"},{"Title":"B","ClassA":"cat2"}]`,
		"key":     "ClassA",
	}

	result, err := srv.handleRecordUpsert(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleRecordUpsert_InvalidJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"json":    "not valid json",
	}

	result, err := srv.handleRecordUpsert(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleSiteCreate_WithJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Issues", body["ReferenceType"])
		fmt.Fprint(w, `{"StatusCode":200,"Id":999}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"parent_id":      float64(100),
		"title":          "IssueTable",
		"reference_type": "Issues",
		"body":           "Test body",
		"json":           `{"SiteSettings":{"Columns":[{"ColumnName":"ClassA"}]}}`,
	}

	result, err := srv.handleSiteCreate(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleSiteCreate_InvalidJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"parent_id": float64(100),
		"title":     "Test",
		"json":      "bad json",
	}

	result, err := srv.handleSiteCreate(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleRecordImport_WithMapping(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Id":1,"Message":"Created"}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"csv":     "Name,Category\nAlice,A\nBob,B",
		"mapping": "columns:\n  Name: Title\n  Category: ClassA",
	}

	result, err := srv.handleRecordImport(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleRecordImport_MissingCSV(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
	}

	result, err := srv.handleRecordImport(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleRecordImport_InvalidMapping(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"csv":     "A,B\n1,2",
		"mapping": "{{invalid yaml",
	}

	result, err := srv.handleRecordImport(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleRecordImport_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Id":1,"Message":"Created"}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"csv":     "Title,ClassA\nRecord1,Cat1\nRecord2,Cat2",
	}

	result, err := srv.handleRecordImport(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, `"imported": 2`)
}
