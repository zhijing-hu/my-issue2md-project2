// Package cli provides the main execution flow for issue2md
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bigwhite/issue2md/internal/converter"
	"github.com/bigwhite/issue2md/internal/github"
	"github.com/bigwhite/issue2md/internal/parser"
)

// CLI represents the main CLI application
// It contains instances of all the required components
type CLI struct {
	gitClient  github.Client
	converter  converter.Converter
	parser     parser.Parser
	config     *config.Config
}

// NewCLI creates a new CLI instance with initialized components
type NewCLI struct{}

// Create initializes a new CLI instance
func (n NewCLI) Create() (*CLI, error) {
	// Initialize GitHub client with empty token (for now)
	// In real usage, this would get token from environment
	client, err := createGitHubClient("")
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Initialize converter
	conv, err := createConverter(false)
	if err != nil {
		return nil, fmt.Errorf("failed to create converter: %w", err)
	}

	// Initialize parser
	parserInstance := parser.NewParser()

	return &CLI{
		gitClient:  client,
		converter:  conv,
		parser:     parserInstance,
	}, nil
}

// Run executes the main CLI flow
func (c *CLI) Run(ctx context.Context, args *Args) error {
	// Show help/version/error-codes if requested (no need to validate)
	if args.Help {
		ShowHelp()
		return nil
	}
	if args.Version {
		ShowVersion()
		return nil
	}
	if args.ErrorCodes {
		ShowErrorCodes()
		return nil
	}

	// Validate arguments (this will fail if URL is missing/invalid)
	if err := args.Validate(); err != nil {
		return fmt.Errorf("argument validation failed: %w", err)
	}

	// Parse the GitHub URL to extract resource information
	resource, err := c.parser.Parse(args.URL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Create converter with appropriate reaction setting
	conv, err := createConverter(args.EnableReactions)
	if err != nil {
		return fmt.Errorf("failed to create converter with reactions setting: %w", err)
	}

	// For now, use the parser's resource information and create mock content
	// In a real implementation, this would fetch from GitHub API and convert
	owner := resource.Owner
	repo := resource.Repo

	var markdown string
	switch resource.ResourceType {
	case parser.TypeIssue:
		mockIssue := createMockIssue(owner, repo, resource.Number, resource.RawURL)
		var err error
		markdown, err = conv.ConvertIssue(mockIssue)
		if err != nil {
			return fmt.Errorf("failed to convert issue: %w", err)
		}

	case parser.TypePull:
		mockPR := createMockPullRequest(owner, repo, resource.Number, resource.RawURL)
		var err error
		markdown, err = conv.ConvertPullRequest(mockPR)
		if err != nil {
			return fmt.Errorf("failed to convert pull request: %w", err)
		}

	case parser.TypeDiscussion:
		mockDiscussion := createMockDiscussion(owner, repo, resource.Number, resource.RawURL)
		var err error
		markdown, err = conv.ConvertDiscussion(mockDiscussion)
		if err != nil {
			return fmt.Errorf("failed to convert discussion: %w", err)
		}

	default:
		return fmt.Errorf("unsupported resource type: %s", resource.ResourceType)
	}

	// Output the result (file or stdout)
	if args.Output != "" {
		if err := c.writeOutputFile(args.Output, markdown); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Successfully wrote markdown to: %s\n", args.Output)
	} else {
		// Write to stdout
		fmt.Println(markdown)
	}

	return nil
}

// createMockIssue creates a test issue for demonstration purposes
func createMockIssue(owner, repo string, number int, url string) *github.Issue {
	return &github.Issue{
		Title:      fmt.Sprintf("Test Issue #%d", number),
		Number:     number,
		State:      "open",
		Body:       fmt.Sprintf("This is a test issue created from %s/%s", owner, repo),
		CreatedAt:  time.Now().Add(-24 * time.Hour), // 1 day ago
		UpdatedAt:  time.Now(),
		Author:     github.User{Login: "testuser"},
		HTMLURL:    url,
		Reactions:    github.Reactions{PlusOne: 5, Heart: 2},
	}
}

// createMockPullRequest creates a test PR for demonstration purposes
func createMockPullRequest(owner, repo string, number int, url string) *github.PullRequest {
	return &github.PullRequest{
		Issue: github.Issue{
			Title:      fmt.Sprintf("Test Pull Request #%d", number),
			Number:     number,
			State:      "open",
			Body:       fmt.Sprintf("This is a test pull request for %s/%s", owner, repo),
			CreatedAt:  time.Now().Add(-24 * time.Hour),
			UpdatedAt:  time.Now(),
			Author:     github.User{Login: "testuser"},
			HTMLURL:    url,
			Reactions:    github.Reactions{Laugh: 3, Hurray: 1},
		},
	}
}

// createMockDiscussion creates a test discussion for demonstration purposes
func createMockDiscussion(owner, repo string, number int, url string) *github.Discussion {
	return &github.Discussion{
		Issue: github.Issue{
			Title:      fmt.Sprintf("Test Discussion #%d", number),
			Number:     number,
			State:      "open",
			Body:       fmt.Sprintf("This is a test discussion in %s/%s", owner, repo),
			CreatedAt:  time.Now().Add(-24 * time.Hour),
			UpdatedAt:  time.Now(),
			Author:     github.User{Login: "testuser"},
			HTMLURL:    url,
			Reactions:    github.Reactions{Rocket: 2, Eyes: 1},
		},
	}
}

// parseGitHubURL extracts owner and repo from GitHub URL
func (c *CLI) parseGitHubURL(url string) (string, string, error) {
	// Remove protocol and split path
	cleanURL := strings.TrimPrefix(url, "https://")
	parts := strings.Split(cleanURL, "/")

	if len(parts) < 4 {
		return "", "", fmt.Errorf("malformed GitHub URL: %s", url)
	}

	// Expect: github.com/owner/repo/type/number
	// parts[0] = "github.com", parts[1] = owner, parts[2] = repo
	owner := parts[1]
	repo := parts[2]

	return owner, repo, nil
}

// Helper functions for dependency injection
// These would be replaced with real client/converter creation in production

// createGitHubClient creates a GitHub client instance
func createGitHubClient(token string) (github.Client, error) {
	var newClient github.NewClient = func(opts ...github.ClientOption) (github.Client, error) {
		client := &github.GitHubClient{}
		// Apply options (simplified for now)
		return client, nil
	}
	return newClient(github.WithToken(token), github.WithBaseURL("https://api.github.com")).Create()
}

// createConverter creates a converter instance
func createConverter(enableReactions bool) (converter.Converter, error) {
	var newConv converter.NewConverter = func(opts ...converter.ConverterOption) (converter.Converter, error) {
		conv := &converter.ConverterImpl{
			enableReactions: enableReactions,
		}
		conv.LoadDefaultTemplates()
		return conv, nil
	}
	return newConv(converter.WithReactions(enableReactions)).Create()
}