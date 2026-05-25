package sitediff

import (
	"fmt"
	"io"
	"strings"
)

// TextFormatter renders a SiteDiff in human-readable form using +/-/~/→
// markers and a unified-diff body for text changes. Output is deterministic.
type TextFormatter struct{}

func (TextFormatter) Format(w io.Writer, d *SiteDiff) error {
	if d == nil || !d.HasChanges() {
		_, err := fmt.Fprintln(w, "No changes.")
		return err
	}

	if len(d.HeaderChanges) > 0 {
		fmt.Fprintln(w, "== Header ==")
		for _, c := range d.HeaderChanges {
			writeChange(w, c, "  ")
		}
		fmt.Fprintln(w)
	}

	if len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Moved) > 0 {
		fmt.Fprintln(w, "== Sites ==")
		for _, r := range d.Added {
			fmt.Fprintf(w, "  + (added)   [%d] %s\n", r.SiteID, r.Title)
		}
		for _, r := range d.Removed {
			fmt.Fprintf(w, "  - (removed) [%d] %s\n", r.SiteID, r.Title)
		}
		for _, m := range d.Moved {
			fmt.Fprintf(w, "  → (moved)   [%d] %s   ParentId: %d → %d\n",
				m.SiteID, m.Title, m.OldParentID, m.NewParentID)
		}
		fmt.Fprintln(w)
	}

	for _, p := range d.Modified {
		fmt.Fprintf(w, "== [%d] %s ==\n", p.SiteID, p.Title)
		for _, c := range p.Changes {
			writeChange(w, c, "  ")
		}
		fmt.Fprintln(w)
	}
	return nil
}

func writeChange(w io.Writer, c FieldChange, indent string) {
	switch c.Kind {
	case ChangeAdded:
		fmt.Fprintf(w, "%s+ %s = %s\n", indent, c.Path, formatValue(c.NewValue))
	case ChangeRemoved:
		fmt.Fprintf(w, "%s- %s = %s\n", indent, c.Path, formatValue(c.OldValue))
	case ChangeModified:
		fmt.Fprintf(w, "%s~ %s: %s → %s\n", indent, c.Path,
			formatValue(c.OldValue), formatValue(c.NewValue))
	case ChangeTextDiff:
		fmt.Fprintf(w, "%s~ %s:\n", indent, c.Path)
		// Indent each line of the unified diff so it nests under the change.
		body := strings.TrimRight(c.UnifiedDiff, "\n")
		for _, line := range strings.Split(body, "\n") {
			fmt.Fprintf(w, "%s    %s\n", indent, line)
		}
	}
}

// formatValue renders a scalar (or nested object) as a single short line.
// Long objects are summarised so the output stays readable even when an
// entire SiteSettings subtree was added or removed.
func formatValue(v any) string {
	switch x := v.(type) {
	case nil:
		return "(none)"
	case string:
		if strings.ContainsAny(x, "\n\r") {
			return fmt.Sprintf("(multiline, %d chars)", len(x))
		}
		if len(x) > 80 {
			return fmt.Sprintf("%q…", x[:77])
		}
		return fmt.Sprintf("%q", x)
	case map[string]any:
		return fmt.Sprintf("(object, %d keys)", len(x))
	case []any:
		return fmt.Sprintf("(array, %d items)", len(x))
	case bool:
		return fmt.Sprintf("%v", x)
	default:
		return fmt.Sprintf("%v", x)
	}
}
