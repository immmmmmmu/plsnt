package sitediff

import (
	"encoding/json"
	"io"
)

// JSONFormatter renders a SiteDiff as pretty-printed JSON. Field order
// matches the struct declaration (json tags drive Marshal); since the diff
// itself was produced deterministically, the output is stable across runs.
type JSONFormatter struct{}

func (JSONFormatter) Format(w io.Writer, d *SiteDiff) error {
	if d == nil {
		d = &SiteDiff{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(d)
}
