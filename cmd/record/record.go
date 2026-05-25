package record

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/config"
	encdetect "github.com/immmmmmmu/plsnt/internal/encoding"
	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/format"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
	"github.com/immmmmmmu/plsnt/internal/validate"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Manage records",
	}

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newUpsertCmd())
	cmd.AddCommand(newImportCmd())
	cmd.AddCommand(newBulkDeleteCmd())
	return cmd
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <record-id>",
		Short: "Get a single record by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.RecordID(args[0]); err != nil {
				return err
			}
			recordID, _ := strconv.ParseInt(args[0], 10, 64)

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			jsonPayload, _ := cmd.Flags().GetString("json")
			if jsonPayload != "" {
				if err := validate.JSONSyntax(jsonPayload); err != nil {
					return err
				}
				var payload map[string]any
				if err := json.Unmarshal([]byte(jsonPayload), &payload); err != nil {
					return errs.New(errs.CodeInvalidInput, "invalid JSON payload").
						WithSuggestion("Provide valid JSON, e.g. --json '{}'")
				}
				svc := pleasanter.NewRecordService(client)
				result, err := svc.GetRaw(context.Background(), recordID, payload)
				if err != nil {
					return err
				}
				return outputResult(cmd, result)
			}

			svc := pleasanter.NewRecordService(client)
			resp, err := svc.Get(context.Background(), recordID)
			if err != nil {
				return err
			}

			return outputResponse(cmd, resp)
		},
	}
	return cmd
}

func newListCmd() *cobra.Command {
	var siteID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List records from a site",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.SiteID(siteID); err != nil {
				return err
			}
			sid, _ := strconv.ParseInt(siteID, 10, 64)

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			jsonPayload, _ := cmd.Flags().GetString("json")
			if jsonPayload != "" {
				if err := validate.JSONSyntax(jsonPayload); err != nil {
					return err
				}
				var payload map[string]any
				if err := json.Unmarshal([]byte(jsonPayload), &payload); err != nil {
					return errs.New(errs.CodeInvalidInput, "invalid JSON payload")
				}
				svc := pleasanter.NewRecordService(client)
				result, err := svc.ListRaw(context.Background(), sid, payload)
				if err != nil {
					return err
				}
				return outputResult(cmd, result)
			}

			svc := pleasanter.NewRecordService(client)

			viewJSON, _ := cmd.Flags().GetString("view")
			var view *pleasanter.View
			if viewJSON != "" {
				view = &pleasanter.View{}
				if err := json.Unmarshal([]byte(viewJSON), view); err != nil {
					return errs.New(errs.CodeInvalidInput, "invalid --view JSON").
						WithSuggestion(`Example: --view '{"ColumnFilterHash":{"ClassA":"Red"}}'`)
				}
			}

			allPages, _ := cmd.Flags().GetBool("all-pages")

			var resp *pleasanter.APIResponse
			if allPages {
				resp, err = svc.ListAll(context.Background(), pleasanter.ListOptions{
					SiteID: sid,
					View:   view,
				})
			} else {
				resp, err = svc.List(context.Background(), pleasanter.ListOptions{
					SiteID: sid,
					View:   view,
				})
			}
			if err != nil {
				return err
			}

			return outputResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&siteID, "site-id", "", "site ID (required)")
	_ = cmd.MarkFlagRequired("site-id")
	cmd.Flags().String("view", "", "View filter JSON")
	cmd.Flags().Bool("all-pages", false, "fetch all pages automatically")

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

func outputResponse(cmd *cobra.Command, resp *pleasanter.APIResponse) error {
	output, _ := cmd.Flags().GetString("output")
	fieldsStr, _ := cmd.Flags().GetString("fields")

	var fields []string
	if fieldsStr != "" {
		for _, f := range strings.Split(fieldsStr, ",") {
			if trimmed := strings.TrimSpace(f); trimmed != "" {
				fields = append(fields, trimmed)
			}
		}
	}

	data := resp.Response.Data
	if len(fields) > 0 {
		data = filterFields(data, fields)
	}

	// Show metadata on stderr for human-friendly formats only
	outputLower := strings.ToLower(output)
	if outputLower != "count" && outputLower != "ids" {
		fmt.Fprintf(os.Stderr, "TotalCount: %d, Offset: %d, PageSize: %d\n",
			resp.Response.TotalCount, resp.Response.Offset, resp.Response.PageSize)
	}

	formatter, err := format.New(output, fields)
	if err != nil {
		return err
	}

	return formatter.Format(os.Stdout, data)
}

func outputResult(cmd *cobra.Command, result map[string]any) error {
	output, _ := cmd.Flags().GetString("output")
	fieldsStr, _ := cmd.Flags().GetString("fields")

	var fields []string
	if fieldsStr != "" {
		for _, f := range strings.Split(fieldsStr, ",") {
			if trimmed := strings.TrimSpace(f); trimmed != "" {
				fields = append(fields, trimmed)
			}
		}
	}

	// Unwrap Pleasanter API response structure if present
	var dataToFormat any = result
	if data, meta := format.UnwrapAPIResponse(result); meta != nil {
		dataToFormat = data
		outputLower := strings.ToLower(output)
		if outputLower != "count" && outputLower != "ids" {
			fmt.Fprintf(os.Stderr, "TotalCount: %v, Offset: %v, PageSize: %v\n",
				meta["TotalCount"], meta["Offset"], meta["PageSize"])
		}
	}

	formatter, err := format.New(output, fields)
	if err != nil {
		return err
	}

	return formatter.Format(os.Stdout, dataToFormat)
}

func filterFields(records []pleasanter.Record, fields []string) []pleasanter.Record {
	// For structured records, field filtering is handled by the formatter.
	// This is a pass-through; the formatter's resolveHeaders does the actual filtering.
	return records
}

func newCreateCmd() *cobra.Command {
	var siteID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new record",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.SiteID(siteID); err != nil {
				return err
			}
			sid, _ := strconv.ParseInt(siteID, 10, 64)

			payload, err := readPayload(cmd)
			if err != nil {
				return err
			}

			client, err := newClientNoRetry(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewRecordService(client)
			result, err := svc.Create(context.Background(), sid, payload)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}

	cmd.Flags().StringVar(&siteID, "site-id", "", "site ID (required)")
	_ = cmd.MarkFlagRequired("site-id")
	return cmd
}

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <record-id>",
		Short: "Update an existing record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.RecordID(args[0]); err != nil {
				return err
			}
			recordID, _ := strconv.ParseInt(args[0], 10, 64)

			payload, err := readPayload(cmd)
			if err != nil {
				return err
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewRecordService(client)
			result, err := svc.Update(context.Background(), recordID, payload)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}
	return cmd
}

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <record-id>",
		Short: "Delete a record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.RecordID(args[0]); err != nil {
				return err
			}
			recordID, _ := strconv.ParseInt(args[0], 10, 64)

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewRecordService(client)
			result, err := svc.Delete(context.Background(), recordID)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}
	return cmd
}

func newUpsertCmd() *cobra.Command {
	var siteID string
	var keysFlag string

	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "Bulk upsert records (update-or-insert)",
		Long: `Bulk upsert records using the Pleasanter bulkupsert API.

When --keys is provided, --json (or stdin) should be a JSON array of record objects.
When --keys is not provided, --json (or stdin) should be a JSON object with "Keys" and "Data" fields.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.SiteID(siteID); err != nil {
				return err
			}
			sid, _ := strconv.ParseInt(siteID, 10, 64)

			jsonPayload, _ := cmd.Flags().GetString("json")

			// Read from stdin if no --json flag and stdin has data
			if jsonPayload == "" {
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					data, err := io.ReadAll(io.LimitReader(os.Stdin, 10*1024*1024))
					if err != nil {
						return errs.New(errs.CodeInternalError, "failed to read stdin")
					}
					jsonPayload = string(data)
				}
			}

			if jsonPayload == "" {
				return errs.New(errs.CodeInvalidInput, "JSON payload is required").
					WithSuggestion("Use --json '[{...}]' with --keys, or --json '{\"Keys\":[...],\"Data\":[...]}', or pipe JSON via stdin")
			}

			if err := validate.JSONSyntax(jsonPayload); err != nil {
				return err
			}

			var keys []string
			var records []map[string]any

			if keysFlag != "" {
				// Option B: --keys provided, JSON is an array of record objects
				for _, k := range strings.Split(keysFlag, ",") {
					if trimmed := strings.TrimSpace(k); trimmed != "" {
						keys = append(keys, trimmed)
					}
				}
				if len(keys) == 0 {
					return errs.New(errs.CodeInvalidInput, "--keys must contain at least one column name").
						WithSuggestion("Example: --keys ClassA or --keys ClassA,ClassB")
				}

				if err := json.Unmarshal([]byte(jsonPayload), &records); err != nil {
					return errs.New(errs.CodeInvalidInput, "when --keys is provided, JSON must be an array of record objects").
						WithSuggestion(`Example: --json '[{"Title":"Item 1","ClassHash":{"ClassA":"key1"}}]'`)
				}
			} else {
				// Option A: full payload with Keys and Data
				var fullPayload map[string]any
				if err := json.Unmarshal([]byte(jsonPayload), &fullPayload); err != nil {
					return errs.New(errs.CodeInvalidInput, "without --keys, JSON must be an object with Keys and Data fields").
						WithSuggestion(`Example: --json '{"Keys":["ClassA"],"Data":[...]}'`)
				}

				rawKeys, ok := fullPayload["Keys"]
				if !ok {
					return errs.New(errs.CodeInvalidInput, "JSON payload missing required 'Keys' field").
						WithSuggestion(`Include "Keys": ["ClassA"] in the JSON payload, or use --keys flag`)
				}
				keysSlice, ok := rawKeys.([]any)
				if !ok {
					return errs.New(errs.CodeInvalidInput, "'Keys' must be a JSON array of strings").
						WithSuggestion(`Example: "Keys": ["ClassA", "ClassB"]`)
				}
				for _, k := range keysSlice {
					s, ok := k.(string)
					if !ok {
						return errs.New(errs.CodeInvalidInput, "'Keys' array must contain only strings")
					}
					keys = append(keys, s)
				}

				rawData, ok := fullPayload["Data"]
				if !ok {
					return errs.New(errs.CodeInvalidInput, "JSON payload missing required 'Data' field").
						WithSuggestion(`Include "Data": [{...}] in the JSON payload`)
				}
				dataSlice, ok := rawData.([]any)
				if !ok {
					return errs.New(errs.CodeInvalidInput, "'Data' must be a JSON array of record objects")
				}
				for _, item := range dataSlice {
					rec, ok := item.(map[string]any)
					if !ok {
						return errs.New(errs.CodeInvalidInput, "each element in 'Data' must be a JSON object")
					}
					records = append(records, rec)
				}
			}

			if len(records) == 0 {
				return errs.New(errs.CodeInvalidInput, "no records provided for upsert").
					WithSuggestion("Provide at least one record in the data array")
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewRecordService(client)
			result, err := svc.BulkUpsert(context.Background(), sid, keys, records)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}

	cmd.Flags().StringVar(&siteID, "site-id", "", "site ID (required)")
	_ = cmd.MarkFlagRequired("site-id")
	cmd.Flags().StringVar(&keysFlag, "keys", "", "comma-separated column names for upsert matching (e.g. ClassA,ClassB)")
	return cmd
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

func newImportCmd() *cobra.Command {
	var siteID string
	var filePath string
	var mappingPath string
	var keysFlag string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import records from a CSV file with column mapping",
		Long: `Import records from a CSV file using a YAML mapping file to map CSV columns
to Pleasanter fields.

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

			utf8Reader, detectedEnc, err := encdetect.NewReader(f)
			if err != nil {
				return errs.New(errs.CodeInternalError,
					fmt.Sprintf("failed to detect encoding: %v", err)).
					WithSuggestion("Check that the CSV file is not corrupted")
			}
			if detectedEnc != "UTF-8" {
				fmt.Fprintf(os.Stderr, "Detected encoding: %s (converted to UTF-8)\n", detectedEnc)
			}

			records, err := pleasanter.ParseCSV(utf8Reader, mapping)
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

			fmt.Fprintf(os.Stderr, "Imported %d record(s)\n", len(results))

			if len(results) == 1 {
				return outputResult(cmd, results[0])
			}

			// For multiple results, output as JSON array
			combined := map[string]any{
				"Results": results,
				"Count":   len(results),
			}
			return outputResult(cmd, combined)
		},
	}

	cmd.Flags().StringVar(&siteID, "site-id", "", "site ID (required)")
	_ = cmd.MarkFlagRequired("site-id")
	cmd.Flags().StringVar(&filePath, "file", "", "CSV file path (required)")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringVar(&mappingPath, "mapping", "", "mapping YAML file path (required)")
	_ = cmd.MarkFlagRequired("mapping")
	cmd.Flags().StringVar(&keysFlag, "keys", "", "comma-separated upsert key columns (optional)")

	return cmd
}

const maxBulkDeleteIDs = 1000
const bulkDeleteConfirmThreshold = 100

func newBulkDeleteCmd() *cobra.Command {
	var siteID string
	var idsFlag string
	var viewFlag string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "bulk-delete",
		Short: "Delete multiple records at once",
		Long: `Delete multiple records by IDs or by View filter.

When using --ids, specify comma-separated record IDs (max 1000).
When using --view, specify a View filter JSON. This requires --confirm
because it could delete many records.

Examples:
  plsnt record bulk-delete --site-id 100 --ids 101,102,103
  plsnt record bulk-delete --site-id 100 --view '{"ColumnFilterHash":{"ClassA":"old"}}' --confirm`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.SiteID(siteID); err != nil {
				return err
			}
			sid, _ := strconv.ParseInt(siteID, 10, 64)

			hasIDs := idsFlag != ""
			hasView := viewFlag != ""

			if hasIDs && hasView {
				return errs.New(errs.CodeInvalidInput, "--ids and --view are mutually exclusive").
					WithSuggestion("Use either --ids or --view, not both")
			}
			if !hasIDs && !hasView {
				return errs.New(errs.CodeInvalidInput, "either --ids or --view is required").
					WithSuggestion("Use --ids 101,102 or --view '{\"ColumnFilterHash\":{...}}'")
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}
			svc := pleasanter.NewRecordService(client)

			if hasIDs {
				parts := strings.Split(idsFlag, ",")
				var recordIDs []int64
				for _, p := range parts {
					trimmed := strings.TrimSpace(p)
					if trimmed == "" {
						continue
					}
					id, err := strconv.ParseInt(trimmed, 10, 64)
					if err != nil || id <= 0 {
						return errs.New(errs.CodeInvalidInput,
							fmt.Sprintf("invalid record ID: %q", trimmed)).
							WithSuggestion("Record IDs must be positive integers")
					}
					recordIDs = append(recordIDs, id)
				}
				if len(recordIDs) == 0 {
					return errs.New(errs.CodeInvalidInput, "no valid record IDs provided").
						WithSuggestion("Example: --ids 101,102,103")
				}
				if len(recordIDs) > maxBulkDeleteIDs {
					return errs.New(errs.CodeInvalidInput,
						fmt.Sprintf("too many IDs: %d (maximum %d per request)", len(recordIDs), maxBulkDeleteIDs)).
						WithSuggestion(fmt.Sprintf("Split into batches of %d or fewer", maxBulkDeleteIDs))
				}
				if len(recordIDs) > bulkDeleteConfirmThreshold && !confirm {
					return errs.New(errs.CodeInvalidInput,
						fmt.Sprintf("deleting %d records requires --confirm flag", len(recordIDs))).
						WithSuggestion("Add --confirm to proceed with bulk deletion")
				}

				result, err := svc.BulkDelete(context.Background(), sid, recordIDs)
				if err != nil {
					return err
				}
				return outputResult(cmd, result)
			}

			// --view mode
			if !confirm {
				return errs.New(errs.CodeInvalidInput,
					"--view deletion requires --confirm flag (could delete many records)").
					WithSuggestion("Add --confirm to proceed with view-based bulk deletion")
			}

			if err := validate.JSONSyntax(viewFlag); err != nil {
				return err
			}

			var view map[string]any
			if err := json.Unmarshal([]byte(viewFlag), &view); err != nil {
				return errs.New(errs.CodeInvalidInput, "invalid --view JSON").
					WithSuggestion(`Example: --view '{"ColumnFilterHash":{"ClassA":"old"}}'`)
			}

			result, err := svc.BulkDeleteByView(context.Background(), sid, view)
			if err != nil {
				return err
			}
			return outputResult(cmd, result)
		},
	}

	cmd.Flags().StringVar(&siteID, "site-id", "", "site ID (required)")
	_ = cmd.MarkFlagRequired("site-id")
	cmd.Flags().StringVar(&idsFlag, "ids", "", "comma-separated record IDs to delete (max 1000)")
	cmd.Flags().StringVar(&viewFlag, "view", "", "View filter JSON (deletes all matching records)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm dangerous deletion")

	return cmd
}

func readPayload(cmd *cobra.Command) (map[string]any, error) {
	jsonPayload, _ := cmd.Flags().GetString("json")

	// Read from stdin if no --json flag and stdin has data
	if jsonPayload == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(io.LimitReader(os.Stdin, 10*1024*1024)) // 10MB limit
			if err != nil {
				return nil, errs.New(errs.CodeInternalError, "failed to read stdin")
			}
			jsonPayload = string(data)
		}
	}

	if jsonPayload == "" {
		return nil, errs.New(errs.CodeInvalidInput, "JSON payload is required").
			WithSuggestion("Use --json '{...}' or pipe JSON via stdin")
	}

	if err := validate.JSONSyntax(jsonPayload); err != nil {
		return nil, err
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(jsonPayload), &payload); err != nil {
		return nil, errs.New(errs.CodeInvalidInput, "invalid JSON payload").
			WithSuggestion("Provide valid JSON object, e.g. --json '{\"Title\": \"test\"}'")
	}

	return payload, nil
}
