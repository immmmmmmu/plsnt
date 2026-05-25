package pleasanter

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

// MappingConfig defines how CSV columns map to Pleasanter fields.
type MappingConfig struct {
	Columns  map[string]string `yaml:"columns"`
	Defaults map[string]any    `yaml:"defaults,omitempty"`
}

// LoadMapping reads and parses a YAML mapping file.
func LoadMapping(path string) (*MappingConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("failed to read mapping file: %v", err)).
			WithSuggestion("Check that the mapping file path is correct and readable")
	}

	var cfg MappingConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("failed to parse mapping YAML: %v", err)).
			WithSuggestion("Check the mapping file syntax")
	}

	if len(cfg.Columns) == 0 {
		return nil, errs.New(errs.CodeInvalidInput,
			"mapping file must define at least one column mapping").
			WithSuggestion("Add a 'columns' section with CSV-to-Pleasanter field mappings")
	}

	return &cfg, nil
}

// hashPrefixes maps Pleasanter field name prefixes to their Hash field names.
var hashPrefixes = map[string]string{
	"Class":       "ClassHash",
	"Num":         "NumHash",
	"Date":        "DateHash",
	"Description": "DescriptionHash",
	"Check":       "CheckHash",
}

// topLevelFields are Pleasanter fields that go at the top level of the record payload.
var topLevelFields = map[string]bool{
	"Title":  true,
	"Body":   true,
	"Status": true,
}

// classifyField determines where a Pleasanter field should be placed in the payload.
// It returns (hashName, fieldName, isHash).
// For example: "ClassA" -> ("ClassHash", "ClassA", true)
//
//	"Title"  -> ("", "Title", false)
func classifyField(field string) (string, string, bool) {
	if topLevelFields[field] {
		return "", field, false
	}

	for prefix, hashName := range hashPrefixes {
		if strings.HasPrefix(field, prefix) && len(field) > len(prefix) {
			return hashName, field, true
		}
	}

	// Unknown fields go at top level
	return "", field, false
}

// ParseCSV reads CSV data and applies the mapping to produce record payloads.
func ParseCSV(r io.Reader, mapping *MappingConfig) ([]map[string]any, error) {
	reader := csv.NewReader(r)

	// Read header row
	header, err := reader.Read()
	if err != nil {
		return nil, errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("failed to read CSV header: %v", err)).
			WithSuggestion("Ensure the CSV file has a header row")
	}

	// Build column index: CSV column index -> Pleasanter field name
	colMap := make(map[int]string)
	for i, col := range header {
		col = strings.TrimSpace(col)
		if target, ok := mapping.Columns[col]; ok {
			colMap[i] = target
		}
	}

	if len(colMap) == 0 {
		return nil, errs.New(errs.CodeInvalidInput,
			"no CSV columns matched the mapping").
			WithSuggestion("Check that column names in the mapping file match the CSV header exactly")
	}

	var records []map[string]any
	rowNum := 1 // 1-indexed (header is row 0)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errs.New(errs.CodeInvalidInput,
				fmt.Sprintf("failed to read CSV row %d: %v", rowNum+1, err)).
				WithSuggestion("Check the CSV file for formatting issues")
		}
		rowNum++

		record := make(map[string]any)

		// Apply defaults first
		for key, val := range mapping.Defaults {
			record[key] = deepCopyAny(val)
		}

		// Apply mapped columns
		for colIdx, field := range colMap {
			if colIdx >= len(row) {
				continue
			}
			value := row[colIdx]

			hashName, fieldName, isHash := classifyField(field)
			if isHash {
				hash, ok := record[hashName].(map[string]any)
				if !ok {
					hash = make(map[string]any)
					record[hashName] = hash
				}
				hash[fieldName] = value
			} else {
				record[fieldName] = value
			}
		}

		records = append(records, record)
	}

	return records, nil
}

// deepCopyAny creates a deep copy of a value to prevent shared references between records.
func deepCopyAny(v any) any {
	switch val := v.(type) {
	case map[string]any:
		cp := make(map[string]any, len(val))
		for k, v2 := range val {
			cp[k] = deepCopyAny(v2)
		}
		return cp
	case []any:
		cp := make([]any, len(val))
		for i, v2 := range val {
			cp[i] = deepCopyAny(v2)
		}
		return cp
	default:
		return v
	}
}
