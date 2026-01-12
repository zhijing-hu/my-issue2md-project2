// Package cli provides table-driven tests for argument parsing
package cli

import (
	"os"
	"testing"
)

// TestParseArgs tests command-line argument parsing functionality
func TestParseArgs(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedArgs   *Args
		expectError    bool
		expectedURL    string
		expectedOutput string
	}{
		{
			name: "Valid issue URL with output file",
			args: []string{"--output", "test.md", "https://github.com/owner/repo/issues/42"},
			expectedURL: "https://github.com/owner/repo/issues/42",
			expectedOutput: "test.md",
		},
		{
			name: "Valid pull request URL with shorthand output",
			args: []string{"-o", "pr-output.md", "https://github.com/owner/repo/pull/17"},
			expectedURL: "https://github.com/owner/repo/pull/17",
			expectedOutput: "pr-output.md",
		},
		{
			name: "Valid discussion URL with reactions enabled",
			args: []string{"--enable-reactions", "https://github.com/owner/repo/discussions/8"},
			expectedURL: "https://github.com/owner/repo/discussions/8",
		},
		{
			name: "Missing URL (should still parse without error)",
			args: []string{"--enable-reactions"},
			expectedURL: "",
		},
		{
			name: "Help flag",
			args: []string{"--help"},
			expectError: false,
		},
		{
			name: "Version flag",
			args: []string{"--version"},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original args to restore later
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set up test command-line arguments
			os.Args = append([]string{"issue2md"}, tc.args...)

			// Parse the arguments
			args, err := ParseArgs()
			if tc.expectError && err == nil {
				t.Error("Expected an error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tc.expectedURL != "" {
				if args.URL != tc.expectedURL {
					t.Errorf("Expected URL %q, got %q", tc.expectedURL, args.URL)
				}
			}

			if tc.expectedOutput != "" {
				if args.Output != tc.expectedOutput {
					t.Errorf("Expected output %q, got %q", tc.expectedOutput, args.Output)
				}
			}

			// Check specific flag values
			if tc.name == "Help flag" && !args.Help {
				t.Error("Expected Help flag to be true")
			}

			if tc.name == "Version flag" && !args.Version {
				t.Error("Expected Version flag to be true")
			}

			if tc.name == "Valid discussion URL with reactions enabled" && !args.EnableReactions {
				t.Error("Expected EnableReactions flag to be true")
			}
		})
	}
}

// TestValidateArgs tests argument validation logic
func TestValidateArgs(t *testing.T) {
	testCases := []struct {
		name        string
		args        *Args
		expectError bool
		errorSubstr string
	}{
		{
			name: "Valid issue URL",
			args: &Args{
				URL: "https://github.com/owner/repo/issues/42",
			},
			expectError: false,
		},
		{
			name: "Valid pull request URL",
			args: &Args{
				URL: "https://github.com/owner/repo/pull/17",
			},
			expectError: false,
		},
		{
			name: "Valid discussion URL",
			args: &Args{
				URL: "https://github.com/owner/repo/discussions/8",
			},
			expectError: false,
		},
		{
			name: "Missing URL",
			args: &Args{
				URL: "",
			},
			expectError: true,
			errorSubstr: "missing required parameter",
		},
		{
			name: "Invalid URL format",
			args: &Args{
				URL: "not-a-url",
			},
			expectError: true,
			errorSubstr: "invalid GitHub URL format",
		},
		{
			name: "Unrecognized GitHub URL format",
			args: &Args{
				URL: "https://github.com/owner/repo/invalid/123",
			},
			expectError: true,
			errorSubstr: "unrecognized GitHub URL format",
		},
		{
			name: "Help requested (no validation)",
			args: &Args{
				URL:    "",
				Help:   true,
			},
			expectError: false,
		},
		{
			name: "Version requested (no validation)",
			args: &Args{
				URL:     "",
				Version: true,
			},
			expectError: false,
		},
		{
			name: "Error codes requested (no validation)",
			args: &Args{
				URL:         "",
				ErrorCodes:  true,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if tc.expectError && err == nil {
				t.Error("Expected an error but got none")
				return
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tc.expectError && err != nil {
				if tc.errorSubstr != "" && !containsSubstring(err.Error(), tc.errorSubstr) {
					t.Errorf("Expected error to contain %q, got %q", tc.errorSubstr, err.Error())
				}
			}
		})
	}
}

// Helper function for substring check
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