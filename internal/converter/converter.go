// Package converter handles conversion of GitHub resources to markdown format
package converter

import (
	"fmt"
	"github.com/bigwhite/issue2md/internal/github"
)

// Converter defines the interface for converting GitHub entities to Markdown
// This interface supports conversion of Issues, Pull Requests, and Discussions
type Converter interface {
	// ConvertIssue converts a GitHub Issue into Markdown format
	ConvertIssue(issue *github.Issue) (string, error)

	// ConvertPullRequest converts a GitHub Pull Request into Markdown format
	ConvertPullRequest(pr *github.PullRequest) (string, error)

	// ConvertDiscussion converts a GitHub Discussion into Markdown format
	ConvertDiscussion(discussion *github.Discussion) (string, error)
}

// ConverterOption defines a function type for configuring the converter
type ConverterOption func(*converterImpl) error

// Converter implementation
const (
	mainTemplateName     = "main"
	commentTemplateName  = "comment"
	reactionTemplateName = "reaction"
)

// Main template for GitHub resource markdown output
// This template follows the specification in spec.md §2.2
const mainTemplate = `{{.Title}}
===================

**Issue #{{.Number}}** by **{{.Author}}** | **{{.State}}** | {{.CreatedAt}} | {{.UpdatedAt}}

{{if .Description}}
## Description

{{.Description}}

{{end}}

{{if .Comments}}
## Comments

{{range .Comments}}

### Comment by {{.Author}} | {{.CreatedAt}}

{{.Body}}

{{end}}

{{end}}

{{if .Reactions.enabled}}
## Reactions

{{range .Reactions}}
- {{.Content}}: {{.Count}}
{{end}}

{{end}}

{{if .URL}}
[View on GitHub]({{.URL}})

{{end}}
`

// Comment template for individual comment rendering
const commentTemplate = `### Comment by {{.Author}} | {{.CreatedAt}}

{{.Body}}

`

// Reaction template for reaction rendering
const reactionTemplate = "- {{.Content}}: {{.Count}}"

// converterImpl is the concrete implementation of the Converter interface
type converterImpl struct {
	templates      map[string]string
	enableReactions bool
}

// WithReactions enables reactions in the markdown output
func WithReactions(enable bool) ConverterOption {
	return func(c *converterImpl) error {
		c.enableReactions = enable
		return nil
	}
}

// NewConverter creates a new converter instance with provided options
type NewConverter func(opts ...ConverterOption) (Converter, error)

// Create creates a new converter instance with provided configuration options
func (n NewConverter) Create(opts ...ConverterOption) (Converter, error) {
	converter := &converterImpl{
		enableReactions: false, // default: reactions disabled
		templates:      make(map[string]string),
	}

	// Load default templates
	converter.loadDefaultTemplates()

	// Apply options
	for _, opt := range opts {
		if err := opt(converter); err != nil {
			return nil, err
		}
	}

	return converter, nil
}

// embellishReactions formats reactions map based on enableReactions flag
func (c *converterImpl) embellishReactions(reactions map[string]int, enabled bool) map[string]interface{} {
	result := make(map[string]interface{})

	if !enabled || reactions == nil {
		result["enabled"] = false
		result["items"] = []interface{}{}
		return result
	}

	// Convert reactions map to slice for template processing
	var reactionItems []map[string]interface{}
	for emoji, count := range reactions {
		reactionItems = append(reactionItems, map[string]interface{}{
			"Content": emoji,
			"Count":  count,
		})
	}

	result["enabled"] = true
	result["items"] = reactionItems
	return result
}

// embellishReactionsFromStruct formats Reactions struct based on enableReactions flag
func (c *converterImpl) embellishReactionsFromStruct(reactions github.Reactions, enabled bool) map[string]interface{} {
	result := make(map[string]interface{})

	if !enabled || reactions == (github.Reactions{}) {
		result["enabled"] = false
		result["items"] = []interface{}{}
		return result
	}

	// Convert Reactions struct to slice for template processing
	var reactionItems []map[string]interface{}

	// Map the reaction struct fields to emoji content
	reactionMap := map[string]string{
		"PlusOne": "👍",
		"MinusOne": "👎",
		"Laugh": "😄",
		"Hurray": "🎉",
		"Confused": "😕",
		"Heart": "❤️",
		"Rocket": "🚀",
		"Eyes": "👀",
	}

	for fieldName, emoji := range reactionMap {
		// Use reflection or direct field access to get the count
		var count int
		switch fieldName {
		case "PlusOne":
			count = reactions.PlusOne
		case "MinusOne":
			count = reactions.MinusOne
		case "Laugh":
			count = reactions.Laugh
		case "Hurray":
			count = reactions.Hurray
		case "Confused":
			count = reactions.Confused
		case "Heart":
			count = reactions.Heart
		case "Rocket":
			count = reactions.Rocket
		case "Eyes":
			count = reactions.Eyes
		}

		if count > 0 {
			reactionItems = append(reactionItems, map[string]interface{}{
			"Content": emoji,
			"Count":  count,
		})
		}
	}

	result["enabled"] = true
	result["items"] = reactionItems
	return result
}

// loadDefaultTemplates loads the default Markdown templates
func (c *converterImpl) loadDefaultTemplates() {
	c.templates[mainTemplateName] = mainTemplate
	c.templates[commentTemplateName] = commentTemplate
	c.templates[reactionTemplateName] = reactionTemplate
}

// ConvertIssue converts a GitHub Issue into Markdown format
func (c *converterImpl) ConvertIssue(issue *github.Issue) (string, error) {
	data := map[string]interface{}{
		"Title":       issue.Title,
		"Number":      issue.Number,
		"Author":      issue.Author.Login,
		"State":       issue.State,
		"CreatedAt":   issue.CreatedAt,
		"UpdatedAt":   issue.UpdatedAt,
		"Description": issue.Body,
		"URL":         issue.HTMLURL,
		"Comments":    c.convertComments(issue.Comments),
		"Reactions":   c.embellishReactionsFromStruct(issue.Reactions, c.enableReactions),
	}

	// Simple implementation - return formatted markdown
	// In a real implementation, you would use proper template rendering
	return fmt.Sprintf("# %s\n\nIssue #%d by %s | %s | %s | %s\n\n%s\n\n%s",
		data["Title"], data["Number"], data["Author"], data["State"], data["CreatedAt"], data["UpdatedAt"],
		data["Description"], data["URL"]), nil
}

// ConvertPullRequest converts a GitHub Pull Request into Markdown format
func (c *converterImpl) ConvertPullRequest(pr *github.PullRequest) (string, error) {
	data := map[string]interface{}{
		"Title":       pr.Title,
		"Number":      pr.Number,
		"Author":      pr.Author.Login,
		"State":       pr.State,
		"CreatedAt":   pr.CreatedAt,
		"UpdatedAt":   pr.UpdatedAt,
		"Description": pr.Body,
		"URL":         pr.HTMLURL,
		"Comments":    c.convertComments(pr.Comments),
		"Reactions":   c.embellishReactionsFromStruct(pr.Reactions, c.enableReactions),
	}

	return fmt.Sprintf("# %s\n\nPull Request #%d by %s | %s | %s | %s\n\n%s\n\n%s",
		data["Title"], data["Number"], data["Author"], data["State"], data["CreatedAt"], data["UpdatedAt"],
		data["Description"], data["URL"]), nil
}

// ConvertDiscussion converts a GitHub Discussion into Markdown format
func (c *converterImpl) ConvertDiscussion(discussion *github.Discussion) (string, error) {
	data := map[string]interface{}{
		"Title":       discussion.Title,
		"Number":      discussion.Number,
		"Author":      discussion.Author.Login,
		"State":       discussion.State,
		"CreatedAt":   discussion.CreatedAt,
		"UpdatedAt":   discussion.UpdatedAt,
		"Description": discussion.Body,
		"URL":         discussion.HTMLURL,
		"Comments":    c.convertComments(discussion.Comments),
		"Reactions":   c.embellishReactionsFromStruct(discussion.Reactions, c.enableReactions),
	}

	return fmt.Sprintf("# %s\n\nDiscussion #%d by %s | %s | %s | %s\n\n%s\n\n%s",
		data["Title"], data["Number"], data["Author"], data["State"], data["CreatedAt"], data["UpdatedAt"],
		data["Description"], data["URL"]), nil
}

// convertComments converts comments into markdown format
func (c *converterImpl) convertComments(comments []github.Comment) []map[string]interface{} {
	var result []map[string]interface{}
	if comments == nil {
		return result
	}

	for _, comment := range comments {
		result = append(result, map[string]interface{}{
			"Author":    comment.Author.Login,
			"CreatedAt": comment.CreatedAt,
			"Body":      comment.Body,
		})
	}

	return result
}