package pleasanter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/api"
)

func TestGetSchema_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SiteResponse{
			StatusCode: 200,
			Response: SiteResponseBody{
				Data: SiteData{
					SiteId: 100,
					Title:  "Test Table",
					SiteSettings: SiteSettings{
						ReferenceType: "Results",
						Columns: []ColumnDef{
							{ColumnName: "ClassA", LabelText: "Category", ChoicesText: "Red\nBlue\nGreen"},
							{ColumnName: "NumA", LabelText: "Amount", Required: true},
							{ColumnName: "DateA", LabelText: "Due Date"},
							{ColumnName: "DescriptionA", LabelText: "Notes"},
							{ColumnName: "CheckA", LabelText: "Approved"},
							{ColumnName: "Title", LabelText: "Title"},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSchemaService(client)

	info, err := svc.GetSchema(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	if info.SiteID != 100 {
		t.Errorf("expected SiteID 100, got %d", info.SiteID)
	}
	if info.Title != "Test Table" {
		t.Errorf("expected title 'Test Table', got %q", info.Title)
	}
	if info.TableType != "Results" {
		t.Errorf("expected TableType 'Results', got %q", info.TableType)
	}
	if len(info.Columns) != 6 {
		t.Fatalf("expected 6 columns, got %d", len(info.Columns))
	}

	// Check ClassA
	if info.Columns[0].Type != "Class" {
		t.Errorf("expected ClassA type 'Class', got %q", info.Columns[0].Type)
	}
	if len(info.Columns[0].Choices) != 3 {
		t.Errorf("expected 3 choices, got %d", len(info.Columns[0].Choices))
	}

	// Check NumA
	if info.Columns[1].Type != "Num" {
		t.Errorf("expected NumA type 'Num', got %q", info.Columns[1].Type)
	}
	if !info.Columns[1].Required {
		t.Error("expected NumA to be required")
	}

	// Check standard field
	if info.Columns[5].Type != "Standard" {
		t.Errorf("expected Title type 'Standard', got %q", info.Columns[5].Type)
	}
}

func TestGetSchema_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SiteResponse{StatusCode: 404}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSchemaService(client)

	_, err := svc.GetSchema(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found site")
	}
}

func TestGetSchema_EmptyData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SiteResponse{
			StatusCode: 200,
			Response:   SiteResponseBody{Data: SiteData{}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewSchemaService(client)

	_, err := svc.GetSchema(context.Background(), 100)
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestGetSchema_APIKeyInjected(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)

		resp := SiteResponse{
			StatusCode: 200,
			Response: SiteResponseBody{
				Data: SiteData{SiteId: 1, SiteSettings: SiteSettings{ReferenceType: "Results"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "my-secret-key", "1.1", srv.Client())
	svc := NewSchemaService(client)

	svc.GetSchema(context.Background(), 1)

	if receivedBody["ApiKey"] != "my-secret-key" {
		t.Errorf("expected ApiKey 'my-secret-key', got %v", receivedBody["ApiKey"])
	}
}

func TestClassifyColumnType(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"ClassA", "Class"},
		{"ClassZ", "Class"},
		{"NumA", "Num"},
		{"DateA", "Date"},
		{"DescriptionA", "Description"},
		{"CheckA", "Check"},
		{"AttachmentsA", "Attachments"},
		{"Title", "Standard"},
		{"Status", "Standard"},
	}
	for _, tt := range tests {
		got := classifyColumnType(tt.name)
		if got != tt.expected {
			t.Errorf("classifyColumnType(%q) = %q, want %q", tt.name, got, tt.expected)
		}
	}
}

func TestParseChoices(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"Red\nBlue\nGreen", 3},
		{"Single", 1},
		{"A\n\nB\n", 2},
		{"", 0},
	}
	for _, tt := range tests {
		got := parseChoices(tt.input)
		if len(got) != tt.expected {
			t.Errorf("parseChoices(%q) returned %d items, want %d", tt.input, len(got), tt.expected)
		}
	}
}
