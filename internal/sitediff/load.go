package sitediff

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

// DefaultMaxBytes is the default per-file input size limit (50 MiB).
const DefaultMaxBytes int64 = 50 << 20

// Load reads a SitePackage JSON file from disk into a generic map.
//
// Numeric values are decoded as json.Number so that the diff preserves
// their textual form (e.g. "1.017" never becomes 1.0169999...). The reader
// is wrapped in io.LimitReader; oversize input returns CodeValidationError.
//
// maxBytes <= 0 falls back to DefaultMaxBytes.
func Load(path string, maxBytes int64) (map[string]any, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errs.New(errs.CodeValidationError,
				fmt.Sprintf("file not found: %s", path)).
				WithSuggestion("Verify the path. Web UI exports are typically saved under Downloads/.")
		}
		return nil, errs.Wrap(err, errs.CodeValidationError)
	}
	defer f.Close()

	// LimitReader gives us +1 byte to detect overflow without slurping more.
	limited := io.LimitReader(f, maxBytes+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeInternalError)
	}
	if int64(len(buf)) > maxBytes {
		return nil, errs.New(errs.CodeValidationError,
			fmt.Sprintf("file %s exceeds size limit (%d bytes)", path, maxBytes)).
			WithSuggestion("Pass --max-size to raise the limit, e.g. --max-size 100MB")
	}

	// Pleasanter Web UI exports include a UTF-8 BOM (EF BB BF). encoding/json
	// rejects it as "invalid character 'ï'", so strip it transparently.
	buf = bytes.TrimPrefix(buf, []byte{0xEF, 0xBB, 0xBF})

	dec := json.NewDecoder(bytes.NewReader(buf))
	dec.UseNumber()

	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		return nil, errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("invalid JSON in %s: %v", path, err)).
			WithSuggestion("Confirm the file is a valid Pleasanter SitePackage JSON export.")
	}
	if out == nil {
		return nil, errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("file %s parses as JSON null, expected an object", path))
	}
	return out, nil
}
