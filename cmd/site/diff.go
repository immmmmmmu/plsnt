package site

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/sitediff"
)

func newDiffCmd() *cobra.Command {
	var ro runOpts

	cmd := &cobra.Command{
		Use:   "diff <old.json> <new.json>",
		Short: "Compare two SitePackage JSON files locally (no API calls)",
		Long: `Diff two SitePackage JSON files exported from the Pleasanter Web UI.

Pleasanter's REST API does not expose SitePackage import/export — the only
way to obtain a SitePackage JSON is the Web UI's "Export site package"
action. This command takes two such files and reports the structural
differences between them.

The diff absorbs cosmetic noise:
  * Array elements are matched by semantic key (ColumnName, Title, Name)
    so reordered Columns / Scripts / Views do not surface as changes.
  * SiteId, TenantId, exporter timestamps and similar metadata are ignored
    by default. Use --no-ignore-default to see them.

Examples:
  plsnt site diff before.json after.json
  plsnt site diff before.json after.json --format markdown > review.md
  plsnt site diff before.json after.json --ignore LabelText
  plsnt site diff before.json after.json --ignore-path /Sites/*/SiteSettings/Version
  plsnt site diff before.json after.json --exit-code      # diff(1)-style exit 1`,
		Args:          cobra.ExactArgs(2),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runDiff(args[0], args[1], ro)
			if err == nil {
				return nil
			}
			// ErrDifferencesFound is the only error we surface silently —
			// the diff body has already been printed on stdout. For every
			// other error we print to stderr ourselves (we asked cobra to
			// stay quiet via SilenceErrors).
			if errors.Is(err, sitediff.ErrDifferencesFound) {
				return err
			}
			fmt.Fprintln(os.Stderr, "Error:", err)
			return err
		},
	}

	cmd.Flags().StringVar(&ro.format, "format", "text", "output format: text|json|markdown")
	cmd.Flags().StringVar(&ro.ignoreKeys, "ignore", "", "comma-separated leaf keys to ignore at any depth")
	cmd.Flags().StringVar(&ro.ignorePaths, "ignore-path", "", "comma-separated JSONPath globs to ignore (e.g. /Sites/*/SiteSettings/Version)")
	cmd.Flags().BoolVar(&ro.noIgnoreDefault, "no-ignore-default", false, "disable the built-in ignore list (SiteId, timestamps, exporter metadata)")
	cmd.Flags().StringVar(&ro.matchBy, "match-by", "auto", "Sites[] matching: auto (SiteId then Title), siteid, or title")
	cmd.Flags().BoolVar(&ro.exitCode, "exit-code", false, "exit 1 if differences are found (diff(1) style; exit 0 otherwise)")
	cmd.Flags().BoolVar(&ro.strict, "strict", false, "treat duplicate semantic keys as errors")
	cmd.Flags().BoolVar(&ro.includePerms, "include-permissions", false, "compare Permissions[] arrays (off by default — produces noisy diffs)")
	cmd.Flags().StringVar(&ro.maxSize, "max-size", "50MB", "per-file size limit (e.g. 10MB, 200MB, 1GB)")
	return cmd
}

type runOpts struct {
	format          string
	ignoreKeys      string
	ignorePaths     string
	noIgnoreDefault bool
	matchBy         string
	exitCode        bool
	strict          bool
	includePerms    bool
	maxSize         string
}

func runDiff(oldPath, newPath string, ro runOpts) error {
	maxBytes, err := parseSize(ro.maxSize)
	if err != nil {
		return err
	}
	opts := sitediff.Options{
		IgnoreKeys:         splitCSV(ro.ignoreKeys),
		IgnorePaths:        splitCSV(ro.ignorePaths),
		NoDefaultIgnore:    ro.noIgnoreDefault,
		MatchSitesBy:       ro.matchBy,
		Strict:             ro.strict,
		IncludePermissions: ro.includePerms,
		MaxBytes:           maxBytes,
	}

	oldPkg, err := sitediff.Load(oldPath, opts.MaxBytes)
	if err != nil {
		return err
	}
	newPkg, err := sitediff.Load(newPath, opts.MaxBytes)
	if err != nil {
		return err
	}

	d, err := sitediff.Diff(oldPkg, newPkg, opts)
	if err != nil {
		return errs.New(errs.CodeInternalError, err.Error())
	}

	f, err := sitediff.NewFormatter(ro.format)
	if err != nil {
		return errs.New(errs.CodeValidationError, err.Error()).
			WithSuggestion("Use one of: text, json, markdown")
	}
	if err := f.Format(os.Stdout, d); err != nil {
		return errs.New(errs.CodeInternalError, err.Error())
	}

	if ro.exitCode && d.HasChanges() {
		return sitediff.ErrDifferencesFound
	}
	return nil
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseSize accepts strings like "50MB", "1.5GB", "1024KB", or a plain
// integer (bytes). Case-insensitive. Empty string → 0 (caller falls back
// to the package default).
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	upper := strings.ToUpper(s)
	mult := int64(1)
	switch {
	case strings.HasSuffix(upper, "GB"):
		mult = 1 << 30
		s = upper[:len(upper)-2]
	case strings.HasSuffix(upper, "MB"):
		mult = 1 << 20
		s = upper[:len(upper)-2]
	case strings.HasSuffix(upper, "KB"):
		mult = 1 << 10
		s = upper[:len(upper)-2]
	case strings.HasSuffix(upper, "B"):
		s = upper[:len(upper)-1]
	default:
		s = upper
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errs.New(errs.CodeValidationError, "max-size: numeric component is empty").
			WithSuggestion("Use forms like 50MB, 1GB, or a raw byte count")
	}
	dot := strings.Index(s, ".")
	if dot < 0 {
		whole, err := parseInt64(s)
		if err != nil {
			return 0, sizeErr(s)
		}
		return whole * mult, nil
	}
	whole, err := parseInt64(s[:dot])
	if err != nil {
		return 0, sizeErr(s)
	}
	rest := s[dot+1:]
	if rest == "" {
		return whole * mult, nil
	}
	if len(rest) > 3 {
		rest = rest[:3]
	}
	for len(rest) < 3 {
		rest += "0"
	}
	frac, err := parseInt64(rest)
	if err != nil {
		return 0, sizeErr(s)
	}
	return whole*mult + frac*mult/1000, nil
}

func parseInt64(s string) (int64, error) {
	var n int64
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, errs.New(errs.CodeValidationError, "not a positive integer")
		}
		n = n*10 + int64(r-'0')
	}
	return n, nil
}

func sizeErr(s string) error {
	return errs.New(errs.CodeValidationError, "invalid --max-size: "+s).
		WithSuggestion("Use forms like 50MB, 1GB, or a raw byte count")
}
