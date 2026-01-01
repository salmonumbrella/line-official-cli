package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newShopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shop",
		Short: "Manage LINE Shop features",
		Long:  "Access LINE Shop features including mission stickers.",
	}

	cmd.AddCommand(newShopMissionCmd())

	return cmd
}

func newShopMissionCmd() *cobra.Command {
	return newShopMissionCmdWithClient(nil)
}

func newShopMissionCmdWithClient(client *api.Client) *cobra.Command {
	var userID string
	var productID string
	var productType string
	var sendMessage bool

	cmd := &cobra.Command{
		Use:   "mission",
		Short: "Send a mission sticker to a user",
		Long: `Send a mission sticker to a user as a reward for completing a mission.

This feature is for LINE accounts that have mission sticker campaigns.
The user must have completed the required mission to receive the sticker.`,
		Example: `  # Send a mission sticker to a user
  line shop mission --to U1234567890abcdef --product-id 12345 --product-type STICKER

  # Send a mission sticker with a present message
  line shop mission --to U1234567890abcdef --product-id 12345 --product-type STICKER --send-message

  # JSON output
  line shop mission --to U1234567890abcdef --product-id 12345 --product-type STICKER --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if userID == "" {
				return fmt.Errorf("--to is required")
			}
			if productID == "" {
				return fmt.Errorf("--product-id is required")
			}
			if productType == "" {
				return fmt.Errorf("--product-type is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.SendMissionSticker(cmd.Context(), userID, productID, productType, sendMessage); err != nil {
				return fmt.Errorf("failed to send mission sticker: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"to":                 userID,
					"productId":          productID,
					"productType":        productType,
					"sendPresentMessage": sendMessage,
					"status":             "sent",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Mission sticker sent to user %s (product: %s)\n", userID, productID)
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "to", "", "User ID to send the sticker to (required)")
	cmd.Flags().StringVar(&productID, "product-id", "", "Mission sticker product ID (required)")
	cmd.Flags().StringVar(&productType, "product-type", "", "Product type (e.g., STICKER) (required)")
	cmd.Flags().BoolVar(&sendMessage, "send-message", false, "Include a present message")

	return cmd
}
