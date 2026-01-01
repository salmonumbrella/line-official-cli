package cmd

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type serveFlags struct {
	Port    int
	Secret  string
	Forward string
	Quiet   bool
}

// LineWebhookEvent represents a single LINE webhook event
type LineWebhookEvent struct {
	Type              string          `json:"type"`
	Timestamp         int64           `json:"timestamp"`
	Source            *EventSource    `json:"source,omitempty"`
	ReplyToken        string          `json:"replyToken,omitempty"`
	Message           json.RawMessage `json:"message,omitempty"`
	Postback          json.RawMessage `json:"postback,omitempty"`
	Beacon            json.RawMessage `json:"beacon,omitempty"`
	Link              json.RawMessage `json:"link,omitempty"`
	Things            json.RawMessage `json:"things,omitempty"`
	Members           json.RawMessage `json:"members,omitempty"`
	Unsend            json.RawMessage `json:"unsend,omitempty"`
	VideoPlayComplete json.RawMessage `json:"videoPlayComplete,omitempty"`
}

// EventSource represents the source of a webhook event
type EventSource struct {
	Type    string `json:"type"`
	UserID  string `json:"userId,omitempty"`
	GroupID string `json:"groupId,omitempty"`
	RoomID  string `json:"roomId,omitempty"`
}

// LineWebhookPayload represents the full webhook request body
type LineWebhookPayload struct {
	Destination string             `json:"destination"`
	Events      []LineWebhookEvent `json:"events"`
}

func newWebhookServeCmd() *cobra.Command {
	sf := &serveFlags{}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a local webhook server for development",
		Long: `Start a local HTTP server to receive LINE webhook events during development.

The server logs incoming webhook events in a human-readable format,
making it easy to debug and test your LINE bot.

If --secret is provided, the server validates webhook signatures using HMAC-SHA256.
If --forward is provided, events are forwarded to the specified URL after logging.`,
		Example: `  # Basic: just log events
  line webhook serve

  # With signature validation
  line webhook serve --secret YOUR_CHANNEL_SECRET

  # Forward to local app after logging
  line webhook serve --forward http://localhost:3000/webhook

  # Custom port
  line webhook serve --port 9000

  # Quiet mode - only show errors
  line webhook serve --quiet`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookServe(cmd, sf)
		},
	}

	cmd.Flags().IntVarP(&sf.Port, "port", "p", 8080, "Port to listen on")
	cmd.Flags().StringVar(&sf.Secret, "secret", "", "Channel secret for signature validation")
	cmd.Flags().StringVar(&sf.Forward, "forward", "", "URL to forward events to after logging")
	cmd.Flags().BoolVarP(&sf.Quiet, "quiet", "q", false, "Only show errors, no event logging")

	return cmd
}

func runWebhookServe(cmd *cobra.Command, sf *serveFlags) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	// Create webhook handler
	handler := &webhookHandler{
		secret:  sf.Secret,
		forward: sf.Forward,
		quiet:   sf.Quiet,
		out:     out,
		errOut:  errOut,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", handler.handleWebhook)
	mux.HandleFunc("/", handler.handleRoot)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", sf.Port),
		Handler: mux,
	}

	// Channel to receive shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Print startup message
	url := fmt.Sprintf("http://localhost:%d/webhook", sf.Port)
	_, _ = fmt.Fprintf(out, "Webhook server listening on %s\n", url)
	_, _ = fmt.Fprintf(out, "Press Ctrl+C to stop\n")
	if sf.Secret != "" {
		_, _ = fmt.Fprintf(out, "Signature validation: enabled\n")
	}
	if sf.Forward != "" {
		_, _ = fmt.Fprintf(out, "Forwarding to: %s\n", sf.Forward)
	}
	_, _ = fmt.Fprintf(out, "\n")

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-shutdown:
		_, _ = fmt.Fprintf(out, "\nShutting down...\n")
	case <-cmd.Context().Done():
		_, _ = fmt.Fprintf(out, "\nShutting down...\n")
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	return nil
}

type webhookHandler struct {
	secret  string
	forward string
	quiet   bool
	out     io.Writer
	errOut  io.Writer
}

func (h *webhookHandler) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintln(w, "LINE Webhook Server")
	_, _ = fmt.Fprintln(w, "POST to /webhook to send events")
}

func (h *webhookHandler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Only accept POST requests
	if r.Method != http.MethodPost {
		h.logError(timestamp, r.Method, "/webhook", http.StatusMethodNotAllowed, "Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logError(timestamp, r.Method, "/webhook", http.StatusBadRequest, "Failed to read body")
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Validate signature if secret is provided
	if h.secret != "" {
		signature := r.Header.Get("X-Line-Signature")
		if signature == "" {
			h.logError(timestamp, r.Method, "/webhook", http.StatusUnauthorized, "Missing X-Line-Signature header")
			http.Error(w, "Missing signature", http.StatusUnauthorized)
			return
		}

		if !h.validateSignature(body, signature) {
			h.logError(timestamp, r.Method, "/webhook", http.StatusForbidden, "Invalid signature")
			http.Error(w, "Invalid signature", http.StatusForbidden)
			return
		}
	}

	// Parse and log events
	var payload LineWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		// If JSON parsing fails, still log the raw body
		h.logRequest(timestamp, http.StatusOK)
		if !h.quiet {
			_, _ = fmt.Fprintf(h.out, "Raw body: %s\n\n", string(body))
		}
	} else {
		h.logRequest(timestamp, http.StatusOK)
		if !h.quiet {
			h.logPayload(&payload)
		}
	}

	// Forward to another URL if configured
	if h.forward != "" {
		if err := h.forwardRequest(body, r.Header); err != nil {
			_, _ = fmt.Fprintf(h.errOut, "Forward error: %v\n", err)
		}
	}

	// Return 200 OK
	w.WriteHeader(http.StatusOK)
}

func (h *webhookHandler) validateSignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(body)
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (h *webhookHandler) logRequest(timestamp string, status int) {
	if h.quiet {
		return
	}
	_, _ = fmt.Fprintf(h.out, "[%s] POST /webhook - %d OK\n", timestamp, status)
}

func (h *webhookHandler) logError(timestamp, method, path string, status int, message string) {
	_, _ = fmt.Fprintf(h.errOut, "[%s] %s %s - %d %s\n", timestamp, method, path, status, message)
}

func (h *webhookHandler) logPayload(payload *LineWebhookPayload) {
	if payload.Destination != "" {
		_, _ = fmt.Fprintf(h.out, "Destination: %s\n", payload.Destination)
	}

	if len(payload.Events) == 0 {
		_, _ = fmt.Fprintf(h.out, "Events: (none)\n\n")
		return
	}

	for i, event := range payload.Events {
		if len(payload.Events) > 1 {
			_, _ = fmt.Fprintf(h.out, "--- Event %d ---\n", i+1)
		}
		h.logEvent(&event)
	}
	_, _ = fmt.Fprintf(h.out, "\n")
}

func (h *webhookHandler) logEvent(event *LineWebhookEvent) {
	_, _ = fmt.Fprintf(h.out, "Event Type: %s\n", event.Type)

	if event.Source != nil {
		_, _ = fmt.Fprintf(h.out, "Source: %s", event.Source.Type)
		switch event.Source.Type {
		case "user":
			if event.Source.UserID != "" {
				_, _ = fmt.Fprintf(h.out, " (User: %s)", event.Source.UserID)
			}
		case "group":
			if event.Source.GroupID != "" {
				_, _ = fmt.Fprintf(h.out, " (Group: %s", event.Source.GroupID)
				if event.Source.UserID != "" {
					_, _ = fmt.Fprintf(h.out, ", User: %s", event.Source.UserID)
				}
				_, _ = fmt.Fprintf(h.out, ")")
			}
		case "room":
			if event.Source.RoomID != "" {
				_, _ = fmt.Fprintf(h.out, " (Room: %s", event.Source.RoomID)
				if event.Source.UserID != "" {
					_, _ = fmt.Fprintf(h.out, ", User: %s", event.Source.UserID)
				}
				_, _ = fmt.Fprintf(h.out, ")")
			}
		}
		_, _ = fmt.Fprintf(h.out, "\n")
	}

	if event.ReplyToken != "" {
		_, _ = fmt.Fprintf(h.out, "Reply Token: %s\n", event.ReplyToken)
	}

	// Log event-specific data
	if len(event.Message) > 0 {
		_, _ = fmt.Fprintf(h.out, "Message: %s\n", formatJSON(event.Message))
	}
	if len(event.Postback) > 0 {
		_, _ = fmt.Fprintf(h.out, "Postback: %s\n", formatJSON(event.Postback))
	}
	if len(event.Beacon) > 0 {
		_, _ = fmt.Fprintf(h.out, "Beacon: %s\n", formatJSON(event.Beacon))
	}
	if len(event.Link) > 0 {
		_, _ = fmt.Fprintf(h.out, "Link: %s\n", formatJSON(event.Link))
	}
	if len(event.Things) > 0 {
		_, _ = fmt.Fprintf(h.out, "Things: %s\n", formatJSON(event.Things))
	}
	if len(event.Members) > 0 {
		_, _ = fmt.Fprintf(h.out, "Members: %s\n", formatJSON(event.Members))
	}
	if len(event.Unsend) > 0 {
		_, _ = fmt.Fprintf(h.out, "Unsend: %s\n", formatJSON(event.Unsend))
	}
	if len(event.VideoPlayComplete) > 0 {
		_, _ = fmt.Fprintf(h.out, "VideoPlayComplete: %s\n", formatJSON(event.VideoPlayComplete))
	}
}

func formatJSON(raw json.RawMessage) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return string(raw)
	}
	return buf.String()
}

func (h *webhookHandler) forwardRequest(body []byte, headers http.Header) error {
	req, err := http.NewRequest(http.MethodPost, h.forward, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create forward request: %w", err)
	}

	// Copy relevant headers
	req.Header.Set("Content-Type", "application/json")
	if sig := headers.Get("X-Line-Signature"); sig != "" {
		req.Header.Set("X-Line-Signature", sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("forward request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if !h.quiet {
		_, _ = fmt.Fprintf(h.out, "Forwarded to %s: %s\n", h.forward, resp.Status)
	}

	return nil
}
