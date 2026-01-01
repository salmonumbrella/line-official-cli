package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_LinkRichMenuToUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/bulk/link" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req struct {
			RichMenuID string   `json:"richMenuId"`
			UserIDs    []string `json:"userIds"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to parse request body: %v", err)
		}
		if req.RichMenuID != "richmenu-123" {
			t.Errorf("expected richmenu-123, got %s", req.RichMenuID)
		}
		if len(req.UserIDs) != 2 {
			t.Errorf("expected 2 user IDs, got %d", len(req.UserIDs))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.LinkRichMenuToUsers(context.Background(), "richmenu-123", []string{"U1", "U2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UnlinkRichMenuFromUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/bulk/unlink" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req struct {
			UserIDs []string `json:"userIds"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to parse request body: %v", err)
		}
		if len(req.UserIDs) != 3 {
			t.Errorf("expected 3 user IDs, got %d", len(req.UserIDs))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.UnlinkRichMenuFromUsers(context.Background(), []string{"U1", "U2", "U3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_RichMenuBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/batch" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req struct {
			Operations      []RichMenuBatchOperation `json:"operations"`
			ResumeRequestID string                   `json:"resumeRequestId,omitempty"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to parse request body: %v", err)
		}
		if len(req.Operations) != 2 {
			t.Errorf("expected 2 operations, got %d", len(req.Operations))
		}
		if req.Operations[0].Type != "link" {
			t.Errorf("expected link, got %s", req.Operations[0].Type)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"requestId":"req-abc123"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	operations := []RichMenuBatchOperation{
		{Type: "link", RichMenuID: "richmenu-123", UserIDs: []string{"U1", "U2"}},
		{Type: "unlink", UserIDs: []string{"U3"}},
	}

	requestID, err := client.RichMenuBatch(context.Background(), operations, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestID != "req-abc123" {
		t.Errorf("expected req-abc123, got %s", requestID)
	}
}

func TestClient_RichMenuBatch_WithResume(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Operations      []RichMenuBatchOperation `json:"operations"`
			ResumeRequestID string                   `json:"resumeRequestId,omitempty"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to parse request body: %v", err)
		}
		if req.ResumeRequestID != "previous-req" {
			t.Errorf("expected previous-req, got %s", req.ResumeRequestID)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"requestId":"new-req"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	operations := []RichMenuBatchOperation{
		{Type: "link", RichMenuID: "richmenu-123", UserIDs: []string{"U1"}},
	}

	requestID, err := client.RichMenuBatch(context.Background(), operations, "previous-req")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestID != "new-req" {
		t.Errorf("expected new-req, got %s", requestID)
	}
}

func TestClient_ValidateRichMenuBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/validate/batch" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	operations := []RichMenuBatchOperation{
		{Type: "link", RichMenuID: "richmenu-123", UserIDs: []string{"U1"}},
	}

	err := client.ValidateRichMenuBatch(context.Background(), operations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetRichMenuBatchProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/progress/batch" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("requestId") != "req-123" {
			t.Errorf("expected requestId=req-123, got %s", r.URL.Query().Get("requestId"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"phase":"succeeded","acceptedTime":"2024-01-01T00:00:00Z","completedTime":"2024-01-01T00:01:00Z"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	progress, err := client.GetRichMenuBatchProgress(context.Background(), "req-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if progress.Phase != "succeeded" {
		t.Errorf("expected succeeded, got %s", progress.Phase)
	}
	if progress.AcceptedTime != "2024-01-01T00:00:00Z" {
		t.Errorf("unexpected acceptedTime: %s", progress.AcceptedTime)
	}
	if progress.CompletedTime != "2024-01-01T00:01:00Z" {
		t.Errorf("unexpected completedTime: %s", progress.CompletedTime)
	}
}

func TestClient_ValidateRichMenu(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/validate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var menu CreateRichMenuRequest
		if err := json.Unmarshal(body, &menu); err != nil {
			t.Errorf("failed to parse request body: %v", err)
		}
		if menu.Name != "Test Menu" {
			t.Errorf("expected Test Menu, got %s", menu.Name)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	menu := &CreateRichMenuRequest{
		Size:        RichMenuSize{Width: 2500, Height: 1686},
		Selected:    false,
		Name:        "Test Menu",
		ChatBarText: "Tap here",
		Areas:       []RichMenuArea{},
	}

	err := client.ValidateRichMenu(context.Background(), menu)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DownloadRichMenuImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/richmenu-123/content" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fake-image-data"))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	// Override baseURL - the function will detect this is not production and won't override
	client.baseURL = server.URL

	data, contentType, err := client.DownloadRichMenuImage(context.Background(), "richmenu-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "fake-image-data" {
		t.Errorf("unexpected data: %s", string(data))
	}
	if contentType != "image/png" {
		t.Errorf("expected image/png, got %s", contentType)
	}
}

func TestClient_DownloadRichMenuImage_JPEG(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fake-jpeg-data"))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	data, contentType, err := client.DownloadRichMenuImage(context.Background(), "richmenu-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "fake-jpeg-data" {
		t.Errorf("unexpected data: %s", string(data))
	}
	if contentType != "image/jpeg" {
		t.Errorf("expected image/jpeg, got %s", contentType)
	}
}

func TestClient_GetRichMenuList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/richmenu/list" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"richmenus":[{"richMenuId":"rm-1","name":"Menu 1"},{"richMenuId":"rm-2","name":"Menu 2"}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	menus, err := client.GetRichMenuList(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(menus) != 2 {
		t.Errorf("expected 2 menus, got %d", len(menus))
	}
	if menus[0].RichMenuID != "rm-1" {
		t.Errorf("expected rm-1, got %s", menus[0].RichMenuID)
	}
}

func TestClient_DeleteRichMenu(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/richmenu-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.DeleteRichMenu(context.Background(), "richmenu-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteRichMenu_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Rich menu not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.DeleteRichMenu(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SetDefaultRichMenu(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/user/all/richmenu/richmenu-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.SetDefaultRichMenu(context.Background(), "richmenu-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SetDefaultRichMenu_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Rich menu not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.SetDefaultRichMenu(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CancelDefaultRichMenu(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/user/all/richmenu" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.CancelDefaultRichMenu(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CancelDefaultRichMenu_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Default rich menu not set"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.CancelDefaultRichMenu(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetDefaultRichMenuID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/user/all/richmenu" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"richMenuId":"richmenu-default"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	menuID, err := client.GetDefaultRichMenuID(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if menuID != "richmenu-default" {
		t.Errorf("expected richmenu-default, got %s", menuID)
	}
}

func TestClient_GetDefaultRichMenuID_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Default rich menu not set"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetDefaultRichMenuID(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UploadRichMenuImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/richmenu-123/content" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "image/png" {
			t.Errorf("expected Content-Type image/png, got %s", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "fake-image-data" {
			t.Errorf("unexpected body: %s", string(body))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.UploadRichMenuImage(context.Background(), "richmenu-123", "image/png", []byte("fake-image-data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UploadRichMenuImage_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid image format"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.UploadRichMenuImage(context.Background(), "richmenu-123", "image/gif", []byte("bad-data"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetRichMenu(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/richmenu-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"richMenuId":"richmenu-123","name":"Test Menu","chatBarText":"Tap here","selected":true,"size":{"width":2500,"height":1686},"areas":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	menu, err := client.GetRichMenu(context.Background(), "richmenu-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if menu.RichMenuID != "richmenu-123" {
		t.Errorf("expected richmenu-123, got %s", menu.RichMenuID)
	}
	if menu.Name != "Test Menu" {
		t.Errorf("expected Test Menu, got %s", menu.Name)
	}
	if menu.ChatBarText != "Tap here" {
		t.Errorf("expected Tap here, got %s", menu.ChatBarText)
	}
	if !menu.Selected {
		t.Error("expected selected to be true")
	}
	if menu.Size.Width != 2500 || menu.Size.Height != 1686 {
		t.Errorf("unexpected size: %dx%d", menu.Size.Width, menu.Size.Height)
	}
}

func TestClient_GetRichMenu_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Rich menu not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetRichMenu(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_LinkRichMenuToUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/user/U123456/richmenu/richmenu-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.LinkRichMenuToUser(context.Background(), "U123456", "richmenu-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_LinkRichMenuToUser_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"User not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.LinkRichMenuToUser(context.Background(), "nonexistent", "richmenu-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UnlinkRichMenuFromUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/user/U123456/richmenu" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.UnlinkRichMenuFromUser(context.Background(), "U123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UnlinkRichMenuFromUser_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"User not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.UnlinkRichMenuFromUser(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetUserRichMenu(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/user/U123456/richmenu" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"richMenuId":"richmenu-user"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	menuID, err := client.GetUserRichMenu(context.Background(), "U123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if menuID != "richmenu-user" {
		t.Errorf("expected richmenu-user, got %s", menuID)
	}
}

func TestClient_GetUserRichMenu_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"User not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetUserRichMenu(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateRichMenuAlias(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/alias" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req CreateRichMenuAliasRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to parse request body: %v", err)
		}
		if req.RichMenuAliasID != "alias-main" {
			t.Errorf("expected alias-main, got %s", req.RichMenuAliasID)
		}
		if req.RichMenuID != "richmenu-123" {
			t.Errorf("expected richmenu-123, got %s", req.RichMenuID)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.CreateRichMenuAlias(context.Background(), "alias-main", "richmenu-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateRichMenuAlias_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"message":"Alias already exists"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.CreateRichMenuAlias(context.Background(), "alias-existing", "richmenu-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetRichMenuAlias(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/alias/alias-main" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"richMenuAliasId":"alias-main","richMenuId":"richmenu-123"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	alias, err := client.GetRichMenuAlias(context.Background(), "alias-main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alias.RichMenuAliasID != "alias-main" {
		t.Errorf("expected alias-main, got %s", alias.RichMenuAliasID)
	}
	if alias.RichMenuID != "richmenu-123" {
		t.Errorf("expected richmenu-123, got %s", alias.RichMenuID)
	}
}

func TestClient_GetRichMenuAlias_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Alias not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.GetRichMenuAlias(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateRichMenuAlias(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/alias/alias-main" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req UpdateRichMenuAliasRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to parse request body: %v", err)
		}
		if req.RichMenuID != "richmenu-456" {
			t.Errorf("expected richmenu-456, got %s", req.RichMenuID)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.UpdateRichMenuAlias(context.Background(), "alias-main", "richmenu-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UpdateRichMenuAlias_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Alias not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.UpdateRichMenuAlias(context.Background(), "nonexistent", "richmenu-456")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_DeleteRichMenuAlias(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/alias/alias-main" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.DeleteRichMenuAlias(context.Background(), "alias-main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteRichMenuAlias_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Alias not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.DeleteRichMenuAlias(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_ListRichMenuAliases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/richmenu/alias/list" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"aliases":[{"richMenuAliasId":"alias-1","richMenuId":"rm-1"},{"richMenuAliasId":"alias-2","richMenuId":"rm-2"}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	aliases, err := client.ListRichMenuAliases(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(aliases))
	}
	if aliases[0].RichMenuAliasID != "alias-1" {
		t.Errorf("expected alias-1, got %s", aliases[0].RichMenuAliasID)
	}
	if aliases[1].RichMenuID != "rm-2" {
		t.Errorf("expected rm-2, got %s", aliases[1].RichMenuID)
	}
}

func TestClient_ListRichMenuAliases_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"Internal server error"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	_, err := client.ListRichMenuAliases(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_ListRichMenuAliases_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"aliases":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	aliases, err := client.ListRichMenuAliases(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(aliases) != 0 {
		t.Errorf("expected 0 aliases, got %d", len(aliases))
	}
}
