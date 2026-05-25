package sitediff

import (
	"fmt"
	"io"
	"strings"
)

// Formatter renders a SiteDiff as bytes on the given writer. Implementations
// must be deterministic — Diff already sorts its output, but renderers must
// avoid map iteration in Go.
type Formatter interface {
	Format(w io.Writer, d *SiteDiff) error
}

// NewFormatter returns the formatter registered under name. Recognised
// names: "text" (default), "json", "markdown".
func NewFormatter(name string) (Formatter, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "text":
		return &TextFormatter{}, nil
	case "json":
		return &JSONFormatter{}, nil
	case "markdown", "md":
		return &MarkdownFormatter{}, nil
	default:
		return nil, fmt.Errorf("unknown format: %q (want one of: text, json, markdown)", name)
	}
}
