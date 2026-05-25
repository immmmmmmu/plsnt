package batch

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

// BatchDef represents the top-level batch YAML definition.
type BatchDef struct {
	Name      string            `yaml:"name"`
	Variables map[string]string `yaml:"variables,omitempty"`
	Steps     []StepDef         `yaml:"steps"`
}

// StepDef represents a single step in a batch definition.
type StepDef struct {
	Name      string         `yaml:"name"`
	Command   string         `yaml:"command"`
	DependsOn []string       `yaml:"depends_on,omitempty"`
	Args      map[string]any `yaml:"args,omitempty"`
}

// ParseFile reads a YAML batch file, validates it, and expands variables.
func ParseFile(path string) (*BatchDef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.New(errs.CodeInvalidInput, fmt.Sprintf("failed to read batch file: %v", err))
	}

	return Parse(data)
}

// Parse parses YAML bytes into a BatchDef, validates, and expands variables.
func Parse(data []byte) (*BatchDef, error) {
	batch, err := ParseRaw(data)
	if err != nil {
		return nil, err
	}

	batch.ExpandVariables()

	return batch, nil
}

// ParseRaw parses and validates YAML bytes without expanding variables.
func ParseRaw(data []byte) (*BatchDef, error) {
	var batch BatchDef
	if err := yaml.Unmarshal(data, &batch); err != nil {
		return nil, errs.New(errs.CodeInvalidInput, fmt.Sprintf("failed to parse batch YAML: %v", err))
	}

	if err := batch.Validate(); err != nil {
		return nil, err
	}

	return &batch, nil
}

// Validate checks the batch definition for structural errors.
func (b *BatchDef) Validate() error {
	if len(b.Steps) == 0 {
		return errs.New(errs.CodeValidationError, "batch definition must have at least one step")
	}

	// Check for duplicate step names
	seen := make(map[string]bool, len(b.Steps))
	for _, step := range b.Steps {
		if step.Name == "" {
			return errs.New(errs.CodeValidationError, "step name must not be empty")
		}
		if seen[step.Name] {
			return errs.New(errs.CodeValidationError, fmt.Sprintf("duplicate step name: %q", step.Name))
		}
		seen[step.Name] = true
	}

	// Check that all depends_on references exist
	for _, step := range b.Steps {
		for _, dep := range step.DependsOn {
			if !seen[dep] {
				return errs.New(errs.CodeValidationError, fmt.Sprintf("step %q depends on unknown step %q", step.Name, dep))
			}
		}
	}

	// Check for circular dependencies via topological sort
	if _, err := TopologicalSort(b.Steps); err != nil {
		return err
	}

	return nil
}

// ExpandVariables replaces {{var_name}} placeholders in all step args with
// variable values from the batch definition.
func (b *BatchDef) ExpandVariables() {
	if len(b.Variables) == 0 {
		return
	}

	for i := range b.Steps {
		b.Steps[i].Command = expandString(b.Steps[i].Command, b.Variables)
		b.Steps[i].Args = expandArgsMap(b.Steps[i].Args, b.Variables)
	}
}

func expandArgsMap(args map[string]any, vars map[string]string) map[string]any {
	if args == nil {
		return nil
	}

	result := make(map[string]any, len(args))
	for k, v := range args {
		result[k] = expandValue(v, vars)
	}
	return result
}

func expandValue(v any, vars map[string]string) any {
	switch val := v.(type) {
	case string:
		return expandString(val, vars)
	case map[string]any:
		return expandArgsMap(val, vars)
	case []any:
		expanded := make([]any, len(val))
		for i, item := range val {
			expanded[i] = expandValue(item, vars)
		}
		return expanded
	default:
		return v
	}
}

func expandString(s string, vars map[string]string) string {
	result := s
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}

// TopologicalSort orders steps respecting depends_on constraints using Kahn's algorithm.
// Returns an error if a cycle is detected.
func TopologicalSort(steps []StepDef) ([]StepDef, error) {
	nameToIdx := make(map[string]int, len(steps))
	for i, s := range steps {
		nameToIdx[s.Name] = i
	}

	// Build in-degree map and adjacency list
	inDegree := make([]int, len(steps))
	// adj[i] = list of step indices that depend on step i
	adj := make([][]int, len(steps))
	for i := range adj {
		adj[i] = []int{}
	}

	for i, s := range steps {
		inDegree[i] = len(s.DependsOn)
		for _, dep := range s.DependsOn {
			depIdx := nameToIdx[dep]
			adj[depIdx] = append(adj[depIdx], i)
		}
	}

	// Collect nodes with in-degree 0
	var queue []int
	for i, d := range inDegree {
		if d == 0 {
			queue = append(queue, i)
		}
	}

	var sorted []StepDef
	for len(queue) > 0 {
		idx := queue[0]
		queue = queue[1:]
		sorted = append(sorted, steps[idx])

		for _, next := range adj[idx] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(sorted) != len(steps) {
		return nil, errs.New(errs.CodeValidationError, "circular dependency detected in batch steps")
	}

	return sorted, nil
}
