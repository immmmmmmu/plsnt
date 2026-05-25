package migrate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

func testSchema() *pleasanter.SchemaInfo {
	return &pleasanter.SchemaInfo{
		SiteID:    123,
		Title:     "Test Site",
		TableType: "Results",
		Columns: []pleasanter.SchemaColumn{
			{ColumnName: "ClassA", LabelText: "Category", Type: "Class"},
			{ColumnName: "ClassB", LabelText: "SubCategory", Type: "Class"},
			{ColumnName: "NumA", LabelText: "Amount", Type: "Num"},
			{ColumnName: "Title", LabelText: "Title", Type: "Standard"},
			{ColumnName: "DateA", LabelText: "DueDate", Type: "Date"},
		},
	}
}

func TestGenerateMapping_ExactColumnName(t *testing.T) {
	headers := []string{"ClassA", "NumA", "Title"}
	result := GenerateMapping(headers, testSchema())

	if len(result.Mapped) != 3 {
		t.Fatalf("expected 3 mapped columns, got %d", len(result.Mapped))
	}
	if result.Mapped["ClassA"] != "ClassA" {
		t.Errorf("expected ClassA -> ClassA, got %s", result.Mapped["ClassA"])
	}
	if result.Mapped["NumA"] != "NumA" {
		t.Errorf("expected NumA -> NumA, got %s", result.Mapped["NumA"])
	}
	if result.Mapped["Title"] != "Title" {
		t.Errorf("expected Title -> Title, got %s", result.Mapped["Title"])
	}
	if len(result.Unmapped) != 0 {
		t.Errorf("expected 0 unmapped columns, got %d: %v", len(result.Unmapped), result.Unmapped)
	}
}

func TestGenerateMapping_LabelTextMatch(t *testing.T) {
	headers := []string{"Category", "Amount", "DueDate"}
	result := GenerateMapping(headers, testSchema())

	if len(result.Mapped) != 3 {
		t.Fatalf("expected 3 mapped columns, got %d", len(result.Mapped))
	}
	if result.Mapped["Category"] != "ClassA" {
		t.Errorf("expected Category -> ClassA, got %s", result.Mapped["Category"])
	}
	if result.Mapped["Amount"] != "NumA" {
		t.Errorf("expected Amount -> NumA, got %s", result.Mapped["Amount"])
	}
	if result.Mapped["DueDate"] != "DateA" {
		t.Errorf("expected DueDate -> DateA, got %s", result.Mapped["DueDate"])
	}
}

func TestGenerateMapping_CaseInsensitiveMatch(t *testing.T) {
	headers := []string{"classa", "NUMA", "category", "duedate"}
	result := GenerateMapping(headers, testSchema())

	if len(result.Mapped) != 4 {
		t.Fatalf("expected 4 mapped columns, got %d", len(result.Mapped))
	}
	if result.Mapped["classa"] != "ClassA" {
		t.Errorf("expected classa -> ClassA, got %s", result.Mapped["classa"])
	}
	if result.Mapped["NUMA"] != "NumA" {
		t.Errorf("expected NUMA -> NumA, got %s", result.Mapped["NUMA"])
	}
	if result.Mapped["category"] != "ClassA" {
		t.Errorf("expected category -> ClassA, got %s", result.Mapped["category"])
	}
	if result.Mapped["duedate"] != "DateA" {
		t.Errorf("expected duedate -> DateA, got %s", result.Mapped["duedate"])
	}
}

func TestGenerateMapping_UnmappedColumns(t *testing.T) {
	headers := []string{"ClassA", "UnknownField", "AnotherField"}
	result := GenerateMapping(headers, testSchema())

	if len(result.Mapped) != 1 {
		t.Fatalf("expected 1 mapped column, got %d", len(result.Mapped))
	}
	if len(result.Unmapped) != 2 {
		t.Fatalf("expected 2 unmapped columns, got %d", len(result.Unmapped))
	}
	if result.Unmapped[0] != "UnknownField" {
		t.Errorf("expected first unmapped to be UnknownField, got %s", result.Unmapped[0])
	}
	if result.Unmapped[1] != "AnotherField" {
		t.Errorf("expected second unmapped to be AnotherField, got %s", result.Unmapped[1])
	}
}

func TestGenerateMapping_EmptyHeaders(t *testing.T) {
	headers := []string{"", "  ", "ClassA", ""}
	result := GenerateMapping(headers, testSchema())

	if len(result.Mapped) != 1 {
		t.Fatalf("expected 1 mapped column, got %d", len(result.Mapped))
	}
	if result.Mapped["ClassA"] != "ClassA" {
		t.Errorf("expected ClassA -> ClassA, got %s", result.Mapped["ClassA"])
	}
	if len(result.Unmapped) != 0 {
		t.Errorf("expected 0 unmapped columns, got %d", len(result.Unmapped))
	}
}

func TestGenerateMapping_NoMatchingColumns(t *testing.T) {
	headers := []string{"Foo", "Bar", "Baz"}
	result := GenerateMapping(headers, testSchema())

	if len(result.Mapped) != 0 {
		t.Errorf("expected 0 mapped columns, got %d", len(result.Mapped))
	}
	if len(result.Unmapped) != 3 {
		t.Errorf("expected 3 unmapped columns, got %d", len(result.Unmapped))
	}
}

func TestReadCSVHeaders(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "test.csv")

	content := "Name,Age,City\nAlice,30,Tokyo\nBob,25,Osaka\n"
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	headers, err := ReadCSVHeaders(csvPath)
	if err != nil {
		t.Fatalf("ReadCSVHeaders returned error: %v", err)
	}

	if len(headers) != 3 {
		t.Fatalf("expected 3 headers, got %d", len(headers))
	}
	expected := []string{"Name", "Age", "City"}
	for i, h := range headers {
		if h != expected[i] {
			t.Errorf("header[%d]: expected %q, got %q", i, expected[i], h)
		}
	}
}

func TestReadCSVHeaders_FileNotFound(t *testing.T) {
	_, err := ReadCSVHeaders("/nonexistent/path/test.csv")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestReadCSVHeaders_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "empty.csv")

	if err := os.WriteFile(csvPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write empty CSV: %v", err)
	}

	_, err := ReadCSVHeaders(csvPath)
	if err == nil {
		t.Fatal("expected error for empty CSV file")
	}
}

func TestWriteMappingYAML_MappedOnly(t *testing.T) {
	result := &MappingResult{
		Mapped: map[string]string{
			"Category": "ClassA",
			"Amount":   "NumA",
		},
	}

	var buf bytes.Buffer
	if err := WriteMappingYAML(&buf, result); err != nil {
		t.Fatalf("WriteMappingYAML returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "columns:") {
		t.Error("output should contain 'columns:'")
	}
	if !strings.Contains(output, "Category: ClassA") {
		t.Error("output should contain 'Category: ClassA'")
	}
	if !strings.Contains(output, "Amount: NumA") {
		t.Error("output should contain 'Amount: NumA'")
	}
	if strings.Contains(output, "# Unmapped") {
		t.Error("output should not contain unmapped comment section")
	}
}

func TestWriteMappingYAML_WithUnmapped(t *testing.T) {
	result := &MappingResult{
		Mapped: map[string]string{
			"Category": "ClassA",
		},
		Unmapped: []string{"UnknownB", "UnknownA"},
	}

	var buf bytes.Buffer
	if err := WriteMappingYAML(&buf, result); err != nil {
		t.Fatalf("WriteMappingYAML returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# Unmapped CSV columns") {
		t.Error("output should contain unmapped comment header")
	}
	// Unmapped columns should be sorted alphabetically
	idxA := strings.Index(output, "UnknownA")
	idxB := strings.Index(output, "UnknownB")
	if idxA == -1 || idxB == -1 {
		t.Fatal("output should contain both unmapped column names")
	}
	if idxA > idxB {
		t.Error("unmapped columns should be sorted alphabetically (UnknownA before UnknownB)")
	}
}
