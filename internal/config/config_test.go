package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NoConfigFile(t *testing.T) {
	// Set XDG_CONFIG_HOME to a non-existent directory
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Also override HOME to prevent loading ~/.line-cli.yaml
	t.Setenv("HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should return empty config
	if cfg.Account != "" {
		t.Errorf("Account = %q, want empty", cfg.Account)
	}
	if cfg.Output != "" {
		t.Errorf("Output = %q, want empty", cfg.Output)
	}
	if cfg.Debug {
		t.Error("Debug = true, want false")
	}
	if cfg.ConfigPath() != "" {
		t.Errorf("ConfigPath() = %q, want empty", cfg.ConfigPath())
	}
}

func TestLoad_XDGConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", t.TempDir()) // Different dir to ensure XDG takes precedence

	// Create config file in XDG location
	configDir := filepath.Join(tmpDir, AppName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	content := `account: test-account
output: json
debug: true
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Account != "test-account" {
		t.Errorf("Account = %q, want %q", cfg.Account, "test-account")
	}
	if cfg.Output != "json" {
		t.Errorf("Output = %q, want %q", cfg.Output, "json")
	}
	if !cfg.Debug {
		t.Error("Debug = false, want true")
	}
	if cfg.ConfigPath() != configPath {
		t.Errorf("ConfigPath() = %q, want %q", cfg.ConfigPath(), configPath)
	}
}

func TestLoad_FallbackConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	// Clear XDG_CONFIG_HOME to test standard ~/.config fallback
	t.Setenv("XDG_CONFIG_HOME", "")

	// Create config file in ~/.config/line-cli/
	configDir := filepath.Join(tmpDir, ".config", AppName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	content := `account: fallback-account
output: text
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Account != "fallback-account" {
		t.Errorf("Account = %q, want %q", cfg.Account, "fallback-account")
	}
	if cfg.Output != "text" {
		t.Errorf("Output = %q, want %q", cfg.Output, "text")
	}
	if cfg.ConfigPath() != configPath {
		t.Errorf("ConfigPath() = %q, want %q", cfg.ConfigPath(), configPath)
	}
}

func TestLoad_HomeDotFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", "")

	// Create ~/.line-cli.yaml
	configPath := filepath.Join(tmpDir, ".line-cli.yaml")
	content := `account: home-account
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Account != "home-account" {
		t.Errorf("Account = %q, want %q", cfg.Account, "home-account")
	}
	if cfg.ConfigPath() != configPath {
		t.Errorf("ConfigPath() = %q, want %q", cfg.ConfigPath(), configPath)
	}
}

func TestLoad_XDGTakesPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create both XDG and home dotfile
	xdgDir := filepath.Join(tmpDir, AppName)
	if err := os.MkdirAll(xdgDir, 0755); err != nil {
		t.Fatal(err)
	}
	xdgPath := filepath.Join(xdgDir, "config.yaml")
	if err := os.WriteFile(xdgPath, []byte("account: xdg-account\n"), 0644); err != nil {
		t.Fatal(err)
	}

	homePath := filepath.Join(tmpDir, ".line-cli.yaml")
	if err := os.WriteFile(homePath, []byte("account: home-account\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// XDG should take precedence
	if cfg.Account != "xdg-account" {
		t.Errorf("Account = %q, want %q", cfg.Account, "xdg-account")
	}
	if cfg.ConfigPath() != xdgPath {
		t.Errorf("ConfigPath() = %q, want %q", cfg.ConfigPath(), xdgPath)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", t.TempDir())

	configDir := filepath.Join(tmpDir, AppName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	// Invalid YAML
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for invalid YAML")
	}
}

func TestLoad_PartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", t.TempDir())

	configDir := filepath.Join(tmpDir, AppName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Only set some values
	configPath := filepath.Join(configDir, "config.yaml")
	content := `account: partial-account
# output not set
# debug not set
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Account != "partial-account" {
		t.Errorf("Account = %q, want %q", cfg.Account, "partial-account")
	}
	if cfg.Output != "" {
		t.Errorf("Output = %q, want empty", cfg.Output)
	}
	if cfg.Debug {
		t.Error("Debug = true, want false")
	}
}

func TestExampleConfig(t *testing.T) {
	example := ExampleConfig()
	if example == "" {
		t.Error("ExampleConfig() returned empty string")
	}
	// Should contain key config options
	if !contains(example, "account") {
		t.Error("ExampleConfig() should mention 'account'")
	}
	if !contains(example, "output") {
		t.Error("ExampleConfig() should mention 'output'")
	}
	if !contains(example, "debug") {
		t.Error("ExampleConfig() should mention 'debug'")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath() error = %v", err)
	}
	if path == "" {
		t.Error("DefaultConfigPath() returned empty string")
	}
	if !contains(path, AppName) {
		t.Errorf("DefaultConfigPath() = %q, should contain %q", path, AppName)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
