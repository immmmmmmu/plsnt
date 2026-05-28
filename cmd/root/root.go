package root

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/immmmmmmu/plsnt/cmd/access"
	"github.com/immmmmmmu/plsnt/cmd/batch"
	cmdconfig "github.com/immmmmmmu/plsnt/cmd/config"
	"github.com/immmmmmmu/plsnt/cmd/dept"
	"github.com/immmmmmmu/plsnt/cmd/group"
	initcmd "github.com/immmmmmmu/plsnt/cmd/init"
	cmdmcp "github.com/immmmmmmu/plsnt/cmd/mcp"
	"github.com/immmmmmmu/plsnt/cmd/migrate"
	"github.com/immmmmmmu/plsnt/cmd/record"
	"github.com/immmmmmmu/plsnt/cmd/schema"
	"github.com/immmmmmmu/plsnt/cmd/site"
	"github.com/immmmmmmu/plsnt/cmd/user"
	"github.com/immmmmmmu/plsnt/cmd/version"
	"github.com/immmmmmmu/plsnt/cmd/workflow"
)

var (
	profile    string
	output     string
	fields     string
	verbose    bool
	silent     bool
	dryRun     bool
	jsonInput  string
	logFile    string
	insecure   bool
	isTTY      bool
)

var rootCmd = &cobra.Command{
	Use:   "plsnt",
	Short: "Pleasanter CLI tool - agent-first design",
	Long:  "plsnt is a CLI tool for operating Pleasanter via REST API. Designed for AI agents first, humans second.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		isTTY = term.IsTerminal(int(os.Stdout.Fd()))
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "profile name")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output format (json/table/csv/ndjson/count/ids)")
	rootCmd.PersistentFlags().StringVar(&fields, "fields", "", "comma-separated field names to include")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "suppress non-error output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show request without executing")
	rootCmd.PersistentFlags().StringVar(&jsonInput, "json", "", "raw JSON payload")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "log file path")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "skip TLS certificate verification")

	rootCmd.AddCommand(access.NewCmd())
	rootCmd.AddCommand(batch.NewCmd())
	rootCmd.AddCommand(initcmd.NewCmd())
	rootCmd.AddCommand(version.NewCmd())
	rootCmd.AddCommand(cmdconfig.NewCmd())
	rootCmd.AddCommand(schema.NewCmd())
	rootCmd.AddCommand(record.NewCmd())
	rootCmd.AddCommand(site.NewCmd())
	rootCmd.AddCommand(user.NewCmd())
	rootCmd.AddCommand(group.NewCmd())
	rootCmd.AddCommand(dept.NewCmd())
	rootCmd.AddCommand(cmdmcp.NewCmd())
	rootCmd.AddCommand(migrate.NewCmd())
	rootCmd.AddCommand(workflow.NewCmd())
}

func Execute() error {
	return rootCmd.Execute()
}

func IsTTY() bool {
	return isTTY
}
