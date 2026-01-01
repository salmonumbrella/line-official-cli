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

func TestMessageAggregationUsageCmd_Execute(t *testing.T) {
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"numOfCustomAggregationUnits": 25,
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageAggregationUsageCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/v2/bot/message/aggregation/info" {
		t.Errorf("expected path /v2/bot/message/aggregation/info, got %s", capturedPath)
	}

	output := out.String()
	if !strings.Contains(output, "25") {
		t.Errorf("expected output to contain '25', got %s", output)
	}
}

func TestMessageAggregationUsageCmd_Execute_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"numOfCustomAggregationUnits": 42,
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newMessageAggregationUsageCmdWithClient(client)

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
	if result["numOfCustomAggregationUnits"].(float64) != 42 {
		t.Errorf("expected numOfCustomAggregationUnits=42, got %v", result["numOfCustomAggregationUnits"])
	}
}

func TestMessageAggregationListCmd_Execute(t *testing.T) {
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"customAggregationUnits": []string{"unit1", "unit2", "unit3"},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageAggregationListCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/v2/bot/message/aggregation/list" {
		t.Errorf("expected path /v2/bot/message/aggregation/list, got %s", capturedPath)
	}

	output := out.String()
	if !strings.Contains(output, "unit1") {
		t.Errorf("expected output to contain 'unit1', got %s", output)
	}
	if !strings.Contains(output, "unit2") {
		t.Errorf("expected output to contain 'unit2', got %s", output)
	}
	if !strings.Contains(output, "(3)") {
		t.Errorf("expected output to contain '(3)', got %s", output)
	}
}

func TestMessageAggregationListCmd_Execute_WithPagination(t *testing.T) {
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"customAggregationUnits": []string{"unit4", "unit5"},
			"next":                   "next-cursor-token",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageAggregationListCmdWithClient(client)
	cmd.SetArgs([]string{"--limit", "10", "--start", "cursor-abc"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedPath, "limit=10") {
		t.Errorf("expected path to contain 'limit=10', got %s", capturedPath)
	}
	if !strings.Contains(capturedPath, "start=cursor-abc") {
		t.Errorf("expected path to contain 'start=cursor-abc', got %s", capturedPath)
	}

	output := out.String()
	if !strings.Contains(output, "next-cursor-token") {
		t.Errorf("expected output to contain next cursor, got %s", output)
	}
}

func TestMessageAggregationListCmd_Execute_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"customAggregationUnits": []string{},
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageAggregationListCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No aggregation units found") {
		t.Errorf("expected output to contain 'No aggregation units found', got %s", output)
	}
}

func TestMessageAggregationListCmd_Execute_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"customAggregationUnits": []string{"unitA", "unitB"},
			"next":                   "next-page",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newMessageAggregationListCmdWithClient(client)

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
	units := result["customAggregationUnits"].([]any)
	if len(units) != 2 {
		t.Errorf("expected 2 units, got %d", len(units))
	}
	if result["next"] != "next-page" {
		t.Errorf("expected next=next-page, got %v", result["next"])
	}
}

func TestMessageAggregationUsageCmd_Execute_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Internal server error",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageAggregationUsageCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to get aggregation usage") {
		t.Errorf("expected error to contain 'failed to get aggregation usage', got %v", err)
	}
}

func TestMessageAggregationListCmd_Execute_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Unauthorized",
		})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageAggregationListCmdWithClient(client)

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to get aggregation unit list") {
		t.Errorf("expected error to contain 'failed to get aggregation unit list', got %v", err)
	}
}
