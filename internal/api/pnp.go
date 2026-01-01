package api

import "context"

// PNPPushRequest represents the request body for PNP push messages.
// PNP (Phone Number Push) sends messages to users by phone number instead of LINE user ID.
type PNPPushRequest struct {
	To       string `json:"to"`       // Phone number with country code (e.g., "+819012345678")
	Messages []any  `json:"messages"` // Same message format as regular push
}

// PNPPushMessage sends a text message to a user identified by phone number.
// POST /bot/pnp/push (note: different base path than /v2/bot/...)
// Requires PNP enabled channel.
func (c *Client) PNPPushMessage(ctx context.Context, phoneNumber, text string) error {
	req := PNPPushRequest{
		To:       phoneNumber,
		Messages: []any{TextMessage{Type: "text", Text: text}},
	}
	_, err := c.Post(ctx, "/bot/pnp/push", req)
	return err
}
