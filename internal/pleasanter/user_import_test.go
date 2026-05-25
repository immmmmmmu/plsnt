package pleasanter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/api"
)

func TestParseUserCSV_Success(t *testing.T) {
	csv := "LoginId,Name,Password,DeptId,FirstAndLastNameOrder\nuser001,Yamada Taro,P@ssw0rd,1,1\nuser002,Suzuki Hanako,P@ssw0rd,2,1\n"
	users, err := ParseUserCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if users[0]["LoginId"] != "user001" {
		t.Errorf("expected LoginId 'user001', got %v", users[0]["LoginId"])
	}
	if users[0]["Name"] != "Yamada Taro" {
		t.Errorf("expected Name 'Yamada Taro', got %v", users[0]["Name"])
	}
	if users[0]["Password"] != "P@ssw0rd" {
		t.Errorf("expected Password 'P@ssw0rd', got %v", users[0]["Password"])
	}
	if users[0]["DeptId"] != int64(1) {
		t.Errorf("expected DeptId 1, got %v", users[0]["DeptId"])
	}
	if users[0]["FirstAndLastNameOrder"] != int64(1) {
		t.Errorf("expected FirstAndLastNameOrder 1, got %v", users[0]["FirstAndLastNameOrder"])
	}

	if users[1]["LoginId"] != "user002" {
		t.Errorf("expected LoginId 'user002', got %v", users[1]["LoginId"])
	}
	if users[1]["DeptId"] != int64(2) {
		t.Errorf("expected DeptId 2, got %v", users[1]["DeptId"])
	}
}

func TestParseUserCSV_RequiredOnly(t *testing.T) {
	csv := "LoginId,Name,Password\nuser001,Yamada,secret\n"
	users, err := ParseUserCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if _, ok := users[0]["DeptId"]; ok {
		t.Error("DeptId should not be set when column is absent")
	}
	if _, ok := users[0]["FirstAndLastNameOrder"]; ok {
		t.Error("FirstAndLastNameOrder should not be set when column is absent")
	}
}

func TestParseUserCSV_OptionalEmpty(t *testing.T) {
	csv := "LoginId,Name,Password,DeptId,FirstAndLastNameOrder\nuser001,Yamada,secret,,\n"
	users, err := ParseUserCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if _, ok := users[0]["DeptId"]; ok {
		t.Error("DeptId should not be set when value is empty")
	}
	if _, ok := users[0]["FirstAndLastNameOrder"]; ok {
		t.Error("FirstAndLastNameOrder should not be set when value is empty")
	}
}

func TestParseUserCSV_MissingRequiredColumn(t *testing.T) {
	csv := "LoginId,Password\nuser001,secret\n"
	_, err := ParseUserCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing required column Name")
	}
	if !strings.Contains(err.Error(), "Name") {
		t.Errorf("expected error mentioning 'Name', got: %v", err)
	}
}

func TestParseUserCSV_EmptyInput(t *testing.T) {
	_, err := ParseUserCSV(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseUserCSV_HeaderOnly(t *testing.T) {
	csv := "LoginId,Name,Password\n"
	users, err := ParseUserCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users for header-only CSV, got %d", len(users))
	}
}

func TestParseUserCSV_WhitespaceHeaders(t *testing.T) {
	csv := " LoginId , Name , Password \nuser001,Yamada,secret\n"
	users, err := ParseUserCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0]["LoginId"] != "user001" {
		t.Errorf("expected LoginId 'user001', got %v", users[0]["LoginId"])
	}
}

func TestUserBulkCreate_Success(t *testing.T) {
	var callCount int
	var receivedBodies []map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		receivedBodies = append(receivedBodies, body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Id":100,"Message":"Created"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	users := []map[string]any{
		{"LoginId": "user001", "Name": "Yamada", "Password": "pass1"},
		{"LoginId": "user002", "Name": "Suzuki", "Password": "pass2"},
	}

	results, err := svc.BulkCreate(context.Background(), users)
	if err != nil {
		t.Fatalf("BulkCreate failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
	if receivedBodies[0]["LoginId"] != "user001" {
		t.Errorf("expected first LoginId 'user001', got %v", receivedBodies[0]["LoginId"])
	}
	if receivedBodies[1]["LoginId"] != "user002" {
		t.Errorf("expected second LoginId 'user002', got %v", receivedBodies[1]["LoginId"])
	}
}

func TestUserBulkCreate_PartialFailure(t *testing.T) {
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"StatusCode":500,"Message":"Internal Server Error"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"StatusCode":200,"Id":100,"Message":"Created"}`))
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	users := []map[string]any{
		{"LoginId": "user001", "Name": "Yamada", "Password": "pass1"},
		{"LoginId": "user002", "Name": "Suzuki", "Password": "pass2"},
		{"LoginId": "user003", "Name": "Tanaka", "Password": "pass3"},
	}

	results, err := svc.BulkCreate(context.Background(), users)
	if err == nil {
		t.Fatal("expected error on partial failure")
	}
	// First user should have succeeded before the failure
	if len(results) != 1 {
		t.Errorf("expected 1 successful result before failure, got %d", len(results))
	}
	if !strings.Contains(err.Error(), "user002") {
		t.Errorf("expected error to mention 'user002', got: %v", err)
	}
}

func TestUserBulkCreate_EmptySlice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no API call expected for empty input")
	}))
	defer srv.Close()

	client := api.NewWithHTTPClient(srv.URL, "key", "1.1", srv.Client())
	svc := NewUserService(client)

	results, err := svc.BulkCreate(context.Background(), []map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
