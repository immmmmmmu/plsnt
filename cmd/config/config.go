package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/config"
	"github.com/immmmmmmu/plsnt/internal/errs"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration profiles",
	}

	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newUseCmd())
	cmd.AddCommand(newTestCmd())
	cmd.AddCommand(newMCPSetupCmd())
	return cmd
}

func newSetCmd() *cobra.Command {
	var (
		url        string
		apiKey     string
		apiVersion string
		name       string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set profile configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				name = "default"
			}

			cfgPath := config.DefaultPath()
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			p, exists := cfg.Profiles[name]
			if !exists {
				p = &config.Profile{}
				cfg.Profiles[name] = p
			}

			if url != "" {
				// Warn on HTTP
				if strings.HasPrefix(url, "http://") {
					fmt.Fprintln(os.Stderr, "WARNING: Using HTTP (not HTTPS). API key will be sent in plain text.")
				}
				p.URL = url
			}
			if apiKey != "" {
				p.APIKey = apiKey
			}
			if apiVersion != "" {
				p.APIVersion = apiVersion
			}

			if cfg.CurrentProfile == "" {
				cfg.CurrentProfile = name
			}

			if err := config.Save(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Profile %q saved to %s\n", name, cfgPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "default", "profile name")
	cmd.Flags().StringVar(&url, "url", "", "Pleasanter server URL")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key")
	cmd.Flags().StringVar(&apiVersion, "api-version", "1.1", "API version")

	return cmd
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.DefaultPath())
			if err != nil {
				return err
			}

			if len(cfg.Profiles) == 0 {
				fmt.Fprintln(os.Stderr, "No profiles configured. Use 'plsnt config set' to create one.")
				return nil
			}

			for name, p := range cfg.Profiles {
				active := " "
				if name == cfg.CurrentProfile {
					active = "*"
				}
				fmt.Printf("%s %s\n  url: %s\n  api_key: %s\n  api_version: %s\n",
					active, name, p.URL, config.MaskAPIKey(p.APIKey), p.APIVersion)
			}
			return nil
		},
	}
}

func newUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use [profile-name]",
		Short: "Switch active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfgPath := config.DefaultPath()
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			if _, ok := cfg.Profiles[name]; !ok {
				e := errs.New(errs.CodeValidationError,
					fmt.Sprintf("Profile %q not found", name))
				profiles := make([]string, 0, len(cfg.Profiles))
				for k := range cfg.Profiles {
					profiles = append(profiles, k)
				}
				return e.WithSuggestion(
					fmt.Sprintf("Available profiles: %s", strings.Join(profiles, ", ")))
			}

			cfg.CurrentProfile = name
			if err := config.Save(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Switched to profile %q\n", name)
			return nil
		},
	}
}

func newTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test connection to Pleasanter server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.DefaultPath())
			if err != nil {
				return err
			}

			profile, name, err := cfg.ActiveProfile()
			if err != nil {
				return errs.New(errs.CodeValidationError, err.Error()).
					WithSuggestion("Run 'plsnt config set' to configure a profile")
			}

			url, apiKey, apiVersion := profile.Resolve()
			if url == "" || apiKey == "" {
				return errs.New(errs.CodeValidationError, "URL and API key are required").
					WithSuggestion("Run 'plsnt config set --url <url> --api-key <key>'")
			}

			fmt.Fprintf(os.Stderr, "Testing connection to %s (profile: %s)...\n", url, name)

			var opts []api.Option
			if insecure, _ := cmd.Flags().GetBool("insecure"); insecure {
				opts = append(opts, api.WithInsecure())
			}
			client := api.New(url, apiKey, apiVersion, opts...)

			// Use /api/users/get as a connectivity test. This endpoint requires auth
			// and returns user list, confirming both connectivity and API key validity.
			resp, err := client.PostRaw(context.Background(), "/api/users/get", map[string]any{})
			if err != nil {
				return err
			}

			statusCode := 0
			if sc, ok := resp["StatusCode"]; ok {
				switch v := sc.(type) {
				case float64:
					statusCode = int(v)
				case json.Number:
					n, _ := v.Int64()
					statusCode = int(n)
				}
			}

			if statusCode == 200 {
				fmt.Fprintln(os.Stderr, "Connection successful. API key is valid.")
			} else {
				fmt.Fprintf(os.Stderr, "Server responded but returned status %d.\n", statusCode)
			}
			return nil
		},
	}
}
