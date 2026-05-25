package sitediff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func mustDiff(t *testing.T, oldS, newS string, opts Options) *SiteDiff {
	t.Helper()
	d, err := Diff(mustParse(t, oldS), mustParse(t, newS), opts)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func runFormat(t *testing.T, f Formatter, d *SiteDiff) string {
	t.Helper()
	var buf bytes.Buffer
	if err := f.Format(&buf, d); err != nil {
		t.Fatalf("format: %v", err)
	}
	return buf.String()
}

func TestNewFormatter(t *testing.T) {
	cases := map[string]string{
		"":         "*sitediff.TextFormatter",
		"text":     "*sitediff.TextFormatter",
		"json":     "*sitediff.JSONFormatter",
		"markdown": "*sitediff.MarkdownFormatter",
		"md":       "*sitediff.MarkdownFormatter",
	}
	for in, want := range cases {
		f, err := NewFormatter(in)
		if err != nil {
			t.Errorf("NewFormatter(%q): %v", in, err)
			continue
		}
		got := fmt.Sprintf("%T", f)
		// fmt.Sprintf returns the fully-qualified package name; trim it for clarity.
		got = strings.Replace(got, "github.com/immmmmmmu/plsnt/internal/", "", 1)
		if got != want {
			t.Errorf("NewFormatter(%q) = %s, want %s", in, got, want)
		}
	}
	if _, err := NewFormatter("yaml"); err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestTextFormatter_NoChanges(t *testing.T) {
	out := runFormat(t, &TextFormatter{}, &SiteDiff{})
	if !strings.Contains(out, "No changes") {
		t.Errorf("expected 'No changes', got %q", out)
	}
}

func TestTextFormatter_ColumnsModified(t *testing.T) {
	d := mustDiff(t,
		`{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Columns":[{"ColumnName":"A","LabelText":"old"}]}}]}`,
		`{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Columns":[{"ColumnName":"A","LabelText":"new"}]}}]}`,
		Options{},
	)
	out := runFormat(t, &TextFormatter{}, d)
	if !strings.Contains(out, "[1] S") {
		t.Errorf("missing site header:\n%s", out)
	}
	if !strings.Contains(out, "ColumnName=A") || !strings.Contains(out, "\"old\"") || !strings.Contains(out, "\"new\"") {
		t.Errorf("missing change details:\n%s", out)
	}
}

func TestTextFormatter_AddedSite(t *testing.T) {
	d := mustDiff(t,
		`{"Sites":[{"SiteId":1,"Title":"keep"}]}`,
		`{"Sites":[{"SiteId":1,"Title":"keep"},{"SiteId":2,"Title":"new"}]}`,
		Options{},
	)
	out := runFormat(t, &TextFormatter{}, d)
	if !strings.Contains(out, "+ (added)") || !strings.Contains(out, "[2] new") {
		t.Errorf("missing added marker:\n%s", out)
	}
}

func TestTextFormatter_UnifiedDiffIndented(t *testing.T) {
	d := mustDiff(t,
		`{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Scripts":[{"Title":"x","Body":"a\nb\nc\n"}]}}]}`,
		`{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Scripts":[{"Title":"x","Body":"a\nB\nc\n"}]}}]}`,
		Options{},
	)
	out := runFormat(t, &TextFormatter{}, d)
	if !strings.Contains(out, "    -b") || !strings.Contains(out, "    +B") {
		t.Errorf("unified diff not indented properly:\n%s", out)
	}
}

func TestJSONFormatter_RoundTrip(t *testing.T) {
	d := mustDiff(t,
		`{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Columns":[{"ColumnName":"A","LabelText":"old"}]}}]}`,
		`{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Columns":[{"ColumnName":"A","LabelText":"new"}]}}]}`,
		Options{},
	)
	out := runFormat(t, &JSONFormatter{}, d)
	var got SiteDiff
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out)
	}
	if len(got.Modified) != 1 {
		t.Errorf("round-trip lost modifications: %+v", got)
	}
}

func TestMarkdownFormatter_NoChanges(t *testing.T) {
	out := runFormat(t, &MarkdownFormatter{}, &SiteDiff{})
	if !strings.Contains(out, "_No changes._") {
		t.Errorf("expected italic 'No changes', got %q", out)
	}
}

func TestMarkdownFormatter_DiffCodeBlock(t *testing.T) {
	d := mustDiff(t,
		`{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Scripts":[{"Title":"x","Body":"a\nb\nc\n"}]}}]}`,
		`{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Scripts":[{"Title":"x","Body":"a\nB\nc\n"}]}}]}`,
		Options{},
	)
	out := runFormat(t, &MarkdownFormatter{}, d)
	if !strings.Contains(out, "```diff") {
		t.Errorf("missing fenced ```diff block:\n%s", out)
	}
	if !strings.Contains(out, "## Site `[1]` S") {
		t.Errorf("missing per-site heading:\n%s", out)
	}
}
