package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/config"
	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
	"github.com/immmmmmmu/plsnt/internal/workflow/master"
)

const maxMasterFileSize = 10 * 1024 * 1024 // 10MB

func newMasterCmd() *cobra.Command {
	var (
		siteID int64
		file   string
		key    string
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "master",
		Short: "Import master data from CSV",
		Long:  "Import master data from a CSV file into a Pleasanter table using key-based upsert (update existing, create new)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if siteID <= 0 {
				return errs.New(errs.CodeValidationError,
					fmt.Sprintf("site-id must be a positive integer, got: %d", siteID)).
					WithSuggestion("Specify a valid site ID, e.g. --site-id 32200")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open CSV file
			f, err := os.Open(file)
			if err != nil {
				return errs.New(errs.CodeInvalidInput,
					fmt.Sprintf("failed to open file: %v", err)).
					WithSuggestion("Check that the file path is correct and readable")
			}
			defer f.Close()

			// Limit file size to 10MB
			limitedReader := io.LimitReader(f, maxMasterFileSize)

			// Create API client (no retry for create/update)
			client, err := newMasterClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewRecordService(client)

			// Create importer
			importer := master.NewImporter(svc, master.Options{
				SiteID: siteID,
				Key:    key,
				DryRun: dryRun,
			})

			// Execute import
			result, err := importer.ImportCSV(context.Background(), limitedReader)
			if err != nil {
				return errs.Wrap(err, errs.CodeInternalError)
			}

			// Output JSON result
			writer := cmd.OutOrStdout()
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				return errs.New(errs.CodeInternalError, fmt.Sprintf("failed to marshal result: %v", err))
			}
			fmt.Fprintln(writer, string(jsonBytes))

			return nil
		},
	}

	cmd.Flags().Int64Var(&siteID, "site-id", 0, "target table site ID (required)")
	cmd.Flags().StringVarP(&file, "file", "f", "", "CSV file path (required)")
	cmd.Flags().StringVarP(&key, "key", "k", "ClassA", "upsert key column name")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show import plan without executing")
	_ = cmd.MarkFlagRequired("site-id")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

// newMasterClient creates an API client with retry disabled (for create/update operations).
func newMasterClient(cmd *cobra.Command) (api.Client, error) {
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
