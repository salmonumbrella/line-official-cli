package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/salmonumbrella/line-official-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// mockSecretsStore is a mock implementation of secrets.Store for testing
type mockSecretsStore struct {
	accounts    map[string]secrets.Credentials
	accountMeta map[string]secrets.AccountInfo
	setErr      error
	getErr      error
	delErr      error
	listErr     error
	primaryErr  error
}

func newMockStore() *mockSecretsStore {
	return &mockSecretsStore{
		accounts:    make(map[string]secrets.Credentials),
		accountMeta: make(map[string]secrets.AccountInfo),
	}
}

func (m *mockSecretsStore) Set(name string, creds secrets.Credentials, botName string) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.accounts[name] = creds
	m.accountMeta[name] = secrets.AccountInfo{
		Name:    name,
		BotName: botName,
	}
	return nil
}

func (m *mockSecretsStore) Get(name string) (*secrets.Credentials, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	creds, ok := m.accounts[name]
	if !ok {
		return nil, errors.New("account not found")
	}
	return &creds, nil
}

func (m *mockSecretsStore) Delete(name string) error {
	if m.delErr != nil {
		return m.delErr
	}
	delete(m.accounts, name)
	return nil
}

func (m *mockSecretsStore) List() ([]secrets.AccountInfo, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	accounts := make([]secrets.AccountInfo, 0, len(m.accounts))
	for name := range m.accounts {
		if meta, ok := m.accountMeta[name]; ok {
			accounts = append(accounts, meta)
		} else {
			accounts = append(accounts, secrets.AccountInfo{Name: name})
		}
	}
	return accounts, nil
}

func (m *mockSecretsStore) SetPrimary(name string) error {
	if m.primaryErr != nil {
		return m.primaryErr
	}
	if _, ok := m.accounts[name]; !ok {
		return errors.New("account not found")
	}
	// Update all accounts
	for n, meta := range m.accountMeta {
		meta.IsPrimary = (n == name)
		m.accountMeta[n] = meta
	}
	return nil
}

func (m *mockSecretsStore) GetPrimary() (string, error) {
	if m.primaryErr != nil {
		return "", m.primaryErr
	}
	for name, meta := range m.accountMeta {
		if meta.IsPrimary {
			return name, nil
		}
	}
	// Return first account as fallback
	for name := range m.accounts {
		return name, nil
	}
	return "", nil
}

func TestAuthCmd_RequiresSubcommand(t *testing.T) {
	cmd := newAuthCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAuthCmd_HasSubcommands(t *testing.T) {
	cmd := newAuthCmd()
	subcommands := cmd.Commands()
	if len(subcommands) != 4 {
		t.Errorf("expected 4 subcommands, got %d", len(subcommands))
	}
	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}
	expected := []string{"login", "logout", "status", "list"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestAuthLoginCmd_Structure(t *testing.T) {
	cmd := newAuthLoginCmd()
	if cmd.Use != "login" {
		t.Errorf("expected Use 'login', got %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description")
	}
	if cmd.Long == "" {
		t.Error("expected Long description")
	}
	if cmd.Example == "" {
		t.Error("expected Example")
	}
}

func TestAuthLoginCmd_Flags(t *testing.T) {
	cmd := newAuthLoginCmd()
	tokenFlag := cmd.Flags().Lookup("token")
	if tokenFlag == nil {
		t.Fatal("expected --token flag")
	}
	if tokenFlag.Usage == "" {
		t.Error("expected --token flag to have usage text")
	}
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Fatal("expected --name flag")
	}
	if nameFlag.Usage == "" {
		t.Error("expected --name flag to have usage text")
	}
}

func TestAuthLogoutCmd_Structure(t *testing.T) {
	cmd := newAuthLogoutCmd()
	if cmd.Use != "logout" {
		t.Errorf("expected Use 'logout', got %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description")
	}
	if cmd.Long == "" {
		t.Error("expected Long description")
	}
}

func TestAuthLogoutCmd_Flags(t *testing.T) {
	cmd := newAuthLogoutCmd()
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("expected --name flag")
	}
}

func TestAuthStatusCmd_Structure(t *testing.T) {
	cmd := newAuthStatusCmd()
	if cmd.Use != "status" {
		t.Errorf("expected Use 'status', got %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description")
	}
	if cmd.Long == "" {
		t.Error("expected Long description")
	}
}

func TestAuthStatusCmd_NoNameFlag(t *testing.T) {
	cmd := newAuthStatusCmd()
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag != nil {
		t.Error("--name flag should be removed from status command")
	}
}

func TestAuthListCmd_Structure(t *testing.T) {
	cmd := newAuthListCmd()
	if cmd.Use != "list" {
		t.Errorf("expected Use 'list', got %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description")
	}
	if cmd.Long == "" {
		t.Error("expected Long description")
	}
}

func TestAuthCmd_HelpOutput(t *testing.T) {
	cmd := newAuthCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "login") {
		t.Error("expected help to mention 'login'")
	}
	if !strings.Contains(output, "logout") {
		t.Error("expected help to mention 'logout'")
	}
	if !strings.Contains(output, "status") {
		t.Error("expected help to mention 'status'")
	}
	if !strings.Contains(output, "list") {
		t.Error("expected help to mention 'list'")
	}
}

func TestAuthLoginCmd_HelpOutput(t *testing.T) {
	cmd := newAuthLoginCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "--token") {
		t.Error("expected help to mention '--token'")
	}
	if !strings.Contains(output, "--name") {
		t.Error("expected help to mention '--name'")
	}
}

func TestAuthSubcommands_Metadata(t *testing.T) {
	tests := []struct {
		name            string
		cmdFunc         func() *cobra.Command
		expectedUse     string
		expectedHasLong bool
	}{
		{"login", newAuthLoginCmd, "login", true},
		{"logout", newAuthLogoutCmd, "logout", true},
		{"status", newAuthStatusCmd, "status", true},
		{"list", newAuthListCmd, "list", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFunc()
			if cmd.Use != tt.expectedUse {
				t.Errorf("expected Use '%s', got '%s'", tt.expectedUse, cmd.Use)
			}
			if tt.expectedHasLong && cmd.Long == "" {
				t.Error("expected Long description to be non-empty")
			}
			if cmd.RunE == nil {
				t.Error("expected RunE to be defined")
			}
		})
	}
}

func TestAuthLoginCmd_WithToken_Success(t *testing.T) {
	store := newMockStore()
	cmd := newAuthLoginCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--token", "test-token-123", "--name", "my-account"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Logged in as my-account") {
		t.Errorf("expected 'Logged in as my-account' in output, got: %s", output)
	}
	creds, err := store.Get("my-account")
	if err != nil {
		t.Fatalf("expected credentials to be stored: %v", err)
	}
	if creds.ChannelAccessToken != "test-token-123" {
		t.Errorf("expected token 'test-token-123', got: %s", creds.ChannelAccessToken)
	}
}

func TestAuthLoginCmd_WithToken_DefaultName(t *testing.T) {
	store := newMockStore()
	cmd := newAuthLoginCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--token", "test-token-123"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Logged in as default") {
		t.Errorf("expected 'Logged in as default' in output, got: %s", output)
	}
	creds, err := store.Get("default")
	if err != nil {
		t.Fatalf("expected credentials to be stored: %v", err)
	}
	if creds.ChannelAccessToken != "test-token-123" {
		t.Errorf("expected token 'test-token-123', got: %s", creds.ChannelAccessToken)
	}
}

func TestAuthLoginCmd_WithToken_StoreError(t *testing.T) {
	store := newMockStore()
	store.setErr = errors.New("keychain locked")
	cmd := newAuthLoginCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--token", "test-token-123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for store failure")
	}
	if !strings.Contains(err.Error(), "failed to save credentials") {
		t.Errorf("expected 'failed to save credentials' in error, got: %v", err)
	}
}

func TestAuthLogoutCmd_Success(t *testing.T) {
	store := newMockStore()
	_ = store.Set("my-account", secrets.Credentials{ChannelAccessToken: "token123"}, "")
	cmd := newAuthLogoutCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--name", "my-account"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Logged out: my-account") {
		t.Errorf("expected 'Logged out: my-account' in output, got: %s", output)
	}
	_, err = store.Get("my-account")
	if err == nil {
		t.Error("expected account to be deleted")
	}
}

func TestAuthLogoutCmd_DefaultName(t *testing.T) {
	store := newMockStore()
	_ = store.Set("default", secrets.Credentials{ChannelAccessToken: "token123"}, "")
	cmd := newAuthLogoutCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Logged out: default") {
		t.Errorf("expected 'Logged out: default' in output, got: %s", output)
	}
}

func TestAuthLogoutCmd_DeleteError(t *testing.T) {
	store := newMockStore()
	store.delErr = errors.New("permission denied")
	cmd := newAuthLogoutCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--name", "my-account"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for delete failure")
	}
	if !strings.Contains(err.Error(), "failed to remove credentials") {
		t.Errorf("expected 'failed to remove credentials' in error, got: %v", err)
	}
}

func TestAuthStatusCmd_ShowsActiveAccount(t *testing.T) {
	store := newMockStore()
	_ = store.Set("my-account", secrets.Credentials{ChannelAccessToken: "abcdefgh12345678xyz9"}, "My Bot")
	cmd := newAuthStatusCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Active account: my-account") {
		t.Errorf("expected 'Active account: my-account' in output, got: %s", output)
	}
	if !strings.Contains(output, "All accounts:") {
		t.Errorf("expected 'All accounts:' in output, got: %s", output)
	}
	if !strings.Contains(output, "my-account") {
		t.Errorf("expected 'my-account' in output, got: %s", output)
	}
}

func TestAuthStatusCmd_ShowsPrimaryAccount(t *testing.T) {
	store := newMockStore()
	_ = store.Set("my-account", secrets.Credentials{ChannelAccessToken: "abcdefgh12345678xyz9"}, "My Bot")
	_ = store.SetPrimary("my-account")
	cmd := newAuthStatusCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Active account: my-account (primary)") {
		t.Errorf("expected 'Active account: my-account (primary)' in output, got: %s", output)
	}
}

func TestAuthStatusCmd_ShowsFirstAccountFallback(t *testing.T) {
	store := newMockStore()
	_ = store.Set("first-account", secrets.Credentials{ChannelAccessToken: "token1"}, "")
	cmd := newAuthStatusCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "(first account)") {
		t.Errorf("expected '(first account)' source indicator in output, got: %s", output)
	}
}

func TestAuthStatusCmd_ShowsAccountFromFlag(t *testing.T) {
	store := newMockStore()
	_ = store.Set("account-a", secrets.Credentials{ChannelAccessToken: "token1"}, "")
	_ = store.Set("account-b", secrets.Credentials{ChannelAccessToken: "token2"}, "")

	// Save and set the global flag
	oldAccount := flags.Account
	flags.Account = "account-b"
	defer func() { flags.Account = oldAccount }()

	cmd := newAuthStatusCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Active account: account-b") {
		t.Errorf("expected 'Active account: account-b' in output, got: %s", output)
	}
	if !strings.Contains(output, "(from --account flag or LINE_ACCOUNT env)") {
		t.Errorf("expected flag source indicator in output, got: %s", output)
	}
}

func TestAuthStatusCmd_NotLoggedIn(t *testing.T) {
	store := newMockStore()
	cmd := newAuthStatusCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Not logged in") {
		t.Errorf("expected 'Not logged in' in output, got: %s", output)
	}
	if !strings.Contains(output, "Run: line auth login") {
		t.Errorf("expected hint in output, got: %s", output)
	}
}

func TestAuthStatusCmd_ShowsBotName(t *testing.T) {
	store := newMockStore()
	_ = store.Set("my-account", secrets.Credentials{ChannelAccessToken: "token123"}, "My Super Bot")
	cmd := newAuthStatusCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "My Super Bot") {
		t.Errorf("expected bot name 'My Super Bot' in output, got: %s", output)
	}
}

func TestAuthStatusCmd_MultipleAccounts(t *testing.T) {
	store := newMockStore()
	_ = store.Set("account-1", secrets.Credentials{ChannelAccessToken: "token1"}, "Bot 1")
	_ = store.Set("account-2", secrets.Credentials{ChannelAccessToken: "token2"}, "Bot 2")
	_ = store.SetPrimary("account-1")
	cmd := newAuthStatusCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "account-1") {
		t.Errorf("expected 'account-1' in output, got: %s", output)
	}
	if !strings.Contains(output, "account-2") {
		t.Errorf("expected 'account-2' in output, got: %s", output)
	}
	if !strings.Contains(output, "* account-1") {
		t.Errorf("expected active account marker '* account-1' in output, got: %s", output)
	}
}

func TestAuthListCmd_NoAccounts(t *testing.T) {
	store := newMockStore()
	cmd := newAuthListCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "No accounts configured") {
		t.Errorf("expected 'No accounts configured' in output, got: %s", output)
	}
	if !strings.Contains(output, "Run: line auth login") {
		t.Errorf("expected hint in output, got: %s", output)
	}
}

func TestAuthListCmd_OneAccount(t *testing.T) {
	store := newMockStore()
	_ = store.Set("my-account", secrets.Credentials{ChannelAccessToken: "token123"}, "")
	cmd := newAuthListCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Configured accounts:") {
		t.Errorf("expected 'Configured accounts:' in output, got: %s", output)
	}
	if !strings.Contains(output, "my-account") {
		t.Errorf("expected 'my-account' in output, got: %s", output)
	}
}

func TestAuthListCmd_MultipleAccounts(t *testing.T) {
	store := newMockStore()
	_ = store.Set("account-1", secrets.Credentials{ChannelAccessToken: "token1"}, "")
	_ = store.Set("account-2", secrets.Credentials{ChannelAccessToken: "token2"}, "")
	_ = store.Set("account-3", secrets.Credentials{ChannelAccessToken: "token3"}, "")
	cmd := newAuthListCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Configured accounts:") {
		t.Errorf("expected 'Configured accounts:' in output, got: %s", output)
	}
	if !strings.Contains(output, "account-1") {
		t.Errorf("expected 'account-1' in output, got: %s", output)
	}
	if !strings.Contains(output, "account-2") {
		t.Errorf("expected 'account-2' in output, got: %s", output)
	}
	if !strings.Contains(output, "account-3") {
		t.Errorf("expected 'account-3' in output, got: %s", output)
	}
}

func TestAuthListCmd_TableOutput(t *testing.T) {
	store := newMockStore()
	_ = store.Set("account-1", secrets.Credentials{ChannelAccessToken: "token1"}, "")
	_ = store.Set("account-2", secrets.Credentials{ChannelAccessToken: "token2"}, "")
	oldOutput := flags.Output
	flags.Output = "table"
	defer func() { flags.Output = oldOutput }()
	cmd := newAuthListCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "ACCOUNT") {
		t.Errorf("expected 'ACCOUNT' header in table output, got: %s", output)
	}
	if !strings.Contains(output, "account-1") {
		t.Errorf("expected 'account-1' in output, got: %s", output)
	}
}

func TestAuthListCmd_JsonOutput(t *testing.T) {
	store := newMockStore()
	_ = store.Set("account-1", secrets.Credentials{ChannelAccessToken: "token1"}, "My Bot")
	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()
	cmd := newAuthListCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	// JSON output should contain the account name
	if !strings.Contains(output, "account-1") {
		t.Errorf("expected 'account-1' in json output, got: %s", output)
	}
	// JSON output should contain bot name
	if !strings.Contains(output, "My Bot") {
		t.Errorf("expected 'My Bot' in json output, got: %s", output)
	}
	// Should be valid JSON (starts with [ for array)
	if !strings.HasPrefix(strings.TrimSpace(output), "[") {
		t.Errorf("expected JSON array output, got: %s", output)
	}
}

func TestAuthListCmd_ListError(t *testing.T) {
	store := newMockStore()
	store.listErr = errors.New("keychain unavailable")
	cmd := newAuthListCmdWithStore(store)
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for list failure")
	}
	if !strings.Contains(err.Error(), "failed to list accounts") {
		t.Errorf("expected 'failed to list accounts' in error, got: %v", err)
	}
}
