package cmd

import (
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

// newMessagePushCmd creates a push message command.
func newMessagePushCmd() *cobra.Command {
	return newMessagePushCmdWithClient(nil)
}

// newMessagePushCmdWithClient creates a push message command with an optional API client for testing.
func newMessagePushCmdWithClient(client *api.Client) *cobra.Command {
	var userID string
	var text string
	var flexJSON string
	var altText string
	var imageURL string
	var previewURL string
	var packageID string
	var stickerID string
	var videoURL string
	var audioURL string
	var duration int
	var locationTitle string
	var locationAddress string
	var lat float64
	var lng float64

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push a message to a user",
		Long:  "Send a text, flex, image, video, audio, location, or sticker message directly to a specific user.",
		Example: `  # Send a text message
  line message push --to U1234567890abcdef --text "Hello!"

  # Send a flex message from JSON
  line message push --to U1234567890abcdef --flex '{"type":"bubble",...}'

  # Send an image message
  line message push --to U1234567890abcdef --image https://example.com/image.jpg

  # Send a video message
  line message push --to U1234567890abcdef --video https://example.com/video.mp4 --preview https://example.com/preview.jpg

  # Send an audio message
  line message push --to U1234567890abcdef --audio https://example.com/audio.m4a --duration 60000

  # Send a location message
  line message push --to U1234567890abcdef --location-title "Tokyo Tower" --location-address "4-2-8 Shiba-koen, Minato-ku, Tokyo" --lat 35.6586 --lng 139.7454

  # Send a sticker
  line message push --to U1234567890abcdef --sticker-package 446 --sticker-id 1988`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if userID == "" {
				return fmt.Errorf("--to is required: specify a user ID")
			}

			// Validate exactly one message type is specified
			if err := requireExactlyOneFlag([]FlagCheck{
				{Name: "--text", Set: text != ""},
				{Name: "--flex", Set: flexJSON != ""},
				{Name: "--image", Set: imageURL != ""},
				{Name: "--video", Set: videoURL != ""},
				{Name: "--audio", Set: audioURL != ""},
				{Name: "--location-*", Set: locationTitle != "" || locationAddress != "" || lat != 0 || lng != 0},
				{Name: "--sticker-*", Set: packageID != "" || stickerID != ""},
			}); err != nil {
				return err
			}

			// Validate sticker flags are used together
			if (packageID != "" && stickerID == "") || (packageID == "" && stickerID != "") {
				return fmt.Errorf("--sticker-package and --sticker-id must be used together")
			}

			target := messageTarget{Type: "push", UserID: userID}
			return dispatchMessage(cmd, client, target, text, flexJSON, altText, imageURL, previewURL, videoURL, audioURL, duration, locationTitle, locationAddress, lat, lng, packageID, stickerID)
		},
	}

	cmd.Flags().StringVar(&userID, "to", "", "User ID to send message to (required)")
	cmd.Flags().StringVar(&text, "text", "", "Text message content")
	cmd.Flags().StringVar(&flexJSON, "flex", "", "Flex message JSON")
	cmd.Flags().StringVar(&altText, "alt-text", "Flex message", "Alt text for flex messages (shown in notifications)")
	cmd.Flags().StringVar(&imageURL, "image", "", "Image URL to send")
	cmd.Flags().StringVar(&videoURL, "video", "", "Video URL to send")
	cmd.Flags().StringVar(&audioURL, "audio", "", "Audio URL to send")
	cmd.Flags().IntVar(&duration, "duration", 0, "Audio duration in milliseconds (required for --audio)")
	cmd.Flags().StringVar(&previewURL, "preview", "", "Preview image URL (required for --video, defaults to --image for images)")
	cmd.Flags().StringVar(&locationTitle, "location-title", "", "Location title")
	cmd.Flags().StringVar(&locationAddress, "location-address", "", "Location address")
	cmd.Flags().Float64Var(&lat, "lat", 0, "Latitude for location message")
	cmd.Flags().Float64Var(&lng, "lng", 0, "Longitude for location message")
	cmd.Flags().StringVar(&packageID, "sticker-package", "", "Sticker package ID")
	cmd.Flags().StringVar(&stickerID, "sticker-id", "", "Sticker ID")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

// newMessageBroadcastCmd creates a broadcast message command.
func newMessageBroadcastCmd() *cobra.Command {
	return newMessageBroadcastCmdWithClient(nil)
}

// newMessageBroadcastCmdWithClient creates a broadcast message command with an optional API client for testing.
func newMessageBroadcastCmdWithClient(client *api.Client) *cobra.Command {
	var text string
	var flexJSON string
	var altText string
	var imageURL string
	var previewURL string
	var packageID string
	var stickerID string
	var videoURL string
	var audioURL string
	var duration int
	var locationTitle string
	var locationAddress string
	var lat float64
	var lng float64

	cmd := &cobra.Command{
		Use:   "broadcast",
		Short: "Broadcast a message to all followers",
		Long:  "Send a text, flex, image, video, audio, location, or sticker message to all users who follow your LINE Official Account.",
		Example: `  # Broadcast a text message
  line message broadcast --text "Hello everyone!"

  # Broadcast a flex message
  line message broadcast --flex '{"type":"bubble",...}'

  # Broadcast an image
  line message broadcast --image https://example.com/image.jpg

  # Broadcast a video
  line message broadcast --video https://example.com/video.mp4 --preview https://example.com/preview.jpg

  # Broadcast an audio message
  line message broadcast --audio https://example.com/audio.m4a --duration 60000

  # Broadcast a location
  line message broadcast --location-title "Tokyo Tower" --location-address "4-2-8 Shiba-koen, Minato-ku, Tokyo" --lat 35.6586 --lng 139.7454

  # Broadcast a sticker
  line message broadcast --sticker-package 446 --sticker-id 1988`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate exactly one message type is specified
			if err := requireExactlyOneFlag([]FlagCheck{
				{Name: "--text", Set: text != ""},
				{Name: "--flex", Set: flexJSON != ""},
				{Name: "--image", Set: imageURL != ""},
				{Name: "--video", Set: videoURL != ""},
				{Name: "--audio", Set: audioURL != ""},
				{Name: "--location-*", Set: locationTitle != "" || locationAddress != "" || lat != 0 || lng != 0},
				{Name: "--sticker-*", Set: packageID != "" || stickerID != ""},
			}); err != nil {
				return err
			}

			// Validate sticker flags are used together
			if (packageID != "" && stickerID == "") || (packageID == "" && stickerID != "") {
				return fmt.Errorf("--sticker-package and --sticker-id must be used together")
			}

			// Require confirmation for broadcast unless --yes is set
			if !flags.Yes {
				_, _ = fmt.Fprint(cmd.OutOrStdout(), "This will broadcast to ALL followers. Continue? [y/N]: ")
				var response string
				_, _ = fmt.Fscanln(cmd.InOrStdin(), &response)
				if response != "y" && response != "Y" && response != "yes" {
					return fmt.Errorf("broadcast cancelled")
				}
			}

			target := messageTarget{Type: "broadcast"}
			return dispatchMessage(cmd, client, target, text, flexJSON, altText, imageURL, previewURL, videoURL, audioURL, duration, locationTitle, locationAddress, lat, lng, packageID, stickerID)
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "Text message content")
	cmd.Flags().StringVar(&flexJSON, "flex", "", "Flex message JSON")
	cmd.Flags().StringVar(&altText, "alt-text", "Flex message", "Alt text for flex messages (shown in notifications)")
	cmd.Flags().StringVar(&imageURL, "image", "", "Image URL to broadcast")
	cmd.Flags().StringVar(&videoURL, "video", "", "Video URL to broadcast")
	cmd.Flags().StringVar(&audioURL, "audio", "", "Audio URL to broadcast")
	cmd.Flags().IntVar(&duration, "duration", 0, "Audio duration in milliseconds (required for --audio)")
	cmd.Flags().StringVar(&previewURL, "preview", "", "Preview image URL (required for --video, defaults to --image for images)")
	cmd.Flags().StringVar(&locationTitle, "location-title", "", "Location title")
	cmd.Flags().StringVar(&locationAddress, "location-address", "", "Location address")
	cmd.Flags().Float64Var(&lat, "lat", 0, "Latitude for location message")
	cmd.Flags().Float64Var(&lng, "lng", 0, "Longitude for location message")
	cmd.Flags().StringVar(&packageID, "sticker-package", "", "Sticker package ID")
	cmd.Flags().StringVar(&stickerID, "sticker-id", "", "Sticker ID")

	return cmd
}

// newMessageMulticastCmd creates a multicast message command.
func newMessageMulticastCmd() *cobra.Command {
	return newMessageMulticastCmdWithClient(nil)
}

// newMessageMulticastCmdWithClient creates a multicast message command with an optional API client for testing.
func newMessageMulticastCmdWithClient(client *api.Client) *cobra.Command {
	var userIDs []string
	var text string
	var flexJSON string
	var altText string
	var imageURL string
	var previewURL string
	var packageID string
	var stickerID string
	var videoURL string
	var audioURL string
	var duration int
	var locationTitle string
	var locationAddress string
	var lat float64
	var lng float64

	cmd := &cobra.Command{
		Use:   "multicast",
		Short: "Send message to multiple users",
		Long:  "Send a text, flex, image, video, audio, location, or sticker message to multiple users (max 500 per request).",
		Example: `  # Send text to multiple users
  line message multicast --to U123,U456,U789 --text "Hello!"

  # Send flex message
  line message multicast --to U123,U456 --flex '{"type":"bubble",...}'

  # Send an image
  line message multicast --to U123,U456 --image https://example.com/image.jpg

  # Send a video
  line message multicast --to U123,U456 --video https://example.com/video.mp4 --preview https://example.com/preview.jpg

  # Send an audio message
  line message multicast --to U123,U456 --audio https://example.com/audio.m4a --duration 60000

  # Send a location
  line message multicast --to U123,U456 --location-title "Tokyo Tower" --location-address "4-2-8 Shiba-koen, Minato-ku, Tokyo" --lat 35.6586 --lng 139.7454

  # Send a sticker
  line message multicast --to U123,U456 --sticker-package 446 --sticker-id 1988`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(userIDs) == 0 {
				return fmt.Errorf("--to is required: specify comma-separated user IDs")
			}
			if len(userIDs) > 500 {
				return fmt.Errorf("too many users: max 500 per request, got %d", len(userIDs))
			}

			// Validate exactly one message type is specified
			if err := requireExactlyOneFlag([]FlagCheck{
				{Name: "--text", Set: text != ""},
				{Name: "--flex", Set: flexJSON != ""},
				{Name: "--image", Set: imageURL != ""},
				{Name: "--video", Set: videoURL != ""},
				{Name: "--audio", Set: audioURL != ""},
				{Name: "--location-*", Set: locationTitle != "" || locationAddress != "" || lat != 0 || lng != 0},
				{Name: "--sticker-*", Set: packageID != "" || stickerID != ""},
			}); err != nil {
				return err
			}

			// Validate sticker flags are used together
			if (packageID != "" && stickerID == "") || (packageID == "" && stickerID != "") {
				return fmt.Errorf("--sticker-package and --sticker-id must be used together")
			}

			target := messageTarget{Type: "multicast", UserIDs: userIDs}
			return dispatchMessage(cmd, client, target, text, flexJSON, altText, imageURL, previewURL, videoURL, audioURL, duration, locationTitle, locationAddress, lat, lng, packageID, stickerID)
		},
	}

	cmd.Flags().StringSliceVar(&userIDs, "to", nil, "Comma-separated user IDs (required, max 500)")
	cmd.Flags().StringVar(&text, "text", "", "Text message content")
	cmd.Flags().StringVar(&flexJSON, "flex", "", "Flex message JSON")
	cmd.Flags().StringVar(&altText, "alt-text", "Flex message", "Alt text for flex messages")
	cmd.Flags().StringVar(&imageURL, "image", "", "Image URL to send")
	cmd.Flags().StringVar(&videoURL, "video", "", "Video URL to send")
	cmd.Flags().StringVar(&audioURL, "audio", "", "Audio URL to send")
	cmd.Flags().IntVar(&duration, "duration", 0, "Audio duration in milliseconds (required for --audio)")
	cmd.Flags().StringVar(&previewURL, "preview", "", "Preview image URL (required for --video, defaults to --image for images)")
	cmd.Flags().StringVar(&locationTitle, "location-title", "", "Location title")
	cmd.Flags().StringVar(&locationAddress, "location-address", "", "Location address")
	cmd.Flags().Float64Var(&lat, "lat", 0, "Latitude for location message")
	cmd.Flags().Float64Var(&lng, "lng", 0, "Longitude for location message")
	cmd.Flags().StringVar(&packageID, "sticker-package", "", "Sticker package ID")
	cmd.Flags().StringVar(&stickerID, "sticker-id", "", "Sticker ID")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}
