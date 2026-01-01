package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newLIFFCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liff",
		Short: "Manage LIFF (LINE Front-end Framework) apps",
		Long:  "Create, list, update, and delete LIFF apps for your channel.",
	}

	cmd.AddCommand(newLIFFListCmd())
	cmd.AddCommand(newLIFFCreateCmd())
	cmd.AddCommand(newLIFFUpdateCmd())
	cmd.AddCommand(newLIFFDeleteCmd())

	return cmd
}

func newLIFFListCmd() *cobra.Command {
	return newLIFFListCmdWithClient(nil)
}

func newLIFFListCmdWithClient(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all LIFF apps",
		Long:  "Get a list of all LIFF apps registered on the channel.",
		Example: `  # List all LIFF apps
  line liff list

  # Output as JSON
  line liff list --output json

  # Output as table
  line liff list --output table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			apps, err := c.GetAllLIFFApps(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list LIFF apps: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"apps": apps}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if len(apps) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No LIFF apps found")
				return nil
			}

			if flags.Output == "table" {
				table := NewTable("LIFF ID", "TYPE", "URL", "DESCRIPTION")
				for _, app := range apps {
					table.AddRow(app.LIFFID, app.View.Type, app.View.URL, app.Description)
				}
				table.Render(cmd.OutOrStdout())
				return nil
			}

			// Default text output
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Found %d LIFF app(s):\n\n", len(apps))
			for _, app := range apps {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "LIFF ID:     %s\n", app.LIFFID)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "View Type:   %s\n", app.View.Type)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "URL:         %s\n", app.View.URL)
				if app.Description != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", app.Description)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
			}

			return nil
		},
	}

	return cmd
}

func newLIFFCreateCmd() *cobra.Command {
	return newLIFFCreateCmdWithClient(nil)
}

func newLIFFCreateCmdWithClient(client *api.Client) *cobra.Command {
	var viewType string
	var url string
	var description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new LIFF app",
		Long: `Create a new LIFF app with the specified view configuration.

View types:
  compact - Opens in the bottom half of the screen (like a modal)
  tall    - Opens in a taller panel (about 80% of the screen)
  full    - Opens in full screen mode`,
		Example: `  # Create a compact LIFF app
  line liff create --type compact --url https://example.com/liff

  # Create a full screen LIFF app with description
  line liff create --type full --url https://example.com/app --description "My LIFF App"

  # Output as JSON
  line liff create --type tall --url https://example.com/liff --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if viewType == "" {
				return fmt.Errorf("--type is required (compact, tall, or full)")
			}
			if viewType != "compact" && viewType != "tall" && viewType != "full" {
				return fmt.Errorf("--type must be one of: compact, tall, full")
			}
			if url == "" {
				return fmt.Errorf("--url is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			req := &api.AddLIFFAppRequest{
				View: api.LIFFView{
					Type: viewType,
					URL:  url,
				},
				Description: description,
			}

			liffID, err := c.AddLIFFApp(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("failed to create LIFF app: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"liffId":      liffID,
					"view":        req.View,
					"description": description,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created LIFF app: %s\n", liffID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "View Type:        %s\n", viewType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "URL:              %s\n", url)
			if description != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description:      %s\n", description)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&viewType, "type", "", "View type: compact, tall, or full (required)")
	cmd.Flags().StringVar(&url, "url", "", "LIFF app URL (required)")
	cmd.Flags().StringVar(&description, "description", "", "LIFF app description (optional)")

	return cmd
}

func newLIFFUpdateCmd() *cobra.Command {
	return newLIFFUpdateCmdWithClient(nil)
}

func newLIFFUpdateCmdWithClient(client *api.Client) *cobra.Command {
	var liffID string
	var viewType string
	var url string
	var description string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a LIFF app",
		Long: `Update an existing LIFF app with new view configuration.

View types:
  compact - Opens in the bottom half of the screen (like a modal)
  tall    - Opens in a taller panel (about 80% of the screen)
  full    - Opens in full screen mode`,
		Example: `  # Update a LIFF app's URL
  line liff update --id 1234567890-abcdefgh --type full --url https://example.com/new-liff

  # Update with a new description
  line liff update --id 1234567890-abcdefgh --type compact --url https://example.com/liff --description "Updated app"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if liffID == "" {
				return fmt.Errorf("--id is required")
			}
			if viewType == "" {
				return fmt.Errorf("--type is required (compact, tall, or full)")
			}
			if viewType != "compact" && viewType != "tall" && viewType != "full" {
				return fmt.Errorf("--type must be one of: compact, tall, full")
			}
			if url == "" {
				return fmt.Errorf("--url is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			req := &api.UpdateLIFFAppRequest{
				View: api.LIFFView{
					Type: viewType,
					URL:  url,
				},
				Description: description,
			}

			if err := c.UpdateLIFFApp(cmd.Context(), liffID, req); err != nil {
				return fmt.Errorf("failed to update LIFF app: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"liffId":      liffID,
					"view":        req.View,
					"description": description,
					"status":      "updated",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated LIFF app: %s\n", liffID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "View Type:        %s\n", viewType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "URL:              %s\n", url)
			if description != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description:      %s\n", description)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&liffID, "id", "", "LIFF app ID (required)")
	cmd.Flags().StringVar(&viewType, "type", "", "View type: compact, tall, or full (required)")
	cmd.Flags().StringVar(&url, "url", "", "LIFF app URL (required)")
	cmd.Flags().StringVar(&description, "description", "", "LIFF app description (optional)")

	return cmd
}

func newLIFFDeleteCmd() *cobra.Command {
	return newLIFFDeleteCmdWithClient(nil)
}

func newLIFFDeleteCmdWithClient(client *api.Client) *cobra.Command {
	var liffID string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a LIFF app",
		Long:  "Delete a LIFF app from the channel. This action cannot be undone.",
		Example: `  # Delete a LIFF app (requires confirmation)
  line liff delete --id 1234567890-abcdefgh --yes

  # Output as JSON
  line liff delete --id 1234567890-abcdefgh --yes --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if liffID == "" {
				return fmt.Errorf("--id is required")
			}

			if !flags.Yes {
				return fmt.Errorf("use --yes to confirm deleting the LIFF app")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.DeleteLIFFApp(cmd.Context(), liffID); err != nil {
				return fmt.Errorf("failed to delete LIFF app: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"liffId": liffID,
					"status": "deleted",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted LIFF app: %s\n", liffID)
			return nil
		},
	}

	cmd.Flags().StringVar(&liffID, "id", "", "LIFF app ID (required)")

	return cmd
}
