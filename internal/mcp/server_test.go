package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/immmmmmmu/plsnt/internal/config"
)

func TestNewServer_RegistersTools(t *testing.T) {
	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles: map[string]*config.Profile{
			"test": {URL: "http://localhost", APIKey: "key"},
		},
	}

	// Use nil client — we only check tool registration, not execution.
	srv := NewServer("1.0.0", nil, cfg, "test")
	assert.NotNil(t, srv)
	assert.NotNil(t, srv.MCPServer())
}

func TestNewServer_WithLogFile(t *testing.T) {
	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles: map[string]*config.Profile{
			"test": {URL: "http://localhost", APIKey: "key"},
		},
	}

	srv := NewServer("1.0.0", nil, cfg, "test", WithLogFile("/tmp/test.log"))
	assert.Equal(t, "/tmp/test.log", srv.logFile)
}
