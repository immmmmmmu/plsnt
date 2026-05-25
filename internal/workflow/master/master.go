package master

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

// RecordServicer is the interface for record operations needed by the master importer.
type RecordServicer interface {
	ListAll(ctx context.Context, opts pleasanter.ListOptions) (*pleasanter.APIResponse, error)
	Create(ctx context.Context, siteID int64, payload map[string]any) (map[string]any, error)
	Update(ctx context.Context, recordID int64, payload map[string]any) (map[string]any, error)
}

// Options configures the master import operation.
type Options struct {
	SiteID int64
	Key    string
	DryRun bool
}

// Importer performs master data import from CSV.
type Importer struct {
	svc  RecordServicer
	opts Options
}

// NewImporter creates a new master data importer.
func NewImporter(svc RecordServicer, opts Options) *Importer {
	return &Importer{svc: svc, opts: opts}
}

// classRegex matches ClassA through ClassZ.
var classRegex = regexp.MustCompile(`^Class[A-Z]$`)

// numRegex matches NumA through NumZ.
var numRegex = regexp.MustCompile(`^Num[A-Z]$`)

// dateRegex matches DateA through DateZ.
var dateRegex = regexp.MustCompile(`^Date[A-Z]$`)

// descRegex matches DescriptionA through DescriptionZ.
var descRegex = regexp.MustCompile(`^Description[A-Z]$`)

// checkRegex matches CheckA through CheckZ.
var checkRegex = regexp.MustCompile(`^Check[A-Z]$`)

// ImportCSV reads a CSV from the reader and performs upsert against Pleasanter.
func (imp *Importer) ImportCSV(ctx context.Context, r io.Reader) (*MasterResult, error) {
	// Read all data for BOM detection
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Strip UTF-8 BOM if present
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	csvReader := csv.NewReader(bytes.NewReader(data))

	// Read header
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Read all rows
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV data: %w", err)
	}

	// Fetch existing records for key-based matching
	existingMap, err := imp.buildExistingMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing records: %w", err)
	}

	result := &MasterResult{
		SiteID: imp.opts.SiteID,
	}

	for _, row := range rows {
		payload := imp.buildPayload(headers, row)

		// Extract key value for matching
		keyValue := imp.extractKeyValue(headers, row)

		if existingRecord, found := existingMap[keyValue]; found {
			// Update existing record
			recordID := getRecordID(existingRecord)
			if !imp.opts.DryRun {
				if _, err := imp.svc.Update(ctx, recordID, payload); err != nil {
					result.Errors++
					continue
				}
			}
			result.Updated++
		} else {
			// Create new record
			if !imp.opts.DryRun {
				if _, err := imp.svc.Create(ctx, imp.opts.SiteID, payload); err != nil {
					result.Errors++
					continue
				}
			}
			result.Added++
		}
	}

	return result, nil
}

// buildExistingMap fetches all existing records and indexes them by key column value.
func (imp *Importer) buildExistingMap(ctx context.Context) (map[string]pleasanter.Record, error) {
	resp, err := imp.svc.ListAll(ctx, pleasanter.ListOptions{
		SiteID: imp.opts.SiteID,
	})
	if err != nil {
		return nil, err
	}

	existing := make(map[string]pleasanter.Record, len(resp.Response.Data))
	for _, rec := range resp.Response.Data {
		keyVal := extractKeyFromRecord(rec, imp.opts.Key)
		if keyVal != "" {
			existing[keyVal] = rec
		}
	}
	return existing, nil
}

// extractKeyFromRecord extracts the key column value from an existing record.
func extractKeyFromRecord(rec pleasanter.Record, key string) string {
	if key == "Title" {
		return rec.Title
	}
	if classRegex.MatchString(key) {
		if rec.ClassHash != nil {
			return rec.ClassHash[key]
		}
		return ""
	}
	if numRegex.MatchString(key) {
		if rec.NumHash != nil {
			return string(rec.NumHash[key])
		}
		return ""
	}
	if dateRegex.MatchString(key) {
		if rec.DateHash != nil {
			return rec.DateHash[key]
		}
		return ""
	}
	if descRegex.MatchString(key) {
		if rec.DescriptionHash != nil {
			return rec.DescriptionHash[key]
		}
		return ""
	}
	return ""
}

// extractKeyValue extracts the key column value from a CSV row.
func (imp *Importer) extractKeyValue(headers []string, row []string) string {
	for i, h := range headers {
		if h == imp.opts.Key && i < len(row) {
			return row[i]
		}
	}
	return ""
}

// buildPayload constructs a Pleasanter API payload from CSV headers and a row.
func (imp *Importer) buildPayload(headers []string, row []string) map[string]any {
	payload := make(map[string]any)
	classHash := make(map[string]string)
	numHash := make(map[string]json.Number)
	dateHash := make(map[string]string)
	descHash := make(map[string]string)
	checkHash := make(map[string]bool)

	for i, h := range headers {
		if i >= len(row) {
			break
		}
		val := row[i]

		switch {
		case h == "Title":
			payload["Title"] = val
		case h == "Body":
			payload["Body"] = val
		case classRegex.MatchString(h):
			classHash[h] = val
		case numRegex.MatchString(h):
			numHash[h] = json.Number(val)
		case dateRegex.MatchString(h):
			dateHash[h] = val
		case descRegex.MatchString(h):
			descHash[h] = val
		case checkRegex.MatchString(h):
			checkHash[h] = strings.EqualFold(val, "true")
		}
	}

	if len(classHash) > 0 {
		payload["ClassHash"] = classHash
	}
	if len(numHash) > 0 {
		payload["NumHash"] = numHash
	}
	if len(dateHash) > 0 {
		payload["DateHash"] = dateHash
	}
	if len(descHash) > 0 {
		payload["DescriptionHash"] = descHash
	}
	if len(checkHash) > 0 {
		payload["CheckHash"] = checkHash
	}

	return payload
}

// getRecordID returns the record ID from a Record.
// Pleasanter uses ResultId for Results tables and IssueId for Issues tables.
func getRecordID(rec pleasanter.Record) int64 {
	if rec.ResultId > 0 {
		return rec.ResultId
	}
	return rec.IssueId
}
