package version

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCmd_Default(t *testing.T) {
	// Reset to defaults
	Version = "dev"
	Commit = "unknown"
	Date = "unknown"

	cmd := NewCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "plsnt dev (unknown, unknown)\n", buf.String())
}

func TestVersionCmd_WithLdflags(t *testing.T) {
	// Simulate ldflags injection
	Version = "0.1.0"
	Commit = "abc1234"
	Date = "2026-03-19T12:00:00Z"

	cmd := NewCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "plsnt v0.1.0 (abc1234, 2026-03-19T12:00:00Z)\n", buf.String())

	// Restore defaults
	Version = "dev"
	Commit = "unknown"
	Date = "unknown"
}

func TestFormatVersion_Dev(t *testing.T) {
	got := FormatVersion("dev", "unknown", "unknown")
	assert.Equal(t, "plsnt dev (unknown, unknown)", got)
}

func TestFormatVersion_Release(t *testing.T) {
	got := FormatVersion("0.1.0", "abc1234", "2026-03-19T12:00:00Z")
	assert.Equal(t, "plsnt v0.1.0 (abc1234, 2026-03-19T12:00:00Z)", got)
}

func TestFormatVersion_Prerelease(t *testing.T) {
	got := FormatVersion("1.0.0-rc.1", "def5678", "2026-03-19")
	assert.Equal(t, "plsnt v1.0.0-rc.1 (def5678, 2026-03-19)", got)
}
