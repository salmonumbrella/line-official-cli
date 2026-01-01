package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestMessageCmd_RequiresSubcommand(t *testing.T) {
	cmd := newMessageCmd()

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

func TestMessagePushCmd_RequiresTo(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"message", "push", "--text", "hello"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --to flag")
	}
}

func TestMessageAggregationCmd_HasSubcommands(t *testing.T) {
	cmd := newMessageAggregationCmd()

	// Check that subcommands exist
	subcommands := cmd.Commands()
	if len(subcommands) != 2 {
		t.Errorf("expected 2 subcommands, got %d", len(subcommands))
	}

	// Verify subcommand names
	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	if !names["usage"] {
		t.Error("expected 'usage' subcommand")
	}
	if !names["list"] {
		t.Error("expected 'list' subcommand")
	}
}

func TestMessageAggregationListCmd_Flags(t *testing.T) {
	cmd := newMessageAggregationListCmd()

	// Check limit flag exists
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("expected --limit flag")
	}

	// Check start flag exists
	startFlag := cmd.Flags().Lookup("start")
	if startFlag == nil {
		t.Error("expected --start flag")
	}
}

// Edge case tests for helper functions

// Tests for formatMessageOutput coverage
func TestFormatMessageOutput_TextOutput_AllTargetTypes(t *testing.T) {
	tests := []struct {
		name        string
		target      messageTarget
		msgType     string
		extraFields map[string]any
		wantOutput  string
	}{
		{
			name:       "push text message",
			target:     messageTarget{Type: "push", UserID: "U12345"},
			msgType:    "text",
			wantOutput: "Message sent to U12345",
		},
		{
			name:       "push image message",
			target:     messageTarget{Type: "push", UserID: "U12345"},
			msgType:    "image",
			wantOutput: "Image sent to U12345",
		},
		{
			name:       "broadcast text message",
			target:     messageTarget{Type: "broadcast"},
			msgType:    "text",
			wantOutput: "Broadcast sent",
		},
		{
			name:       "broadcast flex message",
			target:     messageTarget{Type: "broadcast"},
			msgType:    "flex",
			wantOutput: "Flex broadcast sent",
		},
		{
			name:       "multicast text message",
			target:     messageTarget{Type: "multicast", UserIDs: []string{"U1", "U2", "U3"}},
			msgType:    "text",
			wantOutput: "Message sent to 3 users",
		},
		{
			name:       "multicast sticker message",
			target:     messageTarget{Type: "multicast", UserIDs: []string{"U1", "U2"}},
			msgType:    "sticker",
			wantOutput: "Sticker sent to 2 users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global flags
			oldOutput := flags.Output
			flags.Output = "text"
			defer func() { flags.Output = oldOutput }()

			cmd := newMessageCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := formatMessageOutput(cmd, tt.target, tt.msgType, tt.extraFields)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out.String(), tt.wantOutput) {
				t.Errorf("expected output to contain %q, got %q", tt.wantOutput, out.String())
			}
		})
	}
}

func TestFormatMessageOutput_JSONOutput_WithExtraFields(t *testing.T) {
	tests := []struct {
		name        string
		target      messageTarget
		msgType     string
		extraFields map[string]any
		wantKey     string
		wantValue   any
	}{
		{
			name:        "push with extra fields",
			target:      messageTarget{Type: "push", UserID: "U123"},
			msgType:     "audio",
			extraFields: map[string]any{"duration": 60000},
			wantKey:     "duration",
			wantValue:   float64(60000),
		},
		{
			name:        "multicast JSON output",
			target:      messageTarget{Type: "multicast", UserIDs: []string{"U1", "U2", "U3", "U4", "U5"}},
			msgType:     "location",
			extraFields: map[string]any{"title": "Test Place", "lat": 35.6, "lng": 139.7},
			wantKey:     "recipients",
			wantValue:   float64(5),
		},
		{
			name:    "broadcast JSON output",
			target:  messageTarget{Type: "broadcast"},
			msgType: "video",
			wantKey: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = "json"
			defer func() { flags.Output = oldOutput }()

			cmd := newMessageCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := formatMessageOutput(cmd, tt.target, tt.msgType, tt.extraFields)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var result map[string]any
			if err := json.Unmarshal(out.Bytes(), &result); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			if result["type"] != tt.msgType {
				t.Errorf("expected type=%q, got %v", tt.msgType, result["type"])
			}

			if tt.wantValue != nil {
				if result[tt.wantKey] != tt.wantValue {
					t.Errorf("expected %s=%v, got %v", tt.wantKey, tt.wantValue, result[tt.wantKey])
				}
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "lowercase word",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "message type text",
			input:    "text",
			expected: "Text",
		},
		{
			name:     "message type image",
			input:    "image",
			expected: "Image",
		},
		{
			name:     "message type flex",
			input:    "flex",
			expected: "Flex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := capitalize(tt.input)
			if result != tt.expected {
				t.Errorf("capitalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
