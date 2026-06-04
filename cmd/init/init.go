// Package initcmd implements the `plsnt init` command which bootstraps the Core
// plsnt skill set into a project (or user) directory for a chosen agent app,
// and optionally configures a connection profile and .mcp.json.
package initcmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/immmmmmmu/plsnt/internal/bootstrap"
	"github.com/immmmmmmu/plsnt/internal/config"
	"github.com/immmmmmmu/plsnt/internal/errs"
)

type options struct {
	baseDir     string
	agent       bootstrap.Agent
	url         string
	apiKey      string
	profileName string
	withMCP     bool
	stdout      io.Writer
	stderr      io.Writer
}

// NewCmd builds the `plsnt init` cobra command.
func NewCmd() *cobra.Command {
	var (
		agentFlag   string
		scopeFlag   string
		url         string
		apiKey      string
		profileName string
		withMCP     bool
		assumeYes   bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap the Core plsnt skill set into this project",
		Long: `init writes the Core plsnt skills into the current project (or your home
directory) so an AI agent can operate plsnt effectively.

The skills are plain Markdown and work with any agent app:
  --agent claude    .claude/skills/<name>/SKILL.md (default)
  --agent codex     AGENTS.md
  --agent gemini    GEMINI.md
  --agent generic   AGENTS.md

It can also save a connection profile (--url/--api-key) and generate .mcp.json (--mcp).
Run interactively (a TTY) to be prompted for any value not passed as a flag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			interactive := !assumeYes && isStdinTTY()
			in := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			// Resolve agent.
			if agentFlag == "" && interactive {
				agentFlag = promptChoice(in, out, "Target agent app",
					[]string{"claude", "codex", "gemini", "generic"}, "claude")
			}
			if agentFlag == "" {
				agentFlag = "claude"
			}
			agent, err := bootstrap.ParseAgent(agentFlag)
			if err != nil {
				return errs.New(errs.CodeValidationError, err.Error())
			}

			// Resolve scope -> baseDir.
			if scopeFlag == "" && interactive {
				scopeFlag = promptChoice(in, out, "Install location",
					[]string{"project", "user"}, "project")
			}
			if scopeFlag == "" {
				scopeFlag = "project"
			}
			baseDir, err := resolveBaseDir(scopeFlag)
			if err != nil {
				return err
			}

			// Optionally collect connection settings.
			if interactive {
				if url == "" {
					url = promptString(in, out, "Pleasanter URL (blank to skip)", "")
				}
				if url != "" && apiKey == "" {
					apiKey = promptString(in, out, "API key (blank to skip)", "")
				}
				if !cmd.Flags().Changed("mcp") {
					withMCP = promptYesNo(in, out, "Generate .mcp.json?", false)
				}
			}

			return runInit(options{
				baseDir:     baseDir,
				agent:       agent,
				url:         url,
				apiKey:      apiKey,
				profileName: profileName,
				withMCP:     withMCP,
				stdout:      out,
				stderr:      cmd.ErrOrStderr(),
			})
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "target agent app: claude, codex, gemini, generic (default claude)")
	cmd.Flags().StringVar(&scopeFlag, "scope", "", "install location: project (./) or user (~/) (default project)")
	cmd.Flags().StringVar(&url, "url", "", "Pleasanter server URL to save in a profile")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key to save in a profile")
	cmd.Flags().StringVar(&profileName, "profile", "default", "profile name to save")
	cmd.Flags().BoolVar(&withMCP, "mcp", false, "also generate .mcp.json for the plsnt MCP server")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "non-interactive: use flags and defaults without prompting")

	return cmd
}

// runInit performs the bootstrap with fully-resolved options (no prompting).
func runInit(opts options) error {
	written, err := bootstrap.Install(opts.baseDir, opts.agent)
	if err != nil {
		return errs.Wrap(err, errs.CodeInternalError)
	}

	skills, _ := bootstrap.SkillNames()
	rules, _ := bootstrap.Rules()
	if opts.agent == bootstrap.AgentClaude {
		agents, _ := bootstrap.Agents()
		cmds, _ := bootstrap.Commands()
		fmt.Fprintf(opts.stderr, "Installed for agent %q into %s: %d skills, %d sub-agents, %d commands, %d rules\n",
			opts.agent, opts.baseDir, len(skills), len(agents), len(cmds), len(rules))
	} else {
		fmt.Fprintf(opts.stderr, "Installed for agent %q into %s: %d skills + %d rules folded into the bundle\n",
			opts.agent, opts.baseDir, len(skills), len(rules))
	}
	for _, p := range written {
		rel, relErr := filepath.Rel(opts.baseDir, p)
		if relErr != nil {
			rel = p
		}
		fmt.Fprintf(opts.stderr, "  %s\n", rel)
	}

	// Save a connection profile when credentials are supplied.
	if opts.url != "" || opts.apiKey != "" {
		if err := saveProfile(opts); err != nil {
			return err
		}
	}

	// Generate .mcp.json when requested.
	if opts.withMCP {
		if err := writeMCPConfig(opts); err != nil {
			return err
		}
	}

	fmt.Fprintln(opts.stderr, "Done. Your agent can now use the plsnt skills.")
	return nil
}

func saveProfile(opts options) error {
	name := opts.profileName
	if name == "" {
		name = "default"
	}
	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	p, ok := cfg.Profiles[name]
	if !ok {
		p = &config.Profile{}
		cfg.Profiles[name] = p
	}
	if opts.url != "" {
		if config.IsHTTP(opts.url) {
			fmt.Fprintln(opts.stderr, "WARNING: Using HTTP (not HTTPS). API key will be sent in plain text.")
		}
		p.URL = opts.url
	}
	if opts.apiKey != "" {
		p.APIKey = opts.apiKey
	}
	if cfg.CurrentProfile == "" {
		cfg.CurrentProfile = name
	}
	if err := config.Save(cfgPath, cfg); err != nil {
		return err
	}
	fmt.Fprintf(opts.stderr, "Profile %q saved to %s\n", name, cfgPath)
	return nil
}

func writeMCPConfig(opts options) error {
	var mcpConfig *config.MCPServerConfig
	if opts.profileName != "" && opts.profileName != "default" {
		mcpConfig = config.GeneratePlsntMCPConfigWithProfile(opts.profileName)
	} else {
		mcpConfig = config.GeneratePlsntMCPConfig()
	}
	jsonBytes, err := mcpConfig.MarshalJSON()
	if err != nil {
		return errs.Wrap(err, errs.CodeInternalError)
	}
	jsonBytes = append(jsonBytes, '\n')

	path := filepath.Join(opts.baseDir, ".mcp.json")
	if err := os.WriteFile(path, jsonBytes, 0o644); err != nil {
		return errs.Wrap(err, errs.CodeInternalError)
	}
	fmt.Fprintf(opts.stderr, "Wrote %s\n", path)
	return nil
}

func resolveBaseDir(scope string) (string, error) {
	switch scope {
	case "project", "":
		cwd, err := os.Getwd()
		if err != nil {
			return "", errs.Wrap(err, errs.CodeInternalError)
		}
		return cwd, nil
	case "user":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errs.Wrap(err, errs.CodeInternalError)
		}
		return home, nil
	default:
		return "", errs.New(errs.CodeValidationError,
			fmt.Sprintf("unknown scope %q (valid: project, user)", scope))
	}
}

func isStdinTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func promptChoice(in *bufio.Reader, out io.Writer, label string, choices []string, def string) string {
	fmt.Fprintf(out, "%s [%s] (default %s): ", label, strings.Join(choices, "/"), def)
	line, _ := in.ReadString('\n')
	val := strings.TrimSpace(line)
	if val == "" {
		return def
	}
	for _, c := range choices {
		if strings.EqualFold(val, c) {
			return c
		}
	}
	return def
}

func promptString(in *bufio.Reader, out io.Writer, label, def string) string {
	if def != "" {
		fmt.Fprintf(out, "%s (default %s): ", label, def)
	} else {
		fmt.Fprintf(out, "%s: ", label)
	}
	line, _ := in.ReadString('\n')
	val := strings.TrimSpace(line)
	if val == "" {
		return def
	}
	return val
}

func promptYesNo(in *bufio.Reader, out io.Writer, label string, def bool) bool {
	d := "y/N"
	if def {
		d = "Y/n"
	}
	fmt.Fprintf(out, "%s [%s]: ", label, d)
	line, _ := in.ReadString('\n')
	val := strings.ToLower(strings.TrimSpace(line))
	if val == "" {
		return def
	}
	return val == "y" || val == "yes"
}
