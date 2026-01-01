package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/salmonumbrella/line-official-cli/internal/api/generated"
	"github.com/spf13/cobra"
)

func newInsightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insight",
		Short: "View analytics and insights",
		Long:  "Get statistics about followers, message delivery, and demographics.",
	}

	cmd.AddCommand(newInsightFollowersCmd())
	cmd.AddCommand(newInsightMessagesCmd())
	cmd.AddCommand(newInsightDemographicsCmd())
	cmd.AddCommand(newInsightEventsCmd())
	cmd.AddCommand(newInsightUnitStatsCmd())

	return cmd
}

func newInsightFollowersCmd() *cobra.Command {
	return newInsightFollowersCmdWithClient(nil)
}

func newInsightFollowersCmdWithClient(client *api.Client) *cobra.Command {
	var date string

	cmd := &cobra.Command{
		Use:   "followers",
		Short: "Get follower statistics",
		Long:  "Get follower count, targeted reaches, and blocks for a specific date.",
		Example: `  # Get yesterday's follower stats
  line insight followers

  # Get stats for a specific date
  line insight followers --date 20250101`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if date == "" {
				// Default to yesterday (insight data has 1-day delay)
				date = time.Now().AddDate(0, 0, -1).Format("20060102")
			}

			// Validate date format
			if len(date) != 8 {
				return fmt.Errorf("date must be in YYYYMMDD format (e.g., 20250101)")
			}
			if _, err := time.Parse("20060102", date); err != nil {
				return fmt.Errorf("invalid date: must be in YYYYMMDD format (e.g., 20250101)")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			stats, err := c.GetFollowerStats(cmd.Context(), date)
			if err != nil {
				return fmt.Errorf("failed to get follower stats: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(stats)
			}

			if stats.Status != nil && *stats.Status == generated.GetNumberOfFollowersResponseStatusReady {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Follower Stats (%s):\n", date)
				if stats.Followers != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Followers:        %d\n", *stats.Followers)
				}
				if stats.TargetedReaches != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Targeted Reaches: %d\n", *stats.TargetedReaches)
				}
				if stats.Blocks != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Blocks:           %d\n", *stats.Blocks)
				}
			} else {
				status := "unknown"
				if stats.Status != nil {
					status = string(*stats.Status)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Stats not ready for %s (status: %s)\n", date, status)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&date, "date", "", "Date in YYYYMMDD format (default: yesterday)")

	return cmd
}

func newInsightMessagesCmd() *cobra.Command {
	return newInsightMessagesCmdWithClient(nil)
}

func newInsightMessagesCmdWithClient(client *api.Client) *cobra.Command {
	var date string

	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Get message delivery statistics",
		Long:  "Get message delivery counts by type for a specific date.",
		Example: `  # Get yesterday's message stats
  line insight messages

  # Get stats for a specific date
  line insight messages --date 20250101`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if date == "" {
				date = time.Now().AddDate(0, 0, -1).Format("20060102")
			}

			// Validate date format
			if len(date) != 8 {
				return fmt.Errorf("date must be in YYYYMMDD format (e.g., 20250101)")
			}
			if _, err := time.Parse("20060102", date); err != nil {
				return fmt.Errorf("invalid date: must be in YYYYMMDD format (e.g., 20250101)")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			stats, err := c.GetMessageDeliveryStats(cmd.Context(), date)
			if err != nil {
				return fmt.Errorf("failed to get message stats: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(stats)
			}

			if stats.Status != nil && *stats.Status == generated.GetNumberOfMessageDeliveriesResponseStatusReady {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Message Delivery Stats (%s):\n", date)

				var total int64
				if stats.Broadcast != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Broadcast:      %d\n", *stats.Broadcast)
					total += *stats.Broadcast
				}
				if stats.Targeting != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Targeting:      %d\n", *stats.Targeting)
					total += *stats.Targeting
				}
				if stats.AutoResponse != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Auto Response:  %d\n", *stats.AutoResponse)
					total += *stats.AutoResponse
				}
				if stats.WelcomeResponse != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Welcome:        %d\n", *stats.WelcomeResponse)
					total += *stats.WelcomeResponse
				}
				if stats.Chat != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Chat:           %d\n", *stats.Chat)
					total += *stats.Chat
				}
				// API-sent messages
				var apiTotal int64
				if stats.ApiPush != nil {
					apiTotal += *stats.ApiPush
				}
				if stats.ApiReply != nil {
					apiTotal += *stats.ApiReply
				}
				if stats.ApiBroadcast != nil {
					apiTotal += *stats.ApiBroadcast
				}
				if stats.ApiMulticast != nil {
					apiTotal += *stats.ApiMulticast
				}
				if stats.ApiNarrowcast != nil {
					apiTotal += *stats.ApiNarrowcast
				}
				if apiTotal > 0 {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  API:            %d\n", apiTotal)
					total += apiTotal
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Total:          %d\n", total)
			} else {
				status := "unknown"
				if stats.Status != nil {
					status = string(*stats.Status)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Stats not ready for %s (status: %s)\n", date, status)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&date, "date", "", "Date in YYYYMMDD format (default: yesterday)")

	return cmd
}

func newInsightDemographicsCmd() *cobra.Command {
	return newInsightDemographicsCmdWithClient(nil)
}

func newInsightDemographicsCmdWithClient(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demographics",
		Short: "Get friend demographics",
		Long:  "Get demographic breakdown of friends by age, gender, area, and more.",
		Example: `  # Get friend demographics
  line insight demographics`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			demo, err := c.GetFriendsDemographics(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get demographics: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(demo)
			}

			if demo.Available != nil && !*demo.Available {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Demographics data not available.")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Note: Requires at least 20 friends for data to be available.")
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Friend Demographics:")

			// Gender breakdown
			if demo.Genders != nil && len(*demo.Genders) > 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\n  Gender:")
				for _, g := range *demo.Genders {
					if g.Gender != nil && g.Percentage != nil {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "    %-10s %5.1f%%\n", *g.Gender, *g.Percentage)
					}
				}
			}

			// Age breakdown
			if demo.Ages != nil && len(*demo.Ages) > 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\n  Age:")
				for _, a := range *demo.Ages {
					if a.Age != nil && a.Percentage != nil {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "    %-12s %5.1f%%\n", *a.Age, *a.Percentage)
					}
				}
			}

			// App type breakdown
			if demo.AppTypes != nil && len(*demo.AppTypes) > 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\n  Platform:")
				for _, t := range *demo.AppTypes {
					if t.AppType != nil && t.Percentage != nil {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "    %-10s %5.1f%%\n", *t.AppType, *t.Percentage)
					}
				}
			}

			// Area breakdown (top 5)
			if demo.Areas != nil && len(*demo.Areas) > 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\n  Top Areas:")
				count := 0
				for _, a := range *demo.Areas {
					if count >= 5 {
						break
					}
					if a.Area != nil && a.Percentage != nil {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "    %-20s %5.1f%%\n", *a.Area, *a.Percentage)
						count++
					}
				}
			}

			// Subscription period
			if demo.SubscriptionPeriods != nil && len(*demo.SubscriptionPeriods) > 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\n  Subscription Period:")
				for _, s := range *demo.SubscriptionPeriods {
					if s.SubscriptionPeriod != nil && s.Percentage != nil {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "    %-15s %5.1f%%\n", *s.SubscriptionPeriod, *s.Percentage)
					}
				}
			}

			return nil
		},
	}

	return cmd
}

func newInsightEventsCmd() *cobra.Command {
	return newInsightEventsCmdWithClient(nil)
}

func newInsightEventsCmdWithClient(client *api.Client) *cobra.Command {
	var requestID string

	cmd := &cobra.Command{
		Use:   "events",
		Short: "Get message event statistics",
		Long:  "Get detailed event statistics (impressions, clicks) for a specific message request.",
		Example: `  # Get events for a message request
  line insight events --request-id xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`,
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

			stats, err := c.GetMessageEventStats(cmd.Context(), requestID)
			if err != nil {
				return fmt.Errorf("failed to get event stats: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(stats)
			}

			if stats.Overview != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Message Event Statistics:")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Delivered:          %d\n", stats.Overview.Delivered)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Unique Impressions: %d\n", stats.Overview.UniqueImpression)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Unique Clicks:      %d\n", stats.Overview.UniqueClick)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&requestID, "request-id", "", "Message request ID (required)")
	_ = cmd.MarkFlagRequired("request-id")

	return cmd
}

func newInsightUnitStatsCmd() *cobra.Command {
	return newInsightUnitStatsCmdWithClient(nil)
}

func newInsightUnitStatsCmdWithClient(client *api.Client) *cobra.Command {
	var unit string
	var from string
	var to string

	cmd := &cobra.Command{
		Use:   "unit-stats",
		Short: "Get statistics per aggregation unit",
		Long:  "Get event statistics aggregated by a custom aggregation unit for a date range.",
		Example: `  # Get stats for a specific unit over the past week
  line insight unit-stats --unit campaign-2024 --from 20251224 --to 20251231`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if unit == "" {
				return fmt.Errorf("--unit is required")
			}
			if from == "" {
				return fmt.Errorf("--from is required (format: YYYYMMDD)")
			}
			if to == "" {
				return fmt.Errorf("--to is required (format: YYYYMMDD)")
			}

			// Validate date formats
			if len(from) != 8 {
				return fmt.Errorf("--from must be in YYYYMMDD format (e.g., 20250101)")
			}
			if _, err := time.Parse("20060102", from); err != nil {
				return fmt.Errorf("invalid --from date: must be in YYYYMMDD format")
			}
			if len(to) != 8 {
				return fmt.Errorf("--to must be in YYYYMMDD format (e.g., 20250101)")
			}
			if _, err := time.Parse("20060102", to); err != nil {
				return fmt.Errorf("invalid --to date: must be in YYYYMMDD format")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			stats, err := c.GetStatisticsPerUnit(cmd.Context(), unit, from, to)
			if err != nil {
				return fmt.Errorf("failed to get unit statistics: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(stats)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Statistics for unit '%s' (%s to %s):\n", unit, from, to)
			if stats.Overview != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Unique Impressions:  %d\n", stats.Overview.UniqueImpression)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Unique Clicks:       %d\n", stats.Overview.UniqueClick)
				if stats.Overview.UniqueMediaPlayed > 0 {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Unique Media Played: %d\n", stats.Overview.UniqueMediaPlayed)
				}
				if stats.Overview.UniqueMediaPlayedComplete > 0 {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Media Played 100%%:   %d\n", stats.Overview.UniqueMediaPlayedComplete)
				}
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  No statistics available for this unit and date range.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&unit, "unit", "", "Custom aggregation unit name (required)")
	cmd.Flags().StringVar(&from, "from", "", "Start date in YYYYMMDD format (required)")
	cmd.Flags().StringVar(&to, "to", "", "End date in YYYYMMDD format (required)")
	_ = cmd.MarkFlagRequired("unit")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}
