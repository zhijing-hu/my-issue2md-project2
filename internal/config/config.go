// Package config provides configuration management
// This file contains functions for secure environment variable handling
package config

import (
	"fmt"
	"os"
)

// GetToken securely retrieves the GitHub token from environment variables
// This is the ONLY way tokens should be accessed according to the security requirements
func GetToken() (string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}
	// Return immediately, no caching (as per security requirement §6.2.1)
	return token, nil
}