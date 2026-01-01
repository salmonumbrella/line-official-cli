package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/salmonumbrella/line-official-cli/internal/api"
)

func TestRichMenuCmd_HasSubcommands(t *testing.T) {
	cmd := newRichMenuCmd()

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("expected richmenu to have subcommands")
	}

	// Verify key subcommands exist
	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expectedSubcommands := []string{
		"list", "create", "delete", "set-default", "cancel-default",
		"upload-image", "get", "link", "unlink", "alias", "bulk", "batch",
		"validate", "download-image",
	}

	for _, expected := range expectedSubcommands {
		if !names[expected] {
			t.Errorf("expected '%s' subcommand", expected)
		}
	}
}

func TestRichMenuCmd_Aliases(t *testing.T) {
	cmd := newRichMenuCmd()

	if len(cmd.Aliases) == 0 {
		t.Error("expected richmenu to have aliases")
	}

	hasRM := false
	for _, alias := range cmd.Aliases {
		if alias == "rm" {
			hasRM = true
			break
		}
	}

	if !hasRM {
		t.Error("expected 'rm' alias for richmenu command")
	}
}

func TestRichMenuListCmd_NoFlags(t *testing.T) {
	cmd := newRichMenuListCmd()

	// List command should run without any required flags
	// (it will fail due to missing API client, but flags are valid)
	if cmd.Use != "list" {
		t.Errorf("expected Use to be 'list', got %s", cmd.Use)
	}
}

func TestRichMenuCreateCmd_Flags(t *testing.T) {
	cmd := newRichMenuCreateCmd()

	// Check required flags exist
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("expected --name flag")
	}

	actionsFlag := cmd.Flags().Lookup("actions")
	if actionsFlag == nil {
		t.Error("expected --actions flag")
	}

	sizeFlag := cmd.Flags().Lookup("size")
	if sizeFlag == nil {
		t.Fatal("expected --size flag")
	}

	// Check default value for size
	if sizeFlag.DefValue != "full" {
		t.Errorf("expected size default to be 'full', got %s", sizeFlag.DefValue)
	}
}

func TestRichMenuCreateCmd_RequiresName(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"richmenu", "create", "--actions", `[{"type":"message"}]`})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --name flag")
	}
}

func TestRichMenuCreateCmd_RequiresActions(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"richmenu", "create", "--name", "Test Menu"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --actions flag")
	}
}

func TestRichMenuCreateCmd_ValidatesSize(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"richmenu", "create", "--name", "Test", "--actions", `[]`, "--size", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid size")
	}
}

func TestRichMenuDeleteCmd_Flags(t *testing.T) {
	cmd := newRichMenuDeleteCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}
}

func TestRichMenuDeleteCmd_RequiresID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"richmenu", "delete"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestRichMenuSetDefaultCmd_Flags(t *testing.T) {
	cmd := newRichMenuSetDefaultCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}
}

func TestRichMenuSetDefaultCmd_RequiresID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"richmenu", "set-default"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestRichMenuCancelDefaultCmd_NoRequiredFlags(t *testing.T) {
	cmd := newRichMenuCancelDefaultCmd()

	// cancel-default has no required flags
	if cmd.Use != "cancel-default" {
		t.Errorf("expected Use to be 'cancel-default', got %s", cmd.Use)
	}
}

func TestRichMenuUploadImageCmd_Flags(t *testing.T) {
	cmd := newRichMenuUploadImageCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}

	imageFlag := cmd.Flags().Lookup("image")
	if imageFlag == nil {
		t.Error("expected --image flag")
	}
}

func TestRichMenuUploadImageCmd_RequiresBothFlags(t *testing.T) {
	cmd := NewRootCmd()

	tests := []struct {
		name string
		args []string
	}{
		{"missing both", []string{"richmenu", "upload-image"}},
		{"missing image", []string{"richmenu", "upload-image", "--id", "richmenu-123"}},
		{"missing id", []string{"richmenu", "upload-image", "--image", "test.png"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Error("expected error for missing required flags")
			}
		})
	}
}

func TestRichMenuGetCmd_Flags(t *testing.T) {
	cmd := newRichMenuGetCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}
}

func TestRichMenuLinkCmd_Flags(t *testing.T) {
	cmd := newRichMenuLinkCmd()

	userFlag := cmd.Flags().Lookup("user")
	if userFlag == nil {
		t.Error("expected --user flag")
	}

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}
}

func TestRichMenuUnlinkCmd_Flags(t *testing.T) {
	cmd := newRichMenuUnlinkCmd()

	userFlag := cmd.Flags().Lookup("user")
	if userFlag == nil {
		t.Error("expected --user flag")
	}
}

func TestRichMenuAliasCmd_HasSubcommands(t *testing.T) {
	cmd := newRichMenuAliasCmd()

	subcommands := cmd.Commands()
	if len(subcommands) != 5 {
		t.Errorf("expected 5 alias subcommands, got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expected := []string{"create", "get", "update", "delete", "list"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestRichMenuAliasCreateCmd_Flags(t *testing.T) {
	cmd := newRichMenuAliasCreateCmd()

	aliasFlag := cmd.Flags().Lookup("alias")
	if aliasFlag == nil {
		t.Error("expected --alias flag")
	}

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}
}

func TestRichMenuAliasGetCmd_Flags(t *testing.T) {
	cmd := newRichMenuAliasGetCmd()

	aliasFlag := cmd.Flags().Lookup("alias")
	if aliasFlag == nil {
		t.Error("expected --alias flag")
	}
}

func TestRichMenuAliasUpdateCmd_Flags(t *testing.T) {
	cmd := newRichMenuAliasUpdateCmd()

	aliasFlag := cmd.Flags().Lookup("alias")
	if aliasFlag == nil {
		t.Error("expected --alias flag")
	}

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}
}

func TestRichMenuAliasDeleteCmd_Flags(t *testing.T) {
	cmd := newRichMenuAliasDeleteCmd()

	aliasFlag := cmd.Flags().Lookup("alias")
	if aliasFlag == nil {
		t.Error("expected --alias flag")
	}
}

func TestRichMenuBulkCmd_HasSubcommands(t *testing.T) {
	cmd := newRichMenuBulkCmd()

	subcommands := cmd.Commands()
	if len(subcommands) != 2 {
		t.Errorf("expected 2 bulk subcommands, got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	if !names["link"] {
		t.Error("expected 'link' subcommand")
	}
	if !names["unlink"] {
		t.Error("expected 'unlink' subcommand")
	}
}

func TestRichMenuBulkLinkCmd_Flags(t *testing.T) {
	cmd := newRichMenuBulkLinkCmd()

	menuFlag := cmd.Flags().Lookup("menu")
	if menuFlag == nil {
		t.Error("expected --menu flag")
	}

	usersFlag := cmd.Flags().Lookup("users")
	if usersFlag == nil {
		t.Error("expected --users flag")
	}
}

func TestRichMenuBulkUnlinkCmd_Flags(t *testing.T) {
	cmd := newRichMenuBulkUnlinkCmd()

	usersFlag := cmd.Flags().Lookup("users")
	if usersFlag == nil {
		t.Error("expected --users flag")
	}
}

func TestRichMenuBatchCmd_HasSubcommands(t *testing.T) {
	cmd := newRichMenuBatchCmd()

	subcommands := cmd.Commands()
	if len(subcommands) != 2 {
		t.Errorf("expected 2 batch subcommands, got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	if !names["validate"] {
		t.Error("expected 'validate' subcommand")
	}
	if !names["status"] {
		t.Error("expected 'status' subcommand")
	}
}

func TestRichMenuBatchCmd_Flags(t *testing.T) {
	cmd := newRichMenuBatchCmd()

	operationsFlag := cmd.Flags().Lookup("operations")
	if operationsFlag == nil {
		t.Error("expected --operations flag")
	}

	resumeFlag := cmd.Flags().Lookup("resume")
	if resumeFlag == nil {
		t.Error("expected --resume flag")
	}
}

func TestRichMenuBatchValidateCmd_Flags(t *testing.T) {
	cmd := newRichMenuBatchValidateCmd()

	operationsFlag := cmd.Flags().Lookup("operations")
	if operationsFlag == nil {
		t.Error("expected --operations flag")
	}
}

func TestRichMenuBatchStatusCmd_Flags(t *testing.T) {
	cmd := newRichMenuBatchStatusCmd()

	requestFlag := cmd.Flags().Lookup("request")
	if requestFlag == nil {
		t.Error("expected --request flag")
	}
}

func TestRichMenuValidateCmd_Flags(t *testing.T) {
	cmd := newRichMenuValidateCmd()

	fileFlag := cmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Error("expected --file flag")
	}
}

func TestRichMenuDownloadImageCmd_Flags(t *testing.T) {
	cmd := newRichMenuDownloadImageCmd()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag")
	}

	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("expected --output flag")
	}
}

// Table-driven test for required flag validation
func TestRichMenuCommands_RequiredFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectsError bool
	}{
		// delete requires --id
		{"delete without id", []string{"richmenu", "delete"}, true},
		// set-default requires --id
		{"set-default without id", []string{"richmenu", "set-default"}, true},
		// get requires --id
		{"get without id", []string{"richmenu", "get"}, true},
		// link requires --user and --id
		{"link without flags", []string{"richmenu", "link"}, true},
		{"link without user", []string{"richmenu", "link", "--id", "rm-123"}, true},
		{"link without id", []string{"richmenu", "link", "--user", "U123"}, true},
		// unlink requires --user
		{"unlink without user", []string{"richmenu", "unlink"}, true},
		// alias create requires --alias and --id
		{"alias create without flags", []string{"richmenu", "alias", "create"}, true},
		// alias get requires --alias
		{"alias get without alias", []string{"richmenu", "alias", "get"}, true},
		// alias update requires --alias and --id
		{"alias update without flags", []string{"richmenu", "alias", "update"}, true},
		// alias delete requires --alias
		{"alias delete without alias", []string{"richmenu", "alias", "delete"}, true},
		// batch validate requires --operations
		{"batch validate without operations", []string{"richmenu", "batch", "validate"}, true},
		// batch status requires --request
		{"batch status without request", []string{"richmenu", "batch", "status"}, true},
		// validate requires --file
		{"validate without file", []string{"richmenu", "validate"}, true},
		// download-image requires --id
		{"download-image without id", []string{"richmenu", "download-image"}, true},
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

// Execution tests using httptest

func TestRichMenuListCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bot/richmenu/list":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"richmenus": []map[string]any{
					{
						"richMenuId":  "rm-123",
						"name":        "Test Menu",
						"chatBarText": "Menu",
						"size":        map[string]int{"width": 2500, "height": 1686},
						"areas":       []any{},
					},
				},
			})
		case "/v2/bot/user/all/richmenu":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"richMenuId": "rm-123"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
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
			checkText: "rm-123",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"richMenuId"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuListCmdWithClient(client)
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

func TestRichMenuListCmd_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bot/richmenu/list":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"richmenus": []any{}})
		case "/v2/bot/user/all/richmenu":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "no default"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	defer func() { flags.Output = oldOutput }()
	flags.Output = "text"

	cmd := newRichMenuListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "No rich menus found") {
		t.Errorf("output should indicate no menus found, got: %s", out.String())
	}
}

func TestRichMenuGetCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/richmenu/rm-123" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"richMenuId":  "rm-123",
				"name":        "Test Menu",
				"chatBarText": "Menu",
				"size":        map[string]int{"width": 2500, "height": 1686},
				"areas":       []any{},
				"selected":    false,
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
			checkText: "rm-123",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"richMenuId"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuGetCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "rm-123"})
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

func TestRichMenuGetCmd_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "rm-nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent menu")
	}
}

func TestRichMenuDeleteCmd_Execute(t *testing.T) {
	var deletedID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/v2/bot/richmenu/") {
			deletedID = strings.TrimPrefix(r.URL.Path, "/v2/bot/richmenu/")
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
			checkText: "Deleted rich menu: rm-456",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"deleted"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuDeleteCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "rm-456"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if deletedID != "rm-456" {
				t.Errorf("expected menu rm-456 to be deleted, got %s", deletedID)
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestRichMenuDeleteCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuDeleteCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "rm-nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent menu")
	}
}

// Alias command execution tests

func TestRichMenuAliasCreateCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/bot/richmenu/alias" {
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
			checkText: "Created alias 'test-alias' -> rm-123",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"status": "created"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuAliasCreateCmdWithClient(client)
			cmd.SetArgs([]string{"--alias", "test-alias", "--id", "rm-123"})
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

func TestRichMenuAliasGetCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/bot/richmenu/alias/test-alias" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"richMenuAliasId": "test-alias",
				"richMenuId":      "rm-123",
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
			checkText: "Rich Menu:  rm-123",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"richMenuId"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuAliasGetCmdWithClient(client)
			cmd.SetArgs([]string{"--alias", "test-alias"})
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

func TestRichMenuAliasUpdateCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/bot/richmenu/alias/test-alias" {
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
			checkText: "Updated alias 'test-alias' -> rm-456",
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

			cmd := newRichMenuAliasUpdateCmdWithClient(client)
			cmd.SetArgs([]string{"--alias", "test-alias", "--id", "rm-456"})
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

func TestRichMenuAliasDeleteCmd_Execute(t *testing.T) {
	var deletedAlias string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/v2/bot/richmenu/alias/") {
			deletedAlias = strings.TrimPrefix(r.URL.Path, "/v2/bot/richmenu/alias/")
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
			checkText: "Deleted alias: test-alias",
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
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuAliasDeleteCmdWithClient(client)
			cmd.SetArgs([]string{"--alias", "test-alias"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if deletedAlias != "test-alias" {
				t.Errorf("expected alias test-alias to be deleted, got %s", deletedAlias)
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestRichMenuAliasListCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/bot/richmenu/alias/list" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"aliases": []map[string]string{
					{"richMenuAliasId": "alias-1", "richMenuId": "rm-123"},
					{"richMenuAliasId": "alias-2", "richMenuId": "rm-456"},
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
			checkText: "alias-1 -> rm-123",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"aliases"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuAliasListCmdWithClient(client)
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

func TestRichMenuAliasListCmd_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/bot/richmenu/alias/list" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"aliases": []any{}})
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

	cmd := newRichMenuAliasListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "No aliases found") {
		t.Errorf("output should indicate no aliases found, got: %s", out.String())
	}
}

// Bulk command execution tests

func TestRichMenuBulkLinkCmd_Execute(t *testing.T) {
	var receivedMenuID string
	var receivedUserIDs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/bot/richmenu/bulk/link" {
			var req struct {
				RichMenuID string   `json:"richMenuId"`
				UserIDs    []string `json:"userIds"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			receivedMenuID = req.RichMenuID
			receivedUserIDs = req.UserIDs
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	userIDs := []string{"U001", "U002", "U003"}

	tests := []struct {
		name      string
		output    string
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			checkText: "Linked rich menu rm-123 to 3 users",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"userCount": 3`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuBulkLinkCmdWithClient(client, userIDs)
			cmd.SetArgs([]string{"--menu", "rm-123"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if receivedMenuID != "rm-123" {
				t.Errorf("expected menu ID rm-123, got %s", receivedMenuID)
			}
			if len(receivedUserIDs) != 3 {
				t.Errorf("expected 3 user IDs, got %d", len(receivedUserIDs))
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestRichMenuBulkUnlinkCmd_Execute(t *testing.T) {
	var receivedUserIDs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/bot/richmenu/bulk/unlink" {
			var req struct {
				UserIDs []string `json:"userIds"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			receivedUserIDs = req.UserIDs
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	userIDs := []string{"U001", "U002"}

	tests := []struct {
		name      string
		output    string
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			checkText: "Unlinked rich menus from 2 users",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"userCount": 2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuBulkUnlinkCmdWithClient(client, userIDs)
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(receivedUserIDs) != 2 {
				t.Errorf("expected 2 user IDs, got %d", len(receivedUserIDs))
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

// Validate command execution test

func TestRichMenuValidateCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/bot/richmenu/validate" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	menu := &api.CreateRichMenuRequest{
		Name:        "Test Menu",
		ChatBarText: "Menu",
		Size:        api.RichMenuSize{Width: 2500, Height: 1686},
		Areas:       []api.RichMenuArea{},
	}

	tests := []struct {
		name      string
		output    string
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			checkText: "Rich menu definition valid: Test Menu",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"valid": true`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuValidateCmdWithClient(client, menu)
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

func TestRichMenuValidateCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid menu"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	menu := &api.CreateRichMenuRequest{
		Name: "Invalid Menu",
	}

	cmd := newRichMenuValidateCmdWithClient(client, menu)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid menu")
	}
}

// Download image command execution test

func TestRichMenuDownloadImageCmd_Execute(t *testing.T) {
	imageData := []byte("fake-png-data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/bot/richmenu/rm-123/content" {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(imageData)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Create a temp file for output
	tmpDir := t.TempDir()
	outputPath := tmpDir + "/test-image.png"

	tests := []struct {
		name      string
		output    string
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			checkText: "Downloaded image to",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"contentType": "image/png"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuDownloadImageCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "rm-123", "--output", outputPath})
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

func TestRichMenuDownloadImageCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuDownloadImageCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "rm-nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent image")
	}
}

// Additional tests for table output

func TestRichMenuListCmd_TableOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bot/richmenu/list":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"richmenus": []map[string]any{
					{
						"richMenuId":  "rm-123",
						"name":        "Test Menu",
						"chatBarText": "Menu",
						"size":        map[string]int{"width": 2500, "height": 1686},
						"areas":       []any{},
					},
				},
			})
		case "/v2/bot/user/all/richmenu":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"richMenuId": "rm-123"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	defer func() { flags.Output = oldOutput }()
	flags.Output = "table"

	cmd := newRichMenuListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Table output should contain column headers
	output := out.String()
	if !strings.Contains(output, "ID") || !strings.Contains(output, "NAME") {
		t.Errorf("table output should contain headers, got: %s", output)
	}
}

func TestRichMenuAliasListCmd_TableOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/bot/richmenu/alias/list" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"aliases": []map[string]string{
					{"richMenuAliasId": "alias-1", "richMenuId": "rm-123"},
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

	cmd := newRichMenuAliasListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "ALIAS") {
		t.Errorf("table output should contain ALIAS header, got: %s", output)
	}
}

// Tests for API error handling

func TestRichMenuAliasCreateCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid alias"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuAliasCreateCmdWithClient(client)
	cmd.SetArgs([]string{"--alias", "test-alias", "--id", "rm-123"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for failed alias creation")
	}
}

func TestRichMenuAliasGetCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "alias not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuAliasGetCmdWithClient(client)
	cmd.SetArgs([]string{"--alias", "nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent alias")
	}
}

func TestRichMenuAliasUpdateCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "update failed"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuAliasUpdateCmdWithClient(client)
	cmd.SetArgs([]string{"--alias", "test-alias", "--id", "rm-123"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for failed alias update")
	}
}

func TestRichMenuAliasDeleteCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "alias not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuAliasDeleteCmdWithClient(client)
	cmd.SetArgs([]string{"--alias", "nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent alias")
	}
}

func TestRichMenuAliasListCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuAliasListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

// Tests for bulk operations errors

func TestRichMenuBulkLinkCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	userIDs := []string{"U001", "U002"}

	cmd := newRichMenuBulkLinkCmdWithClient(client, userIDs)
	cmd.SetArgs([]string{"--menu", "rm-123"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for failed bulk link")
	}
}

func TestRichMenuBulkUnlinkCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	userIDs := []string{"U001", "U002"}

	cmd := newRichMenuBulkUnlinkCmdWithClient(client, userIDs)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for failed bulk unlink")
	}
}

func TestRichMenuBulkLinkCmd_EmptyUserIDs(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	// Pass empty slice to trigger the "no user IDs" error
	cmd := newRichMenuBulkLinkCmdWithClient(client, []string{})
	cmd.SetArgs([]string{"--menu", "rm-123"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty user IDs")
	}
	if !strings.Contains(err.Error(), "no user IDs") {
		t.Errorf("expected 'no user IDs' error, got: %v", err)
	}
}

func TestRichMenuBulkUnlinkCmd_EmptyUserIDs(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	// Pass empty slice to trigger the "no user IDs" error
	cmd := newRichMenuBulkUnlinkCmdWithClient(client, []string{})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty user IDs")
	}
	if !strings.Contains(err.Error(), "no user IDs") {
		t.Errorf("expected 'no user IDs' error, got: %v", err)
	}
}

// Tests for download image with JPEG content type

func TestRichMenuDownloadImageCmd_JPEG(t *testing.T) {
	imageData := []byte("fake-jpeg-data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/bot/richmenu/rm-123/content" {
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write(imageData)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	tmpDir := t.TempDir()
	outputPath := tmpDir + "/test-image.jpg"

	oldOutput := flags.Output
	defer func() { flags.Output = oldOutput }()
	flags.Output = "text"

	cmd := newRichMenuDownloadImageCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "rm-123", "--output", outputPath})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Downloaded image") {
		t.Errorf("output should confirm download, got: %s", out.String())
	}
}

// Tests for list command API error

func TestRichMenuListCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "unauthorized"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

// Tests for readUserIDsFromFile

func TestReadUserIDsFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/users.txt"

	// Test with valid content including comments and blank lines
	content := `# This is a comment
U001
U002

# Another comment
U003
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	userIDs, err := readUserIDsFromFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(userIDs) != 3 {
		t.Errorf("expected 3 user IDs, got %d", len(userIDs))
	}

	expected := []string{"U001", "U002", "U003"}
	for i, id := range expected {
		if userIDs[i] != id {
			t.Errorf("expected userIDs[%d] to be %s, got %s", i, id, userIDs[i])
		}
	}
}

func TestReadUserIDsFromFile_NonExistent(t *testing.T) {
	_, err := readUserIDsFromFile("/nonexistent/path/users.txt")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestReadUserIDsFromFile_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/empty.txt"

	if err := os.WriteFile(filePath, []byte("# Only comments\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	userIDs, err := readUserIDsFromFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(userIDs) != 0 {
		t.Errorf("expected 0 user IDs, got %d", len(userIDs))
	}
}

// Tests for readBatchOperationsFromFile

func TestReadBatchOperationsFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/ops.json"

	content := `[
  {"type": "link", "richMenuId": "rm-123", "userIds": ["U001", "U002"]},
  {"type": "unlink", "userIds": ["U003"]}
]`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	ops, err := readBatchOperationsFromFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ops) != 2 {
		t.Errorf("expected 2 operations, got %d", len(ops))
	}

	if ops[0].Type != "link" {
		t.Errorf("expected first op type 'link', got %s", ops[0].Type)
	}
	if ops[1].Type != "unlink" {
		t.Errorf("expected second op type 'unlink', got %s", ops[1].Type)
	}
}

func TestReadBatchOperationsFromFile_NonExistent(t *testing.T) {
	_, err := readBatchOperationsFromFile("/nonexistent/path/ops.json")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestReadBatchOperationsFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/invalid.json"

	if err := os.WriteFile(filePath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := readBatchOperationsFromFile(filePath)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' error, got: %v", err)
	}
}

// Tests for download image with default filename (without --output flag)

func TestRichMenuDownloadImageCmd_DefaultFilename(t *testing.T) {
	imageData := []byte("fake-png-data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/bot/richmenu/rm-123/content" {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(imageData)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Use a temp directory as working directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	oldOutput := flags.Output
	defer func() { flags.Output = oldOutput }()
	flags.Output = "text"

	cmd := newRichMenuDownloadImageCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "rm-123"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that file was created with default name
	if _, err := os.Stat(tmpDir + "/rm-123.png"); os.IsNotExist(err) {
		t.Error("expected file rm-123.png to be created")
	}
}

// Test for list menus that include the default

func TestRichMenuListCmd_WithDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bot/richmenu/list":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"richmenus": []map[string]any{
					{
						"richMenuId":  "rm-123",
						"name":        "Main Menu",
						"chatBarText": "Menu",
						"size":        map[string]int{"width": 2500, "height": 1686},
						"areas":       []any{},
					},
					{
						"richMenuId":  "rm-456",
						"name":        "Secondary",
						"chatBarText": "Secondary",
						"size":        map[string]int{"width": 2500, "height": 843},
						"areas":       []any{},
					},
				},
			})
		case "/v2/bot/user/all/richmenu":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"richMenuId": "rm-123"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	defer func() { flags.Output = oldOutput }()
	flags.Output = "text"

	cmd := newRichMenuListCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Check that default is marked
	if !strings.Contains(output, "(default)") {
		t.Errorf("output should contain '(default)' marker, got: %s", output)
	}
	// Check that asterisk prefix is used for default
	if !strings.Contains(output, "* rm-123") {
		t.Errorf("output should have asterisk for default menu, got: %s", output)
	}
}

// Tests for set-default command

func TestRichMenuSetDefaultCmd_Execute(t *testing.T) {
	var receivedID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v2/bot/user/all/richmenu/") {
			receivedID = strings.TrimPrefix(r.URL.Path, "/v2/bot/user/all/richmenu/")
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
			checkText: "Set default rich menu: rm-123",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"defaultRichMenuId": "rm-123"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuSetDefaultCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "rm-123"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if receivedID != "rm-123" {
				t.Errorf("expected ID rm-123, got %s", receivedID)
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestRichMenuSetDefaultCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuSetDefaultCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "rm-nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent menu")
	}
}

// Tests for cancel-default command

func TestRichMenuCancelDefaultCmd_Execute(t *testing.T) {
	var cancelCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/v2/bot/user/all/richmenu" {
			cancelCalled = true
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
			checkText: "Cancelled default rich menu",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"defaultRichMenuId": null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cancelCalled = false
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuCancelDefaultCmdWithClient(client)
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !cancelCalled {
				t.Error("expected cancel endpoint to be called")
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestRichMenuCancelDefaultCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuCancelDefaultCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

// Tests for link command

func TestRichMenuLinkCmd_Execute(t *testing.T) {
	var receivedUserID, receivedMenuID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v2/bot/user/") {
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) >= 6 {
				receivedUserID = parts[4]
				receivedMenuID = parts[6]
			}
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
			checkText: "Linked rich menu rm-123 to user U456",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"status": "linked"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuLinkCmdWithClient(client)
			cmd.SetArgs([]string{"--user", "U456", "--id", "rm-123"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if receivedUserID != "U456" {
				t.Errorf("expected user ID U456, got %s", receivedUserID)
			}
			if receivedMenuID != "rm-123" {
				t.Errorf("expected menu ID rm-123, got %s", receivedMenuID)
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestRichMenuLinkCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuLinkCmdWithClient(client)
	cmd.SetArgs([]string{"--user", "U456", "--id", "rm-123"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for failed link")
	}
}

// Tests for unlink command

func TestRichMenuUnlinkCmd_Execute(t *testing.T) {
	var receivedUserID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/v2/bot/user/") {
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) >= 5 {
				receivedUserID = parts[4]
			}
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
			checkText: "Unlinked rich menu from user U789",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"status": "unlinked"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuUnlinkCmdWithClient(client)
			cmd.SetArgs([]string{"--user", "U789"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if receivedUserID != "U789" {
				t.Errorf("expected user ID U789, got %s", receivedUserID)
			}

			if !strings.Contains(out.String(), tt.checkText) {
				t.Errorf("output should contain %q, got: %s", tt.checkText, out.String())
			}
		})
	}
}

func TestRichMenuUnlinkCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "user not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuUnlinkCmdWithClient(client)
	cmd.SetArgs([]string{"--user", "U789"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

// Tests for batch status command

func TestRichMenuBatchStatusCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/bot/richmenu/progress/batch" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"phase":         "succeeded",
				"acceptedTime":  "2024-01-01T00:00:00Z",
				"completedTime": "2024-01-01T00:01:00Z",
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
			checkText: "Phase:          succeeded",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"phase": "succeeded"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuBatchStatusCmdWithClient(client)
			cmd.SetArgs([]string{"--request", "req-123"})
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

func TestRichMenuBatchStatusCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "request not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuBatchStatusCmdWithClient(client)
	cmd.SetArgs([]string{"--request", "nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent request")
	}
}

func TestRichMenuBatchStatusCmd_OngoingPhase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/bot/richmenu/progress/batch" {
			w.Header().Set("Content-Type", "application/json")
			// No completedTime for ongoing phase
			_ = json.NewEncoder(w).Encode(map[string]string{
				"phase":        "ongoing",
				"acceptedTime": "2024-01-01T00:00:00Z",
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
	flags.Output = "text"

	cmd := newRichMenuBatchStatusCmdWithClient(client)
	cmd.SetArgs([]string{"--request", "req-123"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "Completed Time:") {
		t.Error("output should NOT contain 'Completed Time:' for ongoing phase")
	}
}

// Tests for batch validate command

func TestRichMenuBatchValidateCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/bot/richmenu/validate/batch" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	operations := []api.RichMenuBatchOperation{
		{Type: "link", RichMenuID: "rm-123", UserIDs: []string{"U001", "U002"}},
		{Type: "unlink", UserIDs: []string{"U003"}},
	}

	tests := []struct {
		name      string
		output    string
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			checkText: "Batch operations valid (2 operations)",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"valid": true`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuBatchValidateCmdWithClient(client, operations)
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

func TestRichMenuBatchValidateCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid operations"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	operations := []api.RichMenuBatchOperation{
		{Type: "invalid", UserIDs: []string{"U001"}},
	}

	cmd := newRichMenuBatchValidateCmdWithClient(client, operations)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid operations")
	}
}

func TestRichMenuBatchValidateCmd_EmptyOperations(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	operations := []api.RichMenuBatchOperation{}

	cmd := newRichMenuBatchValidateCmdWithClient(client, operations)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty operations")
	}
	if !strings.Contains(err.Error(), "no operations") {
		t.Errorf("expected 'no operations' error, got: %v", err)
	}
}

// Tests for create command

func TestRichMenuCreateCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/bot/richmenu" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"richMenuId": "rm-created-123"})
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
			checkText: "Created rich menu: Test Menu (ID: rm-created-123)",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"richMenuId": "rm-created-123"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuCreateCmdWithClient(client)
			cmd.SetArgs([]string{"--name", "Test Menu", "--actions", `[{"type":"message","label":"Help","text":"help"}]`})
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

func TestRichMenuCreateCmd_CompactSize(t *testing.T) {
	var receivedHeight int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/bot/richmenu" {
			var req api.CreateRichMenuRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			receivedHeight = req.Size.Height
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"richMenuId": "rm-compact"})
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

	cmd := newRichMenuCreateCmdWithClient(client)
	cmd.SetArgs([]string{"--name", "Compact Menu", "--actions", `[{"type":"message"}]`, "--size", "compact"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedHeight != 843 {
		t.Errorf("expected compact height 843, got %d", receivedHeight)
	}
}

func TestRichMenuCreateCmd_InvalidActions(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newRichMenuCreateCmdWithClient(client)
	cmd.SetArgs([]string{"--name", "Test", "--actions", "not valid json"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid actions JSON")
	}
	if !strings.Contains(err.Error(), "invalid actions JSON") {
		t.Errorf("expected 'invalid actions JSON' error, got: %v", err)
	}
}

func TestRichMenuCreateCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid menu"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newRichMenuCreateCmdWithClient(client)
	cmd.SetArgs([]string{"--name", "Test", "--actions", `[{"type":"message"}]`})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for failed create")
	}
}

// Tests for upload-image command

func TestRichMenuUploadImageCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v2/bot/richmenu/") && strings.HasSuffix(r.URL.Path, "/content") {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	imageData := []byte("fake-image-data")

	tests := []struct {
		name      string
		output    string
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			checkText: "Image uploaded to rich menu: rm-123",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"status": "uploaded"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuUploadImageCmdWithClient(client, imageData)
			cmd.SetArgs([]string{"--id", "rm-123"})
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

func TestRichMenuUploadImageCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid image"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	imageData := []byte("fake-image-data")

	cmd := newRichMenuUploadImageCmdWithClient(client, imageData)
	cmd.SetArgs([]string{"--id", "rm-123"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for failed upload")
	}
}

// Tests for batch command

func TestRichMenuBatchCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/bot/richmenu/batch" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"requestId": "batch-req-123"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	operations := []api.RichMenuBatchOperation{
		{Type: "link", RichMenuID: "rm-123", UserIDs: []string{"U001", "U002"}},
		{Type: "unlink", UserIDs: []string{"U003"}},
	}

	tests := []struct {
		name      string
		output    string
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			checkText: "Batch submitted: batch-req-123 (2 operations)",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"requestId": "batch-req-123"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newRichMenuBatchCmdWithClient(client, operations)
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

func TestRichMenuBatchCmd_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "batch failed"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	operations := []api.RichMenuBatchOperation{
		{Type: "link", RichMenuID: "rm-123", UserIDs: []string{"U001"}},
	}

	cmd := newRichMenuBatchCmdWithClient(client, operations)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for failed batch")
	}
}

func TestRichMenuBatchCmd_EmptyOperations(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	operations := []api.RichMenuBatchOperation{}

	cmd := newRichMenuBatchCmdWithClient(client, operations)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty operations")
	}
	if !strings.Contains(err.Error(), "no operations") {
		t.Errorf("expected 'no operations' error, got: %v", err)
	}
}
