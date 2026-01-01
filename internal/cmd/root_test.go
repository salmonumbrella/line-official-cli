package cmd

import (
	"bytes"
	"context"
	"testing"
)

func TestGetDefault(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "first non-empty",
			values:   []string{"first", "second", "third"},
			expected: "first",
		},
		{
			name:     "second when first empty",
			values:   []string{"", "second", "third"},
			expected: "second",
		},
		{
			name:     "third when first two empty",
			values:   []string{"", "", "third"},
			expected: "third",
		},
		{
			name:     "all empty returns empty",
			values:   []string{"", "", ""},
			expected: "",
		},
		{
			name:     "empty slice returns empty",
			values:   []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDefault(tt.values...)
			if result != tt.expected {
				t.Errorf("getDefault(%v) = %q, want %q", tt.values, result, tt.expected)
			}
		})
	}
}

func TestGetDefaultBool(t *testing.T) {
	tests := []struct {
		name     string
		cfgVal   bool
		fallback bool
		expected bool
	}{
		{
			name:     "cfgVal true returns true",
			cfgVal:   true,
			fallback: false,
			expected: true,
		},
		{
			name:     "cfgVal true overrides fallback true",
			cfgVal:   true,
			fallback: true,
			expected: true,
		},
		{
			name:     "cfgVal false returns fallback false",
			cfgVal:   false,
			fallback: false,
			expected: false,
		},
		{
			name:     "cfgVal false returns fallback true",
			cfgVal:   false,
			fallback: true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDefaultBool(tt.cfgVal, tt.fallback)
			if result != tt.expected {
				t.Errorf("getDefaultBool(%v, %v) = %v, want %v", tt.cfgVal, tt.fallback, result, tt.expected)
			}
		})
	}
}

func TestRequireAccount_ExplicitFlag(t *testing.T) {
	f := &rootFlags{Account: "my-account"}

	account, err := requireAccount(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account != "my-account" {
		t.Errorf("expected account='my-account', got %q", account)
	}
}

func TestNewRootCmd_HasSubcommands(t *testing.T) {
	cmd := NewRootCmd()

	// Check that key subcommands exist
	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("expected subcommands, got none")
	}

	// Verify some key subcommand names
	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expectedCommands := []string{"message", "auth", "config", "version", "completion"}
	for _, expected := range expectedCommands {
		if !names[expected] {
			t.Errorf("expected %q subcommand", expected)
		}
	}
}

func TestNewRootCmd_FlagsExist(t *testing.T) {
	cmd := NewRootCmd()

	// Check persistent flags
	accountFlag := cmd.PersistentFlags().Lookup("account")
	if accountFlag == nil {
		t.Error("expected --account flag")
	}

	outputFlag := cmd.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Error("expected --output flag")
	}

	debugFlag := cmd.PersistentFlags().Lookup("debug")
	if debugFlag == nil {
		t.Error("expected --debug flag")
	}

	dryRunFlag := cmd.PersistentFlags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("expected --dry-run flag")
	}

	yesFlag := cmd.PersistentFlags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("expected --yes flag")
	}

	// Check yes has short flag
	if yesFlag.Shorthand != "y" {
		t.Errorf("expected --yes shorthand to be 'y', got %q", yesFlag.Shorthand)
	}
}

func TestExecute_HelpCommand(t *testing.T) {
	err := Execute([]string{"--help"})
	// --help exits with nil error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteContext_HelpCommand(t *testing.T) {
	ctx := context.Background()
	err := ExecuteContext(ctx, []string{"--help"})
	// --help exits with nil error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecute_VersionCommand(t *testing.T) {
	err := Execute([]string{"version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteContext_VersionCommand(t *testing.T) {
	ctx := context.Background()
	err := ExecuteContext(ctx, []string{"version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRootCmd_Execute(t *testing.T) {
	cmd := NewRootCmd()

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
