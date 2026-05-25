package access

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	accesspkg "github.com/immmmmmmu/plsnt/internal/access"
	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/config"
	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

// NewCmd creates the "access" command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access",
		Short: "Import from Access database files (.mdb/.accdb)",
		Long: `Work with Microsoft Access database files using mdbtools.

Requires mdbtools to be installed: sudo apt install mdbtools`,
	}
	cmd.AddCommand(newTablesCmd())
	cmd.AddCommand(newExportCmd())
	cmd.AddCommand(newImportCmd())
	return cmd
}

// plsnt access tables <file.mdb>
func newTablesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tables <file>",
		Short: "List tables in an Access database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := args[0]
			if err := validateAccessFile(dbPath); err != nil {
				return err
			}

			reader := accesspkg.NewAccessReader()
			ctx := context.Background()

			if err := reader.CheckMDBTools(ctx); err != nil {
				return err
			}

			tables, err := reader.ListTables(ctx, dbPath)
			if err != nil {
				return err
			}

			if len(tables) == 0 {
				fmt.Fprintln(os.Stderr, "No tables found")
				return nil
			}

			// Output as JSON array for agent consumption
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(tables)
		},
	}
	return cmd
}

// plsnt access export <file.mdb> <table-name>
func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <file> <table-name>",
		Short: "Export a table as CSV to stdout",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := args[0]
			tableName := args[1]

			if err := validateAccessFile(dbPath); err != nil {
				return err
			}

			reader := accesspkg.NewAccessReader()
			ctx := context.Background()

			if err := reader.CheckMDBTools(ctx); err != nil {
				return err
			}

			csvBytes, err := reader.ExportTable(ctx, dbPath, tableName)
			if err != nil {
				return err
			}

			_, err = os.Stdout.Write(csvBytes)
			return err
		},
	}
	return cmd
}

// plsnt access import <file.mdb> <table-name> --site-id <id> --mapping <yaml> [--keys ClassA]
func newImportCmd() *cobra.Command {
	var siteID string
	var mappingPath string
	var keysFlag string

	cmd := &cobra.Command{
		Use:   "import <file> <table-name>",
		Short: "Export table from Access and import to Pleasanter",
		Long: `Export a table from an Access database as CSV, apply column mapping,
and import the records into Pleasanter.

When --keys is provided, the bulkupsert API is used for update-or-insert behavior.
Without --keys, records are created one by one.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := args[0]
			tableName := args[1]

			if err := validateAccessFile(dbPath); err != nil {
				return err
			}

			sid, err := strconv.ParseInt(siteID, 10, 64)
			if err != nil || sid <= 0 {
				return errs.New(errs.CodeInvalidInput,
					fmt.Sprintf("invalid site ID: %s", siteID)).
					WithSuggestion("Site ID must be a positive integer")
			}

			// Load mapping
			mapping, err := pleasanter.LoadMapping(mappingPath)
			if err != nil {
				return err
			}

			// Export table from Access
			reader := accesspkg.NewAccessReader()
			ctx := context.Background()

			if err := reader.CheckMDBTools(ctx); err != nil {
				return err
			}

			csvBytes, err := reader.ExportTable(ctx, dbPath, tableName)
			if err != nil {
				return err
			}

			// Parse CSV through existing import pipeline
			records, err := pleasanter.ParseCSV(bytes.NewReader(csvBytes), mapping)
			if err != nil {
				return err
			}

			if len(records) == 0 {
				return errs.New(errs.CodeInvalidInput,
					fmt.Sprintf("table %q contains no data rows", tableName)).
					WithSuggestion("Check the table contents with 'plsnt access export'")
			}

			// Parse keys
			var keys []string
			if keysFlag != "" {
				for _, k := range strings.Split(keysFlag, ",") {
					if trimmed := strings.TrimSpace(k); trimmed != "" {
						keys = append(keys, trimmed)
					}
				}
			}

			// Create client and import
			client, err := newClient(cmd, keys)
			if err != nil {
				return err
			}

			svc := pleasanter.NewRecordService(client)
			results, err := svc.Import(ctx, sid, records, keys)
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Imported %d record(s) from table %q\n", len(results), tableName)

			// Output results
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if len(results) == 1 {
				return enc.Encode(results[0])
			}
			combined := map[string]any{
				"Results": results,
				"Count":   len(results),
			}
			return enc.Encode(combined)
		},
	}

	cmd.Flags().StringVar(&siteID, "site-id", "", "site ID (required)")
	_ = cmd.MarkFlagRequired("site-id")
	cmd.Flags().StringVar(&mappingPath, "mapping", "", "mapping YAML file path (required)")
	_ = cmd.MarkFlagRequired("mapping")
	cmd.Flags().StringVar(&keysFlag, "keys", "", "comma-separated upsert key columns (optional)")

	return cmd
}

func validateAccessFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("file not found: %s", path)).
			WithSuggestion("Check the file path")
	}
	lower := strings.ToLower(path)
	if !strings.HasSuffix(lower, ".mdb") && !strings.HasSuffix(lower, ".accdb") {
		return errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("unsupported file type: %s", path)).
			WithSuggestion("Provide a .mdb or .accdb file")
	}
	return nil
}

func newClient(cmd *cobra.Command, keys []string) (api.Client, error) {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, err
	}

	profileFlag, _ := cmd.Flags().GetString("profile")
	profile, _, err := cfg.ActiveProfileWithOverride(profileFlag)
	if err != nil {
		return nil, errs.New(errs.CodeValidationError, err.Error()).
			WithSuggestion("Run 'plsnt config set' to configure a profile")
	}

	url, apiKey, apiVersion := profile.Resolve()
	if url == "" || apiKey == "" {
		return nil, errs.New(errs.CodeValidationError, "URL and API key are required").
			WithSuggestion("Run 'plsnt config set --url <url> --api-key <key>'")
	}

	var opts []api.Option
	if insecure, _ := cmd.Flags().GetBool("insecure"); insecure {
		opts = append(opts, api.WithInsecure())
	}
	if len(keys) > 0 {
		return api.New(url, apiKey, apiVersion, opts...), nil
	}
	opts = append(opts, api.WithRetryDisabled())
	return api.New(url, apiKey, apiVersion, opts...), nil
}
