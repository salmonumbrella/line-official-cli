package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TokenResponse represents a token issuance response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	KeyID       string `json:"key_id,omitempty"`
}

// TokenInfo represents token verification info
type TokenInfo struct {
	ClientID  string `json:"client_id"`
	ExpiresIn int    `json:"expires_in"`
	Scope     string `json:"scope,omitempty"`
}

// KeyIDsResponse represents the response from listing valid token key IDs
type KeyIDsResponse struct {
	Kids []string `json:"kids"`
}

// postForm sends a form-encoded POST request (no Bearer auth)
func (c *Client) postForm(ctx context.Context, urlPath string, data url.Values) ([]byte, error) {
	bodyStr := data.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+urlPath, strings.NewReader(bodyStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c.debugLogRequest(req, []byte(bodyStr))

	// In dry-run mode, return mock success without sending request
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
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// getNoAuth sends a GET request without Bearer auth
func (c *Client) getNoAuth(ctx context.Context, urlPath string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.debugLogRequest(req, nil)

	// In dry-run mode, return mock success without sending request
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
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// IssueChannelToken issues a short-lived channel access token (v2)
// POST https://api.line.me/v2/oauth/accessToken
// Content-Type: application/x-www-form-urlencoded
// Body: grant_type=client_credentials&client_id=xxx&client_secret=xxx
func (c *Client) IssueChannelToken(ctx context.Context, clientID, clientSecret string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	respBody, err := c.postForm(ctx, "/v2/oauth/accessToken", data)
	if err != nil {
		return nil, err
	}

	var resp TokenResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// VerifyChannelToken verifies a channel access token (v2)
// POST https://api.line.me/v2/oauth/verify
// Body: access_token=xxx
func (c *Client) VerifyChannelToken(ctx context.Context, accessToken string) (*TokenInfo, error) {
	data := url.Values{}
	data.Set("access_token", accessToken)

	respBody, err := c.postForm(ctx, "/v2/oauth/verify", data)
	if err != nil {
		return nil, err
	}

	var resp TokenInfo
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// RevokeChannelToken revokes a channel access token (v2)
// POST https://api.line.me/v2/oauth/revoke
// Body: access_token=xxx
func (c *Client) RevokeChannelToken(ctx context.Context, accessToken string) error {
	data := url.Values{}
	data.Set("access_token", accessToken)

	_, err := c.postForm(ctx, "/v2/oauth/revoke", data)
	return err
}

// IssueChannelTokenByJWT issues a channel access token using JWT (v2.1)
// POST https://api.line.me/oauth2/v2.1/token
// Body: grant_type=client_credentials&client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer&client_assertion=<JWT>
func (c *Client) IssueChannelTokenByJWT(ctx context.Context, jwt string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", jwt)

	respBody, err := c.postForm(ctx, "/oauth2/v2.1/token", data)
	if err != nil {
		return nil, err
	}

	var resp TokenResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// VerifyChannelTokenByJWT verifies a v2.1 token
// GET https://api.line.me/oauth2/v2.1/verify?access_token=xxx
func (c *Client) VerifyChannelTokenByJWT(ctx context.Context, accessToken string) (*TokenInfo, error) {
	path := "/oauth2/v2.1/verify?access_token=" + url.QueryEscape(accessToken)

	respBody, err := c.getNoAuth(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp TokenInfo
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// RevokeChannelTokenByJWT revokes a v2.1 token
// POST https://api.line.me/oauth2/v2.1/revoke
// Body: access_token=xxx&client_id=xxx&client_secret=xxx
func (c *Client) RevokeChannelTokenByJWT(ctx context.Context, accessToken, clientID, clientSecret string) error {
	data := url.Values{}
	data.Set("access_token", accessToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	_, err := c.postForm(ctx, "/oauth2/v2.1/revoke", data)
	return err
}

// GetAllValidTokenKeyIDs gets all valid channel access token key IDs
// GET https://api.line.me/oauth2/v2.1/tokens/kid?client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer&client_assertion=<JWT>
func (c *Client) GetAllValidTokenKeyIDs(ctx context.Context, jwt string) ([]string, error) {
	path := "/oauth2/v2.1/tokens/kid?" +
		"client_assertion_type=" + url.QueryEscape("urn:ietf:params:oauth:client-assertion-type:jwt-bearer") +
		"&client_assertion=" + url.QueryEscape(jwt)

	respBody, err := c.getNoAuth(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp KeyIDsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp.Kids, nil
}

// IssueStatelessToken issues a stateless channel access token (v3)
// POST https://api.line.me/oauth2/v3/token
// Content-Type: application/x-www-form-urlencoded
// Body: grant_type=client_credentials&client_id=xxx&client_secret=xxx
// Note: Stateless tokens cannot be revoked and expire in 15 minutes.
func (c *Client) IssueStatelessToken(ctx context.Context, clientID, clientSecret string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	respBody, err := c.postForm(ctx, "/oauth2/v3/token", data)
	if err != nil {
		return nil, err
	}

	var resp TokenResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}
