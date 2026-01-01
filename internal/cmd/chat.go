package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newChatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Chat features",
		Long:  "Chat features including loading animation and mark as read.",
	}

	cmd.AddCommand(newChatLoadingCmd())
	cmd.AddCommand(newChatMarkReadCmd())
	return cmd
}

func newChatLoadingCmd() *cobra.Command {
	return newChatLoadingCmdWithClient(nil)
}

func newChatLoadingCmdWithClient(client *api.Client) *cobra.Command {
	var userID string
	var seconds int

	cmd := &cobra.Command{
		Use:   "loading",
		Short: "Show loading animation",
		Long:  "Display a loading animation in a chat. The animation indicates the bot is processing and will be shown for the specified duration (default 5 seconds, max 60 seconds).",
		Example: `  # Show loading animation for default 5 seconds
  line chat loading --user U1234567890abcdef

  # Show loading animation for 10 seconds
  line chat loading --user U1234567890abcdef --seconds 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if userID == "" {
				return fmt.Errorf("--user is required")
			}

			if seconds < 1 || seconds > 60 {
				return fmt.Errorf("--seconds must be between 1 and 60")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			err := c.ShowLoadingAnimation(cmd.Context(), userID, seconds)
			if err != nil {
				return fmt.Errorf("failed to show loading animation: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"success": true,
					"chatId":  userID,
					"seconds": seconds,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Loading animation started for %d seconds\n", seconds)
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID / chat ID (required)")
	cmd.Flags().IntVar(&seconds, "seconds", 5, "Loading duration in seconds (1-60)")
	_ = cmd.MarkFlagRequired("user")

	return cmd
}

func newChatMarkReadCmd() *cobra.Command {
	return newChatMarkReadCmdWithClient(nil)
}

func newChatMarkReadCmdWithClient(client *api.Client) *cobra.Command {
	var userID string
	var chatToken string

	cmd := &cobra.Command{
		Use:   "mark-read",
		Short: "Mark messages as read",
		Long: `Mark messages as read using either a user ID or a chat token.

Use --user to mark all messages from a specific user as read (uses /v2/bot/message/markAsRead).
Use --token to mark messages as read using a chat token from a webhook event (uses /v2/bot/chat/markAsRead).

Only one of --user or --token should be specified.`,
		Example: `  # Mark all messages from a user as read
  line chat mark-read --user U1234567890abcdef

  # Mark messages as read using a chat token from webhook
  line chat mark-read --token abc123xyz`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if userID == "" && chatToken == "" {
				return fmt.Errorf("either --user or --token is required")
			}
			if userID != "" && chatToken != "" {
				return fmt.Errorf("only one of --user or --token should be specified")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if chatToken != "" {
				// Token-based mark as read
				err := c.MarkMessagesAsReadByToken(cmd.Context(), chatToken)
				if err != nil {
					return fmt.Errorf("failed to mark messages as read: %w", err)
				}

				if flags.Output == "json" {
					result := map[string]any{
						"success": true,
						"token":   chatToken,
					}
					enc := json.NewEncoder(cmd.OutOrStdout())
					enc.SetIndent("", "  ")
					return enc.Encode(result)
				}

				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Messages marked as read")
				return nil
			}

			// User-based mark as read
			err := c.MarkMessagesAsRead(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("failed to mark messages as read: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"success": true,
					"userId":  userID,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Messages marked as read")
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID (mutually exclusive with --token)")
	cmd.Flags().StringVar(&chatToken, "token", "", "Chat token from webhook event (mutually exclusive with --user)")

	return cmd
}
