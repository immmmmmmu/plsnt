package encoding

import (
	"bytes"
	"io"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// DetectAndConvert reads data, detects if it's Shift-JIS, and converts to UTF-8 if needed.
// Detection heuristic: check for BOM, then validate as UTF-8, then try Shift-JIS decode.
// Returns the converted data, the detected encoding name, and any error.
func DetectAndConvert(data []byte) ([]byte, string, error) {
	// 1. Check UTF-8 BOM
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		return data[3:], "UTF-8 (BOM)", nil
	}

	// 2. Check if valid UTF-8
	if isValidUTF8(data) {
		return data, "UTF-8", nil
	}

	// 3. Try Shift-JIS decode
	decoded, err := decodeShiftJIS(data)
	if err == nil {
		return decoded, "Shift-JIS", nil
	}

	// 4. Fallback: return as-is
	return data, "unknown", nil
}

// isValidUTF8 checks whether all bytes form valid UTF-8 sequences.
func isValidUTF8(data []byte) bool {
	return utf8.Valid(data)
}

// decodeShiftJIS converts Shift-JIS encoded bytes to UTF-8.
func decodeShiftJIS(data []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(data), japanese.ShiftJIS.NewDecoder())
	return io.ReadAll(reader)
}

// NewReader wraps an io.Reader with auto-detection.
// It reads all content, detects the encoding, and returns a new reader with UTF-8 content
// along with the detected encoding name.
func NewReader(r io.Reader) (io.Reader, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}
	converted, enc, err := DetectAndConvert(data)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(converted), enc, nil
}
