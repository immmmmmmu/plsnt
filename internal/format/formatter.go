package format

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

// Formatter writes structured data to an io.Writer in a specific format.
type Formatter interface {
	Format(w io.Writer, data any) error
}

// New returns a Formatter for the given format name.
// Supported: "json", "ndjson", "table", "csv", "count", "ids".
func New(format string, fields []string) (Formatter, error) {
	switch strings.ToLower(format) {
	case "json", "":
		return &jsonFormatter{}, nil
	case "ndjson":
		return &ndjsonFormatter{}, nil
	case "table":
		return &tableFormatter{fields: fields}, nil
	case "csv":
		return &csvFormatter{fields: fields}, nil
	case "count":
		return &countFormatter{fields: fields}, nil
	case "ids":
		return &idsFormatter{fields: fields}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %q", format)
	}
}

type jsonFormatter struct{}

func (f *jsonFormatter) Format(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

type ndjsonFormatter struct{}

func (f *ndjsonFormatter) Format(w io.Writer, data any) error {
	rows, err := toRows(data)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			return err
		}
	}
	return nil
}

type tableFormatter struct {
	fields []string
}

func (f *tableFormatter) Format(w io.Writer, data any) error {
	rows, err := toRows(data)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Fprintln(w, "No data")
		return nil
	}

	rows = flattenHashFields(rows)
	headers := f.resolveHeaders(rows)
	if len(headers) == 0 {
		fmt.Fprintln(w, "No data")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	// Header row
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, h)
	}
	fmt.Fprintln(tw)

	// Separator
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, strings.Repeat("-", len(h)))
	}
	fmt.Fprintln(tw)

	// Data rows
	for _, row := range rows {
		for i, h := range headers {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, formatValue(row[h]))
		}
		fmt.Fprintln(tw)
	}

	return tw.Flush()
}

type csvFormatter struct {
	fields []string
}

func (f *csvFormatter) Format(w io.Writer, data any) error {
	rows, err := toRows(data)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	rows = flattenHashFields(rows)
	headers := f.resolveHeaders(rows)
	if len(headers) == 0 {
		return nil
	}

	cw := csv.NewWriter(w)
	if err := cw.Write(headers); err != nil {
		return err
	}

	for _, row := range rows {
		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = formatValue(row[h])
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func (f *tableFormatter) resolveHeaders(rows []map[string]any) []string {
	if len(f.fields) > 0 {
		return f.fields
	}
	return collectHeaders(rows)
}

func (f *csvFormatter) resolveHeaders(rows []map[string]any) []string {
	if len(f.fields) > 0 {
		return f.fields
	}
	return collectHeaders(rows)
}

func collectHeaders(rows []map[string]any) []string {
	seen := map[string]bool{}
	var headers []string
	for _, row := range rows {
		for k := range row {
			if !seen[k] {
				seen[k] = true
				headers = append(headers, k)
			}
		}
	}
	sort.Strings(headers)
	return headers
}

// hashFieldSuffixes are the Hash field suffixes used by Pleasanter.
var hashFieldSuffixes = []string{
	"Hash",
}

// isHashField checks if a field name ends with "Hash" (e.g. ClassHash, NumHash).
func isHashField(name string) bool {
	for _, suffix := range hashFieldSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// flattenHashFields expands Hash fields (e.g. ClassHash: {"ClassA":"v1","ClassB":"v2"})
// into individual columns (ClassA, ClassB). Returns new rows and updated headers.
func flattenHashFields(rows []map[string]any) []map[string]any {
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		newRow := make(map[string]any, len(row))
		for k, v := range row {
			if isHashField(k) {
				if m, ok := v.(map[string]any); ok {
					for subKey, subVal := range m {
						newRow[subKey] = subVal
					}
					continue
				}
			}
			newRow[k] = v
		}
		result = append(result, newRow)
	}
	return result
}

// UnwrapAPIResponse detects Pleasanter API response structure
// {"Response":{"Data":[...],...},"StatusCode":200} and extracts Data.
// Returns (unwrapped data, metadata) if the structure matches,
// or (original data, nil) if it doesn't.
func UnwrapAPIResponse(result map[string]any) (any, map[string]any) {
	resp, ok := result["Response"]
	if !ok {
		return result, nil
	}
	respMap, ok := resp.(map[string]any)
	if !ok {
		return result, nil
	}
	data, exists := respMap["Data"]
	if !exists {
		return result, nil
	}

	// Build metadata from the Response (excluding Data)
	meta := make(map[string]any, len(respMap)-1)
	for k, v := range respMap {
		if k != "Data" {
			meta[k] = v
		}
	}
	return data, meta
}

func toRows(data any) ([]map[string]any, error) {
	switch v := data.(type) {
	case []map[string]any:
		return v, nil
	case map[string]any:
		return []map[string]any{v}, nil
	case []any:
		rows := make([]map[string]any, 0, len(v))
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expected map, got %T", item)
			}
			rows = append(rows, m)
		}
		return rows, nil
	default:
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		var result any
		if err := json.Unmarshal(b, &result); err != nil {
			return nil, err
		}
		return toRows(result)
	}
}

type countFormatter struct {
	fields []string
}

func (f *countFormatter) Format(w io.Writer, data any) error {
	rows, err := toRows(data)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%d\n", len(rows))
	return nil
}

type idsFormatter struct {
	fields []string
}

func (f *idsFormatter) Format(w io.Writer, data any) error {
	rows, err := toRows(data)
	if err != nil {
		return err
	}
	for _, row := range rows {
		id := extractID(row)
		if id != "" {
			fmt.Fprintf(w, "%s\n", id)
		}
	}
	return nil
}

func extractID(row map[string]any) string {
	for _, key := range []string{"ResultId", "IssueId", "SiteId", "Id"} {
		if v, ok := row[key]; ok {
			return formatValue(v)
		}
	}
	return ""
}

func formatValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case json.Number:
		return val.String()
	case float64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case map[string]any, []any:
		b, _ := json.Marshal(val)
		return string(b)
	default:
		return fmt.Sprintf("%v", val)
	}
}
