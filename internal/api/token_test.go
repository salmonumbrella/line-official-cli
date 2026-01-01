package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClient_IssueChannelToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/oauth/accessToken" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}

		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))

		if values.Get("grant_type") != "client_credentials" {
			t.Errorf("expected grant_type=client_credentials, got %s", values.Get("grant_type"))
		}
		if values.Get("client_id") != "123456" {
			t.Errorf("expected client_id=123456, got %s", values.Get("client_id"))
		}
		if values.Get("client_secret") != "secret123" {
			t.Errorf("expected client_secret=secret123, got %s", values.Get("client_secret"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"token-abc","expires_in":2592000,"token_type":"Bearer"}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	resp, err := client.IssueChannelToken(context.Background(), "123456", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "token-abc" {
		t.Errorf("expected token-abc, got %s", resp.AccessToken)
	}
	if resp.ExpiresIn != 2592000 {
		t.Errorf("expected 2592000, got %d", resp.ExpiresIn)
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("expected Bearer, got %s", resp.TokenType)
	}
}

func TestClient_VerifyChannelToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/oauth/verify" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))

		if values.Get("access_token") != "token-abc" {
			t.Errorf("expected access_token=token-abc, got %s", values.Get("access_token"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"client_id":"123456","expires_in":2591999,"scope":"profile"}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	info, err := client.VerifyChannelToken(context.Background(), "token-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ClientID != "123456" {
		t.Errorf("expected 123456, got %s", info.ClientID)
	}
	if info.ExpiresIn != 2591999 {
		t.Errorf("expected 2591999, got %d", info.ExpiresIn)
	}
	if info.Scope != "profile" {
		t.Errorf("expected profile, got %s", info.Scope)
	}
}

func TestClient_RevokeChannelToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/oauth/revoke" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))

		if values.Get("access_token") != "token-abc" {
			t.Errorf("expected access_token=token-abc, got %s", values.Get("access_token"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	err := client.RevokeChannelToken(context.Background(), "token-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_IssueChannelTokenByJWT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/oauth2/v2.1/token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))

		if values.Get("grant_type") != "client_credentials" {
			t.Errorf("expected grant_type=client_credentials, got %s", values.Get("grant_type"))
		}
		if values.Get("client_assertion_type") != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
			t.Errorf("unexpected client_assertion_type: %s", values.Get("client_assertion_type"))
		}
		if values.Get("client_assertion") != "jwt-token-xyz" {
			t.Errorf("expected client_assertion=jwt-token-xyz, got %s", values.Get("client_assertion"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"v2.1-token-abc","expires_in":2592000,"token_type":"Bearer","key_id":"kid-123"}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	resp, err := client.IssueChannelTokenByJWT(context.Background(), "jwt-token-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "v2.1-token-abc" {
		t.Errorf("expected v2.1-token-abc, got %s", resp.AccessToken)
	}
	if resp.KeyID != "kid-123" {
		t.Errorf("expected kid-123, got %s", resp.KeyID)
	}
}

func TestClient_VerifyChannelTokenByJWT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/oauth2/v2.1/verify" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("access_token") != "v2.1-token-abc" {
			t.Errorf("expected access_token=v2.1-token-abc, got %s", r.URL.Query().Get("access_token"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"client_id":"123456","expires_in":2591999}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	info, err := client.VerifyChannelTokenByJWT(context.Background(), "v2.1-token-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ClientID != "123456" {
		t.Errorf("expected 123456, got %s", info.ClientID)
	}
	if info.ExpiresIn != 2591999 {
		t.Errorf("expected 2591999, got %d", info.ExpiresIn)
	}
}

func TestClient_RevokeChannelTokenByJWT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/oauth2/v2.1/revoke" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))

		if values.Get("access_token") != "v2.1-token-abc" {
			t.Errorf("expected access_token=v2.1-token-abc, got %s", values.Get("access_token"))
		}
		if values.Get("client_id") != "123456" {
			t.Errorf("expected client_id=123456, got %s", values.Get("client_id"))
		}
		if values.Get("client_secret") != "secret123" {
			t.Errorf("expected client_secret=secret123, got %s", values.Get("client_secret"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	err := client.RevokeChannelTokenByJWT(context.Background(), "v2.1-token-abc", "123456", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetAllValidTokenKeyIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/oauth2/v2.1/tokens/kid" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("client_assertion_type") != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
			t.Errorf("unexpected client_assertion_type: %s", r.URL.Query().Get("client_assertion_type"))
		}
		if r.URL.Query().Get("client_assertion") != "jwt-token-xyz" {
			t.Errorf("expected client_assertion=jwt-token-xyz, got %s", r.URL.Query().Get("client_assertion"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"kids":["kid-1","kid-2","kid-3"]}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	kids, err := client.GetAllValidTokenKeyIDs(context.Background(), "jwt-token-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kids) != 3 {
		t.Errorf("expected 3 kids, got %d", len(kids))
	}
	if kids[0] != "kid-1" {
		t.Errorf("expected kid-1, got %s", kids[0])
	}
	if kids[1] != "kid-2" {
		t.Errorf("expected kid-2, got %s", kids[1])
	}
	if kids[2] != "kid-3" {
		t.Errorf("expected kid-3, got %s", kids[2])
	}
}

func TestClient_GetAllValidTokenKeyIDs_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"kids":[]}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	kids, err := client.GetAllValidTokenKeyIDs(context.Background(), "jwt-token-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kids) != 0 {
		t.Errorf("expected 0 kids, got %d", len(kids))
	}
}

func TestClient_IssueChannelToken_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_request","error_description":"Invalid client_id"}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	_, err := client.IssueChannelToken(context.Background(), "invalid", "secret")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_VerifyChannelToken_Expired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_request","error_description":"access token expired"}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	_, err := client.VerifyChannelToken(context.Background(), "expired-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_IssueStatelessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/oauth2/v3/token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}

		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))

		if values.Get("grant_type") != "client_credentials" {
			t.Errorf("expected grant_type=client_credentials, got %s", values.Get("grant_type"))
		}
		if values.Get("client_id") != "123456" {
			t.Errorf("expected client_id=123456, got %s", values.Get("client_id"))
		}
		if values.Get("client_secret") != "secret123" {
			t.Errorf("expected client_secret=secret123, got %s", values.Get("client_secret"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"stateless-token-xyz","expires_in":900,"token_type":"Bearer"}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	resp, err := client.IssueStatelessToken(context.Background(), "123456", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "stateless-token-xyz" {
		t.Errorf("expected stateless-token-xyz, got %s", resp.AccessToken)
	}
	if resp.ExpiresIn != 900 {
		t.Errorf("expected 900 seconds (15 minutes), got %d", resp.ExpiresIn)
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("expected Bearer, got %s", resp.TokenType)
	}
}

func TestClient_IssueStatelessToken_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_request","error_description":"Invalid client_id"}`))
	}))
	defer server.Close()

	client := NewClient("", false, false)
	client.baseURL = server.URL

	_, err := client.IssueStatelessToken(context.Background(), "invalid", "secret")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
