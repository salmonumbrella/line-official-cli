package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newModuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module",
		Short: "Manage LINE module integration",
		Long: `Manage LINE Official Account Manager module integration.

Module channels allow LINE Official Account Manager (OAM) to extend functionality
through a modular architecture. These commands support detaching modules and
controlling chat ownership between the Primary Channel and module channels.`,
	}

	cmd.AddCommand(newModuleDetachCmd())
	cmd.AddCommand(newModuleAcquireCmd())
	cmd.AddCommand(newModuleReleaseCmd())
	cmd.AddCommand(newModuleTokenCmd())
	cmd.AddCommand(newModuleBotsCmd())

	return cmd
}

func newModuleDetachCmd() *cobra.Command {
	return newModuleDetachCmdWithClient(nil)
}

func newModuleDetachCmdWithClient(client *api.Client) *cobra.Command {
	var botID string

	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach module from LINE Official Account",
		Long: `Detach the module channel from a LINE Official Account.

The module channel admin calls this API to unlink the module channel
from a LINE Official Account. This is a destructive action that requires
explicit confirmation.`,
		Example: `  # Detach module (with confirmation prompt)
  line module detach --bot-id U1234567890abcdef

  # Detach module (skip confirmation for automation)
  line module detach --bot-id U1234567890abcdef --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if botID == "" {
				return fmt.Errorf("--bot-id is required")
			}

			// Require confirmation for detach unless --yes is set
			if !flags.Yes {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "This will detach the module from bot %s. Continue? [y/N]: ", botID)
				var response string
				_, _ = fmt.Fscanln(cmd.InOrStdin(), &response)
				if response != "y" && response != "Y" && response != "yes" {
					return fmt.Errorf("detach cancelled")
				}
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.DetachModule(cmd.Context(), botID); err != nil {
				return fmt.Errorf("failed to detach module: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"success": true,
					"botId":   botID,
					"action":  "detached",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Module detached from bot %s\n", botID)
			return nil
		},
	}

	cmd.Flags().StringVar(&botID, "bot-id", "", "User ID of the LINE Official Account bot (required)")
	_ = cmd.MarkFlagRequired("bot-id")

	return cmd
}

func newModuleAcquireCmd() *cobra.Command {
	return newModuleAcquireCmdWithClient(nil)
}

func newModuleAcquireCmdWithClient(client *api.Client) *cobra.Command {
	var chatID string
	var noExpiry bool

	cmd := &cobra.Command{
		Use:   "acquire",
		Short: "Acquire chat control for module",
		Long: `Acquire chat control for a module channel.

When the Primary Channel has chat control, the module channel can call this
API to acquire chat control. The chatId can be a userId, roomId, or groupId.

By default, chat control will return to the Primary Channel after the time
limit (TTL) has passed. Use --no-expiry to keep control indefinitely.`,
		Example: `  # Acquire chat control for a user (with TTL expiry)
  line module acquire --chat U1234567890abcdef

  # Acquire chat control without expiry
  line module acquire --chat U1234567890abcdef --no-expiry`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if chatID == "" {
				return fmt.Errorf("--chat is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			// When noExpiry is true, we set expired=false (control doesn't expire)
			// When noExpiry is false (default), we set expired=true (control returns after TTL)
			expired := !noExpiry

			if err := c.AcquireModuleChatControl(cmd.Context(), chatID, expired); err != nil {
				return fmt.Errorf("failed to acquire chat control: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"success":  true,
					"chatId":   chatID,
					"action":   "acquired",
					"noExpiry": noExpiry,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if noExpiry {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Chat control acquired for %s (no expiry)\n", chatID)
			} else {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Chat control acquired for %s\n", chatID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&chatID, "chat", "", "Chat ID (userId, roomId, or groupId) (required)")
	cmd.Flags().BoolVar(&noExpiry, "no-expiry", false, "Keep control indefinitely (no TTL)")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}

func newModuleReleaseCmd() *cobra.Command {
	return newModuleReleaseCmdWithClient(nil)
}

func newModuleReleaseCmdWithClient(client *api.Client) *cobra.Command {
	var chatID string

	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release chat control for module",
		Long: `Release chat control for a module channel.

When the module channel has chat control, it can call this API to return
chat control to the Primary Channel. The chatId can be a userId, roomId,
or groupId.`,
		Example: `  # Release chat control for a user
  line module release --chat U1234567890abcdef`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if chatID == "" {
				return fmt.Errorf("--chat is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.ReleaseModuleChatControl(cmd.Context(), chatID); err != nil {
				return fmt.Errorf("failed to release chat control: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"success": true,
					"chatId":  chatID,
					"action":  "released",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Chat control released for %s\n", chatID)
			return nil
		},
	}

	cmd.Flags().StringVar(&chatID, "chat", "", "Chat ID (userId, roomId, or groupId) (required)")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}

func newModuleTokenCmd() *cobra.Command {
	return newModuleTokenCmdWithClient(nil)
}

func newModuleTokenCmdWithClient(client *api.Client) *cobra.Command {
	var code string
	var redirectURI string
	var clientID string
	var clientSecret string

	cmd := &cobra.Command{
		Use:   "token",
		Short: "Exchange authorization code for module access token",
		Long: `Exchange an authorization code for a module access token.

This is used by LINE Official Account Manager integrations to obtain
an access token after the user has authorized the module. The authorization
code is received via the redirect URI during the OAuth flow.

Note: This endpoint does not require an existing access token since
it is used to obtain one.`,
		Example: `  # Exchange authorization code for access token
  line module token --code AUTH_CODE --redirect-uri https://example.com/callback \
    --client-id 1234567890 --client-secret abc123def456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				// Create a minimal client (no auth token needed for this endpoint)
				c = api.NewClient("", flags.Debug, flags.DryRun)
			}

			resp, err := c.ExchangeModuleToken(cmd.Context(), code, redirectURI, clientID, clientSecret)
			if err != nil {
				return fmt.Errorf("failed to exchange token: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Access Token: %s\n", resp.AccessToken)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Token Type:   %s\n", resp.TokenType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Expires In:   %d seconds\n", resp.ExpiresIn)
			if resp.RefreshToken != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Refresh Token: %s\n", resp.RefreshToken)
			}
			if resp.Scope != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Scope:        %s\n", resp.Scope)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&code, "code", "", "Authorization code received from OAuth flow (required)")
	cmd.Flags().StringVar(&redirectURI, "redirect-uri", "", "Redirect URI used in the authorization request (required)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Channel ID (required)")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Channel secret (required)")

	_ = cmd.MarkFlagRequired("code")
	_ = cmd.MarkFlagRequired("redirect-uri")
	_ = cmd.MarkFlagRequired("client-id")
	_ = cmd.MarkFlagRequired("client-secret")

	return cmd
}

func newModuleBotsCmd() *cobra.Command {
	return newModuleBotsCmdWithClient(nil)
}

func newModuleBotsCmdWithClient(client *api.Client) *cobra.Command {
	var limit int
	var start string

	cmd := &cobra.Command{
		Use:   "bots",
		Short: "List bots with attached module channels",
		Long: `List LINE Official Account bots that have module channels attached.

This endpoint is used by LINE Official Account Manager integrations to see
which bots have modules attached to them.`,
		Example: `  # List all bots with modules
  line module bots

  # List with limit
  line module bots --limit 10

  # Paginate through results
  line module bots --start <continuation-token>

  # Output as JSON
  line module bots --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			resp, err := c.GetBotsWithModules(cmd.Context(), limit, start)
			if err != nil {
				return fmt.Errorf("failed to list bots: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			// Text output
			if len(resp.Bots) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No bots with attached modules found.")
				return nil
			}

			for _, bot := range resp.Bots {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", bot.DisplayName)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  User ID:    %s\n", bot.UserID)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Basic ID:   %s\n", bot.BasicID)
				if bot.PremiumID != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Premium ID: %s\n", bot.PremiumID)
				}
				if bot.PictureURL != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Picture:    %s\n", bot.PictureURL)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
			}

			if resp.Next != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "More results available. Use --start %s to continue.\n", resp.Next)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of bots to return (default 100, max 100)")
	cmd.Flags().StringVar(&start, "start", "", "Continuation token for pagination")

	return cmd
}
