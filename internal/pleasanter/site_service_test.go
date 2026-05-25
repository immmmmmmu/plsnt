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

func TestSiteGet_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"TotalCount":1}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	result, err := svc.Get(context.Background(), 1234)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if receivedPath != "/api/items/1234/getsite" {
		t.Errorf("expected path /api/items/1234/getsite, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestSiteCreate_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Id":5555,"Message":"Created"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	payload := map[string]any{
		"Title":    "New Site",
		"SiteType": "Results",
	}
	result, err := svc.Create(context.Background(), 100, payload)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if receivedPath != "/api/items/100/createsite" {
		t.Errorf("expected path /api/items/100/createsite, got %s", receivedPath)
	}
	if receivedBody["Title"] != "New Site" {
		t.Errorf("expected Title 'New Site', got %v", receivedBody["Title"])
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestSiteUpdate_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"Updated"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	payload := map[string]any{
		"Title": "Updated Site",
	}
	result, err := svc.Update(context.Background(), 5678, payload)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if receivedPath != "/api/items/5678/updatesite" {
		t.Errorf("expected path /api/items/5678/updatesite, got %s", receivedPath)
	}
	if receivedBody["Title"] != "Updated Site" {
		t.Errorf("expected Title 'Updated Site', got %v", receivedBody["Title"])
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestSiteDelete_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"Deleted"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	result, err := svc.Delete(context.Background(), 9999)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if receivedPath != "/api/items/9999/deletesite" {
		t.Errorf("expected path /api/items/9999/deletesite, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestSiteCopy_Success(t *testing.T) {
	var paths []string
	var createBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/items/1000/getsite":
			w.Write([]byte(`{
				"StatusCode": 200,
				"Response": {
					"Data": {
						"Title": "Original Site",
						"Body": "Description",
						"ReferenceType": "Results",
						"SiteSettings": {"EditorColumnHash": {}}
					}
				}
			}`))
		case "/api/items/2000/createsite":
			json.NewDecoder(r.Body).Decode(&createBody)
			w.Write([]byte(`{"StatusCode":200,"Id":3000,"Message":"Created"}`))
		}
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	result, err := svc.Copy(context.Background(), 1000, 2000, nil)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(paths))
	}
	if paths[0] != "/api/items/1000/getsite" {
		t.Errorf("expected first request to /api/items/1000/getsite, got %s", paths[0])
	}
	if paths[1] != "/api/items/2000/createsite" {
		t.Errorf("expected second request to /api/items/2000/createsite, got %s", paths[1])
	}

	if createBody["Title"] != "Original Site" {
		t.Errorf("expected Title 'Original Site', got %v", createBody["Title"])
	}
	if createBody["Body"] != "Description" {
		t.Errorf("expected Body 'Description', got %v", createBody["Body"])
	}
	if createBody["ReferenceType"] != "Results" {
		t.Errorf("expected ReferenceType 'Results', got %v", createBody["ReferenceType"])
	}
	if createBody["SiteSettings"] == nil {
		t.Error("expected SiteSettings to be forwarded")
	}

	if result["Id"] == nil {
		t.Error("expected Id in create response")
	}
}

func TestSiteCopy_WithOverrides(t *testing.T) {
	var createBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/items/1000/getsite":
			w.Write([]byte(`{
				"StatusCode": 200,
				"Response": {
					"Data": {
						"Title": "Original Site",
						"Body": "Original Body",
						"ReferenceType": "Results"
					}
				}
			}`))
		case "/api/items/2000/createsite":
			json.NewDecoder(r.Body).Decode(&createBody)
			w.Write([]byte(`{"StatusCode":200,"Id":3000,"Message":"Created"}`))
		}
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	overrides := map[string]any{
		"Title": "Copied Site",
	}
	_, err := svc.Copy(context.Background(), 1000, 2000, overrides)
	if err != nil {
		t.Fatalf("Copy with overrides failed: %v", err)
	}

	if createBody["Title"] != "Copied Site" {
		t.Errorf("expected Title 'Copied Site', got %v", createBody["Title"])
	}
	if createBody["Body"] != "Original Body" {
		t.Errorf("expected Body 'Original Body', got %v", createBody["Body"])
	}
}

func TestSiteCopy_GetSourceFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500,"Message":"Internal Server Error"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	_, err := svc.Copy(context.Background(), 1000, 2000, nil)
	if err == nil {
		t.Fatal("expected error when source get fails, got nil")
	}
}

func TestSiteCopy_NoResponseField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	_, err := svc.Copy(context.Background(), 1000, 2000, nil)
	if err == nil {
		t.Fatal("expected error when Response field is missing, got nil")
	}
}

func TestSiteCopy_UsesGetsiteNotGet(t *testing.T) {
	var paths []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/items/1000/getsite":
			w.Write([]byte(`{
				"StatusCode": 200,
				"Response": {
					"Data": {
						"Title": "Source",
						"ReferenceType": "Results"
					}
				}
			}`))
		case "/api/items/2000/createsite":
			w.Write([]byte(`{"StatusCode":200,"Id":3000}`))
		default:
			t.Errorf("unexpected path called: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	_, err := svc.Copy(context.Background(), 1000, 2000, nil)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Verify the first call uses /getsite, not /get
	if len(paths) < 1 {
		t.Fatal("expected at least 1 request")
	}
	if paths[0] != "/api/items/1000/getsite" {
		t.Errorf("expected first request to use /getsite endpoint, got %s", paths[0])
	}
	// Ensure /get was never called
	for _, p := range paths {
		if p == "/api/items/1000/get" {
			t.Error("Copy should use /getsite, not /get, for retrieving source site")
		}
	}
}

func TestSiteCopy_ExtractsFieldsFromResponseData(t *testing.T) {
	var createBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/items/1000/getsite":
			// Response.Data contains the site fields (nested under Response)
			w.Write([]byte(`{
				"StatusCode": 200,
				"Response": {
					"Data": {
						"Title": "Source Title",
						"Body": "Source Body",
						"ReferenceType": "Issues",
						"SiteSettings": {"EditorColumnHash": {"ClassA": true}},
						"SiteId": 1000,
						"CreatedTime": "2025-01-01"
					}
				}
			}`))
		case "/api/items/2000/createsite":
			json.NewDecoder(r.Body).Decode(&createBody)
			w.Write([]byte(`{"StatusCode":200,"Id":3000}`))
		}
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	_, err := svc.Copy(context.Background(), 1000, 2000, nil)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Should extract Title, Body, ReferenceType, SiteSettings from Response.Data
	if createBody["Title"] != "Source Title" {
		t.Errorf("expected Title 'Source Title', got %v", createBody["Title"])
	}
	if createBody["Body"] != "Source Body" {
		t.Errorf("expected Body 'Source Body', got %v", createBody["Body"])
	}
	if createBody["ReferenceType"] != "Issues" {
		t.Errorf("expected ReferenceType 'Issues', got %v", createBody["ReferenceType"])
	}
	if createBody["SiteSettings"] == nil {
		t.Error("expected SiteSettings to be extracted from Response.Data")
	}

	// SiteId and CreatedTime should NOT be copied (they are not in the allowed fields list)
	if createBody["SiteId"] != nil {
		t.Error("SiteId should not be copied to new site payload")
	}
	if createBody["CreatedTime"] != nil {
		t.Error("CreatedTime should not be copied to new site payload")
	}
}

func TestSiteCopy_NoResponseData_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Response exists but Data is missing
		w.Write([]byte(`{"StatusCode":200,"Response":{"TotalCount":0}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	_, err := svc.Copy(context.Background(), 1000, 2000, nil)
	if err == nil {
		t.Fatal("expected error when Response.Data is missing, got nil")
	}
}

func TestSiteSearch_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"TotalCount":2,"Data":[{"SiteId":101,"Title":"Test Site A"},{"SiteId":102,"Title":"Test Site B"}]}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	result, err := svc.Search(context.Background(), 500, "Test")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if receivedPath != "/api/items/500/get" {
		t.Errorf("expected path /api/items/500/get, got %s", receivedPath)
	}

	// Verify View.ColumnFilterHash.Title is sent in the request body
	view, ok := receivedBody["View"].(map[string]any)
	if !ok {
		t.Fatal("expected View in request body")
	}
	filterHash, ok := view["ColumnFilterHash"].(map[string]any)
	if !ok {
		t.Fatal("expected ColumnFilterHash in View")
	}
	if filterHash["Title"] != "Test" {
		t.Errorf("expected Title filter 'Test', got %v", filterHash["Title"])
	}

	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestSiteSearch_EmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"TotalCount":0,"Data":[]}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	result, err := svc.Search(context.Background(), 500, "nonexistent")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	resp, ok := result["Response"].(map[string]any)
	if !ok {
		t.Fatal("expected Response in result")
	}
	totalCount := fmt.Sprintf("%v", resp["TotalCount"])
	if totalCount != "0" {
		t.Errorf("expected TotalCount 0, got %v", totalCount)
	}
}

func TestSiteSearch_EndpointUsesParentSiteID(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"TotalCount":0}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSiteService(client)

	_, err := svc.Search(context.Background(), 99999, "keyword")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if receivedPath != "/api/items/99999/get" {
		t.Errorf("expected path /api/items/99999/get, got %s", receivedPath)
	}
}
