// Package parser provides URL parsing functionality for GitHub resources
package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser interface defines methods for parsing GitHub URLs
type Parser interface {
	// Parse separates URL into components
	// Parameters: url string - complete GitHub URL
	// Returns: (*Resource, error) - parsed resource identifier
	// Error: E001 (invalid URL)
	Parse(url string) (*Resource, error)

	// Validate checks URL format validity
	// Parameters: url string
	// Returns: bool - true if valid GitHub resource URL
	Validate(url string) bool
}

// NewParser creates a new Parser implementation
type githubParser struct{}

// NewParser returns a new Parser instance
func NewParser() Parser {
	return &githubParser{}
}

// Parse implements the Parse method for Parser interface
func (p *githubParser) Parse(urlStr string) (*Resource, error) {
	// Use helper function for URL splitting and basic validation
	parts, err := p.splitAndValidatePath(urlStr)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	// Extract components from URL parts
	owner := parts[3]  // github.com/owner...
	repo := parts[4]   // github.com/owner/repo...
	resourceType := ResourceType(parts[5]) // issues/pull/discussions
	numberStr := parts[6] // issue/PR/discussion number

	// Convert number to int
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return nil, fmt.Errorf("invalid resource number: %w", err)
	}

	// Validate resource type
	switch resourceType {
	case TypeIssue, TypePull, TypeDiscussion:
		// Valid types
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	// Return parsed resource
	return &Resource{
		Owner:       owner,
		Repo:        repo,
		ResourceType: resourceType,
		Number:      number,
		RawURL:      urlStr,
	}, nil
}

// splitAndValidatePath splits URL and validates basic GitHub URL structure
func (p *githubParser) splitAndValidatePath(urlStr string) ([]string, error) {
	// Split the URL by "/" to extract components
	parts := strings.Split(urlStr, "/")

	// Check minimum required parts for valid GitHub resource URLs
	// Expected format: https://github.com/owner/repo/type/number
	if len(parts) < 7 {
		return nil, fmt.Errorf("invalid GitHub URL format")
	}

	// Validate that we got to github.com
	if parts[2] != "github.com" {
		return nil, fmt.Errorf("invalid GitHub URL format")
	}

	// Check URL has https://
	if parts[0] != "https:" || parts[1] != "" {
		return nil, fmt.Errorf("invalid GitHub URL format")
	}

	return parts, nil
}

// Validate implements the Validate method for Parser interface
func (p *githubParser) Validate(urlStr string) bool {
	// Stub implementation
	return false
}