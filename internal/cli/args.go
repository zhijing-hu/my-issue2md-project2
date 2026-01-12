// Package cli provides command-line argument parsing for issue2md
package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Args represents parsed command-line arguments for issue2md
// This struct holds all required and optional parameters
// according to the specification in spec.md §3.1
type Args struct {
	// Required parameters
	URL string // GitHub resource URL (Issue/PR/Discussion)

	// Optional parameters
	Output         string // Output file path (default: print to stdout)
	EnableReactions bool  // Enable reactions in markdown output
	Help           bool   // Show help information
	Version        bool   // Show version information
	ErrorCodes     bool   // Show error codes information
}

// ParseArgs parses command-line arguments into an Args struct
func ParseArgs() (*Args, error) {
	args := &Args{}

	// Define command-line flags
	flagSet := flag.NewFlagSet("issue2md", flag.ContinueOnError)
	flagSet.StringVar(&args.Output, "output", "", "Output file path")
	flagSet.StringVar(&args.Output, "o", "", "Output file path (shorthand)")
	flagSet.BoolVar(&args.EnableReactions, "enable-reactions", false, "Enable reactions in markdown output")
	flagSet.BoolVar(&args.Help, "help", false, "Show help information")
	flagSet.BoolVar(&args.Help, "h", false, "Show help information (shorthand)")
	flagSet.BoolVar(&args.Version, "version", false, "Show version information")
	flagSet.BoolVar(&args.ErrorCodes, "error-codes", false, "Show error codes information")

	// Custom usage to prevent default flag printing on error
	flagSet.Usage = func() {
		// Show help will be handled by our validate/usage printing
	}

	// Parse the command-line arguments
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Get the non-flag arguments (URL should be the first non-flag arg)
	if flagSet.NArg() > 0 {
		args.URL = flagSet.Arg(0)
	}

	// Don't validate here - let the caller handle validation
	// This allows showing help/version without URL validation errors

	return args, nil
}

// Validate checks that the parsed arguments are valid
func (a *Args) Validate() error {
	// If help, version, or error-codes are requested, no validation needed
	if a.Help || a.Version || a.ErrorCodes {
		return nil
	}

	// URL is required for normal operation
	if a.URL == "" {
		return errors.New("missing required parameter: GitHub resource URL")
	}

	// Validate URL format (basic check)
	if !strings.HasPrefix(a.URL, "https://github.com/") {
		return fmt.Errorf("invalid GitHub URL format: %s", a.URL)
	}

	// Validate URL contains valid resource type
	expectedPaths := []string{"issues/", "pull/", "discussions/"}
	isValid := false
	for _, path := range expectedPaths {
		if strings.Contains(a.URL, path) {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("unrecognized GitHub URL format: %s. Expected format like: https://github.com/owner/repo/issues/1", a.URL)
	}

	return nil
}

// ShowHelp prints usage information
func ShowHelp() {
	fmt.Print(`issue2md - Convert GitHub issues, PRs, and discussions to Markdown

Usage:
  issue2md [options] <GitHub-URL>

Arguments:
  <GitHub-URL>    Required. GitHub issue, PR, or discussion URL
                 Examples:
                   https://github.com/owner/repo/issues/1
                   https://github.com/owner/repo/pull/42
                   https://github.com/owner/repo/discussions/8

Options:
  -o, --output <file>             Output to file instead of stdout
  --enable-reactions              Include reaction counts in markdown
  -h, --help                      Show this help message
  --version                       Show version information
  --error-codes                   Show error code explanations

Examples:
  issue2md https://github.com/owner/repo/issues/123
  issue2md --enable-reactions https://github.com/owner/repo/pull/456 > pr.md
  issue2md -o output.md https://github.com/owner/repo/discussions/789
`)
}

// ShowVersion prints version information
func ShowVersion() {
	fmt.Println("issue2md version 1.0.0")
	fmt.Println("GitHub resource to Markdown converter")
	fmt.Println("© 2026 issue2md project")
}

// ShowErrorCodes prints error code information
func ShowErrorCodes() {
	fmt.Print(`issue2md Error Codes:

E001: Invalid URL format
E002: Network connection failure
E003: Resource not found (404)
E004: Permission denied (403)
E005: Invalid GitHub API token
E006: Invalid input parameter

For detailed troubleshooting, see the documentation.
`)
}