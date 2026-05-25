package pleasanter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/api"
)

func TestGroupList_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"Data":[{"GroupId":1,"GroupName":"Admin"},{"GroupId":2,"GroupName":"Dev"}]}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewGroupService(client)

	result, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if receivedPath != "/api/groups/get" {
		t.Errorf("expected path /api/groups/get, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestGroupGet_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"Data":{"GroupId":10,"GroupName":"Testers"}}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewGroupService(client)

	result, err := svc.Get(context.Background(), 10)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if receivedPath != "/api/groups/10/get" {
		t.Errorf("expected path /api/groups/10/get, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestGroupCreate_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Id":99,"Message":"Created"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewGroupService(client)

	payload := map[string]any{
		"GroupName": "NewGroup",
	}
	result, err := svc.Create(context.Background(), payload)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if receivedPath != "/api/groups/create" {
		t.Errorf("expected path /api/groups/create, got %s", receivedPath)
	}
	if receivedBody["GroupName"] != "NewGroup" {
		t.Errorf("expected GroupName 'NewGroup', got %v", receivedBody["GroupName"])
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestGroupUpdate_Success(t *testing.T) {
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
	svc := NewGroupService(client)

	payload := map[string]any{
		"GroupName": "UpdatedGroup",
	}
	result, err := svc.Update(context.Background(), 10, payload)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if receivedPath != "/api/groups/10/update" {
		t.Errorf("expected path /api/groups/10/update, got %s", receivedPath)
	}
	if receivedBody["GroupName"] != "UpdatedGroup" {
		t.Errorf("expected GroupName 'UpdatedGroup', got %v", receivedBody["GroupName"])
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestGroupList_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500,"Message":"Internal Server Error"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewGroupService(client)

	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestGroupGet_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewGroupService(client)

	_, err := svc.Get(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestGroupCreate_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewGroupService(client)

	_, err := svc.Create(context.Background(), map[string]any{"GroupName": "test"})
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestGroupUpdate_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewGroupService(client)

	_, err := svc.Update(context.Background(), 10, map[string]any{"GroupName": "test"})
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestGroupDelete_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewGroupService(client)

	_, err := svc.Delete(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestGroupCreate_POSTMethod(t *testing.T) {
	var receivedMethod string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Id":1}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewGroupService(client)

	svc.Create(context.Background(), map[string]any{"GroupName": "test"})
	if receivedMethod != "POST" {
		t.Errorf("expected POST method, got %s", receivedMethod)
	}
}

func TestGroupDelete_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"Deleted"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewGroupService(client)

	result, err := svc.Delete(context.Background(), 55)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if receivedPath != "/api/groups/55/delete" {
		t.Errorf("expected path /api/groups/55/delete, got %s", receivedPath)
	}
	// The only keys should be ApiKey/ApiVersion (injected by client); no user payload fields
	for key := range receivedBody {
		if key != "ApiKey" && key != "ApiVersion" {
			t.Errorf("unexpected key in delete payload: %s", key)
		}
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}
