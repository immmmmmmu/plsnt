package access

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

// mockCommander is a test double for Commander.
type mockCommander struct {
	// outputs maps "name arg1 arg2..." to the output bytes.
	outputs map[string][]byte
	// errors maps "name arg1 arg2..." to an error.
	errors map[string]error
}

func newMockCommander() *mockCommander {
	return &mockCommander{
		outputs: make(map[string][]byte),
		errors:  make(map[string]error),
	}
}

func (m *mockCommander) On(name string, args ...string) *mockCommanderCall {
	key := commandKey(name, args...)
	return &mockCommanderCall{m: m, key: key}
}

type mockCommanderCall struct {
	m   *mockCommander
	key string
}

func (c *mockCommanderCall) Return(output []byte, err error) {
	if err != nil {
		c.m.errors[c.key] = err
	} else {
		c.m.outputs[c.key] = output
	}
}

func (m *mockCommander) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	key := commandKey(name, args...)
	if err, ok := m.errors[key]; ok {
		return nil, err
	}
	if out, ok := m.outputs[key]; ok {
		return out, nil
	}
	return nil, fmt.Errorf("unexpected command: %s", key)
}

func commandKey(name string, args ...string) string {
	parts := append([]string{name}, args...)
	return strings.Join(parts, " ")
}

func TestCheckMDBTools_Installed(t *testing.T) {
	mock := newMockCommander()
	mock.On("mdb-tables", "--version").Return([]byte("0.7.1\n"), nil)

	reader := NewAccessReaderWithCommander(mock)
	err := reader.CheckMDBTools(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestCheckMDBTools_NotInstalled(t *testing.T) {
	mock := newMockCommander()
	mock.On("mdb-tables", "--version").Return(nil, fmt.Errorf("exec: \"mdb-tables\": executable file not found in $PATH"))

	reader := NewAccessReaderWithCommander(mock)
	err := reader.CheckMDBTools(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mdbtools is not installed") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestListTables_Success(t *testing.T) {
	mock := newMockCommander()
	mock.On("mdb-tables", "-1", "test.mdb").Return([]byte("Users\nOrders\nProducts\n"), nil)

	reader := NewAccessReaderWithCommander(mock)
	tables, err := reader.ListTables(context.Background(), "test.mdb")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(tables) != 3 {
		t.Fatalf("expected 3 tables, got %d", len(tables))
	}
	expected := []string{"Users", "Orders", "Products"}
	for i, name := range expected {
		if tables[i] != name {
			t.Errorf("table[%d]: expected %q, got %q", i, name, tables[i])
		}
	}
}

func TestListTables_Empty(t *testing.T) {
	mock := newMockCommander()
	mock.On("mdb-tables", "-1", "empty.mdb").Return([]byte(""), nil)

	reader := NewAccessReaderWithCommander(mock)
	tables, err := reader.ListTables(context.Background(), "empty.mdb")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(tables) != 0 {
		t.Fatalf("expected 0 tables, got %d", len(tables))
	}
}

func TestListTables_Error(t *testing.T) {
	mock := newMockCommander()
	mock.On("mdb-tables", "-1", "bad.mdb").Return(nil, fmt.Errorf("not a valid MDB file"))

	reader := NewAccessReaderWithCommander(mock)
	_, err := reader.ListTables(context.Background(), "bad.mdb")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to list tables") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestExportTable_Success(t *testing.T) {
	csvData := []byte("id,name,email\n1,Alice,alice@example.com\n2,Bob,bob@example.com\n")
	mock := newMockCommander()
	mock.On("mdb-export", "test.mdb", "Users").Return(csvData, nil)

	reader := NewAccessReaderWithCommander(mock)
	out, err := reader.ExportTable(context.Background(), "test.mdb", "Users")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !bytes.Equal(out, csvData) {
		t.Fatalf("unexpected output: %s", string(out))
	}
}

func TestExportTable_Error(t *testing.T) {
	mock := newMockCommander()
	mock.On("mdb-export", "test.mdb", "NoSuchTable").Return(nil, fmt.Errorf("table not found"))

	reader := NewAccessReaderWithCommander(mock)
	_, err := reader.ExportTable(context.Background(), "test.mdb", "NoSuchTable")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to export table") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNewAccessReader(t *testing.T) {
	reader := NewAccessReader()
	if reader == nil {
		t.Fatal("NewAccessReader returned nil")
	}
	if reader.cmd == nil {
		t.Fatal("cmd field should not be nil")
	}
}

func TestRealCommander_Run_Success(t *testing.T) {
	cmd := &realCommander{}
	// Use "true" command which exists on all Unix systems and always succeeds
	_, err := cmd.Run(context.Background(), "true")
	if err != nil {
		t.Fatalf("expected no error from 'true' command, got: %v", err)
	}
}

func TestRealCommander_Run_Failure(t *testing.T) {
	cmd := &realCommander{}
	_, err := cmd.Run(context.Background(), "false")
	if err == nil {
		t.Fatal("expected error from 'false' command, got nil")
	}
}

func TestRealCommander_Run_NotFound(t *testing.T) {
	cmd := &realCommander{}
	_, err := cmd.Run(context.Background(), "nonexistent-command-xyz-12345")
	if err == nil {
		t.Fatal("expected error for nonexistent command, got nil")
	}
}

func TestExportTable_IntegrationWithParseCSV(t *testing.T) {
	// Test that exported CSV can be parsed through the existing import pipeline
	csvData := []byte("id,name,category\n1,Widget,ClassA-val\n2,Gadget,ClassA-val2\n")
	mock := newMockCommander()
	mock.On("mdb-export", "test.mdb", "Products").Return(csvData, nil)

	reader := NewAccessReaderWithCommander(mock)
	out, err := reader.ExportTable(context.Background(), "test.mdb", "Products")
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	mapping := &pleasanter.MappingConfig{
		Columns: map[string]string{
			"name":     "Title",
			"category": "ClassA",
		},
	}

	records, err := pleasanter.ParseCSV(bytes.NewReader(out), mapping)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Verify first record
	if records[0]["Title"] != "Widget" {
		t.Errorf("expected Title=Widget, got %v", records[0]["Title"])
	}
	classHash, ok := records[0]["ClassHash"].(map[string]any)
	if !ok {
		t.Fatal("expected ClassHash to be a map")
	}
	if classHash["ClassA"] != "ClassA-val" {
		t.Errorf("expected ClassA=ClassA-val, got %v", classHash["ClassA"])
	}
}
