package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Coupon represents a LINE coupon
type Coupon struct {
	CouponID             string                `json:"couponId"`
	Title                string                `json:"title"`
	Description          string                `json:"description,omitempty"`
	Status               string                `json:"status,omitempty"` // DRAFT, RUNNING, CLOSED
	ImageURL             string                `json:"imageUrl,omitempty"`
	StartTimestamp       int64                 `json:"startTimestamp,omitempty"`
	EndTimestamp         int64                 `json:"endTimestamp,omitempty"`
	Timezone             string                `json:"timezone,omitempty"`
	Visibility           string                `json:"visibility,omitempty"` // UNLISTED
	MaxUseCountPerTicket int                   `json:"maxUseCountPerTicket,omitempty"`
	MaxTicketPerUser     int                   `json:"maxTicketPerUser,omitempty"`
	CreatedTimestamp     int64                 `json:"createdTimestamp,omitempty"`
	Reward               *CouponReward         `json:"reward,omitempty"`
	AcquisitionCondition *AcquisitionCondition `json:"acquisitionCondition,omitempty"`
}

// CouponReward represents the reward configuration for a coupon
type CouponReward struct {
	Type      string           `json:"type"` // "discount", "cashback", etc.
	PriceInfo *CouponPriceInfo `json:"priceInfo,omitempty"`
}

// CouponPriceInfo represents pricing information for a coupon reward
type CouponPriceInfo struct {
	Type        string `json:"type"`                  // "fixed", "percentage"
	FixedAmount int    `json:"fixedAmount,omitempty"` // amount in smallest currency unit
	Rate        int    `json:"rate,omitempty"`        // 1-99 for percentage
}

// AcquisitionCondition represents the condition for acquiring a coupon
type AcquisitionCondition struct {
	Type               string `json:"type"`                         // "normal", "lottery"
	LotteryProbability int    `json:"lotteryProbability,omitempty"` // 1-100
}

// CreateCouponRequest represents the request to create a new coupon
type CreateCouponRequest struct {
	Title                string                `json:"title"`
	Description          string                `json:"description,omitempty"`
	ImageURL             string                `json:"imageUrl,omitempty"`
	StartTimestamp       int64                 `json:"startTimestamp,omitempty"`
	EndTimestamp         int64                 `json:"endTimestamp"`
	Timezone             string                `json:"timezone,omitempty"`
	Visibility           string                `json:"visibility,omitempty"`
	MaxUseCountPerTicket int                   `json:"maxUseCountPerTicket,omitempty"`
	MaxTicketPerUser     int                   `json:"maxTicketPerUser,omitempty"`
	Reward               *CouponReward         `json:"reward,omitempty"`
	AcquisitionCondition *AcquisitionCondition `json:"acquisitionCondition,omitempty"`
}

// CouponListResponse represents the response from listing coupons
type CouponListResponse struct {
	Coupons []Coupon `json:"items"`
	Next    string   `json:"next,omitempty"`
}

// createCouponResponse represents the response when creating a coupon
type createCouponResponse struct {
	CouponID string `json:"couponId"`
}

// ListCoupons gets all coupons with optional status filter
// GET /v2/bot/coupon?status=RUNNING&limit=20&start=...
func (c *Client) ListCoupons(ctx context.Context, status []string, limit int, start string) (*CouponListResponse, error) {
	params := url.Values{}
	if len(status) > 0 {
		params.Set("status", strings.Join(status, ","))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if start != "" {
		params.Set("start", start)
	}

	path := "/v2/bot/coupon"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp CouponListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse coupons: %w", err)
	}
	return &resp, nil
}

// CreateCoupon creates a new coupon
// POST /v2/bot/coupon
// Returns couponId
func (c *Client) CreateCoupon(ctx context.Context, req *CreateCouponRequest) (string, error) {
	data, err := c.Post(ctx, "/v2/bot/coupon", req)
	if err != nil {
		return "", err
	}

	var resp createCouponResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.CouponID, nil
}

// GetCoupon gets details of a specific coupon
// GET /v2/bot/coupon/{couponId}
func (c *Client) GetCoupon(ctx context.Context, couponID string) (*Coupon, error) {
	data, err := c.Get(ctx, "/v2/bot/coupon/"+couponID)
	if err != nil {
		return nil, err
	}

	var coupon Coupon
	if err := json.Unmarshal(data, &coupon); err != nil {
		return nil, fmt.Errorf("failed to parse coupon: %w", err)
	}
	return &coupon, nil
}

// CloseCoupon discontinues a coupon
// PUT /v2/bot/coupon/{couponId}/close
func (c *Client) CloseCoupon(ctx context.Context, couponID string) error {
	_, err := c.Put(ctx, "/v2/bot/coupon/"+couponID+"/close", nil)
	return err
}
