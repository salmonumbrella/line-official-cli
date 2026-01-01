package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetWebhookEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/channel/webhook/endpoint" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"endpoint":"https://example.com/webhook","active":true}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	info, err := client.GetWebhookEndpoint(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Endpoint != "https://example.com/webhook" {
		t.Errorf("expected endpoint 'https://example.com/webhook', got %s", info.Endpoint)
	}
	if !info.Active {
		t.Errorf("expected active true, got false")
	}
}

func TestClient_GetWebhookEndpoint_Inactive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"endpoint":"https://example.com/webhook","active":false}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	info, err := client.GetWebhookEndpoint(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Active {
		t.Errorf("expected active false, got true")
	}
}

func TestClient_GetWebhookEndpoint_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Invalid channel access token"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetWebhookEndpoint(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SetWebhookEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/channel/webhook/endpoint" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var req SetWebhookEndpointRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Endpoint != "https://example.com/new-webhook" {
			t.Errorf("expected endpoint 'https://example.com/new-webhook', got %s", req.Endpoint)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.SetWebhookEndpoint(context.Background(), "https://example.com/new-webhook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SetWebhookEndpoint_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid webhook URL"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.SetWebhookEndpoint(context.Background(), "invalid-url")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_TestWebhookEndpoint_WithEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/channel/webhook/test" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req["endpoint"] != "https://example.com/test-webhook" {
			t.Errorf("expected endpoint 'https://example.com/test-webhook', got %s", req["endpoint"])
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"timestamp":"2024-01-15T10:30:00.000Z","statusCode":200,"reason":"OK","detail":"Webhook test successful"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.TestWebhookEndpoint(context.Background(), "https://example.com/test-webhook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Timestamp != "2024-01-15T10:30:00.000Z" {
		t.Errorf("expected timestamp '2024-01-15T10:30:00.000Z', got %s", resp.Timestamp)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected statusCode 200, got %d", resp.StatusCode)
	}
	if resp.Reason != "OK" {
		t.Errorf("expected reason 'OK', got %s", resp.Reason)
	}
	if resp.Detail != "Webhook test successful" {
		t.Errorf("expected detail 'Webhook test successful', got %s", resp.Detail)
	}
}

func TestClient_TestWebhookEndpoint_WithoutEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/channel/webhook/test" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// When endpoint is empty, request body should be null
		var req interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req != nil {
			t.Errorf("expected nil request body when no endpoint provided, got %v", req)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"timestamp":"2024-01-15T10:30:00.000Z","statusCode":200,"reason":"OK","detail":"Webhook test successful"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.TestWebhookEndpoint(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
}

func TestClient_TestWebhookEndpoint_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":false,"timestamp":"2024-01-15T10:30:00.000Z","statusCode":500,"reason":"Internal Server Error","detail":"Webhook endpoint returned error"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.TestWebhookEndpoint(context.Background(), "https://example.com/failing-webhook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Errorf("expected success false, got true")
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected statusCode 500, got %d", resp.StatusCode)
	}
	if resp.Reason != "Internal Server Error" {
		t.Errorf("expected reason 'Internal Server Error', got %s", resp.Reason)
	}
}

func TestClient_TestWebhookEndpoint_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Invalid channel access token"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.TestWebhookEndpoint(context.Background(), "https://example.com/webhook")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
