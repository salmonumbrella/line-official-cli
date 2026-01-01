package cmd

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestVersionCmd_Execute(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check for version info header
	if !strings.Contains(output, "line-cli") {
		t.Error("expected 'line-cli' in version output")
	}

	// Check for commit info
	if !strings.Contains(output, "commit:") {
		t.Error("expected 'commit:' in version output")
	}

	// Check for build date
	if !strings.Contains(output, "built:") {
		t.Error("expected 'built:' in version output")
	}

	// Check for Go version
	if !strings.Contains(output, "go:") {
		t.Error("expected 'go:' in version output")
	}
	if !strings.Contains(output, runtime.Version()) {
		t.Errorf("expected Go version %s in output", runtime.Version())
	}

	// Check for OS/arch info
	if !strings.Contains(output, "os/arch:") {
		t.Error("expected 'os/arch:' in version output")
	}
	if !strings.Contains(output, runtime.GOOS) {
		t.Errorf("expected OS %s in output", runtime.GOOS)
	}
	if !strings.Contains(output, runtime.GOARCH) {
		t.Errorf("expected arch %s in output", runtime.GOARCH)
	}
}

func TestVersionCmd_DefaultValues(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// When built without ldflags, should show dev defaults
	if !strings.Contains(output, "dev") && !strings.Contains(output, "v") {
		t.Error("expected version string (dev or semver) in output")
	}
}
