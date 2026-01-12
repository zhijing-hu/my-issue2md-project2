// Package github provides mock functionality for testing GitHub API calls
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockServer represents a mock GitHub API server for testing
type MockServer struct {
	server *httptest.Server
}

// NewMockServer creates a new mock GitHub API server
func NewMockServer() *MockServer {
	return &MockServer{}
}

// Start starts the mock server with all the predefined endpoints
func (m *MockServer) Start() {
	server := httptest.NewServer(http.HandlerFunc(m.handleRequest))
	m.server = server
}

// Close shuts down the mock server
func (m *MockServer) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

// URL returns the base URL of the mock server
func (m *MockServer) URL() string {
	if m.server != nil {
		return m.server.URL
	}
	return ""
}

// handleRequest handles incoming requests and serves appropriate mock responses
func (m *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Parse the URL path to match GitHub API format: /repos/{owner}/{repo}/{resource}/{number}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	// We expect paths in the format: /repos/{owner}/{repo}/{resourceType}/{number}
	if len(parts) < 5 || parts[0] != "repos" {
		http.Error(w, "Invalid request path", http.StatusNotFound)
		return
	}

	owner := parts[1]
	repo := parts[2]
	resourceType := parts[3] // "issues", "pulls", or "discussions"
	number := parts[4]

	// Mock different GitHub API endpoints
	switch resourceType {
	case "issues", "pulls", "discussions", "pull":
		m.handleResourceRequest(w, r, owner, repo, number, resourceType)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (m *MockServer) handleResourceRequest(w http.ResponseWriter, r *http.Request, owner, repo, number, resourceType string) {
	// Simulate different error scenarios based on query parameters
	errorScenario := r.URL.Query().Get("error")

	switch errorScenario {
	case "E002": // Network error scenario
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"message": "Network error", "documentation_url": "https://docs.github.com"}`)
	case "E003": // Not found (404) scenario
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"message": "Not Found", "documentation_url": "https://docs.github.com"}`)
	case "E004": // Permission error (403) scenario
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, `{"message": "Resource protected by organization SAMR policy", "documentation_url": "https://docs.github.com"}`)
	case "E005": // Rate limit exceeded (403) scenario
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, `{"message": "API rate limit exceeded for user ID 123456789", "documentation_url": "https://docs.github.com"}`)
	case "E006": // Not modified (304) scenario
		w.WriteHeader(http.StatusNotModified)
		fmt.Fprintf(w, "")
	default:
		// Success scenario - return appropriate mock data
		m.serveSuccessResponse(w, r, resourceType)
	}
}

// serveSuccessResponse serves appropriate mock data for different resource types
func (m *MockServer) serveSuccessResponse(w http.ResponseWriter, r *http.Request, resourceType string) {
	switch resourceType {
	case "issues":
		m.serveMockIssue(w, r)
	case "pulls", "pull":
		m.serveMockPullRequest(w, r)
	case "discussions":
		m.serveMockDiscussion(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

// TableDrivenTestCases represents different test scenarios for the mock server
// These are used for testing various error conditions and success cases
var TableDrivenTestCases = []struct {
	name           string
	expectedStatus int
	endpoint       string
	setupFunc      func(*MockServer)
}{{
		name:           "Success Case",
		expectedStatus: http.StatusOK,
		endpoint:       "/issues/testowner/testrepo/42",
		setupFunc: func(s *MockServer) {
			// Default success case setup
		},
	}, {
		name:           "E003 Not Found",
		expectedStatus: http.StatusNotFound,
		endpoint:       "/issues/testowner/testrepo/999?error=E003",
		setupFunc:      nil,
	}, {
		name:           "E004 Permission Error",
		expectedStatus: http.StatusForbidden,
		endpoint:       "/pulls/testowner/testrepo/1?error=E004",
		setupFunc:      nil,
	}, {
		name:           "E005 Rate Limit Exceeded",
		expectedStatus: http.StatusForbidden,
		endpoint:       "/discussions/testowner/testrepo/1?error=E005",
		setupFunc:      nil,
	}, {
		name:           "E002 Network Error",
		expectedStatus: http.StatusInternalServerError,
		endpoint:       "/issues/testowner/testrepo/42?error=E002",
		setupFunc:      nil,
	}}

// Mock data structures
var mockIssue = map[string]interface{}{
	"id":          123456789,
	"number":      42,
	"title":       "Test Issue",
	"state":       "open",
	"body":        "This is a test issue created for mock purposes",
	"created_at":  "2023-01-15T10:30:00Z",
	"updated_at":  "2023-01-16T09:15:00Z",
	"closed_at":   nil,
	"user": map[string]interface{}{
		"login":    "testuser",
		"id":       123456,
		"avatar_url": "https://avatars.githubusercontent.com/u/123456",
	},
	"comments": 5,
	"html_url": "https://api.github.com/repos/testowner/testrepo/issues/42",
}

var mockPullRequest = map[string]interface{}{
	"id":          987654321,
	"number":      17,
	"title":       "Test Pull Request",
	"state":       "open",
	"body":        "This is a test pull request for mocking purposes",
	"created_at":  "2023-01-14T14:20:00Z",
	"updated_at":  "2023-01-15T13:45:00Z",
	"closed_at":   nil,
	"merged":      false,
	"merged_at":   nil,
	"user": map[string]interface{}{
		"login":    "testuser",
		"id":       789012,
		"avatar_url": "https://avatars.githubusercontent.com/u/789012",
	},
	"comments": 3,
	"html_url": "https://api.github.com/repos/testowner/testrepo/pulls/17",
}

var mockDiscussion = map[string]interface{}{
	"id":          567890123,
	"number":      8,
	"title":       "Test Discussion",
	"state":       "open",
	"body":        "This is a test discussion thread",
	"created_at":  "2023-01-13T12:00:00Z",
	"updated_at":  "2023-01-14T11:30:00Z",
	"closed_at":   nil,
	"user": map[string]interface{}{
		"login":    "testuser",
		"id":       345678,
		"avatar_url": "https://avatars.githubusercontent.com/u/345678",
	},
	"comments": 7,
	"html_url": "https://api.github.com/repos/testowner/testrepo/discussions/8",
}

func (m *MockServer) serveMockIssue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mockIssue)
}

func (m *MockServer) serveMockPullRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mockPullRequest)
}

func (m *MockServer) serveMockDiscussion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mockDiscussion)
}