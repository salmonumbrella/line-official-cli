package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ValidateMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/validate/push" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req ValidateMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(req.Messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(req.Messages))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	messages := []json.RawMessage{
		json.RawMessage(`{"type":"text","text":"Hello"}`),
	}
	err := client.ValidateMessage(context.Background(), "push", messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ValidateReplyMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/validate/reply" {
			t.Errorf("unexpected path: %s, expected /v2/bot/message/validate/reply", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req ValidateMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(req.Messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(req.Messages))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	messages := []json.RawMessage{
		json.RawMessage(`{"type":"text","text":"Hello"}`),
	}
	err := client.ValidateReplyMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ValidatePushMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/validate/push" {
			t.Errorf("unexpected path: %s, expected /v2/bot/message/validate/push", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	messages := []json.RawMessage{
		json.RawMessage(`{"type":"text","text":"Hello"}`),
	}
	err := client.ValidatePushMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ValidateMulticastMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/validate/multicast" {
			t.Errorf("unexpected path: %s, expected /v2/bot/message/validate/multicast", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	messages := []json.RawMessage{
		json.RawMessage(`{"type":"text","text":"Hello"}`),
	}
	err := client.ValidateMulticastMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ValidateNarrowcastMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/validate/narrowcast" {
			t.Errorf("unexpected path: %s, expected /v2/bot/message/validate/narrowcast", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	messages := []json.RawMessage{
		json.RawMessage(`{"type":"text","text":"Hello"}`),
	}
	err := client.ValidateNarrowcastMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ValidateBroadcastMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/validate/broadcast" {
			t.Errorf("unexpected path: %s, expected /v2/bot/message/validate/broadcast", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	messages := []json.RawMessage{
		json.RawMessage(`{"type":"text","text":"Hello"}`),
	}
	err := client.ValidateBroadcastMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ValidateMessage_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid message format"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	messages := []json.RawMessage{
		json.RawMessage(`{"type":"invalid"}`),
	}
	err := client.ValidateMessage(context.Background(), "push", messages)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_ValidateMessage_MultipleMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ValidateMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(req.Messages) != 3 {
			t.Errorf("expected 3 messages, got %d", len(req.Messages))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	messages := []json.RawMessage{
		json.RawMessage(`{"type":"text","text":"Hello"}`),
		json.RawMessage(`{"type":"text","text":"World"}`),
		json.RawMessage(`{"type":"sticker","packageId":"446","stickerId":"1988"}`),
	}
	err := client.ValidateMessage(context.Background(), "broadcast", messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetMessageQuota(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/quota" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"type":"limited","value":1000}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	quota, err := client.GetMessageQuota(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quota.Type != "limited" {
		t.Errorf("expected type 'limited', got %s", quota.Type)
	}
	if quota.Value != 1000 {
		t.Errorf("expected value 1000, got %d", quota.Value)
	}
}

func TestClient_GetMessageConsumption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/quota/consumption" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"totalUsage":500}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	consumption, err := client.GetMessageConsumption(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if consumption.TotalUsage != 500 {
		t.Errorf("expected total usage 500, got %d", consumption.TotalUsage)
	}
}

func TestClient_GetDeliveryStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/delivery/push" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20251230" {
			t.Errorf("unexpected date: %s", r.URL.Query().Get("date"))
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","success":100}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetDeliveryStats(context.Background(), "push", "20251230")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Status != "ready" {
		t.Errorf("expected status 'ready', got %s", stats.Status)
	}
	if stats.Success != 100 {
		t.Errorf("expected success 100, got %d", stats.Success)
	}
}

func TestClient_NarrowcastTextMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/narrowcast" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req NarrowcastMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(req.Messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(req.Messages))
		}
		if req.Recipient == nil {
			t.Error("expected recipient, got nil")
		} else if req.Recipient.AudienceGroupID != 12345 {
			t.Errorf("expected audience group ID 12345, got %d", req.Recipient.AudienceGroupID)
		}

		// LINE API returns request ID in X-Line-Request-Id header, not in response body
		w.Header().Set("X-Line-Request-Id", "req-123")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.NarrowcastTextMessage(context.Background(), "Hello", 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RequestID != "req-123" {
		t.Errorf("expected request ID 'req-123', got %s", resp.RequestID)
	}
}

func TestClient_GetNarrowcastProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/progress/narrowcast" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("requestId") != "req-123" {
			t.Errorf("unexpected requestId: %s", r.URL.Query().Get("requestId"))
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"phase":"succeeded","successCount":100,"failureCount":5}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	progress, err := client.GetNarrowcastProgress(context.Background(), "req-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if progress["phase"] != "succeeded" {
		t.Errorf("expected phase 'succeeded', got %v", progress["phase"])
	}
}

func TestClient_GetAggregationUnitUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/aggregation/info" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"numOfCustomAggregationUnits":42}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	usage, err := client.GetAggregationUnitUsage(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage.NumOfCustomAggregationUnits != 42 {
		t.Errorf("expected 42, got %d", usage.NumOfCustomAggregationUnits)
	}
}

func TestClient_GetAggregationUnitNameList(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		start         string
		expectedPath  string
		expectedQuery string
	}{
		{
			name:          "no params",
			limit:         0,
			start:         "",
			expectedPath:  "/v2/bot/message/aggregation/list",
			expectedQuery: "",
		},
		{
			name:          "with limit",
			limit:         10,
			start:         "",
			expectedPath:  "/v2/bot/message/aggregation/list",
			expectedQuery: "limit=10",
		},
		{
			name:          "with start",
			limit:         0,
			start:         "abc123",
			expectedPath:  "/v2/bot/message/aggregation/list",
			expectedQuery: "start=abc123",
		},
		{
			name:          "with limit and start",
			limit:         20,
			start:         "xyz789",
			expectedPath:  "/v2/bot/message/aggregation/list",
			expectedQuery: "limit=20&start=xyz789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.expectedPath {
					t.Errorf("unexpected path: %s, expected: %s", r.URL.Path, tt.expectedPath)
				}
				if r.URL.RawQuery != tt.expectedQuery {
					t.Errorf("unexpected query: %s, expected: %s", r.URL.RawQuery, tt.expectedQuery)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"customAggregationUnits":["unit1","unit2"],"next":"nextCursor"}`))
			}))
			defer server.Close()

			client := NewClient("test-token", false, false)
			client.baseURL = server.URL

			resp, err := client.GetAggregationUnitNameList(context.Background(), tt.limit, tt.start)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(resp.CustomAggregationUnits) != 2 {
				t.Errorf("expected 2 units, got %d", len(resp.CustomAggregationUnits))
			}
			if resp.CustomAggregationUnits[0] != "unit1" {
				t.Errorf("expected unit1, got %s", resp.CustomAggregationUnits[0])
			}
			if resp.Next != "nextCursor" {
				t.Errorf("expected nextCursor, got %s", resp.Next)
			}
		})
	}
}

func TestClient_GetAggregationUnitNameList_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"customAggregationUnits":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.GetAggregationUnitNameList(context.Background(), 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.CustomAggregationUnits) != 0 {
		t.Errorf("expected 0 units, got %d", len(resp.CustomAggregationUnits))
	}
	if resp.Next != "" {
		t.Errorf("expected empty next, got %s", resp.Next)
	}
}

// Per-type delivery stats tests

func TestClient_GetReplyMessageStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/delivery/reply" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20251230" {
			t.Errorf("expected date=20251230, got %s", r.URL.Query().Get("date"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","success":100,"failure":5,"requestCount":105}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetReplyMessageStats(context.Background(), "20251230")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Status != "ready" {
		t.Errorf("expected status ready, got %s", stats.Status)
	}
	if stats.Success != 100 {
		t.Errorf("expected success 100, got %d", stats.Success)
	}
	if stats.Failure != 5 {
		t.Errorf("expected failure 5, got %d", stats.Failure)
	}
	if stats.RequestCount != 105 {
		t.Errorf("expected requestCount 105, got %d", stats.RequestCount)
	}
}

func TestClient_GetPushMessageStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/delivery/push" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20251230" {
			t.Errorf("expected date=20251230, got %s", r.URL.Query().Get("date"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","success":200}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetPushMessageStats(context.Background(), "20251230")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Status != "ready" {
		t.Errorf("expected status ready, got %s", stats.Status)
	}
	if stats.Success != 200 {
		t.Errorf("expected success 200, got %d", stats.Success)
	}
}

func TestClient_GetMulticastMessageStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/delivery/multicast" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20251229" {
			t.Errorf("expected date=20251229, got %s", r.URL.Query().Get("date"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","success":50,"failure":2}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetMulticastMessageStats(context.Background(), "20251229")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Status != "ready" {
		t.Errorf("expected status ready, got %s", stats.Status)
	}
	if stats.Success != 50 {
		t.Errorf("expected success 50, got %d", stats.Success)
	}
	if stats.Failure != 2 {
		t.Errorf("expected failure 2, got %d", stats.Failure)
	}
}

func TestClient_GetBroadcastMessageStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/delivery/broadcast" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20251228" {
			t.Errorf("expected date=20251228, got %s", r.URL.Query().Get("date"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","success":1000}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetBroadcastMessageStats(context.Background(), "20251228")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Status != "ready" {
		t.Errorf("expected status ready, got %s", stats.Status)
	}
	if stats.Success != 1000 {
		t.Errorf("expected success 1000, got %d", stats.Success)
	}
}

func TestClient_GetPNPMessageStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/delivery/pnp" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20251227" {
			t.Errorf("expected date=20251227, got %s", r.URL.Query().Get("date"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","success":500,"failure":10,"requestCount":510}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetPNPMessageStats(context.Background(), "20251227")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Status != "ready" {
		t.Errorf("expected status ready, got %s", stats.Status)
	}
	if stats.Success != 500 {
		t.Errorf("expected success 500, got %d", stats.Success)
	}
	if stats.Failure != 10 {
		t.Errorf("expected failure 10, got %d", stats.Failure)
	}
	if stats.RequestCount != 510 {
		t.Errorf("expected requestCount 510, got %d", stats.RequestCount)
	}
}

func TestClient_GetDeliveryStats_Unready(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"unready"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	stats, err := client.GetReplyMessageStats(context.Background(), "20251231")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Status != "unready" {
		t.Errorf("expected status unready, got %s", stats.Status)
	}
	if stats.Success != 0 {
		t.Errorf("expected success 0 for unready status, got %d", stats.Success)
	}
}

func TestClient_GetDeliveryStats_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid date format"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetPushMessageStats(context.Background(), "invalid-date")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_MarkMessagesAsReadByToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/chat/markAsRead" {
			t.Errorf("unexpected path: %s, expected /v2/bot/chat/markAsRead", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req MarkAsReadByTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Chat.Token != "test-chat-token-123" {
			t.Errorf("expected token 'test-chat-token-123', got %s", req.Chat.Token)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.MarkMessagesAsReadByToken(context.Background(), "test-chat-token-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_MarkMessagesAsReadByToken_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid token"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.MarkMessagesAsReadByToken(context.Background(), "invalid-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
