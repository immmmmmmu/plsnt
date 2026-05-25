package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMCPURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{
			name:     "without trailing slash",
			baseURL:  "http://example.com",
			expected: "http://example.com/mcp",
		},
		{
			name:     "with trailing slash",
			baseURL:  "http://example.com/",
			expected: "http://example.com/mcp",
		},
		{
			name:     "with subpath without trailing slash",
			baseURL:  "https://example.com/pleasanter",
			expected: "https://example.com/pleasanter/mcp",
		},
		{
			name:     "with subpath with trailing slash",
			baseURL:  "https://example.com/pleasanter/",
			expected: "https://example.com/pleasanter/mcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildMCPURL(tt.baseURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsHTTP(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "http URL",
			url:      "http://example.com",
			expected: true,
		},
		{
			name:     "https URL",
			url:      "https://example.com",
			expected: false,
		},
		{
			name:     "http with port",
			url:      "http://localhost:8080",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHTTP(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateMCPConfig(t *testing.T) {
	config := GenerateMCPConfig("http://example.com")

	require.NotNil(t, config)
	require.Contains(t, config.MCPServers, "pleasanter")

	server := config.MCPServers["pleasanter"]
	assert.Equal(t, "http", server.Type)
	assert.Equal(t, "http://example.com/mcp", server.URL)
	assert.Equal(t, "${PLEASANTER_API_KEY}", server.Headers[mcpAuthHeader])
}

func TestGenerateMCPConfig_NoRealAPIKey(t *testing.T) {
	config := GenerateMCPConfig("http://example.com")

	jsonBytes, err := json.Marshal(config)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	// Should contain environment variable reference, not a real key
	assert.Contains(t, jsonStr, "${PLEASANTER_API_KEY}")
	// Should not contain any real API key patterns
	assert.NotContains(t, jsonStr, "sk-")
	assert.NotContains(t, jsonStr, "api_key")
}

func TestGenerateMCPConfigWithKey_EmbedsAPIKey(t *testing.T) {
	apiKey := "my-secret-api-key-12345"
	cfg := GenerateMCPConfigWithKey("http://example.com", apiKey)

	require.NotNil(t, cfg)
	require.Contains(t, cfg.MCPServers, "pleasanter")

	server := cfg.MCPServers["pleasanter"]
	assert.Equal(t, "http", server.Type)
	assert.Equal(t, "http://example.com/mcp", server.URL)
	assert.Equal(t, apiKey, server.Headers[mcpAuthHeader])
}

func TestGenerateMCPConfigWithKey_DoesNotContainEnvRef(t *testing.T) {
	apiKey := "real-key-value"
	cfg := GenerateMCPConfigWithKey("http://example.com", apiKey)

	jsonBytes, err := json.Marshal(cfg)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.NotContains(t, jsonStr, "${PLEASANTER_API_KEY}")
	assert.Contains(t, jsonStr, apiKey)
}

func TestMCPServerConfig_MarshalJSON(t *testing.T) {
	config := GenerateMCPConfig("https://example.com/pleasanter")

	jsonBytes, err := config.MarshalJSON()
	require.NoError(t, err)

	// Verify it's valid JSON
	var parsed map[string]any
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)

	// Verify structure
	mcpServers, ok := parsed["mcpServers"].(map[string]any)
	require.True(t, ok, "mcpServers should be a map")

	pleasanter, ok := mcpServers["pleasanter"].(map[string]any)
	require.True(t, ok, "pleasanter should be a map")

	assert.Equal(t, "http", pleasanter["type"])
	assert.Equal(t, "https://example.com/pleasanter/mcp", pleasanter["url"])

	headers, ok := pleasanter["headers"].(map[string]any)
	require.True(t, ok, "headers should be a map")
	assert.Equal(t, "${PLEASANTER_API_KEY}", headers["X-API-Key"])
}
