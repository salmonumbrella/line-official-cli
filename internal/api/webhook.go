package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type WebhookEndpointInfo struct {
	Endpoint string `json:"endpoint"`
	Active   bool   `json:"active"`
}

type SetWebhookEndpointRequest struct {
	Endpoint string `json:"endpoint"`
}

type TestWebhookResponse struct {
	Success    bool   `json:"success"`
	Timestamp  string `json:"timestamp"`
	StatusCode int    `json:"statusCode"`
	Reason     string `json:"reason"`
	Detail     string `json:"detail"`
}

func (c *Client) GetWebhookEndpoint(ctx context.Context) (*WebhookEndpointInfo, error) {
	data, err := c.Get(ctx, "/v2/bot/channel/webhook/endpoint")
	if err != nil {
		return nil, err
	}
	var info WebhookEndpointInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &info, nil
}

func (c *Client) SetWebhookEndpoint(ctx context.Context, endpoint string) error {
	req := SetWebhookEndpointRequest{Endpoint: endpoint}
	_, err := c.Put(ctx, "/v2/bot/channel/webhook/endpoint", req)
	return err
}

func (c *Client) TestWebhookEndpoint(ctx context.Context, endpoint string) (*TestWebhookResponse, error) {
	var req map[string]string
	if endpoint != "" {
		req = map[string]string{"endpoint": endpoint}
	}
	data, err := c.Post(ctx, "/v2/bot/channel/webhook/test", req)
	if err != nil {
		return nil, err
	}
	var resp TestWebhookResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}
