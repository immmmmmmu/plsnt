package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

func TestNewCmd_HasDeploySubcommand(t *testing.T) {
	cmd := NewCmd()
	assert.Equal(t, "workflow", cmd.Use)
	assert.NotEmpty(t, cmd.Long)

	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "deploy" {
			found = true
			break
		}
	}
	assert.True(t, found, "workflow should have deploy subcommand")
}

func TestDeployCmd_MissingFlags(t *testing.T) {
	// When both required flags are missing, PreRunE catches template="" first
	cmd := NewCmd()
	cmd.SetArgs([]string{"deploy"})
	err := cmd.Execute()
	require.Error(t, err)
	// Either PreRunE validation or cobra required flag check should fire
	assert.Error(t, err)
}

func TestDeployCmd_FolderIDZeroValidation(t *testing.T) {
	// Create a temp template to pass template validation
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "templates", "workflow"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "templates", "workflow", "test.yaml"), []byte("name: test\nsteps:\n  - name: s1\n    command: site create\n"), 0o644))

	// Change to tmpDir so template resolution works
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	cmd := NewCmd()
	cmd.SetArgs([]string{"deploy", "--template", "test", "--folder-id", "0"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "folder-id must be a positive integer")
}

func TestDeployCmd_FolderIDZeroIncludesTenantRootHint(t *testing.T) {
	// Issue #12: folder-id=0 must surface the tenant root limitation hint
	// so users know to create a parent folder via Web UI first.
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "templates", "workflow"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "templates", "workflow", "test.yaml"), []byte("name: test\nsteps:\n  - name: s1\n    command: site create\n"), 0o644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	cmd := NewCmd()
	cmd.SetArgs([]string{"deploy", "--template", "test", "--folder-id", "0"})
	err := cmd.Execute()
	require.Error(t, err)

	cliErr, ok := err.(*errs.CLIError)
	require.True(t, ok, "expected *errs.CLIError, got %T", err)
	assert.Contains(t, cliErr.ErrorBody.Suggestion, "tenant root")
	assert.Contains(t, cliErr.ErrorBody.Suggestion, "Web UI")
	assert.Contains(t, cliErr.ErrorBody.Suggestion, "https://pleasanter.org/en/manual/api-site-create")
	assert.Contains(t, cliErr.ErrorBody.Suggestion, "--folder-id")
}

func TestDeployCmd_TemplateNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	cmd := NewCmd()
	cmd.SetArgs([]string{"deploy", "--template", "nonexistent", "--folder-id", "12345"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeployCmd_InvalidSetFormat(t *testing.T) {
	// Create a temp template
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "templates", "workflow"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "templates", "workflow", "test.yaml"), []byte("name: test\nsteps:\n  - name: s1\n    command: site create\n"), 0o644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	cmd := NewCmd()
	cmd.SetArgs([]string{"deploy", "--template", "test", "--folder-id", "12345", "--set", "no-equals-sign"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --set format")
}

func TestDeployCmd_Flags(t *testing.T) {
	cmd := newDeployCmd()

	// Verify all expected flags exist
	assert.NotNil(t, cmd.Flags().Lookup("template"))
	assert.NotNil(t, cmd.Flags().Lookup("folder-id"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("set"))

	// Verify shorthand flags
	assert.Equal(t, "t", cmd.Flags().Lookup("template").Shorthand)
	assert.Equal(t, "f", cmd.Flags().Lookup("folder-id").Shorthand)
	assert.Equal(t, "s", cmd.Flags().Lookup("set").Shorthand)
}

// --- TASK-0010: Template resolution + batch engine integration tests ---

func TestDeployCmd_TemplateResolution(t *testing.T) {
	// テンプレート名から templates/workflow/{name}.yaml を解決できること
	tmpDir := t.TempDir()
	tmplDir := filepath.Join(tmpDir, "templates", "workflow")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))

	// Create a valid batch YAML template
	tmplYAML := `name: test-template
variables:
  folder_id: "0"
steps:
  - name: create-test
    command: site create
    args:
      parent-id: "{{folder_id}}"
      json: '{"Title":"Test","ReferenceType":"Results","SiteSettings":{}}'
`
	require.NoError(t, os.WriteFile(filepath.Join(tmplDir, "mytemplate.yaml"), []byte(tmplYAML), 0o644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	// resolveTemplatePath should find the template
	path, err := resolveTemplatePath("mytemplate")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("templates", "workflow", "mytemplate.yaml"), path)
}

func TestDeployCmd_TemplateResolutionNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	_, err := resolveTemplatePath("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeployCmd_DryRun(t *testing.T) {
	// --dry-run でテーブル作成せずに計画を表示
	tmpDir := t.TempDir()
	tmplDir := filepath.Join(tmpDir, "templates", "workflow")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))

	tmplYAML := `name: dry-run-test
variables:
  folder_id: "0"
steps:
  - name: create-dept
    command: site create
    args:
      parent-id: "{{folder_id}}"
      json: '{"Title":"部署マスタ","ReferenceType":"Results","SiteSettings":{}}'
  - name: create-position
    command: site create
    args:
      parent-id: "{{folder_id}}"
      json: '{"Title":"役職マスタ","ReferenceType":"Results","SiteSettings":{}}'
`
	require.NoError(t, os.WriteFile(filepath.Join(tmplDir, "drytest.yaml"), []byte(tmplYAML), 0o644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	cmd := NewCmd()
	buf := &safeBuffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deploy", "--template", "drytest", "--folder-id", "99999", "--dry-run"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// DryRun should show step names
	assert.Contains(t, output, "create-dept")
	assert.Contains(t, output, "create-position")
	// DryRun should show DRY-RUN marker
	assert.Contains(t, output, "DRY-RUN")
}

func TestDeployCmd_DryRunWithSetVars(t *testing.T) {
	// --set で変数を上書きしてfolder_idを渡せる
	tmpDir := t.TempDir()
	tmplDir := filepath.Join(tmpDir, "templates", "workflow")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))

	tmplYAML := `name: set-test
variables:
  folder_id: "0"
  custom_var: "default"
steps:
  - name: create-test
    command: site create
    args:
      parent-id: "{{folder_id}}"
      json: '{"Title":"{{custom_var}}","ReferenceType":"Results","SiteSettings":{}}'
`
	require.NoError(t, os.WriteFile(filepath.Join(tmplDir, "settest.yaml"), []byte(tmplYAML), 0o644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	cmd := NewCmd()
	buf := &safeBuffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deploy", "--template", "settest", "--folder-id", "55555",
		"--set", "custom_var=overridden", "--dry-run"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// folder_id should be overridden to 55555
	assert.Contains(t, output, "55555")
	// custom_var should be overridden
	assert.Contains(t, output, "overridden")
}

func TestDeployCmd_DryRunOutputJSON(t *testing.T) {
	// --dry-run の結果がJSON形式で出力される
	tmpDir := t.TempDir()
	tmplDir := filepath.Join(tmpDir, "templates", "workflow")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))

	tmplYAML := `name: json-output-test
variables:
  folder_id: "0"
steps:
  - name: create-single
    command: site create
    args:
      parent-id: "{{folder_id}}"
      json: '{"Title":"テスト","ReferenceType":"Results","SiteSettings":{}}'
`
	require.NoError(t, os.WriteFile(filepath.Join(tmplDir, "jsontest.yaml"), []byte(tmplYAML), 0o644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	cmd := NewCmd()
	buf := &safeBuffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deploy", "--template", "jsontest", "--folder-id", "12345", "--dry-run"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should contain JSON result at the end
	// Parse the last JSON object from output
	var result DeployResult
	// Find the JSON portion (after DRY-RUN lines)
	lines := splitLines(output)
	var jsonPart string
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] != "" && lines[i][0] == '{' {
			jsonPart = lines[i]
			break
		}
	}
	require.NotEmpty(t, jsonPart, "should contain JSON output, got: %s", output)
	err = json.Unmarshal([]byte(jsonPart), &result)
	require.NoError(t, err, "JSON output should be parsable, got: %s", jsonPart)
	assert.Equal(t, "json-output-test", result.TemplateName)
	assert.Equal(t, int64(12345), result.FolderID)
	assert.True(t, result.DryRun)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, "create-single", result.Steps[0].Name)
}

func TestDeployCmd_FolderIDOverridesVariable(t *testing.T) {
	// --folder-id の値が variables の folder_id を上書きする
	tmpDir := t.TempDir()
	tmplDir := filepath.Join(tmpDir, "templates", "workflow")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))

	tmplYAML := `name: folder-override
variables:
  folder_id: "99999"
steps:
  - name: create-test
    command: site create
    args:
      parent-id: "{{folder_id}}"
      json: '{"Title":"Test","ReferenceType":"Results","SiteSettings":{}}'
`
	require.NoError(t, os.WriteFile(filepath.Join(tmplDir, "foldertest.yaml"), []byte(tmplYAML), 0o644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	cmd := NewCmd()
	buf := &safeBuffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deploy", "--template", "foldertest", "--folder-id", "11111", "--dry-run"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should use 11111 (from --folder-id), not 99999 (from template default)
	assert.Contains(t, output, "11111")
	assert.NotContains(t, output, "99999")
}

// --- helpers ---

// safeBuffer is a simple bytes.Buffer wrapper for test output capture
type safeBuffer struct {
	data []byte
}

func (b *safeBuffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *safeBuffer) String() string {
	return string(b.data)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
