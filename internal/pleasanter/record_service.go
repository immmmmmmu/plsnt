package pleasanter

import (
	"context"
	"fmt"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/errs"
)

const (
	// maxListAllRecords is the safety limit for ListAll (10000 records).
	maxListAllRecords = 10000
	// defaultPageSize is the default page size used by Pleasanter API.
	defaultPageSize = 200
)

// RecordService provides record operations.
type RecordService struct {
	client api.Client
}

// NewRecordService creates a new RecordService.
func NewRecordService(client api.Client) *RecordService {
	return &RecordService{client: client}
}

// Get retrieves a single record by ID.
func (s *RecordService) Get(ctx context.Context, recordID int64) (*APIResponse, error) {
	endpoint := fmt.Sprintf("/api/items/%d/get", recordID)

	var resp APIResponse
	if err := s.client.Post(ctx, endpoint, &struct{}{}, &resp); err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, errs.New(errs.CodeRecordNotFound,
			fmt.Sprintf("record %d not found", recordID)).
			WithSuggestion("Check the record ID")
	}

	return &resp, nil
}

// ListOptions configures a record list request.
type ListOptions struct {
	SiteID int64
	Offset int
	View   *View
}

// List retrieves records from a site.
func (s *RecordService) List(ctx context.Context, opts ListOptions) (*APIResponse, error) {
	endpoint := fmt.Sprintf("/api/items/%d/get", opts.SiteID)

	req := map[string]any{}
	if opts.Offset > 0 {
		req["Offset"] = opts.Offset
	}
	if opts.View != nil {
		req["View"] = opts.View
	}

	var resp APIResponse
	if err := s.client.Post(ctx, endpoint, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ListAll retrieves all records by automatically paging through results.
// It stops when all records are fetched or the safety limit (10000 records) is reached.
func (s *RecordService) ListAll(ctx context.Context, opts ListOptions) (*APIResponse, error) {
	opts.Offset = 0
	first, err := s.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	allRecords := make([]Record, 0, first.Response.TotalCount)
	allRecords = append(allRecords, first.Response.Data...)

	pageSize := first.Response.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	totalCount := first.Response.TotalCount

	for len(allRecords) < totalCount && len(allRecords) < maxListAllRecords {
		opts.Offset = len(allRecords)
		page, err := s.List(ctx, opts)
		if err != nil {
			return nil, err
		}
		if len(page.Response.Data) == 0 {
			break
		}
		allRecords = append(allRecords, page.Response.Data...)
	}

	return &APIResponse{
		StatusCode: first.StatusCode,
		Response: ResponseBody{
			Offset:     0,
			PageSize:   pageSize,
			TotalCount: totalCount,
			Data:       allRecords,
		},
	}, nil
}

// ListRaw retrieves records using a raw JSON payload (for --json bypass).
func (s *RecordService) ListRaw(ctx context.Context, siteID int64, payload map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/get", siteID)
	return s.client.PostRaw(ctx, endpoint, payload)
}

// GetRaw retrieves a single record using raw JSON payload.
func (s *RecordService) GetRaw(ctx context.Context, recordID int64, payload map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/get", recordID)
	return s.client.PostRaw(ctx, endpoint, payload)
}

// Create creates a new record in the given site.
// IMPORTANT: create is non-idempotent, so retry must be disabled.
func (s *RecordService) Create(ctx context.Context, siteID int64, payload map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/create", siteID)
	return s.noRetryClient().PostRaw(ctx, endpoint, payload)
}

// Update updates an existing record.
func (s *RecordService) Update(ctx context.Context, recordID int64, payload map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/update", recordID)
	return s.client.PostRaw(ctx, endpoint, payload)
}

// Delete deletes an existing record.
func (s *RecordService) Delete(ctx context.Context, recordID int64) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/delete", recordID)
	return s.client.PostRaw(ctx, endpoint, map[string]any{})
}

// BulkUpsert performs a bulk upsert (update-or-insert) for multiple records.
// Keys specifies which columns serve as the unique key for matching.
// Data is the array of record objects to upsert.
func (s *RecordService) BulkUpsert(ctx context.Context, siteID int64, keys []string, data []map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/bulkupsert", siteID)
	payload := map[string]any{
		"Keys": keys,
		"Data": data,
	}
	return s.client.PostRaw(ctx, endpoint, payload)
}

// Import imports records into a site. If keys are specified, it uses bulkupsert;
// otherwise it creates records one by one (non-idempotent).
func (s *RecordService) Import(ctx context.Context, siteID int64, records []map[string]any, keys []string) ([]map[string]any, error) {
	if len(keys) > 0 {
		result, err := s.BulkUpsert(ctx, siteID, keys, records)
		if err != nil {
			return nil, err
		}
		return []map[string]any{result}, nil
	}

	var results []map[string]any
	for _, rec := range records {
		result, err := s.Create(ctx, siteID, rec)
		if err != nil {
			return results, fmt.Errorf("failed to create record %d: %w", len(results)+1, err)
		}
		results = append(results, result)
	}
	return results, nil
}

// BulkDelete deletes multiple records by IDs.
func (s *RecordService) BulkDelete(ctx context.Context, siteID int64, recordIDs []int64) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/bulkdelete", siteID)
	payload := map[string]any{
		"SelectedIds": recordIDs,
	}
	return s.client.PostRaw(ctx, endpoint, payload)
}

// BulkDeleteByView deletes records matching a View filter.
func (s *RecordService) BulkDeleteByView(ctx context.Context, siteID int64, view map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/bulkdelete", siteID)
	payload := map[string]any{
		"View": view,
	}
	return s.client.PostRaw(ctx, endpoint, payload)
}

// noRetryClient returns a client with retry disabled for non-idempotent operations.
func (s *RecordService) noRetryClient() api.Client {
	return s.client.NoRetry()
}
