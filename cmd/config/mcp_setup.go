package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/immmmmmmu/plsnt/internal/config"
	"github.com/immmmmmmu/plsnt/internal/errs"
)

func newMCPSetupCmd() *cobra.Command {
	var (
		profileName string
		outputPath  string
		embedKey    bool
		serverType  string
		desktop     bool
	)

	cmd := &cobra.Command{
		Use:   "mcp-setup",
		Short: "Generate .mcp.json for MCP Server configuration",
		Long: `Generate a .mcp.json configuration file for connecting Claude Code or Claude Desktop to MCP servers.

Supported --server modes:
  pleasanter  Pleasanter MCP Server (HTTP, default)
  plsnt       plsnt MCP Server (stdio)
  both        Both servers in one config`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.DefaultPath())
			if err != nil {
				return err
			}

			profile, pName, err := cfg.ActiveProfileWithOverride(profileName)
			if err != nil {
				return errs.New(errs.CodeValidationError, err.Error()).
					WithSuggestion("Run 'plsnt config set' to configure a profile")
			}

			url, apiKey, _ := profile.Resolve()

			// For plsnt-only or desktop, URL is not required.
			if serverType != "plsnt" && !desktop {
				if url == "" {
					return errs.New(errs.CodeValidationError, "profile URL is empty").
						WithSuggestion("Run 'plsnt config set --url <url>'")
				}
				if config.IsHTTP(url) {
					fmt.Fprintln(cmd.ErrOrStderr(),
						"WARNING: Using HTTP (not HTTPS). MCP communication will not be encrypted.")
				}
			}

			var mcpConfig *config.MCPServerConfig

			switch {
			case desktop:
				// Claude Desktop: always stdio
				if pName != "default" && pName != "" {
					mcpConfig = config.GeneratePlsntMCPConfigWithProfile(pName)
				} else {
					mcpConfig = config.GenerateDesktopConfig()
				}
			case serverType == "plsnt":
				if pName != "default" && pName != "" {
					mcpConfig = config.GeneratePlsntMCPConfigWithProfile(pName)
				} else {
					mcpConfig = config.GeneratePlsntMCPConfig()
				}
			case serverType == "both":
				if url == "" {
					return errs.New(errs.CodeValidationError, "profile URL is required for 'both' mode").
						WithSuggestion("Run 'plsnt config set --url <url>'")
				}
				mcpConfig = config.GenerateBothMCPConfig(url)
			default:
				// "pleasanter" (default)
				if embedKey {
					if apiKey == "" {
						return errs.New(errs.CodeValidationError, "API key is not configured").
							WithSuggestion("Run 'plsnt config set --api-key <key>'")
					}
					fmt.Fprintln(cmd.ErrOrStderr(),
						"WARNING: API key is embedded directly. Ensure .mcp.json is in .gitignore.")
					mcpConfig = config.GenerateMCPConfigWithKey(url, apiKey)
				} else {
					mcpConfig = config.GenerateMCPConfig(url)
				}
			}

			jsonBytes, err := mcpConfig.MarshalJSON()
			if err != nil {
				return errs.Wrap(err, errs.CodeInternalError)
			}
			jsonBytes = append(jsonBytes, '\n')

			if outputPath != "" {
				if err := os.WriteFile(outputPath, jsonBytes, 0644); err != nil {
					return errs.Wrap(err, errs.CodeInternalError)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "Written to %s\n", outputPath)
				return nil
			}

			_, err = cmd.OutOrStdout().Write(jsonBytes)
			return err
		},
	}

	cmd.Flags().StringVarP(&profileName, "profile", "p", "", "profile name")
	cmd.Flags().StringVar(&outputPath, "output", "", "output file path")
	cmd.Flags().BoolVar(&embedKey, "embed-key", false, "embed API key directly instead of environment variable reference")
	cmd.Flags().StringVar(&serverType, "server", "pleasanter", "MCP server type: pleasanter, plsnt, or both")
	cmd.Flags().BoolVar(&desktop, "desktop", false, "generate Claude Desktop config (stdio only)")

	return cmd
}
