package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newMessageQuotaCmd() *cobra.Command {
	return newMessageQuotaCmdWithClient(nil)
}

func newMessageQuotaCmdWithClient(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quota",
		Short: "Get message quota and usage",
		Long:  "Show the monthly message limit and current usage for your LINE Official Account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			quota, err := c.GetMessageQuota(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get quota: %w", err)
			}

			consumption, err := c.GetMessageConsumption(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get consumption: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"type":  quota.Type,
					"limit": quota.Value,
					"used":  consumption.TotalUsage,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if quota.Type == "limited" && quota.Value > 0 {
				pct := float64(consumption.TotalUsage) / float64(quota.Value) * 100
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Message Quota: %d/month\n", quota.Value)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Used: %d (%.1f%%)\n", consumption.TotalUsage, pct)
			} else if quota.Type == "limited" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Message Quota: 0/month\n")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Used: %d\n", consumption.TotalUsage)
			} else {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Message Quota: Unlimited\n")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Used: %d\n", consumption.TotalUsage)
			}
			return nil
		},
	}

	return cmd
}

func newMessageDeliveryStatsCmd() *cobra.Command {
	return newMessageDeliveryStatsCmdWithClient(nil)
}

func newMessageDeliveryStatsCmdWithClient(client *api.Client) *cobra.Command {
	var date string
	var messageType string

	cmd := &cobra.Command{
		Use:   "delivery-stats",
		Short: "Get message delivery statistics",
		Long: `Get the number of messages sent on a specific date.
Date format: YYYYMMDD (e.g., 20251230)

Supported message types:
  reply     - Reply messages sent in response to webhook events
  push      - Push messages sent to specific users
  multicast - Messages sent to multiple users at once
  broadcast - Messages sent to all followers
  pnp       - Push notification push messages`,
		Example: `  # Get reply message stats for today
  line message delivery-stats --type reply --date 20251230

  # Get broadcast stats
  line message delivery-stats --type broadcast --date 20251229

  # Get PNP message stats
  line message delivery-stats --type pnp --date 20251230`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if date == "" {
				return fmt.Errorf("--date is required (format: YYYYMMDD)")
			}
			if messageType == "" {
				return fmt.Errorf("--type is required (reply|push|multicast|broadcast|pnp)")
			}

			validTypes := map[string]bool{"reply": true, "push": true, "multicast": true, "broadcast": true, "pnp": true}
			if !validTypes[messageType] {
				return fmt.Errorf("--type must be one of: reply, push, multicast, broadcast, pnp")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			stats, err := getDeliveryStatsByType(cmd, c, messageType, date)
			if err != nil {
				return fmt.Errorf("failed to get delivery stats: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"type":         messageType,
					"date":         date,
					"status":       stats.Status,
					"success":      stats.Success,
					"failure":      stats.Failure,
					"requestCount": stats.RequestCount,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type:          %s\n", messageType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Date:          %s\n", date)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:        %s\n", stats.Status)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Success:       %d\n", stats.Success)
			if stats.Failure > 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Failure:       %d\n", stats.Failure)
			}
			if stats.RequestCount > 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Request Count: %d\n", stats.RequestCount)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&date, "date", "", "Date in YYYYMMDD format (required)")
	cmd.Flags().StringVar(&messageType, "type", "", "Message type: reply|push|multicast|broadcast|pnp (required)")
	_ = cmd.MarkFlagRequired("date")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

// getDeliveryStatsByType calls the appropriate per-type API method
func getDeliveryStatsByType(cmd *cobra.Command, client *api.Client, messageType, date string) (*api.DeliveryStats, error) {
	ctx := cmd.Context()
	switch messageType {
	case "reply":
		return client.GetReplyMessageStats(ctx, date)
	case "push":
		return client.GetPushMessageStats(ctx, date)
	case "multicast":
		return client.GetMulticastMessageStats(ctx, date)
	case "broadcast":
		return client.GetBroadcastMessageStats(ctx, date)
	case "pnp":
		return client.GetPNPMessageStats(ctx, date)
	default:
		return nil, fmt.Errorf("unsupported message type: %s", messageType)
	}
}
