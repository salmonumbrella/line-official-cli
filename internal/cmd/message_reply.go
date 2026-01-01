package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newMessageReplyCmd() *cobra.Command {
	return newMessageReplyCmdWithClient(nil)
}

func newMessageReplyCmdWithClient(client *api.Client) *cobra.Command {
	var replyToken string
	var text string
	var flexJSON string
	var altText string

	cmd := &cobra.Command{
		Use:   "reply",
		Short: "Reply to a webhook event",
		Long:  "Send a reply message using a reply token from a webhook event. Reply tokens expire after 1 minute.",
		Example: `  # Reply with text
  line message reply --token <replyToken> --text "Thanks for your message!"

  # Reply with flex message
  line message reply --token <replyToken> --flex '{"type":"bubble",...}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if replyToken == "" {
				return fmt.Errorf("--token is required")
			}
			if text == "" && flexJSON == "" {
				return fmt.Errorf("specify --text or --flex")
			}
			if text != "" && flexJSON != "" {
				return fmt.Errorf("specify either --text or --flex, not both")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if text != "" {
				if err := c.ReplyTextMessage(cmd.Context(), replyToken, text); err != nil {
					return fmt.Errorf("failed to reply: %w", err)
				}
			} else {
				if err := c.ReplyFlexMessage(cmd.Context(), replyToken, altText, json.RawMessage(flexJSON)); err != nil {
					return fmt.Errorf("failed to reply: %w", err)
				}
			}

			if flags.Output == "json" {
				result := map[string]any{"status": "sent"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Reply sent")
			return nil
		},
	}

	cmd.Flags().StringVar(&replyToken, "token", "", "Reply token from webhook event (required)")
	cmd.Flags().StringVar(&text, "text", "", "Text message content")
	cmd.Flags().StringVar(&flexJSON, "flex", "", "Flex message JSON")
	cmd.Flags().StringVar(&altText, "alt-text", "Flex message", "Alt text for flex messages")
	_ = cmd.MarkFlagRequired("token")

	return cmd
}
