package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Manage webhook settings",
		Long:  "Get, set, and test your LINE bot webhook endpoint.",
	}

	cmd.AddCommand(newWebhookGetCmd())
	cmd.AddCommand(newWebhookSetCmd())
	cmd.AddCommand(newWebhookTestCmd())
	cmd.AddCommand(newWebhookServeCmd())
	return cmd
}

func newWebhookGetCmd() *cobra.Command {
	return newWebhookGetCmdWithClient(nil)
}

func newWebhookGetCmdWithClient(client *api.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get current webhook endpoint",
		Long:  "Get the currently configured webhook endpoint URL and status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			info, err := c.GetWebhookEndpoint(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get webhook: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Endpoint: %s\n", info.Endpoint)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Active:   %v\n", info.Active)
			return nil
		},
	}
}

func newWebhookSetCmd() *cobra.Command {
	return newWebhookSetCmdWithClient(nil)
}

func newWebhookSetCmdWithClient(client *api.Client) *cobra.Command {
	var endpoint string

	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Set webhook endpoint URL",
		Long:    "Configure the webhook endpoint URL for your LINE bot.",
		Example: `  line webhook set --url https://example.com/webhook`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if endpoint == "" {
				return fmt.Errorf("--url is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.SetWebhookEndpoint(cmd.Context(), endpoint); err != nil {
				return fmt.Errorf("failed to set webhook: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"endpoint": endpoint, "status": "set"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Webhook set to: %s\n", endpoint)
			return nil
		},
	}

	cmd.Flags().StringVar(&endpoint, "url", "", "Webhook URL (required)")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func newWebhookTestCmd() *cobra.Command {
	return newWebhookTestCmdWithClient(nil)
}

func newWebhookTestCmdWithClient(client *api.Client) *cobra.Command {
	var endpoint string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test webhook endpoint",
		Long:  "Send a test request to verify webhook connectivity.",
		Example: `  # Test current webhook
  line webhook test

  # Test a specific URL
  line webhook test --url https://example.com/webhook`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			resp, err := c.TestWebhookEndpoint(cmd.Context(), endpoint)
			if err != nil {
				return fmt.Errorf("failed to test webhook: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			if resp.Success {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Webhook test: SUCCESS")
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Webhook test: FAILED")
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:    %d %s\n", resp.StatusCode, resp.Reason)
			if resp.Detail != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Detail:    %s\n", resp.Detail)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&endpoint, "url", "", "Specific URL to test (optional)")

	return cmd
}
