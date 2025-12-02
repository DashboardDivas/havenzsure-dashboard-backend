package user

import "context"

// MeService for handling current user related operations
type MeService interface {
	GetCurrentUser(ctx context.Context, externalID string) (*User, error)
	UpdateProfile(ctx context.Context, externalID string, in *UpdateMeProfileInput) (*User, error)
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

func (s *meService) UpdateProfile(ctx context.Context, externalID string, in *UpdateMeProfileInput) (*User, error) {
	// 1) use externalID to get User
	u, err := s.userSvc.GetUserByExternalID(ctx, externalID)
	if err != nil {
		return nil, err
	}

	// 2) Compose UpdateUserInput with only phone / imageUrl
	upd := &UpdateUserInput{
		Phone:    in.Phone,
		ImageURL: in.ImageURL,
	}

	// 3) Delegate to UserService.UpdateUser for validation + repo.Update
	return s.userSvc.UpdateUser(ctx, u.ID, upd)
}
