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

func TestRoomCmd_RequiresSubcommand(t *testing.T) {
	cmd := newRoomCmd()

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

func TestRoomCmd_HasSubcommands(t *testing.T) {
	cmd := newRoomCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 3 {
		t.Errorf("expected at least 3 subcommands (members, profile, leave), got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expected := []string{"members", "profile", "leave"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestRoomMembersCmd_RequiresRoomID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"room", "members"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestRoomProfileCmd_RequiresRoomID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"room", "profile", "--user", "U123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestRoomProfileCmd_RequiresUserID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"room", "profile", "--id", "R123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --user flag")
	}
}

func TestRoomLeaveCmd_RequiresRoomID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"room", "leave"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestRoomLeaveCmd_HasIDFlag(t *testing.T) {
	cmd := newRoomLeaveCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag for leave command")
	}
}

// Execution tests using mock servers

func TestRoomMembersCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/members/count") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"count": 42})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/members/ids") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"memberIds": []string{"U001", "U002", "U003"},
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
		all       bool
		wantJSON  bool
		checkText string
	}{
		{
			name:      "text output count only",
			output:    "text",
			all:       false,
			wantJSON:  false,
			checkText: "Members:  42",
		},
		{
			name:      "text output with all members",
			output:    "text",
			all:       true,
			wantJSON:  false,
			checkText: "U001",
		},
		{
			name:     "json output",
			output:   "json",
			all:      false,
			wantJSON: true,
		},
		{
			name:     "json output with all members",
			output:   "json",
			all:      true,
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = tt.output
			defer func() { flags.Output = oldOutput }()

			cmd := newRoomMembersCmdWithClient(client)
			args := []string{"--id", "R123456789"}
			if tt.all {
				args = append(args, "--all")
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
				if result["count"].(float64) != 42 {
					t.Errorf("expected count 42, got: %v", result["count"])
				}
				if tt.all {
					if result["memberIds"] == nil {
						t.Error("expected memberIds in output")
					}
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestRoomMembersCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid room"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRoomMembersCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "R999999999"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get member count") {
		t.Errorf("expected 'failed to get member count' in error, got: %v", err)
	}
}

func TestRoomMembersCmd_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/members/count") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"count": 5})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/members/ids") {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			if callCount == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"memberIds": []string{"U001", "U002"},
					"next":      "cursor123",
				})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"memberIds": []string{"U003", "U004", "U005"},
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

	cmd := newRoomMembersCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "R123456789", "--all"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	memberIds := result["memberIds"].([]any)
	if len(memberIds) != 5 {
		t.Errorf("expected 5 member IDs after pagination, got: %d", len(memberIds))
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got: %d", callCount)
	}
}

func TestRoomMembersCmd_PaginationError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/members/count") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"count": 5})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/members/ids") {
			callCount++
			if callCount == 1 {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"memberIds": []string{"U001", "U002"},
					"next":      "cursor123",
				})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Server error"})
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRoomMembersCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "R123456789", "--all"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for pagination failure")
	}
	if !strings.Contains(err.Error(), "failed to get member IDs") {
		t.Errorf("expected 'failed to get member IDs' in error, got: %v", err)
	}
}

func TestRoomProfileCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/member/") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"userId":      "U123456789",
				"displayName": "Test User",
				"pictureUrl":  "https://example.com/pic.jpg",
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

			cmd := newRoomProfileCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "R123456789", "--user", "U123456789"})
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
				if result["userId"] != "U123456789" {
					t.Errorf("expected userId 'U123456789', got: %v", result["userId"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestRoomProfileCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "User not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRoomProfileCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "R123456789", "--user", "U999999999"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get member profile") {
		t.Errorf("expected 'failed to get member profile' in error, got: %v", err)
	}
}

func TestRoomProfileCmd_NoPictureURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/member/") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"userId":      "U123456789",
				"displayName": "Test User",
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

	cmd := newRoomProfileCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "R123456789", "--user", "U123456789"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "Picture URL") {
		t.Error("did not expect Picture URL in output when not provided")
	}
	if !strings.Contains(output, "Display Name: Test User") {
		t.Errorf("expected 'Display Name: Test User' in output, got: %s", output)
	}
}

func TestRoomLeaveCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/leave") && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
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
			checkText: "Left room: R123456789",
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
			flags.Yes = true
			defer func() {
				flags.Output = oldOutput
				flags.Yes = oldYes
			}()

			cmd := newRoomLeaveCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "R123456789"})
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
				if result["roomId"] != "R123456789" {
					t.Errorf("expected roomId 'R123456789', got: %v", result["roomId"])
				}
				if result["status"] != "left" {
					t.Errorf("expected status 'left', got: %v", result["status"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestRoomLeaveCmd_RequiresYesFlag(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	oldYes := flags.Yes
	flags.Yes = false
	defer func() { flags.Yes = oldYes }()

	cmd := newRoomLeaveCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "R123456789"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --yes flag is not set")
	}
	if !strings.Contains(err.Error(), "use --yes to confirm") {
		t.Errorf("expected 'use --yes to confirm' in error, got: %v", err)
	}
}

func TestRoomLeaveCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Cannot leave room"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldYes := flags.Yes
	flags.Yes = true
	defer func() { flags.Yes = oldYes }()

	cmd := newRoomLeaveCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "R999999999"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to leave room") {
		t.Errorf("expected 'failed to leave room' in error, got: %v", err)
	}
}
