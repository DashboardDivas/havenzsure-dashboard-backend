package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey string

const authUserKey contextKey = "authUser"

// AuthUser is the user information injected into request context
// Contains complete user data queried from DB
type AuthUser struct {
	ID           uuid.UUID  `json:"id"`
	Code         string     `json:"code"`
	Email        string     `json:"email"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	ExternalID   string     `json:"externalId"`   // GCIP UID
	RoleCode     string     `json:"roleCode"`     // "superadmin", "admin", "adjuster", "bodyman"
	RoleName     string     `json:"roleName"`     // "Super Administrator", etc.
	ShopID       *uuid.UUID `json:"shopId"`       // May be nil
	ShopCode     *string    `json:"shopCode"`     // May be nil
	TokenVersion int        `json:"tokenVersion"` // For token revocation
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
	return u.RoleCode == "superadmin"
}

// IsAdmin checks if user is Admin (includes SuperAdmin)
func (u *AuthUser) IsAdmin() bool {
	return u.RoleCode == "admin" || u.RoleCode == "superadmin"
}

// HasShop checks if user is assigned to a shop
func (u *AuthUser) HasShop() bool {
	return u.ShopID != nil
}
