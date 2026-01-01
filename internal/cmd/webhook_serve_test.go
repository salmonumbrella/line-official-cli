package cmd

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebhookServeCmd_Flags(t *testing.T) {
	cmd := newWebhookServeCmd()

	// Check default values
	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("expected --port flag")
	}
	if portFlag.DefValue != "8080" {
		t.Errorf("expected default port 8080, got %s", portFlag.DefValue)
	}

	secretFlag := cmd.Flags().Lookup("secret")
	if secretFlag == nil {
		t.Fatal("expected --secret flag")
	}

	forwardFlag := cmd.Flags().Lookup("forward")
	if forwardFlag == nil {
		t.Fatal("expected --forward flag")
	}

	quietFlag := cmd.Flags().Lookup("quiet")
	if quietFlag == nil {
		t.Fatal("expected --quiet flag")
	}
}

func TestWebhookHandler_HandleRoot(t *testing.T) {
	handler := &webhookHandler{
		out:    io.Discard,
		errOut: io.Discard,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.handleRoot(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "LINE Webhook Server") {
		t.Error("expected response to contain 'LINE Webhook Server'")
	}
}

func TestWebhookHandler_HandleWebhook_MethodNotAllowed(t *testing.T) {
	handler := &webhookHandler{
		out:    io.Discard,
		errOut: io.Discard,
	}

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", resp.StatusCode)
	}
}

func TestWebhookHandler_HandleWebhook_ValidEvent(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
	}

	payload := LineWebhookPayload{
		Destination: "U1234567890",
		Events: []LineWebhookEvent{
			{
				Type:       "message",
				ReplyToken: "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
				Source: &EventSource{
					Type:   "user",
					UserID: "Udeadbeef",
				},
				Message: json.RawMessage(`{"type":"text","text":"Hello"}`),
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	output := buf.String()
	if !strings.Contains(output, "Event Type: message") {
		t.Error("expected output to contain event type")
	}
	if !strings.Contains(output, "User: Udeadbeef") {
		t.Error("expected output to contain user ID")
	}
	if !strings.Contains(output, `"text":"Hello"`) {
		t.Error("expected output to contain message text")
	}
}

func TestWebhookHandler_HandleWebhook_MultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
	}

	payload := LineWebhookPayload{
		Destination: "U1234567890",
		Events: []LineWebhookEvent{
			{Type: "message"},
			{Type: "follow"},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	output := buf.String()
	if !strings.Contains(output, "--- Event 1 ---") {
		t.Error("expected output to contain event separator for event 1")
	}
	if !strings.Contains(output, "--- Event 2 ---") {
		t.Error("expected output to contain event separator for event 2")
	}
}

func TestWebhookHandler_HandleWebhook_QuietMode(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
		quiet:  true,
	}

	payload := LineWebhookPayload{
		Destination: "U1234567890",
		Events: []LineWebhookEvent{
			{Type: "message"},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output in quiet mode, got: %s", output)
	}
}

func TestWebhookHandler_ValidateSignature_Valid(t *testing.T) {
	secret := "test-channel-secret"
	handler := &webhookHandler{
		secret: secret,
		out:    io.Discard,
		errOut: io.Discard,
	}

	body := []byte(`{"events":[]}`)

	// Generate valid signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !handler.validateSignature(body, signature) {
		t.Error("expected signature to be valid")
	}
}

func TestWebhookHandler_ValidateSignature_Invalid(t *testing.T) {
	handler := &webhookHandler{
		secret: "test-channel-secret",
		out:    io.Discard,
		errOut: io.Discard,
	}

	body := []byte(`{"events":[]}`)
	invalidSignature := base64.StdEncoding.EncodeToString([]byte("invalid"))

	if handler.validateSignature(body, invalidSignature) {
		t.Error("expected signature to be invalid")
	}
}

func TestWebhookHandler_HandleWebhook_SignatureValidation(t *testing.T) {
	secret := "test-channel-secret"
	var buf bytes.Buffer
	handler := &webhookHandler{
		secret: secret,
		out:    &buf,
		errOut: io.Discard,
	}

	body := []byte(`{"events":[{"type":"message"}]}`)

	// Generate valid signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Line-Signature", signature)
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestWebhookHandler_HandleWebhook_MissingSignature(t *testing.T) {
	handler := &webhookHandler{
		secret: "test-channel-secret",
		out:    io.Discard,
		errOut: io.Discard,
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	// No X-Line-Signature header
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestWebhookHandler_HandleWebhook_InvalidSignature(t *testing.T) {
	handler := &webhookHandler{
		secret: "test-channel-secret",
		out:    io.Discard,
		errOut: io.Discard,
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Line-Signature", "invalid-signature")
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestWebhookHandler_Forward(t *testing.T) {
	// Create a test server to receive forwarded requests
	var receivedBody []byte
	var receivedSignature string
	forwardServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		receivedSignature = r.Header.Get("X-Line-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer forwardServer.Close()

	var buf bytes.Buffer
	handler := &webhookHandler{
		forward: forwardServer.URL,
		out:     &buf,
		errOut:  io.Discard,
	}

	body := []byte(`{"events":[{"type":"message"}]}`)
	signature := "test-signature"

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Line-Signature", signature)
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify the forward server received the request
	if string(receivedBody) != string(body) {
		t.Errorf("expected forwarded body %s, got %s", string(body), string(receivedBody))
	}
	if receivedSignature != signature {
		t.Errorf("expected forwarded signature %s, got %s", signature, receivedSignature)
	}

	// Verify forward log message
	if !strings.Contains(buf.String(), "Forwarded to") {
		t.Error("expected output to contain forward status")
	}
}

func TestWebhookHandler_HandleWebhook_GroupSource(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
	}

	payload := LineWebhookPayload{
		Events: []LineWebhookEvent{
			{
				Type: "message",
				Source: &EventSource{
					Type:    "group",
					GroupID: "C1234567890",
					UserID:  "U0987654321",
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	output := buf.String()
	if !strings.Contains(output, "Group: C1234567890") {
		t.Errorf("expected output to contain group ID, got: %s", output)
	}
	if !strings.Contains(output, "User: U0987654321") {
		t.Errorf("expected output to contain user ID, got: %s", output)
	}
}

func TestWebhookHandler_HandleWebhook_RoomSource(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
	}

	payload := LineWebhookPayload{
		Events: []LineWebhookEvent{
			{
				Type: "message",
				Source: &EventSource{
					Type:   "room",
					RoomID: "R1234567890",
					UserID: "U0987654321",
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	output := buf.String()
	if !strings.Contains(output, "Room: R1234567890") {
		t.Errorf("expected output to contain room ID, got: %s", output)
	}
}

func TestWebhookHandler_HandleWebhook_AllEventTypes(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
	}

	payload := LineWebhookPayload{
		Events: []LineWebhookEvent{
			{
				Type:              "message",
				Message:           json.RawMessage(`{"type":"text","text":"hi"}`),
				Postback:          json.RawMessage(`{"data":"action=buy"}`),
				Beacon:            json.RawMessage(`{"hwid":"d41d8cd98f","type":"enter"}`),
				Link:              json.RawMessage(`{"result":"ok"}`),
				Things:            json.RawMessage(`{"deviceId":"t123"}`),
				Members:           json.RawMessage(`[{"userId":"U123"}]`),
				Unsend:            json.RawMessage(`{"messageId":"m123"}`),
				VideoPlayComplete: json.RawMessage(`{"trackingId":"track1"}`),
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	output := buf.String()
	expectedFields := []string{
		"Message:",
		"Postback:",
		"Beacon:",
		"Link:",
		"Things:",
		"Members:",
		"Unsend:",
		"VideoPlayComplete:",
	}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("expected output to contain %s, got: %s", field, output)
		}
	}
}

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    json.RawMessage
		expected string
	}{
		{
			name:     "compact JSON",
			input:    json.RawMessage(`{"type":"text","text":"Hello"}`),
			expected: `{"type":"text","text":"Hello"}`,
		},
		{
			name:     "JSON with whitespace",
			input:    json.RawMessage(`{  "type" : "text" ,  "text" : "Hello"  }`),
			expected: `{"type":"text","text":"Hello"}`,
		},
		{
			name:     "invalid JSON returns original",
			input:    json.RawMessage(`{not valid json`),
			expected: `{not valid json`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatJSON(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestWebhookHandler_LogPayload_NoEvents(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
	}

	payload := &LineWebhookPayload{
		Events: []LineWebhookEvent{},
	}

	handler.logPayload(payload)

	output := buf.String()
	if !strings.Contains(output, "(none)") {
		t.Errorf("expected output to contain '(none)' for empty events, got: %s", output)
	}
}

func TestWebhookHandler_LogPayload_NoDestination(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
	}

	payload := &LineWebhookPayload{
		Destination: "",
		Events: []LineWebhookEvent{
			{Type: "message"},
		},
	}

	handler.logPayload(payload)

	output := buf.String()
	// Should not contain destination line when empty
	if strings.Contains(output, "Destination:") {
		t.Errorf("expected no destination line when empty, got: %s", output)
	}
}

func TestWebhookHandler_ForwardRequest_InvalidURL(t *testing.T) {
	handler := &webhookHandler{
		forward: "://invalid-url",
		out:     io.Discard,
		errOut:  io.Discard,
	}

	err := handler.forwardRequest([]byte(`{}`), http.Header{})
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestWebhookHandler_ForwardRequest_ServerError(t *testing.T) {
	// Create a server that always returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	var buf bytes.Buffer
	handler := &webhookHandler{
		forward: server.URL,
		out:     &buf,
		errOut:  io.Discard,
	}

	headers := http.Header{}
	headers.Set("X-Line-Signature", "test-sig")

	err := handler.forwardRequest([]byte(`{"events":[]}`), headers)
	// No error expected even on 500 - we just log the status
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "500") {
		t.Errorf("expected output to contain status 500, got: %s", output)
	}
}

func TestWebhookHandler_ForwardRequest_QuietMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var buf bytes.Buffer
	handler := &webhookHandler{
		forward: server.URL,
		out:     &buf,
		errOut:  io.Discard,
		quiet:   true,
	}

	err := handler.forwardRequest([]byte(`{}`), http.Header{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// In quiet mode, should not log the forward status
	if buf.Len() != 0 {
		t.Errorf("expected no output in quiet mode, got: %s", buf.String())
	}
}

func TestWebhookHandler_HandleWebhook_InvalidJSON(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
	}

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("not valid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 even for invalid JSON, got %d", resp.StatusCode)
	}

	output := buf.String()
	if !strings.Contains(output, "Raw body:") {
		t.Errorf("expected 'Raw body:' in output for invalid JSON, got: %s", output)
	}
}

func TestWebhookHandler_HandleWebhook_GroupWithoutUserID(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
	}

	payload := LineWebhookPayload{
		Events: []LineWebhookEvent{
			{
				Type: "join",
				Source: &EventSource{
					Type:    "group",
					GroupID: "C1234567890",
					// UserID not set
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	output := buf.String()
	if !strings.Contains(output, "Group: C1234567890") {
		t.Errorf("expected group ID in output, got: %s", output)
	}
	// Should not contain user ID since it's not set
	if strings.Contains(output, "User:") {
		t.Errorf("should not contain User: when not set, got: %s", output)
	}
}

func TestWebhookHandler_HandleWebhook_RoomWithoutUserID(t *testing.T) {
	var buf bytes.Buffer
	handler := &webhookHandler{
		out:    &buf,
		errOut: io.Discard,
	}

	payload := LineWebhookPayload{
		Events: []LineWebhookEvent{
			{
				Type: "join",
				Source: &EventSource{
					Type:   "room",
					RoomID: "R1234567890",
					// UserID not set
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleWebhook(w, req)

	output := buf.String()
	if !strings.Contains(output, "Room: R1234567890") {
		t.Errorf("expected room ID in output, got: %s", output)
	}
}
