package pleasanter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/api"
)

func TestRecordGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/items/5678/get" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		resp := APIResponse{
			StatusCode: 200,
			Response: ResponseBody{
				TotalCount: 1,
				Data: []Record{
					{ResultId: 5678, SiteId: 100, Title: "Test Record"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	resp, err := svc.Get(context.Background(), 5678)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected StatusCode 200, got %d", resp.StatusCode)
	}
	if len(resp.Response.Data) != 1 {
		t.Fatalf("expected 1 record, got %d", len(resp.Response.Data))
	}
	if resp.Response.Data[0].Title != "Test Record" {
		t.Errorf("expected 'Test Record', got %q", resp.Response.Data[0].Title)
	}
}

func TestRecordGet_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{StatusCode: 404}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	_, err := svc.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found record")
	}
}

func TestRecordList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/items/100/get" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		resp := APIResponse{
			StatusCode: 200,
			Response: ResponseBody{
				Offset:     0,
				PageSize:   200,
				TotalCount: 2,
				Data: []Record{
					{ResultId: 1, SiteId: 100, Title: "Record A"},
					{ResultId: 2, SiteId: 100, Title: "Record B"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	resp, err := svc.List(context.Background(), ListOptions{SiteID: 100})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if resp.Response.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", resp.Response.TotalCount)
	}
	if len(resp.Response.Data) != 2 {
		t.Errorf("expected 2 records, got %d", len(resp.Response.Data))
	}
}

func TestRecordList_WithOffset(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		resp := APIResponse{StatusCode: 200, Response: ResponseBody{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	svc.List(context.Background(), ListOptions{SiteID: 100, Offset: 200})

	if receivedBody["Offset"] == nil {
		t.Error("expected Offset in request body")
	}
}

func TestRecordList_WithView(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		resp := APIResponse{StatusCode: 200, Response: ResponseBody{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	view := &View{
		ColumnFilterHash: map[string]string{"ClassA": "Red"},
	}
	svc.List(context.Background(), ListOptions{SiteID: 100, View: view})

	if receivedBody["View"] == nil {
		t.Error("expected View in request body")
	}
}

func TestRecordCreate_Success(t *testing.T) {
	var receivedPath string
	var receivedMethod string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Id":9999,"Message":"Created"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	result, err := svc.Create(context.Background(), 100, map[string]any{
		"Title": "New Record",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if receivedPath != "/api/items/100/create" {
		t.Errorf("expected path /api/items/100/create, got %s", receivedPath)
	}
	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST method, got %s", receivedMethod)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestRecordCreate_PayloadPassthrough(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	payload := map[string]any{
		"Title": "Passthrough Title",
		"ClassHash": map[string]any{
			"ClassA": "ValueA",
			"ClassB": "ValueB",
		},
	}
	_, err := svc.Create(context.Background(), 200, payload)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if receivedBody["Title"] != "Passthrough Title" {
		t.Errorf("expected Title 'Passthrough Title', got %v", receivedBody["Title"])
	}
	classHash, ok := receivedBody["ClassHash"].(map[string]any)
	if !ok {
		t.Fatal("expected ClassHash to be a map")
	}
	if classHash["ClassA"] != "ValueA" {
		t.Errorf("expected ClassA 'ValueA', got %v", classHash["ClassA"])
	}
}

func TestRecordUpdate_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"Updated"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	result, err := svc.Update(context.Background(), 5678, map[string]any{
		"Title": "Updated Title",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if receivedPath != "/api/items/5678/update" {
		t.Errorf("expected path /api/items/5678/update, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestRecordUpdate_PayloadPassthrough(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	payload := map[string]any{
		"Title": "Updated Title",
		"ClassHash": map[string]any{
			"ClassA": "NewValue",
		},
	}
	_, err := svc.Update(context.Background(), 5678, payload)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if receivedBody["Title"] != "Updated Title" {
		t.Errorf("expected Title 'Updated Title', got %v", receivedBody["Title"])
	}
	classHash, ok := receivedBody["ClassHash"].(map[string]any)
	if !ok {
		t.Fatal("expected ClassHash to be a map")
	}
	if classHash["ClassA"] != "NewValue" {
		t.Errorf("expected ClassA 'NewValue', got %v", classHash["ClassA"])
	}
}

func TestRecordDelete_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"Deleted"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	result, err := svc.Delete(context.Background(), 9999)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if receivedPath != "/api/items/9999/delete" {
		t.Errorf("expected path /api/items/9999/delete, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestRecordDelete_EmptyPayload(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	_, err := svc.Delete(context.Background(), 9999)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// The only key should be ApiKey (injected by client); no user payload fields
	for key := range receivedBody {
		if key != "ApiKey" && key != "ApiVersion" {
			t.Errorf("unexpected key in delete payload: %s", key)
		}
	}
}

func TestListAll_MultiplePages(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		offset := 0
		if v, ok := body["Offset"]; ok {
			offset = int(v.(float64))
		}

		pageSize := 2
		totalCount := 5

		var data []Record
		for i := offset; i < offset+pageSize && i < totalCount; i++ {
			data = append(data, Record{
				ResultId: int64(i + 1),
				SiteId:   100,
				Title:    fmt.Sprintf("Record %d", i+1),
			})
		}

		resp := APIResponse{
			StatusCode: 200,
			Response: ResponseBody{
				Offset:     offset,
				PageSize:   pageSize,
				TotalCount: totalCount,
				Data:       data,
			},
		}
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	resp, err := svc.ListAll(context.Background(), ListOptions{SiteID: 100})
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(resp.Response.Data) != 5 {
		t.Errorf("expected 5 records, got %d", len(resp.Response.Data))
	}
	if resp.Response.TotalCount != 5 {
		t.Errorf("expected TotalCount 5, got %d", resp.Response.TotalCount)
	}
	// 3 pages: [0,1], [2,3], [4]
	if callCount != 3 {
		t.Errorf("expected 3 API calls, got %d", callCount)
	}
}

func TestListAll_SinglePage(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := APIResponse{
			StatusCode: 200,
			Response: ResponseBody{
				Offset:     0,
				PageSize:   200,
				TotalCount: 3,
				Data: []Record{
					{ResultId: 1, SiteId: 100, Title: "A"},
					{ResultId: 2, SiteId: 100, Title: "B"},
					{ResultId: 3, SiteId: 100, Title: "C"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	resp, err := svc.ListAll(context.Background(), ListOptions{SiteID: 100})
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(resp.Response.Data) != 3 {
		t.Errorf("expected 3 records, got %d", len(resp.Response.Data))
	}
	if callCount != 1 {
		t.Errorf("expected 1 API call (single page), got %d", callCount)
	}
}

func TestListAll_SafetyLimit(t *testing.T) {
	// Server always returns 200 records and claims TotalCount is 20000 (exceeds safety limit)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		offset := 0
		if v, ok := body["Offset"]; ok {
			offset = int(v.(float64))
		}

		pageSize := 200
		data := make([]Record, pageSize)
		for i := 0; i < pageSize; i++ {
			data[i] = Record{
				ResultId: int64(offset + i + 1),
				SiteId:   100,
				Title:    fmt.Sprintf("Record %d", offset+i+1),
			}
		}

		resp := APIResponse{
			StatusCode: 200,
			Response: ResponseBody{
				Offset:     offset,
				PageSize:   pageSize,
				TotalCount: 20000,
				Data:       data,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	resp, err := svc.ListAll(context.Background(), ListOptions{SiteID: 100})
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	// Safety limit is 10000
	if len(resp.Response.Data) != 10000 {
		t.Errorf("expected safety limit of 10000 records, got %d", len(resp.Response.Data))
	}
}

func TestBulkUpsert_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"BulkUpserted"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	keys := []string{"ClassA"}
	data := []map[string]any{
		{"Title": "Item 1", "ClassHash": map[string]any{"ClassA": "key1"}, "NumHash": map[string]any{"NumA": 100}},
		{"Title": "Item 2", "ClassHash": map[string]any{"ClassA": "key2"}, "NumHash": map[string]any{"NumA": 200}},
	}

	result, err := svc.BulkUpsert(context.Background(), 100, keys, data)
	if err != nil {
		t.Fatalf("BulkUpsert failed: %v", err)
	}

	if receivedPath != "/api/items/100/bulkupsert" {
		t.Errorf("expected path /api/items/100/bulkupsert, got %s", receivedPath)
	}

	// Verify Keys in payload
	rawKeys, ok := receivedBody["Keys"].([]any)
	if !ok {
		t.Fatal("expected Keys to be an array")
	}
	if len(rawKeys) != 1 || rawKeys[0] != "ClassA" {
		t.Errorf("expected Keys [ClassA], got %v", rawKeys)
	}

	// Verify Data in payload
	rawData, ok := receivedBody["Data"].([]any)
	if !ok {
		t.Fatal("expected Data to be an array")
	}
	if len(rawData) != 2 {
		t.Errorf("expected 2 data items, got %d", len(rawData))
	}

	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
	if result["Message"] != "BulkUpserted" {
		t.Errorf("expected Message 'BulkUpserted', got %v", result["Message"])
	}
}

func TestBulkUpsert_MultipleKeys(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	keys := []string{"ClassA", "ClassB"}
	data := []map[string]any{
		{"Title": "Item 1", "ClassHash": map[string]any{"ClassA": "a1", "ClassB": "b1"}},
	}

	_, err := svc.BulkUpsert(context.Background(), 200, keys, data)
	if err != nil {
		t.Fatalf("BulkUpsert failed: %v", err)
	}

	rawKeys, ok := receivedBody["Keys"].([]any)
	if !ok {
		t.Fatal("expected Keys to be an array")
	}
	if len(rawKeys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(rawKeys))
	}
	if rawKeys[0] != "ClassA" || rawKeys[1] != "ClassB" {
		t.Errorf("expected Keys [ClassA, ClassB], got %v", rawKeys)
	}
}

func TestBulkUpsert_PayloadContainsApiKey(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "test-api-key", "1.1", srv.Client())
	svc := NewRecordService(client)

	_, err := svc.BulkUpsert(context.Background(), 100, []string{"ClassA"}, []map[string]any{
		{"Title": "Test"},
	})
	if err != nil {
		t.Fatalf("BulkUpsert failed: %v", err)
	}

	// PostRaw injects ApiKey into the payload
	if receivedBody["ApiKey"] != "test-api-key" {
		t.Errorf("expected ApiKey 'test-api-key', got %v", receivedBody["ApiKey"])
	}
}

func TestBulkUpsert_EmptyData(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	_, err := svc.BulkUpsert(context.Background(), 100, []string{"ClassA"}, []map[string]any{})
	if err != nil {
		t.Fatalf("BulkUpsert failed: %v", err)
	}

	rawData, ok := receivedBody["Data"].([]any)
	if !ok {
		t.Fatal("expected Data to be an array")
	}
	if len(rawData) != 0 {
		t.Errorf("expected empty Data array, got %d items", len(rawData))
	}
}

func TestBulkDelete_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"BulkDeleted"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	recordIDs := []int64{101, 102, 103}
	result, err := svc.BulkDelete(context.Background(), 100, recordIDs)
	if err != nil {
		t.Fatalf("BulkDelete failed: %v", err)
	}

	if receivedPath != "/api/items/100/bulkdelete" {
		t.Errorf("expected path /api/items/100/bulkdelete, got %s", receivedPath)
	}

	rawIDs, ok := receivedBody["SelectedIds"].([]any)
	if !ok {
		t.Fatal("expected SelectedIds to be an array")
	}
	if len(rawIDs) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(rawIDs))
	}
	expectedIDs := []float64{101, 102, 103}
	for i, expected := range expectedIDs {
		if rawIDs[i].(float64) != expected {
			t.Errorf("expected ID %v at index %d, got %v", expected, i, rawIDs[i])
		}
	}

	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
	if result["Message"] != "BulkDeleted" {
		t.Errorf("expected Message 'BulkDeleted', got %v", result["Message"])
	}
}

func TestBulkDelete_PayloadContainsApiKey(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "test-api-key", "1.1", srv.Client())
	svc := NewRecordService(client)

	_, err := svc.BulkDelete(context.Background(), 100, []int64{101})
	if err != nil {
		t.Fatalf("BulkDelete failed: %v", err)
	}

	if receivedBody["ApiKey"] != "test-api-key" {
		t.Errorf("expected ApiKey 'test-api-key', got %v", receivedBody["ApiKey"])
	}
}

func TestBulkDeleteByView_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"BulkDeleted"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	view := map[string]any{
		"ColumnFilterHash": map[string]any{
			"ClassA": "to-delete",
		},
	}
	result, err := svc.BulkDeleteByView(context.Background(), 100, view)
	if err != nil {
		t.Fatalf("BulkDeleteByView failed: %v", err)
	}

	if receivedPath != "/api/items/100/bulkdelete" {
		t.Errorf("expected path /api/items/100/bulkdelete, got %s", receivedPath)
	}

	rawView, ok := receivedBody["View"].(map[string]any)
	if !ok {
		t.Fatal("expected View to be a map")
	}
	filterHash, ok := rawView["ColumnFilterHash"].(map[string]any)
	if !ok {
		t.Fatal("expected ColumnFilterHash to be a map")
	}
	if filterHash["ClassA"] != "to-delete" {
		t.Errorf("expected ClassA 'to-delete', got %v", filterHash["ClassA"])
	}

	if result["Message"] != "BulkDeleted" {
		t.Errorf("expected Message 'BulkDeleted', got %v", result["Message"])
	}
}

func TestBulkDeleteByView_PayloadContainsApiKey(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "test-api-key", "1.1", srv.Client())
	svc := NewRecordService(client)

	view := map[string]any{
		"ColumnFilterHash": map[string]any{"ClassA": "old"},
	}
	_, err := svc.BulkDeleteByView(context.Background(), 200, view)
	if err != nil {
		t.Fatalf("BulkDeleteByView failed: %v", err)
	}

	if receivedBody["ApiKey"] != "test-api-key" {
		t.Errorf("expected ApiKey 'test-api-key', got %v", receivedBody["ApiKey"])
	}
}

func TestBulkDelete_EmptyIDs(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	_, err := svc.BulkDelete(context.Background(), 100, []int64{})
	if err != nil {
		t.Fatalf("BulkDelete with empty IDs failed: %v", err)
	}

	rawIDs, ok := receivedBody["SelectedIds"].([]any)
	if !ok {
		t.Fatal("expected SelectedIds to be an array")
	}
	if len(rawIDs) != 0 {
		t.Errorf("expected empty SelectedIds array, got %d items", len(rawIDs))
	}
}

func TestRecordListRaw_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"TotalCount":1,"Data":[{"Title":"raw"}]}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	result, err := svc.ListRaw(context.Background(), 100, map[string]any{})
	if err != nil {
		t.Fatalf("ListRaw failed: %v", err)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}
