package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/line-official-cli/internal/api"
	"github.com/spf13/cobra"
)

func newCouponCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coupon",
		Short: "Manage coupons",
		Long:  "Create, list, and manage coupons for your LINE Official Account.",
	}

	cmd.AddCommand(newCouponListCmd())
	cmd.AddCommand(newCouponCreateCmd())
	cmd.AddCommand(newCouponGetCmd())
	cmd.AddCommand(newCouponCloseCmd())

	return cmd
}

func newCouponListCmd() *cobra.Command {
	return newCouponListCmdWithClient(nil)
}

func newCouponListCmdWithClient(client *api.Client) *cobra.Command {
	var status string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all coupons",
		Long:  "Get a list of all coupons associated with your LINE Official Account.",
		Example: `  # List all coupons
  line coupon list

  # List only running coupons
  line coupon list --status running

  # List with limit
  line coupon list --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Convert status to uppercase for API (do this before client creation)
			var statusFilter []string
			if status != "" {
				switch status {
				case "running", "RUNNING":
					statusFilter = []string{"RUNNING"}
				case "draft", "DRAFT":
					statusFilter = []string{"DRAFT"}
				case "closed", "CLOSED":
					statusFilter = []string{"CLOSED"}
				default:
					return fmt.Errorf("invalid status: %s (use running, draft, or closed)", status)
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

			resp, err := c.ListCoupons(cmd.Context(), statusFilter, limit, "")
			if err != nil {
				return fmt.Errorf("failed to list coupons: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			}

			if len(resp.Coupons) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No coupons found")
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Coupons:")
			for _, coupon := range resp.Coupons {
				statusStr := ""
				if coupon.Status != "" {
					statusStr = fmt.Sprintf(" [%s]", coupon.Status)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s  %s%s\n", coupon.CouponID, coupon.Title, statusStr)
			}

			if resp.Next != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nMore coupons available. Use pagination to fetch more.\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status: running, draft, or closed")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of coupons to return")

	return cmd
}

func newCouponCreateCmd() *cobra.Command {
	return newCouponCreateCmdWithClient(nil)
}

func newCouponCreateCmdWithClient(client *api.Client) *cobra.Command {
	var title string
	var startTimestamp int64
	var endTimestamp int64
	var description string
	var imageURL string
	var discount int
	var timezone string
	var maxUse int
	var visibility string
	var acquisitionCondition string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new coupon",
		Long: `Create a new coupon with the specified parameters.

Required fields:
  --title          Coupon title
  --start          Start timestamp (Unix milliseconds)
  --end            End timestamp (Unix milliseconds)
  --max-use        Maximum number of times a user can use this coupon
  --visibility     Visibility setting: PUBLIC or UNLISTED
  --acquisition    How users can acquire the coupon: normal or lottery`,
		Example: `  # Create a basic coupon with fixed discount
  line coupon create --title "Summer Sale" \
    --start 1704067200000 --end 1735689600000 \
    --max-use 1 --visibility PUBLIC --acquisition normal \
    --discount 500

  # Create an unlisted lottery coupon
  line coupon create --title "Lucky Draw" \
    --start 1704067200000 --end 1735689600000 \
    --max-use 1 --visibility UNLISTED --acquisition lottery \
    --discount 1000

  # Create a coupon with timezone and description
  line coupon create --title "Welcome" \
    --start 1704067200000 --end 1735689600000 \
    --max-use 1 --visibility PUBLIC --acquisition normal \
    --timezone Asia/Tokyo --description "New user discount"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate required fields
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if startTimestamp == 0 {
				return fmt.Errorf("--start is required (Unix timestamp in milliseconds)")
			}
			if endTimestamp == 0 {
				return fmt.Errorf("--end is required (Unix timestamp in milliseconds)")
			}
			if maxUse <= 0 {
				return fmt.Errorf("--max-use is required (must be > 0)")
			}
			if visibility == "" {
				return fmt.Errorf("--visibility is required (PUBLIC or UNLISTED)")
			}
			if acquisitionCondition == "" {
				return fmt.Errorf("--acquisition is required (normal or lottery)")
			}

			// Validate visibility
			visibility = strings.ToUpper(visibility)
			if visibility != "PUBLIC" && visibility != "UNLISTED" {
				return fmt.Errorf("invalid --visibility: %s (use PUBLIC or UNLISTED)", visibility)
			}

			// Validate acquisition condition
			acquisitionCondition = strings.ToLower(acquisitionCondition)
			if acquisitionCondition != "normal" && acquisitionCondition != "lottery" {
				return fmt.Errorf("invalid --acquisition: %s (use normal or lottery)", acquisitionCondition)
			}

			// Validate timestamps
			if startTimestamp >= endTimestamp {
				return fmt.Errorf("--start must be before --end")
			}

			c := client
			if c == nil {
				var err error
				c, err = newAPIClient()
				if err != nil {
					return err
				}
			}

			req := &api.CreateCouponRequest{
				Title:                title,
				StartTimestamp:       startTimestamp,
				EndTimestamp:         endTimestamp,
				Description:          description,
				ImageURL:             imageURL,
				Timezone:             timezone,
				MaxUseCountPerTicket: maxUse,
				Visibility:           visibility,
				AcquisitionCondition: &api.AcquisitionCondition{
					Type: acquisitionCondition,
				},
			}

			// Add fixed discount reward if specified
			if discount > 0 {
				req.Reward = &api.CouponReward{
					Type: "discount",
					PriceInfo: &api.CouponPriceInfo{
						Type:        "fixed",
						FixedAmount: discount,
					},
				}
			}

			couponID, err := c.CreateCoupon(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("failed to create coupon: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"couponId": couponID,
					"title":    title,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created coupon: %s (ID: %s)\n", title, couponID)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Coupon title (required)")
	cmd.Flags().Int64Var(&startTimestamp, "start", 0, "Start timestamp in milliseconds (required)")
	cmd.Flags().Int64Var(&endTimestamp, "end", 0, "End timestamp in milliseconds (required)")
	cmd.Flags().IntVar(&maxUse, "max-use", 0, "Max times a user can use this coupon (required)")
	cmd.Flags().StringVar(&visibility, "visibility", "", "Visibility: PUBLIC or UNLISTED (required)")
	cmd.Flags().StringVar(&acquisitionCondition, "acquisition", "", "Acquisition type: normal or lottery (required)")
	cmd.Flags().StringVar(&description, "description", "", "Coupon description")
	cmd.Flags().StringVar(&imageURL, "image", "", "Image URL for the coupon")
	cmd.Flags().IntVar(&discount, "discount", 0, "Fixed discount amount")
	cmd.Flags().StringVar(&timezone, "timezone", "", "Timezone (e.g., Asia/Tokyo)")

	return cmd
}

func newCouponGetCmd() *cobra.Command {
	return newCouponGetCmdWithClient(nil)
}

func newCouponGetCmdWithClient(client *api.Client) *cobra.Command {
	var couponID string

	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get coupon details",
		Long:    "Get detailed information about a specific coupon.",
		Example: `  line coupon get --id coupon-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if couponID == "" {
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

			coupon, err := c.GetCoupon(cmd.Context(), couponID)
			if err != nil {
				return fmt.Errorf("failed to get coupon: %w", err)
			}

			if flags.Output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(coupon)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", coupon.CouponID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Title:       %s\n", coupon.Title)
			if coupon.Description != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", coupon.Description)
			}
			if coupon.Status != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:      %s\n", coupon.Status)
			}
			if coupon.StartTimestamp > 0 {
				startTime := time.UnixMilli(coupon.StartTimestamp)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Start:       %s\n", startTime.Format(time.RFC3339))
			}
			if coupon.EndTimestamp > 0 {
				endTime := time.UnixMilli(coupon.EndTimestamp)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "End:         %s\n", endTime.Format(time.RFC3339))
			}
			if coupon.Reward != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Reward Type: %s\n", coupon.Reward.Type)
				if coupon.Reward.PriceInfo != nil {
					if coupon.Reward.PriceInfo.FixedAmount > 0 {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Discount:    %d\n", coupon.Reward.PriceInfo.FixedAmount)
					}
					if coupon.Reward.PriceInfo.Rate > 0 {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Rate:        %d%%\n", coupon.Reward.PriceInfo.Rate)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&couponID, "id", "", "Coupon ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newCouponCloseCmd() *cobra.Command {
	return newCouponCloseCmdWithClient(nil)
}

func newCouponCloseCmdWithClient(client *api.Client) *cobra.Command {
	var couponID string

	cmd := &cobra.Command{
		Use:     "close",
		Short:   "Close a coupon",
		Long:    "Discontinue a coupon, preventing further distribution.",
		Example: `  line coupon close --id coupon-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if couponID == "" {
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

			if err := c.CloseCoupon(cmd.Context(), couponID); err != nil {
				return fmt.Errorf("failed to close coupon: %w", err)
			}

			if flags.Output == "json" {
				result := map[string]any{
					"couponId": couponID,
					"status":   "closed",
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Closed coupon: %s\n", couponID)
			return nil
		},
	}

	cmd.Flags().StringVar(&couponID, "id", "", "Coupon ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
