package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newMembershipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "membership",
		Short: "Manage memberships (Japan-only)",
		Long: `Manage LINE Official Account memberships.
Note: This feature is only available for accounts in Japan.`,
	}

	cmd.AddCommand(newMembershipPlansCmd())
	cmd.AddCommand(newMembershipStatusCmd())
	cmd.AddCommand(newMembershipUsersCmd())
	return cmd
}

func newMembershipPlansCmd() *cobra.Command {
	return newMembershipPlansCmdWithClient(nil)
}

func newMembershipPlansCmdWithClient(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Short: "List membership plans",
		Long:  "Get a list of membership plans offered by your LINE Official Account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			plans, err := c.GetMembershipPlans(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get membership plans: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{"plans": plans})
			}

			if len(plans) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No membership plans found")
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Membership Plans:")
			for _, plan := range plans {
				status := ""
				if plan.IsPublished && plan.IsInSale {
					status = " (active)"
				} else if plan.IsPublished {
					status = " (published, not for sale)"
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %d: %s - %d %s%s\n",
					plan.MembershipID, plan.Title, plan.Price, plan.Currency, status)
			}
			return nil
		},
	}

	return cmd
}

func newMembershipStatusCmd() *cobra.Command {
	return newMembershipStatusCmdWithClient(nil)
}

func newMembershipStatusCmdWithClient(client *api.Client) *cobra.Command {
	var userID string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get user's membership status",
		Long:  "Check a user's membership subscription status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if userID == "" {
				return fmt.Errorf("--user is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			memberships, err := c.GetUserMembershipStatus(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("failed to get membership status: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{"userId": userID, "memberships": memberships})
			}

			if len(memberships) == 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "User %s has no memberships\n", userID)
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "User: %s\n", userID)
			for _, m := range memberships {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Membership %d: %s\n", m.MembershipID, m.SubscriptionState)
				if m.StartTime > 0 {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "    Started: %s\n", time.Unix(m.StartTime/1000, 0).Format(time.RFC3339))
				}
				if m.EndTime > 0 {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "    Ends: %s\n", time.Unix(m.EndTime/1000, 0).Format(time.RFC3339))
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user")

	return cmd
}

func newMembershipUsersCmd() *cobra.Command {
	return newMembershipUsersCmdWithClient(nil)
}

func newMembershipUsersCmdWithClient(client *api.Client) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "users",
		Short: "List membership subscribers",
		Long:  "Get a list of users who have joined memberships.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			var allUserIDs []string
			var next string
			for {
				resp, err := c.GetMembershipUsers(cmd.Context(), next)
				if err != nil {
					return fmt.Errorf("failed to get membership users: %w", err)
				}
				allUserIDs = append(allUserIDs, resp.MemberIDs...)
				if resp.Next == "" || !all {
					break
				}
				next = resp.Next
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{"count": len(allUserIDs), "userIds": allUserIDs})
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Membership Subscribers: %d\n", len(allUserIDs))
			if all {
				for _, id := range allUserIDs {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), id)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Fetch and list all user IDs (paginated)")

	return cmd
}
