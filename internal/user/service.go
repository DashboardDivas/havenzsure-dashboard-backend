package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	firabaseAuth "firebase.google.com/go/v4/auth"
	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/platform/auth"
	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/shop"
	"github.com/google/uuid"
)

// Regular expressions for validation
var (
	emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^[0-9]{3}-[0-9]{3}-[0-9]{4}$`)
	imageRegex = regexp.MustCompile(`^https?://`)
)

// -------------------- Service Structs & Interfaces -------------------- //
// EmailSender defines how the user service sends emails.
// This allows us to swap implementations later (Gmail SMTP, SendGrid, etc.).
type EmailSender interface {
	// SendWelcomePasswordSetup sends the initial "welcome + set password" email.
	SendWelcomePasswordSetup(ctx context.Context, email, firstName, link string) error

	// SendPasswordSetupReminder sends a reminder email with the same reset link.
	SendPasswordSetupReminder(ctx context.Context, email, firstName, link string) error
}

// logEmailSender is the default implementation used in dev/tests.
// It only logs the email contents instead of actually sending.
type logEmailSender struct{}

func (s *logEmailSender) SendWelcomePasswordSetup(ctx context.Context, email, firstName, link string) error {
	log.Printf("[DEV] Would send WELCOME email to %s (name=%s) with link: %s", email, firstName, link)
	return nil
}

func (s *logEmailSender) SendPasswordSetupReminder(ctx context.Context, email, firstName, link string) error {
	log.Printf("[DEV] Would send REMINDER email to %s (name=%s) with link: %s", email, firstName, link)
	return nil
}

// UserService defines business logic for users
type UserService interface {
	CreateUser(ctx context.Context, actor *auth.AuthUser, in *CreateUserInput) (*User, error)
	GetUserByID(ctx context.Context, actor *auth.AuthUser, id uuid.UUID) (*User, error)
	GetUserByExternalID(ctx context.Context, externalID string) (*User, error)
	ListUsers(ctx context.Context, actor *auth.AuthUser, limit, offset int) ([]*User, error)
	UpdateUser(ctx context.Context, actor *auth.AuthUser, id uuid.UUID, in *UpdateUserInput) (*User, error)
	DeactivateUser(ctx context.Context, actor *auth.AuthUser, id uuid.UUID) error
	ReactivateUser(ctx context.Context, actor *auth.AuthUser, id uuid.UUID) error
	ResendPasswordSetupLink(ctx context.Context, actor *auth.AuthUser, userID uuid.UUID) error
}

type service struct {
	repo        Repository
	shopService shop.ShopService
	emailSender EmailSender
}

// -------------------- Service Constructors -------------------- //
// NewService constructs a UserService with default email sender (logs emails).
func NewService(repo Repository, shopSvc shop.ShopService,
) UserService {
	return &service{
		repo:        repo,
		shopService: shopSvc,
		emailSender: &logEmailSender{},
	}
}

// NewServiceWithEmailSender lets the caller inject a concrete EmailSender
// (e.g. Gmail SMTP, SendGrid, etc.)
func NewServiceWithEmailSender(repo Repository, shopSvc shop.ShopService, sender EmailSender) UserService {
	if sender == nil {
		sender = &logEmailSender{}
	}

	return &service{
		repo:        repo,
		shopService: shopSvc,
		emailSender: sender,
	}
}

// -------------------- Service Methods -------------------- //
// CreateUser creates a new user using PASSWORDLESS flow
// Flow:
//  1. Validate input and permissions
//  2. Create user in Firebase/GCIP (without password)
//  3. Get Firebase UID (external_id)
//  4. Save to database
//  5. Generate password reset link
//  6. Send welcome email with password setup link (async)
func (s *service) CreateUser(ctx context.Context, actor *auth.AuthUser, in *CreateUserInput) (*User, error) {
	if in == nil {
		return nil, ErrInvalidInput
	}

	// 1. Get current user (the one performing the creation)
	currentUser := actor
	if currentUser == nil {
		return nil, fmt.Errorf("unauthorized: no auth user in context")
	}

	// 2. Normalize input
	normalizeUser(in)

	// 3. Convert roleCode to roleID
	if in.RoleCode == "" {
		return nil, NewValidationError("roleCode", "cannot be blank")
	}
	roleID, err := s.repo.GetRoleIDByCode(ctx, in.RoleCode)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, NewValidationError("roleCode", "invalid role code")
		}
		return nil, fmt.Errorf("lookup roleCode: %w", err)
	}
	in.RoleID = roleID

	// 4. Permission check: Admin cannot create admin or superadmin and shop enforced
	if currentUser.IsAdminOnly() {
		if in.RoleCode == auth.RoleAdmin || in.RoleCode == auth.RoleSuperAdmin {
			return nil, NewValidationError("roleCode", "admin cannot create admin or superadmin users")
		}
	}

	// 5. Handle shop assignment based on current user's role
	switch {
	case currentUser.IsSuperAdmin():
		// SuperAdmin: Can assign any shop or none
		// If shopCode is provided, convert it to shopID
		if in.ShopCode != nil && *in.ShopCode != "" {
			shopID, err := s.shopService.GetShopIDByCode(ctx, *in.ShopCode)
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					return nil, NewValidationError("shopCode", "invalid shop code")
				}
				return nil, fmt.Errorf("lookup shopCode: %w", err)
			}
			in.ShopID = &shopID
		}

		// Non-superadmin users must be assigned to a shop
		if in.RoleCode != auth.RoleSuperAdmin && in.ShopID == nil {
			return nil, NewValidationError("shopCode", "non-superadmin users must be assigned to a shop")
		}

	case currentUser.IsAdminOnly():
		if !currentUser.HasShop() {
			return nil, fmt.Errorf("admin user is missing shop assignment")
		}

		// Admin: Can only assign users to their own shop
		// If shopCode is provided, must match their own shop
		if in.ShopCode != nil && *in.ShopCode != "" {
			shopID, err := s.shopService.GetShopIDByCode(ctx, *in.ShopCode)
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					return nil, NewValidationError("shopCode", "invalid shop code")
				}
				return nil, fmt.Errorf("lookup shopCode: %w", err)
			}

			if currentUser.ShopID == nil || *currentUser.ShopID != shopID {
				return nil, NewValidationError("shopCode", "admin can only assign users to their own shop")
			}

			in.ShopID = &shopID
		} else {
			// If shopCode is not provided, use the admin's own shop
			in.ShopID = currentUser.ShopID
		}
	default:
		// Should not happen: only admins/superadmins are allowed to hit this service.
		// Keep as a safety net in case routing or middleware is misconfigured.
		return nil, NewValidationError("permissions", "insufficient permissions to create users")
	}

	// 6. Validate input
	if err := validateUser(in); err != nil {
		return nil, err
	}

	// 7. Create user in Firebase/GCIP WITHOUT password (passwordless flow)
	firebaseUID, err := auth.CreateFirebaseUserPasswordless(
		ctx,
		in.Email,
		in.FirstName,
		in.LastName,
	)
	if err != nil {
		if firabaseAuth.IsEmailAlreadyExists(err) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("failed to create user in Firebase: %w", err)
	}

	// 8. Set ExternalID from Firebase (this is the Firebase UID)
	in.ExternalID = firebaseUID
	in.EmailVerified = false // This will be set to true when user first logs in

	// 9. Create user in database
	user, err := s.repo.Create(ctx, in)
	if err != nil {
		// IMPORTANT: Rollback - Delete Firebase user if DB creation fails
		if deleteErr := auth.DeleteFirebaseUser(ctx, firebaseUID); deleteErr != nil {
			log.Printf("ERROR: Failed to rollback Firebase user creation for UID %s: %v", firebaseUID, deleteErr)
		}
		return nil, fmt.Errorf("service create user: %w", err)
	}

	// 10. Generate password setup link
	passwordSetupLink, err := auth.GeneratePasswordResetLink(ctx, user.Email)
	if err != nil {
		log.Printf("WARNING: Failed to generate password setup link for %s: %v", user.Email, err)
		// Don't fail user creation, just log the error
		// Admin can resend the link later
	}

	// 11. Send welcome email with password setup link (async, non-blocking)
	if s.emailSender != nil && passwordSetupLink != "" {
		go func(email, firstName, link string) {
			// Use background context to prevent cancellation from HTTP request.
			emailCtx := context.Background()
			if err := s.emailSender.SendWelcomePasswordSetup(emailCtx, email, firstName, link); err != nil {
				log.Printf("Failed to send welcome email to %s: %v", email, err)
			}
		}(user.Email, user.FirstName, passwordSetupLink)
	} else {
		log.Printf("No email sender configured, skipping welcome email for %s", user.Email)
	}

	log.Printf("Successfully created user: %s (ID: %s, Firebase UID: %s)", user.Email, user.ID, firebaseUID)
	return user, nil
}

func (s *service) GetUserByID(ctx context.Context, actor *auth.AuthUser, id uuid.UUID) (*User, error) {
	// Get current user for permission check
	currentUser := actor
	if currentUser == nil {
		return nil, fmt.Errorf("unauthorized: no auth user in context")
	}

	// Get target user
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service get user by id: %w", err)
	}

	// Apply visibility rules
	if !s.canViewUser(currentUser, u) {
		// Hide existence of user
		return nil, fmt.Errorf("service get user by id: %w", err)
	}

	return u, nil
}

func (s *service) GetUserByExternalID(ctx context.Context, externalID string) (*User, error) {
	u, err := s.repo.GetByExternalID(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("service get user by external id: %w", err)
	}
	return u, nil
}

func (s *service) ListUsers(ctx context.Context, actor *auth.AuthUser, limit, offset int) ([]*User, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	list, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service list users: %w", err)
	}

	// Get current user
	currentUser := actor
	if currentUser == nil {
		return nil, fmt.Errorf("unauthorized: no auth user in context")
	}

	// Apply visibility rules
	if currentUser.IsAdminOnly() {
		filtered := make([]*User, 0, len(list))
		for _, u := range list {
			if !s.canViewUser(currentUser, u) {
				continue
			}
			filtered = append(filtered, u)
		}
		return filtered, nil
	}

	// SuperAdmin can see everyone
	return list, nil
}

func (s *service) UpdateUser(ctx context.Context, actor *auth.AuthUser, id uuid.UUID, in *UpdateUserInput) (*User, error) {
	if in == nil {
		return nil, ErrInvalidInput
	}
	in.EmailVerified = nil // Prevent emailVerified from being updated here

	// Get current user
	currentUser := actor
	if currentUser == nil {
		return nil, fmt.Errorf("unauthorized: no auth user in context")
	}

	// Get target user
	targetUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service get user: %w", err)
	}

	// Permission checks - split into two steps for clarity
	// 1. Check basic management permission (can current user manage target user?)
	if err := s.canManageUser(currentUser, targetUser); err != nil {
		return nil, err
	}

	// 2. Check field update permission (can these specific fields be modified?)
	if err := s.checkFieldUpdatePermission(currentUser, targetUser, in); err != nil {
		return nil, err
	}

	normalizeUpdateUser(in)

	// Convert roleCode to roleID if provided
	if in.RoleCode != nil && *in.RoleCode != "" {
		roleID, err := s.repo.GetRoleIDByCode(ctx, *in.RoleCode)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, NewValidationError("roleCode", "invalid role code")
			}
			return nil, fmt.Errorf("lookup roleCode: %w", err)
		}
		in.RoleID = &roleID
	}

	// Convert shopCode to shopID if provided
	if in.ShopCode != nil && *in.ShopCode != "" {
		shopID, err := s.shopService.GetShopIDByCode(ctx, *in.ShopCode)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, NewValidationError("shopCode", "invalid shop code")
			}
			return nil, fmt.Errorf("lookup shopCode: %w", err)
		}
		in.ShopID = &shopID
	}

	if err := validateUpdateUser(in); err != nil {
		return nil, err
	}

	user, err := s.repo.Update(ctx, id, in)
	if err != nil {
		return nil, fmt.Errorf("service update user: %w", err)
	}
	return user, nil
}

func (s *service) DeactivateUser(ctx context.Context, actor *auth.AuthUser, id uuid.UUID) error {
	// Get current user
	currentUser := actor
	if currentUser == nil {
		return fmt.Errorf("unauthorized: no auth user in context")
	}

	// Get target user
	targetUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("service get user: %w", err)
	}

	// Permission check - only need to verify basic management permission
	// Deactivation doesn't involve field updates, so we don't need field permission check
	if err := s.canManageUser(currentUser, targetUser); err != nil {
		return err
	}

	// Cannot deactivate self
	if targetUser.ID == currentUser.ID {
		return NewValidationError("permissions", "cannot deactivate yourself")
	}

	// Deactivate in database
	if err := s.repo.Deactivate(ctx, id, &currentUser.ID); err != nil {
		return fmt.Errorf("service deactivate user: %w", err)
	}

	// Also disable in Firebase/GCIP
	if err := auth.DisableFirebaseUser(ctx, targetUser.ExternalID); err != nil {
		log.Printf("WARNING: Failed to disable Firebase user %s: %v", targetUser.ExternalID, err)
		// Continue anyway - DB is source of truth
	}

	return nil
}

func (s *service) ReactivateUser(ctx context.Context, actor *auth.AuthUser, id uuid.UUID) error {
	// Get current user
	currentUser := actor
	if currentUser == nil {
		return fmt.Errorf("unauthorized: no auth user in context")
	}

	// Get target user
	targetUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("service get user: %w", err)
	}

	// Permission check - only need to verify basic management permission
	if err := s.canManageUser(currentUser, targetUser); err != nil {
		return err
	}

	// Cannot reactivate self (shouldn't happen but just in case)
	if targetUser.ID == currentUser.ID {
		return NewValidationError("permissions", "cannot reactivate yourself")
	}

	// Reactivate in database
	if err := s.repo.Reactivate(ctx, id); err != nil {
		return fmt.Errorf("service reactivate user: %w", err)
	}

	// Also enable in Firebase/GCIP
	if err := auth.EnableFirebaseUser(ctx, targetUser.ExternalID); err != nil {
		log.Printf("WARNING: Failed to enable Firebase user %s: %v", targetUser.ExternalID, err)
		// Continue anyway - DB is source of truth
	}

	return nil
}

// ResendPasswordSetupLink generates and sends a new password setup link
// Used when user didn't receive initial email or link expired
func (s *service) ResendPasswordSetupLink(ctx context.Context, actor *auth.AuthUser, userID uuid.UUID) error {
	// Get current user (for permission check)
	currentUser := actor
	if currentUser == nil {
		return fmt.Errorf("unauthorized: no auth user in context")
	}

	// Get target user
	targetUser, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("service get user: %w", err)
	}

	// Permission check - only need to verify basic management permission
	if err := s.canManageUser(currentUser, targetUser); err != nil {
		return err
	}

	// Generate new password reset link
	passwordSetupLink, err := auth.ResendPasswordResetLink(ctx, targetUser.Email)
	if err != nil {
		return fmt.Errorf("failed to generate password setup link: %w", err)
	}

	// Send reminder email (async, non-blocking)
	if s.emailSender != nil && passwordSetupLink != "" {
		go func(email, firstName, link string) {
			emailCtx := context.Background()
			if err := s.emailSender.SendPasswordSetupReminder(emailCtx, email, firstName, link); err != nil {
				log.Printf("Failed to send password setup reminder to %s: %v", email, err)
			}
		}(targetUser.Email, targetUser.FirstName, passwordSetupLink)
	} else {
		log.Printf("No email sender configured, skipping password setup reminder for %s", targetUser.Email)
	}

	log.Printf("Password setup link resent for user: %s (ID: %s)", targetUser.Email, userID)
	return nil
}

// -------------------- Permission Helpers -------------------- //
// canViewUser decides whether the actor is allowed to SEE the target user.
// It controls visibility rules (e.g. hide superadmins from admins).
func (s *service) canViewUser(currentUser *auth.AuthUser, targetUser *User) bool {
	// SuperAdmin can see everyone
	if currentUser.IsSuperAdmin() {
		return true
	}

	// Admin visibility rules
	if currentUser.IsAdminOnly() {
		// Admin CANNOT see SuperAdmins
		if targetUser.Role.Code == auth.RoleSuperAdmin {
			return false
		}
		// Admin CAN see:
		//  - themselves
		//  - other admins (read-only handled elsewhere)
		//  - adjusters / bodymen in the same shop
		if targetUser.Role.Code == auth.RoleAdmin {
			return true
		}
		// For adjuster/bodyman, must be in the same shop
		return targetUser.ShopID != nil && currentUser.ShopID != nil &&
			*targetUser.ShopID == *currentUser.ShopID
	}

	// For now, other roles should not normally hit this service-level method.
	// If they do, default to false to be safe.
	return false
}

// canManageUser checks basic management permissions (whether the current user can manage the target user)
// This function only cares about "who can manage whom", not what specific operation is being performed
func (s *service) canManageUser(currentUser *auth.AuthUser, targetUser *User) error {
	// SuperAdmin: No restrictions
	if currentUser.IsSuperAdmin() {
		return nil
	}

	// Admin restrictions
	if currentUser.RoleCode == auth.RoleAdmin {

		// Cannot manage superadmins
		if targetUser.Role.Code == auth.RoleSuperAdmin {
			return NewValidationError("permissions", "cannot manage superadmin users")
		}

		// Cannot manage other admins (except self)
		if targetUser.Role.Code == auth.RoleAdmin && targetUser.ID != currentUser.ID {
			return NewValidationError("permissions", "cannot manage other admin users")
		}

		// Admin can manage staff only within own shop
		if targetUser.Role.Code != auth.RoleAdmin { // staff or lower
			if currentUser.ShopID == nil || targetUser.ShopID == nil {
				return NewValidationError("permissions", "shop assignment missing")
			}
			if *currentUser.ShopID != *targetUser.ShopID {
				return NewValidationError("permissions", "cannot manage users from another shop")
			}
		}
	}

	return nil
}

// checkFieldUpdatePermission checks field update permissions (role-related restrictions)
// This function only cares about "what fields are being modified", ensuring no rules are violated
func (s *service) checkFieldUpdatePermission(currentUser *auth.AuthUser, targetUser *User, updates *UpdateUserInput) error {
	// SuperAdmin: No restrictions
	if currentUser.IsSuperAdmin() {
		return nil
	}

	// Admin restrictions
	if currentUser.IsAdminOnly() && updates.ShopCode != nil {
		// Cannot change shop assignment
		newCode := strings.TrimSpace(*updates.ShopCode)
		oldCode := ""
		if targetUser.Shop != nil {
			oldCode = strings.TrimSpace(targetUser.Shop.Code)
		}
		if !strings.EqualFold(newCode, oldCode) {
			return NewValidationError("shopCode", "cannot change shop assignment")
		}
		// Cannot promote anyone to admin
		if updates.RoleCode != nil {
			newRole := strings.TrimSpace(*updates.RoleCode)
			oldRole := strings.TrimSpace(targetUser.Role.Code)

			// If role is unchanged, skip further checks
			if strings.EqualFold(newRole, oldRole) {
			} else {
				// Cannot promote anyone to admin
				if newRole == auth.RoleAdmin {
					return NewValidationError("roleCode", "cannot promote users to admin role")
				}

				// Cannot promote anyone to superadmin
				if newRole == auth.RoleSuperAdmin {
					return NewValidationError("roleCode", "cannot promote users to superadmin role")
				}

				// Cannot change own role
				if targetUser.ID == currentUser.ID {
					return NewValidationError("roleCode", "cannot change your own role")
				}
			}
		}
	}

	return nil
}

// -------------------- Validation Helpers -------------------- //
// normalizeUser trims and formats fields before validation
func normalizeUser(in *CreateUserInput) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	if in.Phone != nil {
		trim := strings.TrimSpace(*in.Phone)
		if trim == "" {
			in.Phone = nil
		} else {
			in.Phone = &trim
		}
	}
	if in.ImageURL != nil {
		trim := strings.TrimSpace(*in.ImageURL)
		if trim == "" {
			in.ImageURL = nil
		} else {
			in.ImageURL = &trim
		}
	}
	if in.RoleCode != "" {
		in.RoleCode = strings.TrimSpace(in.RoleCode)
	}
	if in.ShopCode != nil {
		trim := strings.TrimSpace(*in.ShopCode)
		if trim == "" {
			in.ShopCode = nil
		} else {
			in.ShopCode = &trim
		}
	}
}

// normalizeUpdateUser trims and formats fields before validation
func normalizeUpdateUser(in *UpdateUserInput) {
	if in.FirstName != nil {
		trim := strings.TrimSpace(*in.FirstName)
		in.FirstName = &trim
	}
	if in.LastName != nil {
		trim := strings.TrimSpace(*in.LastName)
		in.LastName = &trim
	}
	if in.Phone != nil {
		trim := strings.TrimSpace(*in.Phone)
		if trim == "" {
			in.Phone = nil
		} else {
			in.Phone = &trim
		}
	}
	if in.ImageURL != nil {
		trim := strings.TrimSpace(*in.ImageURL)
		if trim == "" {
			in.ImageURL = nil
		} else {
			in.ImageURL = &trim
		}
	}
	if in.RoleCode != nil {
		trim := strings.TrimSpace(*in.RoleCode)
		if trim == "" {
			in.RoleCode = nil
		} else {
			in.RoleCode = &trim
		}
	}
	if in.ShopCode != nil {
		trim := strings.TrimSpace(*in.ShopCode)
		if trim == "" {
			in.ShopCode = nil
		} else {
			in.ShopCode = &trim
		}
	}
}

// validateUser performs schema-like validation
func validateUser(in *CreateUserInput) error {

	if in.Email == "" || !emailRegex.MatchString(in.Email) {
		return NewValidationError("email", "invalid or missing email")
	}
	if in.FirstName == "" {
		return NewValidationError("firstName", "cannot be blank")
	}
	if in.LastName == "" {
		return NewValidationError("lastName", "cannot be blank")
	}
	if in.RoleID == uuid.Nil {
		return NewValidationError("roleCode", "must be provided") // since we use roleCode from frontend
	}
	if in.Phone != nil && !phoneRegex.MatchString(*in.Phone) {
		return NewValidationError("phone", "invalid format (expected NNN-NNN-NNNN)")
	}
	if in.ImageURL != nil && !imageRegex.MatchString(*in.ImageURL) {
		return NewValidationError("imageUrl", "must start with http:// or https://")
	}
	return nil
}

// validateUpdateUser performs schema-like validation for updates
func validateUpdateUser(in *UpdateUserInput) error {

	if in.FirstName != nil && *in.FirstName == "" {
		return NewValidationError("firstName", "cannot be blank")
	}

	if in.LastName != nil && *in.LastName == "" {
		return NewValidationError("lastName", "cannot be blank")
	}

	if in.RoleID != nil && *in.RoleID == uuid.Nil {
		return NewValidationError("roleCode", "must be provided") // since we use roleCode from frontend
	}

	if in.Phone != nil && !phoneRegex.MatchString(*in.Phone) {
		return NewValidationError("phone", "invalid format (expected NNN-NNN-NNNN)")
	}

	if in.ImageURL != nil && !imageRegex.MatchString(*in.ImageURL) {
		return NewValidationError("imageUrl", "must start with http:// or https://")
	}

	return nil
}
