package errs

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	e := New(CodeValidationError, "invalid site ID")

	if e.Success {
		t.Error("expected success to be false")
	}
	if e.ErrorBody.Code != CodeValidationError {
		t.Errorf("expected code %s, got %s", CodeValidationError, e.ErrorBody.Code)
	}
	if e.ErrorBody.Message != "invalid site ID" {
		t.Errorf("expected message 'invalid site ID', got '%s'", e.ErrorBody.Message)
	}
}

func TestError(t *testing.T) {
	e := New(CodeRecordNotFound, "record 123 not found")
	if e.Error() != "record 123 not found" {
		t.Errorf("expected 'record 123 not found', got '%s'", e.Error())
	}
}

func TestError_AppendsSuggestionAsHint(t *testing.T) {
	// Issue #12: Error() must surface the suggestion as a Hint line so cobra
	// renders it on stderr alongside the primary error message.
	e := New(CodeValidationError, "parent-id must be a positive integer, got: \"0\"").
		WithSuggestion("Pleasanter API does not support tenant root creation; create a parent folder first.")

	got := e.Error()
	want := "parent-id must be a positive integer, got: \"0\"\nHint: Pleasanter API does not support tenant root creation; create a parent folder first."
	if got != want {
		t.Errorf("Error() mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestMessage_ReturnsBareMessage(t *testing.T) {
	e := New(CodeValidationError, "bare message").WithSuggestion("hint text")
	if e.Message() != "bare message" {
		t.Errorf("Message() = %q, want %q", e.Message(), "bare message")
	}
}

func TestNewWithSuggestion(t *testing.T) {
	e := NewWithSuggestion(
		CodeRecordNotFound,
		"record not found",
		"use plsnt record list to check available records",
	)

	if e.ErrorBody.Suggestion != "use plsnt record list to check available records" {
		t.Errorf("unexpected suggestion: %s", e.ErrorBody.Suggestion)
	}
}

func TestNewWithStatus(t *testing.T) {
	e := NewWithStatus(CodeAuthError, "authentication failed", 401)

	if e.ErrorBody.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", e.ErrorBody.StatusCode)
	}
}

func TestWithSuggestion(t *testing.T) {
	e := New(CodeRecordNotFound, "not found")
	e2 := e.WithSuggestion("try another ID")

	if e2.ErrorBody.Suggestion != "try another ID" {
		t.Errorf("unexpected suggestion: %s", e2.ErrorBody.Suggestion)
	}
	// original should be unchanged (immutability)
	if e.ErrorBody.Suggestion != "" {
		t.Error("original error was mutated")
	}
}

func TestWithReceived(t *testing.T) {
	e := New(CodeInvalidInput, "invalid character in ID")
	e2 := e.WithReceived("5678?fields=Title")

	if e2.ErrorBody.Received != "5678?fields=Title" {
		t.Errorf("unexpected received: %s", e2.ErrorBody.Received)
	}
	if e.ErrorBody.Received != "" {
		t.Error("original error was mutated")
	}
}

func TestWriteJSON(t *testing.T) {
	e := NewWithSuggestion(
		CodeRecordNotFound,
		"record 9999 not found in site 1234",
		"plsnt record list --site-id 1234 --fields IssueId,Title",
	)
	e.ErrorBody.StatusCode = 404

	var buf bytes.Buffer
	if err := WriteJSON(&buf, e); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	var parsed CLIError
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}

	if parsed.Success {
		t.Error("expected success false")
	}
	if parsed.ErrorBody.Code != CodeRecordNotFound {
		t.Errorf("expected code %s, got %s", CodeRecordNotFound, parsed.ErrorBody.Code)
	}
	if parsed.ErrorBody.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", parsed.ErrorBody.StatusCode)
	}
	if parsed.ErrorBody.Suggestion == "" {
		t.Error("expected suggestion to be present")
	}
}

func TestWriteJSON_OmitsEmptyFields(t *testing.T) {
	e := New(CodeInternalError, "something broke")

	var buf bytes.Buffer
	if err := WriteJSON(&buf, e); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	errBody, ok := raw["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error field")
	}
	if _, exists := errBody["status_code"]; exists {
		t.Error("status_code should be omitted when zero")
	}
	if _, exists := errBody["suggestion"]; exists {
		t.Error("suggestion should be omitted when empty")
	}
	if _, exists := errBody["received"]; exists {
		t.Error("received should be omitted when empty")
	}
}

func TestExitCode_AllCodes(t *testing.T) {
	tests := []struct {
		code     Code
		expected int
	}{
		{CodeValidationError, 2},
		{CodeInvalidInput, 2},
		{CodeRecordNotFound, 3},
		{CodeSiteNotFound, 3},
		{CodeAuthError, 4},
		{CodePermissionError, 4},
		{CodeConnectionError, 5},
		{CodeTimeoutError, 5},
		{CodeServerError, 6},
		{CodeInternalError, 7},
		{Code("UNKNOWN"), 1},
	}

	for _, tt := range tests {
		got := ExitCode(tt.code)
		if got != tt.expected {
			t.Errorf("ExitCode(%s) = %d, want %d", tt.code, got, tt.expected)
		}
	}
}

func TestWrap(t *testing.T) {
	err := errors.New("connection refused")
	e := Wrap(err, CodeConnectionError)

	if e.ErrorBody.Code != CodeConnectionError {
		t.Errorf("expected code %s, got %s", CodeConnectionError, e.ErrorBody.Code)
	}
	if e.ErrorBody.Message != "connection refused" {
		t.Errorf("expected message 'connection refused', got '%s'", e.ErrorBody.Message)
	}
}
