package pleasanter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/api"
)

func TestUserList_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"Data":[{"UserId":1,"LoginId":"admin"},{"UserId":2,"LoginId":"user1"}]}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	result, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if receivedPath != "/api/users/get" {
		t.Errorf("expected path /api/users/get, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestUserGet_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Response":{"Data":[{"UserId":42,"LoginId":"testuser"}]}}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	result, err := svc.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if receivedPath != "/api/users/42/get" {
		t.Errorf("expected path /api/users/42/get, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestUserCreate_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Id":100,"Message":"Created"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	result, err := svc.Create(context.Background(), map[string]any{
		"LoginId":  "newuser",
		"Password": "password123",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if receivedPath != "/api/users/create" {
		t.Errorf("expected path /api/users/create, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
	if receivedBody["LoginId"] != "newuser" {
		t.Errorf("expected LoginId 'newuser', got %v", receivedBody["LoginId"])
	}
}

func TestUserUpdate_Success(t *testing.T) {
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
	svc := NewUserService(client)

	result, err := svc.Update(context.Background(), 42, map[string]any{
		"Name": "Updated Name",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if receivedPath != "/api/users/42/update" {
		t.Errorf("expected path /api/users/42/update, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
	if receivedBody["Name"] != "Updated Name" {
		t.Errorf("expected Name 'Updated Name', got %v", receivedBody["Name"])
	}
}

func TestUserList_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500,"Message":"Internal Server Error"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestUserGet_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	_, err := svc.Get(context.Background(), 42)
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestUserCreate_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	_, err := svc.Create(context.Background(), map[string]any{"LoginId": "test"})
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestUserUpdate_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	_, err := svc.Update(context.Background(), 42, map[string]any{"Name": "test"})
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestUserDelete_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"StatusCode":500}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	_, err := svc.Delete(context.Background(), 42)
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestUserCreate_POSTMethod(t *testing.T) {
	var receivedMethod string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Id":1}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	svc.Create(context.Background(), map[string]any{"LoginId": "test"})
	if receivedMethod != "POST" {
		t.Errorf("expected POST method, got %s", receivedMethod)
	}
}

func TestUserDelete_Success(t *testing.T) {
	var receivedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"Deleted"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	result, err := svc.Delete(context.Background(), 42)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if receivedPath != "/api/users/42/delete" {
		t.Errorf("expected path /api/users/42/delete, got %s", receivedPath)
	}
	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}
