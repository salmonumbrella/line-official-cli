package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newMessageNarrowcastCmd() *cobra.Command {
	return newMessageNarrowcastCmdWithClient(nil)
}

func newMessageNarrowcastCmdWithClient(client *api.Client) *cobra.Command {
	var text string
	var audienceID int64

	cmd := &cobra.Command{
		Use:   "narrowcast",
		Short: "Send message to targeted users",
		Long: `Send a message to users matching specific criteria.
Can target an audience group or use demographic filters.`,
		Example: `  # Send to an audience group
  line message narrowcast --text "Special offer!" --audience 12345678

  # Check narrowcast progress
  line message narrowcast-status --request-id <id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if text == "" {
				return fmt.Errorf("--text is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			resp, err := c.NarrowcastTextMessage(cmd.Context(), text, audienceID)
			if err != nil {
				return fmt.Errorf("failed to narrowcast: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Narrowcast queued: %s\n", resp.RequestID)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Use 'line message narrowcast-status --request-id <id>' to check progress")
			return nil
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "Text message content (required)")
	cmd.Flags().Int64Var(&audienceID, "audience", 0, "Audience group ID to target")
	_ = cmd.MarkFlagRequired("text")

	return cmd
}

func newMessageNarrowcastStatusCmd() *cobra.Command {
	return newMessageNarrowcastStatusCmdWithClient(nil)
}

func newMessageNarrowcastStatusCmdWithClient(client *api.Client) *cobra.Command {
	var requestID string

	cmd := &cobra.Command{
		Use:   "narrowcast-status",
		Short: "Check narrowcast progress",
		Long:  "Get the progress status of a narrowcast message.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if requestID == "" {
				return fmt.Errorf("--request-id is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			progress, err := c.GetNarrowcastProgress(cmd.Context(), requestID)
			if err != nil {
				return fmt.Errorf("failed to get progress: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(progress)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Phase: %v\n", progress["phase"])
			if count, ok := progress["successCount"]; ok {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Success: %v\n", count)
			}
			if count, ok := progress["failureCount"]; ok {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Failure: %v\n", count)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&requestID, "request-id", "", "Request ID from narrowcast (required)")
	_ = cmd.MarkFlagRequired("request-id")

	return cmd
}
