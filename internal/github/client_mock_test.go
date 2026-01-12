package github

import (
	"net/http"
	"testing"
)

// helper function to create clients
func createClient(opts ...ClientOption) (Client, error) {
	var newClient NewClient = func(o ...ClientOption) (Client, error) {
		client := &githubClient{
			httpClient: http.DefaultClient,
			baseURL:    "https://api.github.com",
		}

		// Apply options
		for _, opt := range o {
			if err := opt(client); err != nil {
				// Preserve the original error code if it's already an APIError
				if apiErr, ok := err.(*APIError); ok {
					return nil, NewAPIError(apiErr.Code, "Failed to configure client", apiErr.Unwrap())
				}
				return nil, NewAPIError(ErrInvalidToken, "Failed to configure client", err)
			}
		}

		return client, nil
	}

	return newClient(opts...)
}

// createClientWithError creates a client that will append error query parameters
func createClientWithError(baseURL string, errorCode string) (Client, error) {
	// Create a custom http client that appends error query parameter
	customClient := &http.Client{
		Transport: &errorInjectingTransport{
			baseTransport: http.DefaultTransport,
			errorCode:     errorCode,
		},
	}
	return createClient(WithBaseURL(baseURL), WithHTTPClient(customClient))
}

// errorInjectingTransport is a custom RoundTripper that adds error query params
type errorInjectingTransport struct {
	baseTransport http.RoundTripper
	errorCode     string
}

func (t *errorInjectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Parse the URL to add query parameter
	if req.URL != nil {
		query := req.URL.Query()
		query.Set("error", t.errorCode)
		req.URL.RawQuery = query.Encode()
	}
	return t.baseTransport.RoundTrip(req)
}

// TestClientWithMockServer verifies that the client can work with our mock server
func TestClientWithMockServer(t *testing.T) {
	// Start the mock server
	mockServer := NewMockServer()
	mockServer.Start()
	defer mockServer.Close()

	// Create a client that uses the mock server
	client, err := createClient(WithBaseURL(mockServer.URL()))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test GetIssue
	t.Run("GetIssue Success", func(t *testing.T) {
		issue, err := client.GetIssue("testowner", "testrepo", 42)
		if err != nil {
			t.Fatalf("GetIssue failed: %v", err)
		}
		// We can't verify the contents without specific parsing, but we can check that we got a response
		if issue == nil {
			t.Error("Expected non-nil issue")
		}
	})

	// Test GetIssue with Not Found error
	t.Run("GetIssue Not Found", func(t *testing.T) {
		// Create a client that will inject error=E003 to trigger not found
		errorClient, err := createClientWithError(mockServer.URL(), "E003")
		if err != nil {
			t.Fatalf("Failed to create error client: %v", err)
		}

		_, err = errorClient.GetIssue("testowner", "testrepo", 42)
		if err == nil {
			t.Error("Expected error for not found issue")
		} else {
			apiErr, ok := err.(*APIError)
			if !ok {
				t.Errorf("Expected APIError, got %T", err)
			} else if apiErr.Code != ErrNotFound {
				t.Errorf("Expected error code E003, got %s", apiErr.Code)
			}
		}
	})

	// Test GetPullRequest
	t.Run("GetPullRequest Success", func(t *testing.T) {
		pr, err := client.GetPullRequest("testowner", "testrepo", 17)
		if err != nil {
			t.Fatalf("GetPullRequest failed: %v", err)
		}
		if pr == nil {
			t.Error("Expected non-nil PR")
		}
	})

	// Test error handling with mock server
	t.Run("Handle Forbidden Error", func(t *testing.T) {
		// Use proper GitHub API format: /repos/owner/repo/issues/number
		mockResp := mockServer.URL() + "/repos/testowner/testrepo/issues/42?error=E004"
		resp, err := http.Get(mockResp)
		if err != nil {
			t.Fatalf("Failed to create mock forbidden request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected 403 status code, got %d", resp.StatusCode)
		}
	})
}

// Simple implementation check to ensure the client works without mock
func TestClientCreation(t *testing.T) {
	// Test client creation with different options
	t.Run("Create Client Default", func(t *testing.T) {
		client, err := createClient()
		if err != nil {
			t.Fatalf("Failed to create default client: %v", err)
		}
		if client == nil {
			t.Error("Expected non-nil client")
		}
	})

	t.Run("Create Client WithToken", func(t *testing.T) {
		client, err := createClient(WithToken("test-token"))
		if err != nil {
			t.Fatalf("Failed to create client with token: %v", err)
		}
		if client == nil {
			t.Error("Expected non-nil client")
		}
	})

	t.Run("Create Client Invalid BaseURL", func(t *testing.T) {
		client, err := createClient(WithBaseURL(":invalid:"))
		if err == nil {
			t.Error("Expected error for invalid base URL")
		} else {
			apiErr, ok := err.(*APIError)
			if !ok {
				t.Errorf("Expected APIError, got %T", err)
			} else if apiErr.Code != ErrInvalidURL {
				t.Errorf("Expected error code E001, got %s", apiErr.Code)
			}
		}
		if client != nil {
			t.Error("Expected nil client when creation fails")
		}
	})
}