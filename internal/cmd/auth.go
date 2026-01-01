package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/auth"
	"github.com/salmonumbrella/line-official-cli/internal/secrets"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Login, logout, and manage LINE Official Account credentials.",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthStatusCmd())
	cmd.AddCommand(newAuthListCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	return newAuthLoginCmdWithStore(nil)
}

func newAuthLoginCmdWithStore(store secrets.Store) *cobra.Command {
	var channelAccessToken string
	var accountName string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login with channel access token",
		Long: `Authenticate with your LINE Official Account.

Opens a browser to enter your channel access token from the LINE Developers Console.
The token will be stored securely in your system keyring.`,
		Example: `  # Interactive login (opens browser)
  line auth login

  # Login with token directly
  line auth login --token YOUR_TOKEN --name my-account`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if store == nil {
				store, err = openSecretsStore()
				if err != nil {
					return fmt.Errorf("failed to open keyring: %w", err)
				}
			}

			if channelAccessToken != "" {
				if accountName == "" {
					accountName = "default"
				}
				err := store.Set(accountName, secrets.Credentials{
					ChannelAccessToken: channelAccessToken,
				}, "") // Empty bot name for direct token login
				if err != nil {
					return fmt.Errorf("failed to save credentials: %w", err)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s\n", accountName)
				return nil
			}

			// Browser flow
			server, err := auth.NewSetupServer(store)
			if err != nil {
				return fmt.Errorf("failed to start auth server: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Opening browser for authentication...")
			result, err := server.Start(cmd.Context())
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully logged in as %s\n", result.AccountName)
			return nil
		},
	}

	cmd.Flags().StringVar(&channelAccessToken, "token", "", "Channel access token")
	cmd.Flags().StringVar(&accountName, "name", "", "Account name")

	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	return newAuthLogoutCmdWithStore(nil)
}

func newAuthLogoutCmdWithStore(store secrets.Store) *cobra.Command {
	var accountName string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		Long:  "Remove the stored channel access token for an account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if accountName == "" {
				accountName = "default"
			}

			var err error
			if store == nil {
				store, err = openSecretsStore()
				if err != nil {
					return fmt.Errorf("failed to open keyring: %w", err)
				}
			}

			if err := store.Delete(accountName); err != nil {
				return fmt.Errorf("failed to remove credentials: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Logged out: %s\n", accountName)
			return nil
		},
	}

	cmd.Flags().StringVar(&accountName, "name", "", "Account name to logout")

	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	return newAuthStatusCmdWithStore(nil)
}

func newAuthStatusCmdWithStore(store secrets.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  "Display which account is currently active and authentication status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if store == nil {
				store, err = openSecretsStore()
				if err != nil {
					return fmt.Errorf("failed to open keyring: %w", err)
				}
			}

			accounts, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list accounts: %w", err)
			}
			if len(accounts) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Not logged in")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Run: line auth login")
				return nil
			}

			// Determine which account would be used
			activeAccount := ""
			source := ""

			if flags.Account != "" {
				activeAccount = flags.Account
				source = "(from --account flag or LINE_ACCOUNT env)"
			} else {
				primary, _ := store.GetPrimary()
				if primary != "" {
					activeAccount = primary
					// Check if it's explicitly primary or just first
					for _, acc := range accounts {
						if acc.Name == primary && acc.IsPrimary {
							source = "(primary)"
							break
						}
					}
					if source == "" {
						source = "(first account)"
					}
				}
			}

			if activeAccount == "" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No active account")
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Active account: %s %s\n", activeAccount, source)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "All accounts:")
			for _, acc := range accounts {
				marker := "  "
				if acc.Name == activeAccount {
					marker = "* "
				}
				primary := ""
				if acc.IsPrimary {
					primary = " (primary)"
				}
				botInfo := ""
				if acc.BotName != "" {
					botInfo = fmt.Sprintf(" - %s", acc.BotName)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s%s\n", marker, acc.Name, botInfo, primary)
			}
			return nil
		},
	}

	return cmd
}

func newAuthListCmd() *cobra.Command {
	return newAuthListCmdWithStore(nil)
}

func newAuthListCmdWithStore(store secrets.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured accounts",
		Long:  "Show all LINE Official Accounts that have been configured.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if store == nil {
				store, err = openSecretsStore()
				if err != nil {
					return fmt.Errorf("failed to open keyring: %w", err)
				}
			}

			accounts, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list accounts: %w", err)
			}

			if len(accounts) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No accounts configured")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Run: line auth login")
				return nil
			}

			if flags.Output == "json" {
				data, err := json.MarshalIndent(accounts, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			if flags.Output == "table" {
				table := NewTable("ACCOUNT", "BOT", "PRIMARY", "CREATED")
				for _, acc := range accounts {
					primary := ""
					if acc.IsPrimary {
						primary = "*"
					}
					created := ""
					if !acc.CreatedAt.IsZero() {
						created = acc.CreatedAt.Format("2006-01-02")
					}
					table.AddRow(acc.Name, acc.BotName, primary, created)
				}
				table.Render(cmd.OutOrStdout())
				return nil
			}

			// Default text output
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Configured accounts:")
			for _, acc := range accounts {
				primary := ""
				if acc.IsPrimary {
					primary = " (primary)"
				}
				botInfo := ""
				if acc.BotName != "" {
					botInfo = fmt.Sprintf(" - %s", acc.BotName)
				}
				created := ""
				if !acc.CreatedAt.IsZero() {
					created = fmt.Sprintf(" [%s]", acc.CreatedAt.Format("2006-01-02"))
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s%s%s%s\n", acc.Name, botInfo, primary, created)
			}
			return nil
		},
	}

	return cmd
}
