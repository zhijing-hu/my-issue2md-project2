// Package github provides GitHub API client functionality
// This file contains error definitions for the GitHub client
package github

import (
	"fmt"
)

// ErrorCode represents the custom error codes defined in the specification
type ErrorCode string

// Error codes as defined in spec.md §2.4
const (
	ErrInvalidURL    ErrorCode = "E001" // Invalid URL
	ErrNetworkFailure ErrorCode = "E002" // Network error
	ErrNotFound       ErrorCode = "E003" // 404 Not Found
	ErrNoPermission   ErrorCode = "E004" // 403 Forbidden
	ErrInvalidToken   ErrorCode = "E005" // Invalid token
	ErrFileWrite      ErrorCode = "E006" // File write failure
)

// APIError represents a custom error with error code and context
type APIError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
}

// Unwrap implements the-error wrapping interface
func (e *APIError) Unwrap() error {
	return e.Cause
}

// NewAPIError creates a new APIError with the given code, message, and cause
func NewAPIError(code ErrorCode, message string, cause error) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}