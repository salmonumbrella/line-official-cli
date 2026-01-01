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

func TestBotCmd_RequiresSubcommand(t *testing.T) {
	cmd := newBotCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBotProfileCmd_RequiresUser(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"bot", "profile"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --user flag")
	}
}

func TestBotLinkTokenCmd_RequiresUser(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"bot", "link-token"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --user flag")
	}
}

// Execution tests using mock servers

func TestBotInfoCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/info" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"userId":      "U123456789",
				"displayName": "Test Bot",
				"basicId":     "@testbot",
				"chatMode":    "chat",
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
			checkText: "Display Name: Test Bot",
		},
		{
			name:     "json output",
			output:   "json",
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore flags
			oldOutput := flags.Output
			flags.Output = tt.output
			defer func() { flags.Output = oldOutput }()

			cmd := newBotInfoCmdWithClient(client)
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
				if result["displayName"] != "Test Bot" {
					t.Errorf("expected displayName 'Test Bot', got: %v", result["displayName"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestBotInfoCmd_WithPremiumID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/info" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"userId":      "U123456789",
				"displayName": "Premium Bot",
				"basicId":     "@premiumbot",
				"premiumId":   "@premium123",
				"chatMode":    "chat",
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

	cmd := newBotInfoCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Premium ID:   @premium123") {
		t.Errorf("expected output to contain Premium ID, got: %s", output)
	}
	if !strings.Contains(output, "Display Name: Premium Bot") {
		t.Errorf("expected output to contain Display Name, got: %s", output)
	}
}

func TestBotInfoCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid token"})
	}))
	defer server.Close()

	client := api.NewClient("bad-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newBotInfoCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get bot info") {
		t.Errorf("expected 'failed to get bot info' in error, got: %v", err)
	}
}

func TestBotProfileCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/profile/U123456789" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"userId":        "U123456789",
				"displayName":   "Test User",
				"statusMessage": "Hello World",
				"language":      "en",
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
			checkText: "Display Name: Test User",
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

			cmd := newBotProfileCmdWithClient(client)
			cmd.SetArgs([]string{"--user", "U123456789"})
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
				if result["displayName"] != "Test User" {
					t.Errorf("expected displayName 'Test User', got: %v", result["displayName"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestBotProfileCmd_WithStatusAndLanguage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/profile/U123456789" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"userId":        "U123456789",
				"displayName":   "Test User",
				"statusMessage": "Hello World",
				"language":      "en",
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

	cmd := newBotProfileCmdWithClient(client)
	cmd.SetArgs([]string{"--user", "U123456789"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Status:       Hello World") {
		t.Errorf("expected output to contain Status, got: %s", output)
	}
	if !strings.Contains(output, "Language:     en") {
		t.Errorf("expected output to contain Language, got: %s", output)
	}
}

func TestBotProfileCmd_WithoutOptionalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/profile/U123456789" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"userId":      "U123456789",
				"displayName": "Minimal User",
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

	cmd := newBotProfileCmdWithClient(client)
	cmd.SetArgs([]string{"--user", "U123456789"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "Status:") {
		t.Errorf("expected output to NOT contain Status when empty, got: %s", output)
	}
	if strings.Contains(output, "Language:") {
		t.Errorf("expected output to NOT contain Language when empty, got: %s", output)
	}
	if !strings.Contains(output, "Display Name: Minimal User") {
		t.Errorf("expected output to contain Display Name, got: %s", output)
	}
}

func TestBotProfileCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "User not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newBotProfileCmdWithClient(client)
	cmd.SetArgs([]string{"--user", "U999999999"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get profile") {
		t.Errorf("expected 'failed to get profile' in error, got: %v", err)
	}
}

func TestBotFollowersCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/followers/ids") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"userIds": []string{"U111", "U222", "U333"},
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
			checkText: "Followers: 3",
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

			cmd := newBotFollowersCmdWithClient(client)
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
				count, ok := result["count"].(float64)
				if !ok || int(count) != 3 {
					t.Errorf("expected count 3, got: %v", result["count"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestBotFollowersCmd_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/followers/ids") {
			w.Header().Set("Content-Type", "application/json")
			callCount++
			if r.URL.Query().Get("start") == "" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"userIds": []string{"U111", "U222"},
					"next":    "page2token",
				})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"userIds": []string{"U333"},
				})
			}
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

	cmd := newBotFollowersCmdWithClient(client)
	cmd.SetArgs([]string{"--all"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got: %d", callCount)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output: %v", err)
	}
	count, ok := result["count"].(float64)
	if !ok || int(count) != 3 {
		t.Errorf("expected count 3, got: %v", result["count"])
	}
}

func TestBotFollowersCmd_WithoutAll_StopsAtFirstPage(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/followers/ids") {
			w.Header().Set("Content-Type", "application/json")
			callCount++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"userIds": []string{"U111", "U222"},
				"next":    "page2token",
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

	cmd := newBotFollowersCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 API call without --all, got: %d", callCount)
	}

	output := out.String()
	if !strings.Contains(output, "Followers: 2") {
		t.Errorf("expected 'Followers: 2', got: %s", output)
	}
}

func TestBotFollowersCmd_TextOutputListsIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/followers/ids") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"userIds": []string{"U111", "U222", "U333"},
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

	cmd := newBotFollowersCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "U111") {
		t.Errorf("expected output to contain U111, got: %s", output)
	}
	if !strings.Contains(output, "U222") {
		t.Errorf("expected output to contain U222, got: %s", output)
	}
	if !strings.Contains(output, "U333") {
		t.Errorf("expected output to contain U333, got: %s", output)
	}
}

func TestBotFollowersCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized"})
	}))
	defer server.Close()

	client := api.NewClient("bad-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newBotFollowersCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get followers") {
		t.Errorf("expected 'failed to get followers' in error, got: %v", err)
	}
}

func TestBotLinkTokenCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/user/U123456789/linkToken" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"linkToken": "abc123xyz",
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
			checkText: "Link Token: abc123xyz",
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

			cmd := newBotLinkTokenCmdWithClient(client)
			cmd.SetArgs([]string{"--user", "U123456789"})
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
				if result["linkToken"] != "abc123xyz" {
					t.Errorf("expected linkToken 'abc123xyz', got: %v", result["linkToken"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestBotLinkTokenCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid user"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newBotLinkTokenCmdWithClient(client)
	cmd.SetArgs([]string{"--user", "U999999999"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to issue link token") {
		t.Errorf("expected 'failed to issue link token' in error, got: %v", err)
	}
}
