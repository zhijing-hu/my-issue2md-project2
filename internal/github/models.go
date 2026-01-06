// Package github provides GitHub API client functionality
// This file contains the core data models for GitHub resources
package github

import "time"

// ResourceType defines the type of GitHub resource
type ResourceType string

// Enum values for different resource types
const (
	TypeIssue      ResourceType = "issues"
	TypePull       ResourceType = "pull"
	TypeDiscussion ResourceType = "discussions"
)

// Repository represents a GitHub repository
type Repository struct {
	Owner    string `json:"owner"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

// User represents a GitHub user
type User struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

// Label represents a GitHub label
type Label struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Branch represents a branch reference
type Branch struct {
	Label string `json:"label"`
	Ref   string `json:"ref"`
	SHA   string `json:"sha"`
}

// Category represents a discussion category
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Reactions represents reaction counts on issues, PRs, or comments
type Reactions struct {
	PlusOne int `json:"THUMBS_UP"`      // 👍
	MinusOne int `json:"THUMBS_DOWN"`   // 👎
	Laugh int `json:"LAUGH"`           // 😄
	Hurray int `json:"HOORAY"`         // 🎉
	Confused int `json:"CONFUSED"`      // 😕
	Heart int `json:"HEART"`           // ❤️
	Rocket int `json:"ROCKET"`         // 🚀
	Eyes int `json:"EYES"`             // 👀
}

// Comment represents a comment on an issue, PR, or discussion
type Comment struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	Author    User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Reactions Reactions `json:"reaction_groups,omitempty"`
}

// Issue represents a GitHub issue
type Issue struct {
	ID           int       `json:"id"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	State        string    `json:"state"` // "open" or "closed"
	Body         string    `json:"body,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Author       User      `json:"user"`
	Comments     []Comment `json:"comments,omitempty"`
	HTMLURL      string    `json:"html_url"`
	Repo         Repository `json:"repository"`
	Labels       []Label   `json:"labels,omitempty"`
	Reactions    Reactions `json:"reactions,omitempty"`
	CommentCount int       `json:"comments"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Issue     // Embedded Issue fields
	Head      Branch `json:"head"` // Source branch
	Base      Branch `json:"base"` // Target branch
	MergeCommit string `json:"merge_commit_sha,omitempty"`
}

// Discussion represents a GitHub discussion
type Discussion struct {
	Issue    // Embedded Issue fields
	Category Category `json:"category"`
}