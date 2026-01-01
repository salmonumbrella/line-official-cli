package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type RoomMemberCount struct {
	Count int `json:"count"`
}

type RoomMemberIDs struct {
	MemberIDs []string `json:"memberIds"`
	Next      string   `json:"next,omitempty"`
}

func (c *Client) GetRoomMemberCount(ctx context.Context, roomID string) (int, error) {
	path := fmt.Sprintf("/v2/bot/room/%s/members/count", roomID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return 0, err
	}
	var resp RoomMemberCount
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Count, nil
}

func (c *Client) GetRoomMemberIDs(ctx context.Context, roomID, start string) (*RoomMemberIDs, error) {
	path := fmt.Sprintf("/v2/bot/room/%s/members/ids", roomID)
	if start != "" {
		path += "?start=" + start
	}
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp RoomMemberIDs
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetRoomMemberProfile(ctx context.Context, roomID, userID string) (*UserProfile, error) {
	path := fmt.Sprintf("/v2/bot/room/%s/member/%s", roomID, userID)
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

func (c *Client) LeaveRoom(ctx context.Context, roomID string) error {
	path := fmt.Sprintf("/v2/bot/room/%s/leave", roomID)
	_, err := c.Post(ctx, path, nil)
	return err
}
