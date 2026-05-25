package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- record_list ---

func TestHandleRecordList_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Response":{"Offset":0,"PageSize":200,"TotalCount":2,"Data":[{"Title":"A"},{"Title":"B"}]}}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"site_id": float64(100)}

	result, err := srv.handleRecordList(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "TotalCount")
}

func TestHandleRecordList_WithView(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		// Verify View was included in the request
		assert.NotNil(t, body["View"])
		fmt.Fprint(w, `{"StatusCode":200,"Response":{"Offset":0,"PageSize":200,"TotalCount":1,"Data":[{"Title":"Filtered"}]}}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"view":    `{"ColumnFilterHash":{"ClassA":"Red"}}`,
	}

	result, err := srv.handleRecordList(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleRecordList_WithJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Response":{"Data":[]}}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"json":    `{"Offset":10}`,
	}

	result, err := srv.handleRecordList(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleRecordList_InvalidView(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"view":    "bad json",
	}

	result, err := srv.handleRecordList(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// --- record_get ---

func TestHandleRecordGet_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Response":{"Data":[{"Title":"Found"}],"TotalCount":1}}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"record_id": float64(999)}

	result, err := srv.handleRecordGet(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleRecordGet_WithJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Response":{"Data":[{"Title":"Raw"}]}}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"record_id": float64(999),
		"json":      `{}`,
	}

	result, err := srv.handleRecordGet(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

// --- record_create ---

func TestHandleRecordCreate_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "New Record", body["Title"])
		fmt.Fprint(w, `{"StatusCode":200,"Id":1001}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"title":   "New Record",
		"body":    "Description",
		"status":  float64(100),
	}

	result, err := srv.handleRecordCreate(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleRecordCreate_WithJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		classHash := body["ClassHash"].(map[string]any)
		assert.Equal(t, "Cat1", classHash["ClassA"])
		fmt.Fprint(w, `{"StatusCode":200,"Id":1002}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"title":   "With Custom",
		"json":    `{"ClassHash":{"ClassA":"Cat1"}}`,
	}

	result, err := srv.handleRecordCreate(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

// --- record_update ---

func TestHandleRecordUpdate_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Updated", body["Title"])
		fmt.Fprint(w, `{"StatusCode":200}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"record_id": float64(999),
		"title":     "Updated",
		"status":    float64(200),
	}

	result, err := srv.handleRecordUpdate(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleRecordUpdate_WithJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"record_id": float64(999),
		"json":      `{"NumHash":{"NumA":42}}`,
	}

	result, err := srv.handleRecordUpdate(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

// --- record_delete ---

func TestHandleRecordDelete_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Message":"Deleted"}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"record_id": float64(999)}

	result, err := srv.handleRecordDelete(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

// --- record_bulk_delete ---

func TestHandleRecordBulkDelete_ByIDs(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Message":"Deleted 3"}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id":    float64(100),
		"record_ids": "1,2,3",
	}

	result, err := srv.handleRecordBulkDelete(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleRecordBulkDelete_ByView(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Message":"Deleted by view"}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"view":    `{"ColumnFilterHash":{"Status":"900"}}`,
	}

	result, err := srv.handleRecordBulkDelete(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleRecordBulkDelete_NoParams(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"site_id": float64(100)}

	result, err := srv.handleRecordBulkDelete(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// --- site_get ---

func TestHandleSiteGet_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Response":{"Data":{"SiteId":100,"Title":"MyTable"}}}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"site_id": float64(100)}

	result, err := srv.handleSiteGet(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "MyTable")
}

// --- site_update ---

func TestHandleSiteUpdate_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"site_id": float64(100),
		"json":    `{"Title":"Updated Site"}`,
	}

	result, err := srv.handleSiteUpdate(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHandleSiteUpdate_MissingJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"site_id": float64(100)}

	result, err := srv.handleSiteUpdate(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// --- site_copy ---

func TestHandleSiteCopy_Success(t *testing.T) {
	callCount := 0
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// getsite response
			fmt.Fprint(w, `{"Response":{"Data":{"Title":"Source","ReferenceType":"Results","SiteSettings":{"ReferenceType":"Results"}}}}`)
		} else {
			// createsite response
			fmt.Fprint(w, `{"StatusCode":200,"Id":200}`)
		}
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"source_site_id": float64(100),
		"parent_id":      float64(50),
		"title":          "Copy of Source",
	}

	result, err := srv.handleSiteCopy(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

// --- site_search ---

func TestHandleSiteSearch_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"StatusCode":200,"Response":{"Data":[{"Title":"Found Site"}]}}`)
	})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"parent_id": float64(50),
		"keyword":   "Found",
	}

	result, err := srv.handleSiteSearch(context.Background(), request)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Found Site")
}

func TestHandleSiteSearch_MissingKeyword(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"parent_id": float64(50)}

	result, err := srv.handleSiteSearch(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
