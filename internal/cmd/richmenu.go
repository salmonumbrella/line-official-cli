package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newRichMenuCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "richmenu",
		Aliases: []string{"rm"},
		Short:   "Manage rich menus",
		Long:    "Create, list, and manage rich menus for your LINE Official Account.",
	}

	cmd.AddCommand(newRichMenuListCmd())
	cmd.AddCommand(newRichMenuCreateCmd())
	cmd.AddCommand(newRichMenuDeleteCmd())
	cmd.AddCommand(newRichMenuSetDefaultCmd())
	cmd.AddCommand(newRichMenuCancelDefaultCmd())
	cmd.AddCommand(newRichMenuUploadImageCmd())
	cmd.AddCommand(newRichMenuGetCmd())
	cmd.AddCommand(newRichMenuLinkCmd())
	cmd.AddCommand(newRichMenuUnlinkCmd())
	cmd.AddCommand(newRichMenuAliasCmd())
	cmd.AddCommand(newRichMenuBulkCmd())
	cmd.AddCommand(newRichMenuBatchCmd())
	cmd.AddCommand(newRichMenuValidateCmd())
	cmd.AddCommand(newRichMenuDownloadImageCmd())

	return cmd
}

func newRichMenuListCmd() *cobra.Command {
	return newRichMenuListCmdWithClient(nil)
}

func newRichMenuListCmdWithClient(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all rich menus",
		Long:  "Get a list of all rich menus associated with your LINE Official Account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}
			return listRichMenusWithClient(cmd, c)
		},
	}

	return cmd
}

func newRichMenuCreateCmd() *cobra.Command {
	return newRichMenuCreateCmdWithClient(nil)
}

func newRichMenuCreateCmdWithClient(client *api.Client) *cobra.Command {
	var chatBarText string
	var actionsJSON string
	var size string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new rich menu",
		Long:  "Create a rich menu with the specified actions and chat bar text.",
		Example: `  # Create a full-size rich menu
  line richmenu create --name "Main Menu" --actions '[{"type":"message","label":"Help","text":"help"}]'

  # Create a compact rich menu
  line richmenu create --name "Menu" --size compact --actions '[...]'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if chatBarText == "" {
				return fmt.Errorf("--name is required")
			}
			if actionsJSON == "" {
				return fmt.Errorf("--actions is required")
			}
			if size != "full" && size != "compact" {
				return fmt.Errorf("--size must be 'full' or 'compact'")
			}

			// Parse actions JSON
			var actions []json.RawMessage
			if err := json.Unmarshal([]byte(actionsJSON), &actions); err != nil {
				return fmt.Errorf("invalid actions JSON: %w", err)
			}

			// Determine dimensions based on size
			height := 1686
			if size == "compact" {
				height = 843
			}

			// Build areas from actions (simplified: single area covering the whole menu)
			areas := make([]api.RichMenuArea, len(actions))
			areaWidth := 2500 / len(actions)
			for i, action := range actions {
				areas[i] = api.RichMenuArea{
					Bounds: api.RichMenuBounds{
						X:      i * areaWidth,
						Y:      0,
						Width:  areaWidth,
						Height: height,
					},
					Action: action,
				}
			}

			req := api.CreateRichMenuRequest{
				Size: api.RichMenuSize{
					Width:  2500,
					Height: height,
				},
				Selected:    false,
				Name:        chatBarText,
				ChatBarText: chatBarText,
				Areas:       areas,
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			richMenuID, err := c.CreateRichMenu(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("failed to create rich menu: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"richMenuId": richMenuID,
					"name":       chatBarText,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created rich menu: %s (ID: %s)\n", chatBarText, richMenuID)
			return nil
		},
	}

	cmd.Flags().StringVar(&chatBarText, "name", "", "Chat bar text / menu name (required)")
	cmd.Flags().StringVar(&actionsJSON, "actions", "", "Actions JSON array (required)")
	cmd.Flags().StringVar(&size, "size", "full", "Menu size: full (2500x1686) or compact (2500x843)")

	return cmd
}

func newRichMenuDeleteCmd() *cobra.Command {
	return newRichMenuDeleteCmdWithClient(nil)
}

func newRichMenuDeleteCmdWithClient(client *api.Client) *cobra.Command {
	var richMenuID string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a rich menu",
		Long:  "Delete a rich menu by its ID.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if richMenuID == "" {
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

			if err := c.DeleteRichMenu(cmd.Context(), richMenuID); err != nil {
				return fmt.Errorf("failed to delete rich menu: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"deleted": richMenuID}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted rich menu: %s\n", richMenuID)
			return nil
		},
	}

	cmd.Flags().StringVar(&richMenuID, "id", "", "Rich menu ID to delete (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newRichMenuSetDefaultCmd() *cobra.Command {
	return newRichMenuSetDefaultCmdWithClient(nil)
}

func newRichMenuSetDefaultCmdWithClient(client *api.Client) *cobra.Command {
	var richMenuID string

	cmd := &cobra.Command{
		Use:   "set-default",
		Short: "Set the default rich menu",
		Long:  "Set a rich menu as the default for all users.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if richMenuID == "" {
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

			if err := c.SetDefaultRichMenu(cmd.Context(), richMenuID); err != nil {
				return fmt.Errorf("failed to set default rich menu: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"defaultRichMenuId": richMenuID}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Set default rich menu: %s\n", richMenuID)
			return nil
		},
	}

	cmd.Flags().StringVar(&richMenuID, "id", "", "Rich menu ID to set as default (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newRichMenuCancelDefaultCmd() *cobra.Command {
	return newRichMenuCancelDefaultCmdWithClient(nil)
}

func newRichMenuCancelDefaultCmdWithClient(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-default",
		Short: "Cancel the default rich menu",
		Long:  "Remove the default rich menu setting.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.CancelDefaultRichMenu(cmd.Context()); err != nil {
				return fmt.Errorf("failed to cancel default rich menu: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"defaultRichMenuId": nil}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Cancelled default rich menu")
			return nil
		},
	}

	return cmd
}

func listRichMenusWithClient(cmd *cobra.Command, client *api.Client) error {
	menus, err := client.GetRichMenuList(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to list rich menus: %w", err)
	}

	// Get default rich menu to mark it
	defaultID, _ := client.GetDefaultRichMenuID(cmd.Context())

	if flags.Output == "json" {
		result := map[string]any{
			"richmenus":       menus,
			"defaultRichMenu": defaultID,
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	if len(menus) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No rich menus found")
		return nil
	}

	if flags.Output == "table" {
		table := NewTable("ID", "NAME", "SIZE", "DEFAULT")
		for _, menu := range menus {
			size := fmt.Sprintf("%dx%d", menu.Size.Width, menu.Size.Height)
			isDefault := ""
			if menu.RichMenuID == defaultID {
				isDefault = "yes"
			}
			table.AddRow(menu.RichMenuID, menu.ChatBarText, size, isDefault)
		}
		table.Render(cmd.OutOrStdout())
		return nil
	}

	// Default text output
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Rich Menus:")
	for _, menu := range menus {
		prefix := "  "
		suffix := ""
		if menu.RichMenuID == defaultID {
			prefix = "* "
			suffix = " (default)"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s%s  %s%s\n", prefix, menu.RichMenuID, menu.ChatBarText, suffix)
	}
	return nil
}

func newRichMenuUploadImageCmd() *cobra.Command {
	return newRichMenuUploadImageCmdWithClient(nil, nil)
}

func newRichMenuUploadImageCmdWithClient(client *api.Client, imageDataOverride []byte) *cobra.Command {
	var richMenuID string
	var imagePath string

	cmd := &cobra.Command{
		Use:   "upload-image",
		Short: "Upload an image for a rich menu",
		Long: `Upload an image file for a rich menu. The image must be:
- PNG or JPEG format
- 2500x1686 pixels (full) or 2500x843 pixels (compact)
- Maximum 1MB file size`,
		Example: `  # Upload an image to a rich menu
  line richmenu upload-image --id richmenu-xxx --image menu.png`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if richMenuID == "" {
				return fmt.Errorf("--id is required")
			}

			var data []byte
			var contentType string

			if imageDataOverride != nil {
				// Use override data (for testing)
				data = imageDataOverride
				contentType = "image/png"
			} else {
				if imagePath == "" {
					return fmt.Errorf("--image is required")
				}

				// Read image file
				var err error
				data, err = os.ReadFile(imagePath)
				if err != nil {
					return fmt.Errorf("failed to read image: %w", err)
				}

				// Check file size (max 1MB)
				if len(data) > 1024*1024 {
					return fmt.Errorf("image file too large: max 1MB, got %d bytes", len(data))
				}

				// Determine content type
				contentType = "image/png"
				ext := strings.ToLower(filepath.Ext(imagePath))
				if ext == ".jpg" || ext == ".jpeg" {
					contentType = "image/jpeg"
				} else if ext != ".png" {
					return fmt.Errorf("unsupported image format: use PNG or JPEG")
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

			if err := c.UploadRichMenuImage(cmd.Context(), richMenuID, contentType, data); err != nil {
				return fmt.Errorf("failed to upload image: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"richMenuId": richMenuID, "status": "uploaded"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Image uploaded to rich menu: %s\n", richMenuID)
			return nil
		},
	}

	cmd.Flags().StringVar(&richMenuID, "id", "", "Rich menu ID (required)")
	cmd.Flags().StringVar(&imagePath, "image", "", "Path to image file (required)")
	_ = cmd.MarkFlagRequired("id")
	// Note: --image is not marked required since imageDataOverride can be used in tests

	return cmd
}

func newRichMenuGetCmd() *cobra.Command {
	return newRichMenuGetCmdWithClient(nil)
}

func newRichMenuGetCmdWithClient(client *api.Client) *cobra.Command {
	var richMenuID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get rich menu details",
		Long:  "Get detailed information about a specific rich menu.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if richMenuID == "" {
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

			menu, err := c.GetRichMenu(cmd.Context(), richMenuID)
			if err != nil {
				return fmt.Errorf("failed to get rich menu: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(menu)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:       %s\n", menu.RichMenuID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:     %s\n", menu.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Size:     %dx%d\n", menu.Size.Width, menu.Size.Height)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Areas:    %d\n", len(menu.Areas))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Selected: %v\n", menu.Selected)
			return nil
		},
	}

	cmd.Flags().StringVar(&richMenuID, "id", "", "Rich menu ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newRichMenuLinkCmd() *cobra.Command {
	return newRichMenuLinkCmdWithClient(nil)
}

func newRichMenuLinkCmdWithClient(client *api.Client) *cobra.Command {
	var userID string
	var richMenuID string

	cmd := &cobra.Command{
		Use:     "link",
		Short:   "Link rich menu to a user",
		Long:    "Assign a specific rich menu to a user (overrides default).",
		Example: `  line richmenu link --user U123... --id richmenu-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if userID == "" {
				return fmt.Errorf("--user is required")
			}
			if richMenuID == "" {
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

			if err := c.LinkRichMenuToUser(cmd.Context(), userID, richMenuID); err != nil {
				return fmt.Errorf("failed to link rich menu: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"userId": userID, "richMenuId": richMenuID, "status": "linked"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Linked rich menu %s to user %s\n", richMenuID, userID)
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	cmd.Flags().StringVar(&richMenuID, "id", "", "Rich menu ID (required)")
	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newRichMenuUnlinkCmd() *cobra.Command {
	return newRichMenuUnlinkCmdWithClient(nil)
}

func newRichMenuUnlinkCmdWithClient(client *api.Client) *cobra.Command {
	var userID string

	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "Unlink rich menu from a user",
		Long:  "Remove the user-specific rich menu (reverts to default).",
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

			if err := c.UnlinkRichMenuFromUser(cmd.Context(), userID); err != nil {
				return fmt.Errorf("failed to unlink rich menu: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"userId": userID, "status": "unlinked"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Unlinked rich menu from user %s\n", userID)
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user")

	return cmd
}

func newRichMenuAliasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage rich menu aliases",
		Long:  "Create, update, and delete human-readable aliases for rich menus.",
	}

	cmd.AddCommand(newRichMenuAliasCreateCmd())
	cmd.AddCommand(newRichMenuAliasGetCmd())
	cmd.AddCommand(newRichMenuAliasUpdateCmd())
	cmd.AddCommand(newRichMenuAliasDeleteCmd())
	cmd.AddCommand(newRichMenuAliasListCmd())
	return cmd
}

func newRichMenuAliasCreateCmd() *cobra.Command {
	return newRichMenuAliasCreateCmdWithClient(nil)
}

func newRichMenuAliasCreateCmdWithClient(client *api.Client) *cobra.Command {
	var aliasID string
	var richMenuID string

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a rich menu alias",
		Long:    "Create a human-readable alias for a rich menu.",
		Example: `  line richmenu alias create --alias main-menu --id richmenu-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if aliasID == "" {
				return fmt.Errorf("--alias is required")
			}
			if richMenuID == "" {
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

			if err := c.CreateRichMenuAlias(cmd.Context(), aliasID, richMenuID); err != nil {
				return fmt.Errorf("failed to create alias: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"aliasId": aliasID, "richMenuId": richMenuID, "status": "created"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created alias '%s' -> %s\n", aliasID, richMenuID)
			return nil
		},
	}

	cmd.Flags().StringVar(&aliasID, "alias", "", "Alias ID/name (required)")
	cmd.Flags().StringVar(&richMenuID, "id", "", "Rich menu ID (required)")
	_ = cmd.MarkFlagRequired("alias")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newRichMenuAliasGetCmd() *cobra.Command {
	return newRichMenuAliasGetCmdWithClient(nil)
}

func newRichMenuAliasGetCmdWithClient(client *api.Client) *cobra.Command {
	var aliasID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get rich menu alias info",
		Long:  "Get the rich menu ID associated with an alias.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if aliasID == "" {
				return fmt.Errorf("--alias is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			alias, err := c.GetRichMenuAlias(cmd.Context(), aliasID)
			if err != nil {
				return fmt.Errorf("failed to get alias: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(alias)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Alias:      %s\n", alias.RichMenuAliasID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Rich Menu:  %s\n", alias.RichMenuID)
			return nil
		},
	}

	cmd.Flags().StringVar(&aliasID, "alias", "", "Alias ID/name (required)")
	_ = cmd.MarkFlagRequired("alias")

	return cmd
}

func newRichMenuAliasUpdateCmd() *cobra.Command {
	return newRichMenuAliasUpdateCmdWithClient(nil)
}

func newRichMenuAliasUpdateCmdWithClient(client *api.Client) *cobra.Command {
	var aliasID string
	var richMenuID string

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update a rich menu alias",
		Long:    "Change which rich menu an alias points to.",
		Example: `  line richmenu alias update --alias main-menu --id richmenu-yyy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if aliasID == "" {
				return fmt.Errorf("--alias is required")
			}
			if richMenuID == "" {
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

			if err := c.UpdateRichMenuAlias(cmd.Context(), aliasID, richMenuID); err != nil {
				return fmt.Errorf("failed to update alias: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"aliasId": aliasID, "richMenuId": richMenuID, "status": "updated"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated alias '%s' -> %s\n", aliasID, richMenuID)
			return nil
		},
	}

	cmd.Flags().StringVar(&aliasID, "alias", "", "Alias ID/name (required)")
	cmd.Flags().StringVar(&richMenuID, "id", "", "New rich menu ID (required)")
	_ = cmd.MarkFlagRequired("alias")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newRichMenuAliasDeleteCmd() *cobra.Command {
	return newRichMenuAliasDeleteCmdWithClient(nil)
}

func newRichMenuAliasDeleteCmdWithClient(client *api.Client) *cobra.Command {
	var aliasID string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a rich menu alias",
		Long:  "Remove a rich menu alias.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if aliasID == "" {
				return fmt.Errorf("--alias is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.DeleteRichMenuAlias(cmd.Context(), aliasID); err != nil {
				return fmt.Errorf("failed to delete alias: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"aliasId": aliasID, "status": "deleted"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted alias: %s\n", aliasID)
			return nil
		},
	}

	cmd.Flags().StringVar(&aliasID, "alias", "", "Alias ID/name (required)")
	_ = cmd.MarkFlagRequired("alias")

	return cmd
}

func newRichMenuAliasListCmd() *cobra.Command {
	return newRichMenuAliasListCmdWithClient(nil)
}

func newRichMenuAliasListCmdWithClient(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all rich menu aliases",
		Long:  "Get a list of all rich menu aliases.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			aliases, err := c.ListRichMenuAliases(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list aliases: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{"aliases": aliases})
			}

			if len(aliases) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No aliases found")
				return nil
			}

			if flags.Output == "table" {
				table := NewTable("ALIAS", "RICH MENU ID")
				for _, alias := range aliases {
					table.AddRow(alias.RichMenuAliasID, alias.RichMenuID)
				}
				table.Render(cmd.OutOrStdout())
				return nil
			}

			// Default text output
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Rich Menu Aliases:")
			for _, alias := range aliases {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s -> %s\n", alias.RichMenuAliasID, alias.RichMenuID)
			}
			return nil
		},
	}

	return cmd
}

// Bulk operations commands

func newRichMenuBulkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk rich menu operations",
		Long:  "Link or unlink rich menus to/from multiple users at once.",
	}

	cmd.AddCommand(newRichMenuBulkLinkCmd())
	cmd.AddCommand(newRichMenuBulkUnlinkCmd())

	return cmd
}

func newRichMenuBulkLinkCmd() *cobra.Command {
	return newRichMenuBulkLinkCmdWithClient(nil, nil)
}

func newRichMenuBulkLinkCmdWithClient(client *api.Client, userIDsOverride []string) *cobra.Command {
	var richMenuID string
	var usersFile string

	cmd := &cobra.Command{
		Use:   "link",
		Short: "Link rich menu to multiple users",
		Long:  "Link a rich menu to multiple users at once. User IDs are read from a file (one per line).",
		Example: `  # Link a menu to users from a file
  line richmenu bulk link --menu richmenu-xxx --users users.txt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if richMenuID == "" {
				return fmt.Errorf("--menu is required")
			}

			var userIDs []string
			if userIDsOverride != nil {
				userIDs = userIDsOverride
			} else {
				if usersFile == "" {
					return fmt.Errorf("--users is required")
				}
				var err error
				userIDs, err = readUserIDsFromFile(usersFile)
				if err != nil {
					return fmt.Errorf("failed to read users file: %w", err)
				}
			}

			if len(userIDs) == 0 {
				return fmt.Errorf("no user IDs found in file")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.LinkRichMenuToUsers(cmd.Context(), richMenuID, userIDs); err != nil {
				return fmt.Errorf("failed to bulk link: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"richMenuId": richMenuID,
					"userCount":  len(userIDs),
					"status":     "linked",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Linked rich menu %s to %d users\n", richMenuID, len(userIDs))
			return nil
		},
	}

	cmd.Flags().StringVar(&richMenuID, "menu", "", "Rich menu ID (required)")
	cmd.Flags().StringVar(&usersFile, "users", "", "File containing user IDs, one per line (required)")
	_ = cmd.MarkFlagRequired("menu")
	// Note: --users is not marked required since userIDsOverride can be used in tests

	return cmd
}

func newRichMenuBulkUnlinkCmd() *cobra.Command {
	return newRichMenuBulkUnlinkCmdWithClient(nil, nil)
}

func newRichMenuBulkUnlinkCmdWithClient(client *api.Client, userIDsOverride []string) *cobra.Command {
	var usersFile string

	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "Unlink rich menus from multiple users",
		Long:  "Unlink rich menus from multiple users at once. User IDs are read from a file (one per line).",
		Example: `  # Unlink menus from users in a file
  line richmenu bulk unlink --users users.txt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var userIDs []string
			if userIDsOverride != nil {
				userIDs = userIDsOverride
			} else {
				if usersFile == "" {
					return fmt.Errorf("--users is required")
				}
				var err error
				userIDs, err = readUserIDsFromFile(usersFile)
				if err != nil {
					return fmt.Errorf("failed to read users file: %w", err)
				}
			}

			if len(userIDs) == 0 {
				return fmt.Errorf("no user IDs found in file")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.UnlinkRichMenuFromUsers(cmd.Context(), userIDs); err != nil {
				return fmt.Errorf("failed to bulk unlink: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"userCount": len(userIDs),
					"status":    "unlinked",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Unlinked rich menus from %d users\n", len(userIDs))
			return nil
		},
	}

	cmd.Flags().StringVar(&usersFile, "users", "", "File containing user IDs, one per line (required)")
	// Note: --users is not marked required since userIDsOverride can be used in tests

	return cmd
}

// readUserIDsFromFile reads user IDs from a file, one per line
func readUserIDsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var userIDs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			userIDs = append(userIDs, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return userIDs, nil
}

// Batch operations commands

func newRichMenuBatchCmd() *cobra.Command {
	return newRichMenuBatchCmdWithClient(nil, nil)
}

func newRichMenuBatchCmdWithClient(client *api.Client, operationsOverride []api.RichMenuBatchOperation) *cobra.Command {
	var operationsFile string
	var resumeRequestID string

	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch rich menu operations",
		Long: `Execute atomic batch operations from a JSON file. The file should contain an array of operations:
[
  {"type": "link", "richMenuId": "richmenu-xxx", "userIds": ["U1", "U2"]},
  {"type": "unlink", "userIds": ["U3", "U4"]}
]`,
		Example: `  # Execute batch operations from a file
  line richmenu batch --operations ops.json

  # Resume a failed batch
  line richmenu batch --operations ops.json --resume abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var operations []api.RichMenuBatchOperation
			if operationsOverride != nil {
				operations = operationsOverride
			} else {
				if operationsFile == "" {
					// No operations file provided, show help (may have subcommands)
					return cmd.Help()
				}

				var err error
				operations, err = readBatchOperationsFromFile(operationsFile)
				if err != nil {
					return fmt.Errorf("failed to read operations file: %w", err)
				}
			}

			if len(operations) == 0 {
				return fmt.Errorf("no operations found in file")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			requestID, err := c.RichMenuBatch(cmd.Context(), operations, resumeRequestID)
			if err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"requestId":      requestID,
					"operationCount": len(operations),
					"status":         "submitted",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Batch submitted: %s (%d operations)\n", requestID, len(operations))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Check progress with: line richmenu batch status --request %s\n", requestID)
			return nil
		},
	}

	cmd.Flags().StringVar(&operationsFile, "operations", "", "JSON file containing batch operations")
	cmd.Flags().StringVar(&resumeRequestID, "resume", "", "Resume a previous batch request")

	cmd.AddCommand(newRichMenuBatchValidateCmd())
	cmd.AddCommand(newRichMenuBatchStatusCmd())

	return cmd
}

func newRichMenuBatchValidateCmd() *cobra.Command {
	return newRichMenuBatchValidateCmdWithClient(nil, nil)
}

func newRichMenuBatchValidateCmdWithClient(client *api.Client, operationsOverride []api.RichMenuBatchOperation) *cobra.Command {
	var operationsFile string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate batch operations",
		Long:  "Validate batch operations without executing them.",
		Example: `  # Validate batch operations
  line richmenu batch validate --operations ops.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var operations []api.RichMenuBatchOperation
			if operationsOverride != nil {
				operations = operationsOverride
			} else {
				if operationsFile == "" {
					return fmt.Errorf("--operations is required")
				}

				var err error
				operations, err = readBatchOperationsFromFile(operationsFile)
				if err != nil {
					return fmt.Errorf("failed to read operations file: %w", err)
				}
			}

			if len(operations) == 0 {
				return fmt.Errorf("no operations found in file")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.ValidateRichMenuBatch(cmd.Context(), operations); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"operationCount": len(operations),
					"valid":          true,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Batch operations valid (%d operations)\n", len(operations))
			return nil
		},
	}

	cmd.Flags().StringVar(&operationsFile, "operations", "", "JSON file containing batch operations (required)")
	// Note: --operations is not marked required since operationsOverride can be used in tests

	return cmd
}

func newRichMenuBatchStatusCmd() *cobra.Command {
	return newRichMenuBatchStatusCmdWithClient(nil)
}

func newRichMenuBatchStatusCmdWithClient(client *api.Client) *cobra.Command {
	var requestID string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get batch operation status",
		Long:  "Get the progress of a batch operation.",
		Example: `  # Check batch status
  line richmenu batch status --request abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if requestID == "" {
				return fmt.Errorf("--request is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			progress, err := c.GetRichMenuBatchProgress(cmd.Context(), requestID)
			if err != nil {
				return fmt.Errorf("failed to get batch status: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"requestId":     requestID,
					"phase":         progress.Phase,
					"acceptedTime":  progress.AcceptedTime,
					"completedTime": progress.CompletedTime,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Request ID:     %s\n", requestID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Phase:          %s\n", progress.Phase)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Accepted Time:  %s\n", progress.AcceptedTime)
			if progress.CompletedTime != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Completed Time: %s\n", progress.CompletedTime)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&requestID, "request", "", "Batch request ID (required)")
	_ = cmd.MarkFlagRequired("request")

	return cmd
}

// readBatchOperationsFromFile reads batch operations from a JSON file
func readBatchOperationsFromFile(path string) ([]api.RichMenuBatchOperation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var operations []api.RichMenuBatchOperation
	if err := json.Unmarshal(data, &operations); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return operations, nil
}

// Validate command

func newRichMenuValidateCmd() *cobra.Command {
	return newRichMenuValidateCmdWithClient(nil, nil)
}

func newRichMenuValidateCmdWithClient(client *api.Client, menuOverride *api.CreateRichMenuRequest) *cobra.Command {
	var menuFile string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a rich menu definition",
		Long:  "Validate a rich menu JSON definition without creating it.",
		Example: `  # Validate a rich menu definition
  line richmenu validate --file menu.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var menu *api.CreateRichMenuRequest
			if menuOverride != nil {
				menu = menuOverride
			} else {
				if menuFile == "" {
					return fmt.Errorf("--file is required")
				}

				data, err := os.ReadFile(menuFile)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}

				menu = &api.CreateRichMenuRequest{}
				if err := json.Unmarshal(data, menu); err != nil {
					return fmt.Errorf("invalid JSON: %w", err)
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

			if err := c.ValidateRichMenu(cmd.Context(), menu); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"valid": true,
					"name":  menu.Name,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Rich menu definition valid: %s\n", menu.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&menuFile, "file", "", "JSON file containing rich menu definition (required)")
	// Note: --file is not marked required since menuOverride can be used in tests

	return cmd
}

// Download image command

func newRichMenuDownloadImageCmd() *cobra.Command {
	return newRichMenuDownloadImageCmdWithClient(nil)
}

func newRichMenuDownloadImageCmdWithClient(client *api.Client) *cobra.Command {
	var richMenuID string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "download-image",
		Short: "Download a rich menu image",
		Long:  "Download the image associated with a rich menu.",
		Example: `  # Download image to default filename
  line richmenu download-image --id richmenu-xxx

  # Download to specific path
  line richmenu download-image --id richmenu-xxx --output menu.png`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if richMenuID == "" {
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

			data, contentType, err := c.DownloadRichMenuImage(cmd.Context(), richMenuID)
			if err != nil {
				return fmt.Errorf("failed to download image: %w", err)
			}

			// Determine output filename
			filename := outputPath
			if filename == "" {
				ext := ".png"
				if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
					ext = ".jpg"
				}
				filename = fmt.Sprintf("%s%s", richMenuID, ext)
			}

			if err := os.WriteFile(filename, data, 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"richMenuId":  richMenuID,
					"filename":    filename,
					"contentType": contentType,
					"size":        len(data),
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Downloaded image to %s (%d bytes)\n", filename, len(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&richMenuID, "id", "", "Rich menu ID (required)")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output file path (default: richmenu-{id}.{ext})")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
