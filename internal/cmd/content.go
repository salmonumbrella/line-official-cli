package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newContentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "content",
		Short: "Download message content",
		Long:  "Download images, videos, and audio files sent by users.",
	}

	cmd.AddCommand(newContentDownloadCmd())
	cmd.AddCommand(newContentPreviewCmd())
	cmd.AddCommand(newContentStatusCmd())
	return cmd
}

func newContentDownloadCmd() *cobra.Command {
	return newContentDownloadCmdWithClient(nil)
}

func newContentDownloadCmdWithClient(client *api.Client) *cobra.Command {
	var messageID string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download content from a message",
		Long:  "Download image, video, or audio content from a message by its ID.",
		Example: `  # Download to current directory (auto-named)
  line content download --message-id 123456789

  # Download to specific file
  line content download --message-id 123456789 --output image.jpg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if messageID == "" {
				return fmt.Errorf("--message-id is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			data, contentType, err := c.GetMessageContent(cmd.Context(), messageID)
			if err != nil {
				return fmt.Errorf("failed to download content: %w", err)
			}

			// Determine filename
			filename := outputPath
			if filename == "" {
				ext := ".bin"
				switch {
				case strings.Contains(contentType, "jpeg"):
					ext = ".jpg"
				case strings.Contains(contentType, "png"):
					ext = ".png"
				case strings.Contains(contentType, "gif"):
					ext = ".gif"
				case strings.Contains(contentType, "mp4"):
					ext = ".mp4"
				case strings.Contains(contentType, "audio"):
					ext = ".m4a"
				}
				filename = messageID + ext
			}

			if err := os.WriteFile(filename, data, 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"messageId":   messageID,
					"contentType": contentType,
					"size":        len(data),
					"file":        filename,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			absPath, _ := filepath.Abs(filename)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Downloaded %s (%d bytes)\n", absPath, len(data))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Content-Type: %s\n", contentType)
			return nil
		},
	}

	cmd.Flags().StringVar(&messageID, "message-id", "", "Message ID (required)")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output file path (auto-named if omitted)")
	_ = cmd.MarkFlagRequired("message-id")

	return cmd
}

func newContentPreviewCmd() *cobra.Command {
	return newContentPreviewCmdWithClient(nil)
}

func newContentPreviewCmdWithClient(client *api.Client) *cobra.Command {
	var messageID string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Download preview image for message content",
		Long:  "Download a preview (thumbnail) image for video or image content by message ID.",
		Example: `  # Download preview to current directory (auto-named)
  line content preview --message-id 123456789

  # Download preview to specific file
  line content preview --message-id 123456789 --output preview.jpg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if messageID == "" {
				return fmt.Errorf("--message-id is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			data, contentType, err := c.GetMessageContentPreview(cmd.Context(), messageID)
			if err != nil {
				return fmt.Errorf("failed to download preview: %w", err)
			}

			// Determine filename
			filename := outputPath
			if filename == "" {
				ext := ".jpg" // Previews are typically JPEG
				switch {
				case strings.Contains(contentType, "png"):
					ext = ".png"
				case strings.Contains(contentType, "gif"):
					ext = ".gif"
				}
				filename = "preview-" + messageID + ext
			}

			if err := os.WriteFile(filename, data, 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"messageId":   messageID,
					"contentType": contentType,
					"size":        len(data),
					"file":        filename,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			absPath, _ := filepath.Abs(filename)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Downloaded preview %s (%d bytes)\n", absPath, len(data))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Content-Type: %s\n", contentType)
			return nil
		},
	}

	cmd.Flags().StringVar(&messageID, "message-id", "", "Message ID (required)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (defaults to preview-{id}.jpg)")
	_ = cmd.MarkFlagRequired("message-id")

	return cmd
}

func newContentStatusCmd() *cobra.Command {
	return newContentStatusCmdWithClient(nil)
}

func newContentStatusCmdWithClient(client *api.Client) *cobra.Command {
	var messageID string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check transcoding status of message content",
		Long:  "Check if video or audio content is ready for download (transcoding status).",
		Example: `  # Check transcoding status
  line content status --message-id 123456789`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if messageID == "" {
				return fmt.Errorf("--message-id is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			status, err := c.GetMessageContentTranscoding(cmd.Context(), messageID)
			if err != nil {
				return fmt.Errorf("failed to get transcoding status: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"messageId": messageID,
					"status":    status.Status,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Message ID: %s\n", messageID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Transcoding Status: %s\n", status.Status)
			return nil
		},
	}

	cmd.Flags().StringVar(&messageID, "message-id", "", "Message ID (required)")
	_ = cmd.MarkFlagRequired("message-id")

	return cmd
}
