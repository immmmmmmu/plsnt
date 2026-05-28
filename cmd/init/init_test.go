package initcmd

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/bootstrap"
	"github.com/immmmmmmu/plsnt/internal/config"
)

func TestRunInit_Claude_InstallsSkills(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer

	err := runInit(options{
		baseDir: dir,
		agent:   bootstrap.AgentClaude,
		stdout:  &out,
		stderr:  &errOut,
	})
	if err != nil {
		t.Fatalf("runInit error: %v", err)
	}

	guide := filepath.Join(dir, ".claude", "skills", "plsnt-guide", "SKILL.md")
	if _, statErr := os.Stat(guide); statErr != nil {
		t.Fatalf("expected %s: %v", guide, statErr)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("Installed")) {
		t.Errorf("expected summary in stderr, got: %s", errOut.String())
	}
}

func TestRunInit_SavesProfile(t *testing.T) {
	dir := t.TempDir()
	cfgHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	var out, errOut bytes.Buffer
	err := runInit(options{
		baseDir:     dir,
		agent:       bootstrap.AgentClaude,
		url:         "https://pleasanter.example.com",
		apiKey:      "secret-key",
		profileName: "default",
		stdout:      &out,
		stderr:      &errOut,
	})
	if err != nil {
		t.Fatalf("runInit error: %v", err)
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		t.Fatalf("config load error: %v", err)
	}
	p, ok := cfg.Profiles["default"]
	if !ok {
		t.Fatalf("profile not saved")
	}
	if p.URL != "https://pleasanter.example.com" || p.APIKey != "secret-key" {
		t.Errorf("profile values wrong: %+v", p)
	}
}

func TestRunInit_WithMCP(t *testing.T) {
	dir := t.TempDir()
	cfgHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	var out, errOut bytes.Buffer
	err := runInit(options{
		baseDir: dir,
		agent:   bootstrap.AgentCodex,
		withMCP: true,
		stdout:  &out,
		stderr:  &errOut,
	})
	if err != nil {
		t.Fatalf("runInit error: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(dir, ".mcp.json")); statErr != nil {
		t.Fatalf("expected .mcp.json: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "AGENTS.md")); statErr != nil {
		t.Fatalf("expected AGENTS.md: %v", statErr)
	}
}

func TestPromptHelpers(t *testing.T) {
	var out bytes.Buffer

	// promptChoice: explicit valid choice.
	r := bufio.NewReader(strings.NewReader("codex\n"))
	if got := promptChoice(r, &out, "Agent", []string{"claude", "codex"}, "claude"); got != "codex" {
		t.Errorf("promptChoice = %q, want codex", got)
	}
	// promptChoice: empty -> default.
	r = bufio.NewReader(strings.NewReader("\n"))
	if got := promptChoice(r, &out, "Agent", []string{"claude", "codex"}, "claude"); got != "claude" {
		t.Errorf("promptChoice empty = %q, want claude", got)
	}
	// promptChoice: invalid -> default.
	r = bufio.NewReader(strings.NewReader("nope\n"))
	if got := promptChoice(r, &out, "Agent", []string{"claude"}, "claude"); got != "claude" {
		t.Errorf("promptChoice invalid = %q, want claude", got)
	}

	// promptString: value and default.
	r = bufio.NewReader(strings.NewReader("https://x\n"))
	if got := promptString(r, &out, "URL", ""); got != "https://x" {
		t.Errorf("promptString = %q", got)
	}
	r = bufio.NewReader(strings.NewReader("\n"))
	if got := promptString(r, &out, "URL", "def"); got != "def" {
		t.Errorf("promptString default = %q", got)
	}

	// promptYesNo.
	r = bufio.NewReader(strings.NewReader("y\n"))
	if !promptYesNo(r, &out, "ok?", false) {
		t.Errorf("promptYesNo y = false")
	}
	r = bufio.NewReader(strings.NewReader("\n"))
	if promptYesNo(r, &out, "ok?", false) {
		t.Errorf("promptYesNo empty = true")
	}
}

func TestNewCmd_NonInteractive(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd := NewCmd()
	cmd.SetArgs([]string{"--yes", "--agent", "codex", "--scope", "project"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(tmp, "AGENTS.md")); statErr != nil {
		t.Fatalf("expected AGENTS.md in cwd: %v", statErr)
	}
}

func TestNewCmd_InvalidAgent(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	cmd := NewCmd()
	cmd.SetArgs([]string{"--yes", "--agent", "bogus"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error for invalid agent")
	}
}

func TestRunInit_HTTPProfileAndMCPProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	var out, errOut bytes.Buffer
	err := runInit(options{
		baseDir:     dir,
		agent:       bootstrap.AgentClaude,
		url:         "http://localhost",
		apiKey:      "k",
		profileName: "production",
		withMCP:     true,
		stdout:      &out,
		stderr:      &errOut,
	})
	if err != nil {
		t.Fatalf("runInit error: %v", err)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("WARNING: Using HTTP")) {
		t.Errorf("expected HTTP warning, got: %s", errOut.String())
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	if !bytes.Contains(data, []byte("production")) {
		t.Errorf(".mcp.json should reference profile 'production': %s", data)
	}
}

func TestResolveBaseDir(t *testing.T) {
	cwd, _ := os.Getwd()
	if got, err := resolveBaseDir("project"); err != nil || got != cwd {
		t.Errorf("project: got %q err %v, want %q", got, err, cwd)
	}
	home, _ := os.UserHomeDir()
	if got, err := resolveBaseDir("user"); err != nil || got != home {
		t.Errorf("user: got %q err %v, want %q", got, err, home)
	}
	if _, err := resolveBaseDir("bogus"); err == nil {
		t.Errorf("bogus scope: expected error")
	}
}
