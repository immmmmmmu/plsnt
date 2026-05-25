package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/immmmmmmu/plsnt/internal/config"
)

func TestHandleProfilesResource(t *testing.T) {
	cfg := &config.Config{
		CurrentProfile: "default",
		Profiles: map[string]*config.Profile{
			"default":    {URL: "http://localhost", APIKey: "key1"},
			"production": {URL: "https://prod.example.com", APIKey: "key2"},
		},
	}

	srv := NewServer("test", nil, cfg, "default")

	request := mcp.ReadResourceRequest{}
	request.Params.URI = "plsnt://profiles"

	contents, err := srv.handleProfilesResource(context.Background(), request)
	require.NoError(t, err)
	require.Len(t, contents, 1)

	text := contents[0].(mcp.TextResourceContents).Text
	assert.Contains(t, text, "default")
	assert.Contains(t, text, "production")
	assert.Contains(t, text, "http://localhost")
}

func TestHandleTemplatesResource_NoDir(t *testing.T) {
	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles:       map[string]*config.Profile{"test": {URL: "http://localhost", APIKey: "key"}},
	}

	srv := NewServer("test", nil, cfg, "test")

	request := mcp.ReadResourceRequest{}
	request.Params.URI = "plsnt://templates"

	// templates/ directory may not exist in test context — should return empty list.
	contents, err := srv.handleTemplatesResource(context.Background(), request)
	require.NoError(t, err)
	require.Len(t, contents, 1)

	text := contents[0].(mcp.TextResourceContents).Text
	// Should be valid JSON (null or array of templates).
	assert.True(t, text == "null" || text[0] == '[', "expected null or JSON array, got: %s", text)
}
