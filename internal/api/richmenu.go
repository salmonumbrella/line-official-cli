package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type RichMenu struct {
	RichMenuID  string         `json:"richMenuId"`
	Name        string         `json:"name"`
	Size        RichMenuSize   `json:"size"`
	ChatBarText string         `json:"chatBarText"`
	Selected    bool           `json:"selected"`
	Areas       []RichMenuArea `json:"areas"`
}

type RichMenuSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type RichMenuArea struct {
	Bounds RichMenuBounds  `json:"bounds"`
	Action json.RawMessage `json:"action"`
}

type RichMenuBounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type RichMenuListResponse struct {
	RichMenus []RichMenu `json:"richmenus"`
}

type CreateRichMenuRequest struct {
	Size        RichMenuSize   `json:"size"`
	Selected    bool           `json:"selected"`
	Name        string         `json:"name"`
	ChatBarText string         `json:"chatBarText"`
	Areas       []RichMenuArea `json:"areas"`
}

type CreateRichMenuResponse struct {
	RichMenuID string `json:"richMenuId"`
}

func (c *Client) GetRichMenuList(ctx context.Context) ([]RichMenu, error) {
	data, err := c.Get(ctx, "/v2/bot/richmenu/list")
	if err != nil {
		return nil, err
	}
	var resp RichMenuListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse rich menus: %w", err)
	}
	return resp.RichMenus, nil
}

func (c *Client) CreateRichMenu(ctx context.Context, req CreateRichMenuRequest) (string, error) {
	data, err := c.Post(ctx, "/v2/bot/richmenu", req)
	if err != nil {
		return "", err
	}
	var resp CreateRichMenuResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.RichMenuID, nil
}

func (c *Client) DeleteRichMenu(ctx context.Context, richMenuID string) error {
	_, err := c.Delete(ctx, "/v2/bot/richmenu/"+richMenuID)
	return err
}

func (c *Client) SetDefaultRichMenu(ctx context.Context, richMenuID string) error {
	_, err := c.Post(ctx, "/v2/bot/user/all/richmenu/"+richMenuID, nil)
	return err
}

func (c *Client) CancelDefaultRichMenu(ctx context.Context) error {
	_, err := c.Delete(ctx, "/v2/bot/user/all/richmenu")
	return err
}

func (c *Client) GetDefaultRichMenuID(ctx context.Context) (string, error) {
	data, err := c.Get(ctx, "/v2/bot/user/all/richmenu")
	if err != nil {
		return "", err
	}
	var resp struct {
		RichMenuID string `json:"richMenuId"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.RichMenuID, nil
}

// UploadRichMenuImage uploads an image for a rich menu
// The image must be 2500x1686 (full) or 2500x843 (compact) pixels, PNG or JPEG, max 1MB
func (c *Client) UploadRichMenuImage(ctx context.Context, richMenuID string, contentType string, imageData []byte) error {
	// Use data API endpoint for binary uploads (only switch if using production URL)
	originalBaseURL := c.baseURL
	if c.baseURL == BaseURL {
		c.baseURL = "https://api-data.line.me"
		defer func() { c.baseURL = originalBaseURL }()
	}

	path := "/v2/bot/richmenu/" + richMenuID + "/content"
	_, err := c.PostBinary(ctx, path, contentType, imageData)
	return err
}

func (c *Client) GetRichMenu(ctx context.Context, richMenuID string) (*RichMenu, error) {
	data, err := c.Get(ctx, "/v2/bot/richmenu/"+richMenuID)
	if err != nil {
		return nil, err
	}
	var menu RichMenu
	if err := json.Unmarshal(data, &menu); err != nil {
		return nil, fmt.Errorf("failed to parse rich menu: %w", err)
	}
	return &menu, nil
}

func (c *Client) LinkRichMenuToUser(ctx context.Context, userID, richMenuID string) error {
	path := fmt.Sprintf("/v2/bot/user/%s/richmenu/%s", userID, richMenuID)
	_, err := c.Post(ctx, path, nil)
	return err
}

func (c *Client) UnlinkRichMenuFromUser(ctx context.Context, userID string) error {
	path := fmt.Sprintf("/v2/bot/user/%s/richmenu", userID)
	_, err := c.Delete(ctx, path)
	return err
}

func (c *Client) GetUserRichMenu(ctx context.Context, userID string) (string, error) {
	path := fmt.Sprintf("/v2/bot/user/%s/richmenu", userID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return "", err
	}
	var resp struct {
		RichMenuID string `json:"richMenuId"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.RichMenuID, nil
}

// Rich Menu Alias types
type RichMenuAlias struct {
	RichMenuAliasID string `json:"richMenuAliasId"`
	RichMenuID      string `json:"richMenuId"`
}

type RichMenuAliasListResponse struct {
	Aliases []RichMenuAlias `json:"aliases"`
}

type CreateRichMenuAliasRequest struct {
	RichMenuAliasID string `json:"richMenuAliasId"`
	RichMenuID      string `json:"richMenuId"`
}

type UpdateRichMenuAliasRequest struct {
	RichMenuID string `json:"richMenuId"`
}

func (c *Client) CreateRichMenuAlias(ctx context.Context, aliasID, richMenuID string) error {
	req := CreateRichMenuAliasRequest{
		RichMenuAliasID: aliasID,
		RichMenuID:      richMenuID,
	}
	_, err := c.Post(ctx, "/v2/bot/richmenu/alias", req)
	return err
}

func (c *Client) GetRichMenuAlias(ctx context.Context, aliasID string) (*RichMenuAlias, error) {
	path := fmt.Sprintf("/v2/bot/richmenu/alias/%s", aliasID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var alias RichMenuAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &alias, nil
}

func (c *Client) UpdateRichMenuAlias(ctx context.Context, aliasID, richMenuID string) error {
	path := fmt.Sprintf("/v2/bot/richmenu/alias/%s", aliasID)
	req := UpdateRichMenuAliasRequest{RichMenuID: richMenuID}
	_, err := c.Post(ctx, path, req)
	return err
}

func (c *Client) DeleteRichMenuAlias(ctx context.Context, aliasID string) error {
	path := fmt.Sprintf("/v2/bot/richmenu/alias/%s", aliasID)
	_, err := c.Delete(ctx, path)
	return err
}

func (c *Client) ListRichMenuAliases(ctx context.Context) ([]RichMenuAlias, error) {
	data, err := c.Get(ctx, "/v2/bot/richmenu/alias/list")
	if err != nil {
		return nil, err
	}
	var resp RichMenuAliasListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Aliases, nil
}

// Bulk operations - link/unlink menu to/from multiple users at once

// maxBulkUserIDs is the maximum number of user IDs allowed in a single bulk request
const maxBulkUserIDs = 500

// LinkRichMenuToUsers links a rich menu to multiple users at once
// POST /v2/bot/richmenu/bulk/link
func (c *Client) LinkRichMenuToUsers(ctx context.Context, richMenuID string, userIDs []string) error {
	if len(userIDs) > maxBulkUserIDs {
		return fmt.Errorf("too many user IDs: max %d, got %d", maxBulkUserIDs, len(userIDs))
	}
	req := struct {
		RichMenuID string   `json:"richMenuId"`
		UserIDs    []string `json:"userIds"`
	}{
		RichMenuID: richMenuID,
		UserIDs:    userIDs,
	}
	_, err := c.Post(ctx, "/v2/bot/richmenu/bulk/link", req)
	return err
}

// UnlinkRichMenuFromUsers unlinks rich menus from multiple users at once
// POST /v2/bot/richmenu/bulk/unlink
func (c *Client) UnlinkRichMenuFromUsers(ctx context.Context, userIDs []string) error {
	if len(userIDs) > maxBulkUserIDs {
		return fmt.Errorf("too many user IDs: max %d, got %d", maxBulkUserIDs, len(userIDs))
	}
	req := struct {
		UserIDs []string `json:"userIds"`
	}{
		UserIDs: userIDs,
	}
	_, err := c.Post(ctx, "/v2/bot/richmenu/bulk/unlink", req)
	return err
}

// Batch operations - atomically replace or unlink menus for many users

// RichMenuBatchOperation represents a single operation in a batch request
type RichMenuBatchOperation struct {
	Type       string   `json:"type"`                 // "link" or "unlink"
	RichMenuID string   `json:"richMenuId,omitempty"` // required for "link"
	UserIDs    []string `json:"userIds"`
}

// BatchProgress represents the progress of a batch operation
type BatchProgress struct {
	Phase         string `json:"phase"` // "ongoing", "succeeded", "failed"
	AcceptedTime  string `json:"acceptedTime"`
	CompletedTime string `json:"completedTime,omitempty"`
}

// RichMenuBatch executes batch operations atomically
// POST /v2/bot/richmenu/batch
// Returns the requestId for tracking progress
func (c *Client) RichMenuBatch(ctx context.Context, operations []RichMenuBatchOperation, resumeRequestID string) (string, error) {
	req := struct {
		Operations      []RichMenuBatchOperation `json:"operations"`
		ResumeRequestID string                   `json:"resumeRequestId,omitempty"`
	}{
		Operations:      operations,
		ResumeRequestID: resumeRequestID,
	}
	data, err := c.Post(ctx, "/v2/bot/richmenu/batch", req)
	if err != nil {
		return "", err
	}
	var resp struct {
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.RequestID, nil
}

// ValidateRichMenuBatch validates batch operations without executing
// POST /v2/bot/richmenu/validate/batch
func (c *Client) ValidateRichMenuBatch(ctx context.Context, operations []RichMenuBatchOperation) error {
	req := struct {
		Operations []RichMenuBatchOperation `json:"operations"`
	}{
		Operations: operations,
	}
	_, err := c.Post(ctx, "/v2/bot/richmenu/validate/batch", req)
	return err
}

// GetRichMenuBatchProgress gets the progress of a batch operation
// GET /v2/bot/richmenu/progress/batch?requestId=xxx
func (c *Client) GetRichMenuBatchProgress(ctx context.Context, requestID string) (*BatchProgress, error) {
	path := fmt.Sprintf("/v2/bot/richmenu/progress/batch?requestId=%s", requestID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var progress BatchProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &progress, nil
}

// Rich menu validation

// ValidateRichMenu validates a rich menu definition without creating it
// POST /v2/bot/richmenu/validate
func (c *Client) ValidateRichMenu(ctx context.Context, menu *CreateRichMenuRequest) error {
	_, err := c.Post(ctx, "/v2/bot/richmenu/validate", menu)
	return err
}

// Download rich menu image

// DownloadRichMenuImage downloads the image for a rich menu
// GET /v2/bot/richmenu/{richMenuId}/content from api-data.line.me
// Returns: image bytes, content-type, error
func (c *Client) DownloadRichMenuImage(ctx context.Context, richMenuID string) ([]byte, string, error) {
	// Use data API endpoint for binary downloads (only switch if using production URL)
	originalBaseURL := c.baseURL
	if c.baseURL == BaseURL {
		c.baseURL = "https://api-data.line.me"
		defer func() { c.baseURL = originalBaseURL }()
	}

	path := "/v2/bot/richmenu/" + richMenuID + "/content"
	return c.GetBinary(ctx, path)
}
