// Package converter provides table-driven tests for conversion functionality
package converter

import (
	"testing"
	"time"

	"github.com/bigwhite/issue2md/internal/github"
)

// TestConverter_Creation tests basic converter creation
func TestConverter_Creation(t *testing.T) {
	t.Run("Create converter with default options", func(t *testing.T) {
		converter, err := makeConverter()
		if err != nil {
			t.Fatalf("Failed to create converter: %v", err)
		}
		if converter == nil {
			t.Error("Expected non-nil converter")
		}
	})

	t.Run("Create converter with reactions enabled", func(t *testing.T) {
		converter, err := makeConverter(WithReactions(true))
		if err != nil {
			t.Fatalf("Failed to create converter with reactions: %v", err)
		}
		if converter == nil {
			t.Error("Expected non-nil converter")
		}
	})
}

// TestConverter_ConvertIssue tests basic issue conversion
func TestConverter_ConvertIssue(t *testing.T) {
	converter, err := makeConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	issue := createTestIssue()
	result, err := converter.ConvertIssue(&issue)
	if err != nil {
		t.Errorf("ConvertIssue failed: %v", err)
		return
	}

	// Basic verification that output contains expected elements
	if result == "" {
		t.Error("Expected non-empty result from ConvertIssue")
	}

	// Check that key elements are present
	tests := []string{"Test Issue", "#42", "testuser", "open", "This is a test issue"}
	for _, expected := range tests {
		if !containsSubstring(result, expected) {
			t.Errorf("Expected result to contain %q, got: %s", expected, result)
		}
	}
}

// Helper functions for creating test data
func makeConverter(opts ...ConverterOption) (Converter, error) {
	// Create new converter instance directly
	converter := &converterImpl{
		templates:      make(map[string]string),
		enableReactions: false,
	}
	converter.loadDefaultTemplates()
	for _, opt := range opts {
		if err := opt(converter); err != nil {
			return nil, err
		}
	}
	return converter, nil
}

func createTestIssue() github.Issue {
	return github.Issue{
		Title:      "Test Issue",
		Number:     42,
		State:      "open",
		Body:       "This is a test issue",
		CreatedAt:  mockTime("2023-01-15T10:30:00Z"),
		UpdatedAt:  mockTime("2023-01-16T09:15:00Z"),
		Author:     github.User{Login: "testuser"},
		HTMLURL:    "https://github.com/testowner/testrepo/issues/42",
		Reactions:  github.Reactions{PlusOne: 5, Heart: 2}, // some reactions
	}
}

func mockTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}