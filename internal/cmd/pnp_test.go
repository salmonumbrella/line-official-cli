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

func TestPNPCmd_RequiresSubcommand(t *testing.T) {
	cmd := newPNPCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	// Running without subcommand should show help, not error
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPNPCmd_HasSubcommands(t *testing.T) {
	cmd := newPNPCmd()

	subcommands := cmd.Commands()
	if len(subcommands) != 1 {
		t.Errorf("expected 1 subcommand (push), got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expected := []string{"push"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestPNPPushCmd_RequiresFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing both flags",
			args:    []string{},
			wantErr: "--to",
		},
		{
			name:    "missing text flag",
			args:    []string{"--to", "+819012345678"},
			wantErr: "--text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(append([]string{"pnp", "push"}, tt.args...))

			err := cmd.Execute()
			if err == nil {
				t.Error("expected error for missing flags")
			}
		})
	}
}

func TestPNPPushCmd_HasRequiredFlags(t *testing.T) {
	cmd := newPNPPushCmd()

	requiredFlags := []string{"to", "text"}
	for _, flagName := range requiredFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("expected --%s flag for push command", flagName)
		}
	}
}

// Execution tests using mock servers

func TestPNPPushCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bot/pnp/push" && r.Method == http.MethodPost {
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
			checkText: "PNP message sent to +819012345678",
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

			cmd := newPNPPushCmdWithClient(client)
			cmd.SetArgs([]string{"--to", "+819012345678", "--text", "Hello from test!"})
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
				if result["status"] != "sent" {
					t.Errorf("expected status 'sent', got: %v", result["status"])
				}
				if result["to"] != "+819012345678" {
					t.Errorf("expected to '+819012345678', got: %v", result["to"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestPNPPushCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "PNP not enabled"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newPNPPushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "+819012345678", "--text", "Hello from test!"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to send PNP message") {
		t.Errorf("expected 'failed to send PNP message' in error, got: %v", err)
	}
}
