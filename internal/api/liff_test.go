package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetAllLIFFApps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/liff/v1/apps" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"apps":[{"liffId":"1234567890-abcdefgh","view":{"type":"full","url":"https://example.com/liff"},"description":"Test App"}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	apps, err := client.GetAllLIFFApps(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 1 {
		t.Errorf("expected 1 app, got %d", len(apps))
	}
	if apps[0].LIFFID != "1234567890-abcdefgh" {
		t.Errorf("expected liffId '1234567890-abcdefgh', got %s", apps[0].LIFFID)
	}
	if apps[0].View.Type != "full" {
		t.Errorf("expected view type 'full', got %s", apps[0].View.Type)
	}
	if apps[0].View.URL != "https://example.com/liff" {
		t.Errorf("expected URL 'https://example.com/liff', got %s", apps[0].View.URL)
	}
	if apps[0].Description != "Test App" {
		t.Errorf("expected description 'Test App', got %s", apps[0].Description)
	}
}

func TestClient_GetAllLIFFApps_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"apps":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	apps, err := client.GetAllLIFFApps(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 0 {
		t.Errorf("expected 0 apps, got %d", len(apps))
	}
}

func TestClient_AddLIFFApp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/liff/v1/apps" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req AddLIFFAppRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.View.Type != "compact" {
			t.Errorf("expected view type 'compact', got %s", req.View.Type)
		}
		if req.View.URL != "https://example.com/liff" {
			t.Errorf("expected URL 'https://example.com/liff', got %s", req.View.URL)
		}
		if req.Description != "My LIFF App" {
			t.Errorf("expected description 'My LIFF App', got %s", req.Description)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"liffId":"1234567890-newliff"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	req := &AddLIFFAppRequest{
		View: LIFFView{
			Type: "compact",
			URL:  "https://example.com/liff",
		},
		Description: "My LIFF App",
	}

	liffID, err := client.AddLIFFApp(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if liffID != "1234567890-newliff" {
		t.Errorf("expected liffId '1234567890-newliff', got %s", liffID)
	}
}

func TestClient_UpdateLIFFApp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/liff/v1/apps/1234567890-abcdefgh" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var req UpdateLIFFAppRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.View.Type != "tall" {
			t.Errorf("expected view type 'tall', got %s", req.View.Type)
		}
		if req.View.URL != "https://example.com/new-liff" {
			t.Errorf("expected URL 'https://example.com/new-liff', got %s", req.View.URL)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	req := &UpdateLIFFAppRequest{
		View: LIFFView{
			Type: "tall",
			URL:  "https://example.com/new-liff",
		},
		Description: "Updated LIFF App",
	}

	err := client.UpdateLIFFApp(context.Background(), "1234567890-abcdefgh", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteLIFFApp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/liff/v1/apps/1234567890-abcdefgh" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.DeleteLIFFApp(context.Background(), "1234567890-abcdefgh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_AddLIFFApp_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid view URL"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	req := &AddLIFFAppRequest{
		View: LIFFView{
			Type: "compact",
			URL:  "invalid-url",
		},
	}

	_, err := client.AddLIFFApp(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
