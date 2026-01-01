package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage channel access tokens",
		Long:  "Issue, verify, and revoke channel access tokens for LINE OAuth.",
	}

	cmd.AddCommand(newTokenIssueCmd())
	cmd.AddCommand(newTokenVerifyCmd())
	cmd.AddCommand(newTokenRevokeCmd())
	cmd.AddCommand(newTokenIssueJWTCmd())
	cmd.AddCommand(newTokenVerifyJWTCmd())
	cmd.AddCommand(newTokenRevokeJWTCmd())
	cmd.AddCommand(newTokenListKeysCmd())
	cmd.AddCommand(newTokenIssueStatelessCmd())

	return cmd
}

func newTokenIssueCmd() *cobra.Command {
	return newTokenIssueCmdWithClient(nil)
}

func newTokenIssueCmdWithClient(client *api.Client) *cobra.Command {
	var clientID string
	var clientSecret string

	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue a v2 channel access token",
		Long:  "Issue a short-lived channel access token using client credentials (v2 API).",
		Example: `  # Issue a new v2 channel access token
  line token issue --client-id 1234567890 --client-secret abc123

  # Output as JSON
  line token issue --client-id 1234567890 --client-secret abc123 --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientID == "" {
				return fmt.Errorf("--client-id is required")
			}
			if clientSecret == "" {
				return fmt.Errorf("--client-secret is required")
			}

			c := client
			if c == nil {
				// Create a client without auth (token endpoints don't use Bearer auth)
				c = api.NewClient("", flags.Debug, flags.DryRun)
			}

			resp, err := c.IssueChannelToken(cmd.Context(), clientID, clientSecret)
			if err != nil {
				return fmt.Errorf("failed to issue token: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Access Token: %s\n", resp.AccessToken)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Token Type:   %s\n", resp.TokenType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Expires In:   %d seconds\n", resp.ExpiresIn)
			if resp.KeyID != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Key ID:       %s\n", resp.KeyID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", "", "Channel ID (required)")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Channel secret (required)")

	return cmd
}

func newTokenVerifyCmd() *cobra.Command {
	return newTokenVerifyCmdWithClient(nil)
}

func newTokenVerifyCmdWithClient(client *api.Client) *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify a v2 channel access token",
		Long:  "Verify a channel access token and get its information (v2 API).",
		Example: `  # Verify a v2 channel access token
  line token verify --token eyJhbGciOiJ...

  # Output as JSON
  line token verify --token eyJhbGciOiJ... --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				return fmt.Errorf("--token is required")
			}

			c := client
			if c == nil {
				c = api.NewClient("", flags.Debug, flags.DryRun)
			}

			info, err := c.VerifyChannelToken(cmd.Context(), token)
			if err != nil {
				return fmt.Errorf("failed to verify token: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Client ID:  %s\n", info.ClientID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Expires In: %d seconds\n", info.ExpiresIn)
			if info.Scope != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Scope:      %s\n", info.Scope)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Access token to verify (required)")

	return cmd
}

func newTokenRevokeCmd() *cobra.Command {
	return newTokenRevokeCmdWithClient(nil)
}

func newTokenRevokeCmdWithClient(client *api.Client) *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a v2 channel access token",
		Long:  "Revoke a channel access token (v2 API).",
		Example: `  # Revoke a v2 channel access token
  line token revoke --token eyJhbGciOiJ...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				return fmt.Errorf("--token is required")
			}

			c := client
			if c == nil {
				c = api.NewClient("", flags.Debug, flags.DryRun)
			}

			if err := c.RevokeChannelToken(cmd.Context(), token); err != nil {
				return fmt.Errorf("failed to revoke token: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"status": "revoked",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Token revoked successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Access token to revoke (required)")

	return cmd
}

func newTokenIssueJWTCmd() *cobra.Command {
	return newTokenIssueJWTCmdWithClient(nil)
}

func newTokenIssueJWTCmdWithClient(client *api.Client) *cobra.Command {
	var jwt string

	cmd := &cobra.Command{
		Use:   "issue-jwt",
		Short: "Issue a v2.1 channel access token using JWT",
		Long:  "Issue a channel access token using JWT assertion (v2.1 API).",
		Example: `  # Issue a v2.1 channel access token using JWT
  line token issue-jwt --jwt eyJhbGciOiJSUzI1NiI...

  # Output as JSON
  line token issue-jwt --jwt eyJhbGciOiJSUzI1NiI... --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jwt == "" {
				return fmt.Errorf("--jwt is required")
			}

			c := client
			if c == nil {
				c = api.NewClient("", flags.Debug, flags.DryRun)
			}

			resp, err := c.IssueChannelTokenByJWT(cmd.Context(), jwt)
			if err != nil {
				return fmt.Errorf("failed to issue token: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Access Token: %s\n", resp.AccessToken)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Token Type:   %s\n", resp.TokenType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Expires In:   %d seconds\n", resp.ExpiresIn)
			if resp.KeyID != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Key ID:       %s\n", resp.KeyID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&jwt, "jwt", "", "JWT assertion (required)")

	return cmd
}

func newTokenVerifyJWTCmd() *cobra.Command {
	return newTokenVerifyJWTCmdWithClient(nil)
}

func newTokenVerifyJWTCmdWithClient(client *api.Client) *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "verify-jwt",
		Short: "Verify a v2.1 channel access token",
		Long:  "Verify a v2.1 channel access token and get its information.",
		Example: `  # Verify a v2.1 channel access token
  line token verify-jwt --token eyJhbGciOiJ...

  # Output as JSON
  line token verify-jwt --token eyJhbGciOiJ... --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				return fmt.Errorf("--token is required")
			}

			c := client
			if c == nil {
				c = api.NewClient("", flags.Debug, flags.DryRun)
			}

			info, err := c.VerifyChannelTokenByJWT(cmd.Context(), token)
			if err != nil {
				return fmt.Errorf("failed to verify token: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Client ID:  %s\n", info.ClientID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Expires In: %d seconds\n", info.ExpiresIn)
			if info.Scope != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Scope:      %s\n", info.Scope)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Access token to verify (required)")

	return cmd
}

func newTokenRevokeJWTCmd() *cobra.Command {
	return newTokenRevokeJWTCmdWithClient(nil)
}

func newTokenRevokeJWTCmdWithClient(client *api.Client) *cobra.Command {
	var token string
	var clientID string
	var clientSecret string

	cmd := &cobra.Command{
		Use:   "revoke-jwt",
		Short: "Revoke a v2.1 channel access token",
		Long:  "Revoke a v2.1 channel access token.",
		Example: `  # Revoke a v2.1 channel access token
  line token revoke-jwt --token eyJhbGciOiJ... --client-id 1234567890 --client-secret abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				return fmt.Errorf("--token is required")
			}
			if clientID == "" {
				return fmt.Errorf("--client-id is required")
			}
			if clientSecret == "" {
				return fmt.Errorf("--client-secret is required")
			}

			c := client
			if c == nil {
				c = api.NewClient("", flags.Debug, flags.DryRun)
			}

			if err := c.RevokeChannelTokenByJWT(cmd.Context(), token, clientID, clientSecret); err != nil {
				return fmt.Errorf("failed to revoke token: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"status": "revoked",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Token revoked successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Access token to revoke (required)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Channel ID (required)")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Channel secret (required)")

	return cmd
}

func newTokenListKeysCmd() *cobra.Command {
	return newTokenListKeysCmdWithClient(nil)
}

func newTokenListKeysCmdWithClient(client *api.Client) *cobra.Command {
	var jwt string

	cmd := &cobra.Command{
		Use:   "list-keys",
		Short: "List all valid token key IDs",
		Long:  "Get all valid channel access token key IDs (v2.1 API).",
		Example: `  # List all valid token key IDs
  line token list-keys --jwt eyJhbGciOiJSUzI1NiI...

  # Output as JSON
  line token list-keys --jwt eyJhbGciOiJSUzI1NiI... --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jwt == "" {
				return fmt.Errorf("--jwt is required")
			}

			c := client
			if c == nil {
				c = api.NewClient("", flags.Debug, flags.DryRun)
			}

			kids, err := c.GetAllValidTokenKeyIDs(cmd.Context(), jwt)
			if err != nil {
				return fmt.Errorf("failed to list key IDs: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"kids": kids,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if len(kids) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No valid token key IDs found")
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Valid Token Key IDs:")
			for _, kid := range kids {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", kid)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&jwt, "jwt", "", "JWT assertion (required)")

	return cmd
}

func newTokenIssueStatelessCmd() *cobra.Command {
	return newTokenIssueStatelessCmdWithClient(nil)
}

func newTokenIssueStatelessCmdWithClient(client *api.Client) *cobra.Command {
	var clientID string
	var clientSecret string

	cmd := &cobra.Command{
		Use:   "issue-stateless",
		Short: "Issue a stateless v3 channel access token",
		Long: `Issue a stateless channel access token using client credentials (v3 API).

WARNING: Stateless tokens cannot be revoked and expire in 15 minutes.
They are suitable for short-lived operations where revocation is not needed.`,
		Example: `  # Issue a stateless v3 channel access token
  line token issue-stateless --client-id 1234567890 --client-secret abc123

  # Output as JSON
  line token issue-stateless --client-id 1234567890 --client-secret abc123 --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientID == "" {
				return fmt.Errorf("--client-id is required")
			}
			if clientSecret == "" {
				return fmt.Errorf("--client-secret is required")
			}

			c := client
			if c == nil {
				// Create a client without auth (token endpoints don't use Bearer auth)
				c = api.NewClient("", flags.Debug, flags.DryRun)
			}

			// Warn about stateless token limitations
			if flags.Output != "json" {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Note: Stateless tokens cannot be revoked and expire in 15 minutes.")
			}

			resp, err := c.IssueStatelessToken(cmd.Context(), clientID, clientSecret)
			if err != nil {
				return fmt.Errorf("failed to issue stateless token: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Access Token: %s\n", resp.AccessToken)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Token Type:   %s\n", resp.TokenType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Expires In:   %d seconds\n", resp.ExpiresIn)
			if resp.KeyID != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Key ID:       %s\n", resp.KeyID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", "", "Channel ID (required)")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Channel secret (required)")

	return cmd
}
