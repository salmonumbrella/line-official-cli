package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const BaseURL = "https://api.line.me"

type Client struct {
	httpClient         *http.Client
	channelAccessToken string
	baseURL            string
	debug              bool
	dryRun             bool
}

func NewClient(channelAccessToken string, debug bool, dryRun bool) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		channelAccessToken: channelAccessToken,
		baseURL:            BaseURL,
		debug:              debug || dryRun, // dry-run implies debug
		dryRun:             dryRun,
	}
}

// SetBaseURL sets the base URL for API requests (used for testing)
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

const debugMaxBodyLen = 500

// debugLog prints a debug message to stderr with [DEBUG] prefix
func (c *Client) debugLog(format string, args ...any) {
	if !c.debug {
		return
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
}

// debugLogRequest logs HTTP request details (method, URL, headers, body preview)
func (c *Client) debugLogRequest(req *http.Request, body []byte) {
	if !c.debug {
		return
	}
	c.debugLog(">>> %s %s", req.Method, req.URL.String())
	c.debugLogHeaders(">>> ", req.Header, true)
	if len(body) > 0 {
		c.debugLogBody(">>> Body: ", body)
	}
}

// debugLogResponse logs HTTP response details (status, headers, body preview)
func (c *Client) debugLogResponse(resp *http.Response, body []byte) {
	if !c.debug {
		return
	}
	c.debugLog("<<< %s", resp.Status)
	c.debugLogHeaders("<<< ", resp.Header, false)
	if len(body) > 0 {
		c.debugLogBody("<<< Body: ", body)
	}
}

// debugLogHeaders logs headers, redacting Authorization token
func (c *Client) debugLogHeaders(prefix string, headers http.Header, redactAuth bool) {
	for name, values := range headers {
		for _, value := range values {
			if redactAuth && strings.EqualFold(name, "Authorization") {
				// Redact the token but show it's a Bearer token
				if strings.HasPrefix(value, "Bearer ") {
					c.debugLog("%s%s: Bearer [REDACTED]", prefix, name)
				} else {
					c.debugLog("%s%s: [REDACTED]", prefix, name)
				}
			} else {
				c.debugLog("%s%s: %s", prefix, name, value)
			}
		}
	}
}

// debugLogBody logs body content, truncating if too long
func (c *Client) debugLogBody(prefix string, body []byte) {
	bodyStr := string(body)
	if len(bodyStr) > debugMaxBodyLen {
		c.debugLog("%s%s... (%d bytes truncated)", prefix, bodyStr[:debugMaxBodyLen], len(bodyStr)-debugMaxBodyLen)
	} else {
		c.debugLog("%s%s", prefix, bodyStr)
	}
}

// dryRunLog prints a dry-run message to stderr
func (c *Client) dryRunLog(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[DRY-RUN] "+format+"\n", args...)
}

// mockDryRunResponse returns a mock response for dry-run mode
func (c *Client) mockDryRunResponse(method string) *Response {
	c.dryRunLog("Request not sent")
	// Return empty response with 200 status implied
	return &Response{
		Body:    []byte("{}"),
		Headers: make(http.Header),
	}
}

// Response wraps the HTTP response body and headers
type Response struct {
	Body    []byte
	Headers http.Header
}

func (c *Client) doWithHeaders(ctx context.Context, method, path string, body any) (*Response, error) {
	var bodyReader io.Reader
	var bodyData []byte
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyData = data
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)
	req.Header.Set("Content-Type", "application/json")

	c.debugLogRequest(req, bodyData)

	// In dry-run mode, return mock response without sending request
	if c.dryRun {
		return c.mockDryRunResponse(method), nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.debugLogResponse(resp, respBody)

	if resp.StatusCode >= 400 {
		return nil, ParseAPIError(resp.StatusCode, method, path, respBody)
	}

	return &Response{Body: respBody, Headers: resp.Header}, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	resp, err := c.doWithHeaders(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) Post(ctx context.Context, path string, body any) ([]byte, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) PostWithHeaders(ctx context.Context, path string, body any) (*Response, error) {
	return c.doWithHeaders(ctx, http.MethodPost, path, body)
}

func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.do(ctx, http.MethodDelete, path, nil)
}

func (c *Client) Put(ctx context.Context, path string, body any) ([]byte, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

// BotInfo represents information about a LINE Official Account bot
type BotInfo struct {
	UserID         string `json:"userId"`
	BasicID        string `json:"basicId"`
	PremiumID      string `json:"premiumId,omitempty"`
	DisplayName    string `json:"displayName"`
	PictureURL     string `json:"pictureUrl,omitempty"`
	ChatMode       string `json:"chatMode"`
	MarkAsReadMode string `json:"markAsReadMode"`
}

// GetBotInfo retrieves basic information about the LINE Official Account
func (c *Client) GetBotInfo(ctx context.Context) (*BotInfo, error) {
	data, err := c.Get(ctx, "/v2/bot/info")
	if err != nil {
		return nil, err
	}
	var info BotInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse bot info: %w", err)
	}
	return &info, nil
}

// UserProfile represents a LINE user's profile information
type UserProfile struct {
	UserID        string `json:"userId"`
	DisplayName   string `json:"displayName"`
	PictureURL    string `json:"pictureUrl,omitempty"`
	StatusMessage string `json:"statusMessage,omitempty"`
	Language      string `json:"language,omitempty"`
}

// GetUserProfile retrieves profile information for a specific user
func (c *Client) GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	data, err := c.Get(ctx, "/v2/bot/profile/"+userID)
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}
	return &profile, nil
}

// FollowerIDsResponse represents the response from the followers endpoint
type FollowerIDsResponse struct {
	UserIDs []string `json:"userIds"`
	Next    string   `json:"next,omitempty"`
}

// GetFollowerIDs retrieves a list of user IDs of users who have added the bot as a friend
func (c *Client) GetFollowerIDs(ctx context.Context, start string, limit int) (*FollowerIDsResponse, error) {
	path := "/v2/bot/followers/ids"
	if start != "" || limit > 0 {
		path += "?"
		if start != "" {
			path += "start=" + start
		}
		if limit > 0 {
			if start != "" {
				path += "&"
			}
			path += fmt.Sprintf("limit=%d", limit)
		}
	}
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp FollowerIDsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse followers: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetBinary(ctx context.Context, path string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)

	c.debugLogRequest(req, nil)

	// In dry-run mode, return empty binary response
	if c.dryRun {
		c.dryRunLog("Request not sent")
		return []byte{}, "application/octet-stream", nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		c.debugLogResponse(resp, body)
		return nil, "", ParseAPIError(resp.StatusCode, http.MethodGet, path, body)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	// For binary responses, log status and headers but not body (it's binary data)
	c.debugLog("<<< %s", resp.Status)
	c.debugLogHeaders("<<< ", resp.Header, false)
	c.debugLog("<<< Body: [binary data, %d bytes]", len(data))

	contentType := resp.Header.Get("Content-Type")
	return data, contentType, nil
}

func (c *Client) GetMessageContent(ctx context.Context, messageID string) ([]byte, string, error) {
	// Use data API endpoint for content downloads (only swap if using production URL)
	if c.baseURL == BaseURL {
		originalBaseURL := c.baseURL
		c.baseURL = "https://api-data.line.me"
		defer func() { c.baseURL = originalBaseURL }()
	}

	return c.GetBinary(ctx, "/v2/bot/message/"+messageID+"/content")
}

func (c *Client) PostBinary(ctx context.Context, path string, contentType string, data []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)
	req.Header.Set("Content-Type", contentType)

	// Log request with binary body indicator
	c.debugLog(">>> %s %s", req.Method, req.URL.String())
	c.debugLogHeaders(">>> ", req.Header, true)
	c.debugLog(">>> Body: [binary data, %d bytes]", len(data))

	// In dry-run mode, return mock success
	if c.dryRun {
		c.dryRunLog("Request not sent")
		return []byte("{}"), nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.debugLogResponse(resp, respBody)

	if resp.StatusCode >= 400 {
		return nil, ParseAPIError(resp.StatusCode, http.MethodPost, path, respBody)
	}

	return respBody, nil
}

// GetMessageContentPreview downloads preview image for message media.
// Uses the data API endpoint: https://api-data.line.me/v2/bot/message/{messageId}/content/preview
func (c *Client) GetMessageContentPreview(ctx context.Context, messageID string) ([]byte, string, error) {
	// Use data API endpoint for content downloads (only swap if using production URL)
	if c.baseURL == BaseURL {
		originalBaseURL := c.baseURL
		c.baseURL = "https://api-data.line.me"
		defer func() { c.baseURL = originalBaseURL }()
	}

	return c.GetBinary(ctx, "/v2/bot/message/"+messageID+"/content/preview")
}

// TranscodingStatus represents the transcoding status of media content.
type TranscodingStatus struct {
	Status string `json:"status"` // "processing", "succeeded", "failed"
}

// GetMessageContentTranscoding checks if media is ready for download.
// GET /v2/bot/message/{messageId}/content/transcoding
func (c *Client) GetMessageContentTranscoding(ctx context.Context, messageID string) (*TranscodingStatus, error) {
	data, err := c.Get(ctx, "/v2/bot/message/"+messageID+"/content/transcoding")
	if err != nil {
		return nil, err
	}
	var status TranscodingStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse transcoding status: %w", err)
	}
	return &status, nil
}

// LinkTokenResponse represents the response from the link token endpoint
type LinkTokenResponse struct {
	LinkToken string `json:"linkToken"`
}

// IssueLinkToken generates an account linking token for a user.
// POST /v2/bot/user/{userId}/linkToken
func (c *Client) IssueLinkToken(ctx context.Context, userID string) (string, error) {
	data, err := c.Post(ctx, "/v2/bot/user/"+userID+"/linkToken", nil)
	if err != nil {
		return "", err
	}
	var resp LinkTokenResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse link token response: %w", err)
	}
	return resp.LinkToken, nil
}

// PostMultipart sends a multipart/form-data POST request with file content and form fields.
func (c *Client) PostMultipart(ctx context.Context, path string, fieldName, fileName string, fileContent []byte, formFields map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add form fields first
	for key, value := range formFields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("failed to write form field %s: %w", key, err)
		}
	}

	// Add the file
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(fileContent); err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Log multipart request
	c.debugLog(">>> %s %s", req.Method, req.URL.String())
	c.debugLogHeaders(">>> ", req.Header, true)
	c.debugLog(">>> Body: [multipart/form-data, file=%s, %d bytes]", fileName, len(fileContent))

	// In dry-run mode, return mock success
	if c.dryRun {
		c.dryRunLog("Request not sent")
		return []byte("{}"), nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.debugLogResponse(resp, respBody)

	if resp.StatusCode >= 400 {
		return nil, ParseAPIError(resp.StatusCode, http.MethodPost, path, respBody)
	}

	return respBody, nil
}

// PutMultipart sends a multipart/form-data PUT request with file content and form fields.
func (c *Client) PutMultipart(ctx context.Context, path string, fieldName, fileName string, fileContent []byte, formFields map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add form fields first
	for key, value := range formFields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("failed to write form field %s: %w", key, err)
		}
	}

	// Add the file
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(fileContent); err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Log multipart request
	c.debugLog(">>> %s %s", req.Method, req.URL.String())
	c.debugLogHeaders(">>> ", req.Header, true)
	c.debugLog(">>> Body: [multipart/form-data, file=%s, %d bytes]", fileName, len(fileContent))

	// In dry-run mode, return mock success
	if c.dryRun {
		c.dryRunLog("Request not sent")
		return []byte("{}"), nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.debugLogResponse(resp, respBody)

	if resp.StatusCode >= 400 {
		return nil, ParseAPIError(resp.StatusCode, http.MethodPut, path, respBody)
	}

	return respBody, nil
}
