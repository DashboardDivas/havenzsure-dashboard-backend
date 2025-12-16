// Package auth provides request-level identity models (AuthUser) and context helpers.
// It does NOT contain Firebase logic (see firebase.go) or authorization middleware.

package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey string

const authUserKey contextKey = "authUser"

// Canonical role codes used across the system.
const (
	RoleSuperAdmin = "superadmin"
	RoleAdmin      = "admin"
	RoleAdjuster   = "adjuster"
	RoleBodyman    = "bodyman"
)

// AuthUser is the user information injected into request context
// Contains limited fields needed for authorization and auditing
type AuthUser struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	ExternalID   string     `json:"externalId"`       // GCIP UID
	RoleCode     string     `json:"roleCode"`         // "superadmin", "admin", "adjuster", "bodyman"
	ShopID       *uuid.UUID `json:"shopId,omitempty"` // Nullable
	TokenVersion int        `json:"tokenVersion"`     // For token revocation
	IsActive     bool       `json:"isActive"`
}

// SetAuthUser injects AuthUser into context
func SetAuthUser(ctx context.Context, user *AuthUser) context.Context {
	return context.WithValue(ctx, authUserKey, user)
}

// GetAuthUser retrieves AuthUser from context
// Returns error if not found
func GetAuthUser(ctx context.Context) (*AuthUser, error) {
	user, ok := ctx.Value(authUserKey).(*AuthUser)
	if !ok || user == nil {
		return nil, errors.New("no authenticated user in context")
	}
	return user, nil
}

// HasRole checks if user has any of the specified roles
func (u *AuthUser) HasRole(roles ...string) bool {
	for _, role := range roles {
		if u.RoleCode == role {
			return true
		}
	}
	return false
}

// IsSuperAdmin checks if user is SuperAdmin
func (u *AuthUser) IsSuperAdmin() bool {
	return u.RoleCode == RoleSuperAdmin
}

// IsAdminOnly checks if user is exactly Admin (does NOT include SuperAdmin).
func (u *AuthUser) IsAdminOnly() bool {
	return u.RoleCode == RoleAdmin
}

// IsAdminOrAbove checks if user is Admin or SuperAdmin.
func (u *AuthUser) IsAdminOrAbove() bool {
	return u.RoleCode == RoleAdmin || u.RoleCode == RoleSuperAdmin
}

// IsStaff checks if user is a non-admin staff role (adjuster/bodyman).
func (u *AuthUser) IsStaff() bool {
	return u.RoleCode == RoleAdjuster || u.RoleCode == RoleBodyman
}

// HasShop checks if user is assigned to a shop
func (u *AuthUser) HasShop() bool {
	return u.ShopID != nil
}
