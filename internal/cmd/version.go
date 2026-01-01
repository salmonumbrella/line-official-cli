package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information - set by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "line-cli %s\n", version)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  commit: %s\n", commit)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  built:  %s\n", date)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  go:     %s\n", runtime.Version())
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}

	return cmd
}
