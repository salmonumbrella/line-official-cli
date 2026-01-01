package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/salmonumbrella/line-official-cli/internal/config"
	"github.com/spf13/cobra"
)

type rootFlags struct {
	Account string
	Output  string
	Debug   bool
	DryRun  bool // show what would be sent without actually sending
	// Agent-friendly flags
	Yes bool // skip confirmation prompts
}

var flags rootFlags
var cfg *config.Config

func NewRootCmd() *cobra.Command {
	// Load config file (errors are ignored - config is optional)
	cfg, _ = config.Load()
	if cfg == nil {
		cfg = &config.Config{}
	}

	cmd := &cobra.Command{
		Use:   "line",
		Short: "LINE Official Account CLI",
		Long: `A command-line interface for LINE Official Accounts.

Manage messaging, rich menus, audiences, and insights for your
LINE Official Account - built for both humans and AI agents.`,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	// Priority: flags > env vars > config file > defaults
	cmd.PersistentFlags().StringVar(&flags.Account, "account", getDefault(os.Getenv("LINE_ACCOUNT"), cfg.Account, ""), "Account name (or LINE_ACCOUNT env)")
	cmd.PersistentFlags().StringVar(&flags.Output, "output", getDefault(os.Getenv("LINE_OUTPUT"), cfg.Output, "text"), "Output format: text|json|table")
	cmd.PersistentFlags().BoolVar(&flags.Debug, "debug", getDefaultBool(cfg.Debug, false), "Enable debug output")
	cmd.PersistentFlags().BoolVar(&flags.DryRun, "dry-run", false, "Show what would be sent without actually sending")
	cmd.PersistentFlags().BoolVarP(&flags.Yes, "yes", "y", false, "Skip confirmation prompts")

	// Add subcommands
	cmd.AddCommand(newMessageCmd())
	cmd.AddCommand(newRichMenuCmd())
	cmd.AddCommand(newAudienceCmd())
	cmd.AddCommand(newInsightCmd())
	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newBotCmd())
	cmd.AddCommand(newWebhookCmd())
	cmd.AddCommand(newContentCmd())
	cmd.AddCommand(newGroupCmd())
	cmd.AddCommand(newRoomCmd())
	cmd.AddCommand(newMembershipCmd())
	cmd.AddCommand(newCouponCmd())
	cmd.AddCommand(newTokenCmd())
	cmd.AddCommand(newChatCmd())
	cmd.AddCommand(newLIFFCmd())
	cmd.AddCommand(newModuleCmd())
	cmd.AddCommand(newShopCmd())
	cmd.AddCommand(newPNPCmd())
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newCompletionCmd())
	cmd.AddCommand(newConfigCmd())

	return cmd
}

// getDefault returns the first non-empty string, or the fallback.
func getDefault(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// getDefaultBool returns cfgVal if true, otherwise returns fallback.
// This is because we cannot distinguish "not set" from "set to false" in bool.
func getDefaultBool(cfgVal, fallback bool) bool {
	if cfgVal {
		return true
	}
	return fallback
}

func requireAccount(f *rootFlags) (string, error) {
	// 1. Check explicit flag (already includes env var from flag default)
	if f.Account != "" {
		return f.Account, nil
	}

	store, err := openSecretsStore()
	if err != nil {
		return "", fmt.Errorf("failed to access keyring: %w. Use --account or set LINE_ACCOUNT", err)
	}

	// 2. Check for primary account (includes fallback to first account)
	primary, err := store.GetPrimary()
	if err == nil && primary != "" {
		return primary, nil
	}

	// 3. No accounts configured
	return "", fmt.Errorf("no accounts configured. Run: line auth login")
}

func Execute(args []string) error {
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	return cmd.Execute()
}

func ExecuteContext(ctx context.Context, args []string) error {
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}
