package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetMembershipPlans(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/membership/plans" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"memberships":[{"membershipId":12345,"title":"Premium Plan","description":"Full access","benefits":["Exclusive content","Early access"],"price":500,"currency":"JPY","isPublished":true,"isInSale":true},{"membershipId":67890,"title":"Basic Plan","description":"Limited access","price":100,"currency":"JPY","isPublished":true,"isInSale":false}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	plans, err := client.GetMembershipPlans(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plans) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(plans))
	}

	// Verify first plan
	if plans[0].MembershipID != 12345 {
		t.Errorf("expected membershipId 12345, got %d", plans[0].MembershipID)
	}
	if plans[0].Title != "Premium Plan" {
		t.Errorf("expected title 'Premium Plan', got %s", plans[0].Title)
	}
	if plans[0].Description != "Full access" {
		t.Errorf("expected description 'Full access', got %s", plans[0].Description)
	}
	if len(plans[0].Benefits) != 2 {
		t.Errorf("expected 2 benefits, got %d", len(plans[0].Benefits))
	}
	if plans[0].Price != 500 {
		t.Errorf("expected price 500, got %d", plans[0].Price)
	}
	if plans[0].Currency != "JPY" {
		t.Errorf("expected currency 'JPY', got %s", plans[0].Currency)
	}
	if !plans[0].IsPublished {
		t.Error("expected isPublished to be true")
	}
	if !plans[0].IsInSale {
		t.Error("expected isInSale to be true")
	}

	// Verify second plan
	if plans[1].MembershipID != 67890 {
		t.Errorf("expected membershipId 67890, got %d", plans[1].MembershipID)
	}
	if plans[1].IsInSale {
		t.Error("expected isInSale to be false")
	}
}

func TestClient_GetMembershipPlans_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"memberships":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	plans, err := client.GetMembershipPlans(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plans) != 0 {
		t.Errorf("expected 0 plans, got %d", len(plans))
	}
}

func TestClient_GetMembershipPlans_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Invalid access token"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetMembershipPlans(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetUserMembershipStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/users/U1234567890abcdef/membership/subscription" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"memberships":[{"membershipId":12345,"subscriptionState":"ACTIVE","startTime":1609459200,"endTime":1640995200}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	statuses, err := client.GetUserMembershipStatus(context.Background(), "U1234567890abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].MembershipID != 12345 {
		t.Errorf("expected membershipId 12345, got %d", statuses[0].MembershipID)
	}
	if statuses[0].SubscriptionState != "ACTIVE" {
		t.Errorf("expected subscriptionState 'ACTIVE', got %s", statuses[0].SubscriptionState)
	}
	if statuses[0].StartTime != 1609459200 {
		t.Errorf("expected startTime 1609459200, got %d", statuses[0].StartTime)
	}
	if statuses[0].EndTime != 1640995200 {
		t.Errorf("expected endTime 1640995200, got %d", statuses[0].EndTime)
	}
}

func TestClient_GetUserMembershipStatus_NoMembership(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"memberships":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	statuses, err := client.GetUserMembershipStatus(context.Background(), "U1234567890abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 0 {
		t.Errorf("expected 0 statuses, got %d", len(statuses))
	}
}

func TestClient_GetUserMembershipStatus_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"User not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetUserMembershipStatus(context.Background(), "Uinvalid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetMembershipUsers(t *testing.T) {
	tests := []struct {
		name         string
		start        string
		expectedPath string
	}{
		{
			name:         "without start",
			start:        "",
			expectedPath: "/v2/bot/membership/users",
		},
		{
			name:         "with start",
			start:        "cursor123",
			expectedPath: "/v2/bot/membership/users?start=cursor123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath := r.URL.Path
				if r.URL.RawQuery != "" {
					gotPath += "?" + r.URL.RawQuery
				}
				if gotPath != tt.expectedPath {
					t.Errorf("unexpected path: %s, expected: %s", gotPath, tt.expectedPath)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"memberIds":["U123","U456","U789"],"next":"cursor456"}`))
			}))
			defer server.Close()

			client := NewClient("test-token", false, false)
			client.baseURL = server.URL

			resp, err := client.GetMembershipUsers(context.Background(), tt.start)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(resp.MemberIDs) != 3 {
				t.Errorf("expected 3 member IDs, got %d", len(resp.MemberIDs))
			}
			if resp.MemberIDs[0] != "U123" {
				t.Errorf("expected first member ID 'U123', got %s", resp.MemberIDs[0])
			}
			if resp.MemberIDs[1] != "U456" {
				t.Errorf("expected second member ID 'U456', got %s", resp.MemberIDs[1])
			}
			if resp.MemberIDs[2] != "U789" {
				t.Errorf("expected third member ID 'U789', got %s", resp.MemberIDs[2])
			}
			if resp.Next != "cursor456" {
				t.Errorf("expected next 'cursor456', got %s", resp.Next)
			}
		})
	}
}

func TestClient_GetMembershipUsers_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"memberIds":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.GetMembershipUsers(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.MemberIDs) != 0 {
		t.Errorf("expected 0 member IDs, got %d", len(resp.MemberIDs))
	}
	if resp.Next != "" {
		t.Errorf("expected empty next, got %s", resp.Next)
	}
}

func TestClient_GetMembershipUsers_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Membership feature not enabled"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetMembershipUsers(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetMembershipPlans_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetMembershipPlans(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetUserMembershipStatus_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetUserMembershipStatus(context.Background(), "U123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetMembershipUsers_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetMembershipUsers(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
