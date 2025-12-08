package user

import (
	"context"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/platform/auth"
)

// MeService for handling current user related operations
type MeService interface {
	GetCurrentUser(ctx context.Context, externalID string) (*User, error)
	UpdateProfile(ctx context.Context, actor *auth.AuthUser, in *UpdateMeProfileInput) (*User, error)
}

type meService struct {
	userSvc UserService
}

func NewMeService(userSvc UserService) MeService {
	return &meService{userSvc: userSvc}
}

func (s *meService) GetCurrentUser(ctx context.Context, externalID string) (*User, error) {
	// Delegate to UserService.GetUserByExternalID
	return s.userSvc.GetUserByExternalID(ctx, externalID)
}

func (s *meService) UpdateProfile(ctx context.Context, actor *auth.AuthUser, in *UpdateMeProfileInput) (*User, error) {
	// 1) use externalID to get User
	u, err := s.userSvc.GetUserByExternalID(ctx, actor.ExternalID)
	if err != nil {
		return nil, err
	}

	// 2) Compose UpdateUserInput with only phone / imageUrl

	upd := &UpdateUserInput{
		Phone:    in.Phone,
		ImageURL: in.ImageURL,
	}

	// 3) Delegate to UserService.UpdateUser for validation + repo.Update.
	// UpdateProfile constructs an UpdateUserInput that only allows phone/imageUrl to be updated.
	// If additional fields are added to UpdateMeProfileInput in the future, they must still pass
	// UserService.UpdateUser RBAC checks, which enforce all role/shop permission rules.
	return s.userSvc.UpdateUser(ctx, actor, u.ID, upd)
}
