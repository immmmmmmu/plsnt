package group

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

// NewCmd creates the "group" parent command.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage groups",
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewGroupService(client)
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
		Use:   "get <group-id>",
		Short: "Get a single group by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.GroupID(args[0]); err != nil {
				return err
			}
			groupID, _ := strconv.ParseInt(args[0], 10, 64)

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewGroupService(client)
			result, err := svc.Get(context.Background(), groupID)
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
		Short: "Create a new group",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := readPayload(cmd)
			if err != nil {
				return err
			}

			client, err := newClientNoRetry(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewGroupService(client)
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
		Use:   "update <group-id>",
		Short: "Update an existing group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.GroupID(args[0]); err != nil {
				return err
			}
			groupID, _ := strconv.ParseInt(args[0], 10, 64)

			payload, err := readPayload(cmd)
			if err != nil {
				return err
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewGroupService(client)
			result, err := svc.Update(context.Background(), groupID, payload)
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
		Use:   "delete <group-id>",
		Short: "Delete a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.GroupID(args[0]); err != nil {
				return err
			}
			groupID, _ := strconv.ParseInt(args[0], 10, 64)

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewGroupService(client)
			result, err := svc.Delete(context.Background(), groupID)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}
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
			WithSuggestion("Provide valid JSON object, e.g. --json '{\"GroupName\": \"mygroup\"}'")
	}

	return payload, nil
}
