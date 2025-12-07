// Source note:
// This file was partially generated / refactored with assistance from AI (ChatGPT).
// Edited and reviewed by AN-NI HUANG
// Date: 2025-11-22
package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/platform/auth"
	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/user"
)

// AuthMiddleware verifies Firebase/GCIP ID Token
type AuthMiddleware struct {
	userRepo user.Repository
}

// NewAuthMiddleware creates Auth middleware
func NewAuthMiddleware(userRepo user.Repository) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo: userRepo,
	}
}

// Verify is the main authentication middleware
// Flow:
//  1. Extract Bearer token from Header
//  2. Verify token signature with Firebase/GCIP
//  3. Query user data from DB
//  4. Check user status (active/inactive)
//  5. If first login, mark email as verified
//  6. Build AuthUser and inject into context
//  7. Update last_sign_in_at asynchronously
func (m *AuthMiddleware) Verify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 1. Extract Bearer token
		token, err := extractBearerToken(r)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, auth.ErrNoToken)
			return
		}

		// 2. Verify Firebase token signature
		firebaseToken, err := auth.VerifyIDToken(ctx, token)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, auth.ErrInvalidToken)
			return
		}

		// 3. Query user from DB (by external_id)
		dbUser, err := m.userRepo.GetByExternalID(ctx, firebaseToken.UID)
		if err != nil {
			// User may exist in GCIP but not in DB (should not happen)
			writeAuthError(w, http.StatusUnauthorized, auth.ErrUserNotFound)
			return
		}

		// 4. Check if user is deactivated
		if !dbUser.IsActive {
			writeAuthError(w, http.StatusUnauthorized, auth.ErrUserInactive)
			return
		}

		// 5. If this is the first successful login, mark email as verified.
		if !dbUser.EmailVerified {
			go func(u *user.User) {
				ctx := context.Background()

				// Set emailVerified in GCIP
				if err := auth.SetEmailVerified(ctx, u.ExternalID, true); err != nil {
					log.Printf("failed to set GCIP emailVerified for %s: %v", u.ExternalID, err)
				}

				// Update DB record
				if err := m.userRepo.MarkEmailVerified(ctx, u.ID); err != nil {
					log.Printf("failed to mark DB email_verified for %s: %v", u.ID, err)
				}
			}(dbUser)
		}

		// 6. Build AuthUser and inject into context
		authUser := &auth.AuthUser{
			ID:           dbUser.ID,
			Email:        dbUser.Email,
			ExternalID:   dbUser.ExternalID,
			RoleCode:     dbUser.Role.Code,
			ShopID:       dbUser.ShopID,
			TokenVersion: dbUser.TokenVersion,
			IsActive:     dbUser.IsActive,
		}

		// Inject into context
		ctx = auth.SetAuthUser(ctx, authUser)

		// 7. Update last_sign_in_at (asynchronous, non-blocking)
		go m.userRepo.TouchLastSignInNowByExternalID(context.Background(), dbUser.ExternalID)

		// Continue processing request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractBearerToken extracts token from Authorization header
// Format: "Authorization: Bearer <token>"
func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", auth.ErrNoToken
	}

	// Check format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", auth.ErrInvalidToken
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", auth.ErrInvalidToken
	}

	return token, nil
}

// writeAuthError returns unified authentication error format
func writeAuthError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
