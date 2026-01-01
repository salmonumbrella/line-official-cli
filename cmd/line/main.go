package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/salmonumbrella/line-official-cli/internal/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := cmd.ExecuteContext(ctx, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
