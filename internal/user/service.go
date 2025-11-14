package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// UserService defines business logic for users
type UserService interface {
	CreateUser(ctx context.Context, in *CreateUserInput) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByExternalID(ctx context.Context, externalID string) (*User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, in *UpdateUserInput) (*User, error)
	DeactivateUser(ctx context.Context, id uuid.UUID, byUserID *uuid.UUID) error
	ReactivateUser(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

// NewService creates a new UserService
func NewService(repo Repository) UserService {
	return &service{repo: repo}
}

// Regular expressions for validation
var (
	emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^[0-9]{3}-[0-9]{3}-[0-9]{4}$`)
	imageRegex = regexp.MustCompile(`^https?://`)
)

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
	if in.ExternalID == "" {
		return NewValidationError("externalId", "cannot be blank")
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

// CreateUser creates a new user
func (s *service) CreateUser(ctx context.Context, in *CreateUserInput) (*User, error) {
	if in == nil {
		return nil, ErrInvalidInput
	}
	normalizeUser(in)

	// Convert roleCode to roleID
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

	// Convert shopCode to shopID if provided
	if in.ShopCode != nil && *in.ShopCode != "" {
		shopID, err := s.repo.GetShopIDByCode(ctx, *in.ShopCode)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, NewValidationError("shopCode", "invalid shop code")
			}
			return nil, fmt.Errorf("lookup shopCode: %w", err)
		}
		in.ShopID = &shopID
	}

	if err := validateUser(in); err != nil {
		return nil, err
	}
	user, err := s.repo.Create(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("service create user: %w", err)
	}
	return user, nil
}

func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
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

func (s *service) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
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
	return list, nil
}

func (s *service) UpdateUser(ctx context.Context, id uuid.UUID, in *UpdateUserInput) (*User, error) {
	if in == nil {
		return nil, ErrInvalidInput
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
		shopID, err := s.repo.GetShopIDByCode(ctx, *in.ShopCode)
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

func (s *service) DeactivateUser(ctx context.Context, id uuid.UUID, byUserID *uuid.UUID) error {
	if err := s.repo.Deactivate(ctx, id, byUserID); err != nil {
		return fmt.Errorf("service deactivate user: %w", err)
	}
	return nil
}

func (s *service) ReactivateUser(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Reactivate(ctx, id); err != nil {
		return fmt.Errorf("service reactivate user: %w", err)
	}
	return nil
}
