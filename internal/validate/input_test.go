package validate

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

func TestSiteID_Valid(t *testing.T) {
	valid := []string{"1", "100", "1234", "999999"}
	for _, v := range valid {
		if err := SiteID(v); err != nil {
			t.Errorf("SiteID(%q) returned error: %v", v, err)
		}
	}
}

func TestSiteID_Invalid(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"0", "zero"},
		{"-1", "negative"},
		{"abc", "non-numeric"},
		{"", "empty"},
		{"12.5", "decimal"},
	}
	for _, tt := range tests {
		if err := SiteID(tt.input); err == nil {
			t.Errorf("SiteID(%q) [%s] should return error", tt.input, tt.desc)
		}
	}
}

func TestSiteID_QueryParamMixup(t *testing.T) {
	tests := []string{"1234?fields=Title", "1234#anchor", "1234&other=1"}
	for _, v := range tests {
		if err := SiteID(v); err == nil {
			t.Errorf("SiteID(%q) should reject query param chars", v)
		}
	}
}

func TestSiteID_DoubleEncoding(t *testing.T) {
	if err := SiteID("1234%2525"); err == nil {
		t.Error("SiteID with %% should be rejected")
	}
}

func TestParentID_Valid(t *testing.T) {
	valid := []string{"1", "100", "32814", "999999"}
	for _, v := range valid {
		if err := ParentID(v, "parent-id"); err != nil {
			t.Errorf("ParentID(%q) returned error: %v", v, err)
		}
	}
}

func TestParentID_RejectsZeroWithTenantRootHint(t *testing.T) {
	err := ParentID("0", "parent-id")
	if err == nil {
		t.Fatal("ParentID(\"0\") should return error")
	}

	var cliErr *errs.CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *errs.CLIError, got %T", err)
	}

	if cliErr.ErrorBody.Code != errs.CodeValidationError {
		t.Errorf("expected CodeValidationError, got %v", cliErr.ErrorBody.Code)
	}

	if !strings.Contains(cliErr.ErrorBody.Message, "parent-id") {
		t.Errorf("error message should mention flag name, got: %s", cliErr.ErrorBody.Message)
	}

	hint := cliErr.ErrorBody.Suggestion
	wantSubstrings := []string{
		"tenant root",
		"parent_id=0",
		"Web UI",
		"https://pleasanter.org/en/manual/api-site-create",
		"--parent-id",
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(hint, want) {
			t.Errorf("suggestion missing %q\nhint: %s", want, hint)
		}
	}
}

func TestParentID_FlagNameEmbedded(t *testing.T) {
	err := ParentID("0", "folder-id")
	if err == nil {
		t.Fatal("expected error")
	}
	var cliErr *errs.CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *errs.CLIError, got %T", err)
	}
	if !strings.Contains(cliErr.ErrorBody.Message, "folder-id") {
		t.Errorf("expected flag name in message, got: %s", cliErr.ErrorBody.Message)
	}
	if !strings.Contains(cliErr.ErrorBody.Suggestion, "--folder-id") {
		t.Errorf("expected --folder-id in suggestion, got: %s", cliErr.ErrorBody.Suggestion)
	}
}

func TestParentID_Invalid(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"0", "zero"},
		{"-1", "negative"},
		{"abc", "non-numeric"},
		{"", "empty"},
		{"12.5", "decimal"},
	}
	for _, tt := range tests {
		if err := ParentID(tt.input, "parent-id"); err == nil {
			t.Errorf("ParentID(%q) [%s] should return error", tt.input, tt.desc)
		}
	}
}

func TestParentID_QueryParamMixup(t *testing.T) {
	for _, v := range []string{"1234?fields=Title", "1234#anchor", "1234&other=1"} {
		if err := ParentID(v, "parent-id"); err == nil {
			t.Errorf("ParentID(%q) should reject query param chars", v)
		}
	}
}

func TestParentIDInt_Valid(t *testing.T) {
	for _, v := range []int64{1, 100, 32814, 999999} {
		if err := ParentIDInt(v, "folder-id"); err != nil {
			t.Errorf("ParentIDInt(%d) returned error: %v", v, err)
		}
	}
}

func TestParentIDInt_RejectsZeroWithTenantRootHint(t *testing.T) {
	err := ParentIDInt(0, "folder-id")
	if err == nil {
		t.Fatal("ParentIDInt(0) should return error")
	}
	var cliErr *errs.CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *errs.CLIError, got %T", err)
	}
	if !strings.Contains(cliErr.ErrorBody.Suggestion, "tenant root") {
		t.Errorf("expected tenant root hint, got: %s", cliErr.ErrorBody.Suggestion)
	}
}

func TestParentIDInt_RejectsNegative(t *testing.T) {
	if err := ParentIDInt(-1, "folder-id"); err == nil {
		t.Error("ParentIDInt(-1) should return error")
	}
}

func TestRecordID_Valid(t *testing.T) {
	if err := RecordID("5678"); err != nil {
		t.Errorf("RecordID(5678) returned error: %v", err)
	}
}

func TestRecordID_Invalid(t *testing.T) {
	if err := RecordID("0"); err == nil {
		t.Error("RecordID(0) should return error")
	}
	if err := RecordID("abc"); err == nil {
		t.Error("RecordID(abc) should return error")
	}
}

func TestGroupID_Valid(t *testing.T) {
	valid := []string{"1", "100", "999"}
	for _, v := range valid {
		if err := GroupID(v); err != nil {
			t.Errorf("GroupID(%q) returned error: %v", v, err)
		}
	}
}

func TestGroupID_Invalid(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"0", "zero"},
		{"-1", "negative"},
		{"abc", "non-numeric"},
		{"", "empty"},
	}
	for _, tt := range tests {
		if err := GroupID(tt.input); err == nil {
			t.Errorf("GroupID(%q) [%s] should return error", tt.input, tt.desc)
		}
	}
}

func TestGroupID_QueryParamMixup(t *testing.T) {
	if err := GroupID("123?foo=bar"); err == nil {
		t.Error("GroupID with query param chars should be rejected")
	}
}

func TestGroupID_DoubleEncoding(t *testing.T) {
	if err := GroupID("123%25"); err == nil {
		t.Error("GroupID with %% should be rejected")
	}
}

func TestDeptID_Valid(t *testing.T) {
	valid := []string{"1", "10", "500"}
	for _, v := range valid {
		if err := DeptID(v); err != nil {
			t.Errorf("DeptID(%q) returned error: %v", v, err)
		}
	}
}

func TestDeptID_Invalid(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"0", "zero"},
		{"-5", "negative"},
		{"xyz", "non-numeric"},
		{"", "empty"},
	}
	for _, tt := range tests {
		if err := DeptID(tt.input); err == nil {
			t.Errorf("DeptID(%q) [%s] should return error", tt.input, tt.desc)
		}
	}
}

func TestDeptID_QueryParamMixup(t *testing.T) {
	if err := DeptID("10#anchor"); err == nil {
		t.Error("DeptID with query param chars should be rejected")
	}
}

func TestDeptID_DoubleEncoding(t *testing.T) {
	if err := DeptID("10%2F"); err == nil {
		t.Error("DeptID with %% should be rejected")
	}
}

func TestRecordID_QueryParamMixup(t *testing.T) {
	if err := RecordID("5678?x=1"); err == nil {
		t.Error("RecordID with query param chars should be rejected")
	}
}

func TestRecordID_DoubleEncoding(t *testing.T) {
	if err := RecordID("5678%25"); err == nil {
		t.Error("RecordID with %% should be rejected")
	}
}

func TestNoControlChars_Valid(t *testing.T) {
	valid := []string{
		"hello world",
		"line1\nline2",
		"tab\there",
		"carriage\rreturn",
		"日本語テスト",
	}
	for _, v := range valid {
		if err := NoControlChars(v, "test"); err != nil {
			t.Errorf("NoControlChars(%q) returned error: %v", v, err)
		}
	}
}

func TestNoControlChars_Invalid(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"\x00hello", "null byte"},
		{"hello\x01", "SOH"},
		{"mid\x07dle", "bell"},
		{"\x7Ftest", "DEL"},
	}
	for _, tt := range tests {
		if err := NoControlChars(tt.input, "test"); err == nil {
			t.Errorf("NoControlChars(%q) [%s] should return error", tt.desc, tt.desc)
		}
	}
}

func TestFilePath_Valid(t *testing.T) {
	cwd, _ := os.Getwd()
	if err := FilePath("data/output.csv", cwd); err != nil {
		t.Errorf("FilePath for relative path returned error: %v", err)
	}
}

func TestFilePath_PathTraversal(t *testing.T) {
	cwd, _ := os.Getwd()
	tests := []string{
		"../../etc/passwd",
		"../../../.ssh/id_rsa",
		"data/../../secrets",
	}
	for _, v := range tests {
		if err := FilePath(v, cwd); err == nil {
			t.Errorf("FilePath(%q) should reject path traversal", v)
		}
	}
}

func TestFilePath_OutsideCwd(t *testing.T) {
	// A path that resolves outside cwd
	if err := FilePath("/etc/passwd", "/home/user"); err == nil {
		t.Error("FilePath pointing outside cwd should return error")
	}
}

func TestFilePath_AbsoluteInsideCwd(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/subdir/file.csv"
	if err := FilePath(path, dir); err != nil {
		t.Errorf("FilePath for absolute path inside cwd returned error: %v", err)
	}
}

func TestJSONSyntax_WhitespaceOnly(t *testing.T) {
	if err := JSONSyntax("   "); err == nil {
		t.Error("JSONSyntax with whitespace-only input should return error")
	}
}

func TestJSONSyntax_Valid(t *testing.T) {
	valid := []string{
		`{"Title": "test"}`,
		`[{"a": 1}]`,
		`  { "key": "value" }  `,
		"",
	}
	for _, v := range valid {
		if err := JSONSyntax(v); err != nil {
			t.Errorf("JSONSyntax(%q) returned error: %v", v, err)
		}
	}
}

func TestJSONSyntax_Invalid(t *testing.T) {
	tests := []string{
		"not json",
		"123",
		"true",
		`"string"`,
	}
	for _, v := range tests {
		if err := JSONSyntax(v); err == nil {
			t.Errorf("JSONSyntax(%q) should return error", v)
		}
	}
}
