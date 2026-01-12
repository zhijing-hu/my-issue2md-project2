// Package parser provides URL parsing functionality for GitHub resources
package parser

import (
	"testing"
)

// TestParser_Parse tests the Parse functionality with various URL scenarios
func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name         string
		inputURL     string
		expectError  bool
		expectResult *Resource
	}{
		// Test Case 1: Valid Issue URL
		{
			name:        "valid issue URL",
			inputURL:    "https://github.com/owner/repo/issues/123",
			expectError: false,
			expectResult: &Resource{
				Owner:       "owner",
				Repo:        "repo",
				ResourceType: TypeIssue,
				Number:      123,
				RawURL:      "https://github.com/owner/repo/issues/123",
			},
		},
		// Test Case 2: Valid Pull Request URL
		{
			name:        "valid pull request URL",
			inputURL:    "https://github.com/owner/repo/pull/456",
			expectError: false,
			expectResult: &Resource{
				Owner:       "owner",
				Repo:        "repo",
				ResourceType: TypePull,
				Number:      456,
				RawURL:      "https://github.com/owner/repo/pull/456",
			},
		},
		// Test Case 3: Valid Discussion URL
		{
			name:        "valid discussion URL",
			inputURL:    "https://github.com/owner/repo/discussions/789",
			expectError: false,
			expectResult: &Resource{
				Owner:       "owner",
				Repo:        "repo",
				ResourceType: TypeDiscussion,
				Number:      789,
				RawURL:      "https://github.com/owner/repo/discussions/789",
			},
		},
		// Test Case 4: Invalid URL (malformed)
		{
			name:        "invalid malformed URL",
			inputURL:    "not-a-valid-url",
			expectError: true,
			expectResult: nil,
		},
		// Test Case 5: Unsupported URL type (repository homepage)
		{
			name:        "unsupported repository URL",
			inputURL:    "https://github.com/owner/repo",
			expectError: true,
			expectResult: nil,
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.inputURL)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return // If we expect an error, we don't check the result
			}

			// If we don't expect an error, err should be nil
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check the result
			if result == nil && tt.expectResult != nil {
				t.Errorf("Expected non-nil result but got nil")
				return
			}

			if result != nil && tt.expectResult != nil {
				if result.Owner != tt.expectResult.Owner {
					t.Errorf("Expected Owner %q but got %q", tt.expectResult.Owner, result.Owner)
				}
				if result.Repo != tt.expectResult.Repo {
					t.Errorf("Expected Repo %q but got %q", tt.expectResult.Repo, result.Repo)
				}
				if result.ResourceType != tt.expectResult.ResourceType {
					t.Errorf("Expected ResourceType %q but got %q", tt.expectResult.ResourceType, result.ResourceType)
				}
				if result.Number != tt.expectResult.Number {
					t.Errorf("Expected Number %d but got %d", tt.expectResult.Number, result.Number)
				}
				if result.RawURL != tt.expectResult.RawURL {
					t.Errorf("Expected RawURL %q but got %q", tt.expectResult.RawURL, result.RawURL)
				}
			}
		})
	}
}

// TestParser_Validate tests the Validate functionality
func TestParser_Validate(t *testing.T) {
	// There's a stub implementation, so we just test that it can be called
	// without panic
	parser := NewParser()

	// Test with some URLs - should not panic
	testURLs := []string{
		"https://github.com/owner/repo/issues/123",
		"https://github.com/owner/repo/pull/456",
		"invalid-url",
		"",
	}

	for _, testURL := range testURLs {
		result := parser.Validate(testURL)
		// Just ensure it doesn't panic - stub should return false
		_ = result
	}
}