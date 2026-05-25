package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

const mcpAuthHeader = "X-API-Key"

// MCPServerConfig represents the .mcp.json structure for Claude Code.
type MCPServerConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer represents a single MCP server entry.
// For HTTP servers: Type + URL + Headers.
// For stdio servers: Command + Args + Env.
type MCPServer struct {
	Type    string            `json:"type,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// GenerateMCPConfig creates a .mcp.json config from a profile's URL.
// The API key is always output as an environment variable reference.
func GenerateMCPConfig(baseURL string) *MCPServerConfig {
	mcpURL := buildMCPURL(baseURL)

	return &MCPServerConfig{
		MCPServers: map[string]MCPServer{
			"pleasanter": {
				Type: "http",
				URL:  mcpURL,
				Headers: map[string]string{
					mcpAuthHeader: "${PLEASANTER_API_KEY}",
				},
			},
		},
	}
}

// GenerateMCPConfigWithKey creates a .mcp.json config with the API key embedded directly.
func GenerateMCPConfigWithKey(baseURL, apiKey string) *MCPServerConfig {
	mcpURL := buildMCPURL(baseURL)

	return &MCPServerConfig{
		MCPServers: map[string]MCPServer{
			"pleasanter": {
				Type: "http",
				URL:  mcpURL,
				Headers: map[string]string{
					mcpAuthHeader: apiKey,
				},
			},
		},
	}
}

// GeneratePlsntMCPConfig creates a .mcp.json config for the plsnt MCP server (stdio).
func GeneratePlsntMCPConfig() *MCPServerConfig {
	return &MCPServerConfig{
		MCPServers: map[string]MCPServer{
			"plsnt": {
				Command: "plsnt",
				Args:    []string{"mcp", "serve"},
			},
		},
	}
}

// GeneratePlsntMCPConfigWithProfile creates a plsnt MCP config with a specific profile.
func GeneratePlsntMCPConfigWithProfile(profileName string) *MCPServerConfig {
	return &MCPServerConfig{
		MCPServers: map[string]MCPServer{
			"plsnt": {
				Command: "plsnt",
				Args:    []string{"mcp", "serve"},
				Env: map[string]string{
					"PLSNT_PROFILE": profileName,
				},
			},
		},
	}
}

// GenerateBothMCPConfig creates a .mcp.json with both Pleasanter and plsnt MCP servers.
func GenerateBothMCPConfig(baseURL string) *MCPServerConfig {
	mcpURL := buildMCPURL(baseURL)

	return &MCPServerConfig{
		MCPServers: map[string]MCPServer{
			"pleasanter": {
				Type: "http",
				URL:  mcpURL,
				Headers: map[string]string{
					mcpAuthHeader: "${PLEASANTER_API_KEY}",
				},
			},
			"plsnt": {
				Command: "plsnt",
				Args:    []string{"mcp", "serve"},
			},
		},
	}
}

// GenerateDesktopConfig creates a Claude Desktop config for the plsnt MCP server.
func GenerateDesktopConfig() *MCPServerConfig {
	return GeneratePlsntMCPConfig()
}

// MarshalJSON outputs the config as indented JSON.
func (c *MCPServerConfig) MarshalJSON() ([]byte, error) {
	type Alias MCPServerConfig
	return json.MarshalIndent((*Alias)(c), "", "  ")
}

// buildMCPURL appends /mcp to the base URL, handling trailing slash.
func buildMCPURL(baseURL string) string {
	return fmt.Sprintf("%s/mcp", strings.TrimRight(baseURL, "/"))
}

// IsHTTP returns true if the URL uses plain HTTP (not HTTPS).
func IsHTTP(url string) bool {
	return strings.HasPrefix(url, "http://")
}
