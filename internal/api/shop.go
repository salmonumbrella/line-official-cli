package api

import (
	"context"
)

// MissionStickerRequest represents a request to send a mission sticker
type MissionStickerRequest struct {
	To                 string `json:"to"`
	ProductID          string `json:"productId"`
	ProductType        string `json:"productType"`
	SendPresentMessage bool   `json:"sendPresentMessage,omitempty"`
}

// SendMissionSticker sends a mission sticker to a user
// POST /shop/v3/mission
func (c *Client) SendMissionSticker(ctx context.Context, userID, productID, productType string, sendMessage bool) error {
	req := MissionStickerRequest{
		To:                 userID,
		ProductID:          productID,
		ProductType:        productType,
		SendPresentMessage: sendMessage,
	}
	_, err := c.Post(ctx, "/shop/v3/mission", req)
	return err
}
