package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_DetachModule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/channel/detach" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req["botId"] != "U1234567890abcdef" {
			t.Errorf("expected botId 'U1234567890abcdef', got %s", req["botId"])
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.DetachModule(context.Background(), "U1234567890abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DetachModule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Permission denied"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.DetachModule(context.Background(), "U1234567890abcdef")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_AcquireModuleChatControl(t *testing.T) {
	tests := []struct {
		name          string
		chatID        string
		expired       bool
		expectedField string
	}{
		{
			name:          "acquire with TTL expiry",
			chatID:        "U1234567890abcdef",
			expired:       true,
			expectedField: "true",
		},
		{
			name:          "acquire without expiry",
			chatID:        "U1234567890abcdef",
			expired:       false,
			expectedField: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/v2/bot/chat/" + tt.chatID + "/control/acquire"
				if r.URL.Path != expectedPath {
					t.Errorf("unexpected path: %s, expected: %s", r.URL.Path, expectedPath)
				}
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}

				var req map[string]bool
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("failed to decode request: %v", err)
				}
				if req["expired"] != tt.expired {
					t.Errorf("expected expired=%v, got %v", tt.expired, req["expired"])
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer server.Close()

			client := NewClient("test-token", false, false)
			client.baseURL = server.URL

			err := client.AcquireModuleChatControl(context.Background(), tt.chatID, tt.expired)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_AcquireModuleChatControl_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"message":"Chat control already acquired"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.AcquireModuleChatControl(context.Background(), "U1234567890abcdef", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_ReleaseModuleChatControl(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/chat/U1234567890abcdef/control/release" {
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

	err := client.ReleaseModuleChatControl(context.Background(), "U1234567890abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ReleaseModuleChatControl_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"message":"Chat control not held by this channel"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.ReleaseModuleChatControl(context.Background(), "U1234567890abcdef")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_ExchangeModuleToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path
		if r.URL.Path != "/module/auth/v1/token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// Verify content type is form-encoded
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/x-www-form-urlencoded" {
			t.Errorf("expected Content-Type application/x-www-form-urlencoded, got %s", contentType)
		}
		// Verify no Authorization header (this endpoint doesn't use Bearer auth)
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("expected no Authorization header, got %s", auth)
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		if r.FormValue("grant_type") != "authorization_code" {
			t.Errorf("expected grant_type=authorization_code, got %s", r.FormValue("grant_type"))
		}
		if r.FormValue("code") != "test-auth-code" {
			t.Errorf("expected code=test-auth-code, got %s", r.FormValue("code"))
		}
		if r.FormValue("redirect_uri") != "https://example.com/callback" {
			t.Errorf("expected redirect_uri=https://example.com/callback, got %s", r.FormValue("redirect_uri"))
		}
		if r.FormValue("client_id") != "1234567890" {
			t.Errorf("expected client_id=1234567890, got %s", r.FormValue("client_id"))
		}
		if r.FormValue("client_secret") != "secret123" {
			t.Errorf("expected client_secret=secret123, got %s", r.FormValue("client_secret"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"access_token": "module-access-token-xyz",
			"token_type": "Bearer",
			"expires_in": 2592000,
			"refresh_token": "refresh-token-abc",
			"scope": "profile openid"
		}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	resp, err := client.ExchangeModuleToken(
		context.Background(),
		"test-auth-code",
		"https://example.com/callback",
		"1234567890",
		"secret123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.AccessToken != "module-access-token-xyz" {
		t.Errorf("expected access_token 'module-access-token-xyz', got %s", resp.AccessToken)
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("expected token_type 'Bearer', got %s", resp.TokenType)
	}
	if resp.ExpiresIn != 2592000 {
		t.Errorf("expected expires_in 2592000, got %d", resp.ExpiresIn)
	}
	if resp.RefreshToken != "refresh-token-abc" {
		t.Errorf("expected refresh_token 'refresh-token-abc', got %s", resp.RefreshToken)
	}
	if resp.Scope != "profile openid" {
		t.Errorf("expected scope 'profile openid', got %s", resp.Scope)
	}
}

func TestClient_ExchangeModuleToken_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"The authorization code has expired"}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	_, err := client.ExchangeModuleToken(
		context.Background(),
		"expired-code",
		"https://example.com/callback",
		"1234567890",
		"secret123",
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetBotsWithModules(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		start         string
		expectedQuery string
	}{
		{
			name:          "no params",
			limit:         0,
			start:         "",
			expectedQuery: "",
		},
		{
			name:          "with limit",
			limit:         10,
			start:         "",
			expectedQuery: "limit=10",
		},
		{
			name:          "with start",
			limit:         0,
			start:         "abc123",
			expectedQuery: "start=abc123",
		},
		{
			name:          "with both",
			limit:         50,
			start:         "xyz789",
			expectedQuery: "limit=50&start=xyz789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/bot/list" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				// Verify Bearer token auth is used
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-token" {
					t.Errorf("expected Authorization 'Bearer test-token', got %s", auth)
				}
				// Verify query params
				if tt.expectedQuery != "" && r.URL.RawQuery != tt.expectedQuery {
					t.Errorf("expected query '%s', got '%s'", tt.expectedQuery, r.URL.RawQuery)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
					"bots": [
						{
							"userId": "U1234567890abcdef",
							"basicId": "@123abc",
							"premiumId": "@premium",
							"displayName": "Test Bot",
							"pictureUrl": "https://example.com/picture.jpg"
						},
						{
							"userId": "U0987654321fedcba",
							"basicId": "@456def",
							"displayName": "Another Bot"
						}
					],
					"next": "continuation-token-xyz"
				}`))
			}))
			defer server.Close()

			client := NewClient("test-token", false, false)
			client.baseURL = server.URL

			resp, err := client.GetBotsWithModules(context.Background(), tt.limit, tt.start)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(resp.Bots) != 2 {
				t.Errorf("expected 2 bots, got %d", len(resp.Bots))
			}
			if resp.Bots[0].UserID != "U1234567890abcdef" {
				t.Errorf("expected userId 'U1234567890abcdef', got %s", resp.Bots[0].UserID)
			}
			if resp.Bots[0].BasicID != "@123abc" {
				t.Errorf("expected basicId '@123abc', got %s", resp.Bots[0].BasicID)
			}
			if resp.Bots[0].PremiumID != "@premium" {
				t.Errorf("expected premiumId '@premium', got %s", resp.Bots[0].PremiumID)
			}
			if resp.Bots[0].DisplayName != "Test Bot" {
				t.Errorf("expected displayName 'Test Bot', got %s", resp.Bots[0].DisplayName)
			}
			if resp.Bots[0].PictureURL != "https://example.com/picture.jpg" {
				t.Errorf("expected pictureUrl 'https://example.com/picture.jpg', got %s", resp.Bots[0].PictureURL)
			}
			if resp.Bots[1].PremiumID != "" {
				t.Errorf("expected empty premiumId for second bot, got %s", resp.Bots[1].PremiumID)
			}
			if resp.Next != "continuation-token-xyz" {
				t.Errorf("expected next 'continuation-token-xyz', got %s", resp.Next)
			}
		})
	}
}

func TestClient_GetBotsWithModules_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"bots": []}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.GetBotsWithModules(context.Background(), 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Bots) != 0 {
		t.Errorf("expected 0 bots, got %d", len(resp.Bots))
	}
	if resp.Next != "" {
		t.Errorf("expected empty next, got %s", resp.Next)
	}
}

func TestClient_GetBotsWithModules_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Invalid channel access token"}`))
	}))
	defer server.Close()

	client := NewClient("invalid-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetBotsWithModules(context.Background(), 0, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
