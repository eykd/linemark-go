// Package main is the entry point for the lmk CLI application.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/eykd/linemark-go/cmd"
)

func main() {
	// Create a context that is cancelled on SIGINT (Ctrl+C).
	// This enables graceful shutdown for long-running operations.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprint(os.Stderr, cmd.FormatError(err))
		os.Exit(1)
	}
}
