package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/salmonumbrella/line-official-cli/internal/api"
)

func TestMessagePushCmd_Execute_TextMessage(t *testing.T) {
	var capturedBody []byte
	var capturedPath string
	var capturedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U1234567890abcdef", "--text", "Hello!"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/v2/bot/message/push" {
		t.Errorf("expected path /v2/bot/message/push, got %s", capturedPath)
	}
	if capturedMethod != "POST" {
		t.Errorf("expected method POST, got %s", capturedMethod)
	}

	// Check request body
	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	if reqBody["to"] != "U1234567890abcdef" {
		t.Errorf("expected to=U1234567890abcdef, got %v", reqBody["to"])
	}
	messages := reqBody["messages"].([]any)
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	msg := messages[0].(map[string]any)
	if msg["type"] != "text" {
		t.Errorf("expected type=text, got %v", msg["type"])
	}
	if msg["text"] != "Hello!" {
		t.Errorf("expected text=Hello!, got %v", msg["text"])
	}

	// Check output
	output := out.String()
	if !strings.Contains(output, "U1234567890abcdef") {
		t.Errorf("expected output to contain user ID, got %s", output)
	}
}

func TestMessagePushCmd_Execute_ImageMessage(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--image", "https://example.com/image.jpg"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	messages := reqBody["messages"].([]any)
	msg := messages[0].(map[string]any)
	if msg["type"] != "image" {
		t.Errorf("expected type=image, got %v", msg["type"])
	}
	if msg["originalContentUrl"] != "https://example.com/image.jpg" {
		t.Errorf("expected originalContentUrl, got %v", msg["originalContentUrl"])
	}
	// Preview should default to original image
	if msg["previewImageUrl"] != "https://example.com/image.jpg" {
		t.Errorf("expected previewImageUrl to default to originalContentUrl, got %v", msg["previewImageUrl"])
	}
}

func TestMessageBroadcastCmd_Execute_WithYesFlag(t *testing.T) {
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Save and restore global flags
	oldYes := flags.Yes
	flags.Yes = true
	defer func() { flags.Yes = oldYes }()

	cmd := newMessageBroadcastCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "Broadcast message"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/v2/bot/message/broadcast" {
		t.Errorf("expected path /v2/bot/message/broadcast, got %s", capturedPath)
	}

	output := out.String()
	if !strings.Contains(output, "Broadcast sent") {
		t.Errorf("expected output to contain 'Broadcast sent', got %s", output)
	}
}

func TestMessageBroadcastCmd_Execute_CancelledWithoutYes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Ensure --yes is not set
	oldYes := flags.Yes
	flags.Yes = false
	defer func() { flags.Yes = oldYes }()

	cmd := newMessageBroadcastCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "Broadcast message"})

	// Simulate user typing "n" for no
	cmd.SetIn(strings.NewReader("n\n"))

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for cancelled broadcast")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected error to contain 'cancelled', got %v", err)
	}
}

func TestMessageMulticastCmd_Execute(t *testing.T) {
	var capturedBody []byte
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageMulticastCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123,U456,U789", "--text", "Hello all!"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/v2/bot/message/multicast" {
		t.Errorf("expected path /v2/bot/message/multicast, got %s", capturedPath)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}

	userIDs := reqBody["to"].([]any)
	if len(userIDs) != 3 {
		t.Errorf("expected 3 user IDs, got %d", len(userIDs))
	}

	output := out.String()
	if !strings.Contains(output, "3 users") {
		t.Errorf("expected output to contain '3 users', got %s", output)
	}
}

func TestMessagePushCmd_Execute_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid request"}`))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--text", "Hello!"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestMessagePushCmd_Execute_StickerMessage(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--sticker-package", "446", "--sticker-id", "1988"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	messages := reqBody["messages"].([]any)
	msg := messages[0].(map[string]any)
	if msg["type"] != "sticker" {
		t.Errorf("expected type=sticker, got %v", msg["type"])
	}
	if msg["packageId"] != "446" {
		t.Errorf("expected packageId=446, got %v", msg["packageId"])
	}
	if msg["stickerId"] != "1988" {
		t.Errorf("expected stickerId=1988, got %v", msg["stickerId"])
	}
}

func TestMessagePushCmd_Execute_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Save and restore global flags
	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--text", "Hello!"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse output as JSON
	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v, output: %s", err, out.String())
	}
	if result["type"] != "text" {
		t.Errorf("expected type=text in JSON output, got %v", result["type"])
	}
	if result["status"] != "sent" {
		t.Errorf("expected status=sent in JSON output, got %v", result["status"])
	}
	if result["to"] != "U123" {
		t.Errorf("expected to=U123 in JSON output, got %v", result["to"])
	}
}

// Edge case tests for improved coverage

func TestMessagePushCmd_Execute_FlexMessage(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	flexJSON := `{"type":"bubble","body":{"type":"box","layout":"vertical","contents":[]}}`
	cmd.SetArgs([]string{"--to", "U123", "--flex", flexJSON, "--alt-text", "Test flex"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	messages := reqBody["messages"].([]any)
	msg := messages[0].(map[string]any)
	if msg["type"] != "flex" {
		t.Errorf("expected type=flex, got %v", msg["type"])
	}
}

func TestMessagePushCmd_Execute_VideoMessage(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--video", "https://example.com/video.mp4", "--preview", "https://example.com/preview.jpg"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	messages := reqBody["messages"].([]any)
	msg := messages[0].(map[string]any)
	if msg["type"] != "video" {
		t.Errorf("expected type=video, got %v", msg["type"])
	}
}

func TestMessagePushCmd_Execute_VideoWithoutPreview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--video", "https://example.com/video.mp4"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for video without preview")
	}
	if !strings.Contains(err.Error(), "--preview is required") {
		t.Errorf("expected error about preview required, got: %v", err)
	}
}

func TestMessagePushCmd_Execute_AudioMessage(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--audio", "https://example.com/audio.m4a", "--duration", "60000"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	messages := reqBody["messages"].([]any)
	msg := messages[0].(map[string]any)
	if msg["type"] != "audio" {
		t.Errorf("expected type=audio, got %v", msg["type"])
	}
}

func TestMessagePushCmd_Execute_AudioWithoutDuration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--audio", "https://example.com/audio.m4a"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for audio without duration")
	}
	if !strings.Contains(err.Error(), "--duration is required") {
		t.Errorf("expected error about duration required, got: %v", err)
	}
}

func TestMessagePushCmd_Execute_LocationMessage(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--location-title", "Tokyo Tower", "--location-address", "Tokyo, Japan", "--lat", "35.6586", "--lng", "139.7454"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	messages := reqBody["messages"].([]any)
	msg := messages[0].(map[string]any)
	if msg["type"] != "location" {
		t.Errorf("expected type=location, got %v", msg["type"])
	}
}

func TestMessagePushCmd_Execute_LocationMissingTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--location-address", "Tokyo", "--lat", "35.6586", "--lng", "139.7454"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for location without title")
	}
	if !strings.Contains(err.Error(), "--location-title is required") {
		t.Errorf("expected error about title required, got: %v", err)
	}
}

func TestMessagePushCmd_Execute_LocationMissingAddress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--location-title", "Tokyo Tower", "--lat", "35.6586", "--lng", "139.7454"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for location without address")
	}
	if !strings.Contains(err.Error(), "--location-address is required") {
		t.Errorf("expected error about address required, got: %v", err)
	}
}

func TestMessagePushCmd_Execute_LocationMissingCoords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--location-title", "Tokyo Tower", "--location-address", "Tokyo"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for location without coordinates")
	}
	if !strings.Contains(err.Error(), "--lat and --lng are required") {
		t.Errorf("expected error about coordinates required, got: %v", err)
	}
}

func TestMessagePushCmd_Execute_StickerMissingID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessagePushCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123", "--sticker-package", "446"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for sticker without sticker ID")
	}
	if !strings.Contains(err.Error(), "--sticker-package and --sticker-id must be used together") {
		t.Errorf("expected error about sticker IDs, got: %v", err)
	}
}

func TestMessageBroadcastCmd_Execute_ImageMessage_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldYes := flags.Yes
	oldOutput := flags.Output
	flags.Yes = true
	flags.Output = "json"
	defer func() {
		flags.Yes = oldYes
		flags.Output = oldOutput
	}()

	cmd := newMessageBroadcastCmdWithClient(client)
	cmd.SetArgs([]string{"--image", "https://example.com/image.jpg"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}
	if result["type"] != "image" {
		t.Errorf("expected type=image in JSON output, got %v", result["type"])
	}
	if result["status"] != "broadcast" {
		t.Errorf("expected status=broadcast in JSON output, got %v", result["status"])
	}
}

func TestMessageMulticastCmd_Execute_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "json"
	defer func() { flags.Output = oldOutput }()

	cmd := newMessageMulticastCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123,U456", "--text", "Hello!"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}
	if result["type"] != "text" {
		t.Errorf("expected type=text in JSON output, got %v", result["type"])
	}
	if result["status"] != "sent" {
		t.Errorf("expected status=sent in JSON output, got %v", result["status"])
	}
	if result["recipients"] != float64(2) {
		t.Errorf("expected recipients=2 in JSON output, got %v", result["recipients"])
	}
}

func TestMessageMulticastCmd_Execute_TooManyUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageMulticastCmdWithClient(client)

	// Create 501 user IDs
	userIDs := make([]string, 501)
	for i := 0; i < 501; i++ {
		userIDs[i] = "U" + strings.Repeat("0", 32)
	}
	cmd.SetArgs([]string{"--to", strings.Join(userIDs, ","), "--text", "Hello!"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for too many users")
	}
	if !strings.Contains(err.Error(), "max 500") {
		t.Errorf("expected error about max 500, got: %v", err)
	}
}

func TestMessageMulticastCmd_Execute_ImageMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newMessageMulticastCmdWithClient(client)
	cmd.SetArgs([]string{"--to", "U123,U456", "--image", "https://example.com/image.jpg"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "2 users") {
		t.Errorf("expected output to contain '2 users', got: %s", output)
	}
}

func TestMessageBroadcastCmd_Execute_StickerMissingPackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldYes := flags.Yes
	flags.Yes = true
	defer func() { flags.Yes = oldYes }()

	cmd := newMessageBroadcastCmdWithClient(client)
	cmd.SetArgs([]string{"--sticker-id", "1988"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for sticker without package ID")
	}
	if !strings.Contains(err.Error(), "--sticker-package and --sticker-id must be used together") {
		t.Errorf("expected error about sticker IDs, got: %v", err)
	}
}

func TestMessageBroadcastCmd_Execute_ConfirmWithYes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldYes := flags.Yes
	flags.Yes = false
	defer func() { flags.Yes = oldYes }()

	cmd := newMessageBroadcastCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "Broadcast message"})

	// Simulate user typing "yes" (full word)
	cmd.SetIn(strings.NewReader("yes\n"))

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Broadcast sent") {
		t.Errorf("expected 'Broadcast sent' in output, got: %s", out.String())
	}
}

func TestMessageBroadcastCmd_Execute_ConfirmWithY(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldYes := flags.Yes
	flags.Yes = false
	defer func() { flags.Yes = oldYes }()

	cmd := newMessageBroadcastCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "Broadcast message"})

	// Simulate user typing "Y" (uppercase)
	cmd.SetIn(strings.NewReader("Y\n"))

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Broadcast sent") {
		t.Errorf("expected 'Broadcast sent' in output, got: %s", out.String())
	}
}

func TestMessageMulticastCmd_Execute_NoUsers(t *testing.T) {
	client := api.NewClient("test-token", false, false)

	cmd := newMessageMulticastCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "Hello!"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --to flag")
	}
}

func TestMessageBroadcastCmd_Execute_DryRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Server should NOT be called in dry-run mode
		t.Error("unexpected server call in dry-run mode")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, true) // dryRun=true
	client.SetBaseURL(server.URL)

	oldYes := flags.Yes
	flags.Yes = true
	defer func() { flags.Yes = oldYes }()

	cmd := newMessageBroadcastCmdWithClient(client)
	cmd.SetArgs([]string{"--text", "Broadcast message"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
