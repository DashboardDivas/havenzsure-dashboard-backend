package shop

import (
	"errors"
	"fmt"
)

// variable for error messages
var (
	ErrNotFound     = errors.New("shop not found")
	ErrConflict     = errors.New("shop already exists")
	ErrInvalidInput = errors.New("invalid shop input")
)

// ValidationError represents an error due to invalid input data.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface for ValidationError
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Unwrap allows errors.Is to work with ValidationError and match ErrInvalidInput
func (e *ValidationError) Unwrap() error {
	return ErrInvalidInput
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
