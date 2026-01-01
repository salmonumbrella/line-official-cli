package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type MembershipPlan struct {
	MembershipID int64    `json:"membershipId"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Benefits     []string `json:"benefits,omitempty"`
	Price        int64    `json:"price"`
	Currency     string   `json:"currency"`
	IsPublished  bool     `json:"isPublished"`
	IsInSale     bool     `json:"isInSale"`
}

type MembershipPlansResponse struct {
	Memberships []MembershipPlan `json:"memberships"`
}

type UserMembershipStatus struct {
	MembershipID      int64  `json:"membershipId"`
	SubscriptionState string `json:"subscriptionState"`
	StartTime         int64  `json:"startTime,omitempty"`
	EndTime           int64  `json:"endTime,omitempty"`
}

type UserMembershipResponse struct {
	Memberships []UserMembershipStatus `json:"memberships"`
}

type MembershipUsersResponse struct {
	MemberIDs []string `json:"memberIds"`
	Next      string   `json:"next,omitempty"`
}

func (c *Client) GetMembershipPlans(ctx context.Context) ([]MembershipPlan, error) {
	data, err := c.Get(ctx, "/v2/bot/membership/plans")
	if err != nil {
		return nil, err
	}
	var resp MembershipPlansResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Memberships, nil
}

func (c *Client) GetUserMembershipStatus(ctx context.Context, userID string) ([]UserMembershipStatus, error) {
	path := fmt.Sprintf("/v2/bot/users/%s/membership/subscription", userID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp UserMembershipResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Memberships, nil
}

func (c *Client) GetMembershipUsers(ctx context.Context, start string) (*MembershipUsersResponse, error) {
	path := "/v2/bot/membership/users"
	if start != "" {
		path += "?start=" + start
	}
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp MembershipUsersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}
