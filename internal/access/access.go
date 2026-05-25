package access

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

// Commander abstracts command execution for testing.
type Commander interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type realCommander struct{}

func (c *realCommander) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// AccessReader reads Access database files using mdbtools.
type AccessReader struct {
	cmd Commander
}

// NewAccessReader creates a new AccessReader using real command execution.
func NewAccessReader() *AccessReader {
	return &AccessReader{cmd: &realCommander{}}
}

// NewAccessReaderWithCommander creates a new AccessReader with a custom Commander (for testing).
func NewAccessReaderWithCommander(cmd Commander) *AccessReader {
	return &AccessReader{cmd: cmd}
}

// CheckMDBTools checks if mdbtools is installed.
func (a *AccessReader) CheckMDBTools(ctx context.Context) error {
	_, err := a.cmd.Run(ctx, "mdb-tables", "--version")
	if err != nil {
		return errs.New(errs.CodeInternalError,
			"mdbtools is not installed").
			WithSuggestion("Install mdbtools: sudo apt install mdbtools")
	}
	return nil
}

// ListTables returns table names from an Access database file.
func (a *AccessReader) ListTables(ctx context.Context, dbPath string) ([]string, error) {
	out, err := a.cmd.Run(ctx, "mdb-tables", "-1", dbPath)
	if err != nil {
		return nil, errs.New(errs.CodeInternalError,
			fmt.Sprintf("failed to list tables: %v", err)).
			WithSuggestion("Check that the file is a valid Access database (.mdb/.accdb)")
	}

	var tables []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			tables = append(tables, trimmed)
		}
	}
	return tables, nil
}

// ExportTable exports a table as CSV bytes.
func (a *AccessReader) ExportTable(ctx context.Context, dbPath, tableName string) ([]byte, error) {
	out, err := a.cmd.Run(ctx, "mdb-export", dbPath, tableName)
	if err != nil {
		return nil, errs.New(errs.CodeInternalError,
			fmt.Sprintf("failed to export table %q: %v", tableName, err)).
			WithSuggestion("Check the table name with 'plsnt access tables <file>'")
	}
	return out, nil
}
