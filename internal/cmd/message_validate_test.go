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

func TestMessageValidateCmd_Execute_Success(t *testing.T) {
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

	cmd := newMessageValidateCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "push", "--messages", `[{"type":"text","text":"Hello"}]`})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/v2/bot/message/validate/push" {
		t.Errorf("expected path /v2/bot/message/validate/push, got %s", capturedPath)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	messages := reqBody["messages"].([]any)
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	output := out.String()
	if !strings.Contains(output, "Validation passed") {
		t.Errorf("expected output to contain 'Validation passed', got %s", output)
	}
}

func TestMessageValidateCmd_Execute_JSONOutput(t *testing.T) {
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

	cmd := newMessageValidateCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "broadcast", "--messages", `[{"type":"text","text":"Test"}]`})

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
	if result["valid"] != true {
		t.Errorf("expected valid=true, got %v", result["valid"])
	}
	if result["type"] != "broadcast" {
		t.Errorf("expected type=broadcast, got %v", result["type"])
	}
	if result["messageCount"].(float64) != 1 {
		t.Errorf("expected messageCount=1, got %v", result["messageCount"])
	}
}

func TestMessageValidateCmd_Execute_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Invalid message format",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageValidateCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "push", "--messages", `[{"type":"invalid"}]`})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for validation failure")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected error to contain 'validation failed', got %v", err)
	}
}

func TestMessageValidateCmd_Execute_AllTypes(t *testing.T) {
	messageTypes := []string{"reply", "push", "multicast", "narrowcast", "broadcast"}

	for _, msgType := range messageTypes {
		t.Run(msgType, func(t *testing.T) {
			var capturedPath string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{}"))
			}))
			defer server.Close()

			client := api.NewClient("test-token", false, false)
			client.SetBaseURL(server.URL)

			cmd := newMessageValidateCmdWithClient(client)
			cmd.SetArgs([]string{"--type", msgType, "--messages", `[{"type":"text","text":"Test"}]`})

			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error for type %s: %v", msgType, err)
			}

			expectedPath := "/v2/bot/message/validate/" + msgType
			if capturedPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, capturedPath)
			}
		})
	}
}

func TestMessageValidateCmd_Execute_InvalidType(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageValidateCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "invalid", "--messages", `[{"type":"text","text":"Test"}]`})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "must be one of") {
		t.Errorf("expected error to contain 'must be one of', got %v", err)
	}
}

func TestMessageValidateCmd_Execute_InvalidJSON(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageValidateCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "push", "--messages", `not valid json`})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid messages JSON") {
		t.Errorf("expected error to contain 'invalid messages JSON', got %v", err)
	}
}

func TestMessageValidateCmd_Execute_MissingInput(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageValidateCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "push"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing input")
	}
	if !strings.Contains(err.Error(), "--messages or --file is required") {
		t.Errorf("expected error to contain '--messages or --file is required', got %v", err)
	}
}

func TestMessageValidateCmd_Execute_BothMessagesAndFile(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageValidateCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "push", "--messages", `[{"type":"text","text":"Hi"}]`, "--file", "messages.json"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for specifying both --messages and --file")
	}
	if !strings.Contains(err.Error(), "not both") {
		t.Errorf("expected error to contain 'not both', got %v", err)
	}
}

func TestMessageValidateCmd_Execute_FileNotFound(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageValidateCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "push", "--file", "/nonexistent/path/messages.json"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for file not found")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("expected error to contain 'failed to read file', got %v", err)
	}
}

func TestMessageValidateCmd_Execute_ValidationError_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Invalid message format",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newMessageValidateCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "push", "--messages", `[{"type":"invalid"}]`})

	var out bytes.Buffer
	cmd.SetOut(&out)

	// In JSON output mode, validation errors are encoded as JSON and returned without error
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output: %v", err)
	}
	if result["valid"] != false {
		t.Errorf("expected valid=false, got %v", result["valid"])
	}
	if result["error"] == nil {
		t.Error("expected error field in JSON output")
	}
}
