package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type GroupSummary struct {
	GroupID    string `json:"groupId"`
	GroupName  string `json:"groupName"`
	PictureURL string `json:"pictureUrl,omitempty"`
}

type GroupMemberCount struct {
	Count int `json:"count"`
}

type GroupMemberIDs struct {
	MemberIDs []string `json:"memberIds"`
	Next      string   `json:"next,omitempty"`
}

func (c *Client) GetGroupSummary(ctx context.Context, groupID string) (*GroupSummary, error) {
	path := fmt.Sprintf("/v2/bot/group/%s/summary", groupID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var summary GroupSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &summary, nil
}

func (c *Client) GetGroupMemberCount(ctx context.Context, groupID string) (int, error) {
	path := fmt.Sprintf("/v2/bot/group/%s/members/count", groupID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return 0, err
	}
	var resp GroupMemberCount
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Count, nil
}

func (c *Client) GetGroupMemberIDs(ctx context.Context, groupID, start string) (*GroupMemberIDs, error) {
	path := fmt.Sprintf("/v2/bot/group/%s/members/ids", groupID)
	if start != "" {
		path += "?start=" + start
	}
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp GroupMemberIDs
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetGroupMemberProfile(ctx context.Context, groupID, userID string) (*UserProfile, error) {
	path := fmt.Sprintf("/v2/bot/group/%s/member/%s", groupID, userID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &profile, nil
}

func (c *Client) LeaveGroup(ctx context.Context, groupID string) error {
	path := fmt.Sprintf("/v2/bot/group/%s/leave", groupID)
	_, err := c.Post(ctx, path, nil)
	return err
}
