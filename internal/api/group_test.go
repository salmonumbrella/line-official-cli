package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetGroupSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/group/C1234567890abcdef/summary" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"groupId":"C1234567890abcdef","groupName":"Test Group","pictureUrl":"https://example.com/pic.jpg"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	summary, err := client.GetGroupSummary(context.Background(), "C1234567890abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.GroupID != "C1234567890abcdef" {
		t.Errorf("expected groupId 'C1234567890abcdef', got %s", summary.GroupID)
	}
	if summary.GroupName != "Test Group" {
		t.Errorf("expected groupName 'Test Group', got %s", summary.GroupName)
	}
	if summary.PictureURL != "https://example.com/pic.jpg" {
		t.Errorf("expected pictureUrl 'https://example.com/pic.jpg', got %s", summary.PictureURL)
	}
}

func TestClient_GetGroupMemberCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/group/C1234567890abcdef/members/count" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":42}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	count, err := client.GetGroupMemberCount(context.Background(), "C1234567890abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 42 {
		t.Errorf("expected count 42, got %d", count)
	}
}

func TestClient_GetGroupMemberIDs(t *testing.T) {
	tests := []struct {
		name         string
		start        string
		expectedPath string
	}{
		{
			name:         "without start",
			start:        "",
			expectedPath: "/v2/bot/group/C1234567890abcdef/members/ids",
		},
		{
			name:         "with start",
			start:        "cursor123",
			expectedPath: "/v2/bot/group/C1234567890abcdef/members/ids?start=cursor123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := tt.expectedPath
				gotPath := r.URL.Path
				if r.URL.RawQuery != "" {
					gotPath += "?" + r.URL.RawQuery
				}
				if gotPath != expectedPath {
					t.Errorf("unexpected path: %s, expected: %s", gotPath, expectedPath)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"memberIds":["U123","U456"],"next":"cursor456"}`))
			}))
			defer server.Close()

			client := NewClient("test-token", false, false)
			client.baseURL = server.URL

			resp, err := client.GetGroupMemberIDs(context.Background(), "C1234567890abcdef", tt.start)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(resp.MemberIDs) != 2 {
				t.Errorf("expected 2 member IDs, got %d", len(resp.MemberIDs))
			}
			if resp.MemberIDs[0] != "U123" {
				t.Errorf("expected first member ID 'U123', got %s", resp.MemberIDs[0])
			}
			if resp.Next != "cursor456" {
				t.Errorf("expected next 'cursor456', got %s", resp.Next)
			}
		})
	}
}

func TestClient_GetGroupMemberProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/group/C1234567890abcdef/member/U1234567890abcdef" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"userId":"U1234567890abcdef","displayName":"Test User","pictureUrl":"https://example.com/profile.jpg"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	profile, err := client.GetGroupMemberProfile(context.Background(), "C1234567890abcdef", "U1234567890abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.UserID != "U1234567890abcdef" {
		t.Errorf("expected userId 'U1234567890abcdef', got %s", profile.UserID)
	}
	if profile.DisplayName != "Test User" {
		t.Errorf("expected displayName 'Test User', got %s", profile.DisplayName)
	}
	if profile.PictureURL != "https://example.com/profile.jpg" {
		t.Errorf("expected pictureUrl 'https://example.com/profile.jpg', got %s", profile.PictureURL)
	}
}

func TestClient_GetGroupMemberProfile_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"User not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetGroupMemberProfile(context.Background(), "Cinvalid", "Uinvalid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_LeaveGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/group/C1234567890abcdef/leave" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.LeaveGroup(context.Background(), "C1234567890abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_LeaveGroup_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Bot is not in the group"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.LeaveGroup(context.Background(), "Cinvalid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
