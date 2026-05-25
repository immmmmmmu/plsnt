package sitediff

import (
	"fmt"
	"io"
	"strings"
)

// MarkdownFormatter renders a SiteDiff in GitHub-flavoured markdown,
// suitable for pasting into a PR comment. Uses headings per site and a
// fenced code block for unified diffs (with the "diff" language hint so
// GitHub colourises +/- lines).
type MarkdownFormatter struct{}

func (MarkdownFormatter) Format(w io.Writer, d *SiteDiff) error {
	if d == nil || !d.HasChanges() {
		_, err := fmt.Fprintln(w, "_No changes._")
		return err
	}

	fmt.Fprintln(w, "# SitePackage diff")
	fmt.Fprintln(w)

	if len(d.HeaderChanges) > 0 {
		fmt.Fprintln(w, "## Header")
		fmt.Fprintln(w)
		writeMarkdownChanges(w, d.HeaderChanges)
		fmt.Fprintln(w)
	}

	if len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Moved) > 0 {
		fmt.Fprintln(w, "## Sites")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "| Change | SiteId | Title | Note |")
		fmt.Fprintln(w, "|---|---|---|---|")
		for _, r := range d.Added {
			fmt.Fprintf(w, "| ➕ added | %d | %s | |\n", r.SiteID, escapeMarkdownTableCell(r.Title))
		}
		for _, r := range d.Removed {
			fmt.Fprintf(w, "| ➖ removed | %d | %s | |\n", r.SiteID, escapeMarkdownTableCell(r.Title))
		}
		for _, m := range d.Moved {
			fmt.Fprintf(w, "| 🔀 moved | %d | %s | ParentId: %d → %d |\n",
				m.SiteID, escapeMarkdownTableCell(m.Title), m.OldParentID, m.NewParentID)
		}
		fmt.Fprintln(w)
	}

	for _, p := range d.Modified {
		fmt.Fprintf(w, "## Site `[%d]` %s\n\n", p.SiteID, p.Title)
		writeMarkdownChanges(w, p.Changes)
		fmt.Fprintln(w)
	}
	return nil
}

func writeMarkdownChanges(w io.Writer, changes []FieldChange) {
	// Split text-diff changes (rendered as code blocks) from inline scalar
	// changes (rendered as a list) so the surrounding markup stays compact.
	var inline []FieldChange
	var blocks []FieldChange
	for _, c := range changes {
		if c.Kind == ChangeTextDiff {
			blocks = append(blocks, c)
		} else {
			inline = append(inline, c)
		}
	}

	for _, c := range inline {
		switch c.Kind {
		case ChangeAdded:
			fmt.Fprintf(w, "- ➕ `%s` = %s\n", c.Path, formatMarkdownValue(c.NewValue))
		case ChangeRemoved:
			fmt.Fprintf(w, "- ➖ `%s` = %s\n", c.Path, formatMarkdownValue(c.OldValue))
		case ChangeModified:
			fmt.Fprintf(w, "- ✏️ `%s`: %s → %s\n", c.Path,
				formatMarkdownValue(c.OldValue), formatMarkdownValue(c.NewValue))
		}
	}

	for _, c := range blocks {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "**`%s`**\n", c.Path)
		fmt.Fprintln(w, "```diff")
		fmt.Fprint(w, c.UnifiedDiff)
		if !strings.HasSuffix(c.UnifiedDiff, "\n") {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, "```")
	}
}

func formatMarkdownValue(v any) string {
	s := formatValue(v)
	// Already quoted if string; no further escape needed for backticks
	// because formatValue uses %q which escapes them.
	return s
}

func escapeMarkdownTableCell(s string) string {
	// Pipes and newlines break tables. Replace with safer equivalents.
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
