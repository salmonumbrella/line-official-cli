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

func TestMessageReplyCmd_Execute_TextMessage(t *testing.T) {
	var capturedPath string
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageReplyCmdWithClient(client)
	cmd.SetArgs([]string{"--token", "reply-token-123", "--text", "Thanks for your message!"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/v2/bot/message/reply" {
		t.Errorf("expected path /v2/bot/message/reply, got %s", capturedPath)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	if reqBody["replyToken"] != "reply-token-123" {
		t.Errorf("expected replyToken=reply-token-123, got %v", reqBody["replyToken"])
	}
	messages := reqBody["messages"].([]any)
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	msg := messages[0].(map[string]any)
	if msg["type"] != "text" {
		t.Errorf("expected type=text, got %v", msg["type"])
	}
	if msg["text"] != "Thanks for your message!" {
		t.Errorf("expected text='Thanks for your message!', got %v", msg["text"])
	}

	output := out.String()
	if !strings.Contains(output, "Reply sent") {
		t.Errorf("expected output to contain 'Reply sent', got %s", output)
	}
}

func TestMessageReplyCmd_Execute_FlexMessage(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	flexContent := `{"type":"bubble","body":{"type":"box","layout":"vertical","contents":[{"type":"text","text":"Hello"}]}}`

	cmd := newMessageReplyCmdWithClient(client)
	cmd.SetArgs([]string{"--token", "reply-token-456", "--flex", flexContent, "--alt-text", "Custom alt text"})

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
	messages := reqBody["messages"].([]any)
	msg := messages[0].(map[string]any)
	if msg["type"] != "flex" {
		t.Errorf("expected type=flex, got %v", msg["type"])
	}
	if msg["altText"] != "Custom alt text" {
		t.Errorf("expected altText='Custom alt text', got %v", msg["altText"])
	}
}

func TestMessageReplyCmd_Execute_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newMessageReplyCmdWithClient(client)
	cmd.SetArgs([]string{"--token", "reply-token", "--text", "Test"})

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
	if result["status"] != "sent" {
		t.Errorf("expected status=sent, got %v", result["status"])
	}
}

func TestMessageReplyCmd_MissingToken(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageReplyCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "Hello"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --token")
	}
}

func TestMessageReplyCmd_MissingMessage(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageReplyCmdWithClient(client)
	cmd.SetArgs([]string{"--token", "some-token"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing message content")
	}
	if !strings.Contains(err.Error(), "specify --text or --flex") {
		t.Errorf("expected error to contain 'specify --text or --flex', got %v", err)
	}
}

func TestMessageReplyCmd_BothTextAndFlex(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageReplyCmdWithClient(client)
	cmd.SetArgs([]string{"--token", "some-token", "--text", "Hello", "--flex", `{"type":"bubble"}`})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for specifying both --text and --flex")
	}
	if !strings.Contains(err.Error(), "not both") {
		t.Errorf("expected error to contain 'not both', got %v", err)
	}
}

func TestMessageReplyCmd_Execute_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Invalid reply token",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageReplyCmdWithClient(client)
	cmd.SetArgs([]string{"--token", "expired-token", "--text", "Hello"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to reply") {
		t.Errorf("expected error to contain 'failed to reply', got %v", err)
	}
}

func TestMessageReplyCmd_Execute_FlexAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Invalid flex message",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	flexContent := `{"type":"bubble","body":{"type":"box","layout":"vertical","contents":[]}}`
	cmd := newMessageReplyCmdWithClient(client)
	cmd.SetArgs([]string{"--token", "expired-token", "--flex", flexContent})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to reply") {
		t.Errorf("expected error to contain 'failed to reply', got %v", err)
	}
}
