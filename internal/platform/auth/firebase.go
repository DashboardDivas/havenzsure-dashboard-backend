// Source note:
// This file was partially generated / refactored with assistance from AI (ChatGPT).
// Edited and reviewed by AN-NI HUANG
// Date: 2025-11-22
package auth

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var (
	// Singleton pattern: share one Firebase App instance globally
	firebaseApp  *firebase.App // Firebase App instance
	firebaseAuth *auth.Client
	once         sync.Once
	initError    error
)

// InitFirebase initializes Firebase Admin SDK (GCIP)
// Uses GOOGLE_APPLICATION_CREDENTIALS env var pointing to service account JSON file
func InitFirebase(ctx context.Context) error {
	// Ensure singleton initialization
	// No matter how many times called, only initialize once
	once.Do(func() {
		// Load from file path
		credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credPath == "" {
			initError = fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS env var not set")
			return
		}

		log.Printf("Loading Firebase credentials from file: %s", credPath)

		// Initialize Firebase App with credentials
		opt := option.WithCredentialsFile(credPath)
		config := &firebase.Config{
			ProjectID: os.Getenv("FIREBASE_PROJECT_ID"),
		}

		app, err := firebase.NewApp(ctx, config, opt)
		if err != nil {
			initError = fmt.Errorf("failed to initialize Firebase app: %w", err)
			return
		}
		firebaseApp = app

		// Get Auth Client
		// This client will be used for verifying ID tokens and managing users
		authClient, err := app.Auth(ctx)
		if err != nil {
			initError = fmt.Errorf("failed to get Firebase auth client: %w", err)
			return
		}
		firebaseAuth = authClient

		log.Println("Firebase/GCIP initialized successfully")
	})

	return initError
}

// GetAuthClient returns Firebase Auth Client
// Must call InitFirebase() first
func GetAuthClient() (*auth.Client, error) {
	if firebaseAuth == nil {
		return nil, fmt.Errorf("firebase not initialized. call InitFirebase() first")
	}
	return firebaseAuth, nil
}

// VerifyIDToken verifies Firebase ID Token
// This is the main function used by auth middleware
func VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	client, err := GetAuthClient()
	if err != nil {
		return nil, err
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return token, nil
}

// CreateFirebaseUserPasswordless creates a new user in Firebase/GCIP without password
// User will receive a password reset link to set their own password
// Parameters:
//   - email: User's email address
//   - firstName: User's first name
//   - lastName: User's last name
//
// Returns:
//   - uid: Firebase UID (external_id for database)
//   - error: Any error that occurred during creation
func CreateFirebaseUserPasswordless(ctx context.Context, email, firstName, lastName string) (uid string, err error) {
	client, err := GetAuthClient()
	if err != nil {
		return "", err
	}

	// Build display name
	displayName := firstName + " " + lastName

	// Create user WITHOUT password
	// Firebase will generate a random unusable password automatically
	params := (&auth.UserToCreate{}).
		Email(email).
		DisplayName(displayName).
		EmailVerified(false) // User must verify email when setting password

	userRecord, err := client.CreateUser(ctx, params)
	if err != nil {
		return "", err
	}

	log.Printf("Successfully created Firebase user (passwordless): %s (UID: %s)", email, userRecord.UID)
	return userRecord.UID, nil
}

// GeneratePasswordResetLink generates a password reset link for a user
// This link allows user to set their password for the first time
// Parameters:
//   - email: User's email address
//
// Returns:
//   - link: Password reset link (valid for 1 hour by default)
//   - error: Any error that occurred
func GeneratePasswordResetLink(ctx context.Context, email string) (link string, err error) {
	client, err := GetAuthClient()
	if err != nil {
		return "", err
	}

	// Generate password reset link
	// This link is valid for 1 hour by default
	link, err = client.PasswordResetLink(ctx, email)
	if err != nil {
		return "", fmt.Errorf("failed to generate password reset link: %w", err)
	}

	log.Printf("Successfully generated password reset link for: %s", email)
	return link, nil
}

// ResendPasswordResetLink resends password reset link to existing user
// Useful when user didn't receive the initial email or link expired
func ResendPasswordResetLink(ctx context.Context, email string) (link string, err error) {
	return GeneratePasswordResetLink(ctx, email)
}

// DeleteFirebaseUser deletes a user from Firebase/GCIP
// This is used for rollback when database insertion fails
func DeleteFirebaseUser(ctx context.Context, uid string) error {
	client, err := GetAuthClient()
	if err != nil {
		return err
	}

	if err := client.DeleteUser(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete Firebase user: %w", err)
	}

	log.Printf("Successfully deleted Firebase user: UID=%s", uid)
	return nil
}

// UpdateFirebasePassword updates a user's password in Firebase
// Used when user changes their password
// NOTE: Currently unused - reserved for future "Change Password" feature
func UpdateFirebasePassword(ctx context.Context, uid, newPassword string) error {
	client, err := GetAuthClient()
	if err != nil {
		return err
	}

	params := (&auth.UserToUpdate{}).Password(newPassword)
	_, err = client.UpdateUser(ctx, uid, params)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	log.Printf("Successfully updated password for UID: %s", uid)
	return nil
}

// DisableFirebaseUser disables a user account in Firebase
// Used when user is deactivated in the system
func DisableFirebaseUser(ctx context.Context, uid string) error {
	client, err := GetAuthClient()
	if err != nil {
		return err
	}

	params := (&auth.UserToUpdate{}).Disabled(true)
	_, err = client.UpdateUser(ctx, uid, params)
	if err != nil {
		return fmt.Errorf("failed to disable user: %w", err)
	}

	log.Printf("Successfully disabled Firebase user: UID=%s", uid)
	return nil
}

// EnableFirebaseUser enables a previously disabled user account
// Used when user is reactivated in the system
func EnableFirebaseUser(ctx context.Context, uid string) error {
	client, err := GetAuthClient()
	if err != nil {
		return err
	}

	params := (&auth.UserToUpdate{}).Disabled(false)
	_, err = client.UpdateUser(ctx, uid, params)
	if err != nil {
		return fmt.Errorf("failed to enable user: %w", err)
	}

	log.Printf("Successfully enabled Firebase user: UID=%s", uid)
	return nil
}

// SetEmailVerified updates the emailVerified flag for a Firebase/GCIP user.
func SetEmailVerified(ctx context.Context, uid string, verified bool) error {
	client, err := GetAuthClient()
	if err != nil {
		return err
	}

	params := (&auth.UserToUpdate{}).EmailVerified(verified)
	if _, err := client.UpdateUser(ctx, uid, params); err != nil {
		return fmt.Errorf("failed to update emailVerified: %w", err)
	}

	log.Printf("Successfully set emailVerified=%v for Firebase user: UID=%s", verified, uid)
	return nil
}

// SetCustomClaims sets custom claims on a Firebase user
// NOTE: Currently unused - reserved for future use (token version/ roles), not used in current MVP
func SetCustomClaims(ctx context.Context, uid string, claims map[string]interface{}) error {
	client, err := GetAuthClient()
	if err != nil {
		return err
	}

	if err := client.SetCustomUserClaims(ctx, uid, claims); err != nil {
		return fmt.Errorf("failed to set custom claims: %w", err)
	}

	log.Printf("Successfully set custom claims for UID: %s", uid)
	return nil
}
