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

func TestInsightCmd_RequiresSubcommand(t *testing.T) {
	cmd := newInsightCmd()

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

func TestInsightCmd_HasSubcommands(t *testing.T) {
	cmd := newInsightCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 5 {
		t.Errorf("expected at least 5 subcommands (followers, messages, demographics, events, unit-stats), got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expected := []string{"followers", "messages", "demographics", "events", "unit-stats"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestInsightFollowersCmd_HasDateFlag(t *testing.T) {
	cmd := newInsightFollowersCmd()

	// Check --date flag exists (optional, defaults to yesterday)
	dateFlag := cmd.Flags().Lookup("date")
	if dateFlag == nil {
		t.Error("expected --date flag for followers command")
	}
}

func TestInsightMessagesCmd_HasDateFlag(t *testing.T) {
	cmd := newInsightMessagesCmd()

	// Check --date flag exists (optional, defaults to yesterday)
	dateFlag := cmd.Flags().Lookup("date")
	if dateFlag == nil {
		t.Error("expected --date flag for messages command")
	}
}

func TestInsightEventsCmd_RequiresRequestID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"insight", "events"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --request-id flag")
	}
}

func TestInsightUnitStatsCmd_RequiresUnit(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"insight", "unit-stats", "--from", "20251224", "--to", "20251231"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --unit flag")
	}
}

func TestInsightUnitStatsCmd_RequiresFrom(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"insight", "unit-stats", "--unit", "campaign-2024", "--to", "20251231"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --from flag")
	}
}

func TestInsightUnitStatsCmd_RequiresTo(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"insight", "unit-stats", "--unit", "campaign-2024", "--from", "20251224"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --to flag")
	}
}

func TestInsightUnitStatsCmd_Flags(t *testing.T) {
	cmd := newInsightUnitStatsCmd()

	// Check all required flags exist
	flagNames := []string{"unit", "from", "to"}
	for _, flagName := range flagNames {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("expected --%s flag", flagName)
		}
	}
}

// Execution tests using mock servers

func TestInsightFollowersCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/insight/followers") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":          "ready",
				"followers":       1000,
				"targetedReaches": 800,
				"blocks":          50,
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
			checkText: "Follower Stats",
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

			cmd := newInsightFollowersCmdWithClient(client)
			cmd.SetArgs([]string{"--date", "20251230"})
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
				if result["status"] != "ready" {
					t.Errorf("expected status 'ready', got: %v", result["status"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestInsightFollowersCmd_InvalidDate(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newInsightFollowersCmdWithClient(client)
	cmd.SetArgs([]string{"--date", "invalid"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid date format")
	}
	if !strings.Contains(err.Error(), "YYYYMMDD") {
		t.Errorf("expected YYYYMMDD format error, got: %v", err)
	}
}

func TestInsightFollowersCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid token"})
	}))
	defer server.Close()

	client := api.NewClient("bad-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newInsightFollowersCmdWithClient(client)
	cmd.SetArgs([]string{"--date", "20251230"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get follower stats") {
		t.Errorf("expected 'failed to get follower stats' in error, got: %v", err)
	}
}

func TestInsightMessagesCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/insight/message/delivery") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":    "ready",
				"broadcast": int64(100),
				"targeting": int64(200),
				"chat":      int64(50),
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
			checkText: "Message Delivery Stats",
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

			cmd := newInsightMessagesCmdWithClient(client)
			cmd.SetArgs([]string{"--date", "20251230"})
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
				if result["status"] != "ready" {
					t.Errorf("expected status 'ready', got: %v", result["status"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestInsightMessagesCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized"})
	}))
	defer server.Close()

	client := api.NewClient("bad-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newInsightMessagesCmdWithClient(client)
	cmd.SetArgs([]string{"--date", "20251230"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get message stats") {
		t.Errorf("expected 'failed to get message stats' in error, got: %v", err)
	}
}

func TestInsightDemographicsCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/insight/demographic" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"available": true,
				"genders": []map[string]any{
					{"gender": "male", "percentage": 45.0},
					{"gender": "female", "percentage": 55.0},
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
			checkText: "Friend Demographics",
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

			cmd := newInsightDemographicsCmdWithClient(client)
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
				if result["available"] != true {
					t.Errorf("expected available true, got: %v", result["available"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestInsightDemographicsCmd_NotAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/insight/demographic" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"available": false,
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

	cmd := newInsightDemographicsCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "not available") {
		t.Errorf("expected 'not available' message, got: %s", output)
	}
}

func TestInsightDemographicsCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized"})
	}))
	defer server.Close()

	client := api.NewClient("bad-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newInsightDemographicsCmdWithClient(client)
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get demographics") {
		t.Errorf("expected 'failed to get demographics' in error, got: %v", err)
	}
}

func TestInsightEventsCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/insight/message/event") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"overview": map[string]any{
					"requestId":        "abc123",
					"timestamp":        1704067200000,
					"delivered":        1000,
					"uniqueImpression": 800,
					"uniqueClick":      100,
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
			checkText: "Message Event Statistics",
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

			cmd := newInsightEventsCmdWithClient(client)
			cmd.SetArgs([]string{"--request-id", "abc123"})
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
				overview, ok := result["overview"].(map[string]any)
				if !ok {
					t.Errorf("expected overview in response, got: %v", result)
				} else if overview["delivered"].(float64) != 1000 {
					t.Errorf("expected delivered 1000, got: %v", overview["delivered"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestInsightEventsCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newInsightEventsCmdWithClient(client)
	cmd.SetArgs([]string{"--request-id", "nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get event stats") {
		t.Errorf("expected 'failed to get event stats' in error, got: %v", err)
	}
}

func TestInsightUnitStatsCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/insight/message/event/aggregation") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"overview": map[string]any{
					"uniqueImpression":            int64(5000),
					"uniqueClick":                 int64(500),
					"uniqueMediaPlayed":           int64(200),
					"uniqueMediaPlayed100Percent": int64(50),
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
			checkText: "Statistics for unit",
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

			cmd := newInsightUnitStatsCmdWithClient(client)
			cmd.SetArgs([]string{"--unit", "campaign-2024", "--from", "20251224", "--to", "20251231"})
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
				overview, ok := result["overview"].(map[string]any)
				if !ok {
					t.Errorf("expected overview in response, got: %v", result)
				} else if overview["uniqueImpression"].(float64) != 5000 {
					t.Errorf("expected uniqueImpression 5000, got: %v", overview["uniqueImpression"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestInsightUnitStatsCmd_InvalidDateFormat(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "invalid from date",
			args:    []string{"--unit", "test", "--from", "2025-12-24", "--to", "20251231"},
			wantErr: "--from must be in YYYYMMDD",
		},
		{
			name:    "invalid to date",
			args:    []string{"--unit", "test", "--from", "20251224", "--to", "invalid"},
			wantErr: "--to must be in YYYYMMDD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newInsightUnitStatsCmdWithClient(client)
			cmd.SetArgs(tt.args)
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err == nil {
				t.Error("expected error for invalid date format")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestInsightUnitStatsCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid unit"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newInsightUnitStatsCmdWithClient(client)
	cmd.SetArgs([]string{"--unit", "invalid", "--from", "20251224", "--to", "20251231"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get unit statistics") {
		t.Errorf("expected 'failed to get unit statistics' in error, got: %v", err)
	}
}
