package master

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/pleasanter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRecordService implements the RecordServicer interface for testing.
type mockRecordService struct {
	// listAllResp is returned by ListAll.
	listAllResp *pleasanter.APIResponse
	listAllErr  error

	// createCalls tracks Create invocations.
	createCalls []createCall
	createResp  map[string]any
	createErr   error

	// updateCalls tracks Update invocations.
	updateCalls []updateCall
	updateResp  map[string]any
	updateErr   error
}

type createCall struct {
	SiteID  int64
	Payload map[string]any
}

type updateCall struct {
	RecordID int64
	Payload  map[string]any
}

func (m *mockRecordService) ListAll(ctx context.Context, opts pleasanter.ListOptions) (*pleasanter.APIResponse, error) {
	return m.listAllResp, m.listAllErr
}

func (m *mockRecordService) Create(ctx context.Context, siteID int64, payload map[string]any) (map[string]any, error) {
	m.createCalls = append(m.createCalls, createCall{SiteID: siteID, Payload: payload})
	return m.createResp, m.createErr
}

func (m *mockRecordService) Update(ctx context.Context, recordID int64, payload map[string]any) (map[string]any, error) {
	m.updateCalls = append(m.updateCalls, updateCall{RecordID: recordID, Payload: payload})
	return m.updateResp, m.updateErr
}

func newEmptyListResponse() *pleasanter.APIResponse {
	return &pleasanter.APIResponse{
		StatusCode: 200,
		Response: pleasanter.ResponseBody{
			TotalCount: 0,
			Data:       []pleasanter.Record{},
		},
	}
}

func newListResponseWithRecords(records ...pleasanter.Record) *pleasanter.APIResponse {
	return &pleasanter.APIResponse{
		StatusCode: 200,
		Response: pleasanter.ResponseBody{
			TotalCount: len(records),
			Data:       records,
		},
	}
}

func TestImportCSV_NewRecords(t *testing.T) {
	// CSV with 2 new records (no existing records in Pleasanter)
	csv := "Title,ClassA,NumA\nDept A,D001,100\nDept B,D002,200\n"
	reader := strings.NewReader(csv)

	mock := &mockRecordService{
		listAllResp: newEmptyListResponse(),
		createResp:  map[string]any{"Id": json.Number("1001")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, int64(12345), result.SiteID)
	assert.Equal(t, 2, result.Added)
	assert.Equal(t, 0, result.Updated)
	assert.Equal(t, 0, result.Errors)

	// Verify Create was called twice with correct payloads
	require.Len(t, mock.createCalls, 2)

	call1 := mock.createCalls[0]
	assert.Equal(t, int64(12345), call1.SiteID)
	assert.Equal(t, "Dept A", call1.Payload["Title"])
	classHash1, ok := call1.Payload["ClassHash"].(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "D001", classHash1["ClassA"])

	call2 := mock.createCalls[1]
	assert.Equal(t, "Dept B", call2.Payload["Title"])

	// NumHash should use json.Number
	numHash1, ok := call1.Payload["NumHash"].(map[string]json.Number)
	require.True(t, ok)
	assert.Equal(t, json.Number("100"), numHash1["NumA"])
}

func TestImportCSV_UpdateRecords(t *testing.T) {
	// CSV with 2 records that already exist (matched by ClassA key)
	csv := "Title,ClassA\nUpdated Dept A,D001\nUpdated Dept B,D002\n"
	reader := strings.NewReader(csv)

	existingRecords := newListResponseWithRecords(
		pleasanter.Record{
			ResultId:  5001,
			Title:     "Old Dept A",
			ClassHash: map[string]string{"ClassA": "D001"},
		},
		pleasanter.Record{
			ResultId:  5002,
			Title:     "Old Dept B",
			ClassHash: map[string]string{"ClassA": "D002"},
		},
	)

	mock := &mockRecordService{
		listAllResp: existingRecords,
		updateResp:  map[string]any{"Id": json.Number("5001")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 2, result.Updated)
	assert.Equal(t, 0, result.Errors)

	// Verify Update was called with correct record IDs
	require.Len(t, mock.updateCalls, 2)
	assert.Equal(t, int64(5001), mock.updateCalls[0].RecordID)
	assert.Equal(t, "Updated Dept A", mock.updateCalls[0].Payload["Title"])
	assert.Equal(t, int64(5002), mock.updateCalls[1].RecordID)
}

func TestImportCSV_Mixed(t *testing.T) {
	// CSV with 1 existing (D001) and 1 new (D003)
	csv := "Title,ClassA\nUpdated Dept,D001\nNew Dept,D003\n"
	reader := strings.NewReader(csv)

	existingRecords := newListResponseWithRecords(
		pleasanter.Record{
			ResultId:  5001,
			Title:     "Old Dept",
			ClassHash: map[string]string{"ClassA": "D001"},
		},
	)

	mock := &mockRecordService{
		listAllResp: existingRecords,
		createResp:  map[string]any{"Id": json.Number("5010")},
		updateResp:  map[string]any{"Id": json.Number("5001")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Added)
	assert.Equal(t, 1, result.Updated)
	assert.Equal(t, 0, result.Errors)

	require.Len(t, mock.createCalls, 1)
	require.Len(t, mock.updateCalls, 1)
	assert.Equal(t, "New Dept", mock.createCalls[0].Payload["Title"])
	assert.Equal(t, int64(5001), mock.updateCalls[0].RecordID)
}

func TestImportCSV_DryRun(t *testing.T) {
	csv := "Title,ClassA\nDept A,D001\nDept B,D002\n"
	reader := strings.NewReader(csv)

	existingRecords := newListResponseWithRecords(
		pleasanter.Record{
			ResultId:  5001,
			Title:     "Existing",
			ClassHash: map[string]string{"ClassA": "D001"},
		},
	)

	mock := &mockRecordService{
		listAllResp: existingRecords,
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
		DryRun: true,
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Added)
	assert.Equal(t, 1, result.Updated)
	assert.Equal(t, 0, result.Errors)

	// No actual API calls should be made in dry-run mode
	assert.Empty(t, mock.createCalls, "dry-run should not call Create")
	assert.Empty(t, mock.updateCalls, "dry-run should not call Update")
}

func TestImportCSV_EmptyCSV(t *testing.T) {
	csv := "Title,ClassA\n"
	reader := strings.NewReader(csv)

	mock := &mockRecordService{
		listAllResp: newEmptyListResponse(),
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 0, result.Updated)
}

func TestImportCSV_FieldMapping(t *testing.T) {
	// Test all field types: Title, Body, ClassHash, NumHash, DateHash, DescriptionHash, CheckHash
	csv := "Title,Body,ClassA,NumA,DateA,DescriptionA,CheckA\nTest,Desc,code1,42.5,2026-04-01,Long text,true\n"
	reader := strings.NewReader(csv)

	mock := &mockRecordService{
		listAllResp: newEmptyListResponse(),
		createResp:  map[string]any{"Id": json.Number("1")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	_, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	require.Len(t, mock.createCalls, 1)

	payload := mock.createCalls[0].Payload
	assert.Equal(t, "Test", payload["Title"])
	assert.Equal(t, "Desc", payload["Body"])

	classHash := payload["ClassHash"].(map[string]string)
	assert.Equal(t, "code1", classHash["ClassA"])

	numHash := payload["NumHash"].(map[string]json.Number)
	assert.Equal(t, json.Number("42.5"), numHash["NumA"])

	dateHash := payload["DateHash"].(map[string]string)
	assert.Equal(t, "2026-04-01", dateHash["DateA"])

	descHash := payload["DescriptionHash"].(map[string]string)
	assert.Equal(t, "Long text", descHash["DescriptionA"])

	checkHash := payload["CheckHash"].(map[string]bool)
	assert.Equal(t, true, checkHash["CheckA"])
}

func TestImportCSV_CheckHashFalse(t *testing.T) {
	csv := "Title,ClassA,CheckA\nTest,code1,false\n"
	reader := strings.NewReader(csv)

	mock := &mockRecordService{
		listAllResp: newEmptyListResponse(),
		createResp:  map[string]any{"Id": json.Number("1")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	_, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	require.Len(t, mock.createCalls, 1)

	checkHash := mock.createCalls[0].Payload["CheckHash"].(map[string]bool)
	assert.Equal(t, false, checkHash["CheckA"])
}

func TestImportCSV_IssueIdKey(t *testing.T) {
	// When existing record has IssueId instead of ResultId, update should use IssueId
	csv := "Title,ClassA\nUpdated,D001\n"
	reader := strings.NewReader(csv)

	existingRecords := newListResponseWithRecords(
		pleasanter.Record{
			IssueId:   7001,
			Title:     "Old",
			ClassHash: map[string]string{"ClassA": "D001"},
		},
	)

	mock := &mockRecordService{
		listAllResp: existingRecords,
		updateResp:  map[string]any{"Id": json.Number("7001")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Updated)
	require.Len(t, mock.updateCalls, 1)
	assert.Equal(t, int64(7001), mock.updateCalls[0].RecordID)
}

func TestImportCSV_InvalidCSV(t *testing.T) {
	// Header only, no closing newline — this is valid (0 data rows)
	csv := "Title,ClassA"
	reader := strings.NewReader(csv)

	mock := &mockRecordService{
		listAllResp: newEmptyListResponse(),
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Added)
}

func TestImportCSV_TitleKey(t *testing.T) {
	// Use Title as key column
	csv := "Title,ClassA\nExisting Dept,D001\nNew Dept,D002\n"
	reader := strings.NewReader(csv)

	existingRecords := newListResponseWithRecords(
		pleasanter.Record{
			ResultId:  5001,
			Title:     "Existing Dept",
			ClassHash: map[string]string{"ClassA": "OLD001"},
		},
	)

	mock := &mockRecordService{
		listAllResp: existingRecords,
		createResp:  map[string]any{"Id": json.Number("5010")},
		updateResp:  map[string]any{"Id": json.Number("5001")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "Title",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Added)
	assert.Equal(t, 1, result.Updated)
}

func TestImportCSV_NumHashKey(t *testing.T) {
	// Use NumA as key column
	csv := "Title,NumA\nDept A,100\n"
	reader := strings.NewReader(csv)

	existingRecords := newListResponseWithRecords(
		pleasanter.Record{
			ResultId: 5001,
			Title:    "Old Dept",
			NumHash:  map[string]json.Number{"NumA": json.Number("100")},
		},
	)

	mock := &mockRecordService{
		listAllResp: existingRecords,
		updateResp:  map[string]any{"Id": json.Number("5001")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "NumA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 1, result.Updated)
}

func TestImportCSV_DateHashKey(t *testing.T) {
	csv := "Title,DateA\nDept,2026-04-01\n"
	reader := strings.NewReader(csv)

	existingRecords := newListResponseWithRecords(
		pleasanter.Record{
			ResultId: 5001,
			Title:    "Old",
			DateHash: map[string]string{"DateA": "2026-04-01"},
		},
	)

	mock := &mockRecordService{
		listAllResp: existingRecords,
		updateResp:  map[string]any{"Id": json.Number("5001")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "DateA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Updated)
}

func TestImportCSV_DescriptionHashKey(t *testing.T) {
	csv := "Title,DescriptionA\nDept,code-desc\n"
	reader := strings.NewReader(csv)

	existingRecords := newListResponseWithRecords(
		pleasanter.Record{
			ResultId:        5001,
			Title:           "Old",
			DescriptionHash: map[string]string{"DescriptionA": "code-desc"},
		},
	)

	mock := &mockRecordService{
		listAllResp: existingRecords,
		updateResp:  map[string]any{"Id": json.Number("5001")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "DescriptionA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Updated)
}

func TestImportCSV_CreateError(t *testing.T) {
	csv := "Title,ClassA\nDept,D001\n"
	reader := strings.NewReader(csv)

	mock := &mockRecordService{
		listAllResp: newEmptyListResponse(),
		createErr:   fmt.Errorf("api error"),
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 1, result.Errors)
}

func TestImportCSV_UpdateError(t *testing.T) {
	csv := "Title,ClassA\nDept,D001\n"
	reader := strings.NewReader(csv)

	existingRecords := newListResponseWithRecords(
		pleasanter.Record{
			ResultId:  5001,
			Title:     "Old",
			ClassHash: map[string]string{"ClassA": "D001"},
		},
	)

	mock := &mockRecordService{
		listAllResp: existingRecords,
		updateErr:   fmt.Errorf("api error"),
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Updated)
	assert.Equal(t, 1, result.Errors)
}

func TestImportCSV_ListAllError(t *testing.T) {
	csv := "Title,ClassA\nDept,D001\n"
	reader := strings.NewReader(csv)

	mock := &mockRecordService{
		listAllErr: fmt.Errorf("connection error"),
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	_, err := importer.ImportCSV(context.Background(), reader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch existing records")
}

func TestImportCSV_NilHashKeys(t *testing.T) {
	// Existing record with nil hashes should not panic when key is ClassA
	csv := "Title,ClassA\nDept,D001\n"
	reader := strings.NewReader(csv)

	existingRecords := newListResponseWithRecords(
		pleasanter.Record{
			ResultId: 5001,
			Title:    "Old",
			// ClassHash is nil
		},
	)

	mock := &mockRecordService{
		listAllResp: existingRecords,
		createResp:  map[string]any{"Id": json.Number("1")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	// Should be treated as new since nil ClassHash means no match
	assert.Equal(t, 1, result.Added)
}

func TestImportCSV_UTF8BOM(t *testing.T) {
	// CSV with UTF-8 BOM prefix
	csv := "\xEF\xBB\xBFTitle,ClassA\nDept,D001\n"
	reader := strings.NewReader(csv)

	mock := &mockRecordService{
		listAllResp: newEmptyListResponse(),
		createResp:  map[string]any{"Id": json.Number("1")},
	}

	importer := NewImporter(mock, Options{
		SiteID: 12345,
		Key:    "ClassA",
	})

	result, err := importer.ImportCSV(context.Background(), reader)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Added)
	assert.Equal(t, "Dept", mock.createCalls[0].Payload["Title"])
}
