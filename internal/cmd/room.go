package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newRoomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "room",
		Short: "Manage multi-person chats (rooms)",
		Long:  "Get information about rooms your bot is a member of.",
	}

	cmd.AddCommand(newRoomMembersCmd())
	cmd.AddCommand(newRoomProfileCmd())
	cmd.AddCommand(newRoomLeaveCmd())
	return cmd
}

func newRoomMembersCmd() *cobra.Command {
	return newRoomMembersCmdWithClient(nil)
}

func newRoomMembersCmdWithClient(client *api.Client) *cobra.Command {
	var roomID string
	var all bool

	cmd := &cobra.Command{
		Use:   "members",
		Short: "List room members",
		Long:  "Get member count and list of user IDs in a room.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if roomID == "" {
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

			count, err := c.GetRoomMemberCount(cmd.Context(), roomID)
			if err != nil {
				return fmt.Errorf("failed to get member count: %w", err)
			}

			var allMemberIDs []string
			if all {
				var next string
				for {
					resp, err := c.GetRoomMemberIDs(cmd.Context(), roomID, next)
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
				result := map[string]any{"roomId": roomID, "count": count}
				if all {
					result["memberIds"] = allMemberIDs
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Room ID:  %s\n", roomID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Members:  %d\n", count)
			if all {
				for _, id := range allMemberIDs {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), id)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&roomID, "id", "", "Room ID (required)")
	cmd.Flags().BoolVar(&all, "all", false, "List all member IDs")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newRoomProfileCmd() *cobra.Command {
	return newRoomProfileCmdWithClient(nil)
}

func newRoomProfileCmdWithClient(client *api.Client) *cobra.Command {
	var roomID string
	var userID string

	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Get a room member's profile",
		Long:  "Get the profile of a specific member in a room.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if roomID == "" {
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

			profile, err := c.GetRoomMemberProfile(cmd.Context(), roomID, userID)
			if err != nil {
				return fmt.Errorf("failed to get member profile: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(profile)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "User ID:      %s\n", profile.UserID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", profile.DisplayName)
			if profile.PictureURL != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Picture URL:  %s\n", profile.PictureURL)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&roomID, "id", "", "Room ID (required)")
	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("user")

	return cmd
}

func newRoomLeaveCmd() *cobra.Command {
	return newRoomLeaveCmdWithClient(nil)
}

func newRoomLeaveCmdWithClient(client *api.Client) *cobra.Command {
	var roomID string

	cmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave a room",
		Long:  "Make your bot leave a multi-person chat room.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if roomID == "" {
				return fmt.Errorf("--id is required")
			}

			if !flags.Yes {
				return fmt.Errorf("use --yes to confirm leaving the room")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.LeaveRoom(cmd.Context(), roomID); err != nil {
				return fmt.Errorf("failed to leave room: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"roomId": roomID, "status": "left"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Left room: %s\n", roomID)
			return nil
		},
	}

	cmd.Flags().StringVar(&roomID, "id", "", "Room ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
