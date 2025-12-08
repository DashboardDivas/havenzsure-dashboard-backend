package user

import (
	"time"

	"github.com/google/uuid"
)

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

// To do: Change admin-facing user dto here as well
