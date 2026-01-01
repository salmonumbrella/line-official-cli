package cmd

import (
	"strings"
	"testing"
)

func TestRequireExactlyOneFlag(t *testing.T) {
	tests := []struct {
		name        string
		checks      []FlagCheck
		wantErr     bool
		errContains string
	}{
		{
			name: "exactly one set - valid",
			checks: []FlagCheck{
				{Name: "--text", Set: true},
				{Name: "--flex", Set: false},
				{Name: "--image", Set: false},
			},
			wantErr: false,
		},
		{
			name: "second one set - valid",
			checks: []FlagCheck{
				{Name: "--text", Set: false},
				{Name: "--flex", Set: true},
				{Name: "--image", Set: false},
			},
			wantErr: false,
		},
		{
			name: "none set - error",
			checks: []FlagCheck{
				{Name: "--text", Set: false},
				{Name: "--flex", Set: false},
				{Name: "--image", Set: false},
			},
			wantErr:     true,
			errContains: "specify one of",
		},
		{
			name: "multiple set - error",
			checks: []FlagCheck{
				{Name: "--text", Set: true},
				{Name: "--flex", Set: true},
				{Name: "--image", Set: false},
			},
			wantErr:     true,
			errContains: "specify only one",
		},
		{
			name: "all set - error lists all",
			checks: []FlagCheck{
				{Name: "--text", Set: true},
				{Name: "--flex", Set: true},
				{Name: "--image", Set: true},
			},
			wantErr:     true,
			errContains: "--text",
		},
		{
			name:        "empty checks - error",
			checks:      []FlagCheck{},
			wantErr:     true,
			errContains: "specify one of",
		},
		{
			name: "single check set - valid",
			checks: []FlagCheck{
				{Name: "--only", Set: true},
			},
			wantErr: false,
		},
		{
			name: "single check not set - error",
			checks: []FlagCheck{
				{Name: "--only", Set: false},
			},
			wantErr:     true,
			errContains: "specify one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireExactlyOneFlag(tt.checks)
			if tt.wantErr {
				if err == nil {
					t.Errorf("requireExactlyOneFlag() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("requireExactlyOneFlag() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("requireExactlyOneFlag() unexpected error: %v", err)
				}
			}
		})
	}
}
