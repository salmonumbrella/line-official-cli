package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newAudienceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "audience",
		Aliases: []string{"aud"},
		Short:   "Manage audience groups",
		Long:    "Create, list, and manage audience groups for targeted messaging.",
	}

	cmd.AddCommand(newAudienceListCmd())
	cmd.AddCommand(newAudienceGetCmd())
	cmd.AddCommand(newAudienceDeleteCmd())
	cmd.AddCommand(newAudienceCreateCmd())
	cmd.AddCommand(newAudienceAddUsersCmd())
	cmd.AddCommand(newAudienceCreateClickCmd())
	cmd.AddCommand(newAudienceCreateImpressionCmd())
	cmd.AddCommand(newAudienceUpdateDescriptionCmd())
	cmd.AddCommand(newAudienceSharedCmd())

	return cmd
}

func newAudienceListCmd() *cobra.Command {
	return newAudienceListCmdWithClient(nil)
}

func newAudienceListCmdWithClient(client *api.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List audience groups",
		Long:  "Get a list of all audience groups associated with your LINE Official Account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			groups, err := c.GetAudienceGroups(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list audience groups: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(groups)
			}

			if len(groups) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No audience groups found")
				return nil
			}

			if flags.Output == "table" {
				table := NewTable("ID", "DESCRIPTION", "STATUS", "USERS", "CREATED")
				for _, g := range groups {
					var created string
					if g.Created != nil {
						created = time.Unix(*g.Created, 0).Format("2006-01-02")
					}

					var audienceCount string
					if g.AudienceCount != nil {
						audienceCount = fmt.Sprintf("%d", *g.AudienceCount)
					}

					var status string
					if g.Status != nil {
						status = string(*g.Status)
					}

					var description string
					if g.Description != nil {
						description = *g.Description
					}

					var audienceGroupID string
					if g.AudienceGroupId != nil {
						audienceGroupID = fmt.Sprintf("%d", *g.AudienceGroupId)
					}

					table.AddRow(audienceGroupID, description, status, audienceCount, created)
				}
				table.Render(cmd.OutOrStdout())
				return nil
			}

			// Default text output
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Audience Groups:")
			for _, g := range groups {
				var created string
				if g.Created != nil {
					created = time.Unix(*g.Created, 0).Format("2006-01-02")
				} else {
					created = "unknown"
				}

				var audienceCount int64
				if g.AudienceCount != nil {
					audienceCount = *g.AudienceCount
				}

				var status string
				if g.Status != nil {
					status = string(*g.Status)
				} else {
					status = "unknown"
				}

				var description string
				if g.Description != nil {
					description = *g.Description
				}

				var audienceGroupID int64
				if g.AudienceGroupId != nil {
					audienceGroupID = *g.AudienceGroupId
				}

				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %d  %s  (%s, %d users, created %s)\n",
					audienceGroupID, description, status, audienceCount, created)
			}
			return nil
		},
	}
}

func newAudienceGetCmd() *cobra.Command {
	return newAudienceGetCmdWithClient(nil)
}

func newAudienceGetCmdWithClient(client *api.Client) *cobra.Command {
	var audienceGroupID int64

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get audience group details",
		Long:  "Get detailed information about a specific audience group.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if audienceGroupID <= 0 {
				return fmt.Errorf("invalid audience group ID: must be positive")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			resp, err := c.GetAudienceGroup(cmd.Context(), audienceGroupID)
			if err != nil {
				return fmt.Errorf("failed to get audience group: %w", err)
			}
			if resp == nil {
				return fmt.Errorf("audience group not found")
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			g := resp.AudienceGroup
			if g == nil {
				return fmt.Errorf("audience group not found")
			}

			var created string
			if g.Created != nil {
				created = time.Unix(*g.Created, 0).Format("2006-01-02 15:04:05")
			} else {
				created = "unknown"
			}

			var audienceGroupIDVal int64
			if g.AudienceGroupId != nil {
				audienceGroupIDVal = *g.AudienceGroupId
			}

			var description string
			if g.Description != nil {
				description = *g.Description
			}

			var groupType string
			if g.Type != nil {
				groupType = string(*g.Type)
			}

			var status string
			if g.Status != nil {
				status = string(*g.Status)
			}

			var audienceCount int64
			if g.AudienceCount != nil {
				audienceCount = *g.AudienceCount
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:          %d\n", audienceGroupIDVal)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type:        %s\n", groupType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:      %s\n", status)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Users:       %d\n", audienceCount)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created:     %s\n", created)
			return nil
		},
	}

	cmd.Flags().Int64Var(&audienceGroupID, "id", 0, "Audience group ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newAudienceDeleteCmd() *cobra.Command {
	return newAudienceDeleteCmdWithClient(nil)
}

func newAudienceDeleteCmdWithClient(client *api.Client) *cobra.Command {
	var audienceGroupID int64

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an audience group",
		Long:  "Delete an audience group by its ID.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if audienceGroupID <= 0 {
				return fmt.Errorf("invalid audience group ID: must be positive")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.DeleteAudienceGroup(cmd.Context(), audienceGroupID); err != nil {
				return fmt.Errorf("failed to delete audience group: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"deleted": audienceGroupID}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted audience group: %d\n", audienceGroupID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&audienceGroupID, "id", 0, "Audience group ID to delete (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newAudienceCreateCmd() *cobra.Command {
	return newAudienceCreateCmdWithClient(nil)
}

func newAudienceCreateCmdWithClient(client *api.Client) *cobra.Command {
	var description string
	var userIDsFile string
	var userIDs []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an audience group",
		Long: `Create an audience group from a list of user IDs.
User IDs can be provided via --users flag or from a file (one per line).
When using --file, the file is uploaded directly to LINE for better performance with large files.`,
		Example: `  # Create from user IDs
  line audience create --name "VIP Users" --users U123,U456,U789

  # Create from file (bulk upload)
  line audience create --name "Campaign Target" --file users.txt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if description == "" {
				return fmt.Errorf("--name is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			var resp *api.CreateAudienceResponse
			var usersCount int
			var apiErr error

			if userIDsFile != "" {
				// Use file upload API for bulk operations
				data, err := os.ReadFile(userIDsFile)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				for _, line := range strings.Split(string(data), "\n") {
					line = strings.TrimSpace(line)
					if line != "" {
						usersCount++
					}
				}
				if usersCount == 0 {
					return fmt.Errorf("file contains no user IDs")
				}

				resp, apiErr = c.CreateAudienceFromFile(cmd.Context(), description, userIDsFile)
				if apiErr != nil {
					return fmt.Errorf("failed to create audience: %w", apiErr)
				}
			} else if len(userIDs) > 0 {
				usersCount = len(userIDs)
				resp, apiErr = c.CreateAudienceGroup(cmd.Context(), description, userIDs)
				if apiErr != nil {
					return fmt.Errorf("failed to create audience: %w", apiErr)
				}
			} else {
				return fmt.Errorf("specify --users or --file")
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created audience group: %d (%s)\n", resp.AudienceGroupID, description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Users: %d\n", usersCount)
			return nil
		},
	}

	cmd.Flags().StringVar(&description, "name", "", "Audience group name/description (required)")
	cmd.Flags().StringSliceVar(&userIDs, "users", nil, "Comma-separated user IDs")
	cmd.Flags().StringVar(&userIDsFile, "file", "", "File containing user IDs (one per line)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newAudienceAddUsersCmd() *cobra.Command {
	return newAudienceAddUsersCmdWithClient(nil)
}

func newAudienceAddUsersCmdWithClient(client *api.Client) *cobra.Command {
	var audienceGroupID int64
	var userIDs []string
	var userIDsFile string
	var description string

	cmd := &cobra.Command{
		Use:   "add-users",
		Short: "Add users to an existing audience group",
		Long: `Add user IDs to an existing audience group.
User IDs can be provided via --users flag or from a file (one per line).`,
		Example: `  # Add users to audience
  line audience add-users --id 12345 --users U123,U456,U789

  # Add users from file
  line audience add-users --id 12345 --file more-users.txt

  # Add users with description
  line audience add-users --id 12345 --users U123,U456 --description "Added batch 2"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if audienceGroupID <= 0 {
				return fmt.Errorf("invalid audience group ID: must be positive")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			var usersCount int

			if userIDsFile != "" {
				// Use file upload API for bulk operations
				data, err := os.ReadFile(userIDsFile)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				for _, line := range strings.Split(string(data), "\n") {
					line = strings.TrimSpace(line)
					if line != "" {
						usersCount++
					}
				}
				if usersCount == 0 {
					return fmt.Errorf("file contains no user IDs")
				}

				if err := c.AddUsersToAudienceFromFile(cmd.Context(), audienceGroupID, userIDsFile, description); err != nil {
					return fmt.Errorf("failed to add users to audience: %w", err)
				}
			} else if len(userIDs) > 0 {
				usersCount = len(userIDs)
				if err := c.AddUsersToAudience(cmd.Context(), audienceGroupID, userIDs, description); err != nil {
					return fmt.Errorf("failed to add users to audience: %w", err)
				}
			} else {
				return fmt.Errorf("specify --users or --file")
			}

			if flags.Output == "json" {
				result := map[string]any{
					"audienceGroupId": audienceGroupID,
					"usersAdded":      usersCount,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added %d users to audience group %d\n", usersCount, audienceGroupID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&audienceGroupID, "id", 0, "Audience group ID (required)")
	cmd.Flags().StringSliceVar(&userIDs, "users", nil, "Comma-separated user IDs")
	cmd.Flags().StringVar(&userIDsFile, "file", "", "File containing user IDs (one per line)")
	cmd.Flags().StringVar(&description, "description", "", "Description for this upload batch")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newAudienceCreateClickCmd() *cobra.Command {
	return newAudienceCreateClickCmdWithClient(nil)
}

func newAudienceCreateClickCmdWithClient(client *api.Client) *cobra.Command {
	var name string
	var requestID string

	cmd := &cobra.Command{
		Use:   "create-click",
		Short: "Create audience from message click events",
		Long:  "Create an audience group from users who clicked a link in a message.",
		Example: `  # Create click-based audience
  line audience create-click --name "Clicked Campaign Link" --request abc123-def456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
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

			resp, err := c.CreateClickBasedAudience(cmd.Context(), name, requestID)
			if err != nil {
				return fmt.Errorf("failed to create click-based audience: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created click-based audience group: %d (%s)\n", resp.AudienceGroupID, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Audience group name/description (required)")
	cmd.Flags().StringVar(&requestID, "request", "", "Request ID of the message (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("request")

	return cmd
}

func newAudienceCreateImpressionCmd() *cobra.Command {
	return newAudienceCreateImpressionCmdWithClient(nil)
}

func newAudienceCreateImpressionCmdWithClient(client *api.Client) *cobra.Command {
	var name string
	var requestID string

	cmd := &cobra.Command{
		Use:   "create-impression",
		Short: "Create audience from message impression events",
		Long:  "Create an audience group from users who saw (received an impression of) a message.",
		Example: `  # Create impression-based audience
  line audience create-impression --name "Saw Campaign Message" --request abc123-def456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
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

			resp, err := c.CreateImpressionBasedAudience(cmd.Context(), name, requestID)
			if err != nil {
				return fmt.Errorf("failed to create impression-based audience: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created impression-based audience group: %d (%s)\n", resp.AudienceGroupID, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Audience group name/description (required)")
	cmd.Flags().StringVar(&requestID, "request", "", "Request ID of the message (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("request")

	return cmd
}

func newAudienceUpdateDescriptionCmd() *cobra.Command {
	return newAudienceUpdateDescriptionCmdWithClient(nil)
}

func newAudienceUpdateDescriptionCmdWithClient(client *api.Client) *cobra.Command {
	var audienceGroupID int64
	var description string

	cmd := &cobra.Command{
		Use:   "update-description",
		Short: "Update audience group description",
		Long:  "Update the description of an existing audience group.",
		Example: `  # Update audience description
  line audience update-description --id 12345 --description "Updated VIP Users"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if audienceGroupID <= 0 {
				return fmt.Errorf("invalid audience group ID: must be positive")
			}
			if description == "" {
				return fmt.Errorf("--description is required")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.UpdateAudienceDescription(cmd.Context(), audienceGroupID, description); err != nil {
				return fmt.Errorf("failed to update audience description: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"audienceGroupId": audienceGroupID,
					"description":     description,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated description for audience group %d\n", audienceGroupID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&audienceGroupID, "id", 0, "Audience group ID (required)")
	cmd.Flags().StringVar(&description, "description", "", "New description (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func newAudienceSharedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shared",
		Short: "Manage shared audience groups",
		Long:  "List and get details of shared audience groups.",
	}

	cmd.AddCommand(newAudienceSharedListCmd())
	cmd.AddCommand(newAudienceSharedGetCmd())

	return cmd
}

func newAudienceSharedListCmd() *cobra.Command {
	return newAudienceSharedListCmdWithClient(nil)
}

func newAudienceSharedListCmdWithClient(client *api.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List shared audience groups",
		Long:  "Get a list of all shared audience groups.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			groups, err := c.GetSharedAudienceGroups(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list shared audience groups: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(groups)
			}

			if len(groups) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No shared audience groups found")
				return nil
			}

			if flags.Output == "table" {
				table := NewTable("ID", "DESCRIPTION", "STATUS", "USERS", "CREATED")
				for _, g := range groups {
					var created string
					if g.Created != nil {
						created = time.Unix(*g.Created, 0).Format("2006-01-02")
					}

					var audienceCount string
					if g.AudienceCount != nil {
						audienceCount = fmt.Sprintf("%d", *g.AudienceCount)
					}

					var status string
					if g.Status != nil {
						status = string(*g.Status)
					}

					var description string
					if g.Description != nil {
						description = *g.Description
					}

					var audienceGroupID string
					if g.AudienceGroupId != nil {
						audienceGroupID = fmt.Sprintf("%d", *g.AudienceGroupId)
					}

					table.AddRow(audienceGroupID, description, status, audienceCount, created)
				}
				table.Render(cmd.OutOrStdout())
				return nil
			}

			// Default text output
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Shared Audience Groups:")
			for _, g := range groups {
				var created string
				if g.Created != nil {
					created = time.Unix(*g.Created, 0).Format("2006-01-02")
				} else {
					created = "unknown"
				}

				var audienceCount int64
				if g.AudienceCount != nil {
					audienceCount = *g.AudienceCount
				}

				var status string
				if g.Status != nil {
					status = string(*g.Status)
				} else {
					status = "unknown"
				}

				var description string
				if g.Description != nil {
					description = *g.Description
				}

				var audienceGroupID int64
				if g.AudienceGroupId != nil {
					audienceGroupID = *g.AudienceGroupId
				}

				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %d  %s  (%s, %d users, created %s)\n",
					audienceGroupID, description, status, audienceCount, created)
			}
			return nil
		},
	}
}

func newAudienceSharedGetCmd() *cobra.Command {
	return newAudienceSharedGetCmdWithClient(nil)
}

func newAudienceSharedGetCmdWithClient(client *api.Client) *cobra.Command {
	var audienceGroupID int64

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get shared audience group details",
		Long:  "Get detailed information about a specific shared audience group.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if audienceGroupID <= 0 {
				return fmt.Errorf("invalid audience group ID: must be positive")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			resp, err := c.GetSharedAudienceGroup(cmd.Context(), audienceGroupID)
			if err != nil {
				return fmt.Errorf("failed to get shared audience group: %w", err)
			}
			if resp == nil {
				return fmt.Errorf("shared audience group not found")
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			g := resp.AudienceGroup
			if g == nil {
				return fmt.Errorf("shared audience group not found")
			}

			var created string
			if g.Created != nil {
				created = time.Unix(*g.Created, 0).Format("2006-01-02 15:04:05")
			} else {
				created = "unknown"
			}

			var audienceGroupIDVal int64
			if g.AudienceGroupId != nil {
				audienceGroupIDVal = *g.AudienceGroupId
			}

			var description string
			if g.Description != nil {
				description = *g.Description
			}

			var groupType string
			if g.Type != nil {
				groupType = string(*g.Type)
			}

			var status string
			if g.Status != nil {
				status = string(*g.Status)
			}

			var audienceCount int64
			if g.AudienceCount != nil {
				audienceCount = *g.AudienceCount
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:          %d\n", audienceGroupIDVal)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type:        %s\n", groupType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:      %s\n", status)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Users:       %d\n", audienceCount)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created:     %s\n", created)

			// Display owner information if available
			if resp.Owner != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Owner:")
				if resp.Owner.Name != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Name:    %s\n", *resp.Owner.Name)
				}
				if resp.Owner.ServiceType != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Service: %s\n", *resp.Owner.ServiceType)
				}
			}

			return nil
		},
	}

	cmd.Flags().Int64Var(&audienceGroupID, "id", 0, "Shared audience group ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
