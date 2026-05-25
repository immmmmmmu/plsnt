package errs

import (
	"encoding/json"
	"fmt"
	"io"
)

type Code string

const (
	CodeValidationError Code = "VALIDATION_ERROR"
	CodeInvalidInput    Code = "INVALID_INPUT"
	CodeRecordNotFound  Code = "RECORD_NOT_FOUND"
	CodeSiteNotFound    Code = "SITE_NOT_FOUND"
	CodeAuthError       Code = "AUTH_ERROR"
	CodePermissionError Code = "PERMISSION_ERROR"
	CodeConnectionError Code = "CONNECTION_ERROR"
	CodeTimeoutError    Code = "TIMEOUT_ERROR"
	CodeServerError     Code = "SERVER_ERROR"
	CodeInternalError   Code = "INTERNAL_ERROR"
)

type CLIError struct {
	Success    bool   `json:"success"`
	ErrorBody  *Body  `json:"error"`
}

type Body struct {
	Code       Code   `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Received   string `json:"received,omitempty"`
}

func (e *CLIError) Error() string {
	if e.ErrorBody == nil {
		return "unknown error"
	}
	if e.ErrorBody.Suggestion == "" {
		return e.ErrorBody.Message
	}
	return e.ErrorBody.Message + "\nHint: " + e.ErrorBody.Suggestion
}

// Message returns the bare error message without the suggestion hint.
// Use this when comparing or logging only the primary error reason.
func (e *CLIError) Message() string {
	if e.ErrorBody == nil {
		return ""
	}
	return e.ErrorBody.Message
}

func New(code Code, message string) *CLIError {
	return &CLIError{
		Success: false,
		ErrorBody: &Body{
			Code:    code,
			Message: message,
		},
	}
}

func NewWithSuggestion(code Code, message, suggestion string) *CLIError {
	e := New(code, message)
	e.ErrorBody.Suggestion = suggestion
	return e
}

func NewWithStatus(code Code, message string, statusCode int) *CLIError {
	e := New(code, message)
	e.ErrorBody.StatusCode = statusCode
	return e
}

func (e *CLIError) WithSuggestion(suggestion string) *CLIError {
	return &CLIError{
		Success: false,
		ErrorBody: &Body{
			Code:       e.ErrorBody.Code,
			Message:    e.ErrorBody.Message,
			StatusCode: e.ErrorBody.StatusCode,
			Suggestion: suggestion,
			Received:   e.ErrorBody.Received,
		},
	}
}

func (e *CLIError) WithReceived(received string) *CLIError {
	return &CLIError{
		Success: false,
		ErrorBody: &Body{
			Code:       e.ErrorBody.Code,
			Message:    e.ErrorBody.Message,
			StatusCode: e.ErrorBody.StatusCode,
			Suggestion: e.ErrorBody.Suggestion,
			Received:   received,
		},
	}
}

func WriteJSON(w io.Writer, e *CLIError) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(e)
}

func ExitCode(code Code) int {
	switch code {
	case CodeValidationError, CodeInvalidInput:
		return 2
	case CodeRecordNotFound, CodeSiteNotFound:
		return 3
	case CodeAuthError, CodePermissionError:
		return 4
	case CodeConnectionError, CodeTimeoutError:
		return 5
	case CodeServerError:
		return 6
	case CodeInternalError:
		return 7
	default:
		return 1
	}
}

func Wrap(err error, code Code) *CLIError {
	return New(code, fmt.Sprintf("%v", err))
}
