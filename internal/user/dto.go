package user

import (
	"time"

	"github.com/google/uuid"
)

// -------------------- /me DTOs ---------------------

// MeResponse is the DTO for /me response
type MeResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Phone     *string   `json:"phone,omitempty"`
	ImageURL  *string   `json:"imageUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`

	RoleCode string `json:"roleCode"`
	RoleName string `json:"roleName"`

	ShopID   *uuid.UUID `json:"shopId,omitempty"`
	ShopCode *string    `json:"shopCode,omitempty"`
	ShopName *string    `json:"shopName,omitempty"`
}

// ToMeResponse converts User model to MeResponse DTO
func (u *User) ToMeResponse() *MeResponse {
	if u == nil {
		return nil
	}

	var shopID *uuid.UUID
	var shopCode *string
	var shopName *string

	if u.ShopID != nil {
		shopID = u.ShopID
	}

	if u.Shop != nil {
		shopCode = &u.Shop.Code
		shopName = &u.Shop.Name
	}

	return &MeResponse{
		ID:        u.ID,
		Code:      u.Code,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		ImageURL:  u.ImageURL,
		CreatedAt: u.CreatedAt,

		RoleCode: u.Role.Code,
		RoleName: u.Role.Name,

		ShopID:   shopID,
		ShopCode: shopCode,
		ShopName: shopName,
	}
}

// UpdateMeProfileInput represents the fields that the current user
// is allowed to update on their own profile via /users/me.
type UpdateMeProfileInput struct {
	Phone    *string `json:"phone,omitempty"`
	ImageURL *string `json:"imageUrl,omitempty"`
}

// ------------- Admin-facing User DTOs -------------

// UserResponse represents the user data sent to the frontend
type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	Code          string    `json:"code"`
	Email         string    `json:"email"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	Phone         *string   `json:"phone,omitempty"`
	ImageURL      *string   `json:"imageUrl,omitempty"`
	EmailVerified bool      `json:"emailVerified"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Role          Role      `json:"role"`
	Shop          *Shop     `json:"shop,omitempty"`
}

// ToResponse converts User (internal) to UserResponse (for API)
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		Code:          u.Code,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Phone:         u.Phone,
		ImageURL:      u.ImageURL,
		EmailVerified: u.EmailVerified,
		IsActive:      u.IsActive,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		Role:          u.Role,
		Shop:          u.Shop,
	}
}
