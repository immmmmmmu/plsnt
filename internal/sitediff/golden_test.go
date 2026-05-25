package sitediff

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// updateGolden controls whether golden_test re-writes the expected_*.txt
// fixtures with the current output. Run `go test -update ./internal/sitediff/...`
// after a deliberate output format change.
var updateGolden = flag.Bool("update", false, "update golden expected_*.txt files")

// goldenCases enumerates the testdata sub-directories the golden test
// walks. Each directory must contain old.json, new.json, optional
// opts.json (currently we keep options inline below), and the
// expected_*.txt files that this test asserts against.
type goldenCase struct {
	dir  string
	opts Options
}

func goldenCases() []goldenCase {
	return []goldenCase{
		{dir: "realistic_columns"},
		{dir: "sites_added_removed"},
		{dir: "scripts_unified_diff"},
		{dir: "ignore_default", opts: Options{MatchSitesBy: "title"}},
	}
}

func TestGolden_Text(t *testing.T) {
	for _, tc := range goldenCases() {
		t.Run(tc.dir, func(t *testing.T) {
			runGolden(t, tc, "text", &TextFormatter{}, "expected_text.txt")
		})
	}
}

func TestGolden_JSON(t *testing.T) {
	for _, tc := range goldenCases() {
		t.Run(tc.dir, func(t *testing.T) {
			runGolden(t, tc, "json", &JSONFormatter{}, "expected.json")
		})
	}
}

func TestGolden_Markdown(t *testing.T) {
	for _, tc := range goldenCases() {
		t.Run(tc.dir, func(t *testing.T) {
			runGolden(t, tc, "markdown", &MarkdownFormatter{}, "expected_markdown.md")
		})
	}
}

func runGolden(t *testing.T, tc goldenCase, label string, f Formatter, expectedFile string) {
	t.Helper()
	base := filepath.Join("testdata", tc.dir)

	old, err := Load(filepath.Join(base, "old.json"), 0)
	if err != nil {
		t.Fatalf("load old: %v", err)
	}
	newer, err := Load(filepath.Join(base, "new.json"), 0)
	if err != nil {
		t.Fatalf("load new: %v", err)
	}
	d, err := Diff(old, newer, tc.opts)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}

	var buf bytes.Buffer
	if err := f.Format(&buf, d); err != nil {
		t.Fatalf("Format(%s): %v", label, err)
	}
	got := buf.Bytes()

	expectedPath := filepath.Join(base, expectedFile)
	if *updateGolden {
		if err := os.WriteFile(expectedPath, got, 0o600); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated %s", expectedPath)
		return
	}

	want, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to create)", expectedPath, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s mismatch.\nGot:\n%s\nWant:\n%s\n(run with -update to refresh)",
			expectedFile, got, want)
	}
}
