package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigCmd_Execute(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config"})

	// Note: runConfig() writes directly to stdout via fmt.Println,
	// so we only verify the command executes without error.
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigShowCmd_Execute(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "show"})

	// Note: runConfig() writes directly to stdout via fmt.Println,
	// so we only verify the command executes without error.
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigPathCmd_Execute(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "path"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should output path information (either Loaded or Recommended)
	if !strings.Contains(output, "Loaded") && !strings.Contains(output, "Recommended") {
		t.Error("expected path information in output")
	}
}

func TestConfigPathCmd_JSONOutput(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--output", "json", "config", "path"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// JSON output should contain recommended path
	if !strings.Contains(output, "recommended") {
		t.Error("expected 'recommended' key in JSON output")
	}
	if !strings.Contains(output, "{") {
		t.Error("expected JSON object in output")
	}
}

func TestConfigExampleCmd_Execute(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "example"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Example config should contain YAML comments and field examples
	if !strings.Contains(output, "LINE CLI configuration") {
		t.Error("expected config header comment in output")
	}
	if !strings.Contains(output, "account:") {
		t.Error("expected 'account:' field in example config")
	}
	if !strings.Contains(output, "output:") {
		t.Error("expected 'output:' field in example config")
	}
}

func TestConfigCmd_JSONOutput(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--output", "json", "config"})

	// Note: runConfig() writes directly to stdout via fmt.Println,
	// so we only verify the command executes without error.
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigShowCmd_JSONOutput(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--output", "json", "config", "show"})

	// Note: runConfig() writes directly to stdout via fmt.Println,
	// so we only verify the command executes without error.
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigCmd_HasSubcommands(t *testing.T) {
	cmd := newConfigCmd()

	subcommands := cmd.Commands()
	if len(subcommands) != 3 {
		t.Errorf("expected 3 subcommands, got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	if !names["show"] {
		t.Error("expected 'show' subcommand")
	}
	if !names["path"] {
		t.Error("expected 'path' subcommand")
	}
	if !names["example"] {
		t.Error("expected 'example' subcommand")
	}
}
