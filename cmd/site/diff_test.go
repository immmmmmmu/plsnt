package site

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/sitediff"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

// captureStdout temporarily redirects os.Stdout for the duration of fn.
// We avoid testing the cobra Command directly so the test stays focused on
// our actual logic; runDiff is the integration boundary.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	done := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.Bytes()
	}()
	fn()
	_ = w.Close()
	os.Stdout = orig
	return string(<-done)
}

func TestRunDiff_NoChanges(t *testing.T) {
	pkg := `{"HeaderInfo":{},"Sites":[{"SiteId":1,"Title":"S"}]}`
	a := writeTemp(t, "a.json", pkg)
	b := writeTemp(t, "b.json", pkg)
	var err error
	out := captureStdout(t, func() {
		err = runDiff(a, b, runOpts{format: "text", maxSize: "1MB"})
	})
	if err != nil {
		t.Fatalf("runDiff: %v", err)
	}
	if !strings.Contains(out, "No changes") {
		t.Errorf("expected 'No changes', got %q", out)
	}
}

func TestRunDiff_ExitCodeWithDiff(t *testing.T) {
	a := writeTemp(t, "a.json", `{"Sites":[{"SiteId":1,"Title":"old"}]}`)
	b := writeTemp(t, "b.json", `{"Sites":[{"SiteId":1,"Title":"new"}]}`)
	var err error
	captureStdout(t, func() {
		err = runDiff(a, b, runOpts{format: "text", exitCode: true, maxSize: "1MB"})
	})
	if !errors.Is(err, sitediff.ErrDifferencesFound) {
		t.Fatalf("expected ErrDifferencesFound, got %v", err)
	}
}

func TestRunDiff_NoExitCodeMeansNoError(t *testing.T) {
	a := writeTemp(t, "a.json", `{"Sites":[{"SiteId":1,"Title":"old"}]}`)
	b := writeTemp(t, "b.json", `{"Sites":[{"SiteId":1,"Title":"new"}]}`)
	var err error
	captureStdout(t, func() {
		err = runDiff(a, b, runOpts{format: "text", exitCode: false, maxSize: "1MB"})
	})
	if err != nil {
		t.Fatalf("without --exit-code, runDiff must not return error, got %v", err)
	}
}

func TestRunDiff_BadFormat(t *testing.T) {
	a := writeTemp(t, "a.json", `{}`)
	b := writeTemp(t, "b.json", `{}`)
	var err error
	captureStdout(t, func() {
		err = runDiff(a, b, runOpts{format: "yaml", maxSize: "1MB"})
	})
	if err == nil || !strings.Contains(err.Error(), "yaml") {
		t.Fatalf("expected format error, got %v", err)
	}
}

func TestRunDiff_FileNotFound(t *testing.T) {
	err := runDiff("/no/such/old.json", "/no/such/new.json", runOpts{format: "text", maxSize: "1MB"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestRunDiff_MarkdownFormat(t *testing.T) {
	a := writeTemp(t, "a.json", `{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Columns":[{"ColumnName":"A","LabelText":"old"}]}}]}`)
	b := writeTemp(t, "b.json", `{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Columns":[{"ColumnName":"A","LabelText":"new"}]}}]}`)
	var err error
	out := captureStdout(t, func() {
		err = runDiff(a, b, runOpts{format: "markdown", maxSize: "1MB"})
	})
	if err != nil {
		t.Fatalf("runDiff: %v", err)
	}
	if !strings.Contains(out, "# SitePackage diff") {
		t.Errorf("missing markdown header:\n%s", out)
	}
}

func TestParseSize(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		err  bool
	}{
		{"", 0, false},
		{"100", 100, false},
		{"100B", 100, false},
		{"1KB", 1024, false},
		{"50MB", 50 << 20, false},
		{"2GB", 2 << 30, false},
		{"1.5MB", 1024*1024 + 1024*1024/2, false},
		{"  10MB  ", 10 << 20, false},
		{"abc", 0, true},
		{"MB", 0, true},
	}
	for _, tc := range cases {
		got, err := parseSize(tc.in)
		if tc.err {
			if err == nil {
				t.Errorf("parseSize(%q) expected error, got %d", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseSize(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseSize(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestSplitCSV(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"a", []string{"a"}},
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b , ", []string{"a", "b"}},
		{",,", nil},
	}
	for _, tc := range cases {
		got := splitCSV(tc.in)
		if !equalStrings(got, tc.want) {
			t.Errorf("splitCSV(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
