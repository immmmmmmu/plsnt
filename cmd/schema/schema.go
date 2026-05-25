package schema

import (
	"context"
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

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema <site-id>",
		Short: "Show column definitions for a site",
		Long:  "Retrieve and display column definitions (schema) for a Pleasanter site using the getsite API.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			siteIDStr := args[0]
			if err := validate.SiteID(siteIDStr); err != nil {
				return err
			}

			siteID, _ := strconv.ParseInt(siteIDStr, 10, 64)

			cfgPath := config.DefaultPath()
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			profileFlag, _ := cmd.Flags().GetString("profile")
			profile, _, err := cfg.ActiveProfileWithOverride(profileFlag)
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
			svc := pleasanter.NewSchemaService(client)

			info, err := svc.GetSchema(context.Background(), siteID)
			if err != nil {
				return err
			}

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

			formatter, err := format.New(output, fields)
			if err != nil {
				return err
			}

			return formatter.Format(os.Stdout, info)
		},
	}

	return cmd
}
