package mcp

import (
	"github.com/spf13/cobra"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/config"
	"github.com/immmmmmmu/plsnt/internal/errs"
	mcpserver "github.com/immmmmmmu/plsnt/internal/mcp"
)

var version = "dev"

// SetVersion sets the version string used in MCP server info.
func SetVersion(v string) {
	version = v
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP (Model Context Protocol) server",
	}

	cmd.AddCommand(newServeCmd())
	return cmd
}

func newServeCmd() *cobra.Command {
	var logFile string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start MCP server on stdio",
		Long:  "Start a Model Context Protocol server that exposes plsnt CLI tools via stdio JSON-RPC. Compatible with Claude Code and Claude Desktop.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.DefaultPath())
			if err != nil {
				return errs.Wrap(err, errs.CodeInternalError)
			}

			profileFlag, _ := cmd.Flags().GetString("profile")
			profile, profileName, err := cfg.ActiveProfileWithOverride(profileFlag)
			if err != nil {
				return errs.New(errs.CodeValidationError, err.Error()).
					WithSuggestion("Run 'plsnt config set' to configure a profile")
			}

			url, apiKey, apiVersion := profile.Resolve()
			if url == "" || apiKey == "" {
				return errs.New(errs.CodeValidationError, "URL and API key are required").
					WithSuggestion("Run 'plsnt config set --url <url> --api-key <key>'")
			}

			var opts []api.Option
			if insecure, _ := cmd.Flags().GetBool("insecure"); insecure {
				opts = append(opts, api.WithInsecure())
			}
			client := api.New(url, apiKey, apiVersion, opts...)

			var serverOpts []mcpserver.ServerOption
			if logFile != "" {
				serverOpts = append(serverOpts, mcpserver.WithLogFile(logFile))
			}

			srv := mcpserver.NewServer(version, client, cfg, profileName, serverOpts...)
			return srv.Serve()
		},
	}

	cmd.Flags().StringVar(&logFile, "log-file", "", "log file path for verbose output")

	return cmd
}
