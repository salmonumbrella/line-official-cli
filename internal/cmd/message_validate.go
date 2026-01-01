package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newMessageValidateCmd() *cobra.Command {
	return newMessageValidateCmdWithClient(nil)
}

func newMessageValidateCmdWithClient(client *api.Client) *cobra.Command {
	var messageType string
	var messagesJSON string
	var filePath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate message objects",
		Long: `Validate message objects before sending.
Catches formatting errors without actually sending messages.
Provide messages via --messages flag or --file flag (not both).`,
		Example: `  # Validate a text message for push
  line message validate --type push --messages '[{"type":"text","text":"Hello"}]'

  # Validate a flex message for broadcast
  line message validate --type broadcast --messages '[{"type":"flex","altText":"Menu",...}]'

  # Validate from a JSON file
  line message validate --type push --file messages.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if messageType == "" {
				return fmt.Errorf("--type is required (reply|push|multicast|narrowcast|broadcast)")
			}
			if messagesJSON == "" && filePath == "" {
				return fmt.Errorf("--messages or --file is required")
			}
			if messagesJSON != "" && filePath != "" {
				return fmt.Errorf("specify either --messages or --file, not both")
			}

			validTypes := map[string]bool{
				"reply": true, "push": true, "multicast": true,
				"narrowcast": true, "broadcast": true,
			}
			if !validTypes[messageType] {
				return fmt.Errorf("--type must be one of: reply, push, multicast, narrowcast, broadcast")
			}

			// Get messages from file or flag
			var messagesData []byte
			if filePath != "" {
				data, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				messagesData = data
			} else {
				messagesData = []byte(messagesJSON)
			}

			var messages []json.RawMessage
			if err := json.Unmarshal(messagesData, &messages); err != nil {
				return fmt.Errorf("invalid messages JSON: %w", err)
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			if err := c.ValidateMessage(cmd.Context(), messageType, messages); err != nil {
				if flags.Output == "json" {
					result := map[string]any{"valid": false, "error": err.Error()}
					enc := json.NewEncoder(cmd.OutOrStdout())
					enc.SetIndent("", "  ")
					return enc.Encode(result)
				}
				return fmt.Errorf("validation failed: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{"valid": true, "type": messageType, "messageCount": len(messages)}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Validation passed: %d message(s) valid for %s\n", len(messages), messageType)
			return nil
		},
	}

	cmd.Flags().StringVar(&messageType, "type", "", "Message type: reply|push|multicast|narrowcast|broadcast (required)")
	cmd.Flags().StringVar(&messagesJSON, "messages", "", "Messages JSON array")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to JSON file containing messages array")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}
