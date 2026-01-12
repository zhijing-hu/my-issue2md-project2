// Package main provides the entry point for the issue2md CLI tool
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bigwhite/issue2md/internal/cli"
)

func main() {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nReceived interrupt signal, shutting down...")
		cancel()
		os.Exit(1)
	}()

	// Create a new CLI instance
	var newCli cli.NewCLI
	cliInstance, err := newCli.Create()
	if err != nil {
		log.Fatalf("🔥 Failed to create CLI: %v", err)
	}

	// Parse command line arguments
	args, err := cli.ParseArgs()
	if err != nil {
		log.Fatalf("🔥 Failed to parse arguments: %v", err)
	}

	// Run the CLI with parsed arguments
	if err := cliInstance.Run(ctx, args); err != nil {
		log.Fatalf("🔥 Error: %v", err)
	}
}