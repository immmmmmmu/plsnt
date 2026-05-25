package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// ExecutionResult holds the outcome of a single step execution.
type ExecutionResult struct {
	StepName string
	Success  bool
	Output   string
	Error    error
	Duration time.Duration
}

// EngineOptions configures the batch execution engine.
type EngineOptions struct {
	DryRun      bool
	LogFile     string
	Writer      io.Writer // output writer (default: os.Stdout)
	CommandName string    // executable name (default: "plsnt")
	Silent      bool      // suppress scaffold summary output
}

// Engine executes batch definitions step by step.
type Engine struct {
	opts        EngineOptions
	log         io.Writer
	stepOutputs map[string]map[string]string // step name -> key -> value from JSON output
}

// NewEngine creates a new batch execution engine.
func NewEngine(opts EngineOptions) (*Engine, error) {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	if opts.CommandName == "" {
		opts.CommandName = "plsnt"
	}

	var logWriter = io.Discard
	if opts.LogFile != "" {
		f, err := os.Create(opts.LogFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create log file %q: %w", opts.LogFile, err)
		}
		logWriter = f
	}

	return &Engine{
		opts:        opts,
		log:         logWriter,
		stepOutputs: make(map[string]map[string]string),
	}, nil
}

// Execute runs all steps in the batch definition in topological order.
func (e *Engine) Execute(ctx context.Context, batch *BatchDef) ([]ExecutionResult, error) {
	sorted, err := TopologicalSort(batch.Steps)
	if err != nil {
		return nil, err
	}

	var results []ExecutionResult

	for _, step := range sorted {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		// Expand step output references (e.g. {{step_name.Key}})
		step = e.expandStepOutputRefs(step)

		result := e.executeStep(step)
		results = append(results, result)

		// Parse JSON output and store for subsequent steps
		if result.Success {
			e.captureStepOutput(step.Name, result.Output)
		}

		e.writeLog(result)

		if !result.Success {
			return results, fmt.Errorf("step %q failed: %v", step.Name, result.Error)
		}
	}

	return results, nil
}

func (e *Engine) executeStep(step StepDef) ExecutionResult {
	start := time.Now()

	cmdDesc := buildCommandDescription(step)

	if e.opts.DryRun {
		msg := fmt.Sprintf("[DRY-RUN] Step %q: plsnt %s", step.Name, cmdDesc)
		fmt.Fprintln(e.opts.Writer, msg)

		return ExecutionResult{
			StepName: step.Name,
			Success:  true,
			Output:   msg,
			Duration: time.Since(start),
		}
	}

	fmt.Fprintf(e.opts.Writer, "[EXECUTE] Step %q: %s %s\n", step.Name, e.opts.CommandName, cmdDesc)

	args := buildCommandArgs(step)
	cmd := exec.CommandContext(context.Background(), e.opts.CommandName, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		combined := stdout.String() + stderr.String()
		return ExecutionResult{
			StepName: step.Name,
			Success:  false,
			Output:   combined,
			Error:    fmt.Errorf("command failed: %w\n%s", err, combined),
			Duration: time.Since(start),
		}
	}

	return ExecutionResult{
		StepName: step.Name,
		Success:  true,
		Output:   stdout.String(),
		Duration: time.Since(start),
	}
}

func buildCommandDescription(step StepDef) string {
	parts := []string{step.Command}

	if len(step.Args) > 0 {
		// Sort keys for deterministic output
		keys := make([]string, 0, len(step.Args))
		for k := range step.Args {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := step.Args[k]
			switch val := v.(type) {
			case bool:
				if val {
					parts = append(parts, fmt.Sprintf("--%s", k))
				}
			case string:
				parts = append(parts, fmt.Sprintf("--%s %q", k, val))
			default:
				parts = append(parts, fmt.Sprintf("--%s %v", k, val))
			}
		}
	}

	return strings.Join(parts, " ")
}

func buildCommandArgs(step StepDef) []string {
	parts := strings.Fields(step.Command)

	// Ensure JSON output for batch automation (needed for step output references)
	hasOutput := false
	if len(step.Args) > 0 {
		keys := make([]string, 0, len(step.Args))
		for k := range step.Args {
			keys = append(keys, k)
			if k == "output" || k == "o" {
				hasOutput = true
			}
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := step.Args[k]
			switch val := v.(type) {
			case bool:
				if val {
					parts = append(parts, fmt.Sprintf("--%s", k))
				}
			case string:
				parts = append(parts, fmt.Sprintf("--%s", k), val)
			default:
				parts = append(parts, fmt.Sprintf("--%s", k), fmt.Sprintf("%v", val))
			}
		}
	}

	if !hasOutput {
		parts = append(parts, "--output", "json")
	}

	return parts
}

func (e *Engine) writeLog(result ExecutionResult) {
	if e.log == io.Discard {
		return
	}
	status := "OK"
	if !result.Success {
		status = "FAIL"
	}
	fmt.Fprintf(e.log, "[%s] %s (%s) %s\n",
		time.Now().Format(time.RFC3339),
		result.StepName,
		result.Duration,
		status,
	)
}

// StepOutputs returns a copy of the captured step outputs.
// Used by summary generation after batch execution.
func (e *Engine) StepOutputs() map[string]map[string]string {
	result := make(map[string]map[string]string, len(e.stepOutputs))
	for k, v := range e.stepOutputs {
		copied := make(map[string]string, len(v))
		for k2, v2 := range v {
			copied[k2] = v2
		}
		result[k] = copied
	}
	return result
}

// Close closes any resources held by the engine (e.g., log file).
func (e *Engine) Close() error {
	if closer, ok := e.log.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// captureStepOutput parses JSON output from a step and stores top-level
// string/number values for use in subsequent step references.
func (e *Engine) captureStepOutput(stepName, output string) {
	// Try to parse output as JSON object
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		return
	}

	values := make(map[string]string, len(parsed))
	for k, raw := range parsed {
		// Try string first
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			values[k] = s
			continue
		}
		// Try number
		var n json.Number
		if err := json.Unmarshal(raw, &n); err == nil {
			values[k] = n.String()
			continue
		}
	}
	if len(values) > 0 {
		e.stepOutputs[stepName] = values
	}
}

// expandStepOutputRefs replaces {{step_name.Key}} in step args with
// captured output values from previously executed steps.
func (e *Engine) expandStepOutputRefs(step StepDef) StepDef {
	if len(e.stepOutputs) == 0 || len(step.Args) == 0 {
		return step
	}

	expanded := StepDef{
		Name:      step.Name,
		Command:   e.expandOutputString(step.Command),
		DependsOn: step.DependsOn,
		Args:      make(map[string]any, len(step.Args)),
	}

	for k, v := range step.Args {
		expanded.Args[k] = e.expandOutputValue(v)
	}
	return expanded
}

func (e *Engine) expandOutputValue(v any) any {
	switch val := v.(type) {
	case string:
		return e.expandOutputString(val)
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, item := range val {
			result[k] = e.expandOutputValue(item)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = e.expandOutputValue(item)
		}
		return result
	default:
		return v
	}
}

func (e *Engine) expandOutputString(s string) string {
	result := s
	for stepName, values := range e.stepOutputs {
		for key, val := range values {
			placeholder := "{{" + stepName + "." + key + "}}"
			result = strings.ReplaceAll(result, placeholder, val)
		}
	}
	return result
}
