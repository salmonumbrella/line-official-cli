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

func TestTokenCmd_RequiresSubcommand(t *testing.T) {
	cmd := newTokenCmd()

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

func TestTokenCmd_HasSubcommands(t *testing.T) {
	cmd := newTokenCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 4 {
		t.Errorf("expected at least 4 subcommands (issue, verify, revoke, issue-stateless), got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expected := []string{"issue", "verify", "revoke", "issue-stateless"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestTokenIssueCmd_RequiresClientID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"token", "issue", "--client-secret", "secret123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --client-id flag")
	}
}

func TestTokenIssueCmd_RequiresClientSecret(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"token", "issue", "--client-id", "123456"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --client-secret flag")
	}
}

func TestTokenVerifyCmd_RequiresToken(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"token", "verify"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --token flag")
	}
}

func TestTokenRevokeCmd_RequiresToken(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"token", "revoke"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --token flag")
	}
}

func TestTokenStatelessCmd_RequiresClientID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"token", "issue-stateless", "--client-secret", "secret123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --client-id flag")
	}
}

func TestTokenStatelessCmd_RequiresClientSecret(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"token", "issue-stateless", "--client-id", "123456"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --client-secret flag")
	}
}

func TestTokenStatelessCmd_Flags(t *testing.T) {
	cmd := newTokenIssueStatelessCmd()

	// Check all required flags exist
	flags := []string{"client-id", "client-secret"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("expected --%s flag", flagName)
		}
	}
}

// Execution tests with httptest mocks

func TestTokenIssueCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/oauth/accessToken" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token-123",
			"token_type":   "Bearer",
			"expires_in":   2592000,
			"key_id":       "key-abc",
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newTokenIssueCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--client-id", "123456", "--client-secret", "secret123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-token-123") {
		t.Errorf("expected output to contain access token, got: %s", output)
	}
	if !strings.Contains(output, "Bearer") {
		t.Errorf("expected output to contain token type, got: %s", output)
	}
}

func TestTokenIssueCmd_ExecuteJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token-json",
			"token_type":   "Bearer",
			"expires_in":   2592000,
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	// Set JSON output mode
	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newTokenIssueCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--client-id", "123456", "--client-secret", "secret123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp api.TokenResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if resp.AccessToken != "test-token-json" {
		t.Errorf("expected access_token 'test-token-json', got: %s", resp.AccessToken)
	}
}

func TestTokenVerifyCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/oauth/verify" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"client_id":  "1234567890",
			"expires_in": 2591000,
			"scope":      "profile",
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newTokenVerifyCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--token", "some-token"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1234567890") {
		t.Errorf("expected output to contain client ID, got: %s", output)
	}
	if !strings.Contains(output, "profile") {
		t.Errorf("expected output to contain scope, got: %s", output)
	}
}

func TestTokenVerifyCmd_ExecuteJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"client_id":  "client-json-test",
			"expires_in": 1000,
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newTokenVerifyCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--token", "verify-token"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp api.TokenInfo
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if resp.ClientID != "client-json-test" {
		t.Errorf("expected client_id 'client-json-test', got: %s", resp.ClientID)
	}
}

func TestTokenRevokeCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/oauth/revoke" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newTokenRevokeCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--token", "revoke-me"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "revoked successfully") {
		t.Errorf("expected success message, got: %s", output)
	}
}

func TestTokenRevokeCmd_ExecuteJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newTokenRevokeCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--token", "revoke-json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if resp["status"] != "revoked" {
		t.Errorf("expected status 'revoked', got: %v", resp["status"])
	}
}

func TestTokenIssueJWTCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/v2.1/token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "jwt-issued-token",
			"token_type":   "Bearer",
			"expires_in":   2592000,
			"key_id":       "jwt-key-id",
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newTokenIssueJWTCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--jwt", "eyJhbGciOiJSUzI1NiI..."})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "jwt-issued-token") {
		t.Errorf("expected output to contain access token, got: %s", output)
	}
	if !strings.Contains(output, "jwt-key-id") {
		t.Errorf("expected output to contain key ID, got: %s", output)
	}
}

func TestTokenIssueJWTCmd_RequiresJWT(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"token", "issue-jwt"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --jwt flag")
	}
}

func TestTokenVerifyJWTCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/oauth2/v2.1/verify") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"client_id":  "jwt-client-id",
			"expires_in": 86400,
			"scope":      "openid profile",
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newTokenVerifyJWTCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--token", "jwt-verify-token"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "jwt-client-id") {
		t.Errorf("expected output to contain client ID, got: %s", output)
	}
}

func TestTokenVerifyJWTCmd_RequiresToken(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"token", "verify-jwt"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --token flag")
	}
}

func TestTokenRevokeJWTCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/v2.1/revoke" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newTokenRevokeJWTCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--token", "jwt-revoke-token", "--client-id", "123", "--client-secret", "secret"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "revoked successfully") {
		t.Errorf("expected success message, got: %s", output)
	}
}

func TestTokenRevokeJWTCmd_RequiresAllFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing token", []string{"token", "revoke-jwt", "--client-id", "123", "--client-secret", "secret"}},
		{"missing client-id", []string{"token", "revoke-jwt", "--token", "tok", "--client-secret", "secret"}},
		{"missing client-secret", []string{"token", "revoke-jwt", "--token", "tok", "--client-id", "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestTokenListKeysCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/oauth2/v2.1/tokens/kid") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"kids": []string{"key-1", "key-2", "key-3"},
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newTokenListKeysCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--jwt", "jwt-assertion"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "key-1") {
		t.Errorf("expected output to contain key-1, got: %s", output)
	}
	if !strings.Contains(output, "key-2") {
		t.Errorf("expected output to contain key-2, got: %s", output)
	}
}

func TestTokenListKeysCmd_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"kids": []string{},
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newTokenListKeysCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--jwt", "jwt-assertion"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No valid token key IDs found") {
		t.Errorf("expected 'no keys found' message, got: %s", output)
	}
}

func TestTokenListKeysCmd_RequiresJWT(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"token", "list-keys"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --jwt flag")
	}
}

func TestTokenIssueStatelessCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/v3/token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "stateless-token-abc",
			"token_type":   "Bearer",
			"expires_in":   900,
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newTokenIssueStatelessCmdWithClient(client)
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"--client-id", "123456", "--client-secret", "secret123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "stateless-token-abc") {
		t.Errorf("expected output to contain access token, got: %s", output)
	}
	if !strings.Contains(output, "900") {
		t.Errorf("expected output to contain expires_in, got: %s", output)
	}

	// Check stderr for the warning note
	errOutput := errBuf.String()
	if !strings.Contains(errOutput, "Stateless tokens cannot be revoked") {
		t.Errorf("expected warning note in stderr, got: %s", errOutput)
	}
}

func TestTokenIssueStatelessCmd_ExecuteJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "stateless-json-token",
			"token_type":   "Bearer",
			"expires_in":   900,
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newTokenIssueStatelessCmdWithClient(client)
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"--client-id", "123456", "--client-secret", "secret123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp api.TokenResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if resp.AccessToken != "stateless-json-token" {
		t.Errorf("expected access_token 'stateless-json-token', got: %s", resp.AccessToken)
	}

	// In JSON mode, no warning should be printed
	errOutput := errBuf.String()
	if strings.Contains(errOutput, "Stateless tokens") {
		t.Errorf("expected no warning in JSON mode, got: %s", errOutput)
	}
}

func TestTokenCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":             "invalid_client",
			"error_description": "The client ID is invalid",
		})
	}))
	defer server.Close()

	client := api.NewClient("", false, false)
	client.SetBaseURL(server.URL)

	cmd := newTokenIssueCmdWithClient(client)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--client-id", "bad-id", "--client-secret", "secret"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid credentials")
	}
	if !strings.Contains(err.Error(), "failed to issue token") {
		t.Errorf("expected 'failed to issue token' error, got: %v", err)
	}
}
