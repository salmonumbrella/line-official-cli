package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// DetachModuleRequest represents the request body for detaching a module
type DetachModuleRequest struct {
	BotID string `json:"botId"`
}

// AcquireChatControlRequest represents the request body for acquiring chat control
type AcquireChatControlRequest struct {
	Expired bool `json:"expired,omitempty"`
}

// DetachModule detaches the module channel from a LINE Official Account.
// POST /v2/bot/channel/detach
// The module channel admin calls this API to detach the module channel from a LINE Official Account.
func (c *Client) DetachModule(ctx context.Context, botID string) error {
	req := DetachModuleRequest{BotID: botID}
	_, err := c.Post(ctx, "/v2/bot/channel/detach", req)
	return err
}

// AcquireModuleChatControl acquires chat control for a module.
// POST /v2/bot/chat/{chatId}/control/acquire
// When the Primary Channel has chat control, the module channel calls this API to acquire chat control.
// The chatId can be a userId, roomId, or groupId.
func (c *Client) AcquireModuleChatControl(ctx context.Context, chatID string, expired bool) error {
	path := "/v2/bot/chat/" + chatID + "/control/acquire"
	req := AcquireChatControlRequest{Expired: expired}
	_, err := c.Post(ctx, path, req)
	return err
}

// ReleaseModuleChatControl releases chat control for a module.
// POST /v2/bot/chat/{chatId}/control/release
// When the module channel has chat control, the module channel calls this API to return chat control
// to the Primary Channel.
func (c *Client) ReleaseModuleChatControl(ctx context.Context, chatID string) error {
	path := "/v2/bot/chat/" + chatID + "/control/release"
	_, err := c.Post(ctx, path, nil)
	return err
}

// ModuleTokenResponse represents the response from the module token exchange endpoint.
type ModuleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// ModuleBotInfo represents information about a bot with a module channel attached.
type ModuleBotInfo struct {
	UserID      string `json:"userId"`
	BasicID     string `json:"basicId"`
	PremiumID   string `json:"premiumId,omitempty"`
	DisplayName string `json:"displayName"`
	PictureURL  string `json:"pictureUrl,omitempty"`
}

// BotListResponse represents the response from the bot list endpoint.
type BotListResponse struct {
	Bots []ModuleBotInfo `json:"bots"`
	Next string          `json:"next,omitempty"`
}

// GetBotsWithModules returns a list of bots with module channels attached.
// GET /v2/bot/list
func (c *Client) GetBotsWithModules(ctx context.Context, limit int, start string) (*BotListResponse, error) {
	path := "/v2/bot/list"
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if start != "" {
		params.Set("start", start)
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp BotListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse bot list response: %w", err)
	}
	return &resp, nil
}

// ExchangeModuleToken exchanges an authorization code for a module access token.
// POST /module/auth/v1/token
// Content-Type: application/x-www-form-urlencoded
// This endpoint does NOT use Bearer token auth - it's used to obtain a token.
func (c *Client) ExchangeModuleToken(ctx context.Context, code, redirectURI, clientID, clientSecret string) (*ModuleTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	respBody, err := c.postForm(ctx, "/module/auth/v1/token", data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange module token: %w", err)
	}

	var resp ModuleTokenResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse module token response: %w", err)
	}

	return &resp, nil
}
