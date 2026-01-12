// Package github provides GitHub API client functionality
package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// githubClient is the concrete implementation of the Client interface
type githubClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// Client defines the interface for interacting with GitHub API
// This interface supports fetching Issues, Pull Requests, and Discussions
type Client interface {
	// GetIssue fetches an issue from GitHub by owner and repo name and issue number
	// Returns the Issue struct or an error
	GetIssue(owner, repo string, number int) (*Issue, error)

	// GetPullRequest fetches a pull request from GitHub by owner and repo name and PR number
	// Returns the PullRequest struct or an error
	GetPullRequest(owner, repo string, number int) (*PullRequest, error)

	// GetDiscussion fetches a discussion from GitHub by owner and repo name and discussion number
	// Returns the Discussion struct or an error
	GetDiscussion(owner, repo string, number int) (*Discussion, error)
}

// ClientOption defines a function type for configuring the GitHub client
type ClientOption func(*githubClient) error

// WithToken configures the client with a GitHub API token
func WithToken(token string) ClientOption {
	return func(c *githubClient) error {
		c.token = token
		return nil
	}
}

// WithHTTPClient allows customization of the HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *githubClient) error {
		if client == nil {
			return NewAPIError(ErrNetworkFailure, "HTTP client cannot be nil", nil)
		}
		c.httpClient = client
		return nil
	}
}

// WithBaseURL allows customization of the GitHub API base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *githubClient) error {
		if _, err := url.Parse(baseURL); err != nil {
			return NewAPIError(ErrInvalidURL, "Invalid base URL", err)
		}
		c.baseURL = baseURL
		return nil
	}
}

// NewClient creates a new GitHub client with the provided options
type NewClient func(opts ...ClientOption) (Client, error)

// Create a new GitHub client
func (n NewClient) Create(opts ...ClientOption) (Client, error) {
	client := &githubClient{
		httpClient: http.DefaultClient,
		baseURL:    "https://api.github.com",
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, NewAPIError(ErrInvalidToken, "Failed to configure client", err)
		}
	}

	return client, nil
}

// GetIssue fetches an issue from GitHub
func (c *githubClient) GetIssue(owner, repo string, number int) (*Issue, error) {
	var issue Issue
	if err := c.fetchResource(owner, repo, TypeIssue, number, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// GetPullRequest fetches a pull request from GitHub
func (c *githubClient) GetPullRequest(owner, repo string, number int) (*PullRequest, error) {
	var pr PullRequest
	if err := c.fetchResource(owner, repo, TypePull, number, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// GetDiscussion fetches a discussion from GitHub
func (c *githubClient) GetDiscussion(owner, repo string, number int) (*Discussion, error) {
	var discussion Discussion
	if err := c.fetchResource(owner, repo, TypeDiscussion, number, &discussion); err != nil {
		return nil, err
	}
	return &discussion, nil
}

// fetchResource is the generic function to fetch any GitHub resource
func (c *githubClient) fetchResource(owner, repo string, resourceType ResourceType, number int, result interface{}) error {
	// Build the API URL
	apiURL := fmt.Sprintf("%s/repos/%s/%s/%s/%d", c.baseURL, owner, repo, resourceType, number)

	// Create the request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return NewAPIError(ErrInvalidURL, "Failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewAPIError(ErrNetworkFailure, "Failed to execute HTTP request", err)
	}
	defer resp.Body.Close()

	// Handle the response
	return c.handleResponse(resp, result)
}

// handleResponse processes the HTTP response and maps errors
func (c *githubClient) handleResponse(resp *http.Response, result interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewAPIError(ErrNetworkFailure, "Failed to read response body", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// Success case - unmarshal the JSON
		if err := json.Unmarshal(body, result); err != nil {
			return NewAPIError(ErrInvalidURL, "Failed to parse response JSON", err)
		}
		return nil

	case http.StatusNotFound:
		return NewAPIError(ErrNotFound, "Resource not found (404)", nil)

	case http.StatusForbidden:
		// Check if it's a rate limit error or permission error
		errorBody := make(map[string]interface{})
		if err := json.Unmarshal(body, &errorBody); err == nil {
			if msg, ok := errorBody["message"].(string); ok {
				if strings.Contains(msg, "rate limit") {
					return NewAPIError(ErrInvalidToken, "API rate limit exceeded (403)", nil)
				}
				if strings.Contains(msg, "protected by organization SAMR") {
					return NewAPIError(ErrNoPermission, "Resource protected by organization policy (403)", nil)
				}
			}
		}
		return NewAPIError(ErrNoPermission, "Forbidden access (403)", nil)

	case http.StatusNotModified:
		return NewAPIError(ErrInvalidToken, "Not modified (304)", nil)

	case http.StatusUnauthorized:
		return NewAPIError(ErrInvalidToken, "Invalid token (401)", nil)

	case http.StatusInternalServerError:
		return NewAPIError(ErrNetworkFailure, "Internal server error (500)", nil)

	default:
		return NewAPIError(ErrNetworkFailure, fmt.Sprintf("Unexpected status code %d", resp.StatusCode), nil)
	}
}