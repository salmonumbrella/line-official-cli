package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type TextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type FlexMessage struct {
	Type     string          `json:"type"`
	AltText  string          `json:"altText"`
	Contents json.RawMessage `json:"contents"`
}

type ImageMessage struct {
	Type               string `json:"type"`
	OriginalContentURL string `json:"originalContentUrl"`
	PreviewImageURL    string `json:"previewImageUrl"`
}

type StickerMessage struct {
	Type      string `json:"type"`
	PackageID string `json:"packageId"`
	StickerID string `json:"stickerId"`
}

type VideoMessage struct {
	Type               string `json:"type"`
	OriginalContentURL string `json:"originalContentUrl"`
	PreviewImageURL    string `json:"previewImageUrl"`
	TrackingID         string `json:"trackingId,omitempty"`
}

type AudioMessage struct {
	Type               string `json:"type"`
	OriginalContentURL string `json:"originalContentUrl"`
	Duration           int    `json:"duration"`
}

type LocationMessage struct {
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type PushMessageRequest struct {
	To       string `json:"to"`
	Messages []any  `json:"messages"`
}

type BroadcastMessageRequest struct {
	Messages []any `json:"messages"`
}

type MulticastMessageRequest struct {
	To       []string `json:"to"`
	Messages []any    `json:"messages"`
}

type ReplyMessageRequest struct {
	ReplyToken string `json:"replyToken"`
	Messages   []any  `json:"messages"`
}

type QuotaResponse struct {
	Type  string `json:"type"`
	Value int    `json:"value,omitempty"`
}

type ConsumptionResponse struct {
	TotalUsage int `json:"totalUsage"`
}

type DeliveryStatsResponse struct {
	Status  string `json:"status"`
	Success int64  `json:"success"`
}

// DeliveryStats represents detailed delivery statistics for a message type
type DeliveryStats struct {
	Status       string `json:"status"`
	Success      int64  `json:"success,omitempty"`
	Failure      int64  `json:"failure,omitempty"`
	RequestCount int64  `json:"requestCount,omitempty"`
}

func (c *Client) ReplyTextMessage(ctx context.Context, replyToken, text string) error {
	req := ReplyMessageRequest{
		ReplyToken: replyToken,
		Messages:   []any{TextMessage{Type: "text", Text: text}},
	}
	_, err := c.Post(ctx, "/v2/bot/message/reply", req)
	return err
}

func (c *Client) ReplyFlexMessage(ctx context.Context, replyToken, altText string, contents json.RawMessage) error {
	req := ReplyMessageRequest{
		ReplyToken: replyToken,
		Messages:   []any{FlexMessage{Type: "flex", AltText: altText, Contents: contents}},
	}
	_, err := c.Post(ctx, "/v2/bot/message/reply", req)
	return err
}

func (c *Client) GetMessageQuota(ctx context.Context) (*QuotaResponse, error) {
	data, err := c.Get(ctx, "/v2/bot/message/quota")
	if err != nil {
		return nil, err
	}
	var resp QuotaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse quota: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetMessageConsumption(ctx context.Context) (*ConsumptionResponse, error) {
	data, err := c.Get(ctx, "/v2/bot/message/quota/consumption")
	if err != nil {
		return nil, err
	}
	var resp ConsumptionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse consumption: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetDeliveryStats(ctx context.Context, messageType, date string) (*DeliveryStatsResponse, error) {
	path := fmt.Sprintf("/v2/bot/message/delivery/%s?date=%s", messageType, date)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp DeliveryStatsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// getDeliveryStatsForType fetches delivery statistics for a specific message type
func (c *Client) getDeliveryStatsForType(ctx context.Context, messageType, date string) (*DeliveryStats, error) {
	path := fmt.Sprintf("/v2/bot/message/delivery/%s?date=%s", messageType, date)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp DeliveryStats
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse delivery stats: %w", err)
	}
	return &resp, nil
}

// GetReplyMessageStats retrieves delivery statistics for reply messages
// GET /v2/bot/message/delivery/reply?date=YYYYMMDD
func (c *Client) GetReplyMessageStats(ctx context.Context, date string) (*DeliveryStats, error) {
	return c.getDeliveryStatsForType(ctx, "reply", date)
}

// GetPushMessageStats retrieves delivery statistics for push messages
// GET /v2/bot/message/delivery/push?date=YYYYMMDD
func (c *Client) GetPushMessageStats(ctx context.Context, date string) (*DeliveryStats, error) {
	return c.getDeliveryStatsForType(ctx, "push", date)
}

// GetMulticastMessageStats retrieves delivery statistics for multicast messages
// GET /v2/bot/message/delivery/multicast?date=YYYYMMDD
func (c *Client) GetMulticastMessageStats(ctx context.Context, date string) (*DeliveryStats, error) {
	return c.getDeliveryStatsForType(ctx, "multicast", date)
}

// GetBroadcastMessageStats retrieves delivery statistics for broadcast messages
// GET /v2/bot/message/delivery/broadcast?date=YYYYMMDD
func (c *Client) GetBroadcastMessageStats(ctx context.Context, date string) (*DeliveryStats, error) {
	return c.getDeliveryStatsForType(ctx, "broadcast", date)
}

// GetPNPMessageStats retrieves delivery statistics for PNP (push notification push) messages
// GET /v2/bot/message/delivery/pnp?date=YYYYMMDD
func (c *Client) GetPNPMessageStats(ctx context.Context, date string) (*DeliveryStats, error) {
	return c.getDeliveryStatsForType(ctx, "pnp", date)
}

type ValidateMessageRequest struct {
	Messages []json.RawMessage `json:"messages"`
}

func (c *Client) ValidateMessage(ctx context.Context, messageType string, messages []json.RawMessage) error {
	path := fmt.Sprintf("/v2/bot/message/validate/%s", messageType)
	req := ValidateMessageRequest{Messages: messages}
	_, err := c.Post(ctx, path, req)
	return err
}

// ValidateReplyMessage validates message objects for reply endpoint
// POST /v2/bot/message/validate/reply
func (c *Client) ValidateReplyMessage(ctx context.Context, messages []json.RawMessage) error {
	return c.ValidateMessage(ctx, "reply", messages)
}

// ValidatePushMessage validates message objects for push endpoint
// POST /v2/bot/message/validate/push
func (c *Client) ValidatePushMessage(ctx context.Context, messages []json.RawMessage) error {
	return c.ValidateMessage(ctx, "push", messages)
}

// ValidateMulticastMessage validates message objects for multicast endpoint
// POST /v2/bot/message/validate/multicast
func (c *Client) ValidateMulticastMessage(ctx context.Context, messages []json.RawMessage) error {
	return c.ValidateMessage(ctx, "multicast", messages)
}

// ValidateNarrowcastMessage validates message objects for narrowcast endpoint
// POST /v2/bot/message/validate/narrowcast
func (c *Client) ValidateNarrowcastMessage(ctx context.Context, messages []json.RawMessage) error {
	return c.ValidateMessage(ctx, "narrowcast", messages)
}

// ValidateBroadcastMessage validates message objects for broadcast endpoint
// POST /v2/bot/message/validate/broadcast
func (c *Client) ValidateBroadcastMessage(ctx context.Context, messages []json.RawMessage) error {
	return c.ValidateMessage(ctx, "broadcast", messages)
}

type NarrowcastMessageRequest struct {
	Messages  []any                `json:"messages"`
	Recipient *NarrowcastRecipient `json:"recipient,omitempty"`
	Filter    *NarrowcastFilter    `json:"filter,omitempty"`
	Limit     *NarrowcastLimit     `json:"limit,omitempty"`
}

type NarrowcastRecipient struct {
	Type            string `json:"type"`
	AudienceGroupID int64  `json:"audienceGroupId,omitempty"`
}

type NarrowcastFilter struct {
	Demographic *DemographicFilter `json:"demographic,omitempty"`
}

type DemographicFilter struct {
	Type      string   `json:"type,omitempty"`
	OneOf     []any    `json:"oneOf,omitempty"`
	Ages      []string `json:"ages,omitempty"`
	Genders   []string `json:"genders,omitempty"`
	AppTypes  []string `json:"appTypes,omitempty"`
	AreaCodes []string `json:"areaCodes,omitempty"`
}

type NarrowcastLimit struct {
	Max                int  `json:"max,omitempty"`
	UpToRemainingQuota bool `json:"upToRemainingQuota,omitempty"`
}

type NarrowcastResponse struct {
	RequestID string `json:"requestId"`
}

func (c *Client) NarrowcastTextMessage(ctx context.Context, text string, audienceGroupID int64) (*NarrowcastResponse, error) {
	req := NarrowcastMessageRequest{
		Messages: []any{TextMessage{Type: "text", Text: text}},
	}
	if audienceGroupID > 0 {
		req.Recipient = &NarrowcastRecipient{
			Type:            "audience",
			AudienceGroupID: audienceGroupID,
		}
	}
	resp, err := c.PostWithHeaders(ctx, "/v2/bot/message/narrowcast", req)
	if err != nil {
		return nil, err
	}
	// LINE API returns request ID in X-Line-Request-Id header, not in response body
	requestID := resp.Headers.Get("X-Line-Request-Id")
	return &NarrowcastResponse{RequestID: requestID}, nil
}

func (c *Client) GetNarrowcastProgress(ctx context.Context, requestID string) (map[string]any, error) {
	path := fmt.Sprintf("/v2/bot/message/progress/narrowcast?requestId=%s", requestID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp map[string]any
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp, nil
}

// Aggregation unit types

type AggregationUsage struct {
	NumOfCustomAggregationUnits int64 `json:"numOfCustomAggregationUnits"`
}

type AggregationUnitListResponse struct {
	CustomAggregationUnits []string `json:"customAggregationUnits"`
	Next                   string   `json:"next,omitempty"`
}

// GetAggregationUnitUsage gets the number of aggregation units used this month
// GET /v2/bot/message/aggregation/info
func (c *Client) GetAggregationUnitUsage(ctx context.Context) (*AggregationUsage, error) {
	data, err := c.Get(ctx, "/v2/bot/message/aggregation/info")
	if err != nil {
		return nil, err
	}
	var resp AggregationUsage
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse aggregation usage: %w", err)
	}
	return &resp, nil
}

// GetAggregationUnitNameList gets the list of aggregation unit names
// GET /v2/bot/message/aggregation/list?limit=N&start=xxx
func (c *Client) GetAggregationUnitNameList(ctx context.Context, limit int, start string) (*AggregationUnitListResponse, error) {
	path := "/v2/bot/message/aggregation/list"
	params := []string{}
	if limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}
	if start != "" {
		params = append(params, fmt.Sprintf("start=%s", start))
	}
	if len(params) > 0 {
		path += "?"
		for i, p := range params {
			if i > 0 {
				path += "&"
			}
			path += p
		}
	}

	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp AggregationUnitListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse aggregation unit list: %w", err)
	}
	return &resp, nil
}

// LoadingAnimationRequest represents the request body for showing loading animation
type LoadingAnimationRequest struct {
	ChatID         string `json:"chatId"`
	LoadingSeconds int    `json:"loadingSeconds,omitempty"`
}

// ShowLoadingAnimation displays a loading animation in the chat
// POST /v2/bot/chat/loading/start
func (c *Client) ShowLoadingAnimation(ctx context.Context, chatID string, loadingSeconds int) error {
	req := LoadingAnimationRequest{
		ChatID:         chatID,
		LoadingSeconds: loadingSeconds,
	}
	_, err := c.Post(ctx, "/v2/bot/chat/loading/start", req)
	return err
}

// MarkAsReadRequest represents the request body for marking messages as read
type MarkAsReadRequest struct {
	Chat MarkAsReadChat `json:"chat"`
}

// MarkAsReadChat contains the user ID for mark as read request
type MarkAsReadChat struct {
	UserID string `json:"userId"`
}

// MarkMessagesAsRead marks all messages from a user as read
// POST /v2/bot/message/markAsRead
func (c *Client) MarkMessagesAsRead(ctx context.Context, userID string) error {
	req := MarkAsReadRequest{
		Chat: MarkAsReadChat{
			UserID: userID,
		},
	}
	_, err := c.Post(ctx, "/v2/bot/message/markAsRead", req)
	return err
}

// MarkAsReadByTokenRequest represents the request body for marking messages as read by chat token
type MarkAsReadByTokenRequest struct {
	Chat MarkAsReadByTokenChat `json:"chat"`
}

// MarkAsReadByTokenChat contains the chat token for mark as read request
type MarkAsReadByTokenChat struct {
	Token string `json:"token"`
}

// MarkMessagesAsReadByToken marks messages as read using a chat token from a webhook event
// POST /v2/bot/chat/markAsRead
func (c *Client) MarkMessagesAsReadByToken(ctx context.Context, chatToken string) error {
	req := MarkAsReadByTokenRequest{
		Chat: MarkAsReadByTokenChat{
			Token: chatToken,
		},
	}
	_, err := c.Post(ctx, "/v2/bot/chat/markAsRead", req)
	return err
}

// SendMessage sends a message using the specified target type.
// targetType must be "push", "broadcast", or "multicast".
// For "push", userID must be set. For "multicast", userIDs must be set.
func (c *Client) SendMessage(ctx context.Context, targetType string, userID string, userIDs []string, message any) error {
	switch targetType {
	case "push":
		req := PushMessageRequest{
			To:       userID,
			Messages: []any{message},
		}
		_, err := c.Post(ctx, "/v2/bot/message/push", req)
		return err
	case "broadcast":
		req := BroadcastMessageRequest{
			Messages: []any{message},
		}
		_, err := c.Post(ctx, "/v2/bot/message/broadcast", req)
		return err
	case "multicast":
		req := MulticastMessageRequest{
			To:       userIDs,
			Messages: []any{message},
		}
		_, err := c.Post(ctx, "/v2/bot/message/multicast", req)
		return err
	default:
		return fmt.Errorf("unsupported target type: %s", targetType)
	}
}
