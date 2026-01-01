package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/salmonumbrella/line-official-cli/internal/api"
)

func TestMessageNarrowcastCmd_Execute(t *testing.T) {
	var capturedPath string
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("X-Line-Request-Id", "test-request-id-123")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageNarrowcastCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "Special offer!"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/v2/bot/message/narrowcast" {
		t.Errorf("expected path /v2/bot/message/narrowcast, got %s", capturedPath)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	messages := reqBody["messages"].([]any)
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}
	msg := messages[0].(map[string]any)
	if msg["text"] != "Special offer!" {
		t.Errorf("expected text='Special offer!', got %v", msg["text"])
	}

	output := out.String()
	if !strings.Contains(output, "Narrowcast queued") {
		t.Errorf("expected output to contain 'Narrowcast queued', got %s", output)
	}
	if !strings.Contains(output, "test-request-id-123") {
		t.Errorf("expected output to contain request ID, got %s", output)
	}
}

func TestMessageNarrowcastCmd_Execute_WithAudience(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("X-Line-Request-Id", "test-request-id-456")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageNarrowcastCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "VIP offer!", "--audience", "12345678"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}

	recipient := reqBody["recipient"].(map[string]any)
	if recipient["type"] != "audience" {
		t.Errorf("expected recipient type=audience, got %v", recipient["type"])
	}
	if recipient["audienceGroupId"].(float64) != 12345678 {
		t.Errorf("expected audienceGroupId=12345678, got %v", recipient["audienceGroupId"])
	}
}

func TestMessageNarrowcastCmd_Execute_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Line-Request-Id", "json-request-id-789")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newMessageNarrowcastCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "Test message"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output: %v", err)
	}
	if result["requestId"] != "json-request-id-789" {
		t.Errorf("expected requestId=json-request-id-789, got %v", result["requestId"])
	}
}

func TestMessageNarrowcastStatusCmd_Execute(t *testing.T) {
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"phase":        "succeeded",
			"successCount": 150,
			"failureCount": 5,
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageNarrowcastStatusCmdWithClient(client)
	cmd.SetArgs([]string{"--request-id", "test-request-id-123"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedPath, "/v2/bot/message/progress/narrowcast") {
		t.Errorf("expected path to contain '/v2/bot/message/progress/narrowcast', got %s", capturedPath)
	}
	if !strings.Contains(capturedPath, "requestId=test-request-id-123") {
		t.Errorf("expected path to contain 'requestId=test-request-id-123', got %s", capturedPath)
	}

	output := out.String()
	if !strings.Contains(output, "succeeded") {
		t.Errorf("expected output to contain 'succeeded', got %s", output)
	}
	if !strings.Contains(output, "150") {
		t.Errorf("expected output to contain '150', got %s", output)
	}
}

func TestMessageNarrowcastStatusCmd_Execute_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"phase":        "waiting",
			"successCount": 0,
			"failureCount": 0,
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newMessageNarrowcastStatusCmdWithClient(client)
	cmd.SetArgs([]string{"--request-id", "test-id"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output: %v", err)
	}
	if result["phase"] != "waiting" {
		t.Errorf("expected phase=waiting, got %v", result["phase"])
	}
}

func TestMessageNarrowcastCmd_MissingText(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageNarrowcastCmdWithClient(client)
	cmd.SetArgs([]string{})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --text")
	}
}

func TestMessageNarrowcastStatusCmd_MissingRequestID(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageNarrowcastStatusCmdWithClient(client)
	cmd.SetArgs([]string{})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --request-id")
	}
}

func TestMessageNarrowcastCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Invalid request",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageNarrowcastCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "Test message"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to narrowcast") {
		t.Errorf("expected error to contain 'failed to narrowcast', got %v", err)
	}
}

func TestMessageNarrowcastStatusCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Request not found",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageNarrowcastStatusCmdWithClient(client)
	cmd.SetArgs([]string{"--request-id", "nonexistent-id"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to get progress") {
		t.Errorf("expected error to contain 'failed to get progress', got %v", err)
	}
}
