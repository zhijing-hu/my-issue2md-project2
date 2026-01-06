// Package parser provides URL parsing functionality
// This file contains types for URL resource parsing
package parser

// ResourceType defines the type of GitHub resource
// This mirrors the types defined in github package but is defined here for separation of concerns
type ResourceType string

// Enum values for different resource types
const (
	TypeIssue      ResourceType = "issues"
	TypePull       ResourceType = "pull"
	TypeDiscussion ResourceType = "discussions"
)

// Resource represents a parsed GitHub resource from a URL
type Resource struct {
	Owner       string      // Repository owner
	Repo        string      // Repository name
	ResourceType ResourceType // Type of resource (issues, pull, discussions)
	Number      int         // Issue/PR/Discussion number
	RawURL      string      // Original input URL
}