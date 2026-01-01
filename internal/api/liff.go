package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// LIFFView represents the view configuration for a LIFF app
type LIFFView struct {
	Type string `json:"type"` // compact, tall, full
	URL  string `json:"url"`
}

// LIFFApp represents a LIFF (LINE Front-end Framework) application
type LIFFApp struct {
	LIFFID      string   `json:"liffId"`
	View        LIFFView `json:"view"`
	Description string   `json:"description,omitempty"`
}

// LIFFAppsResponse represents the response from listing LIFF apps
type LIFFAppsResponse struct {
	Apps []LIFFApp `json:"apps"`
}

// AddLIFFAppRequest represents a request to add a new LIFF app
type AddLIFFAppRequest struct {
	View        LIFFView `json:"view"`
	Description string   `json:"description,omitempty"`
}

// AddLIFFAppResponse represents the response from adding a LIFF app
type AddLIFFAppResponse struct {
	LIFFID string `json:"liffId"`
}

// UpdateLIFFAppRequest represents a request to update a LIFF app
type UpdateLIFFAppRequest struct {
	View        LIFFView `json:"view"`
	Description string   `json:"description,omitempty"`
}

// GetAllLIFFApps retrieves all LIFF apps for the channel.
// GET https://api.line.me/liff/v1/apps
func (c *Client) GetAllLIFFApps(ctx context.Context) ([]LIFFApp, error) {
	data, err := c.Get(ctx, "/liff/v1/apps")
	if err != nil {
		return nil, err
	}

	var resp LIFFAppsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse LIFF apps response: %w", err)
	}

	return resp.Apps, nil
}

// AddLIFFApp creates a new LIFF app.
// POST https://api.line.me/liff/v1/apps
func (c *Client) AddLIFFApp(ctx context.Context, req *AddLIFFAppRequest) (string, error) {
	data, err := c.Post(ctx, "/liff/v1/apps", req)
	if err != nil {
		return "", err
	}

	var resp AddLIFFAppResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse add LIFF app response: %w", err)
	}

	return resp.LIFFID, nil
}

// UpdateLIFFApp updates an existing LIFF app.
// PUT https://api.line.me/liff/v1/apps/{liffId}
func (c *Client) UpdateLIFFApp(ctx context.Context, liffID string, req *UpdateLIFFAppRequest) error {
	_, err := c.Put(ctx, "/liff/v1/apps/"+liffID, req)
	return err
}

// DeleteLIFFApp deletes a LIFF app.
// DELETE https://api.line.me/liff/v1/apps/{liffId}
func (c *Client) DeleteLIFFApp(ctx context.Context, liffID string) error {
	_, err := c.Delete(ctx, "/liff/v1/apps/"+liffID)
	return err
}
