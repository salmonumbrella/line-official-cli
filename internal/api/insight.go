package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api/generated"
)

// GetFollowerStats returns follower statistics for a given date
// date format: "20250101" (YYYYMMDD)
func (c *Client) GetFollowerStats(ctx context.Context, date string) (*generated.GetNumberOfFollowersResponse, error) {
	path := fmt.Sprintf("/v2/bot/insight/followers?date=%s", date)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp generated.GetNumberOfFollowersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse follower stats: %w", err)
	}
	return &resp, nil
}

// GetMessageDeliveryStats returns message delivery statistics for a given date
// date format: "20250101" (YYYYMMDD)
func (c *Client) GetMessageDeliveryStats(ctx context.Context, date string) (*generated.GetNumberOfMessageDeliveriesResponse, error) {
	path := fmt.Sprintf("/v2/bot/insight/message/delivery?date=%s", date)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp generated.GetNumberOfMessageDeliveriesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse delivery stats: %w", err)
	}
	return &resp, nil
}

// GetFriendsDemographics returns demographic information about friends
func (c *Client) GetFriendsDemographics(ctx context.Context) (*generated.GetFriendsDemographicsResponse, error) {
	data, err := c.Get(ctx, "/v2/bot/insight/demographic")
	if err != nil {
		return nil, err
	}
	var resp generated.GetFriendsDemographicsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse demographics: %w", err)
	}
	return &resp, nil
}

// MessageEventResponse contains message event statistics
type MessageEventResponse struct {
	Overview *MessageEventOverview `json:"overview,omitempty"`
	Messages []MessageEventMessage `json:"messages,omitempty"`
	Clicks   []MessageEventClick   `json:"clicks,omitempty"`
}

// MessageEventOverview contains overview metrics for a message
type MessageEventOverview struct {
	RequestID         string `json:"requestId"`
	Timestamp         int64  `json:"timestamp"`
	Delivered         int64  `json:"delivered"`
	UniqueImpression  int64  `json:"uniqueImpression,omitempty"`
	UniqueClick       int64  `json:"uniqueClick,omitempty"`
	UniqueMediaPlayed int64  `json:"uniqueMediaPlayed,omitempty"`
}

// MessageEventMessage contains per-message metrics
type MessageEventMessage struct {
	Seq                 int   `json:"seq"`
	Impression          int64 `json:"impression"`
	MediaPlayed         int64 `json:"mediaPlayed,omitempty"`
	MediaPlayedComplete int64 `json:"mediaPlayedComplete,omitempty"`
}

// MessageEventClick contains click metrics for a message
type MessageEventClick struct {
	Seq         int    `json:"seq"`
	URL         string `json:"url,omitempty"`
	Click       int64  `json:"click"`
	UniqueClick int64  `json:"uniqueClick"`
}

// GetMessageEventStats returns event statistics for a specific message request
func (c *Client) GetMessageEventStats(ctx context.Context, requestID string) (*MessageEventResponse, error) {
	path := fmt.Sprintf("/v2/bot/insight/message/event?requestId=%s", requestID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp MessageEventResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse event stats: %w", err)
	}
	return &resp, nil
}

// StatisticsPerUnitResponse contains aggregated statistics by custom aggregation unit
type StatisticsPerUnitResponse struct {
	Overview *StatisticsOverview   `json:"overview,omitempty"`
	Messages []MessageEventMessage `json:"messages,omitempty"`
	Clicks   []ClickStatistics     `json:"clicks,omitempty"`
}

// StatisticsOverview contains overview metrics for aggregation unit
type StatisticsOverview struct {
	UniqueImpression          int64 `json:"uniqueImpression,omitempty"`
	UniqueClick               int64 `json:"uniqueClick,omitempty"`
	UniqueMediaPlayed         int64 `json:"uniqueMediaPlayed,omitempty"`
	UniqueMediaPlayedComplete int64 `json:"uniqueMediaPlayed100Percent,omitempty"`
}

// ClickStatistics contains click metrics for an aggregation unit
type ClickStatistics struct {
	Seq                  int    `json:"seq"`
	URL                  string `json:"url,omitempty"`
	Click                int64  `json:"click"`
	UniqueClick          int64  `json:"uniqueClick"`
	UniqueClickOfRequest int64  `json:"uniqueClickOfRequest,omitempty"`
}

// GetStatisticsPerUnit returns event statistics aggregated by custom unit
// GET /v2/bot/insight/message/event/aggregation?customAggregationUnit=<unit>&from=YYYYMMDD&to=YYYYMMDD
func (c *Client) GetStatisticsPerUnit(ctx context.Context, unit, from, to string) (*StatisticsPerUnitResponse, error) {
	path := fmt.Sprintf("/v2/bot/insight/message/event/aggregation?customAggregationUnit=%s&from=%s&to=%s", unit, from, to)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp StatisticsPerUnitResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse statistics per unit: %w", err)
	}
	return &resp, nil
}
