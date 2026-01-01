package api

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestParseAPIError_ValidationError(t *testing.T) {
	body := []byte(`{
		"message": "The request body has 1 error(s)",
		"details": [
			{
				"message": "May not be empty",
				"property": "messages[0].text"
			}
		]
	}`)

	apiErr := ParseAPIError(http.StatusBadRequest, "POST", "/v2/bot/message/push", body)

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Method != "POST" {
		t.Errorf("expected method POST, got %s", apiErr.Method)
	}
	if apiErr.Endpoint != "/v2/bot/message/push" {
		t.Errorf("expected endpoint /v2/bot/message/push, got %s", apiErr.Endpoint)
	}
	if apiErr.Message != "The request body has 1 error(s)" {
		t.Errorf("expected message 'The request body has 1 error(s)', got %s", apiErr.Message)
	}
	if len(apiErr.Details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(apiErr.Details))
	}
	if apiErr.Details[0].Property != "messages[0].text" {
		t.Errorf("expected property 'messages[0].text', got %s", apiErr.Details[0].Property)
	}
	if apiErr.Details[0].Message != "May not be empty" {
		t.Errorf("expected message 'May not be empty', got %s", apiErr.Details[0].Message)
	}
}

func TestParseAPIError_Unauthorized(t *testing.T) {
	body := []byte(`{"message":"Authentication failed"}`)

	apiErr := ParseAPIError(http.StatusUnauthorized, "POST", "/v2/bot/message/push", body)

	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Authentication failed" {
		t.Errorf("expected message 'Authentication failed', got %s", apiErr.Message)
	}
	if !strings.Contains(apiErr.Hint, "channel access token") {
		t.Errorf("expected hint about channel access token, got %s", apiErr.Hint)
	}
}

func TestParseAPIError_RateLimit(t *testing.T) {
	body := []byte(`{"message":"Rate limit exceeded"}`)

	apiErr := ParseAPIError(http.StatusTooManyRequests, "POST", "/v2/bot/message/broadcast", body)

	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Hint, "Rate limit") {
		t.Errorf("expected hint about rate limit, got %s", apiErr.Hint)
	}
}

func TestParseAPIError_NotFound(t *testing.T) {
	body := []byte(`{"message":"Rich menu not found"}`)

	apiErr := ParseAPIError(http.StatusNotFound, "GET", "/v2/bot/richmenu/rm-12345", body)

	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Hint, "not found") {
		t.Errorf("expected hint about resource not found, got %s", apiErr.Hint)
	}
}

func TestParseAPIError_EmptyBody(t *testing.T) {
	body := []byte(`{}`)

	apiErr := ParseAPIError(http.StatusBadRequest, "POST", "/v2/bot/message/push", body)

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
	// Should fallback to status text
	if apiErr.Message != http.StatusText(http.StatusBadRequest) {
		t.Errorf("expected message 'Bad Request', got %s", apiErr.Message)
	}
}

func TestParseAPIError_InvalidJSON(t *testing.T) {
	body := []byte(`not json`)

	apiErr := ParseAPIError(http.StatusBadRequest, "POST", "/v2/bot/message/push", body)

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
	// Should use raw body as message
	if apiErr.Message != "not json" {
		t.Errorf("expected message 'not json', got %s", apiErr.Message)
	}
}

func TestAPIError_Error(t *testing.T) {
	apiErr := &APIError{
		StatusCode: http.StatusUnauthorized,
		Method:     "POST",
		Endpoint:   "/v2/bot/message/push",
		Message:    "Authentication failed",
		Hint:       "Check your channel access token is valid and not expired.",
	}

	errStr := apiErr.Error()

	if !strings.Contains(errStr, "401 Unauthorized") {
		t.Errorf("expected '401 Unauthorized' in error, got %s", errStr)
	}
	if !strings.Contains(errStr, "POST /v2/bot/message/push") {
		t.Errorf("expected 'POST /v2/bot/message/push' in error, got %s", errStr)
	}
	if !strings.Contains(errStr, "Authentication failed") {
		t.Errorf("expected 'Authentication failed' in error, got %s", errStr)
	}
	if !strings.Contains(errStr, "Check your channel access token") {
		t.Errorf("expected hint in error, got %s", errStr)
	}
}

func TestAPIError_ErrorWithDetails(t *testing.T) {
	apiErr := &APIError{
		StatusCode: http.StatusBadRequest,
		Method:     "POST",
		Endpoint:   "/v2/bot/message/push",
		Message:    "The request body has 2 error(s)",
		Details: []ErrorDetail{
			{Property: "messages[0].text", Message: "May not be empty"},
			{Property: "to", Message: "Invalid user ID format"},
		},
		Hint: "Check your request parameters and message format",
	}

	errStr := apiErr.Error()

	if !strings.Contains(errStr, "400 Bad Request") {
		t.Errorf("expected '400 Bad Request' in error, got %s", errStr)
	}
	if !strings.Contains(errStr, "Details:") {
		t.Errorf("expected 'Details:' in error, got %s", errStr)
	}
	if !strings.Contains(errStr, "messages[0].text: May not be empty") {
		t.Errorf("expected validation detail in error, got %s", errStr)
	}
	if !strings.Contains(errStr, "to: Invalid user ID format") {
		t.Errorf("expected validation detail in error, got %s", errStr)
	}
}

func TestIsAPIError(t *testing.T) {
	apiErr := &APIError{StatusCode: 400}
	regularErr := errors.New("regular error")

	if !IsAPIError(apiErr) {
		t.Error("expected IsAPIError to return true for APIError")
	}
	if IsAPIError(regularErr) {
		t.Error("expected IsAPIError to return false for regular error")
	}
	if IsAPIError(nil) {
		t.Error("expected IsAPIError to return false for nil")
	}
}

func TestAsAPIError(t *testing.T) {
	apiErr := &APIError{StatusCode: 400}
	regularErr := errors.New("regular error")

	if AsAPIError(apiErr) != apiErr {
		t.Error("expected AsAPIError to return the same APIError")
	}
	if AsAPIError(regularErr) != nil {
		t.Error("expected AsAPIError to return nil for regular error")
	}
	if AsAPIError(nil) != nil {
		t.Error("expected AsAPIError to return nil for nil")
	}
}

func TestAPIError_StatusCheckers(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		check      func(*APIError) bool
		expected   bool
	}{
		{"IsUnauthorized_true", 401, (*APIError).IsUnauthorized, true},
		{"IsUnauthorized_false", 400, (*APIError).IsUnauthorized, false},
		{"IsForbidden_true", 403, (*APIError).IsForbidden, true},
		{"IsForbidden_false", 401, (*APIError).IsForbidden, false},
		{"IsNotFound_true", 404, (*APIError).IsNotFound, true},
		{"IsNotFound_false", 400, (*APIError).IsNotFound, false},
		{"IsRateLimited_true", 429, (*APIError).IsRateLimited, true},
		{"IsRateLimited_false", 400, (*APIError).IsRateLimited, false},
		{"IsServerError_500", 500, (*APIError).IsServerError, true},
		{"IsServerError_503", 503, (*APIError).IsServerError, true},
		{"IsServerError_400", 400, (*APIError).IsServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &APIError{StatusCode: tt.statusCode}
			if got := tt.check(apiErr); got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestGetHintForStatusCode(t *testing.T) {
	tests := []struct {
		statusCode int
		contains   string
	}{
		{400, "request parameters"},
		{401, "channel access token"},
		{403, "permission"},
		{404, "not found"},
		{409, "conflict"},
		{429, "Rate limit"},
		{500, "server error"},
		{503, "temporarily unavailable"},
		{502, "Server error"}, // generic 5xx
		{418, ""},             // no hint for unusual codes
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.statusCode), func(t *testing.T) {
			hint := getHintForStatusCode(tt.statusCode)
			if tt.contains == "" {
				if hint != "" {
					t.Errorf("expected no hint, got %s", hint)
				}
			} else if !strings.Contains(hint, tt.contains) {
				t.Errorf("expected hint containing '%s', got '%s'", tt.contains, hint)
			}
		})
	}
}
