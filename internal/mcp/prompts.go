package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) registerPrompts() {
	s.mcpServer.AddPrompt(mcp.NewPrompt("create-app",
		mcp.WithPromptDescription("Interactive guide for creating a Pleasanter application (table design → creation → data import)"),
		mcp.WithArgument("app_name",
			mcp.ArgumentDescription("Name of the application to create"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("folder_id",
			mcp.ArgumentDescription("Parent folder site ID"),
			mcp.RequiredArgument(),
		),
	), s.handleCreateAppPrompt)

	s.mcpServer.AddPrompt(mcp.NewPrompt("migrate-csv",
		mcp.WithPromptDescription("Step-by-step guide for migrating CSV data into Pleasanter (mapping generation → validation → execution)"),
		mcp.WithArgument("site_id",
			mcp.ArgumentDescription("Target site ID for migration"),
			mcp.RequiredArgument(),
		),
	), s.handleMigrateCSVPrompt)
}

func (s *Server) handleCreateAppPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	appName := request.Params.Arguments["app_name"]
	folderID := request.Params.Arguments["folder_id"]

	return &mcp.GetPromptResult{
		Description: "Interactive guide for creating a Pleasanter application",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: "I want to create a new Pleasanter application called \"" + appName + "\" under folder " + folderID + ". " +
						"Please guide me through the following steps:\n\n" +
						"1. **Table Design**: Help me define the table structure (columns, types, choices)\n" +
						"2. **Create Site**: Use the `site_create` tool to create the table\n" +
						"3. **Configure Columns**: Use `site_update` to set up column definitions (labels, choices, required fields)\n" +
						"4. **Schema Check**: Use `schema_get` to verify the configuration\n" +
						"5. **Sample Data**: Optionally import sample data using `record_create` or `record_import`\n\n" +
						"Available tools: site_create, site_update, schema_get, record_create, record_import, record_upsert\n\n" +
						"Start by asking me what kind of data this application will manage.",
				},
			},
		},
	}, nil
}

func (s *Server) handleMigrateCSVPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	siteID := request.Params.Arguments["site_id"]

	return &mcp.GetPromptResult{
		Description: "Step-by-step CSV migration guide",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: "I want to migrate CSV data into Pleasanter site " + siteID + ". " +
						"Please guide me through the following steps:\n\n" +
						"1. **Schema Check**: Use `schema_get` to inspect the target table's column definitions\n" +
						"2. **Mapping Definition**: Help me create a YAML mapping between CSV columns and Pleasanter fields\n" +
						"3. **Validation**: Review a sample of the CSV data to check for issues\n" +
						"4. **Dry Run**: If using workflow_master, run with dry_run=true first\n" +
						"5. **Execute**: Use `migrate_execute` or `record_import` to perform the migration\n" +
						"6. **Verify**: Use `record_list` to confirm imported data\n\n" +
						"Available tools: schema_get, migrate_execute, record_import, record_list, workflow_master\n\n" +
						"Start by showing me the target table's schema.",
				},
			},
		},
	}, nil
}

// ensure interface compliance.
var _ server.PromptHandlerFunc = (*Server)(nil).handleCreateAppPrompt
var _ server.PromptHandlerFunc = (*Server)(nil).handleMigrateCSVPrompt
