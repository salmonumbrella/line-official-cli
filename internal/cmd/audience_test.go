package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salmonumbrella/line-official-cli/internal/api"
)

func TestAudienceCmd_HasSubcommands(t *testing.T) {
	cmd := newAudienceCmd()

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("expected audience to have subcommands")
	}

	// Verify key subcommands exist
	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expectedSubcommands := []string{
		"list", "get", "delete", "create", "add-users",
		"create-click", "create-impression", "update-description", "shared",
	}

	for _, expected := range expectedSubcommands {
		if !names[expected] {
			t.Errorf("expected '%s' subcommand", expected)
		}
	}
}

func TestAudienceCmd_Aliases(t *testing.T) {
	cmd := newAudienceCmd()

	if len(cmd.Aliases) == 0 {
		t.Error("expected audience to have aliases")
	}

	hasAud := false
	for _, alias := range cmd.Aliases {
		if alias == "aud" {
			hasAud = true
			break
		}
	}

	if !hasAud {
		t.Error("expected 'aud' alias for audience command")
	}
}

func TestAudienceListCmd_NoRequiredFlags(t *testing.T) {
	cmd := newAudienceListCmd()

	// List command should have no required flags
	if cmd.Use != "list" {
		t.Errorf("expected Use to be 'list', got %s", cmd.Use)
	}
}

func TestAudienceGetCmd_Flags(t *testing.T) {
	cmd := newAudienceGetCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}
}

func TestAudienceGetCmd_RequiresID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "get"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestAudienceGetCmd_ValidatesIDPositive(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "get", "--id", "0"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-positive ID")
	}
}

func TestAudienceDeleteCmd_Flags(t *testing.T) {
	cmd := newAudienceDeleteCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}
}

func TestAudienceDeleteCmd_RequiresID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "delete"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestAudienceDeleteCmd_ValidatesIDPositive(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "delete", "--id", "-1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for negative ID")
	}
}

func TestAudienceCreateCmd_Flags(t *testing.T) {
	cmd := newAudienceCreateCmd()

	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("expected --name flag")
	}

	usersFlag := cmd.Flags().Lookup("users")
	if usersFlag == nil {
		t.Error("expected --users flag")
	}

	fileFlag := cmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Error("expected --file flag")
	}
}

func TestAudienceCreateCmd_RequiresName(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "create", "--users", "U123,U456"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --name flag")
	}
}

func TestAudienceCreateCmd_RequiresUsersOrFile(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "create", "--name", "Test Audience"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --users or --file flag")
	}
}

func TestAudienceAddUsersCmd_Flags(t *testing.T) {
	cmd := newAudienceAddUsersCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}

	usersFlag := cmd.Flags().Lookup("users")
	if usersFlag == nil {
		t.Error("expected --users flag")
	}

	fileFlag := cmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Error("expected --file flag")
	}

	descFlag := cmd.Flags().Lookup("description")
	if descFlag == nil {
		t.Error("expected --description flag")
	}
}

func TestAudienceAddUsersCmd_RequiresID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "add-users", "--users", "U123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestAudienceAddUsersCmd_RequiresUsersOrFile(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "add-users", "--id", "12345"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --users or --file flag")
	}
}

func TestAudienceAddUsersCmd_ValidatesIDPositive(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "add-users", "--id", "0", "--users", "U123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-positive ID")
	}
}

func TestAudienceCreateClickCmd_Flags(t *testing.T) {
	cmd := newAudienceCreateClickCmd()

	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("expected --name flag")
	}

	requestFlag := cmd.Flags().Lookup("request")
	if requestFlag == nil {
		t.Error("expected --request flag")
	}
}

func TestAudienceCreateClickCmd_RequiresName(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "create-click", "--request", "req-123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --name flag")
	}
}

func TestAudienceCreateClickCmd_RequiresRequest(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "create-click", "--name", "Click Audience"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --request flag")
	}
}

func TestAudienceCreateImpressionCmd_Flags(t *testing.T) {
	cmd := newAudienceCreateImpressionCmd()

	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("expected --name flag")
	}

	requestFlag := cmd.Flags().Lookup("request")
	if requestFlag == nil {
		t.Error("expected --request flag")
	}
}

func TestAudienceCreateImpressionCmd_RequiresName(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "create-impression", "--request", "req-456"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --name flag")
	}
}

func TestAudienceCreateImpressionCmd_RequiresRequest(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "create-impression", "--name", "Impression Audience"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --request flag")
	}
}

func TestAudienceUpdateDescriptionCmd_Flags(t *testing.T) {
	cmd := newAudienceUpdateDescriptionCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}

	descFlag := cmd.Flags().Lookup("description")
	if descFlag == nil {
		t.Error("expected --description flag")
	}
}

func TestAudienceUpdateDescriptionCmd_RequiresID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "update-description", "--description", "New Description"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestAudienceUpdateDescriptionCmd_RequiresDescription(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "update-description", "--id", "12345"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --description flag")
	}
}

func TestAudienceUpdateDescriptionCmd_ValidatesIDPositive(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "update-description", "--id", "0", "--description", "Test"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-positive ID")
	}
}

func TestAudienceSharedCmd_HasSubcommands(t *testing.T) {
	cmd := newAudienceSharedCmd()

	subcommands := cmd.Commands()
	if len(subcommands) != 2 {
		t.Errorf("expected 2 shared subcommands, got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	if !names["list"] {
		t.Error("expected 'list' subcommand")
	}
	if !names["get"] {
		t.Error("expected 'get' subcommand")
	}
}

func TestAudienceSharedListCmd_NoRequiredFlags(t *testing.T) {
	cmd := newAudienceSharedListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use to be 'list', got %s", cmd.Use)
	}
}

func TestAudienceSharedGetCmd_Flags(t *testing.T) {
	cmd := newAudienceSharedGetCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}
}

func TestAudienceSharedGetCmd_RequiresID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "shared", "get"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestAudienceSharedGetCmd_ValidatesIDPositive(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"audience", "shared", "get", "--id", "-5"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for negative ID")
	}
}

// Table-driven test for required flag validation
func TestAudienceCommands_RequiredFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectsError bool
	}{
		// get requires --id
		{"get without id", []string{"audience", "get"}, true},
		// delete requires --id
		{"delete without id", []string{"audience", "delete"}, true},
		// create requires --name
		{"create without name", []string{"audience", "create"}, true},
		// add-users requires --id
		{"add-users without id", []string{"audience", "add-users"}, true},
		// create-click requires --name and --request
		{"create-click without flags", []string{"audience", "create-click"}, true},
		{"create-click without name", []string{"audience", "create-click", "--request", "req-123"}, true},
		{"create-click without request", []string{"audience", "create-click", "--name", "Test"}, true},
		// create-impression requires --name and --request
		{"create-impression without flags", []string{"audience", "create-impression"}, true},
		{"create-impression without name", []string{"audience", "create-impression", "--request", "req-456"}, true},
		{"create-impression without request", []string{"audience", "create-impression", "--name", "Test"}, true},
		// update-description requires --id and --description
		{"update-description without flags", []string{"audience", "update-description"}, true},
		{"update-description without id", []string{"audience", "update-description", "--description", "Test"}, true},
		{"update-description without description", []string{"audience", "update-description", "--id", "12345"}, true},
		// shared get requires --id
		{"shared get without id", []string{"audience", "shared", "get"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.expectsError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectsError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Test ID validation
func TestAudienceCommands_IDValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"get with zero id", []string{"audience", "get", "--id", "0"}},
		{"get with negative id", []string{"audience", "get", "--id", "-1"}},
		{"delete with zero id", []string{"audience", "delete", "--id", "0"}},
		{"delete with negative id", []string{"audience", "delete", "--id", "-1"}},
		{"add-users with zero id", []string{"audience", "add-users", "--id", "0", "--users", "U123"}},
		{"update-description with zero id", []string{"audience", "update-description", "--id", "0", "--description", "Test"}},
		{"shared get with zero id", []string{"audience", "shared", "get", "--id", "0"}},
		{"shared get with negative id", []string{"audience", "shared", "get", "--id", "-1"}},
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
				t.Error("expected error for invalid ID")
			}
		})
	}
}

// Execution tests using mock servers

func TestAudienceListCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/audienceGroup/list") {
			w.Header().Set("Content-Type", "application/json")
			audienceGroupID := int64(123456)
			description := "Test Audience"
			status := "READY"
			audienceCount := int64(1000)
			created := int64(1700000000)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroups": []map[string]any{
					{
						"audienceGroupId": audienceGroupID,
						"description":     description,
						"status":          status,
						"audienceCount":   audienceCount,
						"created":         created,
					},
				},
				"hasNextPage": false,
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
			checkText: "Test Audience",
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

			cmd := newAudienceListCmdWithClient(client)
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := out.String()
			if tt.wantJSON {
				var result []any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("expected valid JSON output, got: %s", output)
				}
				if len(result) == 0 {
					t.Error("expected at least one audience group")
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestAudienceListCmd_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/audienceGroup/list") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroups": []any{},
				"hasNextPage":    false,
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

	cmd := newAudienceListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "No audience groups found") {
		t.Errorf("expected 'No audience groups found', got: %s", out.String())
	}
}

func TestAudienceGetCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/12345" {
			w.Header().Set("Content-Type", "application/json")
			audienceGroupID := int64(12345)
			description := "My Test Audience"
			status := "READY"
			groupType := "UPLOAD"
			audienceCount := int64(500)
			created := int64(1700000000)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroup": map[string]any{
					"audienceGroupId": audienceGroupID,
					"description":     description,
					"status":          status,
					"type":            groupType,
					"audienceCount":   audienceCount,
					"created":         created,
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
			checkText: "My Test Audience",
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

			cmd := newAudienceGetCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "12345"})
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
				ag, ok := result["audienceGroup"].(map[string]any)
				if !ok {
					t.Error("expected audienceGroup in response")
				}
				if ag["description"] != "My Test Audience" {
					t.Errorf("expected description 'My Test Audience', got: %v", ag["description"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
				if !strings.Contains(output, "READY") {
					t.Errorf("expected output to contain status 'READY', got: %s", output)
				}
			}
		})
	}
}

func TestAudienceDeleteCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/99999" && r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
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
			checkText: "Deleted audience group: 99999",
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

			cmd := newAudienceDeleteCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "99999"})
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
				if result["deleted"] != float64(99999) {
					t.Errorf("expected deleted '99999', got: %v", result["deleted"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestAudienceCreateCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/upload" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroupId": 77777,
				"type":            "UPLOAD",
				"description":     "New Test Audience",
				"created":         1700000000,
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
			checkText: "Created audience group: 77777",
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

			cmd := newAudienceCreateCmdWithClient(client)
			cmd.SetArgs([]string{"--name", "New Test Audience", "--users", "U123,U456,U789"})
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
				if result["audienceGroupId"] != float64(77777) {
					t.Errorf("expected audienceGroupId '77777', got: %v", result["audienceGroupId"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestAudienceSharedListCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/audienceGroup/shared/list") {
			w.Header().Set("Content-Type", "application/json")
			audienceGroupID := int64(555555)
			description := "Shared Test Audience"
			status := "READY"
			audienceCount := int64(2500)
			created := int64(1700000000)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroups": []map[string]any{
					{
						"audienceGroupId": audienceGroupID,
						"description":     description,
						"status":          status,
						"audienceCount":   audienceCount,
						"created":         created,
					},
				},
				"hasNextPage": false,
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
			checkText: "Shared Test Audience",
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

			cmd := newAudienceSharedListCmdWithClient(client)
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := out.String()
			if tt.wantJSON {
				var result []any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("expected valid JSON output, got: %s", output)
				}
				if len(result) == 0 {
					t.Error("expected at least one shared audience group")
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestAudienceSharedListCmd_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/audienceGroup/shared/list") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroups": []any{},
				"hasNextPage":    false,
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

	cmd := newAudienceSharedListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "No shared audience groups found") {
		t.Errorf("expected 'No shared audience groups found', got: %s", out.String())
	}
}

func TestAudienceSharedGetCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/shared/88888" {
			w.Header().Set("Content-Type", "application/json")
			audienceGroupID := int64(88888)
			description := "Shared Audience Details"
			status := "READY"
			groupType := "UPLOAD"
			audienceCount := int64(3000)
			created := int64(1700000000)
			ownerName := "Test Owner"
			serviceType := "LINE_OA"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroup": map[string]any{
					"audienceGroupId": audienceGroupID,
					"description":     description,
					"status":          status,
					"type":            groupType,
					"audienceCount":   audienceCount,
					"created":         created,
				},
				"owner": map[string]any{
					"name":        ownerName,
					"serviceType": serviceType,
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
			checkText: "Shared Audience Details",
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

			cmd := newAudienceSharedGetCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "88888"})
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
				ag, ok := result["audienceGroup"].(map[string]any)
				if !ok {
					t.Error("expected audienceGroup in response")
				}
				if ag["description"] != "Shared Audience Details" {
					t.Errorf("expected description 'Shared Audience Details', got: %v", ag["description"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
				if !strings.Contains(output, "Test Owner") {
					t.Errorf("expected output to contain owner name 'Test Owner', got: %s", output)
				}
			}
		})
	}
}

// API error tests

func TestAudienceListCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid token"})
	}))
	defer server.Close()

	client := api.NewClient("bad-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newAudienceListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to list audience groups") {
		t.Errorf("expected 'failed to list audience groups' in error, got: %v", err)
	}
}

func TestAudienceGetCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newAudienceGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "99999"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get audience group") {
		t.Errorf("expected 'failed to get audience group' in error, got: %v", err)
	}
}

func TestAudienceDeleteCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newAudienceDeleteCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "11111"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to delete audience group") {
		t.Errorf("expected 'failed to delete audience group' in error, got: %v", err)
	}
}

func TestAudienceSharedListCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Server error"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newAudienceSharedListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to list shared audience groups") {
		t.Errorf("expected 'failed to list shared audience groups' in error, got: %v", err)
	}
}

func TestAudienceSharedGetCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newAudienceSharedGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "55555"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get shared audience group") {
		t.Errorf("expected 'failed to get shared audience group' in error, got: %v", err)
	}
}

// Additional coverage tests

func TestAudienceListCmd_TableOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/audienceGroup/list") {
			w.Header().Set("Content-Type", "application/json")
			audienceGroupID := int64(123456)
			description := "Test Audience"
			status := "READY"
			audienceCount := int64(1000)
			created := int64(1700000000)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroups": []map[string]any{
					{
						"audienceGroupId": audienceGroupID,
						"description":     description,
						"status":          status,
						"audienceCount":   audienceCount,
						"created":         created,
					},
				},
				"hasNextPage": false,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "table"
	defer func() { flags.Output = oldOutput }()

	cmd := newAudienceListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "ID") || !strings.Contains(output, "DESCRIPTION") {
		t.Errorf("expected table headers, got: %s", output)
	}
	if !strings.Contains(output, "Test Audience") {
		t.Errorf("expected audience description, got: %s", output)
	}
}

func TestAudienceListCmd_NilFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/audienceGroup/list") {
			w.Header().Set("Content-Type", "application/json")
			// Return an audience group with nil fields
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroups": []map[string]any{
					{},
				},
				"hasNextPage": false,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Test text output with nil fields
	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newAudienceListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "unknown") {
		t.Errorf("expected 'unknown' for nil fields, got: %s", output)
	}
}

func TestAudienceListCmd_TableNilFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/audienceGroup/list") {
			w.Header().Set("Content-Type", "application/json")
			// Return an audience group with nil fields
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroups": []map[string]any{
					{},
				},
				"hasNextPage": false,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "table"
	defer func() { flags.Output = oldOutput }()

	cmd := newAudienceListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Table should render without crashing even with nil fields
	output := out.String()
	if !strings.Contains(output, "ID") {
		t.Errorf("expected table headers, got: %s", output)
	}
}

func TestAudienceGetCmd_NilResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/12345" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroup": nil,
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

	cmd := newAudienceGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "12345"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nil audienceGroup")
	}
	if !strings.Contains(err.Error(), "audience group not found") {
		t.Errorf("expected 'audience group not found', got: %v", err)
	}
}

func TestAudienceGetCmd_NilFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/12345" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroup": map[string]any{},
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

	cmd := newAudienceGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "12345"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Should handle nil fields gracefully with "unknown"
	if !strings.Contains(output, "unknown") {
		t.Errorf("expected 'unknown' for nil fields, got: %s", output)
	}
}

func TestAudienceCreateClickCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/click" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroupId": 88888,
				"type":            "CLICK",
				"description":     "Click Audience",
				"created":         1700000000,
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
			checkText: "Created click-based audience group: 88888",
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

			cmd := newAudienceCreateClickCmdWithClient(client)
			cmd.SetArgs([]string{"--name", "Click Audience", "--request", "req-123"})
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
				if result["audienceGroupId"] != float64(88888) {
					t.Errorf("expected audienceGroupId 88888, got: %v", result["audienceGroupId"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestAudienceCreateClickCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Bad request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newAudienceCreateClickCmdWithClient(client)
	cmd.SetArgs([]string{"--name", "Test", "--request", "req-123"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to create click-based audience") {
		t.Errorf("expected 'failed to create click-based audience' in error, got: %v", err)
	}
}

func TestAudienceCreateImpressionCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/imp" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroupId": 99999,
				"type":            "IMPRESSION",
				"description":     "Impression Audience",
				"created":         1700000000,
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
			checkText: "Created impression-based audience group: 99999",
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

			cmd := newAudienceCreateImpressionCmdWithClient(client)
			cmd.SetArgs([]string{"--name", "Impression Audience", "--request", "req-456"})
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
				if result["audienceGroupId"] != float64(99999) {
					t.Errorf("expected audienceGroupId 99999, got: %v", result["audienceGroupId"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestAudienceCreateImpressionCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Bad request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newAudienceCreateImpressionCmdWithClient(client)
	cmd.SetArgs([]string{"--name", "Test", "--request", "req-456"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to create impression-based audience") {
		t.Errorf("expected 'failed to create impression-based audience' in error, got: %v", err)
	}
}

func TestAudienceUpdateDescriptionCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/12345/updateDescription" && r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
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
			checkText: "Updated description for audience group 12345",
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

			cmd := newAudienceUpdateDescriptionCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "12345", "--description", "New Description"})
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
				if result["audienceGroupId"] != float64(12345) {
					t.Errorf("expected audienceGroupId 12345, got: %v", result["audienceGroupId"])
				}
				if result["description"] != "New Description" {
					t.Errorf("expected description 'New Description', got: %v", result["description"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestAudienceUpdateDescriptionCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newAudienceUpdateDescriptionCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "12345", "--description", "Test"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to update audience description") {
		t.Errorf("expected 'failed to update audience description' in error, got: %v", err)
	}
}

func TestAudienceAddUsersCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/upload" && r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
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
			checkText: "Added 3 users to audience group 12345",
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

			cmd := newAudienceAddUsersCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "12345", "--users", "U123,U456,U789"})
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
				if result["audienceGroupId"] != float64(12345) {
					t.Errorf("expected audienceGroupId 12345, got: %v", result["audienceGroupId"])
				}
				if result["usersAdded"] != float64(3) {
					t.Errorf("expected usersAdded 3, got: %v", result["usersAdded"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestAudienceAddUsersCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newAudienceAddUsersCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "12345", "--users", "U123"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to add users to audience") {
		t.Errorf("expected 'failed to add users to audience' in error, got: %v", err)
	}
}

func TestAudienceSharedListCmd_TableOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/audienceGroup/shared/list") {
			w.Header().Set("Content-Type", "application/json")
			audienceGroupID := int64(555555)
			description := "Shared Test Audience"
			status := "READY"
			audienceCount := int64(2500)
			created := int64(1700000000)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroups": []map[string]any{
					{
						"audienceGroupId": audienceGroupID,
						"description":     description,
						"status":          status,
						"audienceCount":   audienceCount,
						"created":         created,
					},
				},
				"hasNextPage": false,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "table"
	defer func() { flags.Output = oldOutput }()

	cmd := newAudienceSharedListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "ID") || !strings.Contains(output, "DESCRIPTION") {
		t.Errorf("expected table headers, got: %s", output)
	}
	if !strings.Contains(output, "Shared Test Audience") {
		t.Errorf("expected audience description, got: %s", output)
	}
}

func TestAudienceSharedListCmd_NilFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/audienceGroup/shared/list") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroups": []map[string]any{
					{},
				},
				"hasNextPage": false,
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

	cmd := newAudienceSharedListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "unknown") {
		t.Errorf("expected 'unknown' for nil fields, got: %s", output)
	}
}

func TestAudienceSharedListCmd_TableNilFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/audienceGroup/shared/list") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroups": []map[string]any{
					{},
				},
				"hasNextPage": false,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "table"
	defer func() { flags.Output = oldOutput }()

	cmd := newAudienceSharedListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "ID") {
		t.Errorf("expected table headers, got: %s", output)
	}
}

func TestAudienceSharedGetCmd_NilResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/shared/12345" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroup": nil,
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

	cmd := newAudienceSharedGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "12345"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nil audienceGroup")
	}
	if !strings.Contains(err.Error(), "shared audience group not found") {
		t.Errorf("expected 'shared audience group not found', got: %v", err)
	}
}

func TestAudienceSharedGetCmd_NilFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/shared/12345" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroup": map[string]any{},
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

	cmd := newAudienceSharedGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "12345"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "unknown") {
		t.Errorf("expected 'unknown' for nil fields, got: %s", output)
	}
}

func TestAudienceSharedGetCmd_OwnerWithNilFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/shared/12345" {
			w.Header().Set("Content-Type", "application/json")
			audienceGroupID := int64(12345)
			description := "Shared Audience"
			status := "READY"
			groupType := "UPLOAD"
			audienceCount := int64(1000)
			created := int64(1700000000)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroup": map[string]any{
					"audienceGroupId": audienceGroupID,
					"description":     description,
					"status":          status,
					"type":            groupType,
					"audienceCount":   audienceCount,
					"created":         created,
				},
				"owner": map[string]any{},
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

	cmd := newAudienceSharedGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "12345"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Should print "Owner:" even with nil fields
	if !strings.Contains(output, "Owner:") {
		t.Errorf("expected 'Owner:' in output, got: %s", output)
	}
}

func TestAudienceCreateCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Bad request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newAudienceCreateCmdWithClient(client)
	cmd.SetArgs([]string{"--name", "Test Audience", "--users", "U123"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to create audience") {
		t.Errorf("expected 'failed to create audience' in error, got: %v", err)
	}
}

func TestAudienceCreateCmd_FromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/upload/byFile" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"audienceGroupId": 88888,
				"type":            "UPLOAD",
				"description":     "File Upload Audience",
				"created":         1700000000,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Create temp file with user IDs
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "users.txt")
	if err := os.WriteFile(tmpFile, []byte("U123\nU456\nU789\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

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
			checkText: "Created audience group: 88888",
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

			cmd := newAudienceCreateCmdWithClient(client)
			cmd.SetArgs([]string{"--name", "File Upload Audience", "--file", tmpFile})
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
				if result["audienceGroupId"] != float64(88888) {
					t.Errorf("expected audienceGroupId 88888, got: %v", result["audienceGroupId"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestAudienceCreateCmd_FromFile_EmptyFile(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	// Create empty temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cmd := newAudienceCreateCmdWithClient(client)
	cmd.SetArgs([]string{"--name", "Empty File Audience", "--file", tmpFile})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for empty file")
	}
	if !strings.Contains(err.Error(), "file contains no user IDs") {
		t.Errorf("expected 'file contains no user IDs' in error, got: %v", err)
	}
}

func TestAudienceCreateCmd_FromFile_NonexistentFile(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newAudienceCreateCmdWithClient(client)
	cmd.SetArgs([]string{"--name", "Nonexistent File Audience", "--file", "/nonexistent/path/users.txt"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("expected 'failed to read file' in error, got: %v", err)
	}
}

func TestAudienceCreateCmd_FromFile_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Bad request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Create temp file with user IDs
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "users.txt")
	if err := os.WriteFile(tmpFile, []byte("U123\nU456\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cmd := newAudienceCreateCmdWithClient(client)
	cmd.SetArgs([]string{"--name", "File Audience", "--file", tmpFile})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to create audience") {
		t.Errorf("expected 'failed to create audience' in error, got: %v", err)
	}
}

func TestAudienceAddUsersCmd_FromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/audienceGroup/upload/byFile" && r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Create temp file with user IDs
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "users.txt")
	if err := os.WriteFile(tmpFile, []byte("U123\nU456\nU789\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

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
			checkText: "Added 3 users to audience group 12345",
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

			cmd := newAudienceAddUsersCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "12345", "--file", tmpFile})
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
				if result["audienceGroupId"] != float64(12345) {
					t.Errorf("expected audienceGroupId 12345, got: %v", result["audienceGroupId"])
				}
				if result["usersAdded"] != float64(3) {
					t.Errorf("expected usersAdded 3, got: %v", result["usersAdded"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestAudienceAddUsersCmd_FromFile_EmptyFile(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	// Create empty temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cmd := newAudienceAddUsersCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "12345", "--file", tmpFile})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for empty file")
	}
	if !strings.Contains(err.Error(), "file contains no user IDs") {
		t.Errorf("expected 'file contains no user IDs' in error, got: %v", err)
	}
}

func TestAudienceAddUsersCmd_FromFile_NonexistentFile(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newAudienceAddUsersCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "12345", "--file", "/nonexistent/path/users.txt"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("expected 'failed to read file' in error, got: %v", err)
	}
}

func TestAudienceAddUsersCmd_FromFile_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Create temp file with user IDs
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "users.txt")
	if err := os.WriteFile(tmpFile, []byte("U123\nU456\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cmd := newAudienceAddUsersCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "12345", "--file", tmpFile})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to add users to audience") {
		t.Errorf("expected 'failed to add users to audience' in error, got: %v", err)
	}
}
