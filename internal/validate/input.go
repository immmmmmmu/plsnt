package validate

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

func SiteID(s string) error {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return errs.New(errs.CodeValidationError,
			fmt.Sprintf("SiteID must be a positive integer, got: %q", s)).
			WithSuggestion("Specify a valid site ID, e.g. --site-id 1234")
	}
	return checkIDChars(s, "SiteID")
}

// ParentID validates a parent SiteID used by site/folder creation commands.
// Unlike SiteID, this rejects "0" with an explicit hint about the tenant
// root limitation: the Pleasanter API does not support creating sites at
// parent_id=0, so users must create a parent folder via the Web UI first.
//
// flagName is the user-facing flag (e.g. "parent-id", "folder-id") and is
// embedded in both the error message and the suggestion.
func ParentID(s, flagName string) error {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return errs.New(errs.CodeValidationError,
			fmt.Sprintf("%s must be a positive integer, got: %q", flagName, s)).
			WithSuggestion(tenantRootHint(flagName))
	}
	return checkIDChars(s, flagName)
}

// ParentIDInt is the int64 variant of ParentID, used by commands that
// receive the value via cobra Int64Var.
func ParentIDInt(id int64, flagName string) error {
	if id <= 0 {
		return errs.New(errs.CodeValidationError,
			fmt.Sprintf("%s must be a positive integer, got: %d", flagName, id)).
			WithSuggestion(tenantRootHint(flagName))
	}
	return nil
}

func tenantRootHint(flagName string) string {
	return fmt.Sprintf(
		"Pleasanter API does not support creating sites at the tenant root (parent_id=0). "+
			"Create a parent folder via Web UI first, then specify its SiteId "+
			"(e.g. --%s 1234, or --var folder_id=1234 for `plsnt batch run`). "+
			"See: https://pleasanter.org/en/manual/api-site-create",
		flagName,
	)
}

func RecordID(s string) error {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return errs.New(errs.CodeValidationError,
			fmt.Sprintf("RecordID must be a positive integer, got: %q", s)).
			WithSuggestion("Specify a valid record ID, e.g. --id 5678")
	}
	return checkIDChars(s, "RecordID")
}

func GroupID(s string) error {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return errs.New(errs.CodeValidationError,
			fmt.Sprintf("GroupID must be a positive integer, got: %q", s)).
			WithSuggestion("Specify a valid group ID, e.g. 123")
	}
	return checkIDChars(s, "GroupID")
}

func DeptID(s string) error {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return errs.New(errs.CodeValidationError,
			fmt.Sprintf("DeptID must be a positive integer, got: %q", s)).
			WithSuggestion("Specify a valid department ID, e.g. 10")
	}
	return checkIDChars(s, "DeptID")
}

func checkIDChars(s, name string) error {
	if strings.ContainsAny(s, "?#&") {
		return errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("%s contains invalid character. Query parameters may have been mixed in.", name)).
			WithReceived(s).
			WithSuggestion(fmt.Sprintf("Specify only the numeric %s without query parameters", name))
	}
	if strings.Contains(s, "%") {
		return errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("%s contains '%%' character. Double-encoding may have occurred.", name)).
			WithReceived(s).
			WithSuggestion(fmt.Sprintf("Specify the %s without URL encoding", name))
	}
	return nil
}

func NoControlChars(s, fieldName string) error {
	for i, r := range s {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			return errs.New(errs.CodeInvalidInput,
				fmt.Sprintf("%s contains control character at position %d (U+%04X)", fieldName, i, r)).
				WithSuggestion(fmt.Sprintf("Remove control characters from %s", fieldName))
		}
		if r == 0x7F {
			return errs.New(errs.CodeInvalidInput,
				fmt.Sprintf("%s contains DEL character at position %d", fieldName, i)).
				WithSuggestion(fmt.Sprintf("Remove control characters from %s", fieldName))
		}
	}
	return nil
}

func FilePath(path, cwd string) error {
	if strings.Contains(path, "..") {
		return errs.New(errs.CodeValidationError,
			"Path contains '..' (path traversal detected)").
			WithReceived(path).
			WithSuggestion("Use an absolute path or a path relative to the current directory without '..'")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return errs.Wrap(err, errs.CodeValidationError)
	}

	cwdAbs, err := filepath.Abs(cwd)
	if err != nil {
		return errs.Wrap(err, errs.CodeInternalError)
	}

	if !strings.HasPrefix(abs, cwdAbs) {
		return errs.New(errs.CodeValidationError,
			"Path points outside the current working directory").
			WithReceived(path).
			WithSuggestion("Specify a path within the current directory")
	}

	return nil
}

func JSONSyntax(raw string) error {
	if raw == "" {
		return nil
	}
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) == 0 {
		return errs.New(errs.CodeInvalidInput, "JSON payload is empty")
	}
	if trimmed[0] != '{' && trimmed[0] != '[' {
		return errs.New(errs.CodeInvalidInput,
			"JSON payload must start with '{' or '['").
			WithSuggestion("Provide a valid JSON object, e.g. --json '{\"Title\": \"test\"}'")
	}
	return nil
}
