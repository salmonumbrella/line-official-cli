package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newBotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bot",
		Short: "Get bot information",
		Long:  "Get information about your LINE Official Account bot.",
	}

	cmd.AddCommand(newBotInfoCmd())
	cmd.AddCommand(newBotProfileCmd())
	cmd.AddCommand(newBotFollowersCmd())
	cmd.AddCommand(newBotLinkTokenCmd())
	return cmd
}

func newBotInfoCmd() *cobra.Command {
	return newBotInfoCmdWithClient(nil)
}

func newBotInfoCmdWithClient(client *api.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Get bot info",
		Long:  "Get basic information about your LINE Official Account including user ID, display name, and settings.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			info, err := c.GetBotInfo(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get bot info: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", info.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "User ID:      %s\n", info.UserID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Basic ID:     %s\n", info.BasicID)
			if info.PremiumID != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Premium ID:   %s\n", info.PremiumID)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Chat Mode:    %s\n", info.ChatMode)
			return nil
		},
	}
}

func newBotProfileCmd() *cobra.Command {
	return newBotProfileCmdWithClient(nil)
}

func newBotProfileCmdWithClient(client *api.Client) *cobra.Command {
	var userID string

	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Get user profile",
		Long:  "Get profile information for a specific user by their user ID.",
		Example: `  # Get a user's profile
  line bot profile --user U1234567890abcdef`,
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

			profile, err := c.GetUserProfile(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("failed to get profile: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(profile)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", profile.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "User ID:      %s\n", profile.UserID)
			if profile.StatusMessage != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:       %s\n", profile.StatusMessage)
			}
			if profile.Language != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Language:     %s\n", profile.Language)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user")

	return cmd
}

func newBotFollowersCmd() *cobra.Command {
	return newBotFollowersCmdWithClient(nil)
}

func newBotFollowersCmdWithClient(client *api.Client) *cobra.Command {
	var limit int
	var all bool

	cmd := &cobra.Command{
		Use:   "followers",
		Short: "List follower IDs",
		Long:  "Get a list of user IDs of users who have added your bot as a friend.",
		Example: `  # Get first 100 followers
  line bot followers

  # Get all followers (paginated)
  line bot followers --all`,
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
				resp, err := c.GetFollowerIDs(cmd.Context(), next, limit)
				if err != nil {
					return fmt.Errorf("failed to get followers: %w", err)
				}

				allUserIDs = append(allUserIDs, resp.UserIDs...)

				if !all || resp.Next == "" {
					break
				}
				next = resp.Next
			}

			if flags.Output == "json" {
				result := map[string]any{"userIds": allUserIDs, "count": len(allUserIDs)}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Followers: %d\n", len(allUserIDs))
			for _, id := range allUserIDs {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), id)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "Number of IDs per request (max 1000)")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all followers (paginated)")

	return cmd
}

func newBotLinkTokenCmd() *cobra.Command {
	return newBotLinkTokenCmdWithClient(nil)
}

func newBotLinkTokenCmdWithClient(client *api.Client) *cobra.Command {
	var userID string

	cmd := &cobra.Command{
		Use:   "link-token",
		Short: "Generate account linking token",
		Long:  "Generate an account linking token for a user. This token is used to link a LINE user with an account in your service.",
		Example: `  # Generate link token for a user
  line bot link-token --user U1234567890abcdef`,
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

			linkToken, err := c.IssueLinkToken(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("failed to issue link token: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]string{"linkToken": linkToken}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Link Token: %s\n", linkToken)
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user")

	return cmd
}
