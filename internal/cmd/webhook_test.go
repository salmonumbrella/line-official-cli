package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/salmonumbrella/line-official-cli/internal/api"
)

func TestWebhookCmd_RequiresSubcommand(t *testing.T) {
	cmd := newWebhookCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWebhookCmd_HasSubcommands(t *testing.T) {
	cmd := newWebhookCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 3 {
		t.Errorf("expected at least 3 subcommands (get, set, test), got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	for _, name := range []string{"get", "set", "test"} {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestWebhookSetCmd_RequiresUrl(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"webhook", "set"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --url flag")
	}
}

// Execution tests using mock servers

func TestWebhookGetCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/channel/webhook/endpoint" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"endpoint": "https://example.com/webhook",
				"active":   true,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	tests := []struct {
		name      string
		output    string
		wantJSON  bool
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			wantJSON:  false,
			checkText: "Endpoint: https://example.com/webhook",
		},
		{
			name:     "json output",
			output:   "json",
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = tt.output
			defer func() { flags.Output = oldOutput }()

			cmd := newWebhookGetCmdWithClient(client)
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := out.String()
			if tt.wantJSON {
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("expected valid JSON output, got: %s", output)
				}
				if result["endpoint"] != "https://example.com/webhook" {
					t.Errorf("expected endpoint 'https://example.com/webhook', got: %v", result["endpoint"])
				}
				if result["active"] != true {
					t.Errorf("expected active true, got: %v", result["active"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
				if !strings.Contains(output, "Active:   true") {
					t.Errorf("expected output to contain 'Active:   true', got: %s", output)
				}
			}
		})
	}
}

func TestWebhookGetCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid token"})
	}))
	defer server.Close()

	client := api.NewClient("bad-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newWebhookGetCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get webhook") {
		t.Errorf("expected 'failed to get webhook' in error, got: %v", err)
	}
}

func TestWebhookSetCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/channel/webhook/endpoint" && r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	tests := []struct {
		name      string
		output    string
		wantJSON  bool
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			wantJSON:  false,
			checkText: "Webhook set to: https://example.com/new-webhook",
		},
		{
			name:     "json output",
			output:   "json",
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = tt.output
			defer func() { flags.Output = oldOutput }()

			cmd := newWebhookSetCmdWithClient(client)
			cmd.SetArgs([]string{"--url", "https://example.com/new-webhook"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := out.String()
			if tt.wantJSON {
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("expected valid JSON output, got: %s", output)
				}
				if result["endpoint"] != "https://example.com/new-webhook" {
					t.Errorf("expected endpoint 'https://example.com/new-webhook', got: %v", result["endpoint"])
				}
				if result["status"] != "set" {
					t.Errorf("expected status 'set', got: %v", result["status"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestWebhookSetCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid URL"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newWebhookSetCmdWithClient(client)
	cmd.SetArgs([]string{"--url", "invalid-url"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to set webhook") {
		t.Errorf("expected 'failed to set webhook' in error, got: %v", err)
	}
}

func TestWebhookTestCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/channel/webhook/test" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success":    true,
				"timestamp":  "2024-01-01T00:00:00Z",
				"statusCode": 200,
				"reason":     "OK",
				"detail":     "",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	tests := []struct {
		name      string
		output    string
		wantJSON  bool
		checkText string
	}{
		{
			name:      "text output success",
			output:    "text",
			wantJSON:  false,
			checkText: "Webhook test: SUCCESS",
		},
		{
			name:     "json output",
			output:   "json",
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = tt.output
			defer func() { flags.Output = oldOutput }()

			cmd := newWebhookTestCmdWithClient(client)
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := out.String()
			if tt.wantJSON {
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("expected valid JSON output, got: %s", output)
				}
				if result["success"] != true {
					t.Errorf("expected success true, got: %v", result["success"])
				}
				if result["statusCode"] != float64(200) {
					t.Errorf("expected statusCode 200, got: %v", result["statusCode"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
				if !strings.Contains(output, "Status:    200 OK") {
					t.Errorf("expected output to contain 'Status:    200 OK', got: %s", output)
				}
			}
		})
	}
}

func TestWebhookTestCmd_WithUrl(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/channel/webhook/test" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success":    true,
				"timestamp":  "2024-01-01T00:00:00Z",
				"statusCode": 200,
				"reason":     "OK",
				"detail":     "",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newWebhookTestCmdWithClient(client)
	cmd.SetArgs([]string{"--url", "https://example.com/test-webhook"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Webhook test: SUCCESS") {
		t.Errorf("expected output to contain 'Webhook test: SUCCESS', got: %s", output)
	}
}

func TestWebhookTestCmd_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/channel/webhook/test" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success":    false,
				"timestamp":  "2024-01-01T00:00:00Z",
				"statusCode": 500,
				"reason":     "Internal Server Error",
				"detail":     "Connection refused",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newWebhookTestCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Webhook test: FAILED") {
		t.Errorf("expected output to contain 'Webhook test: FAILED', got: %s", output)
	}
	if !strings.Contains(output, "Detail:    Connection refused") {
		t.Errorf("expected output to contain 'Detail:    Connection refused', got: %s", output)
	}
}

func TestWebhookTestCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid token"})
	}))
	defer server.Close()

	client := api.NewClient("bad-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newWebhookTestCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to test webhook") {
		t.Errorf("expected 'failed to test webhook' in error, got: %v", err)
	}
}
