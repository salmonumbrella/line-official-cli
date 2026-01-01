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

func TestModuleCmd_RequiresSubcommand(t *testing.T) {
	cmd := newModuleCmd()

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

func TestModuleCmd_HasSubcommands(t *testing.T) {
	cmd := newModuleCmd()

	subcommands := cmd.Commands()
	if len(subcommands) != 5 {
		t.Errorf("expected 5 subcommands (detach, acquire, release, token, bots), got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expected := []string{"detach", "acquire", "release", "token", "bots"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestModuleDetachCmd_RequiresBotID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"module", "detach"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --bot-id flag")
	}
}

func TestModuleAcquireCmd_RequiresChatID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"module", "acquire"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --chat flag")
	}
}

func TestModuleReleaseCmd_RequiresChatID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"module", "release"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --chat flag")
	}
}

func TestModuleDetachCmd_HasBotIDFlag(t *testing.T) {
	cmd := newModuleDetachCmd()

	// Check --bot-id flag exists (--yes is a global flag from rootFlags)
	botIDFlag := cmd.Flags().Lookup("bot-id")
	if botIDFlag == nil {
		t.Error("expected --bot-id flag for detach command")
	}
}

func TestModuleAcquireCmd_HasNoExpiryFlag(t *testing.T) {
	cmd := newModuleAcquireCmd()

	// Check --no-expiry flag exists
	noExpiryFlag := cmd.Flags().Lookup("no-expiry")
	if noExpiryFlag == nil {
		t.Error("expected --no-expiry flag for acquire command")
	}
}

func TestModuleTokenCmd_RequiresAllFlags(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"module", "token"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing required flags")
	}
}

func TestModuleTokenCmd_HasRequiredFlags(t *testing.T) {
	cmd := newModuleTokenCmd()

	requiredFlags := []string{"code", "redirect-uri", "client-id", "client-secret"}
	for _, flagName := range requiredFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("expected --%s flag for token command", flagName)
		}
	}
}

// Execution tests using mock servers

func TestModuleDetachCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/channel/detach" && r.Method == http.MethodPost {
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
			checkText: "Module detached from bot U123456789",
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
			oldYes := flags.Yes
			flags.Output = tt.output
			flags.Yes = true // Skip confirmation
			defer func() {
				flags.Output = oldOutput
				flags.Yes = oldYes
			}()

			cmd := newModuleDetachCmdWithClient(client)
			cmd.SetArgs([]string{"--bot-id", "U123456789"})
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
				if result["botId"] != "U123456789" {
					t.Errorf("expected botId 'U123456789', got: %v", result["botId"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestModuleDetachCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid token"})
	}))
	defer server.Close()

	client := api.NewClient("bad-token", false, false)
	client.SetBaseURL(server.URL)

	oldYes := flags.Yes
	flags.Yes = true
	defer func() { flags.Yes = oldYes }()

	cmd := newModuleDetachCmdWithClient(client)
	cmd.SetArgs([]string{"--bot-id", "U123456789"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to detach module") {
		t.Errorf("expected 'failed to detach module' in error, got: %v", err)
	}
}

func TestModuleAcquireCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/chat/") && strings.HasSuffix(r.URL.Path, "/control/acquire") && r.Method == http.MethodPost {
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
		noExpiry  bool
		wantJSON  bool
		checkText string
	}{
		{
			name:      "text output with expiry",
			output:    "text",
			noExpiry:  false,
			wantJSON:  false,
			checkText: "Chat control acquired for U123456789",
		},
		{
			name:      "text output no expiry",
			output:    "text",
			noExpiry:  true,
			wantJSON:  false,
			checkText: "Chat control acquired for U123456789 (no expiry)",
		},
		{
			name:     "json output",
			output:   "json",
			noExpiry: false,
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = tt.output
			defer func() { flags.Output = oldOutput }()

			cmd := newModuleAcquireCmdWithClient(client)
			args := []string{"--chat", "U123456789"}
			if tt.noExpiry {
				args = append(args, "--no-expiry")
			}
			cmd.SetArgs(args)
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
				if result["chatId"] != "U123456789" {
					t.Errorf("expected chatId 'U123456789', got: %v", result["chatId"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestModuleAcquireCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not allowed"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newModuleAcquireCmdWithClient(client)
	cmd.SetArgs([]string{"--chat", "U123456789"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to acquire chat control") {
		t.Errorf("expected 'failed to acquire chat control' in error, got: %v", err)
	}
}

func TestModuleReleaseCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/chat/") && strings.HasSuffix(r.URL.Path, "/control/release") && r.Method == http.MethodPost {
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
			checkText: "Chat control released for U123456789",
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

			cmd := newModuleReleaseCmdWithClient(client)
			cmd.SetArgs([]string{"--chat", "U123456789"})
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
				if result["chatId"] != "U123456789" {
					t.Errorf("expected chatId 'U123456789', got: %v", result["chatId"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestModuleReleaseCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not allowed"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newModuleReleaseCmdWithClient(client)
	cmd.SetArgs([]string{"--chat", "U123456789"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to release chat control") {
		t.Errorf("expected 'failed to release chat control' in error, got: %v", err)
	}
}

func TestModuleTokenCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/module/auth/v1/token" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "test-access-token",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"refresh_token": "test-refresh-token",
				"scope":         "openid profile",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
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
			checkText: "Access Token: test-access-token",
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

			cmd := newModuleTokenCmdWithClient(client)
			cmd.SetArgs([]string{
				"--code", "test-auth-code",
				"--redirect-uri", "https://example.com/callback",
				"--client-id", "1234567890",
				"--client-secret", "secret123",
			})
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
				if result["access_token"] != "test-access-token" {
					t.Errorf("expected access_token 'test-access-token', got: %v", result["access_token"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestModuleTokenCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newModuleTokenCmdWithClient(client)
	cmd.SetArgs([]string{
		"--code", "bad-code",
		"--redirect-uri", "https://example.com/callback",
		"--client-id", "1234567890",
		"--client-secret", "secret123",
	})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to exchange token") {
		t.Errorf("expected 'failed to exchange token' in error, got: %v", err)
	}
}

func TestModuleBotsCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/list" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"bots": []map[string]string{
					{
						"userId":      "U111111111",
						"basicId":     "@testbot1",
						"displayName": "Test Bot 1",
					},
					{
						"userId":      "U222222222",
						"basicId":     "@testbot2",
						"displayName": "Test Bot 2",
					},
				},
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
			checkText: "Display Name: Test Bot 1",
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

			cmd := newModuleBotsCmdWithClient(client)
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
				bots, ok := result["bots"].([]any)
				if !ok || len(bots) != 2 {
					t.Errorf("expected 2 bots, got: %v", result["bots"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestModuleBotsCmd_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/list" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"bots": []map[string]string{},
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

	cmd := newModuleBotsCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No bots with attached modules found") {
		t.Errorf("expected 'No bots with attached modules found' message, got: %s", output)
	}
}

func TestModuleBotsCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid token"})
	}))
	defer server.Close()

	client := api.NewClient("bad-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newModuleBotsCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to list bots") {
		t.Errorf("expected 'failed to list bots' in error, got: %v", err)
	}
}
