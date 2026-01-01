package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_SendMissionSticker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shop/v3/mission" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req["to"] != "U1234567890abcdef" {
			t.Errorf("expected to 'U1234567890abcdef', got %v", req["to"])
		}
		if req["productId"] != "12345" {
			t.Errorf("expected productId '12345', got %v", req["productId"])
		}
		if req["productType"] != "STICKER" {
			t.Errorf("expected productType 'STICKER', got %v", req["productType"])
		}
		if req["sendPresentMessage"] != true {
			t.Errorf("expected sendPresentMessage true, got %v", req["sendPresentMessage"])
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.SendMissionSticker(context.Background(), "U1234567890abcdef", "12345", "STICKER", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SendMissionSticker_WithoutMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		// When sendPresentMessage is false, it's omitted from JSON due to omitempty tag
		// So we check it's either false or nil (not present)
		val, exists := req["sendPresentMessage"]
		if exists && val != false {
			t.Errorf("expected sendPresentMessage false or omitted, got %v", val)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.SendMissionSticker(context.Background(), "U1234567890abcdef", "12345", "STICKER", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SendMissionSticker_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid product ID"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.SendMissionSticker(context.Background(), "U1234567890abcdef", "invalid", "STICKER", false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SendMissionSticker_UserNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"User not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.SendMissionSticker(context.Background(), "Uinvalid", "12345", "STICKER", false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
