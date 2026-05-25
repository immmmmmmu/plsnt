package user

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
	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/format"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
	"github.com/immmmmmmu/plsnt/internal/validate"
)

// NewCmd creates the "user" parent command.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newImportCmd())
	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all users",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewUserService(client)
			result, err := svc.List(context.Background())
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}
	return cmd
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <user-id>",
		Short: "Get a single user by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := parseUserID(args[0])
			if err != nil {
				return err
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewUserService(client)
			result, err := svc.Get(context.Background(), userID)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}
	return cmd
}

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := readPayload(cmd)
			if err != nil {
				return err
			}

			client, err := newClientNoRetry(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewUserService(client)
			result, err := svc.Create(context.Background(), payload)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}
	return cmd
}

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <user-id>",
		Short: "Update an existing user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := parseUserID(args[0])
			if err != nil {
				return err
			}

			payload, err := readPayload(cmd)
			if err != nil {
				return err
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewUserService(client)
			result, err := svc.Update(context.Background(), userID, payload)
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
		Use:   "delete <user-id>",
		Short: "Delete a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := parseUserID(args[0])
			if err != nil {
				return err
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewUserService(client)
			result, err := svc.Delete(context.Background(), userID)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}
	return cmd
}

func newImportCmd() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Bulk-create users from a CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return errs.New(errs.CodeInvalidInput, "file path is required").
					WithSuggestion("Use --file users.csv")
			}

			f, err := os.Open(filePath)
			if err != nil {
				return errs.New(errs.CodeInvalidInput,
					fmt.Sprintf("failed to open file: %s", filePath)).
					WithSuggestion("Check the file path and permissions")
			}
			defer f.Close()

			users, err := pleasanter.ParseUserCSV(f)
			if err != nil {
				return err
			}

			if len(users) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No users found in CSV")
				return nil
			}

			client, err := newClientNoRetry(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewUserService(client)
			results, err := svc.BulkCreate(context.Background(), users)

			// Print summary regardless of error
			fmt.Fprintf(cmd.OutOrStdout(), "Created %d of %d users\n", len(results), len(users))

			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to CSV file")

	return cmd
}

func parseUserID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("invalid user ID: %s", s)).
			WithSuggestion("Provide a positive integer user ID")
	}
	return id, nil
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
			WithSuggestion("Provide valid JSON object, e.g. --json '{\"LoginId\": \"user1\"}'")
	}

	return payload, nil
}
