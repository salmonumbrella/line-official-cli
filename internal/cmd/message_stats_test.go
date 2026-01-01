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

func TestMessageQuotaCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bot/message/quota":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":  "limited",
				"value": 1000,
			})
		case "/v2/bot/message/quota/consumption":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"totalUsage": 250,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageQuotaCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "1000/month") {
		t.Errorf("expected output to contain '1000/month', got %s", output)
	}
	if !strings.Contains(output, "250") {
		t.Errorf("expected output to contain '250', got %s", output)
	}
	if !strings.Contains(output, "25.0%") {
		t.Errorf("expected output to contain '25.0%%', got %s", output)
	}
}

func TestMessageQuotaCmd_Execute_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bot/message/quota":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":  "limited",
				"value": 500,
			})
		case "/v2/bot/message/quota/consumption":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"totalUsage": 100,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newMessageQuotaCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output: %v", err)
	}
	if result["type"] != "limited" {
		t.Errorf("expected type=limited, got %v", result["type"])
	}
	if result["limit"].(float64) != 500 {
		t.Errorf("expected limit=500, got %v", result["limit"])
	}
	if result["used"].(float64) != 100 {
		t.Errorf("expected used=100, got %v", result["used"])
	}
}

func TestMessageQuotaCmd_Execute_Unlimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bot/message/quota":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type": "unlimited",
			})
		case "/v2/bot/message/quota/consumption":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"totalUsage": 5000,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageQuotaCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Unlimited") {
		t.Errorf("expected output to contain 'Unlimited', got %s", output)
	}
}

func TestMessageDeliveryStatsCmd_Execute(t *testing.T) {
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ready",
			"success": 150,
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageDeliveryStatsCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "broadcast", "--date", "20251230"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedPath, "/v2/bot/message/delivery/broadcast") {
		t.Errorf("expected path to contain '/v2/bot/message/delivery/broadcast', got %s", capturedPath)
	}
	if !strings.Contains(capturedPath, "date=20251230") {
		t.Errorf("expected path to contain 'date=20251230', got %s", capturedPath)
	}

	output := out.String()
	if !strings.Contains(output, "broadcast") {
		t.Errorf("expected output to contain 'broadcast', got %s", output)
	}
	if !strings.Contains(output, "150") {
		t.Errorf("expected output to contain '150', got %s", output)
	}
}

func TestMessageDeliveryStatsCmd_Execute_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":       "ready",
			"success":      200,
			"failure":      5,
			"requestCount": 10,
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newMessageDeliveryStatsCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "push", "--date", "20251229"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output: %v", err)
	}
	if result["type"] != "push" {
		t.Errorf("expected type=push, got %v", result["type"])
	}
	if result["date"] != "20251229" {
		t.Errorf("expected date=20251229, got %v", result["date"])
	}
	if result["success"].(float64) != 200 {
		t.Errorf("expected success=200, got %v", result["success"])
	}
}

func TestMessageDeliveryStatsCmd_InvalidType(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageDeliveryStatsCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "invalid", "--date", "20251230"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "must be one of") {
		t.Errorf("expected error to contain 'must be one of', got %v", err)
	}
}

func TestMessageDeliveryStatsCmd_AllTypes(t *testing.T) {
	messageTypes := []string{"reply", "push", "multicast", "broadcast", "pnp"}

	for _, msgType := range messageTypes {
		t.Run(msgType, func(t *testing.T) {
			var capturedPath string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"status":  "ready",
					"success": 100,
				})
			}))
			defer server.Close()

			client := api.NewClient("test-token", false, false)
			client.SetBaseURL(server.URL)

			cmd := newMessageDeliveryStatsCmdWithClient(client)
			cmd.SetArgs([]string{"--type", msgType, "--date", "20251230"})

			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error for type %s: %v", msgType, err)
			}

			expectedPath := "/v2/bot/message/delivery/" + msgType
			if capturedPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, capturedPath)
			}
		})
	}
}

func TestMessageQuotaCmd_Execute_LimitedZeroQuota(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bot/message/quota":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":  "limited",
				"value": 0,
			})
		case "/v2/bot/message/quota/consumption":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"totalUsage": 0,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageQuotaCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "0/month") {
		t.Errorf("expected output to contain '0/month', got %s", output)
	}
}

func TestMessageQuotaCmd_Execute_QuotaAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Server error",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageQuotaCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to get quota") {
		t.Errorf("expected error to contain 'failed to get quota', got %v", err)
	}
}

func TestMessageQuotaCmd_Execute_ConsumptionAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bot/message/quota":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":  "limited",
				"value": 1000,
			})
		case "/v2/bot/message/quota/consumption":
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "Server error",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageQuotaCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to get consumption") {
		t.Errorf("expected error to contain 'failed to get consumption', got %v", err)
	}
}

func TestMessageDeliveryStatsCmd_MissingDate(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageDeliveryStatsCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "broadcast"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --date")
	}
}

func TestMessageDeliveryStatsCmd_MissingType(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageDeliveryStatsCmdWithClient(client)
	cmd.SetArgs([]string{"--date", "20251230"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --type")
	}
}

func TestMessageDeliveryStatsCmd_WithFailureAndRequestCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":       "ready",
			"success":      95,
			"failure":      5,
			"requestCount": 100,
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageDeliveryStatsCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "multicast", "--date", "20251230"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Failure:") {
		t.Errorf("expected output to contain 'Failure:', got %s", output)
	}
	if !strings.Contains(output, "Request Count:") {
		t.Errorf("expected output to contain 'Request Count:', got %s", output)
	}
}

func TestMessageDeliveryStatsCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Server error",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageDeliveryStatsCmdWithClient(client)
	cmd.SetArgs([]string{"--type", "reply", "--date", "20251230"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to get delivery stats") {
		t.Errorf("expected error to contain 'failed to get delivery stats', got %v", err)
	}
}
