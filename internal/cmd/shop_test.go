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

func TestShopCmd_RequiresSubcommand(t *testing.T) {
	cmd := newShopCmd()

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

func TestShopCmd_HasSubcommands(t *testing.T) {
	cmd := newShopCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 1 {
		t.Errorf("expected at least 1 subcommand (mission), got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	if !names["mission"] {
		t.Error("expected 'mission' subcommand")
	}
}

func TestShopMissionStickerCmd_RequiresTo(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"shop", "mission-sticker", "--product-id", "12345", "--product-type", "STICKER"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --to flag")
	}
}

func TestShopMissionStickerCmd_RequiresProductID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"shop", "mission-sticker", "--to", "U123", "--product-type", "STICKER"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --product-id flag")
	}
}

func TestShopMissionStickerCmd_RequiresProductType(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"shop", "mission-sticker", "--to", "U123", "--product-id", "12345"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --product-type flag")
	}
}

func TestShopMissionStickerCmd_Flags(t *testing.T) {
	cmd := newShopMissionCmd()

	// Check all required flags exist
	flagNames := []string{"to", "product-id", "product-type", "send-message"}
	for _, flagName := range flagNames {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("expected --%s flag", flagName)
		}
	}
}

// Execution tests using mock servers

func TestShopMissionCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/shop/v3/mission" && r.Method == http.MethodPost {
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
			checkText: "Mission sticker sent to user U123456789 (product: 12345)",
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

			cmd := newShopMissionCmdWithClient(client)
			cmd.SetArgs([]string{"--to", "U123456789", "--product-id", "12345", "--product-type", "STICKER"})
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
				if result["to"] != "U123456789" {
					t.Errorf("expected to 'U123456789', got: %v", result["to"])
				}
				if result["productId"] != "12345" {
					t.Errorf("expected productId '12345', got: %v", result["productId"])
				}
				if result["productType"] != "STICKER" {
					t.Errorf("expected productType 'STICKER', got: %v", result["productType"])
				}
				if result["status"] != "sent" {
					t.Errorf("expected status 'sent', got: %v", result["status"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestShopMissionCmd_WithSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/shop/v3/mission" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newShopMissionCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123456789", "--product-id", "12345", "--product-type", "STICKER", "--send-message"})
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
	if result["sendPresentMessage"] != true {
		t.Errorf("expected sendPresentMessage true, got: %v", result["sendPresentMessage"])
	}
}

func TestShopMissionCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newShopMissionCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123456789", "--product-id", "12345", "--product-type", "STICKER"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to send mission sticker") {
		t.Errorf("expected 'failed to send mission sticker' in error, got: %v", err)
	}
}

func TestShopMissionCmd_RequiresTo(t *testing.T) {
	cmd := newShopMissionCmd()
	cmd.SetArgs([]string{"--product-id", "12345", "--product-type", "STICKER"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --to flag")
	}
	if !strings.Contains(err.Error(), "--to is required") {
		t.Errorf("expected '--to is required' in error, got: %v", err)
	}
}

func TestShopMissionCmd_RequiresProductID(t *testing.T) {
	cmd := newShopMissionCmd()
	cmd.SetArgs([]string{"--to", "U123", "--product-type", "STICKER"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --product-id flag")
	}
	if !strings.Contains(err.Error(), "--product-id is required") {
		t.Errorf("expected '--product-id is required' in error, got: %v", err)
	}
}

func TestShopMissionCmd_RequiresProductType(t *testing.T) {
	cmd := newShopMissionCmd()
	cmd.SetArgs([]string{"--to", "U123", "--product-id", "12345"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --product-type flag")
	}
	if !strings.Contains(err.Error(), "--product-type is required") {
		t.Errorf("expected '--product-type is required' in error, got: %v", err)
	}
}
