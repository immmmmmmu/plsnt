package config

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/immmmmmmu/plsnt/internal/config"
)

func setupTestConfig(t *testing.T, cfg *config.Config) string {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Dir(cfgPath), 0700)
	require.NoError(t, err)

	err = os.WriteFile(cfgPath, data, 0600)
	require.NoError(t, err)

	return cfgPath
}

func TestMCPSetupCmd_StdoutOutput(t *testing.T) {
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {
				URL:    "https://example.com/pleasanter",
				APIKey: "test-api-key-12345",
			},
		},
	})

	// Override config path for testing
	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	// Ensure the config file is at the expected path
	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		err := os.MkdirAll(filepath.Dir(expectedPath), 0700)
		require.NoError(t, err)
		data, _ := os.ReadFile(cfgPath)
		err = os.WriteFile(expectedPath, data, 0600)
		require.NoError(t, err)
	}

	cmd := newMCPSetupCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify valid JSON output
	var parsed map[string]any
	err = json.Unmarshal(stdout.Bytes(), &parsed)
	require.NoError(t, err)

	// Verify structure
	mcpServers, ok := parsed["mcpServers"].(map[string]any)
	require.True(t, ok)

	pleasanter, ok := mcpServers["pleasanter"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "http", pleasanter["type"])
	assert.Equal(t, "https://example.com/pleasanter/mcp", pleasanter["url"])

	headers, ok := pleasanter["headers"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "${PLEASANTER_API_KEY}", headers["X-API-Key"])
}

func TestMCPSetupCmd_NoAPIKeyInOutput(t *testing.T) {
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {
				URL:    "https://example.com",
				APIKey: "super-secret-key-should-not-appear",
			},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	cmd := newMCPSetupCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.NotContains(t, output, "super-secret-key-should-not-appear")
	assert.Contains(t, output, "${PLEASANTER_API_KEY}")
}

func TestMCPSetupCmd_FileOutput(t *testing.T) {
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {
				URL:    "https://example.com",
				APIKey: "test-key",
			},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	outputFile := filepath.Join(t.TempDir(), ".mcp.json")

	cmd := newMCPSetupCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--output", outputFile})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify file was created
	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	mcpServers, ok := parsed["mcpServers"].(map[string]any)
	require.True(t, ok)
	_, ok = mcpServers["pleasanter"].(map[string]any)
	require.True(t, ok)
}

func TestMCPSetupCmd_HTTPWarning(t *testing.T) {
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {
				URL:    "http://localhost",
				APIKey: "test-key",
			},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	cmd := newMCPSetupCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, stderr.String(), "WARNING")
	assert.Contains(t, stderr.String(), "HTTP")
}

func TestMCPSetupCmd_EmbedKey(t *testing.T) {
	apiKey := "embedded-api-key-98765"
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {
				URL:    "https://example.com",
				APIKey: apiKey,
			},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	cmd := newMCPSetupCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--embed-key"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify API key is embedded directly
	var parsed map[string]any
	err = json.Unmarshal(stdout.Bytes(), &parsed)
	require.NoError(t, err)

	mcpServers := parsed["mcpServers"].(map[string]any)
	pleasanter := mcpServers["pleasanter"].(map[string]any)
	headers := pleasanter["headers"].(map[string]any)
	assert.Equal(t, apiKey, headers["X-API-Key"])

	// Should NOT contain env var reference
	assert.NotContains(t, stdout.String(), "${PLEASANTER_API_KEY}")
}

func TestMCPSetupCmd_EmbedKey_Warning(t *testing.T) {
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {
				URL:    "https://example.com",
				APIKey: "test-key",
			},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	cmd := newMCPSetupCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--embed-key"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Should warn about embedded key
	assert.Contains(t, stderr.String(), "WARNING: API key is embedded directly")
	assert.Contains(t, stderr.String(), ".gitignore")
}

func TestMCPSetupCmd_DefaultNoEmbedKey(t *testing.T) {
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {
				URL:    "https://example.com",
				APIKey: "secret-key-should-not-appear",
			},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	cmd := newMCPSetupCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.NoError(t, err)

	// Default: should use env var reference, not embed key
	assert.Contains(t, stdout.String(), "${PLEASANTER_API_KEY}")
	assert.NotContains(t, stdout.String(), "secret-key-should-not-appear")
}

func TestMCPSetupCmd_EmbedKey_FileOutput(t *testing.T) {
	apiKey := "file-embed-key-55555"
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {
				URL:    "https://example.com",
				APIKey: apiKey,
			},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	outputFile := filepath.Join(t.TempDir(), ".mcp.json")

	cmd := newMCPSetupCmd()
	var stderr bytes.Buffer
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--embed-key", "--output", outputFile})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify file contains embedded key
	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	mcpServers := parsed["mcpServers"].(map[string]any)
	pleasanter := mcpServers["pleasanter"].(map[string]any)
	headers := pleasanter["headers"].(map[string]any)
	assert.Equal(t, apiKey, headers["X-API-Key"])

	// Warning should still appear
	assert.Contains(t, stderr.String(), "WARNING: API key is embedded directly")
}

func TestMCPSetupCmd_ProfileFlag(t *testing.T) {
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {
				URL:    "http://localhost",
				APIKey: "default-key",
			},
			"production": {
				URL:    "https://production.example.com",
				APIKey: "prod-key",
			},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	cmd := newMCPSetupCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-p", "production"})

	err := cmd.Execute()
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(stdout.Bytes(), &parsed)
	require.NoError(t, err)

	mcpServers := parsed["mcpServers"].(map[string]any)
	pleasanter := mcpServers["pleasanter"].(map[string]any)
	assert.Equal(t, "https://production.example.com/mcp", pleasanter["url"])
}

func TestMCPSetupCmd_ServerPlsnt(t *testing.T) {
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {URL: "http://localhost", APIKey: "key"},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	cmd := newMCPSetupCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--server", "plsnt"})

	err := cmd.Execute()
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(stdout.Bytes(), &parsed)
	require.NoError(t, err)

	mcpServers := parsed["mcpServers"].(map[string]any)
	plsnt := mcpServers["plsnt"].(map[string]any)
	assert.Equal(t, "plsnt", plsnt["command"])

	args := plsnt["args"].([]any)
	assert.Contains(t, args, "mcp")
	assert.Contains(t, args, "serve")
}

func TestMCPSetupCmd_ServerBoth(t *testing.T) {
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {URL: "https://example.com", APIKey: "key"},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	cmd := newMCPSetupCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--server", "both"})

	err := cmd.Execute()
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(stdout.Bytes(), &parsed)
	require.NoError(t, err)

	mcpServers := parsed["mcpServers"].(map[string]any)
	_, hasPleasanter := mcpServers["pleasanter"]
	_, hasPlsnt := mcpServers["plsnt"]
	assert.True(t, hasPleasanter, "should have pleasanter server")
	assert.True(t, hasPlsnt, "should have plsnt server")
}

func TestMCPSetupCmd_Desktop(t *testing.T) {
	cfgPath := setupTestConfig(t, &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default": {URL: "http://localhost", APIKey: "key"},
		},
	})

	origEnv := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))
	defer os.Setenv("XDG_CONFIG_HOME", origEnv)

	expectedPath := config.DefaultPath()
	if expectedPath != cfgPath {
		os.MkdirAll(filepath.Dir(expectedPath), 0700)
		data, _ := os.ReadFile(cfgPath)
		os.WriteFile(expectedPath, data, 0600)
	}

	cmd := newMCPSetupCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--desktop"})

	err := cmd.Execute()
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(stdout.Bytes(), &parsed)
	require.NoError(t, err)

	mcpServers := parsed["mcpServers"].(map[string]any)
	plsnt := mcpServers["plsnt"].(map[string]any)
	assert.Equal(t, "plsnt", plsnt["command"])
}
