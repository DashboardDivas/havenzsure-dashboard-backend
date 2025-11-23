package auth

import "errors"

var (
	// ErrNoToken indicates no authorization token provided in request
	ErrNoToken = errors.New("no authorization token provided")

	// ErrInvalidToken indicates token format is invalid or verification failed
	ErrInvalidToken = errors.New("invalid authorization token")

	// ErrTokenExpired indicates token has expired
	ErrTokenExpired = errors.New("token expired")

	// ErrTokenRevoked indicates token has been revoked
	ErrTokenRevoked = errors.New("token revoked - please login again")

	// ErrUserNotFound indicates token is valid but user doesn't exist in DB
	ErrUserNotFound = errors.New("user not found")

	// ErrUserInactive indicates user account has been deactivated
	ErrUserInactive = errors.New("user account is inactive")

	// ErrForbidden indicates insufficient permissions for the operation
	ErrForbidden = errors.New("forbidden: insufficient permissions")

	// ErrNoShopAssignment indicates user not assigned to any shop (required for some operations)
	ErrNoShopAssignment = errors.New("user not assigned to any shop")
)
