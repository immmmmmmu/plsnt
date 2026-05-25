package pleasanter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/api"
)

func TestDeptList_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"Data":[{"DeptId":1,"DeptName":"Sales"}]}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewDeptService(client)

	result, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if receivedPath != "/api/depts/get" {
		t.Errorf("expected path /api/depts/get, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestDeptGet_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"Data":[{"DeptId":10,"DeptName":"Engineering"}]}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewDeptService(client)

	result, err := svc.Get(context.Background(), 10)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if receivedPath != "/api/depts/10/get" {
		t.Errorf("expected path /api/depts/10/get, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestDeptCreate_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Id":10,"Message":"Created"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewDeptService(client)

	result, err := svc.Create(context.Background(), map[string]any{
		"DeptName": "New Department",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if receivedPath != "/api/depts/create" {
		t.Errorf("expected path /api/depts/create, got %s", receivedPath)
	}
	if receivedBody["DeptName"] != "New Department" {
		t.Errorf("expected DeptName 'New Department', got %v", receivedBody["DeptName"])
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestDeptUpdate_Success(t *testing.T) {
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
	svc := NewDeptService(client)

	result, err := svc.Update(context.Background(), 10, map[string]any{
		"DeptName": "Updated Department",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if receivedPath != "/api/depts/10/update" {
		t.Errorf("expected path /api/depts/10/update, got %s", receivedPath)
	}
	if receivedBody["DeptName"] != "Updated Department" {
		t.Errorf("expected DeptName 'Updated Department', got %v", receivedBody["DeptName"])
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestDeptList_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500,"Message":"Internal Server Error"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewDeptService(client)

	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestDeptGet_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewDeptService(client)

	_, err := svc.Get(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestDeptCreate_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewDeptService(client)

	_, err := svc.Create(context.Background(), map[string]any{"DeptName": "test"})
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestDeptUpdate_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewDeptService(client)

	_, err := svc.Update(context.Background(), 10, map[string]any{"DeptName": "test"})
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestDeptDelete_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewDeptService(client)

	_, err := svc.Delete(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestDeptCreate_POSTMethod(t *testing.T) {
	var receivedMethod string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Id":1}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewDeptService(client)

	svc.Create(context.Background(), map[string]any{"DeptName": "test"})
	if receivedMethod != "POST" {
		t.Errorf("expected POST method, got %s", receivedMethod)
	}
}

func TestDeptDelete_Success(t *testing.T) {
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
	svc := NewDeptService(client)

	result, err := svc.Delete(context.Background(), 10)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if receivedPath != "/api/depts/10/delete" {
		t.Errorf("expected path /api/depts/10/delete, got %s", receivedPath)
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
