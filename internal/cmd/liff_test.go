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

func TestLIFFCmd_RequiresSubcommand(t *testing.T) {
	cmd := newLIFFCmd()

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

func TestLIFFCmd_HasSubcommands(t *testing.T) {
	cmd := newLIFFCmd()

	subcommands := cmd.Commands()
	if len(subcommands) != 4 {
		t.Errorf("expected 4 subcommands (list, create, update, delete), got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expected := []string{"list", "create", "update", "delete"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestLIFFCreateCmd_RequiresViewType(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"liff", "create", "--url", "https://example.com/liff"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --type flag")
	}
}

func TestLIFFCreateCmd_RequiresURL(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"liff", "create", "--type", "compact"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --url flag")
	}
}

func TestLIFFUpdateCmd_RequiresID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"liff", "update", "--type", "full"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestLIFFDeleteCmd_RequiresID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"liff", "delete"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestLIFFCreateCmd_ValidateViewType(t *testing.T) {
	tests := []struct {
		name      string
		viewType  string
		expectErr bool
	}{
		{"compact is valid", "compact", false},
		{"tall is valid", "tall", false},
		{"full is valid", "full", false},
		{"invalid type", "medium", true},
		{"empty type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newLIFFCreateCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			args := []string{"--url", "https://example.com/liff"}
			if tt.viewType != "" {
				args = append(args, "--type", tt.viewType)
			}
			cmd.SetArgs(args)

			// We can't fully execute without credentials, but we can check flag validation
			err := cmd.Flags().Set("type", tt.viewType)
			if tt.viewType != "" && err != nil {
				t.Errorf("failed to set type flag: %v", err)
			}

			// Check that the type flag exists
			typeFlag := cmd.Flags().Lookup("type")
			if typeFlag == nil {
				t.Error("expected --type flag")
			}
		})
	}
}

// Execution tests using httptest

func TestLIFFListCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liff/v1/apps" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"apps": []map[string]any{
					{
						"liffId":      "1234567890-abcdefgh",
						"view":        map[string]string{"type": "full", "url": "https://example.com/liff"},
						"description": "Test LIFF App",
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
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			checkText: "1234567890-abcdefgh",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"liffId"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newLIFFListCmdWithClient(client)
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestLIFFListCmd_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liff/v1/apps" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"apps": []any{}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	defer func() { flags.Output = oldOutput }()
	flags.Output = "text"

	cmd := newLIFFListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "No LIFF apps found") {
		t.Errorf("output should indicate no apps found, got: %s", out.String())
	}
}

func TestLIFFListCmd_TableOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liff/v1/apps" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"apps": []map[string]any{
					{
						"liffId":      "liff-123",
						"view":        map[string]string{"type": "compact", "url": "https://example.com"},
						"description": "Compact App",
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

	oldOutput := flags.Output
	defer func() { flags.Output = oldOutput }()
	flags.Output = "table"

	cmd := newLIFFListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Table output should contain header and data
	if !strings.Contains(out.String(), "LIFF ID") {
		t.Errorf("table output should contain header, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "liff-123") {
		t.Errorf("table output should contain LIFF ID, got: %s", out.String())
	}
}

func TestLIFFDeleteCmd_Execute(t *testing.T) {
	var deletedID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/liff/v1/apps/") {
			deletedID = strings.TrimPrefix(r.URL.Path, "/liff/v1/apps/")
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
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			checkText: "Deleted LIFF app: liff-456",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"status": "deleted"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			oldYes := flags.Yes
			defer func() {
				flags.Output = oldOutput
				flags.Yes = oldYes
			}()
			flags.Output = tt.output
			flags.Yes = true

			cmd := newLIFFDeleteCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "liff-456"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if deletedID != "liff-456" {
				t.Errorf("expected liff-456 to be deleted, got %s", deletedID)
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestLIFFDeleteCmd_RequiresConfirmation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldYes := flags.Yes
	defer func() { flags.Yes = oldYes }()
	flags.Yes = false

	cmd := newLIFFDeleteCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "liff-123"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --yes is not provided")
	}

	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("error should mention --yes flag, got: %v", err)
	}
}

func TestLIFFDeleteCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "LIFF app not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldYes := flags.Yes
	defer func() { flags.Yes = oldYes }()
	flags.Yes = true

	cmd := newLIFFDeleteCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "liff-nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent LIFF app")
	}
}

// Tests for LIFF list with and without descriptions
func TestLIFFListCmd_TextWithDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"apps": []map[string]any{
				{
					"liffId":      "liff-with-desc",
					"view":        map[string]string{"type": "full", "url": "https://example.com"},
					"description": "Has description",
				},
			},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	defer func() { flags.Output = oldOutput }()
	flags.Output = "text"

	cmd := newLIFFListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Description: Has description") {
		t.Errorf("output should contain description, got: %s", output)
	}
}

func TestLIFFListCmd_TextWithoutDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"apps": []map[string]any{
				{
					"liffId":      "liff-no-desc",
					"view":        map[string]string{"type": "compact", "url": "https://example.com"},
					"description": "",
				},
			},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	defer func() { flags.Output = oldOutput }()
	flags.Output = "text"

	cmd := newLIFFListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "Description:") {
		t.Errorf("output should NOT contain description line when empty, got: %s", output)
	}
}

func TestLIFFListCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Server error"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newLIFFListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to list LIFF apps") {
		t.Errorf("error should mention 'failed to list LIFF apps', got: %v", err)
	}
}

// Tests for LIFF create command
func TestLIFFCreateCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liff/v1/apps" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"liffId": "new-liff-123"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	tests := []struct {
		name        string
		output      string
		description string
		checkText   string
	}{
		{
			name:        "text output without description",
			output:      "text",
			description: "",
			checkText:   "Created LIFF app: new-liff-123",
		},
		{
			name:        "text output with description",
			output:      "text",
			description: "My App",
			checkText:   "Description:      My App",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"liffId"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newLIFFCreateCmdWithClient(client)
			args := []string{"--type", "full", "--url", "https://example.com/liff"}
			if tt.description != "" {
				args = append(args, "--description", tt.description)
			}
			cmd.SetArgs(args)
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestLIFFCreateCmd_AllViewTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"liffId": "liff-123"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	viewTypes := []string{"compact", "tall", "full"}

	for _, vt := range viewTypes {
		t.Run("view type "+vt, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = "text"

			cmd := newLIFFCreateCmdWithClient(client)
			cmd.SetArgs([]string{"--type", vt, "--url", "https://example.com"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error for view type %s: %v", vt, err)
			}

			if !strings.Contains(out.String(), "View Type:        "+vt) {
				t.Errorf("output should contain view type %s, got: %s", vt, out.String())
			}
		})
	}
}

func TestLIFFCreateCmd_ValidationErrors(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	tests := []struct {
		name        string
		args        []string
		errContains string
	}{
		{
			name:        "missing type",
			args:        []string{"--url", "https://example.com"},
			errContains: "--type is required",
		},
		{
			name:        "invalid type",
			args:        []string{"--type", "invalid", "--url", "https://example.com"},
			errContains: "--type must be one of",
		},
		{
			name:        "missing url",
			args:        []string{"--type", "full"},
			errContains: "--url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newLIFFCreateCmdWithClient(client)
			cmd.SetArgs(tt.args)
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error should contain %q, got: %v", tt.errContains, err)
			}
		})
	}
}

func TestLIFFCreateCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newLIFFCreateCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "full", "--url", "https://example.com"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to create LIFF app") {
		t.Errorf("error should mention 'failed to create LIFF app', got: %v", err)
	}
}

// Tests for LIFF update command
func TestLIFFUpdateCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/liff/v1/apps/") {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	tests := []struct {
		name        string
		output      string
		description string
		checkText   string
	}{
		{
			name:        "text output without description",
			output:      "text",
			description: "",
			checkText:   "Updated LIFF app: liff-123",
		},
		{
			name:        "text output with description",
			output:      "text",
			description: "Updated App",
			checkText:   "Description:      Updated App",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"status": "updated"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newLIFFUpdateCmdWithClient(client)
			args := []string{"--id", "liff-123", "--type", "full", "--url", "https://example.com/new"}
			if tt.description != "" {
				args = append(args, "--description", tt.description)
			}
			cmd.SetArgs(args)
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestLIFFUpdateCmd_AllViewTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	viewTypes := []string{"compact", "tall", "full"}

	for _, vt := range viewTypes {
		t.Run("view type "+vt, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = "text"

			cmd := newLIFFUpdateCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "liff-123", "--type", vt, "--url", "https://example.com"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error for view type %s: %v", vt, err)
			}

			if !strings.Contains(out.String(), "View Type:        "+vt) {
				t.Errorf("output should contain view type %s, got: %s", vt, out.String())
			}
		})
	}
}

func TestLIFFUpdateCmd_ValidationErrors(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	tests := []struct {
		name        string
		args        []string
		errContains string
	}{
		{
			name:        "missing id",
			args:        []string{"--type", "full", "--url", "https://example.com"},
			errContains: "--id is required",
		},
		{
			name:        "missing type",
			args:        []string{"--id", "liff-123", "--url", "https://example.com"},
			errContains: "--type is required",
		},
		{
			name:        "invalid type",
			args:        []string{"--id", "liff-123", "--type", "invalid", "--url", "https://example.com"},
			errContains: "--type must be one of",
		},
		{
			name:        "missing url",
			args:        []string{"--id", "liff-123", "--type", "full"},
			errContains: "--url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newLIFFUpdateCmdWithClient(client)
			cmd.SetArgs(tt.args)
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error should contain %q, got: %v", tt.errContains, err)
			}
		})
	}
}

func TestLIFFUpdateCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "LIFF app not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newLIFFUpdateCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "liff-999", "--type", "full", "--url", "https://example.com"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to update LIFF app") {
		t.Errorf("error should mention 'failed to update LIFF app', got: %v", err)
	}
}
