package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

func (s *Server) registerProfilesResource() {
	resource := mcp.NewResource(
		"plsnt://profiles",
		"Available Profiles",
		mcp.WithResourceDescription("List of configured Pleasanter connection profiles"),
		mcp.WithMIMEType("application/json"),
	)

	s.mcpServer.AddResource(resource, s.handleProfilesResource)
}

func (s *Server) handleProfilesResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	type profileInfo struct {
		Name    string `json:"name"`
		URL     string `json:"url"`
		Active  bool   `json:"active"`
		HasKey  bool   `json:"has_api_key"`
	}

	var profiles []profileInfo
	for name, p := range s.cfg.Profiles {
		url, apiKey, _ := p.Resolve()
		profiles = append(profiles, profileInfo{
			Name:   name,
			URL:    url,
			Active: name == s.profile,
			HasKey: apiKey != "",
		})
	}

	data, _ := json.MarshalIndent(profiles, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

func (s *Server) registerTemplatesResource() {
	resource := mcp.NewResource(
		"plsnt://templates",
		"Available Templates",
		mcp.WithResourceDescription("List of batch and workflow YAML templates"),
		mcp.WithMIMEType("application/json"),
	)

	s.mcpServer.AddResource(resource, s.handleTemplatesResource)
}

func (s *Server) handleTemplatesResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	type templateInfo struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}

	var templates []templateInfo

	// Scan templates/ directory relative to CWD.
	entries, err := os.ReadDir("templates")
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && (filepath.Ext(e.Name()) == ".yaml" || filepath.Ext(e.Name()) == ".yml") {
				templates = append(templates, templateInfo{
					Name: e.Name(),
					Path: filepath.Join("templates", e.Name()),
				})
			}
		}
	}

	data, _ := json.MarshalIndent(templates, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

func (s *Server) registerSchemaResourceTemplate() {
	tmpl := mcp.NewResourceTemplate(
		"plsnt://schema/{siteId}",
		"Table Schema",
		mcp.WithTemplateDescription("Column definitions for a Pleasanter table (dynamic resource)"),
		mcp.WithTemplateMIMEType("application/json"),
	)

	s.mcpServer.AddResourceTemplate(tmpl, s.handleSchemaResourceTemplate)
}

func (s *Server) handleSchemaResourceTemplate(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract siteId from URI: plsnt://schema/12345
	uri := request.Params.URI
	var siteIDStr string
	if _, err := fmt.Sscanf(uri, "plsnt://schema/%s", &siteIDStr); err != nil {
		return nil, fmt.Errorf("invalid schema URI: %s", uri)
	}

	var siteID int64
	if _, err := fmt.Sscanf(siteIDStr, "%d", &siteID); err != nil {
		return nil, fmt.Errorf("invalid site ID in URI: %s", siteIDStr)
	}

	svc := pleasanter.NewSchemaService(s.client)
	schema, err := svc.GetSchema(ctx, siteID)
	if err != nil {
		return nil, err
	}

	data, _ := json.MarshalIndent(schema, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

// ensure interface compliance.
var _ server.ResourceHandlerFunc = (*Server)(nil).handleProfilesResource
var _ server.ResourceHandlerFunc = (*Server)(nil).handleTemplatesResource
var _ server.ResourceTemplateHandlerFunc = (*Server)(nil).handleSchemaResourceTemplate
