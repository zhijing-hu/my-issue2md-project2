package github

import (
	"io"
	"net/http"
	"testing"
)

// TestMockServer verify that the mock server starts and responds correctly
func TestMockServer(t *testing.T) {
	mockServer := NewMockServer()
	mockServer.Start()
	defer mockServer.Close()

	if mockServer.URL() == "" {
		t.Fatal("Mock server URL should not be empty")
	}

	// Test that the server responds to requests
	testCases := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{"Success Issue Response", "/repos/testowner/testrepo/issues/42", http.StatusOK},
		{"Not Found Error", "/repos/testowner/testrepo/issues/999?error=E003", http.StatusNotFound},
		{"Permission Error", "/repos/testowner/testrepo/pulls/1?error=E004", http.StatusForbidden},
		{"Rate Limit Error", "/repos/testowner/testrepo/discussions/1?error=E005", http.StatusForbidden},
		{"Network Error", "/repos/testowner/testrepo/issues/42?error=E002", http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(mockServer.URL() + tc.path)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// For success cases, verify we get JSON response
			if tc.expectedStatus == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}

				if len(body) == 0 {
					t.Error("Expected non-empty response body")
				}
			}
		})
	}
}

// TestTableDrivenTestCases verifies the predefined table-driven test cases
func TestTableDrivenTestCases(t *testing.T) {
	if len(TableDrivenTestCases) != 5 {
		t.Fatalf("Expected 5 table-driven test cases, got %d", len(TableDrivenTestCases))
	}

	expectedNames := []string{
		"Success Case",
		"E003 Not Found",
		"E004 Permission Error",
		"E005 Rate Limit Exceeded",
		"E002 Network Error",
	}

	for i, expectedName := range expectedNames {
		if TableDrivenTestCases[i].name != expectedName {
			t.Errorf("Test case %d: expected name '%s', got '%s'", i, expectedName, TableDrivenTestCases[i].name)
		}
	}
}