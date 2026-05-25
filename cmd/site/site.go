package site

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

// NewCmd creates the "site" parent command.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "site",
		Short: "Manage sites",
	}

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newCopyCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newDiffCmd())
	return cmd
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <site-id>",
		Short: "Get site info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.SiteID(args[0]); err != nil {
				return err
			}
			siteID, _ := strconv.ParseInt(args[0], 10, 64)

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewSiteService(client)
			result, err := svc.Get(context.Background(), siteID)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}
	return cmd
}

func newCreateCmd() *cobra.Command {
	var parentID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new site",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.ParentID(parentID, "parent-id"); err != nil {
				return err
			}
			pid, _ := strconv.ParseInt(parentID, 10, 64)

			payload, err := readPayload(cmd)
			if err != nil {
				return err
			}

			client, err := newClientNoRetry(cmd)
			if err != nil {
				return err
			}

			// Check that parent is a folder (Sites), not a table (Results/Issues)
			svc := pleasanter.NewSiteService(client)
			parentInfo, parentErr := svc.Get(context.Background(), pid)
			if parentErr == nil {
				refType := extractReferenceType(parentInfo)
				if refType != "" && refType != "Sites" {
					fmt.Fprintf(os.Stderr, "WARNING: parent-id %d is a %s table, not a folder (Sites). This may create a record instead of a site.\n", pid, refType)
					fmt.Fprintln(os.Stderr, "Use a folder (ReferenceType: Sites) as parent-id, or verify the result with 'site search'.")
				}
			}

			result, err := svc.Create(context.Background(), pid, payload)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}

	cmd.Flags().StringVar(&parentID, "parent-id", "", "parent site ID (required)")
	_ = cmd.MarkFlagRequired("parent-id")
	return cmd
}

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <site-id>",
		Short: "Update site settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.SiteID(args[0]); err != nil {
				return err
			}
			siteID, _ := strconv.ParseInt(args[0], 10, 64)

			payload, err := readPayload(cmd)
			if err != nil {
				return err
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewSiteService(client)
			result, err := svc.Update(context.Background(), siteID, payload)
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
		Use:   "delete <site-id>",
		Short: "Delete a site",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.SiteID(args[0]); err != nil {
				return err
			}
			siteID, _ := strconv.ParseInt(args[0], 10, 64)

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewSiteService(client)
			result, err := svc.Delete(context.Background(), siteID)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}
	return cmd
}

func newCopyCmd() *cobra.Command {
	var parentID string
	var jsonOverrides string

	cmd := &cobra.Command{
		Use:   "copy <source-site-id>",
		Short: "Copy a site's settings to a new location",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.SiteID(args[0]); err != nil {
				return err
			}
			sourceSiteID, _ := strconv.ParseInt(args[0], 10, 64)

			if err := validate.ParentID(parentID, "parent-id"); err != nil {
				return err
			}
			destParentID, _ := strconv.ParseInt(parentID, 10, 64)

			var overrides map[string]any
			if jsonOverrides != "" {
				if err := validate.JSONSyntax(jsonOverrides); err != nil {
					return err
				}
				if err := json.Unmarshal([]byte(jsonOverrides), &overrides); err != nil {
					return errs.New(errs.CodeInvalidInput, "invalid JSON in --json flag").
						WithSuggestion("Provide valid JSON object, e.g. --json '{\"Title\": \"New Name\"}'")
				}
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewSiteService(client)
			result, err := svc.Copy(context.Background(), sourceSiteID, destParentID, overrides)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}

	cmd.Flags().StringVar(&parentID, "parent-id", "", "destination parent site ID (required)")
	_ = cmd.MarkFlagRequired("parent-id")
	cmd.Flags().StringVar(&jsonOverrides, "json", "", "JSON overrides (e.g. '{\"Title\":\"New Name\"}')")
	return cmd
}

func newSearchCmd() *cobra.Command {
	var parentID string
	var keyword string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search child sites by title keyword",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.ParentID(parentID, "parent-id"); err != nil {
				return err
			}
			pid, _ := strconv.ParseInt(parentID, 10, 64)

			if keyword == "" {
				return errs.New(errs.CodeInvalidInput, "keyword is required").
					WithSuggestion("Specify a search keyword, e.g. --keyword 'my site'")
			}

			client, err := newClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewSiteService(client)
			result, err := svc.Search(context.Background(), pid, keyword)
			if err != nil {
				return err
			}

			return outputResult(cmd, result)
		},
	}

	cmd.Flags().StringVar(&parentID, "parent-id", "", "parent site ID to search under (required)")
	_ = cmd.MarkFlagRequired("parent-id")
	cmd.Flags().StringVar(&keyword, "keyword", "", "title keyword to search for (required)")
	_ = cmd.MarkFlagRequired("keyword")
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
			WithSuggestion("Provide valid JSON object, e.g. --json '{\"Title\": \"test\"}'")
	}

	return payload, nil
}

// extractReferenceType extracts ReferenceType from a getsite API response.
// The response structure is: {"Response": {"Data": {"ReferenceType": "..."}}}
// or it may be at the top level: {"ReferenceType": "..."}
func extractReferenceType(info map[string]any) string {
	// Try top level first
	if rt, ok := info["ReferenceType"].(string); ok {
		return rt
	}
	// Try Response.Data path
	if resp, ok := info["Response"].(map[string]any); ok {
		if data, ok := resp["Data"].(map[string]any); ok {
			if rt, ok := data["ReferenceType"].(string); ok {
				return rt
			}
		}
	}
	return ""
}
