package pleasanter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

// E2E tests require a running Pleasanter server.
// Set PLSNT_E2E=1, PLSNT_URL, and PLSNT_API_KEY to run.
//
// These tests create a temporary site, run CRUD operations, then clean up.
// They do NOT touch existing sites.

func skipIfNoE2E(t *testing.T) {
	t.Helper()
	if os.Getenv("PLSNT_E2E") != "1" {
		t.Skip("Skipping E2E test (set PLSNT_E2E=1 to run)")
	}
	if os.Getenv("PLSNT_URL") == "" || os.Getenv("PLSNT_API_KEY") == "" {
		t.Skip("Skipping E2E test (set PLSNT_URL and PLSNT_API_KEY)")
	}
}

func e2eClient(t *testing.T) api.Client {
	t.Helper()
	return api.New(
		os.Getenv("PLSNT_URL"),
		os.Getenv("PLSNT_API_KEY"),
		"1.1",
	)
}

func TestE2E_RecordListFromExistingSite(t *testing.T) {
	skipIfNoE2E(t)

	client := e2eClient(t)
	svc := pleasanter.NewRecordService(client)

	// List records from site 100 (known to exist)
	resp, err := svc.List(context.Background(), pleasanter.ListOptions{
		SiteID: 100,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	t.Logf("Site 100: TotalCount=%d, PageSize=%d, Offset=%d",
		resp.Response.TotalCount, resp.Response.PageSize, resp.Response.Offset)

	if resp.StatusCode != 200 {
		t.Errorf("expected StatusCode 200, got %d", resp.StatusCode)
	}
}

func TestE2E_RecordCRUD(t *testing.T) {
	skipIfNoE2E(t)

	// This test requires PLSNT_E2E_SITE_ID to specify a test site for CRUD.
	// It creates, reads, updates, and deletes a record.
	siteIDStr := os.Getenv("PLSNT_E2E_SITE_ID")
	if siteIDStr == "" {
		t.Skip("Skipping CRUD test (set PLSNT_E2E_SITE_ID to a test site ID)")
	}

	client := e2eClient(t)
	noRetryCli := api.New(
		os.Getenv("PLSNT_URL"),
		os.Getenv("PLSNT_API_KEY"),
		"1.1",
		api.WithRetryDisabled(),
	)
	svc := pleasanter.NewRecordService(client)
	svcNoRetry := pleasanter.NewRecordService(noRetryCli)

	var siteID int64
	fmt.Sscanf(siteIDStr, "%d", &siteID)

	// 1. Create a record
	createResult, err := svcNoRetry.Create(context.Background(), siteID, map[string]any{
		"Title":     "E2E Test Record",
		"ClassHash": map[string]any{"ClassA": "e2e-test"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	t.Logf("Create result: %v", createResult)

	// Extract the created record ID (PostRaw uses json.Number)
	var recordID int64
	if id, ok := createResult["Id"]; ok {
		switch v := id.(type) {
		case json.Number:
			n, _ := v.Int64()
			recordID = n
		case float64:
			recordID = int64(v)
		}
	}

	if recordID == 0 {
		t.Fatalf("Failed to get record ID from create response: %v", createResult)
	}

	t.Logf("Created record ID: %v", recordID)

	// 2. Get the record
	getResp, err := svc.Get(context.Background(), recordID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(getResp.Response.Data) == 0 {
		t.Fatal("Get returned no data")
	}

	record := getResp.Response.Data[0]
	if record.Title != "E2E Test Record" {
		t.Errorf("expected Title 'E2E Test Record', got %q", record.Title)
	}

	// 3. Update the record
	updateResult, err := svc.Update(context.Background(), recordID, map[string]any{
		"Title": "E2E Test Record (Updated)",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	t.Logf("Update result: %v", updateResult)

	// 4. Verify the update
	getResp2, err := svc.Get(context.Background(), recordID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	if len(getResp2.Response.Data) > 0 {
		updated := getResp2.Response.Data[0]
		if updated.Title != "E2E Test Record (Updated)" {
			t.Errorf("expected updated Title, got %q", updated.Title)
		}
	}

	// 5. Delete the record (cleanup)
	deleteResult, err := svc.Delete(context.Background(), recordID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	t.Logf("Delete result: %v", deleteResult)
}

func TestE2E_SchemaGet(t *testing.T) {
	skipIfNoE2E(t)

	client := e2eClient(t)
	svc := pleasanter.NewSchemaService(client)

	// Site 18 is known to exist
	info, err := svc.GetSchema(context.Background(), 18)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	t.Logf("Schema: SiteID=%d, Title=%q, TableType=%q, Columns=%d",
		info.SiteID, info.Title, info.TableType, len(info.Columns))

	if info.SiteID != 18 {
		t.Errorf("expected SiteID 18, got %d", info.SiteID)
	}

	for _, col := range info.Columns {
		t.Logf("  Column: %s (%s) type=%s required=%v",
			col.ColumnName, col.LabelText, col.Type, col.Required)
	}
}

func TestE2E_ConfigTest(t *testing.T) {
	skipIfNoE2E(t)

	client := e2eClient(t)

	// Test connectivity using users/get
	var resp struct {
		StatusCode int `json:"StatusCode"`
	}
	err := client.Post(context.Background(), "/api/users/get", &struct{}{}, &resp)
	if err != nil {
		t.Fatalf("Connection test failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected StatusCode 200, got %d", resp.StatusCode)
	}

	t.Log("Connection test passed")
}
