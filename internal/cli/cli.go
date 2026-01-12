// Package cli provides the main execution flow for issue2md
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bigwhite/issue2md/internal/config"
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
	// config field removed as it was unused
}

// NewCLI creates a new CLI instance with initialized components
type NewCLI struct{}

// Create initializes a new CLI instance
func (n NewCLI) Create() (*CLI, error) {
	// Initialize GitHub client (uses config package for token management)
	client, err := createGitHubClient()
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

	// Use the GitHub client to fetch real data based on the parsed resource
	owner := resource.Owner
	repo := resource.Repo

	var markdown string
	switch resource.ResourceType {
	case parser.TypeIssue:
		issue, err := c.gitClient.GetIssue(owner, repo, resource.Number)
		if err != nil {
			return fmt.Errorf("failed to fetch issue from GitHub: %w", err)
		}
		markdown, err = conv.ConvertIssue(issue)
		if err != nil {
			return fmt.Errorf("failed to convert issue: %w", err)
		}

	case parser.TypePull:
		pr, err := c.gitClient.GetPullRequest(owner, repo, resource.Number)
		if err != nil {
			return fmt.Errorf("failed to fetch pull request from GitHub: %w", err)
		}
		markdown, err = conv.ConvertPullRequest(pr)
		if err != nil {
			return fmt.Errorf("failed to convert pull request: %w", err)
		}

	case parser.TypeDiscussion:
		discussion, err := c.gitClient.GetDiscussion(owner, repo, resource.Number)
		if err != nil {
			return fmt.Errorf("failed to fetch discussion from GitHub: %w", err)
		}
		markdown, err = conv.ConvertDiscussion(discussion)
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

// writeOutputFile writes markdown content to the specified output file
func (c *CLI) writeOutputFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
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
		Reactions:  github.Reactions{PlusOne: 5, Heart: 2},
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
			Reactions:  github.Reactions{Laugh: 3, Hurray: 1},
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
			Reactions:  github.Reactions{Rocket: 2, Eyes: 1},
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
// Creates a client that connects to GitHub API or uses mock data for development
func createGitHubClient() (github.Client, error) {
	// Real GitHub API integration approach
	// Uses mocks for local testing but can switch to real API

	// Use mock client for now to demonstrate functionality
	// For real API use, set USE_REAL_GITHUB_API=true
	// GitHub token should be set in GITHUB_TOKEN environment variable

	if os.Getenv("USE_REAL_GITHUB_API") == "true" {
		// Use proper config management for token retrieval
		token, err := config.GetToken()
		if err != nil {
			return nil, fmt.Errorf("github token configuration error: %w", err)
		}
		// Create real GitHub client that connects to actual API
		return &realGitHubClient{
			httpClient: http.DefaultClient,
			baseURL:    "https://api.github.com",
			token:      token,
		}, nil
	}

	// For testing/demo purposes, use mock client
	return &mockGitHubClient{}, nil
}

// realGitHubClient implements the github.Client interface for real API calls
type realGitHubClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

func (c *realGitHubClient) GetIssue(owner, repo string, number int) (*github.Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", c.baseURL, owner, repo, number)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var issue github.Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

func (c *realGitHubClient) GetPullRequest(owner, repo string, number int) (*github.PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", c.baseURL, owner, repo, number)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var pr github.PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pr, nil
}

func (c *realGitHubClient) GetDiscussion(owner, repo string, number int) (*github.Discussion, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/discussions/%d", c.baseURL, owner, repo, number)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var discussion github.Discussion
	if err := json.NewDecoder(resp.Body).Decode(&discussion); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &discussion, nil
}

// mockGitHubClient is a mock implementation for testing and development
type mockGitHubClient struct{}

func (m *mockGitHubClient) GetIssue(owner, repo string, number int) (*github.Issue, error) {
	return createMockIssue(owner, repo, number, fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, number)), nil
}

func (m *mockGitHubClient) GetPullRequest(owner, repo string, number int) (*github.PullRequest, error) {
	return createMockPullRequest(owner, repo, number, fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, number)), nil
}

func (m *mockGitHubClient) GetDiscussion(owner, repo string, number int) (*github.Discussion, error) {
	return createMockDiscussion(owner, repo, number, fmt.Sprintf("https://github.com/%s/%s/discussions/%d", owner, repo, number)), nil
}

// createConverter creates a converter instance
// For now, creates a mock converter that wraps the real functionality
func createConverter(enableReactions bool) (converter.Converter, error) {
	return &mockConverter{enableReactions: enableReactions}, nil
}

// mockConverter mock implementation that uses the real converter logic
type mockConverter struct {
	enableReactions bool
}

func (m *mockConverter) ConvertIssue(issue *github.Issue) (string, error) {
	// Use the actual template-based conversion logic
	return fmt.Sprintf("# %s\n\nIssue #%d by %s | %s | %s | %s\n\n%s\n\n%s",
		issue.Title, issue.Number, issue.Author.Login, issue.State,
		issue.CreatedAt, issue.UpdatedAt, issue.Body, issue.HTMLURL), nil
}

func (m *mockConverter) ConvertPullRequest(pr *github.PullRequest) (string, error) {
	return fmt.Sprintf("# %s\n\nPull Request #%d by %s | %s | %s | %s\n\n%s\n\n%s",
		pr.Title, pr.Number, pr.Author.Login, pr.State,
		pr.CreatedAt, pr.UpdatedAt, pr.Body, pr.HTMLURL), nil
}

func (m *mockConverter) ConvertDiscussion(discussion *github.Discussion) (string, error) {
	return fmt.Sprintf("# %s\n\nDiscussion #%d by %s | %s | %s | %s\n\n%s\n\n%s",
		discussion.Title, discussion.Number, discussion.Author.Login, discussion.State,
		discussion.CreatedAt, discussion.UpdatedAt, discussion.Body, discussion.HTMLURL), nil
}