package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage group chats",
		Long:  "Get information about groups your bot is a member of.",
	}

	cmd.AddCommand(newGroupSummaryCmd())
	cmd.AddCommand(newGroupMembersCmd())
	cmd.AddCommand(newGroupMemberProfileCmd())
	cmd.AddCommand(newGroupLeaveCmd())
	return cmd
}

func newGroupSummaryCmd() *cobra.Command {
	return newGroupSummaryCmdWithClient(nil)
}

func newGroupSummaryCmdWithClient(client *api.Client) *cobra.Command {
	var groupID string

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Get group summary",
		Long:  "Get summary information about a group (name, picture).",
		RunE: func(cmd *cobra.Command, args []string) error {
			if groupID == "" {
				return fmt.Errorf("--id is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			summary, err := c.GetGroupSummary(cmd.Context(), groupID)
			if err != nil {
				return fmt.Errorf("failed to get group summary: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(summary)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Group ID:   %s\n", summary.GroupID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Group Name: %s\n", summary.GroupName)
			return nil
		},
	}

	cmd.Flags().StringVar(&groupID, "id", "", "Group ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newGroupMembersCmd() *cobra.Command {
	return newGroupMembersCmdWithClient(nil)
}

func newGroupMembersCmdWithClient(client *api.Client) *cobra.Command {
	var groupID string
	var all bool

	cmd := &cobra.Command{
		Use:   "members",
		Short: "List group members",
		Long:  "Get member count and list of user IDs in a group.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if groupID == "" {
				return fmt.Errorf("--id is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			count, err := c.GetGroupMemberCount(cmd.Context(), groupID)
			if err != nil {
				return fmt.Errorf("failed to get member count: %w", err)
			}

			var allMemberIDs []string
			if all {
				var next string
				for {
					resp, err := c.GetGroupMemberIDs(cmd.Context(), groupID, next)
					if err != nil {
						return fmt.Errorf("failed to get member IDs: %w", err)
					}
					allMemberIDs = append(allMemberIDs, resp.MemberIDs...)
					if resp.Next == "" {
						break
					}
					next = resp.Next
				}
			}

			if flags.Output == "json" {
				result := map[string]any{"groupId": groupID, "count": count}
				if all {
					result["memberIds"] = allMemberIDs
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Group ID: %s\n", groupID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Members:  %d\n", count)
			if all {
				for _, id := range allMemberIDs {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), id)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&groupID, "id", "", "Group ID (required)")
	cmd.Flags().BoolVar(&all, "all", false, "List all member IDs")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newGroupMemberProfileCmd() *cobra.Command {
	return newGroupMemberProfileCmdWithClient(nil)
}

func newGroupMemberProfileCmdWithClient(client *api.Client) *cobra.Command {
	var groupID string
	var userID string

	cmd := &cobra.Command{
		Use:   "member-profile",
		Short: "Get a group member's profile",
		Long:  "Get profile information for a specific member in a group.",
		Example: `  # Get a member's profile
  line group member-profile --id C1234567890abcdef --user U1234567890abcdef`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if groupID == "" {
				return fmt.Errorf("--id is required")
			}
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

			profile, err := c.GetGroupMemberProfile(cmd.Context(), groupID, userID)
			if err != nil {
				return fmt.Errorf("failed to get member profile: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(profile)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "User ID:     %s\n", profile.UserID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", profile.DisplayName)
			if profile.PictureURL != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Picture URL:  %s\n", profile.PictureURL)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&groupID, "id", "", "Group ID (required)")
	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("user")

	return cmd
}

func newGroupLeaveCmd() *cobra.Command {
	return newGroupLeaveCmdWithClient(nil)
}

func newGroupLeaveCmdWithClient(client *api.Client) *cobra.Command {
	var groupID string

	cmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave a group",
		Long:  "Make your bot leave a group chat.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if groupID == "" {
				return fmt.Errorf("--id is required")
			}

			if !flags.Yes {
				return fmt.Errorf("use --yes to confirm leaving the group")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.LeaveGroup(cmd.Context(), groupID); err != nil {
				return fmt.Errorf("failed to leave group: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"groupId": groupID, "status": "left"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Left group: %s\n", groupID)
			return nil
		},
	}

	cmd.Flags().StringVar(&groupID, "id", "", "Group ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
