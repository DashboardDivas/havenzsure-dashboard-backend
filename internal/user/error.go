package user

import (
	"errors"
	"fmt"
)

// Domain-level errors for user operations
var (
	ErrNotFound     = errors.New("user not found")
	ErrConflict     = errors.New("user conflict")
	ErrInvalidInput = errors.New("invalid input")
)

// ValidationError represents validation errors with specific field information
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Unwrap allows errors.Is to work with ValidationError
func (e *ValidationError) Unwrap() error {
	return ErrInvalidInput
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
