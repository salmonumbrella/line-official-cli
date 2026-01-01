package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestClient_GetAudienceGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/list" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audienceGroups":[{"audienceGroupId":12345,"description":"Test Audience","status":"READY","audienceCount":100}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	groups, err := client.GetAudienceGroups(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if *groups[0].AudienceGroupId != 12345 {
		t.Errorf("expected audience group ID 12345, got %d", *groups[0].AudienceGroupId)
	}
}

func TestClient_GetAudienceGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/12345" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audienceGroup":{"audienceGroupId":12345,"description":"Test Audience","status":"READY","audienceCount":100}}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.GetAudienceGroup(context.Background(), 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AudienceGroup == nil {
		t.Fatal("expected audience group, got nil")
	}
	if *resp.AudienceGroup.AudienceGroupId != 12345 {
		t.Errorf("expected audience group ID 12345, got %d", *resp.AudienceGroup.AudienceGroupId)
	}
}

func TestClient_DeleteAudienceGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/12345" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.DeleteAudienceGroup(context.Background(), 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateAudienceGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/upload" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req CreateAudienceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Description != "Test Audience" {
			t.Errorf("expected description 'Test Audience', got %s", req.Description)
		}
		if len(req.Audiences) != 2 {
			t.Errorf("expected 2 audiences, got %d", len(req.Audiences))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audienceGroupId":12345,"type":"UPLOAD","description":"Test Audience","created":1609459200000}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.CreateAudienceGroup(context.Background(), "Test Audience", []string{"U123", "U456"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AudienceGroupID != 12345 {
		t.Errorf("expected audience group ID 12345, got %d", resp.AudienceGroupID)
	}
}

func TestClient_AddUsersToAudience(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/upload" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var req AddUsersToAudienceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.AudienceGroupID != 12345 {
			t.Errorf("expected audience group ID 12345, got %d", req.AudienceGroupID)
		}
		if len(req.Audiences) != 2 {
			t.Errorf("expected 2 audiences, got %d", len(req.Audiences))
		}
		if req.UploadDescription != "Batch 2" {
			t.Errorf("expected upload description 'Batch 2', got %s", req.UploadDescription)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.AddUsersToAudience(context.Background(), 12345, []string{"U789", "U012"}, "Batch 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateClickBasedAudience(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/click" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req CreateClickBasedAudienceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Description != "Click Audience" {
			t.Errorf("expected description 'Click Audience', got %s", req.Description)
		}
		if req.RequestID != "req-123" {
			t.Errorf("expected request ID 'req-123', got %s", req.RequestID)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audienceGroupId":12346,"type":"CLICK","description":"Click Audience","created":1609459200000}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.CreateClickBasedAudience(context.Background(), "Click Audience", "req-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AudienceGroupID != 12346 {
		t.Errorf("expected audience group ID 12346, got %d", resp.AudienceGroupID)
	}
}

func TestClient_CreateImpressionBasedAudience(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/imp" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req CreateImpressionBasedAudienceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Description != "Impression Audience" {
			t.Errorf("expected description 'Impression Audience', got %s", req.Description)
		}
		if req.RequestID != "req-456" {
			t.Errorf("expected request ID 'req-456', got %s", req.RequestID)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audienceGroupId":12347,"type":"IMP","description":"Impression Audience","created":1609459200000}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.CreateImpressionBasedAudience(context.Background(), "Impression Audience", "req-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AudienceGroupID != 12347 {
		t.Errorf("expected audience group ID 12347, got %d", resp.AudienceGroupID)
	}
}

func TestClient_UpdateAudienceDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/12345/updateDescription" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var req UpdateDescriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Description != "Updated Description" {
			t.Errorf("expected description 'Updated Description', got %s", req.Description)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.UpdateAudienceDescription(context.Background(), 12345, "Updated Description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetSharedAudienceGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/shared/list" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audienceGroups":[{"audienceGroupId":99999,"description":"Shared Audience","status":"READY","audienceCount":500}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	groups, err := client.GetSharedAudienceGroups(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if *groups[0].AudienceGroupId != 99999 {
		t.Errorf("expected audience group ID 99999, got %d", *groups[0].AudienceGroupId)
	}
}

func TestClient_GetSharedAudienceGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/shared/99999" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audienceGroup":{"audienceGroupId":99999,"description":"Shared Audience","status":"READY","audienceCount":500},"owner":{"name":"Test Owner","serviceType":"LINE_OA"}}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.GetSharedAudienceGroup(context.Background(), 99999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AudienceGroup == nil {
		t.Fatal("expected audience group, got nil")
	}
	if *resp.AudienceGroup.AudienceGroupId != 99999 {
		t.Errorf("expected audience group ID 99999, got %d", *resp.AudienceGroup.AudienceGroupId)
	}
	if resp.Owner == nil {
		t.Fatal("expected owner, got nil")
	}
	if *resp.Owner.Name != "Test Owner" {
		t.Errorf("expected owner name 'Test Owner', got %s", *resp.Owner.Name)
	}
}

func TestClient_GetAudienceGroups_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audienceGroups":null}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	groups, err := client.GetAudienceGroups(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected empty slice, got %d groups", len(groups))
	}
}

func TestClient_GetSharedAudienceGroups_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audienceGroups":null}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	groups, err := client.GetSharedAudienceGroups(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected empty slice, got %d groups", len(groups))
	}
}

func TestClient_CreateAudienceFromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/upload/byFile" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Verify Content-Type is multipart/form-data
		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			t.Error("expected Content-Type header")
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}

		// Check description field
		description := r.FormValue("description")
		if description != "Test File Audience" {
			t.Errorf("expected description 'Test File Audience', got %s", description)
		}

		// Check file was uploaded
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("failed to get uploaded file: %v", err)
		}
		defer func() { _ = file.Close() }()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audienceGroupId":12348,"type":"UPLOAD","description":"Test File Audience","created":1609459200000}`))
	}))
	defer server.Close()

	// Create a temp file with user IDs
	tempFile, err := createTempFileWithContent("U123\nU456\nU789\n")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer removeTempFile(tempFile)

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.CreateAudienceFromFile(context.Background(), "Test File Audience", tempFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AudienceGroupID != 12348 {
		t.Errorf("expected audience group ID 12348, got %d", resp.AudienceGroupID)
	}
}

func TestClient_AddUsersToAudienceFromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/audienceGroup/upload/byFile" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}

		// Check audience group ID field
		audienceGroupID := r.FormValue("audienceGroupId")
		if audienceGroupID != "12345" {
			t.Errorf("expected audienceGroupId '12345', got %s", audienceGroupID)
		}

		// Check upload description field
		uploadDescription := r.FormValue("uploadDescription")
		if uploadDescription != "Batch 3" {
			t.Errorf("expected uploadDescription 'Batch 3', got %s", uploadDescription)
		}

		// Check file was uploaded
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("failed to get uploaded file: %v", err)
		}
		defer func() { _ = file.Close() }()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a temp file with user IDs
	tempFile, err := createTempFileWithContent("U111\nU222\n")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer removeTempFile(tempFile)

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err = client.AddUsersToAudienceFromFile(context.Background(), 12345, tempFile, "Batch 3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateAudienceFromFile_EmptyFile(t *testing.T) {
	// Create a temp file with only whitespace
	tempFile, err := createTempFileWithContent("   \n\n  \n")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer removeTempFile(tempFile)

	client := NewClient("test-token", false, false)

	_, err = client.CreateAudienceFromFile(context.Background(), "Test", tempFile)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
	if err.Error() != "file contains no user IDs" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestClient_CreateAudienceFromFile_FileNotFound(t *testing.T) {
	client := NewClient("test-token", false, false)

	_, err := client.CreateAudienceFromFile(context.Background(), "Test", "/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// Helper functions for temp file management

func createTempFileWithContent(content string) (string, error) {
	tmpfile, err := os.CreateTemp("", "test-audience-*.txt")
	if err != nil {
		return "", err
	}
	if _, err := tmpfile.WriteString(content); err != nil {
		_ = tmpfile.Close()
		_ = os.Remove(tmpfile.Name())
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		_ = os.Remove(tmpfile.Name())
		return "", err
	}
	return tmpfile.Name(), nil
}

func removeTempFile(path string) {
	_ = os.Remove(path)
}
