package user

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a role in the system
type Role struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	IsSystem  bool      `json:"-"` // backend-only: protect system roles
	CreatedAt time.Time `json:"-"` // backend-only
	UpdatedAt time.Time `json:"-"` // backend-only
}

// User represents a system user
type User struct {
	ID            uuid.UUID  `json:"id"`
	Code          string     `json:"code"`
	Email         string     `json:"email"`
	FirstName     string     `json:"firstName"`
	LastName      string     `json:"lastName"`
	Phone         *string    `json:"phone,omitempty"`
	ImageURL      *string    `json:"imageUrl,omitempty"`
	ExternalID    string     `json:"-"` // GCIP UID (immutable, backend-only)
	EmailVerified bool       `json:"emailVerified"`
	IsActive      bool       `json:"isActive"`
	DeactivatedAt *time.Time `json:"deactivatedAt,omitempty"`
	DeactivatedBy *uuid.UUID `json:"deactivatedBy,omitempty"`
	TokenVersion  int        `json:"-"` // backend-only (force logout mechanism)
	ShopID        *uuid.UUID `json:"shopId,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	LastSignInAt  *time.Time `json:"lastSignInAt,omitempty"`

	// Role - each user must have exactly ONE role (enforced by role_id NOT NULL)
	Role Role `json:"role"`
}

// CreateUserInput represents the input data required to create a new user
type CreateUserInput struct {
	Email         string     `json:"email"`
	FirstName     string     `json:"firstName"`
	LastName      string     `json:"lastName"`
	Phone         *string    `json:"phone,omitempty"`
	ImageURL      *string    `json:"imageUrl,omitempty"`
	ExternalID    string     `json:"externalId"` // required from GCIP
	EmailVerified bool       `json:"emailVerified"`
	ShopID        *uuid.UUID `json:"shopId,omitempty"`
	RoleID        uuid.UUID  `json:"roleId"`
}

// UpdateUserInput represents the fields that can be updated for a user
type UpdateUserInput struct {
	FirstName     *string    `json:"firstName,omitempty"`
	LastName      *string    `json:"lastName,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	ImageURL      *string    `json:"imageUrl,omitempty"`
	EmailVerified *bool      `json:"emailVerified,omitempty"`
	ShopID        *uuid.UUID `json:"shopId,omitempty"`
	RoleID        *uuid.UUID `json:"roleId,omitempty"`
	// IsActive cannot be updated here; use ActivateUser / DeactivateUser methods
}
