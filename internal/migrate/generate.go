package migrate

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

// MappingResult holds the auto-mapping result between CSV columns and Pleasanter fields.
type MappingResult struct {
	Mapped   map[string]string // CSV column -> Pleasanter field
	Unmapped []string          // CSV columns that couldn't be auto-mapped
}

// GenerateMapping auto-maps CSV columns to Pleasanter schema columns.
// Matching priority: exact column name, exact label text, case-insensitive column name, case-insensitive label text.
func GenerateMapping(csvHeaders []string, schema *pleasanter.SchemaInfo) *MappingResult {
	result := &MappingResult{
		Mapped: make(map[string]string),
	}

	for _, header := range csvHeaders {
		header = strings.TrimSpace(header)
		if header == "" {
			continue
		}

		matched := false

		// Pass 1: exact match on column name or label text
		for _, col := range schema.Columns {
			if header == col.ColumnName {
				result.Mapped[header] = col.ColumnName
				matched = true
				break
			}
			if header == col.LabelText {
				result.Mapped[header] = col.ColumnName
				matched = true
				break
			}
		}

		// Pass 2: case-insensitive match
		if !matched {
			headerLower := strings.ToLower(header)
			for _, col := range schema.Columns {
				if strings.ToLower(col.ColumnName) == headerLower || strings.ToLower(col.LabelText) == headerLower {
					result.Mapped[header] = col.ColumnName
					matched = true
					break
				}
			}
		}

		if !matched {
			result.Unmapped = append(result.Unmapped, header)
		}
	}

	return result
}

// ReadCSVHeaders reads only the header row from a CSV file.
func ReadCSVHeaders(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("failed to open CSV file: %v", err)).
			WithSuggestion("Check that the file path is correct and readable")
	}
	defer f.Close()

	reader := csv.NewReader(f)
	header, err := reader.Read()
	if err != nil {
		return nil, errs.New(errs.CodeInvalidInput,
			fmt.Sprintf("failed to read CSV header: %v", err)).
			WithSuggestion("Ensure the CSV file has a header row")
	}

	return header, nil
}

// WriteMappingYAML writes the mapping result as a YAML file.
// Mapped columns are written as a proper YAML mapping.
// Unmapped columns are written as comments for manual editing.
func WriteMappingYAML(w io.Writer, result *MappingResult) error {
	cfg := pleasanter.MappingConfig{
		Columns: result.Mapped,
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return errs.New(errs.CodeInternalError,
			fmt.Sprintf("failed to marshal mapping YAML: %v", err))
	}

	if _, err := w.Write(data); err != nil {
		return errs.New(errs.CodeInternalError,
			fmt.Sprintf("failed to write mapping YAML: %v", err))
	}

	if len(result.Unmapped) > 0 {
		// Sort unmapped columns for deterministic output
		sorted := make([]string, len(result.Unmapped))
		copy(sorted, result.Unmapped)
		sort.Strings(sorted)

		comment := "\n# Unmapped CSV columns (edit manually):\n"
		for _, col := range sorted {
			comment += fmt.Sprintf("#   %s: \"\"\n", col)
		}
		if _, err := io.WriteString(w, comment); err != nil {
			return errs.New(errs.CodeInternalError,
				fmt.Sprintf("failed to write unmapped comments: %v", err))
		}
	}

	return nil
}
