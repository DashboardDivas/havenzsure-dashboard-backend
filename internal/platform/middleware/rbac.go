// Source note:
// This file was partially generated / refactored with assistance from AI (ChatGPT).
// Edited and reviewed by AN-NI HUANG
// Date: 2025-11-22
package middleware

import (
	"context"
	"net/http"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/platform/auth"
	"github.com/google/uuid"
)

type contextKey string

const shopIDKey contextKey = "shopID"

// RequireRole checks if user has any of the specified roles
// Usage:
//
//	r.Use(middleware.RequireRole("superadmin", "admin"))
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authUser, err := auth.GetAuthUser(r.Context())
			if err != nil {
				// Should not happen (already passed auth middleware)
				writeAuthError(w, http.StatusUnauthorized, err)
				return
			}

			// Check role
			if !authUser.HasRole(allowedRoles...) {
				writeAuthError(w, http.StatusForbidden, auth.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireSuperAdmin requires SuperAdmin role
// NOTE: Currently unused. No route for only superadmin yet, but provided for future use
func RequireSuperAdmin() func(http.Handler) http.Handler {
	return RequireRole("superadmin")
}

// RequireAdminOrAbove requires Admin or SuperAdmin
func RequireAdminOrAbove() func(http.Handler) http.Handler {
	return RequireRole("superadmin", "admin")
}

// EnforceShopScope enforces shop-scoped access for non-admin users
// SuperAdmin bypass this check; other roles must have a shop assigned.
// The user's shop ID is injected into context for downstream handlers to filter by shop.
func EnforceShopScope() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authUser, err := auth.GetAuthUser(r.Context())
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, err)
				return
			}

			// Only SuperAdmin is not restricted
			if authUser.IsSuperAdmin() {
				next.ServeHTTP(w, r)
				return
			}

			// Other roles must have shop
			if !authUser.HasShop() {
				writeAuthError(w, http.StatusForbidden, auth.ErrNoShopAssignment)
				return
			}

			// Inject shopID into context for downstream handlers
			ctx := context.WithValue(r.Context(), shopIDKey, *authUser.ShopID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetShopIDFromContext retrieves shop ID from context
// Returns (shopID, true) if shop scope enforced, (uuid.Nil, false) otherwise
// NOTE: Currently unused - will be used in work order service layer for filtering by shop
func GetShopIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	shopID, ok := ctx.Value(shopIDKey).(uuid.UUID)
	return shopID, ok
}
