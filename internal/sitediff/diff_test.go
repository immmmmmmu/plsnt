package sitediff

import (
	"encoding/json"
	"strings"
	"testing"
)

func mustParse(t *testing.T, s string) map[string]any {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return out
}

func TestDiff_NoChanges(t *testing.T) {
	a := `{"HeaderInfo":{},"Sites":[{"SiteId":1,"Title":"A","SiteSettings":{"Version":1.0}}]}`
	d, err := Diff(mustParse(t, a), mustParse(t, a), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if d.HasChanges() {
		t.Errorf("identical packages should have no changes, got %+v", d)
	}
}

func TestDiff_AddedAndRemovedSites(t *testing.T) {
	old := `{"Sites":[{"SiteId":1,"Title":"keep"},{"SiteId":2,"Title":"drop"}]}`
	newer := `{"Sites":[{"SiteId":1,"Title":"keep"},{"SiteId":3,"Title":"new"}]}`
	d, err := Diff(mustParse(t, old), mustParse(t, newer), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Removed) != 1 || d.Removed[0].SiteID != 2 {
		t.Errorf("removed: %+v", d.Removed)
	}
	if len(d.Added) != 1 || d.Added[0].SiteID != 3 {
		t.Errorf("added: %+v", d.Added)
	}
}

func TestDiff_SiteMoved(t *testing.T) {
	old := `{"Sites":[{"SiteId":1,"Title":"sub","ParentId":10}]}`
	newer := `{"Sites":[{"SiteId":1,"Title":"sub","ParentId":20}]}`
	d, err := Diff(mustParse(t, old), mustParse(t, newer), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Moved) != 1 {
		t.Fatalf("moved expected, got %+v", d)
	}
	if d.Moved[0].OldParentID != 10 || d.Moved[0].NewParentID != 20 {
		t.Errorf("move parents: %+v", d.Moved[0])
	}
}

func TestDiff_ColumnsModifiedByColumnName(t *testing.T) {
	old := `{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Columns":[
		{"ColumnName":"ClassA","LabelText":"old"},
		{"ColumnName":"ClassB","LabelText":"keep"}
	]}}]}`
	newer := `{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Columns":[
		{"ColumnName":"ClassB","LabelText":"keep"},
		{"ColumnName":"ClassA","LabelText":"new"}
	]}}]}`
	d, err := Diff(mustParse(t, old), mustParse(t, newer), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Modified) != 1 {
		t.Fatalf("expected 1 modified site, got %d", len(d.Modified))
	}
	changes := d.Modified[0].Changes
	if len(changes) != 1 {
		t.Fatalf("expected 1 change (LabelText only), got %d: %+v", len(changes), changes)
	}
	if !strings.Contains(changes[0].Path, "ColumnName=ClassA") {
		t.Errorf("path should reference ClassA: %s", changes[0].Path)
	}
	if changes[0].Kind != ChangeModified {
		t.Errorf("kind: %v", changes[0].Kind)
	}
}

func TestDiff_ScriptBodyUnifiedDiff(t *testing.T) {
	old := `{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Scripts":[
		{"Title":"hdr","Body":"line1\nline2\nline3\n"}
	]}}]}`
	newer := `{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Scripts":[
		{"Title":"hdr","Body":"line1\nLINE2\nline3\n"}
	]}}]}`
	d, err := Diff(mustParse(t, old), mustParse(t, newer), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Modified) != 1 || len(d.Modified[0].Changes) != 1 {
		t.Fatalf("expected 1 change, got %+v", d.Modified)
	}
	c := d.Modified[0].Changes[0]
	if c.Kind != ChangeTextDiff {
		t.Errorf("expected TextDiff, got %v", c.Kind)
	}
	if !strings.Contains(c.UnifiedDiff, "-line2") || !strings.Contains(c.UnifiedDiff, "+LINE2") {
		t.Errorf("unified diff missing markers:\n%s", c.UnifiedDiff)
	}
}

func TestDiff_DefaultIgnoreSiteId(t *testing.T) {
	// The same site has different SiteIds (env crossing), but Title matches
	// via auto matcher. SiteId itself is in DefaultIgnoreKeys so the leaf
	// should NOT show up as a modification.
	old := `{"Sites":[{"SiteId":11,"Title":"Same","ParentId":1,"SiteSettings":{}}]}`
	newer := `{"Sites":[{"SiteId":99,"Title":"Same","ParentId":1,"SiteSettings":{}}]}`
	d, err := Diff(mustParse(t, old), mustParse(t, newer), Options{MatchSitesBy: "title"})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Added) > 0 || len(d.Removed) > 0 {
		t.Fatalf("expected the sites to match by Title, added=%v removed=%v", d.Added, d.Removed)
	}
	if len(d.Modified) != 0 {
		t.Fatalf("SiteId churn should be ignored by default, got %+v", d.Modified)
	}
}

func TestDiff_CustomIgnoreKey(t *testing.T) {
	old := `{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Columns":[{"ColumnName":"A","LabelText":"old"}]}}]}`
	newer := `{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Columns":[{"ColumnName":"A","LabelText":"new"}]}}]}`
	d, err := Diff(mustParse(t, old), mustParse(t, newer), Options{IgnoreKeys: []string{"LabelText"}})
	if err != nil {
		t.Fatal(err)
	}
	if d.HasChanges() {
		t.Errorf("LabelText should be ignored, got %+v", d)
	}
}

func TestDiff_StableOrdering(t *testing.T) {
	// Run the same diff multiple times to confirm output is deterministic
	// (Go map iteration is randomized; our walker must sort).
	pkg := `{"Sites":[
		{"SiteId":3,"Title":"C"},
		{"SiteId":1,"Title":"A"},
		{"SiteId":2,"Title":"B"}
	]}`
	other := `{"Sites":[
		{"SiteId":1,"Title":"A2"},
		{"SiteId":2,"Title":"B2"},
		{"SiteId":3,"Title":"C2"}
	]}`
	var first string
	for i := 0; i < 10; i++ {
		d, err := Diff(mustParse(t, pkg), mustParse(t, other), Options{})
		if err != nil {
			t.Fatal(err)
		}
		buf, _ := json.Marshal(d)
		s := string(buf)
		if i == 0 {
			first = s
			continue
		}
		if s != first {
			t.Fatalf("output changed between runs:\nfirst: %s\nnow:   %s", first, s)
		}
	}
}

func TestDiff_NumberPrecisionPreserved(t *testing.T) {
	// A float-y precision value that round-trips badly through float64.
	old := `{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Version":1.017}}]}`
	newer := `{"Sites":[{"SiteId":1,"Title":"S","SiteSettings":{"Version":1.017}}]}`
	d, err := Diff(mustParse(t, old), mustParse(t, newer), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if d.HasChanges() {
		t.Errorf("identical Version should not diff: %+v", d)
	}
}

func TestDiff_HeaderIncludeFlagSurfaces(t *testing.T) {
	old := `{"HeaderInfo":{"IncludeNotifications":false},"Sites":[]}`
	newer := `{"HeaderInfo":{"IncludeNotifications":true},"Sites":[]}`
	d, err := Diff(mustParse(t, old), mustParse(t, newer), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.HeaderChanges) != 1 || !strings.Contains(d.HeaderChanges[0].Path, "IncludeNotifications") {
		t.Errorf("header changes wrong: %+v", d.HeaderChanges)
	}
}

func TestDiff_HeaderMetadataIgnored(t *testing.T) {
	old := `{"HeaderInfo":{"PackageTime":"2026-01-01","Server":"localhost","AssemblyVersion":"1.5.2.0"},"Sites":[]}`
	newer := `{"HeaderInfo":{"PackageTime":"2026-04-30","Server":"prod.example.com","AssemblyVersion":"1.5.3.0"},"Sites":[]}`
	d, err := Diff(mustParse(t, old), mustParse(t, newer), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if d.HasChanges() {
		t.Errorf("metadata-only diff should be ignored, got %+v", d)
	}
}
