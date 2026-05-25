package sitediff

import "testing"

func TestIgnoreFilter_LeafKey_Default(t *testing.T) {
	f := NewIgnoreFilter(Options{})
	if !f.ShouldIgnore("Sites/0/SiteId", "SiteId") {
		t.Error("default SiteId should be ignored")
	}
	if !f.ShouldIgnore("HeaderInfo/PackageTime", "PackageTime") {
		t.Error("default PackageTime should be ignored")
	}
	if f.ShouldIgnore("Sites/0/Title", "Title") {
		t.Error("Title should not be ignored by default")
	}
}

func TestIgnoreFilter_LeafKey_Custom(t *testing.T) {
	f := NewIgnoreFilter(Options{IgnoreKeys: []string{"LabelText", "  Body  "}})
	if !f.ShouldIgnore("Sites/0/SiteSettings/Columns/0/LabelText", "LabelText") {
		t.Error("LabelText should be ignored")
	}
	if !f.ShouldIgnore("Sites/0/SiteSettings/Scripts/0/Body", "Body") {
		t.Error("Body (with whitespace in flag) should be ignored after trim")
	}
}

func TestIgnoreFilter_NoDefault(t *testing.T) {
	f := NewIgnoreFilter(Options{NoDefaultIgnore: true})
	if f.ShouldIgnore("Sites/0/SiteId", "SiteId") {
		t.Error("with NoDefaultIgnore, SiteId should not be ignored")
	}
}

func TestIgnoreFilter_PathGlob_Star(t *testing.T) {
	f := NewIgnoreFilter(Options{
		NoDefaultIgnore: true,
		IgnorePaths:     []string{"/Sites/*/SiteSettings/Version"},
	})
	if !f.ShouldIgnore("Sites/0/SiteSettings/Version", "Version") {
		t.Error("path glob with * should match index 0")
	}
	if !f.ShouldIgnore("Sites/12/SiteSettings/Version", "Version") {
		t.Error("path glob with * should match any index")
	}
	if f.ShouldIgnore("Sites/0/SiteSettings/Columns/0/Version", "Version") {
		t.Error("* matches exactly one segment, must not skip Columns/0")
	}
}

func TestIgnoreFilter_PathGlob_DoubleStar(t *testing.T) {
	f := NewIgnoreFilter(Options{
		NoDefaultIgnore: true,
		IgnorePaths:     []string{"/Sites/**/Version"},
	})
	if !f.ShouldIgnore("Sites/0/SiteSettings/Version", "Version") {
		t.Error("** should match nested Version")
	}
	if !f.ShouldIgnore("Sites/0/Version", "Version") {
		t.Error("** should match zero segments (immediate Version)")
	}
}

func TestIgnoreFilter_PathGlob_LeadingSlashOptional(t *testing.T) {
	f := NewIgnoreFilter(Options{
		NoDefaultIgnore: true,
		IgnorePaths:     []string{"Sites/0/Title"}, // no leading slash
	})
	if !f.ShouldIgnore("Sites/0/Title", "Title") {
		t.Error("path without leading slash should still match")
	}
}

func TestIgnoreFilter_EmptyEntriesIgnored(t *testing.T) {
	f := NewIgnoreFilter(Options{
		NoDefaultIgnore: true,
		IgnoreKeys:      []string{"", "  ", "X"},
		IgnorePaths:     []string{"", "/A"},
	})
	if !f.ShouldIgnore("any/path", "X") {
		t.Error("X should be ignored")
	}
	if !f.ShouldIgnore("A", "A") {
		t.Error("/A should match")
	}
}

func TestIgnoreFilter_NilSafe(t *testing.T) {
	var f *IgnoreFilter
	if f.ShouldIgnore("x", "x") {
		t.Error("nil filter should ignore nothing")
	}
}

func TestMatchSegments(t *testing.T) {
	cases := []struct {
		pat  []string
		in   []string
		want bool
	}{
		{[]string{"a"}, []string{"a"}, true},
		{[]string{"a"}, []string{"b"}, false},
		{[]string{"a", "*"}, []string{"a", "x"}, true},
		{[]string{"a", "*"}, []string{"a"}, false},
		{[]string{"a", "*", "c"}, []string{"a", "b", "c"}, true},
		{[]string{"a", "*", "c"}, []string{"a", "c"}, false},
		{[]string{"**"}, []string{}, true},
		{[]string{"**"}, []string{"a", "b", "c"}, true},
		{[]string{"a", "**"}, []string{"a"}, true},
		{[]string{"a", "**"}, []string{"a", "b", "c"}, true},
		{[]string{"a", "**", "c"}, []string{"a", "c"}, true},
		{[]string{"a", "**", "c"}, []string{"a", "b", "c"}, true},
		{[]string{"a", "**", "c"}, []string{"a", "b", "x", "y", "c"}, true},
		{[]string{"a", "**", "c"}, []string{"a", "b", "x", "y"}, false},
	}
	for _, tc := range cases {
		got := matchSegments(tc.pat, tc.in)
		if got != tc.want {
			t.Errorf("matchSegments(%v, %v) = %v, want %v", tc.pat, tc.in, got, tc.want)
		}
	}
}
