package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newMessageAggregationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aggregation",
		Short: "Manage aggregation units",
		Long:  "Get information about aggregation units used for message tracking.",
	}

	cmd.AddCommand(newMessageAggregationUsageCmd())
	cmd.AddCommand(newMessageAggregationListCmd())

	return cmd
}

func newMessageAggregationUsageCmd() *cobra.Command {
	return newMessageAggregationUsageCmdWithClient(nil)
}

func newMessageAggregationUsageCmdWithClient(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Get aggregation unit usage count",
		Long:  "Show the number of custom aggregation units used this month.",
		Example: `  # Get aggregation unit usage
  line message aggregation usage`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			usage, err := c.GetAggregationUnitUsage(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get aggregation usage: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(usage)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Custom Aggregation Units: %d\n", usage.NumOfCustomAggregationUnits)
			return nil
		},
	}

	return cmd
}

func newMessageAggregationListCmd() *cobra.Command {
	return newMessageAggregationListCmdWithClient(nil)
}

func newMessageAggregationListCmdWithClient(client *api.Client) *cobra.Command {
	var limit int
	var start string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List aggregation unit names",
		Long:  "Get the list of custom aggregation unit names.",
		Example: `  # List all aggregation units
  line message aggregation list

  # List with pagination
  line message aggregation list --limit 10 --start <cursor>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			resp, err := c.GetAggregationUnitNameList(cmd.Context(), limit, start)
			if err != nil {
				return fmt.Errorf("failed to get aggregation unit list: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			if len(resp.CustomAggregationUnits) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No aggregation units found")
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Aggregation Units (%d):\n", len(resp.CustomAggregationUnits))
			for _, unit := range resp.CustomAggregationUnits {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", unit)
			}

			if resp.Next != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nMore results available. Use --start %s to continue.\n", resp.Next)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of units to return (default: API default)")
	cmd.Flags().StringVar(&start, "start", "", "Pagination cursor for continued listing")

	return cmd
}
