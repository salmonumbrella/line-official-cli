package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

// messageTarget specifies how to send a message (push/broadcast/multicast)
type messageTarget struct {
	Type    string   // "push", "broadcast", "multicast"
	UserID  string   // for push
	UserIDs []string // for multicast
}

// sendMessage is the generic message sending helper for the command layer.
// It handles client creation, API calls, and output formatting.
// If client is nil, a new client is created using newAPIClient().
func sendMessage(cmd *cobra.Command, client *api.Client, target messageTarget, message any, msgType string, extraFields map[string]any) error {
	if client == nil {
		var err error
		client, err = newAPIClient()
		if err != nil {
			return err
		}
	}

	if err := client.SendMessage(cmd.Context(), target.Type, target.UserID, target.UserIDs, message); err != nil {
		return fmt.Errorf("failed to send %s: %w", msgType, err)
	}

	return formatMessageOutput(cmd, target, msgType, extraFields)
}

// formatMessageOutput formats the output for a sent message.
func formatMessageOutput(cmd *cobra.Command, target messageTarget, msgType string, extraFields map[string]any) error {
	if flags.Output == "json" {
		result := map[string]any{"type": msgType}
		switch target.Type {
		case "push":
			result["status"] = "sent"
			result["to"] = target.UserID
		case "broadcast":
			result["status"] = "broadcast"
		case "multicast":
			result["status"] = "sent"
			result["recipients"] = len(target.UserIDs)
		}
		for k, v := range extraFields {
			result[k] = v
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	// Text output
	switch target.Type {
	case "push":
		if msgType == "text" {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Message sent to %s\n", target.UserID)
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s sent to %s\n", capitalize(msgType), target.UserID)
		}
	case "broadcast":
		if msgType == "text" {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Broadcast sent")
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s broadcast sent\n", capitalize(msgType))
		}
	case "multicast":
		if msgType == "text" {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Message sent to %d users\n", len(target.UserIDs))
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s sent to %d users\n", capitalize(msgType), len(target.UserIDs))
		}
	}
	return nil
}

// capitalize returns the string with first letter capitalized.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}

// dispatchMessage routes to the appropriate message type handler based on which flag is set.
// If client is nil, a new client is created using newAPIClient().
func dispatchMessage(cmd *cobra.Command, client *api.Client, target messageTarget, text, flexJSON, altText, imageURL, previewURL, videoURL, audioURL string, duration int, locationTitle, locationAddress string, lat, lng float64, packageID, stickerID string) error {
	if text != "" {
		msg := api.TextMessage{Type: "text", Text: text}
		return sendMessage(cmd, client, target, msg, "text", nil)
	}
	if flexJSON != "" {
		msg := api.FlexMessage{Type: "flex", AltText: altText, Contents: json.RawMessage(flexJSON)}
		return sendMessage(cmd, client, target, msg, "flex", nil)
	}
	if imageURL != "" {
		if previewURL == "" {
			previewURL = imageURL
		}
		msg := api.ImageMessage{Type: "image", OriginalContentURL: imageURL, PreviewImageURL: previewURL}
		return sendMessage(cmd, client, target, msg, "image", nil)
	}
	if videoURL != "" {
		if previewURL == "" {
			return fmt.Errorf("--preview is required for video messages")
		}
		msg := api.VideoMessage{Type: "video", OriginalContentURL: videoURL, PreviewImageURL: previewURL}
		return sendMessage(cmd, client, target, msg, "video", nil)
	}
	if audioURL != "" {
		if duration <= 0 {
			return fmt.Errorf("--duration is required for audio messages (in milliseconds)")
		}
		msg := api.AudioMessage{Type: "audio", OriginalContentURL: audioURL, Duration: duration}
		return sendMessage(cmd, client, target, msg, "audio", map[string]any{"duration": duration})
	}
	if locationTitle != "" || locationAddress != "" || lat != 0 || lng != 0 {
		if locationTitle == "" {
			return fmt.Errorf("--location-title is required for location messages")
		}
		if locationAddress == "" {
			return fmt.Errorf("--location-address is required for location messages")
		}
		if lat == 0 && lng == 0 {
			return fmt.Errorf("--lat and --lng are required for location messages")
		}
		msg := api.LocationMessage{Type: "location", Title: locationTitle, Address: locationAddress, Latitude: lat, Longitude: lng}
		return sendMessage(cmd, client, target, msg, "location", map[string]any{"title": locationTitle, "address": locationAddress, "lat": lat, "lng": lng})
	}
	// Must be sticker (validation already done in command)
	msg := api.StickerMessage{Type: "sticker", PackageID: packageID, StickerID: stickerID}
	return sendMessage(cmd, client, target, msg, "sticker", map[string]any{"packageId": packageID, "stickerId": stickerID})
}

func newMessageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "message",
		Aliases: []string{"msg"},
		Short:   "Send and manage messages",
		Long:    "Send text and flex messages to users or broadcast to all followers.",
	}

	cmd.AddCommand(newMessagePushCmd())
	cmd.AddCommand(newMessageBroadcastCmd())
	cmd.AddCommand(newMessageMulticastCmd())
	cmd.AddCommand(newMessageReplyCmd())
	cmd.AddCommand(newMessageQuotaCmd())
	cmd.AddCommand(newMessageNarrowcastCmd())
	cmd.AddCommand(newMessageNarrowcastStatusCmd())
	cmd.AddCommand(newMessageDeliveryStatsCmd())
	cmd.AddCommand(newMessageValidateCmd())
	cmd.AddCommand(newMessageAggregationCmd())

	return cmd
}
