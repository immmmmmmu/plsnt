package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNew_SupportedFormats(t *testing.T) {
	for _, name := range []string{"json", "JSON", "ndjson", "table", "csv", "count", "ids", ""} {
		if _, err := New(name, nil); err != nil {
			t.Errorf("New(%q) returned error: %v", name, err)
		}
	}
}

func TestNew_UnsupportedFormat(t *testing.T) {
	if _, err := New("xml", nil); err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestJSON_SingleMap(t *testing.T) {
	f, _ := New("json", nil)
	var buf bytes.Buffer
	data := map[string]any{"Title": "hello", "Status": float64(100)}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if result["Title"] != "hello" {
		t.Errorf("expected Title 'hello', got %v", result["Title"])
	}
}

func TestJSON_Array(t *testing.T) {
	f, _ := New("json", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{"Title": "a"},
		{"Title": "b"},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON array: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestNDJSON_MultipleRecords(t *testing.T) {
	f, _ := New("ndjson", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{"Name": "Alice", "Age": float64(30)},
		{"Name": "Bob", "Age": float64(25)},
		{"Name": "Charlie", "Age": float64(35)},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}

	// Verify first and last record content
	var first map[string]any
	json.Unmarshal([]byte(lines[0]), &first)
	if first["Name"] != "Alice" {
		t.Errorf("expected first record Name 'Alice', got %v", first["Name"])
	}

	var last map[string]any
	json.Unmarshal([]byte(lines[2]), &last)
	if last["Name"] != "Charlie" {
		t.Errorf("expected last record Name 'Charlie', got %v", last["Name"])
	}
}

func TestNDJSON_SingleRecord(t *testing.T) {
	f, _ := New("ndjson", nil)
	var buf bytes.Buffer
	data := map[string]any{"Title": "hello"}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if m["Title"] != "hello" {
		t.Errorf("expected Title 'hello', got %v", m["Title"])
	}
}

func TestNDJSON_EmptyData(t *testing.T) {
	f, _ := New("ndjson", nil)
	var buf bytes.Buffer

	if err := f.Format(&buf, []map[string]any{}); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("empty data should produce no output, got %q", buf.String())
	}
}

func TestTable_BasicOutput(t *testing.T) {
	f, _ := New("table", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{"Name": "Alice", "Age": float64(30)},
		{"Name": "Bob", "Age": float64(25)},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Name") {
		t.Error("table output should contain header 'Name'")
	}
	if !strings.Contains(output, "Alice") {
		t.Error("table output should contain 'Alice'")
	}
	if !strings.Contains(output, "Bob") {
		t.Error("table output should contain 'Bob'")
	}
}

func TestTable_WithFields(t *testing.T) {
	f, _ := New("table", []string{"Name"})
	var buf bytes.Buffer
	data := []map[string]any{
		{"Name": "Alice", "Age": float64(30)},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Name") {
		t.Error("should contain 'Name' header")
	}
	if strings.Contains(output, "Age") {
		t.Error("should not contain 'Age' header when fields filter is set")
	}
}

func TestTable_EmptyData(t *testing.T) {
	f, _ := New("table", nil)
	var buf bytes.Buffer

	if err := f.Format(&buf, []map[string]any{}); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(buf.String(), "No data") {
		t.Error("empty data should show 'No data'")
	}
}

func TestCSV_BasicOutput(t *testing.T) {
	f, _ := New("csv", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{"Name": "Alice", "Age": float64(30)},
		{"Name": "Bob", "Age": float64(25)},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 { // header + 2 rows
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Header should contain Age and Name (sorted)
	if !strings.Contains(lines[0], "Age") || !strings.Contains(lines[0], "Name") {
		t.Errorf("header missing fields: %s", lines[0])
	}
}

func TestCSV_WithFields(t *testing.T) {
	f, _ := New("csv", []string{"Name"})
	var buf bytes.Buffer
	data := []map[string]any{
		{"Name": "Alice", "Age": float64(30)},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if lines[0] != "Name" {
		t.Errorf("expected header 'Name', got %q", lines[0])
	}
	if lines[1] != "Alice" {
		t.Errorf("expected 'Alice', got %q", lines[1])
	}
}

func TestCSV_EmptyData(t *testing.T) {
	f, _ := New("csv", nil)
	var buf bytes.Buffer

	if err := f.Format(&buf, []map[string]any{}); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("empty data should produce no output, got %q", buf.String())
	}
}

func TestFormatValue_Types(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{nil, ""},
		{"hello", "hello"},
		{json.Number("123.456"), "123.456"},
		{float64(42), "42"},
		{true, "true"},
		{false, "false"},
		{map[string]any{"a": "b"}, `{"a":"b"}`},
		{[]any{"x", "y"}, `["x","y"]`},
	}

	for _, tt := range tests {
		got := formatValue(tt.input)
		if got != tt.expected {
			t.Errorf("formatValue(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCount_MultipleRows(t *testing.T) {
	f, _ := New("count", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{"ResultId": float64(1), "Title": "a"},
		{"ResultId": float64(2), "Title": "b"},
		{"ResultId": float64(3), "Title": "c"},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if got := strings.TrimSpace(buf.String()); got != "3" {
		t.Errorf("expected '3', got %q", got)
	}
}

func TestCount_EmptyData(t *testing.T) {
	f, _ := New("count", nil)
	var buf bytes.Buffer

	if err := f.Format(&buf, []map[string]any{}); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if got := strings.TrimSpace(buf.String()); got != "0" {
		t.Errorf("expected '0', got %q", got)
	}
}

func TestCount_SingleRecord(t *testing.T) {
	f, _ := New("count", nil)
	var buf bytes.Buffer
	data := map[string]any{"Id": float64(1)}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if got := strings.TrimSpace(buf.String()); got != "1" {
		t.Errorf("expected '1', got %q", got)
	}
}

func TestIDs_ResultId(t *testing.T) {
	f, _ := New("ids", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{"ResultId": float64(101), "Title": "a"},
		{"ResultId": float64(102), "Title": "b"},
		{"ResultId": float64(103), "Title": "c"},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	expected := []string{"101", "102", "103"}
	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("line %d: expected %q, got %q", i, expected[i], line)
		}
	}
}

func TestIDs_IssueId(t *testing.T) {
	f, _ := New("ids", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{"IssueId": float64(201)},
		{"IssueId": float64(202)},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "201" || lines[1] != "202" {
		t.Errorf("unexpected output: %v", lines)
	}
}

func TestIDs_FallbackToId(t *testing.T) {
	f, _ := New("ids", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{"Id": float64(301), "Title": "x"},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if got := strings.TrimSpace(buf.String()); got != "301" {
		t.Errorf("expected '301', got %q", got)
	}
}

func TestIDs_NoIdField(t *testing.T) {
	f, _ := New("ids", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{"Title": "no id here"},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected no output for records without ID fields, got %q", buf.String())
	}
}

func TestIDs_EmptyData(t *testing.T) {
	f, _ := New("ids", nil)
	var buf bytes.Buffer

	if err := f.Format(&buf, []map[string]any{}); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected no output for empty data, got %q", buf.String())
	}
}

func TestIDs_SiteId(t *testing.T) {
	f, _ := New("ids", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{"SiteId": float64(401)},
		{"SiteId": float64(402)},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "401" || lines[1] != "402" {
		t.Errorf("unexpected output: %v", lines)
	}
}

func TestIDs_PriorityOrder(t *testing.T) {
	f, _ := New("ids", nil)
	var buf bytes.Buffer
	// ResultId should take priority over Id
	data := []map[string]any{
		{"ResultId": float64(501), "Id": float64(999)},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if got := strings.TrimSpace(buf.String()); got != "501" {
		t.Errorf("expected '501' (ResultId takes priority), got %q", got)
	}
}

func TestFlattenHashFields_ExpandsClassHash(t *testing.T) {
	rows := []map[string]any{
		{
			"ResultId":  float64(1),
			"Title":     "Item A",
			"ClassHash": map[string]any{"ClassA": "val1", "ClassB": "val2"},
			"NumHash":   map[string]any{"NumA": float64(100), "NumB": float64(200)},
		},
	}

	result := flattenHashFields(rows)

	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	row := result[0]

	// Original Hash fields should be removed
	if _, exists := row["ClassHash"]; exists {
		t.Error("ClassHash should be removed after flattening")
	}
	if _, exists := row["NumHash"]; exists {
		t.Error("NumHash should be removed after flattening")
	}

	// Individual columns should exist
	if row["ClassA"] != "val1" {
		t.Errorf("expected ClassA 'val1', got %v", row["ClassA"])
	}
	if row["ClassB"] != "val2" {
		t.Errorf("expected ClassB 'val2', got %v", row["ClassB"])
	}
	if row["NumA"] != float64(100) {
		t.Errorf("expected NumA 100, got %v", row["NumA"])
	}
	if row["NumB"] != float64(200) {
		t.Errorf("expected NumB 200, got %v", row["NumB"])
	}

	// Non-hash fields should be preserved
	if row["ResultId"] != float64(1) {
		t.Errorf("expected ResultId 1, got %v", row["ResultId"])
	}
	if row["Title"] != "Item A" {
		t.Errorf("expected Title 'Item A', got %v", row["Title"])
	}
}

func TestFlattenHashFields_DescriptionHash(t *testing.T) {
	rows := []map[string]any{
		{
			"ResultId":        float64(1),
			"DescriptionHash": map[string]any{"DescriptionA": "text1", "DescriptionB": "text2"},
		},
	}

	result := flattenHashFields(rows)
	row := result[0]

	if _, exists := row["DescriptionHash"]; exists {
		t.Error("DescriptionHash should be removed after flattening")
	}
	if row["DescriptionA"] != "text1" {
		t.Errorf("expected DescriptionA 'text1', got %v", row["DescriptionA"])
	}
	if row["DescriptionB"] != "text2" {
		t.Errorf("expected DescriptionB 'text2', got %v", row["DescriptionB"])
	}
}

func TestTable_FlattenedHashFieldHeaders(t *testing.T) {
	f, _ := New("table", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{
			"ResultId":  float64(1),
			"ClassHash": map[string]any{"ClassA": "val1", "ClassB": "val2"},
		},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "ClassHash") {
		t.Error("table output should not contain 'ClassHash' after flattening")
	}
	if !strings.Contains(output, "ClassA") {
		t.Error("table output should contain flattened column 'ClassA'")
	}
	if !strings.Contains(output, "ClassB") {
		t.Error("table output should contain flattened column 'ClassB'")
	}
	if !strings.Contains(output, "val1") {
		t.Error("table output should contain value 'val1'")
	}
}

func TestCSV_FlattenedHashFieldHeaders(t *testing.T) {
	f, _ := New("csv", nil)
	var buf bytes.Buffer
	data := []map[string]any{
		{
			"ResultId":  float64(1),
			"ClassHash": map[string]any{"ClassA": "val1"},
			"NumHash":   map[string]any{"NumA": float64(42)},
		},
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	header := lines[0]

	if strings.Contains(header, "ClassHash") {
		t.Error("CSV header should not contain 'ClassHash' after flattening")
	}
	if strings.Contains(header, "NumHash") {
		t.Error("CSV header should not contain 'NumHash' after flattening")
	}
	if !strings.Contains(header, "ClassA") {
		t.Error("CSV header should contain flattened column 'ClassA'")
	}
	if !strings.Contains(header, "NumA") {
		t.Error("CSV header should contain flattened column 'NumA'")
	}
}

// BUG-4: Pleasanter API response structure should be unwrapped by outputResult,
// but toRows should handle it gracefully when Response.Data is an array.
func TestToRows_PleasanterAPIResponse(t *testing.T) {
	// Simulate the structure returned by PostRaw: {"Response":{"Data":[...]}, "StatusCode":200}
	data := map[string]any{
		"Response": map[string]any{
			"Data": []any{
				map[string]any{"ResultId": float64(1), "Title": "A"},
				map[string]any{"ResultId": float64(2), "Title": "B"},
			},
			"TotalCount": float64(2),
		},
		"StatusCode": float64(200),
	}

	// Without unwrapping, toRows treats the whole map as 1 row
	rows, err := toRows(data)
	if err != nil {
		t.Fatalf("toRows failed: %v", err)
	}
	// This is the current (broken) behavior: 1 row with keys "Response" and "StatusCode"
	if len(rows) != 1 {
		t.Fatalf("expected 1 row (current behavior), got %d", len(rows))
	}
}

// BUG-4: Test UnwrapAPIResponse helper
func TestUnwrapAPIResponse_WithData(t *testing.T) {
	input := map[string]any{
		"Response": map[string]any{
			"Data": []any{
				map[string]any{"ResultId": float64(1), "Title": "A"},
				map[string]any{"ResultId": float64(2), "Title": "B"},
			},
			"TotalCount": float64(2),
			"Offset":     float64(0),
			"PageSize":   float64(200),
		},
		"StatusCode": float64(200),
	}

	data, meta := UnwrapAPIResponse(input)
	rows, err := toRows(data)
	if err != nil {
		t.Fatalf("toRows failed: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows after unwrap, got %d", len(rows))
	}
	if rows[0]["Title"] != "A" {
		t.Errorf("expected Title 'A', got %v", rows[0]["Title"])
	}
	if meta == nil {
		t.Fatal("expected non-nil metadata")
	}
	if meta["TotalCount"] != float64(2) {
		t.Errorf("expected TotalCount 2, got %v", meta["TotalCount"])
	}
}

func TestUnwrapAPIResponse_WithoutData(t *testing.T) {
	// create/update/delete response: {"Id":123, "StatusCode":200, "Message":"..."}
	input := map[string]any{
		"Id":         float64(123),
		"StatusCode": float64(200),
		"Message":    "Created",
	}

	data, meta := UnwrapAPIResponse(input)
	if meta != nil {
		t.Errorf("expected nil metadata for non-API-list response, got %v", meta)
	}
	// Should return the original map unchanged
	m, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", data)
	}
	if m["Id"] != float64(123) {
		t.Errorf("expected Id 123, got %v", m["Id"])
	}
}

func TestUnwrapAPIResponse_EmptyData(t *testing.T) {
	input := map[string]any{
		"Response": map[string]any{
			"Data":       []any{},
			"TotalCount": float64(0),
		},
		"StatusCode": float64(200),
	}

	data, meta := UnwrapAPIResponse(input)
	rows, err := toRows(data)
	if err != nil {
		t.Fatalf("toRows failed: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
	if meta == nil {
		t.Fatal("expected non-nil metadata")
	}
}

func TestToRows_StructInput(t *testing.T) {
	type item struct {
		Name string `json:"Name"`
		Age  int    `json:"Age"`
	}

	rows, err := toRows(item{Name: "Test", Age: 10})
	if err != nil {
		t.Fatalf("toRows failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["Name"] != "Test" {
		t.Errorf("expected Name 'Test', got %v", rows[0]["Name"])
	}
}

// BUG-5: Record struct with empty Hash fields should still include Hash keys
func TestToRows_RecordWithEmptyHash(t *testing.T) {
	type Record struct {
		ResultId  int64             `json:"ResultId"`
		Title     string            `json:"Title"`
		ClassHash map[string]string `json:"ClassHash"`
		NumHash   map[string]string `json:"NumHash"`
	}

	rec := Record{
		ResultId:  1,
		Title:     "Test",
		ClassHash: map[string]string{},
		NumHash:   nil,
	}

	rows, err := toRows(rec)
	if err != nil {
		t.Fatalf("toRows failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	// Empty map should be preserved (not omitted)
	if _, exists := rows[0]["ClassHash"]; !exists {
		t.Error("ClassHash should be present even when empty")
	}
	// nil map should also be present (as null → nil in JSON roundtrip)
	if _, exists := rows[0]["NumHash"]; !exists {
		t.Error("NumHash should be present even when nil")
	}
}
