package shop

import "errors"

// variable for error messages
var (
	ErrNotFound     = errors.New("shop not found")
	ErrConflict     = errors.New("shop already exists")
	ErrInvalidInput = errors.New("invalid shop input")
)
