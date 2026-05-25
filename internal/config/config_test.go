package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPath_WithXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")
	path := DefaultPath()
	expected := "/tmp/xdg-test/plsnt/config.yaml"
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestDefaultPath_WithoutXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	path := DefaultPath()
	if path == "" {
		t.Error("DefaultPath should return a non-empty path")
	}
	// Should end with .config/plsnt/config.yaml
	if !strings.HasSuffix(path, ".config/plsnt/config.yaml") {
		t.Errorf("expected path ending with .config/plsnt/config.yaml, got %q", path)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	// Use YAML that will fail to unmarshal into Config struct: tabs in wrong places
	if err := os.WriteFile(path, []byte("current_profile:\n\t- [invalid\n\t  broken"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	_, err := Load(path)
	if err == nil {
		t.Error("Load should fail on invalid YAML")
	}
}

func TestLoadNilProfiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	// Valid YAML but no profiles key
	if err := os.WriteFile(path, []byte("current_profile: test\n"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Profiles == nil {
		t.Error("Profiles should be initialized even when not in YAML")
	}
}

func TestActiveProfile_DefaultFallback(t *testing.T) {
	t.Setenv("PLSNT_PROFILE", "")
	cfg := &Config{
		CurrentProfile: "",
		Profiles: map[string]*Profile{
			"default": {URL: "http://default.example.com"},
		},
	}
	p, name, err := cfg.ActiveProfile()
	if err != nil {
		t.Fatalf("ActiveProfile failed: %v", err)
	}
	if name != "default" {
		t.Errorf("expected name 'default', got %q", name)
	}
	if p.URL != "http://default.example.com" {
		t.Errorf("expected default URL, got %q", p.URL)
	}
}

func TestResolve_ConfigOnly(t *testing.T) {
	t.Setenv("PLSNT_URL", "")
	t.Setenv("PLSNT_API_KEY", "")
	p := &Profile{
		URL:        "http://config.example.com",
		APIKey:     "config-key",
		APIVersion: "2.0",
	}
	url, apiKey, apiVersion := p.Resolve()
	if url != "http://config.example.com" {
		t.Errorf("expected config URL, got %q", url)
	}
	if apiKey != "config-key" {
		t.Errorf("expected config API key, got %q", apiKey)
	}
	if apiVersion != "2.0" {
		t.Errorf("expected version 2.0, got %q", apiVersion)
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/tmp/plsnt-test-nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("Load should not error on missing file: %v", err)
	}
	if cfg.Profiles == nil {
		t.Error("Profiles should be initialized")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		CurrentProfile: "test",
		Profiles: map[string]*Profile{
			"test": {
				URL:        "http://localhost",
				APIKey:     "secret-key-12345",
				APIVersion: "1.1",
			},
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected permissions 0600, got %o", perm)
	}

	// Check directory permissions
	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat dir failed: %v", err)
	}
	// Note: TempDir may have broader perms; we only verify our subdirectory is accessible
	_ = dirInfo.Mode().Perm()

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.CurrentProfile != "test" {
		t.Errorf("expected current_profile 'test', got %q", loaded.CurrentProfile)
	}

	p, ok := loaded.Profiles["test"]
	if !ok {
		t.Fatal("profile 'test' not found")
	}
	if p.URL != "http://localhost" {
		t.Errorf("expected URL 'http://localhost', got %q", p.URL)
	}
	if p.APIKey != "secret-key-12345" {
		t.Errorf("expected APIKey to be preserved")
	}
}

func TestActiveProfile(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "prod",
		Profiles: map[string]*Profile{
			"prod": {URL: "https://prod.example.com"},
		},
	}

	p, name, err := cfg.ActiveProfile()
	if err != nil {
		t.Fatalf("ActiveProfile failed: %v", err)
	}
	if name != "prod" {
		t.Errorf("expected name 'prod', got %q", name)
	}
	if p.URL != "https://prod.example.com" {
		t.Errorf("expected prod URL, got %q", p.URL)
	}
}

func TestActiveProfile_EnvOverride(t *testing.T) {
	t.Setenv("PLSNT_PROFILE", "staging")

	cfg := &Config{
		CurrentProfile: "prod",
		Profiles: map[string]*Profile{
			"prod":    {URL: "https://prod.example.com"},
			"staging": {URL: "https://staging.example.com"},
		},
	}

	p, name, err := cfg.ActiveProfile()
	if err != nil {
		t.Fatalf("ActiveProfile failed: %v", err)
	}
	if name != "staging" {
		t.Errorf("expected name 'staging', got %q", name)
	}
	if p.URL != "https://staging.example.com" {
		t.Errorf("expected staging URL, got %q", p.URL)
	}
}

func TestActiveProfile_NotFound(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "missing",
		Profiles:       map[string]*Profile{},
	}

	_, _, err := cfg.ActiveProfile()
	if err == nil {
		t.Error("expected error for missing profile")
	}
}

func TestResolve_EnvOverride(t *testing.T) {
	t.Setenv("PLSNT_URL", "http://override.example.com")
	t.Setenv("PLSNT_API_KEY", "env-key")

	p := &Profile{
		URL:        "http://config.example.com",
		APIKey:     "config-key",
		APIVersion: "1.1",
	}

	url, apiKey, apiVersion := p.Resolve()
	if url != "http://override.example.com" {
		t.Errorf("expected env URL, got %q", url)
	}
	if apiKey != "env-key" {
		t.Errorf("expected env API key, got %q", apiKey)
	}
	if apiVersion != "1.1" {
		t.Errorf("expected version 1.1, got %q", apiVersion)
	}
}

func TestResolve_DefaultVersion(t *testing.T) {
	p := &Profile{URL: "http://test.com", APIKey: "key"}
	_, _, version := p.Resolve()
	if version != "1.1" {
		t.Errorf("expected default version 1.1, got %q", version)
	}
}

func TestSaveCreatesFileWith0600(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "secure", "nested")
	path := filepath.Join(subdir, "config.yaml")

	cfg := &Config{
		CurrentProfile: "test",
		Profiles: map[string]*Profile{
			"test": {URL: "http://localhost", APIKey: "key123456789"},
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected file permissions 0600, got %04o", perm)
	}

	dirInfo, err := os.Stat(subdir)
	if err != nil {
		t.Fatalf("Stat dir failed: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0700 {
		t.Errorf("expected directory permissions 0700, got %04o", perm)
	}
}

func TestLoadWarnsOnPermissivePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		CurrentProfile: "test",
		Profiles: map[string]*Profile{
			"test": {URL: "http://localhost", APIKey: "key123456789"},
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Make the file world-readable
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe failed: %v", err)
	}
	os.Stderr = w

	_, loadErr := Load(path)

	w.Close()
	os.Stderr = oldStderr

	if loadErr != nil {
		t.Fatalf("Load failed: %v", loadErr)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}
	r.Close()

	output := buf.String()
	expected := fmt.Sprintf("WARNING: config file %s has permissions 0644, should be 0600\n", path)
	if output != expected {
		t.Errorf("expected stderr %q, got %q", expected, output)
	}
}

func TestLoadNoWarningOnSecurePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		CurrentProfile: "test",
		Profiles: map[string]*Profile{
			"test": {URL: "http://localhost", APIKey: "key123456789"},
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe failed: %v", err)
	}
	os.Stderr = w

	_, loadErr := Load(path)

	w.Close()
	os.Stderr = oldStderr

	if loadErr != nil {
		t.Fatalf("Load failed: %v", loadErr)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}
	r.Close()

	if output := buf.String(); output != "" {
		t.Errorf("expected no stderr output for 0600 file, got %q", output)
	}
}

func TestEnsureSecureWrite(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "a", "b", "c")
	path := filepath.Join(subdir, "data.txt")

	data := []byte("sensitive data")
	if err := ensureSecureWrite(path, data); err != nil {
		t.Fatalf("ensureSecureWrite failed: %v", err)
	}

	// Check file content
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("expected %q, got %q", data, got)
	}

	// Check file permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected file permissions 0600, got %04o", perm)
	}

	// Check directory permissions
	dirInfo, err := os.Stat(subdir)
	if err != nil {
		t.Fatalf("Stat dir failed: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0700 {
		t.Errorf("expected directory permissions 0700, got %04o", perm)
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abcdefghijklmnop", "****...mnop"},
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "****...6789"},
		{"", "****"},
	}
	for _, tt := range tests {
		got := MaskAPIKey(tt.input)
		if got != tt.expected {
			t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
