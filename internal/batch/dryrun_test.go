package batch

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDryRun_ScaffoldCRM(t *testing.T) {
	templatesDir := filepath.Join("..", "..", "templates")
	data, err := os.ReadFile(filepath.Join(templatesDir, "scaffold-crm.yaml"))
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	batch, err := ParseRaw(data)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	batch.Variables["folder_id"] = "12345"
	batch.ExpandVariables()

	var buf bytes.Buffer
	engine, err := NewEngine(EngineOptions{
		DryRun: true,
		Writer: &buf,
	})
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	results, err := engine.Execute(context.Background(), batch)
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	for _, r := range results {
		if !r.Success {
			t.Errorf("step %q failed in dry-run", r.StepName)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Error("expected DRY-RUN marker in output")
	}
	if !strings.Contains(output, "create-customer-master") {
		t.Error("expected create-customer-master step in output")
	}
	if !strings.Contains(output, "12345") {
		t.Error("expected expanded variable value 12345 in output")
	}
}

func TestDryRun_MonthlyReport(t *testing.T) {
	templatesDir := filepath.Join("..", "..", "templates")
	data, err := os.ReadFile(filepath.Join(templatesDir, "monthly-report.yaml"))
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	batch, err := ParseRaw(data)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	var buf bytes.Buffer
	engine, err := NewEngine(EngineOptions{
		DryRun: true,
		Writer: &buf,
	})
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	results, err := engine.Execute(context.Background(), batch)
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}

	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}

	for _, r := range results {
		if !r.Success {
			t.Errorf("step %q failed in dry-run", r.StepName)
		}
	}
}

func TestDryRun_IntegrityCheck(t *testing.T) {
	templatesDir := filepath.Join("..", "..", "templates")
	data, err := os.ReadFile(filepath.Join(templatesDir, "integrity-check.yaml"))
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	batch, err := ParseRaw(data)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	var buf bytes.Buffer
	engine, err := NewEngine(EngineOptions{
		DryRun: true,
		Writer: &buf,
	})
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	results, err := engine.Execute(context.Background(), batch)
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	output := buf.String()
	if !strings.Contains(output, "get-schema") {
		t.Error("expected get-schema step in output")
	}
	if !strings.Contains(output, "count-records") {
		t.Error("expected count-records step in output")
	}
}
