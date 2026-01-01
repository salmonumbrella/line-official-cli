package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newPNPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pnp",
		Short: "Phone Number Push messaging",
		Long: `Send LINE notification messages to users by phone number.

PNP (Phone Number Push) allows sending messages to users identified
by their phone number instead of LINE user ID. Requires a PNP-enabled channel.`,
	}

	cmd.AddCommand(newPNPPushCmd())

	return cmd
}

func newPNPPushCmd() *cobra.Command {
	return newPNPPushCmdWithClient(nil)
}

func newPNPPushCmdWithClient(client *api.Client) *cobra.Command {
	var phoneNumber string
	var text string

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push message to phone number",
		Long: `Send a text message to a user identified by phone number.

The phone number must include the country code (e.g., +819012345678).
Requires a LINE channel with PNP permissions enabled.`,
		Example: `  # Send a text message to a phone number
  line pnp push --to +819012345678 --text "Hello from LINE!"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if phoneNumber == "" {
				return fmt.Errorf("--to is required: specify a phone number with country code")
			}
			if text == "" {
				return fmt.Errorf("--text is required: specify the message text")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.PNPPushMessage(cmd.Context(), phoneNumber, text); err != nil {
				return fmt.Errorf("failed to send PNP message: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"status": "sent", "to": phoneNumber}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "PNP message sent to %s\n", phoneNumber)
			return nil
		},
	}

	cmd.Flags().StringVar(&phoneNumber, "to", "", "Phone number with country code (required, e.g., +819012345678)")
	cmd.Flags().StringVar(&text, "text", "", "Text message content (required)")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("text")

	return cmd
}
