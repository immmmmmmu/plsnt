package sitediff

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempJSON(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return p
}

func TestLoad_Success(t *testing.T) {
	p := writeTempJSON(t, "ok.json", `{"HeaderInfo":{"AssemblyVersion":"1.5.2.0"},"Sites":[]}`)
	got, err := Load(p, 0)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	hdr, ok := got["HeaderInfo"].(map[string]any)
	if !ok {
		t.Fatalf("HeaderInfo missing or wrong type: %T", got["HeaderInfo"])
	}
	if hdr["AssemblyVersion"] != "1.5.2.0" {
		t.Fatalf("AssemblyVersion mismatch: %v", hdr["AssemblyVersion"])
	}
}

func TestLoad_PreservesNumberPrecision(t *testing.T) {
	// Without UseNumber(), 1.017 round-trips through float64 and prints "1.017"
	// but 0.1+0.2 style payloads can lose precision. We assert it stays as
	// json.Number with the original textual form.
	p := writeTempJSON(t, "n.json", `{"Version":1.017,"Big":12345678901234567890}`)
	got, err := Load(p, 0)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	v, ok := got["Version"].(json.Number)
	if !ok {
		t.Fatalf("Version not json.Number: %T (%v)", got["Version"], got["Version"])
	}
	if v.String() != "1.017" {
		t.Fatalf("Version textual form lost: got %q want %q", v.String(), "1.017")
	}
	b, ok := got["Big"].(json.Number)
	if !ok {
		t.Fatalf("Big not json.Number: %T", got["Big"])
	}
	if b.String() != "12345678901234567890" {
		t.Fatalf("Big precision lost: got %q", b.String())
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/no/such/path/sitepkg.json", 0)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Fatalf("error did not mention not-found: %v", err)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	p := writeTempJSON(t, "bad.json", `{not json`)
	_, err := Load(p, 0)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("error did not mention invalid JSON: %v", err)
	}
}

func TestLoad_NullJSON(t *testing.T) {
	p := writeTempJSON(t, "null.json", `null`)
	_, err := Load(p, 0)
	if err == nil {
		t.Fatal("expected error for JSON null")
	}
}

func TestLoad_SizeLimit(t *testing.T) {
	// 200 bytes of payload, limit set to 100.
	p := writeTempJSON(t, "big.json", `{"x":"`+strings.Repeat("a", 200)+`"}`)
	_, err := Load(p, 100)
	if err == nil {
		t.Fatal("expected size-limit error")
	}
	if !strings.Contains(err.Error(), "exceeds size limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoad_NonObjectRoot(t *testing.T) {
	// Top-level array. Our Diff function expects map[string]any, so reject early.
	p := writeTempJSON(t, "arr.json", `[1,2,3]`)
	_, err := Load(p, 0)
	if err == nil {
		t.Fatal("expected error for non-object root")
	}
}

func TestSiteDiff_HasChanges(t *testing.T) {
	cases := []struct {
		name string
		d    *SiteDiff
		want bool
	}{
		{"nil", nil, false},
		{"empty", &SiteDiff{}, false},
		{"header", &SiteDiff{HeaderChanges: []FieldChange{{Path: "x", Kind: ChangeModified}}}, true},
		{"added", &SiteDiff{Added: []SiteRef{{SiteID: 1}}}, true},
		{"removed", &SiteDiff{Removed: []SiteRef{{SiteID: 1}}}, true},
		{"moved", &SiteDiff{Moved: []SiteMove{{SiteRef: SiteRef{SiteID: 1}}}}, true},
		{"modified-empty-changes", &SiteDiff{Modified: []SitePatch{{SiteRef: SiteRef{SiteID: 1}}}}, false},
		{"modified-with-changes", &SiteDiff{Modified: []SitePatch{{
			SiteRef: SiteRef{SiteID: 1},
			Changes: []FieldChange{{Path: "x", Kind: ChangeModified}},
		}}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.d.HasChanges(); got != tc.want {
				t.Errorf("HasChanges = %v, want %v", got, tc.want)
			}
		})
	}
}
