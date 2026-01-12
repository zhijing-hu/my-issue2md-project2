// Package cli provides a simplified working CLI implementation
package cli

import (
	"context"
	"fmt"
	"os"
)

// SimpleCLI is a working minimal CLI implementation
// This focuses on the core flow without complex dependency injection
func SimpleCLI(ctx context.Context, args []string) error {
	// Use existing Args parsing
	parsedArgs, err := ParseArgs()
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Handle informational flags
	if parsedArgs.Help {
		ShowHelp()
		return nil
	}
	if parsedArgs.Version {
		ShowVersion()
		return nil
	}
	if parsedArgs.ErrorCodes {
		ShowErrorCodes()
		return nil
	}

	// Validate arguments
	if err := parsedArgs.Validate(); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	// Parse the URL using existing parser functionality
	resource, err := parser.NewParser().Parse(parsedArgs.URL)
	if err != nil {
		return fmt.Errorf("failed to parse GitHub URL: %w", err)
	}

	// For now, just show what we parsed (production version would fetch real data)
	fmt.Printf("✅ Successfully parsed GitHub %s: %s/%s#%d\n",
		resource.ResourceType, resource.Owner, resource.Repo, resource.Number)

	if parsedArgs.EnableReactions {
		fmt.Println("   👍 Reactions enabled")
	}

	if parsedArgs.Output != "" {
		fmt.Printf("   📄 Will output to: %s\n", parsedArgs.Output)
	}

	// In a full implementation, this would:
	// 1. Fetch from GitHub API using the client
	// 2. Convert to markdown using converter
	// 3. Write to file or stdout

	return nil
}

// MainEntrypoint is the CLI entry point for the application
func MainEntrypoint() {
	// In a real application, this would be in main.go
	// For now, provide a simple way to test the CLI
	args := os.Args[1:]
	if len(args) == 0 {
		args = append(args, "--help")
	}

	// Store original args and restore after
	originalArgs := os.Args
	os.Args = append([]string{"issue2md"}, args...)
	defer func() { os.Args = originalArgs }()

	if err := SimpleCLI(context.Background(), args); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}