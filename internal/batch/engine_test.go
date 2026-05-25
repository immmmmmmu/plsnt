package batch

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngine_DryRun(t *testing.T) {
	batch := &BatchDef{
		Name: "dry-run-test",
		Steps: []StepDef{
			{
				Name:    "step1",
				Command: "record list",
				Args: map[string]any{
					"site-id":   "123",
					"all-pages": true,
				},
			},
			{
				Name:      "step2",
				Command:   "record update",
				DependsOn: []string{"step1"},
				Args: map[string]any{
					"json": `{"Status": 200}`,
				},
			},
		},
	}

	var buf bytes.Buffer
	engine, err := NewEngine(EngineOptions{
		DryRun: true,
		Writer: &buf,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, err := engine.Execute(context.Background(), batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if !r.Success {
			t.Errorf("step %q should be successful in dry-run", r.StepName)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Error("dry-run output should contain [DRY-RUN] prefix")
	}
	if !strings.Contains(output, "step1") {
		t.Error("output should contain step1")
	}
	if !strings.Contains(output, "step2") {
		t.Error("output should contain step2")
	}
}

func TestEngine_NormalExecution(t *testing.T) {
	batch := &BatchDef{
		Name: "exec-test",
		Steps: []StepDef{
			{
				Name:    "step1",
				Command: "site get",
				Args: map[string]any{
					"site-id": "456",
				},
			},
		},
	}

	var buf bytes.Buffer
	engine, err := NewEngine(EngineOptions{
		Writer:      &buf,
		CommandName: "/bin/true",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, err := engine.Execute(context.Background(), batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if !results[0].Success {
		t.Error("step should be successful")
	}

	output := buf.String()
	if !strings.Contains(output, "[EXECUTE]") {
		t.Error("normal execution output should contain [EXECUTE] prefix")
	}
	if !strings.Contains(output, "site get") {
		t.Error("output should contain command")
	}
}

func TestEngine_DependsOnOrdering(t *testing.T) {
	batch := &BatchDef{
		Name: "ordering",
		Steps: []StepDef{
			{
				Name:      "last",
				Command:   "cmd3",
				DependsOn: []string{"middle"},
			},
			{
				Name:      "middle",
				Command:   "cmd2",
				DependsOn: []string{"first"},
			},
			{
				Name:    "first",
				Command: "cmd1",
			},
		},
	}

	var buf bytes.Buffer
	engine, err := NewEngine(EngineOptions{
		Writer:      &buf,
		CommandName: "/bin/true",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, err := engine.Execute(context.Background(), batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Verify ordering
	if results[0].StepName != "first" {
		t.Errorf("expected first step to be %q, got %q", "first", results[0].StepName)
	}
	if results[1].StepName != "middle" {
		t.Errorf("expected second step to be %q, got %q", "middle", results[1].StepName)
	}
	if results[2].StepName != "last" {
		t.Errorf("expected third step to be %q, got %q", "last", results[2].StepName)
	}
}

func TestEngine_LogFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "batch.log")

	batch := &BatchDef{
		Name: "log-test",
		Steps: []StepDef{
			{Name: "s1", Command: "test cmd"},
		},
	}

	var buf bytes.Buffer
	engine, err := NewEngine(EngineOptions{
		Writer:      &buf,
		LogFile:     logPath,
		CommandName: "/bin/true",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = engine.Execute(context.Background(), batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := engine.Close(); err != nil {
		t.Fatalf("failed to close engine: %v", err)
	}

	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(logContent), "s1") {
		t.Error("log file should contain step name")
	}
	if !strings.Contains(string(logContent), "OK") {
		t.Error("log file should contain OK status")
	}
}

func TestEngine_ContextCancellation(t *testing.T) {
	batch := &BatchDef{
		Name: "cancel-test",
		Steps: []StepDef{
			{Name: "s1", Command: "cmd1"},
			{Name: "s2", Command: "cmd2"},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	var buf bytes.Buffer
	engine, err := NewEngine(EngineOptions{Writer: &buf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = engine.Execute(ctx, batch)
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestEngine_InvalidLogFile(t *testing.T) {
	_, err := NewEngine(EngineOptions{
		LogFile: "/nonexistent/dir/batch.log",
	})
	if err == nil {
		t.Fatal("expected error for invalid log file path")
	}
}

func TestBuildCommandDescription(t *testing.T) {
	step := StepDef{
		Name:    "test",
		Command: "record list",
		Args: map[string]any{
			"all-pages": true,
			"site-id":   "123",
		},
	}

	desc := buildCommandDescription(step)
	if !strings.Contains(desc, "record list") {
		t.Error("description should contain command")
	}
	if !strings.Contains(desc, "--all-pages") {
		t.Error("description should contain boolean flag")
	}
	if !strings.Contains(desc, "--site-id") {
		t.Error("description should contain string arg")
	}
}

func TestBuildCommandDescription_NoArgs(t *testing.T) {
	step := StepDef{
		Name:    "test",
		Command: "version",
	}

	desc := buildCommandDescription(step)
	if desc != "version" {
		t.Errorf("expected %q, got %q", "version", desc)
	}
}

func TestEngine_CaptureStepOutput(t *testing.T) {
	engine := &Engine{
		stepOutputs: make(map[string]map[string]string),
	}

	// Test JSON output capture
	engine.captureStepOutput("create-folder", `{"Id":12345,"StatusCode":200}`)
	if engine.stepOutputs["create-folder"]["Id"] != "12345" {
		t.Errorf("expected Id=12345, got %q", engine.stepOutputs["create-folder"]["Id"])
	}
	if engine.stepOutputs["create-folder"]["StatusCode"] != "200" {
		t.Errorf("expected StatusCode=200, got %q", engine.stepOutputs["create-folder"]["StatusCode"])
	}

	// Test non-JSON output (should not error)
	engine.captureStepOutput("non-json", "some plain text output")
	if _, exists := engine.stepOutputs["non-json"]; exists {
		t.Error("non-JSON output should not be captured")
	}
}

func TestEngine_ExpandStepOutputRefs(t *testing.T) {
	engine := &Engine{
		stepOutputs: map[string]map[string]string{
			"create-folder": {"Id": "12345"},
			"create-table":  {"Id": "67890"},
		},
	}

	step := StepDef{
		Name:    "next-step",
		Command: "site create",
		Args: map[string]any{
			"parent": "{{create-folder.Id}}",
			"json":   `{"LinkSiteId":"{{create-table.Id}}"}`,
		},
	}

	expanded := engine.expandStepOutputRefs(step)
	if expanded.Args["parent"] != "12345" {
		t.Errorf("expected parent=12345, got %q", expanded.Args["parent"])
	}
	expectedJSON := `{"LinkSiteId":"67890"}`
	if expanded.Args["json"] != expectedJSON {
		t.Errorf("expected json=%q, got %q", expectedJSON, expanded.Args["json"])
	}
}

func TestEngine_ExpandStepOutputRefs_NoOutputs(t *testing.T) {
	engine := &Engine{
		stepOutputs: make(map[string]map[string]string),
	}

	step := StepDef{
		Name:    "step1",
		Command: "test",
		Args: map[string]any{
			"key": "{{unknown.field}}",
		},
	}

	expanded := engine.expandStepOutputRefs(step)
	if expanded.Args["key"] != "{{unknown.field}}" {
		t.Error("unresolved references should remain as-is")
	}
}

func TestBuildCommandDescription_BoolFalse(t *testing.T) {
	step := StepDef{
		Name:    "test",
		Command: "record list",
		Args: map[string]any{
			"verbose": false,
		},
	}

	desc := buildCommandDescription(step)
	if strings.Contains(desc, "--verbose") {
		t.Error("false boolean should not appear in description")
	}
}
