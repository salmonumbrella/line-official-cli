package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetFollowerStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/insight/followers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20251230" {
			t.Errorf("expected date=20251230, got %s", r.URL.Query().Get("date"))
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","followers":1000,"targetedReaches":800,"blocks":50}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetFollowerStats(context.Background(), "20251230")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Followers == nil || *stats.Followers != 1000 {
		t.Errorf("expected followers 1000, got %v", stats.Followers)
	}
	if stats.TargetedReaches == nil || *stats.TargetedReaches != 800 {
		t.Errorf("expected targetedReaches 800, got %v", stats.TargetedReaches)
	}
	if stats.Blocks == nil || *stats.Blocks != 50 {
		t.Errorf("expected blocks 50, got %v", stats.Blocks)
	}
}

func TestClient_GetMessageDeliveryStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/insight/message/delivery" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20251230" {
			t.Errorf("expected date=20251230, got %s", r.URL.Query().Get("date"))
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","broadcast":100,"targeting":200,"autoResponse":50}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetMessageDeliveryStats(context.Background(), "20251230")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Broadcast == nil || *stats.Broadcast != 100 {
		t.Errorf("expected broadcast 100, got %v", stats.Broadcast)
	}
	if stats.Targeting == nil || *stats.Targeting != 200 {
		t.Errorf("expected targeting 200, got %v", stats.Targeting)
	}
	if stats.AutoResponse == nil || *stats.AutoResponse != 50 {
		t.Errorf("expected autoResponse 50, got %v", stats.AutoResponse)
	}
}

func TestClient_GetFriendsDemographics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/insight/demographic" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"available":true,"genders":[{"gender":"male","percentage":40.5}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	demo, err := client.GetFriendsDemographics(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if demo.Available == nil || !*demo.Available {
		t.Errorf("expected available true, got %v", demo.Available)
	}
	if demo.Genders == nil || len(*demo.Genders) != 1 {
		t.Errorf("expected 1 gender entry, got %v", demo.Genders)
	}
}

func TestClient_GetMessageEventStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/insight/message/event" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("requestId") != "req-123" {
			t.Errorf("expected requestId=req-123, got %s", r.URL.Query().Get("requestId"))
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"overview":{"requestId":"req-123","delivered":1000,"uniqueImpression":800,"uniqueClick":200}}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetMessageEventStats(context.Background(), "req-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Overview == nil {
		t.Fatal("expected overview, got nil")
	}
	if stats.Overview.Delivered != 1000 {
		t.Errorf("expected delivered 1000, got %d", stats.Overview.Delivered)
	}
	if stats.Overview.UniqueImpression != 800 {
		t.Errorf("expected uniqueImpression 800, got %d", stats.Overview.UniqueImpression)
	}
	if stats.Overview.UniqueClick != 200 {
		t.Errorf("expected uniqueClick 200, got %d", stats.Overview.UniqueClick)
	}
}

func TestClient_GetStatisticsPerUnit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/insight/message/event/aggregation" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("customAggregationUnit") != "campaign-2024" {
			t.Errorf("expected customAggregationUnit=campaign-2024, got %s", r.URL.Query().Get("customAggregationUnit"))
		}
		if r.URL.Query().Get("from") != "20251224" {
			t.Errorf("expected from=20251224, got %s", r.URL.Query().Get("from"))
		}
		if r.URL.Query().Get("to") != "20251231" {
			t.Errorf("expected to=20251231, got %s", r.URL.Query().Get("to"))
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"overview":{"uniqueImpression":5000,"uniqueClick":1500,"uniqueMediaPlayed":300,"uniqueMediaPlayed100Percent":100}}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetStatisticsPerUnit(context.Background(), "campaign-2024", "20251224", "20251231")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Overview == nil {
		t.Fatal("expected overview, got nil")
	}
	if stats.Overview.UniqueImpression != 5000 {
		t.Errorf("expected uniqueImpression 5000, got %d", stats.Overview.UniqueImpression)
	}
	if stats.Overview.UniqueClick != 1500 {
		t.Errorf("expected uniqueClick 1500, got %d", stats.Overview.UniqueClick)
	}
	if stats.Overview.UniqueMediaPlayed != 300 {
		t.Errorf("expected uniqueMediaPlayed 300, got %d", stats.Overview.UniqueMediaPlayed)
	}
	if stats.Overview.UniqueMediaPlayedComplete != 100 {
		t.Errorf("expected uniqueMediaPlayedComplete 100, got %d", stats.Overview.UniqueMediaPlayedComplete)
	}
}

func TestClient_GetStatisticsPerUnit_NoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetStatisticsPerUnit(context.Background(), "nonexistent-unit", "20251224", "20251231")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Overview != nil {
		t.Errorf("expected nil overview, got %v", stats.Overview)
	}
}

func TestClient_GetFollowerStats_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid date format"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetFollowerStats(context.Background(), "invalid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
