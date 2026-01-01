package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// APIError represents a structured error from the LINE API.
type APIError struct {
	StatusCode int
	Method     string
	Endpoint   string
	Message    string
	Details    []ErrorDetail
	Hint       string
	RawBody    string
}

// ErrorDetail represents a specific validation error detail from the LINE API.
type ErrorDetail struct {
	Message  string `json:"message"`
	Property string `json:"property"`
}

// lineAPIErrorResponse represents the JSON structure of LINE API error responses.
type lineAPIErrorResponse struct {
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details"`
}

// Error implements the error interface with a formatted, actionable message.
func (e *APIError) Error() string {
	var sb strings.Builder

	// Status line
	sb.WriteString(fmt.Sprintf("API Error: %d %s\n", e.StatusCode, http.StatusText(e.StatusCode)))

	// Endpoint
	sb.WriteString(fmt.Sprintf("Endpoint: %s %s\n", e.Method, e.Endpoint))

	// Message
	if e.Message != "" {
		sb.WriteString(fmt.Sprintf("Message: %s\n", e.Message))
	}

	// Details (validation errors)
	if len(e.Details) > 0 {
		sb.WriteString("Details:\n")
		for _, d := range e.Details {
			if d.Property != "" {
				sb.WriteString(fmt.Sprintf("  - %s: %s\n", d.Property, d.Message))
			} else {
				sb.WriteString(fmt.Sprintf("  - %s\n", d.Message))
			}
		}
	}

	// Hint
	if e.Hint != "" {
		sb.WriteString(fmt.Sprintf("Hint: %s", e.Hint))
	}

	return strings.TrimRight(sb.String(), "\n")
}

// ParseAPIError creates a structured APIError from an HTTP response.
func ParseAPIError(statusCode int, method, endpoint string, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		Method:     method,
		Endpoint:   endpoint,
		RawBody:    string(body),
		Hint:       getHintForStatusCode(statusCode),
	}

	// Try to parse LINE API error response format
	var lineErr lineAPIErrorResponse
	if err := json.Unmarshal(body, &lineErr); err == nil && lineErr.Message != "" {
		apiErr.Message = lineErr.Message
		apiErr.Details = lineErr.Details
	} else {
		// Fallback: use raw body as message if it's not empty and not just "{}"
		rawStr := strings.TrimSpace(string(body))
		if rawStr != "" && rawStr != "{}" {
			apiErr.Message = rawStr
		} else {
			apiErr.Message = http.StatusText(statusCode)
		}
	}

	return apiErr
}

// getHintForStatusCode returns actionable hints for common HTTP status codes.
func getHintForStatusCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest: // 400
		return "Check your request parameters and message format"
	case http.StatusUnauthorized: // 401
		return "Check your channel access token is valid and not expired. Run 'line auth login' to re-authenticate."
	case http.StatusForbidden: // 403
		return "Your channel access token doesn't have permission for this API. Check your LINE Developer Console permissions."
	case http.StatusNotFound: // 404
		return "The requested resource was not found. Check the ID or path is correct."
	case http.StatusConflict: // 409
		return "A resource conflict occurred. The resource may already exist or be in an incompatible state."
	case http.StatusTooManyRequests: // 429
		return "Rate limit exceeded. Wait a moment before retrying, or reduce your request frequency."
	case http.StatusInternalServerError: // 500
		return "LINE API server error. Try again later or check LINE status page."
	case http.StatusServiceUnavailable: // 503
		return "LINE API is temporarily unavailable. Try again later."
	default:
		if statusCode >= 500 {
			return "Server error. Try again later."
		}
		return ""
	}
}

// IsAPIError checks if an error is an APIError.
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// AsAPIError returns the error as an APIError if it is one, otherwise nil.
func AsAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return nil
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden returns true if the error is a 403 Forbidden error.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsNotFound returns true if the error is a 404 Not Found error.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsRateLimited returns true if the error is a 429 Too Many Requests error.
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsServerError returns true if the error is a 5xx server error.
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}
