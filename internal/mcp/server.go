package mcp

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/config"
)

// Server wraps the mcp-go MCPServer and holds plsnt dependencies.
type Server struct {
	mcpServer *server.MCPServer
	client    api.Client
	cfg       *config.Config
	profile   string
	logFile   string
}

// ServerOption configures a Server.
type ServerOption func(*Server)

// WithLogFile sets a log file path for verbose output.
func WithLogFile(path string) ServerOption {
	return func(s *Server) {
		s.logFile = path
	}
}

// NewServer creates a new plsnt MCP server.
func NewServer(version string, client api.Client, cfg *config.Config, profileName string, opts ...ServerOption) *Server {
	s := &Server{
		client:  client,
		cfg:     cfg,
		profile: profileName,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.mcpServer = server.NewMCPServer(
		"plsnt",
		version,
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithInstructions("plsnt MCP Server — Pleasanter CLI tools exposed via MCP. "+
			"Use these tools for CRUD operations, batch execution, workflow management, and data migration on Pleasanter."),
	)

	s.registerTools()
	s.registerResources()
	s.registerPrompts()

	return s
}

// Serve starts the MCP server on stdio.
func (s *Server) Serve() error {
	var stdioOpts []server.StdioOption

	if s.logFile != "" {
		f, err := os.OpenFile(s.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", s.logFile, err)
		}
		defer f.Close()
		stdioOpts = append(stdioOpts, server.WithErrorLogger(log.New(f, "[plsnt-mcp] ", log.LstdFlags)))
	} else {
		stdioOpts = append(stdioOpts, server.WithErrorLogger(log.New(io.Discard, "", 0)))
	}

	return server.ServeStdio(s.mcpServer, stdioOpts...)
}

// ServeWith starts the MCP server reading from in and writing to out (for testing).
func (s *Server) ServeWith(ctx context.Context, in io.Reader, out io.Writer) error {
	stdio := server.NewStdioServer(s.mcpServer)
	stdio.SetErrorLogger(log.New(io.Discard, "", 0))
	return stdio.Listen(ctx, in, out)
}

// MCPServer returns the underlying mcp-go server for testing.
func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}

// registerTools registers all MCP tools. Called during construction.
func (s *Server) registerTools() {
	// Phase 1: Differentiation tools
	s.registerConfigTestTool()
	s.registerSchemaGetTool()
	s.registerSiteCreateTool()
	s.registerRecordUpsertTool()
	s.registerRecordImportTool()
	s.registerBatchRunTool()

	// Phase 2: CRUD + Site management
	s.registerRecordListTool()
	s.registerRecordGetTool()
	s.registerRecordCreateTool()
	s.registerRecordUpdateTool()
	s.registerRecordDeleteTool()
	s.registerRecordBulkDeleteTool()
	s.registerSiteGetTool()
	s.registerSiteUpdateTool()
	s.registerSiteCopyTool()
	s.registerSiteSearchTool()

	// Phase 3: Workflow + Migration
	s.registerWorkflowDeployTool()
	s.registerWorkflowMasterTool()
	s.registerWorkflowExportTool()
	s.registerMigrateExecuteTool()
}

// registerResources registers all MCP resources. Called during construction.
func (s *Server) registerResources() {
	s.registerProfilesResource()
	s.registerTemplatesResource()
	s.registerSchemaResourceTemplate()
}

// successResult creates a successful CallToolResult with text content.
func successResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(text),
		},
	}
}
