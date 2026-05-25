package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestPost_InjectsAuthFields(t *testing.T) {
	var received map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &received)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"StatusCode": 200})
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "test-api-key", "1.1", srv.Client())

	type req struct {
		Title string `json:"Title"`
	}
	var resp map[string]any

	err := client.Post(context.Background(), "/api/items/100/get", &req{Title: "hello"}, &resp)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}

	if received["ApiKey"] != "test-api-key" {
		t.Errorf("expected ApiKey 'test-api-key', got %v", received["ApiKey"])
	}
	if received["ApiVersion"] != "1.1" {
		t.Errorf("expected ApiVersion '1.1', got %v", received["ApiVersion"])
	}
	if received["Title"] != "hello" {
		t.Errorf("expected Title 'hello', got %v", received["Title"])
	}
}

func TestPost_DecodesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"StatusCode": 200,
			"Response": map[string]any{
				"TotalCount": 5,
			},
		})
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())

	type apiResp struct {
		StatusCode int `json:"StatusCode"`
		Response   struct {
			TotalCount int `json:"TotalCount"`
		} `json:"Response"`
	}

	var resp apiResp
	err := client.Post(context.Background(), "/api/items/100/get", &struct{}{}, &resp)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if resp.Response.TotalCount != 5 {
		t.Errorf("expected TotalCount 5, got %d", resp.Response.TotalCount)
	}
}

func TestPostRaw_InjectsAuthAndReturnsMap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var received map[string]any
		json.NewDecoder(r.Body).Decode(&received)

		if received["ApiKey"] != "raw-key" {
			t.Errorf("expected ApiKey 'raw-key', got %v", received["ApiKey"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"StatusCode": 200,
			"Response":   map[string]any{"Data": []any{}},
		})
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "raw-key", "1.1", srv.Client())

	payload := map[string]any{"Title": "test"}
	result, err := client.PostRaw(context.Background(), "/api/items/100/create", payload)
	if err != nil {
		t.Fatalf("PostRaw failed: %v", err)
	}

	if result["StatusCode"] == nil {
		t.Error("expected StatusCode in response")
	}
}

func TestPost_HTTP401_ReturnsAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "bad-key", "1.1", srv.Client())

	var resp map[string]any
	err := client.Post(context.Background(), "/api/items/100/get", &struct{}{}, &resp)
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if got := err.Error(); got == "" {
		t.Error("error message should not be empty")
	}
}

func TestPost_HTTP500_ReturnsServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())

	var resp map[string]any
	err := client.Post(context.Background(), "/api/items/100/get", &struct{}{}, &resp)
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestPost_ContentTypeIsJSON(t *testing.T) {
	var contentType string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"StatusCode": 200})
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())

	var resp map[string]any
	client.Post(context.Background(), "/test", &struct{}{}, &resp)

	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}
}

func TestPost_AllMethodsArePOST(t *testing.T) {
	var method string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"StatusCode": 200})
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())

	var resp map[string]any
	client.Post(context.Background(), "/test", &struct{}{}, &resp)

	if method != http.MethodPost {
		t.Errorf("expected POST, got %s", method)
	}
}

func TestPost_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var resp map[string]any
	err := client.Post(ctx, "/test", &struct{}{}, &resp)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestNew_WithRetryDisabled(t *testing.T) {
	var callCount int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := New(srv.URL, "key", "1.1", WithRetryDisabled())

	var resp map[string]any
	client.Post(context.Background(), "/test", &struct{}{}, &resp)

	if count := atomic.LoadInt32(&callCount); count != 1 {
		t.Errorf("expected 1 call with retry disabled, got %d", count)
	}
}

func TestPostRaw_PreservesJsonNumber(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Value":123.456}`))
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())

	result, err := client.PostRaw(context.Background(), "/test", map[string]any{})
	if err != nil {
		t.Fatalf("PostRaw failed: %v", err)
	}

	val := result["Value"]
	// json.Number preserves the exact decimal representation
	num, ok := val.(json.Number)
	if !ok {
		t.Fatalf("expected json.Number, got %T", val)
	}
	if num.String() != "123.456" {
		t.Errorf("expected 123.456, got %s", num.String())
	}
}

func TestNew_WithInsecure(t *testing.T) {
	// Start a TLS server with a self-signed certificate.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"StatusCode": 200})
	}))
	defer srv.Close()

	// Without insecure, the self-signed cert should cause a failure.
	clientNoInsecure := New(srv.URL, "key", "1.1", WithRetryDisabled())
	var resp map[string]any
	err := clientNoInsecure.Post(context.Background(), "/test", &struct{}{}, &resp)
	if err == nil {
		t.Fatal("expected TLS error when connecting to self-signed cert without insecure")
	}

	// With insecure, the connection should succeed.
	clientInsecure := New(srv.URL, "key", "1.1", WithRetryDisabled(), WithInsecure())
	var resp2 map[string]any
	err = clientInsecure.Post(context.Background(), "/test", &struct{}{}, &resp2)
	if err != nil {
		t.Fatalf("expected no error with insecure, got: %v", err)
	}
}

func TestNew_WithInsecure_FieldIsSet(t *testing.T) {
	// Verify the option sets the insecure field on clientImpl.
	c := &clientImpl{}
	opt := WithInsecure()
	opt(c)
	if !c.insecure {
		t.Error("expected insecure to be true after applying WithInsecure option")
	}
}

func TestDoPost_302EmptyBody_ReturnsSyntheticJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound) // 302
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "key", "1.1", &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	})

	result, err := client.PostRaw(context.Background(), "/api/items/100/createsite", map[string]any{})
	if err != nil {
		t.Fatalf("expected no error for 302, got: %v", err)
	}

	statusCode, ok := result["StatusCode"]
	if !ok {
		t.Fatal("expected StatusCode in synthetic response")
	}
	if fmt.Sprintf("%v", statusCode) != "302" {
		t.Errorf("expected StatusCode 302, got %v", statusCode)
	}
	if result["Message"] == nil {
		t.Error("expected Message in synthetic response")
	}
}

func TestDoPost_302WithJSONBody_ReturnsJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound) // 302
		w.Write([]byte(`{"Id":5555,"Message":"Created"}`))
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "key", "1.1", &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	})

	result, err := client.PostRaw(context.Background(), "/api/items/100/createsite", map[string]any{})
	if err != nil {
		t.Fatalf("expected no error for 302 with JSON body, got: %v", err)
	}

	if fmt.Sprintf("%v", result["Id"]) != "5555" {
		t.Errorf("expected Id 5555, got %v", result["Id"])
	}
	if result["Message"] != "Created" {
		t.Errorf("expected Message 'Created', got %v", result["Message"])
	}
}

func TestDoPost_302WithNonJSONBody_ReturnsSyntheticJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound) // 302
		w.Write([]byte(`<html><body>Redirect</body></html>`))
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "key", "1.1", &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	})

	result, err := client.PostRaw(context.Background(), "/api/items/100/createsite", map[string]any{})
	if err != nil {
		t.Fatalf("expected no error for 302 with HTML body, got: %v", err)
	}

	statusCode, ok := result["StatusCode"]
	if !ok {
		t.Fatal("expected StatusCode in synthetic response")
	}
	if fmt.Sprintf("%v", statusCode) != "302" {
		t.Errorf("expected StatusCode 302, got %v", statusCode)
	}
}

func TestNoRetry_ReturnsWorkingClient(t *testing.T) {
	var callCount int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"StatusCode": 200, "Message": "OK"})
	}))
	defer srv.Close()

	client := NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	noRetryClient := client.NoRetry()

	result, err := noRetryClient.PostRaw(context.Background(), "/api/items/100/get", map[string]any{})
	if err != nil {
		t.Fatalf("NoRetry client PostRaw failed: %v", err)
	}
	if result["Message"] != "OK" {
		t.Errorf("expected Message 'OK', got %v", result["Message"])
	}
	if count := atomic.LoadInt32(&callCount); count != 1 {
		t.Errorf("expected exactly 1 call, got %d", count)
	}
}

func TestPost_BaseURLTrailingSlash(t *testing.T) {
	var requestURL string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURL = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"StatusCode": 200})
	}))
	defer srv.Close()

	// URL with trailing slash
	client := NewWithHTTPClient(srv.URL+"/", "key", "1.1", srv.Client())

	var resp map[string]any
	client.Post(context.Background(), "/api/items/100/get", &struct{}{}, &resp)

	if requestURL != "/api/items/100/get" {
		t.Errorf("expected /api/items/100/get, got %q", requestURL)
	}
}
