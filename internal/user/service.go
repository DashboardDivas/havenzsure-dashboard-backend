package user

import (
	"context"
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
}

// validateUser performs schema-like validation
func validateUser(in *CreateUserInput) error {
	emailRegex := regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	phoneRegex := regexp.MustCompile(`^[0-9]{3}-[0-9]{3}-[0-9]{4}$`)
	imageRegex := regexp.MustCompile(`^https?://`)

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
		return NewValidationError("roleId", "must be provided")
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
