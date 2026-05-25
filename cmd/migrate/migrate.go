package migrate

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/config"
	"github.com/immmmmmmu/plsnt/internal/errs"
	migratelib "github.com/immmmmmmu/plsnt/internal/migrate"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
	"github.com/immmmmmmu/plsnt/internal/validate"
)

// NewCmd creates the migrate command and its subcommands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Data migration tools",
	}
	cmd.AddCommand(newGenerateMappingCmd())
	cmd.AddCommand(newExecuteCmd())
	return cmd
}

func newGenerateMappingCmd() *cobra.Command {
	var filePath string
	var siteID string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "generate-mapping",
		Short: "Generate a mapping YAML from CSV headers and site schema",
		Long: `Reads a CSV file's header row and a site's schema, then generates
a mapping YAML file with best-effort auto-mapping.

Columns are matched by exact column name, exact label text, or
case-insensitive variants. Unmapped columns are written as comments
for manual editing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.SiteID(siteID); err != nil {
				return err
			}
			sid, _ := strconv.ParseInt(siteID, 10, 64)

			headers, err := migratelib.ReadCSVHeaders(filePath)
			if err != nil {
				return err
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			schemaSvc := pleasanter.NewSchemaService(client)
			schema, err := schemaSvc.GetSchema(context.Background(), sid)
			if err != nil {
				return err
			}

			result := migratelib.GenerateMapping(headers, schema)

			var w *os.File
			if outputPath == "" || outputPath == "-" {
				w = os.Stdout
			} else {
				w, err = os.Create(outputPath)
				if err != nil {
					return errs.New(errs.CodeInvalidInput,
						fmt.Sprintf("failed to create output file: %v", err)).
						WithSuggestion("Check that the output path is writable")
				}
				defer w.Close()
			}

			if err := migratelib.WriteMappingYAML(w, result); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Mapping generated: %d mapped, %d unmapped\n",
				len(result.Mapped), len(result.Unmapped))

			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "CSV file path (required)")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringVar(&siteID, "site-id", "", "site ID (required)")
	_ = cmd.MarkFlagRequired("site-id")
	cmd.Flags().StringVar(&outputPath, "output", "", "output YAML path (default: stdout)")

	return cmd
}

func newExecuteCmd() *cobra.Command {
	var filePath string
	var mappingPath string
	var siteID string
	var keysFlag string

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a CSV migration using a mapping file",
		Long: `Applies a mapping file to a CSV and imports records into Pleasanter.

When --keys is provided, the bulkupsert API is used for update-or-insert behavior.
Without --keys, records are created one by one.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.SiteID(siteID); err != nil {
				return err
			}
			sid, _ := strconv.ParseInt(siteID, 10, 64)

			mapping, err := pleasanter.LoadMapping(mappingPath)
			if err != nil {
				return err
			}

			f, err := os.Open(filePath)
			if err != nil {
				return errs.New(errs.CodeInvalidInput,
					fmt.Sprintf("failed to open CSV file: %v", err)).
					WithSuggestion("Check that the file path is correct and readable")
			}
			defer f.Close()

			records, err := pleasanter.ParseCSV(f, mapping)
			if err != nil {
				return err
			}

			if len(records) == 0 {
				return errs.New(errs.CodeInvalidInput, "CSV file contains no data rows").
					WithSuggestion("Ensure the CSV file has at least one data row after the header")
			}

			var keys []string
			if keysFlag != "" {
				for _, k := range strings.Split(keysFlag, ",") {
					if trimmed := strings.TrimSpace(k); trimmed != "" {
						keys = append(keys, trimmed)
					}
				}
			}

			var client api.Client
			if len(keys) > 0 {
				client, err = newClient(cmd)
			} else {
				client, err = newClientNoRetry(cmd)
			}
			if err != nil {
				return err
			}

			svc := pleasanter.NewRecordService(client)
			results, err := svc.Import(context.Background(), sid, records, keys)
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Migration complete: %d record(s) imported\n", len(results))

			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "CSV file path (required)")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringVar(&mappingPath, "mapping", "", "mapping YAML file path (required)")
	_ = cmd.MarkFlagRequired("mapping")
	cmd.Flags().StringVar(&siteID, "site-id", "", "site ID (required)")
	_ = cmd.MarkFlagRequired("site-id")
	cmd.Flags().StringVar(&keysFlag, "keys", "", "comma-separated upsert key columns (optional)")

	return cmd
}

func newClient(cmd *cobra.Command) (api.Client, error) {
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
	return api.New(url, apiKey, apiVersion, opts...), nil
}

func newClientNoRetry(cmd *cobra.Command) (api.Client, error) {
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

	opts := []api.Option{api.WithRetryDisabled()}
	if insecure, _ := cmd.Flags().GetBool("insecure"); insecure {
		opts = append(opts, api.WithInsecure())
	}
	return api.New(url, apiKey, apiVersion, opts...), nil
}
