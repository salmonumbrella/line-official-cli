package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_PNPPushMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the path is /bot/pnp/push (not /v2/bot/...)
		if r.URL.Path != "/bot/pnp/push" {
			t.Errorf("unexpected path: %s, expected /bot/pnp/push", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", authHeader)
		}

		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", contentType)
		}

		var req PNPPushRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Verify phone number format
		if req.To != "+819012345678" {
			t.Errorf("expected to '+819012345678', got %s", req.To)
		}

		// Verify message content
		if len(req.Messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(req.Messages))
		}

		// Check message type and text
		msg, ok := req.Messages[0].(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", req.Messages[0])
		}
		if msg["type"] != "text" {
			t.Errorf("expected type 'text', got %v", msg["type"])
		}
		if msg["text"] != "Hello from PNP" {
			t.Errorf("expected text 'Hello from PNP', got %v", msg["text"])
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.PNPPushMessage(context.Background(), "+819012345678", "Hello from PNP")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_PNPPushMessage_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"PNP not enabled for this channel"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.PNPPushMessage(context.Background(), "+819012345678", "Hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_PNPPushMessage_InvalidPhoneNumber(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid phone number format"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.PNPPushMessage(context.Background(), "invalid", "Hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
