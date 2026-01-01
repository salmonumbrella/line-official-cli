package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salmonumbrella/line-official-cli/internal/api"
)

func TestContentCmd_RequiresSubcommand(t *testing.T) {
	cmd := newContentCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	// Running without subcommand should show help, not error
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestContentCmd_HasSubcommands(t *testing.T) {
	cmd := newContentCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 3 {
		t.Errorf("expected at least 3 subcommands (download, preview, status), got %d", len(subcommands))
	}

	names := make(map[string]bool)
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = true
	}

	expected := []string{"download", "preview", "status"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected '%s' subcommand", name)
		}
	}
}

func TestContentDownloadCmd_RequiresMessageID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"content", "download"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --message-id flag")
	}
}

func TestContentPreviewCmd_RequiresMessageID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"content", "preview"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --message-id flag")
	}
}

func TestContentStatusCmd_RequiresMessageID(t *testing.T) {
	cmd := NewRootCmd()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"content", "status"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --message-id flag")
	}
}

// Execution tests using mock servers

func TestContentDownloadCmd_Execute(t *testing.T) {
	testContent := []byte("fake image content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/message/") && strings.HasSuffix(r.URL.Path, "/content") {
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write(testContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Create temp dir for downloads
	tmpDir, err := os.MkdirTemp("", "content-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Change to temp dir so files are written there
	originalDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(originalDir) }()

	tests := []struct {
		name      string
		output    string
		wantJSON  bool
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			wantJSON:  false,
			checkText: "Downloaded",
		},
		{
			name:     "json output",
			output:   "json",
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = tt.output
			defer func() { flags.Output = oldOutput }()

			cmd := newContentDownloadCmdWithClient(client)
			cmd.SetArgs([]string{"--message-id", "msg123"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := out.String()
			if tt.wantJSON {
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("expected valid JSON output, got: %s", output)
				}
				if result["messageId"] != "msg123" {
					t.Errorf("expected messageId 'msg123', got: %v", result["messageId"])
				}
				if result["contentType"] != "image/jpeg" {
					t.Errorf("expected contentType 'image/jpeg', got: %v", result["contentType"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}

			// Verify file was created
			expectedFile := "msg123.jpg"
			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				t.Errorf("expected file %s to be created", expectedFile)
			} else {
				_ = os.Remove(expectedFile)
			}
		})
	}
}

func TestContentDownloadCmd_CustomOutput(t *testing.T) {
	testContent := []byte("fake image content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/message/") && strings.HasSuffix(r.URL.Path, "/content") {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(testContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Create temp dir for downloads
	tmpDir, err := os.MkdirTemp("", "content-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	outputFile := filepath.Join(tmpDir, "custom.png")

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newContentDownloadCmdWithClient(client)
	cmd.SetArgs([]string{"--message-id", "msg456", "--output", outputFile})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created with custom name
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("expected file %s to be created", outputFile)
	}

	// Verify content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !bytes.Equal(content, testContent) {
		t.Errorf("file content mismatch")
	}
}

func TestContentDownloadCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Content not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newContentDownloadCmdWithClient(client)
	cmd.SetArgs([]string{"--message-id", "nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to download content") {
		t.Errorf("expected 'failed to download content' in error, got: %v", err)
	}
}

func TestContentPreviewCmd_Execute(t *testing.T) {
	testContent := []byte("fake preview content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/message/") && strings.HasSuffix(r.URL.Path, "/content/preview") {
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write(testContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	// Create temp dir for downloads
	tmpDir, err := os.MkdirTemp("", "content-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Change to temp dir so files are written there
	originalDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(originalDir) }()

	tests := []struct {
		name      string
		output    string
		wantJSON  bool
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			wantJSON:  false,
			checkText: "Downloaded preview",
		},
		{
			name:     "json output",
			output:   "json",
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = tt.output
			defer func() { flags.Output = oldOutput }()

			cmd := newContentPreviewCmdWithClient(client)
			cmd.SetArgs([]string{"--message-id", "prev123"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := out.String()
			if tt.wantJSON {
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("expected valid JSON output, got: %s", output)
				}
				if result["messageId"] != "prev123" {
					t.Errorf("expected messageId 'prev123', got: %v", result["messageId"])
				}
				if result["contentType"] != "image/jpeg" {
					t.Errorf("expected contentType 'image/jpeg', got: %v", result["contentType"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}

			// Verify file was created
			expectedFile := "preview-prev123.jpg"
			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				t.Errorf("expected file %s to be created", expectedFile)
			} else {
				_ = os.Remove(expectedFile)
			}
		})
	}
}

func TestContentPreviewCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Preview not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newContentPreviewCmdWithClient(client)
	cmd.SetArgs([]string{"--message-id", "nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to download preview") {
		t.Errorf("expected 'failed to download preview' in error, got: %v", err)
	}
}

func TestContentStatusCmd_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/message/") && strings.HasSuffix(r.URL.Path, "/content/transcoding") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "succeeded",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	tests := []struct {
		name      string
		output    string
		wantJSON  bool
		checkText string
	}{
		{
			name:      "text output",
			output:    "text",
			wantJSON:  false,
			checkText: "Transcoding Status: succeeded",
		},
		{
			name:     "json output",
			output:   "json",
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOutput := flags.Output
			flags.Output = tt.output
			defer func() { flags.Output = oldOutput }()

			cmd := newContentStatusCmdWithClient(client)
			cmd.SetArgs([]string{"--message-id", "status123"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := out.String()
			if tt.wantJSON {
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("expected valid JSON output, got: %s", output)
				}
				if result["messageId"] != "status123" {
					t.Errorf("expected messageId 'status123', got: %v", result["messageId"])
				}
				if result["status"] != "succeeded" {
					t.Errorf("expected status 'succeeded', got: %v", result["status"])
				}
			} else {
				if !strings.Contains(output, tt.checkText) {
					t.Errorf("expected output to contain %q, got: %s", tt.checkText, output)
				}
			}
		})
	}
}

func TestContentStatusCmd_Processing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/bot/message/") && strings.HasSuffix(r.URL.Path, "/content/transcoding") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "processing",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	oldOutput := flags.Output
	flags.Output = "text"
	defer func() { flags.Output = oldOutput }()

	cmd := newContentStatusCmdWithClient(client)
	cmd.SetArgs([]string{"--message-id", "processing123"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "processing") {
		t.Errorf("expected 'processing' status, got: %s", output)
	}
}

func TestContentStatusCmd_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Message not found"})
	}))
	defer server.Close()

	client := api.NewClient("test-token", false, false)
	client.SetBaseURL(server.URL)

	cmd := newContentStatusCmdWithClient(client)
	cmd.SetArgs([]string{"--message-id", "nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "failed to get transcoding status") {
		t.Errorf("expected 'failed to get transcoding status' in error, got: %v", err)
	}
}

func TestContentDownloadCmd_FileExtensions(t *testing.T) {
	tests := []struct {
		contentType string
		expectedExt string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"video/mp4", ".mp4"},
		{"audio/m4a", ".m4a"},
		{"application/octet-stream", ".bin"},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/v2/bot/message/") && strings.HasSuffix(r.URL.Path, "/content") {
					w.Header().Set("Content-Type", tt.contentType)
					_, _ = w.Write([]byte("test content"))
					return
				}
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			client := api.NewClient("test-token", false, false)
			client.SetBaseURL(server.URL)

			// Create temp dir for downloads
			tmpDir, err := os.MkdirTemp("", "content-test")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			// Change to temp dir
			originalDir, _ := os.Getwd()
			_ = os.Chdir(tmpDir)
			defer func() { _ = os.Chdir(originalDir) }()

			oldOutput := flags.Output
			flags.Output = "json"
			defer func() { flags.Output = oldOutput }()

			cmd := newContentDownloadCmdWithClient(client)
			cmd.SetArgs([]string{"--message-id", "ext-test"})
			var out bytes.Buffer
			cmd.SetOut(&out)

			err = cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var result map[string]any
			if err := json.Unmarshal(out.Bytes(), &result); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}

			expectedFile := "ext-test" + tt.expectedExt
			if result["file"] != expectedFile {
				t.Errorf("expected file %q for content-type %q, got: %v", expectedFile, tt.contentType, result["file"])
			}
		})
	}
}
