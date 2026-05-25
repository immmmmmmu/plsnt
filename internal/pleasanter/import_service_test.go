package pleasanter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/api"
)

func TestLoadMapping_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mapping.yaml")
	content := `columns:
  "Name": "ClassA"
  "Amount": "NumA"
  "Title Column": "Title"
defaults:
  ClassHash:
    ClassB: "imported"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadMapping(path)
	if err != nil {
		t.Fatalf("LoadMapping failed: %v", err)
	}

	if len(cfg.Columns) != 3 {
		t.Errorf("expected 3 column mappings, got %d", len(cfg.Columns))
	}
	if cfg.Columns["Name"] != "ClassA" {
		t.Errorf("expected Name->ClassA, got %q", cfg.Columns["Name"])
	}
	if cfg.Columns["Amount"] != "NumA" {
		t.Errorf("expected Amount->NumA, got %q", cfg.Columns["Amount"])
	}
	if cfg.Columns["Title Column"] != "Title" {
		t.Errorf("expected 'Title Column'->Title, got %q", cfg.Columns["Title Column"])
	}

	if cfg.Defaults == nil {
		t.Fatal("expected defaults to be set")
	}
}

func TestLoadMapping_FileNotFound(t *testing.T) {
	_, err := LoadMapping("/nonexistent/mapping.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadMapping_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(":::invalid:::"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadMapping(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadMapping_EmptyColumns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	content := `columns: {}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadMapping(path)
	if err == nil {
		t.Fatal("expected error for empty columns")
	}
}

func TestParseCSV_BasicMapping(t *testing.T) {
	csvData := "Name,Amount,Extra\nAlice,100,ignore\nBob,200,skip\n"
	mapping := &MappingConfig{
		Columns: map[string]string{
			"Name":   "ClassA",
			"Amount": "NumA",
		},
	}

	records, err := ParseCSV(strings.NewReader(csvData), mapping)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// ClassA should be in ClassHash
	rec0 := records[0]
	classHash, ok := rec0["ClassHash"].(map[string]any)
	if !ok {
		t.Fatal("expected ClassHash to be a map")
	}
	if classHash["ClassA"] != "Alice" {
		t.Errorf("expected ClassA 'Alice', got %v", classHash["ClassA"])
	}

	// NumA should be in NumHash
	numHash, ok := rec0["NumHash"].(map[string]any)
	if !ok {
		t.Fatal("expected NumHash to be a map")
	}
	if numHash["NumA"] != "100" {
		t.Errorf("expected NumA '100', got %v", numHash["NumA"])
	}
}

func TestParseCSV_TopLevelFields(t *testing.T) {
	csvData := "MyTitle,MyBody,MyStatus\nTest Title,Test Body,100\n"
	mapping := &MappingConfig{
		Columns: map[string]string{
			"MyTitle":  "Title",
			"MyBody":   "Body",
			"MyStatus": "Status",
		},
	}

	records, err := ParseCSV(strings.NewReader(csvData), mapping)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if rec["Title"] != "Test Title" {
		t.Errorf("expected Title 'Test Title', got %v", rec["Title"])
	}
	if rec["Body"] != "Test Body" {
		t.Errorf("expected Body 'Test Body', got %v", rec["Body"])
	}
	if rec["Status"] != "100" {
		t.Errorf("expected Status '100', got %v", rec["Status"])
	}
}

func TestParseCSV_WithDefaults(t *testing.T) {
	csvData := "Name\nAlice\nBob\n"
	mapping := &MappingConfig{
		Columns: map[string]string{
			"Name": "ClassA",
		},
		Defaults: map[string]any{
			"ClassHash": map[string]any{
				"ClassB": "imported",
			},
		},
	}

	records, err := ParseCSV(strings.NewReader(csvData), mapping)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Both records should have ClassB default and ClassA from CSV
	for i, rec := range records {
		classHash, ok := rec["ClassHash"].(map[string]any)
		if !ok {
			t.Fatalf("record %d: expected ClassHash to be a map", i)
		}
		if classHash["ClassB"] != "imported" {
			t.Errorf("record %d: expected ClassB 'imported', got %v", i, classHash["ClassB"])
		}
		if classHash["ClassA"] == nil {
			t.Errorf("record %d: expected ClassA to be set", i)
		}
	}

	// Verify defaults are not shared between records (deep copy)
	classHash0 := records[0]["ClassHash"].(map[string]any)
	classHash1 := records[1]["ClassHash"].(map[string]any)
	classHash0["ClassB"] = "modified"
	if classHash1["ClassB"] == "modified" {
		t.Error("defaults should be deep-copied, not shared between records")
	}
}

func TestParseCSV_NoMatchingColumns(t *testing.T) {
	csvData := "Foo,Bar\n1,2\n"
	mapping := &MappingConfig{
		Columns: map[string]string{
			"Name": "ClassA",
		},
	}

	_, err := ParseCSV(strings.NewReader(csvData), mapping)
	if err == nil {
		t.Fatal("expected error when no columns match")
	}
}

func TestParseCSV_EmptyCSV(t *testing.T) {
	mapping := &MappingConfig{
		Columns: map[string]string{
			"Name": "ClassA",
		},
	}

	_, err := ParseCSV(strings.NewReader(""), mapping)
	if err == nil {
		t.Fatal("expected error for empty CSV")
	}
}

func TestParseCSV_HeaderOnly(t *testing.T) {
	csvData := "Name,Amount\n"
	mapping := &MappingConfig{
		Columns: map[string]string{
			"Name": "ClassA",
		},
	}

	records, err := ParseCSV(strings.NewReader(csvData), mapping)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("expected 0 records for header-only CSV, got %d", len(records))
	}
}

func TestParseCSV_DateAndDescriptionFields(t *testing.T) {
	csvData := "DateCol,DescCol,CheckCol\n2024-01-01,Some description,true\n"
	mapping := &MappingConfig{
		Columns: map[string]string{
			"DateCol":  "DateA",
			"DescCol":  "DescriptionA",
			"CheckCol": "CheckA",
		},
	}

	records, err := ParseCSV(strings.NewReader(csvData), mapping)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]

	dateHash, ok := rec["DateHash"].(map[string]any)
	if !ok {
		t.Fatal("expected DateHash to be a map")
	}
	if dateHash["DateA"] != "2024-01-01" {
		t.Errorf("expected DateA '2024-01-01', got %v", dateHash["DateA"])
	}

	descHash, ok := rec["DescriptionHash"].(map[string]any)
	if !ok {
		t.Fatal("expected DescriptionHash to be a map")
	}
	if descHash["DescriptionA"] != "Some description" {
		t.Errorf("expected DescriptionA 'Some description', got %v", descHash["DescriptionA"])
	}

	checkHash, ok := rec["CheckHash"].(map[string]any)
	if !ok {
		t.Fatal("expected CheckHash to be a map")
	}
	if checkHash["CheckA"] != "true" {
		t.Errorf("expected CheckA 'true', got %v", checkHash["CheckA"])
	}
}

func TestImport_WithKeys_UsesBulkUpsert(t *testing.T) {
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

	records := []map[string]any{
		{"Title": "Item 1", "ClassHash": map[string]any{"ClassA": "key1"}},
		{"Title": "Item 2", "ClassHash": map[string]any{"ClassA": "key2"}},
	}

	results, err := svc.Import(context.Background(), 100, records, []string{"ClassA"})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if receivedPath != "/api/items/100/bulkupsert" {
		t.Errorf("expected bulkupsert path, got %s", receivedPath)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result (bulkupsert response), got %d", len(results))
	}

	// Verify Keys and Data were sent
	rawKeys, ok := receivedBody["Keys"].([]any)
	if !ok {
		t.Fatal("expected Keys in payload")
	}
	if len(rawKeys) != 1 || rawKeys[0] != "ClassA" {
		t.Errorf("expected Keys [ClassA], got %v", rawKeys)
	}
}

func TestImport_WithoutKeys_CreatesOneByOne(t *testing.T) {
	var paths []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"Created"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	records := []map[string]any{
		{"Title": "Item 1"},
		{"Title": "Item 2"},
		{"Title": "Item 3"},
	}

	results, err := svc.Import(context.Background(), 100, records, nil)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	for _, p := range paths {
		if p != "/api/items/100/create" {
			t.Errorf("expected create path, got %s", p)
		}
	}
}

func TestImport_WithoutKeys_ErrorMidway(t *testing.T) {
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"server error"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Message":"Created"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	records := []map[string]any{
		{"Title": "Item 1"},
		{"Title": "Item 2"},
		{"Title": "Item 3"},
	}

	results, err := svc.Import(context.Background(), 100, records, nil)
	if err == nil {
		t.Fatal("expected error on second record creation")
	}

	// Should have 1 successful result before the error
	if len(results) != 1 {
		t.Errorf("expected 1 successful result before error, got %d", len(results))
	}
}

func TestImport_EmptyRecords_WithKeys(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewRecordService(client)

	results, err := svc.Import(context.Background(), 100, []map[string]any{}, []string{"ClassA"})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result from bulkupsert, got %d", len(results))
	}
}

func TestImport_EmptyRecords_WithoutKeys(t *testing.T) {
	client := api.NewWithHTTPClient("http://unused", "key", "1.1", nil)
	svc := NewRecordService(client)

	results, err := svc.Import(context.Background(), 100, []map[string]any{}, nil)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty records, got %d", len(results))
	}
}
