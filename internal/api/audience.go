package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/salmonumbrella/line-official-cli/internal/api/generated"
)

// GetAudienceGroups returns a list of audience groups
func (c *Client) GetAudienceGroups(ctx context.Context) ([]generated.AudienceGroup, error) {
	data, err := c.Get(ctx, "/v2/bot/audienceGroup/list?page=1&size=40")
	if err != nil {
		return nil, err
	}
	var resp generated.GetAudienceGroupsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse audience groups: %w", err)
	}
	if resp.AudienceGroups == nil {
		return []generated.AudienceGroup{}, nil
	}
	return *resp.AudienceGroups, nil
}

// GetAudienceGroup returns a single audience group by ID
func (c *Client) GetAudienceGroup(ctx context.Context, audienceGroupID int64) (*generated.GetAudienceDataResponse, error) {
	path := fmt.Sprintf("/v2/bot/audienceGroup/%d", audienceGroupID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp generated.GetAudienceDataResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse audience group: %w", err)
	}
	return &resp, nil
}

// DeleteAudienceGroup deletes an audience group
func (c *Client) DeleteAudienceGroup(ctx context.Context, audienceGroupID int64) error {
	path := fmt.Sprintf("/v2/bot/audienceGroup/%d", audienceGroupID)
	_, err := c.Delete(ctx, path)
	return err
}

// CreateAudienceRequest represents a request to create an audience group
type CreateAudienceRequest struct {
	Description       string   `json:"description"`
	IsIfaAudience     bool     `json:"isIfaAudience,omitempty"`
	Audiences         []UserID `json:"audiences,omitempty"`
	UploadDescription string   `json:"uploadDescription,omitempty"`
}

// UserID represents a user ID for audience creation
type UserID struct {
	ID string `json:"id"`
}

// CreateAudienceResponse represents the response from creating an audience group
type CreateAudienceResponse struct {
	AudienceGroupID int64  `json:"audienceGroupId"`
	Type            string `json:"type"`
	Description     string `json:"description"`
	Created         int64  `json:"created"`
}

// CreateAudienceGroup creates a new audience group from user IDs
func (c *Client) CreateAudienceGroup(ctx context.Context, description string, userIDs []string) (*CreateAudienceResponse, error) {
	audiences := make([]UserID, len(userIDs))
	for i, id := range userIDs {
		audiences[i] = UserID{ID: id}
	}

	req := CreateAudienceRequest{
		Description: description,
		Audiences:   audiences,
	}

	data, err := c.Post(ctx, "/v2/bot/audienceGroup/upload", req)
	if err != nil {
		return nil, err
	}

	var resp CreateAudienceResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// AddUsersToAudienceRequest represents a request to add users to an existing audience
type AddUsersToAudienceRequest struct {
	AudienceGroupID   int64    `json:"audienceGroupId"`
	Audiences         []UserID `json:"audiences"`
	UploadDescription string   `json:"uploadDescription,omitempty"`
}

// AddUsersToAudience adds users to an existing audience group
// PUT /v2/bot/audienceGroup/upload
func (c *Client) AddUsersToAudience(ctx context.Context, audienceGroupID int64, userIDs []string, description string) error {
	audiences := make([]UserID, len(userIDs))
	for i, id := range userIDs {
		audiences[i] = UserID{ID: id}
	}

	req := AddUsersToAudienceRequest{
		AudienceGroupID:   audienceGroupID,
		Audiences:         audiences,
		UploadDescription: description,
	}

	_, err := c.Put(ctx, "/v2/bot/audienceGroup/upload", req)
	return err
}

// CreateClickBasedAudienceRequest represents a request to create a click-based audience
type CreateClickBasedAudienceRequest struct {
	Description string `json:"description"`
	RequestID   string `json:"requestId"`
}

// CreateClickBasedAudience creates an audience from users who clicked a message
// POST /v2/bot/audienceGroup/click
func (c *Client) CreateClickBasedAudience(ctx context.Context, description string, requestID string) (*CreateAudienceResponse, error) {
	req := CreateClickBasedAudienceRequest{
		Description: description,
		RequestID:   requestID,
	}

	data, err := c.Post(ctx, "/v2/bot/audienceGroup/click", req)
	if err != nil {
		return nil, err
	}

	var resp CreateAudienceResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// CreateImpressionBasedAudienceRequest represents a request to create an impression-based audience
type CreateImpressionBasedAudienceRequest struct {
	Description string `json:"description"`
	RequestID   string `json:"requestId"`
}

// CreateImpressionBasedAudience creates an audience from users who saw a message
// POST /v2/bot/audienceGroup/imp
func (c *Client) CreateImpressionBasedAudience(ctx context.Context, description string, requestID string) (*CreateAudienceResponse, error) {
	req := CreateImpressionBasedAudienceRequest{
		Description: description,
		RequestID:   requestID,
	}

	data, err := c.Post(ctx, "/v2/bot/audienceGroup/imp", req)
	if err != nil {
		return nil, err
	}

	var resp CreateAudienceResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// UpdateDescriptionRequest represents a request to update an audience description
type UpdateDescriptionRequest struct {
	Description string `json:"description"`
}

// UpdateAudienceDescription updates the description of an audience group
// PUT /v2/bot/audienceGroup/{audienceGroupId}/updateDescription
func (c *Client) UpdateAudienceDescription(ctx context.Context, audienceGroupID int64, description string) error {
	path := fmt.Sprintf("/v2/bot/audienceGroup/%d/updateDescription", audienceGroupID)
	req := UpdateDescriptionRequest{
		Description: description,
	}
	_, err := c.Put(ctx, path, req)
	return err
}

// GetSharedAudienceGroups lists shared audience groups
// GET /v2/bot/audienceGroup/shared/list
func (c *Client) GetSharedAudienceGroups(ctx context.Context) ([]generated.AudienceGroup, error) {
	data, err := c.Get(ctx, "/v2/bot/audienceGroup/shared/list?page=1&size=40")
	if err != nil {
		return nil, err
	}
	var resp generated.GetSharedAudienceGroupsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse shared audience groups: %w", err)
	}
	if resp.AudienceGroups == nil {
		return []generated.AudienceGroup{}, nil
	}
	return *resp.AudienceGroups, nil
}

// GetSharedAudienceGroup gets a shared audience group by ID
// GET /v2/bot/audienceGroup/shared/{audienceGroupId}
func (c *Client) GetSharedAudienceGroup(ctx context.Context, audienceGroupID int64) (*generated.GetSharedAudienceDataResponse, error) {
	path := fmt.Sprintf("/v2/bot/audienceGroup/shared/%d", audienceGroupID)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp generated.GetSharedAudienceDataResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse shared audience group: %w", err)
	}
	return &resp, nil
}

// CreateAudienceFromFile creates an audience by uploading a file of user IDs.
// The file should contain one user ID per line.
// POST /v2/bot/audienceGroup/upload/byFile
func (c *Client) CreateAudienceFromFile(ctx context.Context, description string, filePath string) (*CreateAudienceResponse, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Validate file content - ensure it has user IDs
	lines := strings.Split(string(fileContent), "\n")
	var validLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			validLines = append(validLines, line)
		}
	}
	if len(validLines) == 0 {
		return nil, fmt.Errorf("file contains no user IDs")
	}

	// Rejoin with newlines for upload
	uploadContent := []byte(strings.Join(validLines, "\n"))

	formFields := map[string]string{
		"description": description,
	}

	fileName := filepath.Base(filePath)
	data, err := c.PostMultipart(ctx, "/v2/bot/audienceGroup/upload/byFile", "file", fileName, uploadContent, formFields)
	if err != nil {
		return nil, err
	}

	var resp CreateAudienceResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// AddUsersToAudienceFromFile adds users from a file to an existing audience.
// The file should contain one user ID per line.
// PUT /v2/bot/audienceGroup/upload/byFile
func (c *Client) AddUsersToAudienceFromFile(ctx context.Context, audienceGroupID int64, filePath string, uploadDescription string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Validate file content - ensure it has user IDs
	lines := strings.Split(string(fileContent), "\n")
	var validLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			validLines = append(validLines, line)
		}
	}
	if len(validLines) == 0 {
		return fmt.Errorf("file contains no user IDs")
	}

	// Rejoin with newlines for upload
	uploadContent := []byte(strings.Join(validLines, "\n"))

	formFields := map[string]string{
		"audienceGroupId": fmt.Sprintf("%d", audienceGroupID),
	}
	if uploadDescription != "" {
		formFields["uploadDescription"] = uploadDescription
	}

	fileName := filepath.Base(filePath)
	_, err = c.PutMultipart(ctx, "/v2/bot/audienceGroup/upload/byFile", "file", fileName, uploadContent, formFields)
	return err
}
