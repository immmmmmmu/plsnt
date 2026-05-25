package batch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse_ValidYAML(t *testing.T) {
	yaml := `
name: "test-batch"
variables:
  site_id: "123"
  category: "active"
steps:
  - name: "step1"
    command: "record list"
    args:
      site-id: "{{site_id}}"
      view: '{"ColumnFilterHash":{"ClassA":"{{category}}"}}'
      all-pages: true
  - name: "step2"
    command: "record update"
    depends_on: ["step1"]
    args:
      json: '{"Status": 200}'
  - name: "step3"
    command: "site get"
    depends_on: ["step1"]
    args:
      site-id: "{{site_id}}"
`
	batch, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if batch.Name != "test-batch" {
		t.Errorf("expected name %q, got %q", "test-batch", batch.Name)
	}

	if len(batch.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(batch.Steps))
	}

	// Variables should be expanded
	siteID, ok := batch.Steps[0].Args["site-id"].(string)
	if !ok || siteID != "123" {
		t.Errorf("expected site-id to be expanded to %q, got %v", "123", batch.Steps[0].Args["site-id"])
	}

	view, ok := batch.Steps[0].Args["view"].(string)
	if !ok || !strings.Contains(view, "active") {
		t.Errorf("expected view to contain expanded variable, got %v", view)
	}

	step3SiteID, ok := batch.Steps[2].Args["site-id"].(string)
	if !ok || step3SiteID != "123" {
		t.Errorf("expected step3 site-id to be %q, got %v", "123", batch.Steps[2].Args["site-id"])
	}
}

func TestParse_NoSteps(t *testing.T) {
	yaml := `
name: "empty"
steps: []
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for empty steps")
	}
	if !strings.Contains(err.Error(), "at least one step") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParse_MissingSteps(t *testing.T) {
	yaml := `
name: "no-steps"
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for missing steps")
	}
}

func TestParse_DuplicateStepNames(t *testing.T) {
	yaml := `
name: "dup"
steps:
  - name: "step1"
    command: "record list"
  - name: "step1"
    command: "record get"
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for duplicate step names")
	}
	if !strings.Contains(err.Error(), "duplicate step name") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParse_EmptyStepName(t *testing.T) {
	yaml := `
name: "empty-name"
steps:
  - name: ""
    command: "record list"
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for empty step name")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParse_InvalidDependsOn(t *testing.T) {
	yaml := `
name: "bad-dep"
steps:
  - name: "step1"
    command: "record list"
    depends_on: ["nonexistent"]
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for invalid depends_on reference")
	}
	if !strings.Contains(err.Error(), "unknown step") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParse_CircularDependency(t *testing.T) {
	yaml := `
name: "circular"
steps:
  - name: "a"
    command: "cmd1"
    depends_on: ["b"]
  - name: "b"
    command: "cmd2"
    depends_on: ["a"]
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for circular dependency")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParse_ThreeNodeCycle(t *testing.T) {
	yaml := `
name: "cycle3"
steps:
  - name: "a"
    command: "cmd1"
    depends_on: ["c"]
  - name: "b"
    command: "cmd2"
    depends_on: ["a"]
  - name: "c"
    command: "cmd3"
    depends_on: ["b"]
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for 3-node circular dependency")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	_, err := Parse([]byte(`{invalid yaml`))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParse_VariableExpansionInNestedArgs(t *testing.T) {
	yaml := `
name: "nested"
variables:
  val: "expanded"
steps:
  - name: "step1"
    command: "test"
    args:
      simple: "{{val}}"
      no-var: "plain"
`
	batch, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v := batch.Steps[0].Args["simple"].(string); v != "expanded" {
		t.Errorf("expected %q, got %q", "expanded", v)
	}

	if v := batch.Steps[0].Args["no-var"].(string); v != "plain" {
		t.Errorf("expected %q, got %q", "plain", v)
	}
}

func TestParse_NoVariables(t *testing.T) {
	yaml := `
name: "no-vars"
steps:
  - name: "step1"
    command: "record list"
    args:
      site-id: "456"
`
	batch, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := batch.Steps[0].Args["site-id"].(string); v != "456" {
		t.Errorf("expected %q, got %q", "456", v)
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/batch.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestParseFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "batch.yaml")
	content := `
name: "file-test"
steps:
  - name: "s1"
    command: "site get"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	batch, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if batch.Name != "file-test" {
		t.Errorf("expected name %q, got %q", "file-test", batch.Name)
	}
}

func TestTopologicalSort_LinearChain(t *testing.T) {
	steps := []StepDef{
		{Name: "c", Command: "cmd3", DependsOn: []string{"b"}},
		{Name: "b", Command: "cmd2", DependsOn: []string{"a"}},
		{Name: "a", Command: "cmd1"},
	}

	sorted, err := TopologicalSort(steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// a must come before b, b before c
	idx := make(map[string]int)
	for i, s := range sorted {
		idx[s.Name] = i
	}

	if idx["a"] >= idx["b"] {
		t.Errorf("expected a before b, got a=%d b=%d", idx["a"], idx["b"])
	}
	if idx["b"] >= idx["c"] {
		t.Errorf("expected b before c, got b=%d c=%d", idx["b"], idx["c"])
	}
}

func TestTopologicalSort_ParallelSteps(t *testing.T) {
	steps := []StepDef{
		{Name: "a", Command: "cmd1"},
		{Name: "b", Command: "cmd2"},
		{Name: "c", Command: "cmd3"},
	}

	sorted, err := TopologicalSort(steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("expected 3 steps, got %d", len(sorted))
	}
}

func TestTopologicalSort_DiamondDependency(t *testing.T) {
	// a -> b, a -> c, b -> d, c -> d
	steps := []StepDef{
		{Name: "d", Command: "cmd4", DependsOn: []string{"b", "c"}},
		{Name: "b", Command: "cmd2", DependsOn: []string{"a"}},
		{Name: "c", Command: "cmd3", DependsOn: []string{"a"}},
		{Name: "a", Command: "cmd1"},
	}

	sorted, err := TopologicalSort(steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	idx := make(map[string]int)
	for i, s := range sorted {
		idx[s.Name] = i
	}

	if idx["a"] >= idx["b"] || idx["a"] >= idx["c"] {
		t.Error("a must come before b and c")
	}
	if idx["b"] >= idx["d"] || idx["c"] >= idx["d"] {
		t.Error("b and c must come before d")
	}
}

func TestTopologicalSort_CycleDetection(t *testing.T) {
	steps := []StepDef{
		{Name: "a", Command: "cmd1", DependsOn: []string{"b"}},
		{Name: "b", Command: "cmd2", DependsOn: []string{"a"}},
	}

	_, err := TopologicalSort(steps)
	if err == nil {
		t.Fatal("expected error for cycle")
	}
}

func TestParseRaw_DoesNotExpandVariables(t *testing.T) {
	yaml := `
name: "raw"
variables:
  val: "hello"
steps:
  - name: "s1"
    command: "test"
    args:
      key: "{{val}}"
`
	batch, err := ParseRaw([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ParseRaw should NOT expand variables
	if v := batch.Steps[0].Args["key"].(string); v != "{{val}}" {
		t.Errorf("expected unexpanded %q, got %q", "{{val}}", v)
	}
}
