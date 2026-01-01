package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show configuration",
		Long: `Show the current configuration and config file location.

Configuration is loaded from (in order of priority):
  1. Command-line flags (highest)
  2. Environment variables
  3. Config file
  4. Built-in defaults (lowest)

Config file locations (first found is used):
  - $XDG_CONFIG_HOME/line-cli/config.yaml
  - ~/.config/line-cli/config.yaml
  - ~/.line-cli.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfig()
		},
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigPathCmd())
	cmd.AddCommand(newConfigExampleCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfig()
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.Output == "json" {
				type pathOutput struct {
					Loaded      string `json:"loaded,omitempty"`
					Recommended string `json:"recommended"`
				}
				out := pathOutput{
					Loaded: cfg.ConfigPath(),
				}
				if recommended, err := config.DefaultConfigPath(); err == nil {
					out.Recommended = recommended
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			if path := cfg.ConfigPath(); path != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Loaded:      %s\n", path)
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Loaded:      (none)")
			}
			if recommended, err := config.DefaultConfigPath(); err == nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Recommended: %s\n", recommended)
			}
			return nil
		},
	}
}

func newConfigExampleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "example",
		Short: "Print example config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), config.ExampleConfig())
			return nil
		},
	}
}

func runConfig() error {
	if flags.Output == "json" {
		type configOutput struct {
			ConfigPath string `json:"config_path,omitempty"`
			Account    string `json:"account,omitempty"`
			Output     string `json:"output"`
			Debug      bool   `json:"debug"`
		}
		out := configOutput{
			ConfigPath: cfg.ConfigPath(),
			Account:    cfg.Account,
			Output:     getDefault(cfg.Output, "text"),
			Debug:      cfg.Debug,
		}
		enc := json.NewEncoder(nil)
		enc.SetIndent("", "  ")
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Text output
	fmt.Println("Configuration")
	fmt.Println("=============")

	if path := cfg.ConfigPath(); path != "" {
		fmt.Printf("Config file: %s\n", path)
	} else {
		fmt.Println("Config file: (not found)")
		if recommended, err := config.DefaultConfigPath(); err == nil {
			fmt.Printf("             Create at: %s\n", recommended)
		}
	}

	fmt.Println()
	fmt.Println("Values (from config file):")

	if cfg.Account != "" {
		fmt.Printf("  account: %s\n", cfg.Account)
	} else {
		fmt.Println("  account: (not set)")
	}

	if cfg.Output != "" {
		fmt.Printf("  output:  %s\n", cfg.Output)
	} else {
		fmt.Println("  output:  (not set, default: text)")
	}

	fmt.Printf("  debug:   %v\n", cfg.Debug)

	fmt.Println()
	fmt.Println("Run 'line config example' to see an example config file.")

	return nil
}
