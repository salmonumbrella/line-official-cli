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

func TestCouponCmd_RequiresSubcommand(t *testing.T) {
	cmd := newCouponCmd()

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

func TestCouponCmd_HasSubcommands(t *testing.T) {
	cmd := newCouponCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 4 {
		t.Errorf("expected at least 4 subcommands (list, create, get, close), got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expected := []string{"list", "create", "get", "close"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestCouponGetCmd_RequiresCouponID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"coupon", "get"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

func TestCouponCloseCmd_RequiresCouponID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"coupon", "close"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --id flag")
	}
}

// Execution tests using mock servers

func TestCouponListCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/coupon" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"couponId": "coupon-001", "title": "Summer Sale", "status": "RUNNING"},
					{"couponId": "coupon-002", "title": "Winter Deal", "status": "DRAFT"},
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
			checkText: "Summer Sale",
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

			cmd := newCouponListCmdWithClient(client)
			cmd.SetArgs([]string{})
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
				items := result["items"].([]any)
				if len(items) != 2 {
					t.Errorf("expected 2 coupons, got: %d", len(items))
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestCouponListCmd_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponListCmdWithClient(client)
	cmd.SetArgs([]string{})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No coupons found") {
		t.Errorf("expected 'No coupons found' message, got: %s", output)
	}
}

func TestCouponListCmd_WithStatusFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"couponId": "coupon-001", "title": "Active Coupon", "status": "RUNNING"},
			},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponListCmdWithClient(client)
	cmd.SetArgs([]string{"--status", "running"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Active Coupon") {
		t.Errorf("expected 'Active Coupon' in output, got: %s", output)
	}
}

func TestCouponListCmd_InvalidStatus(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newCouponListCmdWithClient(client)
	cmd.SetArgs([]string{"--status", "invalid"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("expected 'invalid status' in error, got: %v", err)
	}
}

func TestCouponListCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newCouponListCmdWithClient(client)
	cmd.SetArgs([]string{})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to list coupons") {
		t.Errorf("expected 'failed to list coupons' in error, got: %v", err)
	}
}

func TestCouponGetCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/coupon/") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"couponId":       "coupon-001",
				"title":          "Summer Sale",
				"description":    "Get 20% off",
				"status":         "RUNNING",
				"startTimestamp": 1704067200000,
				"endTimestamp":   1735689600000,
				"reward": map[string]any{
					"type": "discount",
					"priceInfo": map[string]any{
						"type":        "fixed",
						"fixedAmount": 500,
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
		wantJSON  bool
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			wantJSON:  false,
			checkText: "Title:       Summer Sale",
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

			cmd := newCouponGetCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "coupon-001"})
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
				if result["couponId"] != "coupon-001" {
					t.Errorf("expected couponId 'coupon-001', got: %v", result["couponId"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestCouponGetCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Coupon not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newCouponGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "coupon-999"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get coupon") {
		t.Errorf("expected 'failed to get coupon' in error, got: %v", err)
	}
}

func TestCouponCloseCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/close") && r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
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
			checkText: "Closed coupon: coupon-001",
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

			cmd := newCouponCloseCmdWithClient(client)
			cmd.SetArgs([]string{"--id", "coupon-001"})
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
				if result["couponId"] != "coupon-001" {
					t.Errorf("expected couponId 'coupon-001', got: %v", result["couponId"])
				}
				if result["status"] != "closed" {
					t.Errorf("expected status 'closed', got: %v", result["status"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestCouponCloseCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Cannot close coupon"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newCouponCloseCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "coupon-999"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to close coupon") {
		t.Errorf("expected 'failed to close coupon' in error, got: %v", err)
	}
}

func TestCouponListCmd_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"couponId": "coupon-001", "title": "Coupon 1", "status": "RUNNING"},
			},
			"next": "cursor-abc123",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponListCmdWithClient(client)
	cmd.SetArgs([]string{})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "More coupons available") {
		t.Errorf("expected pagination message, got: %s", output)
	}
}

// Additional status filter tests
func TestCouponListCmd_StatusDraft(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"couponId": "coupon-001", "title": "Draft Coupon", "status": "DRAFT"},
			},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponListCmdWithClient(client)
	cmd.SetArgs([]string{"--status", "draft"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Draft Coupon") {
		t.Errorf("expected 'Draft Coupon' in output, got: %s", output)
	}
}

func TestCouponListCmd_StatusClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"couponId": "coupon-001", "title": "Closed Coupon", "status": "CLOSED"},
			},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponListCmdWithClient(client)
	cmd.SetArgs([]string{"--status", "closed"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Closed Coupon") {
		t.Errorf("expected 'Closed Coupon' in output, got: %s", output)
	}
}

func TestCouponListCmd_StatusUppercase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"couponId": "coupon-001", "title": "Test Coupon", "status": "RUNNING"},
			},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	statuses := []string{"RUNNING", "DRAFT", "CLOSED"}

	for _, status := range statuses {
		t.Run("uppercase "+status, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = "text"
			defer func() { flags.Output = oldOutput }()

			cmd := newCouponListCmdWithClient(client)
			cmd.SetArgs([]string{"--status", status})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error for status %s: %v", status, err)
			}
		})
	}
}

func TestCouponListCmd_CouponWithoutStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"couponId": "coupon-001", "title": "No Status Coupon"},
			},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponListCmdWithClient(client)
	cmd.SetArgs([]string{})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No Status Coupon") {
		t.Errorf("expected 'No Status Coupon' in output, got: %s", output)
	}
	// Should not have status brackets when status is empty
	if strings.Contains(output, "[]") {
		t.Errorf("should not show empty brackets for missing status, got: %s", output)
	}
}

// Additional coupon get tests for different branches
func TestCouponGetCmd_WithDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"couponId":    "coupon-001",
			"title":       "Test Coupon",
			"description": "Test Description",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "coupon-001"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Description: Test Description") {
		t.Errorf("expected description in output, got: %s", output)
	}
}

func TestCouponGetCmd_WithoutDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"couponId": "coupon-001",
			"title":    "Test Coupon",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "coupon-001"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "Description:") {
		t.Errorf("should not show description line when empty, got: %s", output)
	}
}

func TestCouponGetCmd_WithRateDiscount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"couponId": "coupon-001",
			"title":    "Percentage Coupon",
			"reward": map[string]any{
				"type": "discount",
				"priceInfo": map[string]any{
					"type": "rate",
					"rate": 20,
				},
			},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "coupon-001"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Rate:        20%") {
		t.Errorf("expected rate percentage in output, got: %s", output)
	}
}

func TestCouponGetCmd_WithRewardNoPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"couponId": "coupon-001",
			"title":    "Free Item Coupon",
			"reward": map[string]any{
				"type": "free_item",
			},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "coupon-001"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Reward Type: free_item") {
		t.Errorf("expected reward type in output, got: %s", output)
	}
}

func TestCouponGetCmd_WithoutTimestamps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"couponId": "coupon-001",
			"title":    "No Timestamp Coupon",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "coupon-001"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "Start:") || strings.Contains(output, "End:") {
		t.Errorf("should not show timestamp lines when missing, got: %s", output)
	}
}

func TestCouponGetCmd_WithoutStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"couponId": "coupon-001",
			"title":    "No Status Coupon",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newCouponGetCmdWithClient(client)
	cmd.SetArgs([]string{"--id", "coupon-001"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "Status:") {
		t.Errorf("should not show status line when missing, got: %s", output)
	}
}

// Tests for coupon create command
func TestCouponCreateCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/bot/coupon" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"couponId": "new-coupon-123"})
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
			checkText: "Created coupon: Test Coupon (ID: new-coupon-123)",
		},
		{
			name:      "json output",
			output:    "json",
			checkText: `"couponId"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			defer func() { flags.Output = oldOutput }()
			flags.Output = tt.output

			cmd := newCouponCreateCmdWithClient(client)
			cmd.SetArgs([]string{
				"--title", "Test Coupon",
				"--start", "1704067200000",
				"--end", "1735689600000",
				"--max-use", "1",
				"--visibility", "PUBLIC",
				"--acquisition", "normal",
			})
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

func TestCouponCreateCmd_WithDiscount(t *testing.T) {
	var receivedReward map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if reward, ok := body["reward"]; ok {
				receivedReward = reward.(map[string]any)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"couponId": "coupon-123"})
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

	cmd := newCouponCreateCmdWithClient(client)
	cmd.SetArgs([]string{
		"--title", "Discount Coupon",
		"--start", "1704067200000",
		"--end", "1735689600000",
		"--max-use", "1",
		"--visibility", "PUBLIC",
		"--acquisition", "normal",
		"--discount", "500",
	})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedReward == nil {
		t.Fatal("expected reward to be sent")
	}
	if receivedReward["type"] != "discount" {
		t.Errorf("expected reward type 'discount', got: %v", receivedReward["type"])
	}
}

func TestCouponCreateCmd_ValidationErrors(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	tests := []struct {
		name        string
		args        []string
		errContains string
	}{
		{
			name:        "missing title",
			args:        []string{"--start", "1704067200000", "--end", "1735689600000", "--max-use", "1", "--visibility", "PUBLIC", "--acquisition", "normal"},
			errContains: "--title is required",
		},
		{
			name:        "missing start",
			args:        []string{"--title", "Test", "--end", "1735689600000", "--max-use", "1", "--visibility", "PUBLIC", "--acquisition", "normal"},
			errContains: "--start is required",
		},
		{
			name:        "missing end",
			args:        []string{"--title", "Test", "--start", "1704067200000", "--max-use", "1", "--visibility", "PUBLIC", "--acquisition", "normal"},
			errContains: "--end is required",
		},
		{
			name:        "missing max-use",
			args:        []string{"--title", "Test", "--start", "1704067200000", "--end", "1735689600000", "--visibility", "PUBLIC", "--acquisition", "normal"},
			errContains: "--max-use is required",
		},
		{
			name:        "missing visibility",
			args:        []string{"--title", "Test", "--start", "1704067200000", "--end", "1735689600000", "--max-use", "1", "--acquisition", "normal"},
			errContains: "--visibility is required",
		},
		{
			name:        "missing acquisition",
			args:        []string{"--title", "Test", "--start", "1704067200000", "--end", "1735689600000", "--max-use", "1", "--visibility", "PUBLIC"},
			errContains: "--acquisition is required",
		},
		{
			name:        "invalid visibility",
			args:        []string{"--title", "Test", "--start", "1704067200000", "--end", "1735689600000", "--max-use", "1", "--visibility", "INVALID", "--acquisition", "normal"},
			errContains: "invalid --visibility",
		},
		{
			name:        "invalid acquisition",
			args:        []string{"--title", "Test", "--start", "1704067200000", "--end", "1735689600000", "--max-use", "1", "--visibility", "PUBLIC", "--acquisition", "invalid"},
			errContains: "invalid --acquisition",
		},
		{
			name:        "start after end",
			args:        []string{"--title", "Test", "--start", "1735689600000", "--end", "1704067200000", "--max-use", "1", "--visibility", "PUBLIC", "--acquisition", "normal"},
			errContains: "--start must be before --end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCouponCreateCmdWithClient(client)
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

func TestCouponCreateCmd_VisibilityVariants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"couponId": "coupon-123"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Test both uppercase and lowercase visibility values
	visibilities := []string{"PUBLIC", "public", "UNLISTED", "unlisted"}

	for _, vis := range visibilities {
		t.Run("visibility "+vis, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = "text"
			defer func() { flags.Output = oldOutput }()

			cmd := newCouponCreateCmdWithClient(client)
			cmd.SetArgs([]string{
				"--title", "Test",
				"--start", "1704067200000",
				"--end", "1735689600000",
				"--max-use", "1",
				"--visibility", vis,
				"--acquisition", "normal",
			})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error for visibility %s: %v", vis, err)
			}
		})
	}
}

func TestCouponCreateCmd_AcquisitionVariants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"couponId": "coupon-123"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	acquisitions := []string{"normal", "lottery"}

	for _, acq := range acquisitions {
		t.Run("acquisition "+acq, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = "text"
			defer func() { flags.Output = oldOutput }()

			cmd := newCouponCreateCmdWithClient(client)
			cmd.SetArgs([]string{
				"--title", "Test",
				"--start", "1704067200000",
				"--end", "1735689600000",
				"--max-use", "1",
				"--visibility", "PUBLIC",
				"--acquisition", acq,
			})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error for acquisition %s: %v", acq, err)
			}
		})
	}
}

func TestCouponCreateCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newCouponCreateCmdWithClient(client)
	cmd.SetArgs([]string{
		"--title", "Test",
		"--start", "1704067200000",
		"--end", "1735689600000",
		"--max-use", "1",
		"--visibility", "PUBLIC",
		"--acquisition", "normal",
	})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to create coupon") {
		t.Errorf("error should mention 'failed to create coupon', got: %v", err)
	}
}
