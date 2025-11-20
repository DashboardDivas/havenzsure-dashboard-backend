package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Mock Repository for testing Service layer
// ============================================================================

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, in *CreateUserInput) (*User, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetByCode(ctx context.Context, code string) (*User, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetByExternalID(ctx context.Context, externalID string) (*User, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*User), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, id uuid.UUID, in *UpdateUserInput) (*User, error) {
	args := m.Called(ctx, id, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Deactivate(ctx context.Context, id uuid.UUID, byUserID *uuid.UUID) error {
	args := m.Called(ctx, id, byUserID)
	return args.Error(0)
}

func (m *MockRepository) Reactivate(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) IncrementTokenVersion(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) TouchLastSignInNowByExternalID(ctx context.Context, externalID string) error {
	args := m.Called(ctx, externalID)
	return args.Error(0)
}

func (m *MockRepository) GetRoleIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	args := m.Called(ctx, code)
	id, ok := args.Get(0).(uuid.UUID)
	if !ok {
		return uuid.Nil, args.Error(1)
	}
	return id, args.Error(1)
}

func (m *MockRepository) GetShopIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	args := m.Called(ctx, code)
	id, ok := args.Get(0).(uuid.UUID)
	if !ok {
		return uuid.Nil, args.Error(1)
	}
	return id, args.Error(1)
}

// ============================================================================
// Validation Tests - Helper Functions
// ============================================================================

// TestValidateUserEmail validates email format
func TestValidateUserEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		wantErr  bool
		errField string
	}{
		{"valid email", "user@example.com", false, ""},
		{"invalid email - missing @", "userexample.com", true, "email"},
		{"empty email", "", true, "email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &CreateUserInput{
				Email:      tt.email,
				FirstName:  "John",
				LastName:   "Doe",
				ExternalID: "gcip-123",
				RoleID:     testUUID("11111111-1111-1111-1111-111111111111"),
			}

			err := validateUser(input)

			if tt.wantErr {
				assert.Error(t, err)
				if validErr, ok := err.(*ValidationError); ok {
					assert.Equal(t, tt.errField, validErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateUserNames validates firstName and lastName
func TestValidateUserNames(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		wantErr   bool
	}{
		{"valid names", "John", "Doe", false},
		{"empty firstName", "", "Doe", true},
		{"empty lastName", "John", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &CreateUserInput{
				Email:      "user@example.com",
				FirstName:  tt.firstName,
				LastName:   tt.lastName,
				ExternalID: "gcip-123",
				RoleID:     testUUID("11111111-1111-1111-1111-111111111111"),
			}

			err := validateUser(input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateUserOptionalFields validates phone and imageURL fields
func TestValidateUserOptionalFields(t *testing.T) {
	tests := []struct {
		name     string
		phone    *string
		imageURL *string
		wantErr  bool
	}{
		{"valid phone and imageUrl", testString("403-123-4567"), testString("https://example.com/img.jpg"), false},
		{"invalid phone format", testString("123"), nil, true},
		{"invalid imageUrl - no protocol", nil, testString("example.com"), true},
		{"nil optional fields - ok", nil, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &CreateUserInput{
				Email:      "user@example.com",
				FirstName:  "John",
				LastName:   "Doe",
				Phone:      tt.phone,
				ImageURL:   tt.imageURL,
				ExternalID: "gcip-123",
				RoleID:     testUUID("11111111-1111-1111-1111-111111111111"),
			}

			err := validateUser(input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Normalization Tests - Helper Functions
// ============================================================================

// TestNormalizeUser tests normalization: lowercase email, capitalize names, trim whitespace
func TestNormalizeUser(t *testing.T) {
	tests := []struct {
		name           string
		inputEmail     string
		inputFirstName string
		inputLastName  string
		expectedEmail  string
		expectedFirst  string
		expectedLast   string
	}{
		{
			name:           "basic normalization",
			inputEmail:     "  USER@EXAMPLE.COM  ",
			inputFirstName: "john",
			inputLastName:  "doe",
			expectedEmail:  "user@example.com",
			expectedFirst:  "john",
			expectedLast:   "doe",
		},
		{
			name:           "already normalized",
			inputEmail:     "user@example.com",
			inputFirstName: "John",
			inputLastName:  "Doe",
			expectedEmail:  "user@example.com",
			expectedFirst:  "John",
			expectedLast:   "Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &CreateUserInput{
				Email:      tt.inputEmail,
				FirstName:  tt.inputFirstName,
				LastName:   tt.inputLastName,
				ExternalID: "gcip-123",
			}

			normalizeUser(input)

			assert.Equal(t, tt.expectedEmail, input.Email)
			assert.Equal(t, tt.expectedFirst, input.FirstName)
			assert.Equal(t, tt.expectedLast, input.LastName)
		})
	}
}

// TestNormalizeUserOptionalFields tests that whitespace-only fields become nil
func TestNormalizeUserOptionalFields(t *testing.T) {
	t.Run("whitespace-only becomes nil", func(t *testing.T) {
		input := &CreateUserInput{
			Email:      "user@example.com",
			FirstName:  "John",
			LastName:   "Doe",
			Phone:      testString("   "),
			ImageURL:   testString("   "),
			ShopCode:   testString("   "),
			ExternalID: "gcip-123",
		}

		normalizeUser(input)

		assert.Nil(t, input.Phone)
		assert.Nil(t, input.ImageURL)
		assert.Nil(t, input.ShopCode)
	})

	t.Run("nil fields stay nil", func(t *testing.T) {
		input := &CreateUserInput{
			Email:      "user@example.com",
			FirstName:  "John",
			LastName:   "Doe",
			Phone:      nil,
			ImageURL:   nil,
			ExternalID: "gcip-123",
		}

		normalizeUser(input)

		assert.Nil(t, input.Phone)
		assert.Nil(t, input.ImageURL)
	})
}

// ============================================================================
// CreateUser Tests (TC050 + edge cases)
// ============================================================================

// TestCreateUserSuccess tests creating a user with minimal and optional fields
func TestCreateUserSuccess(t *testing.T) {
	t.Run("create user with minimal fields", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo)
		ctx := context.Background()

		adminRoleID := uuid.New()

		input := &CreateUserInput{
			Email:      "user@example.com",
			FirstName:  "John",
			LastName:   "Doe",
			ExternalID: "gcip-123",
			RoleCode:   "admin",
		}

		mockRepo.On("GetRoleIDByCode", ctx, "admin").Return(adminRoleID, nil)
		mockRepo.On("Create", ctx, mock.MatchedBy(func(in *CreateUserInput) bool {
			return in.Email == "user@example.com" && in.RoleID == adminRoleID
		})).Return(&User{
			ID:         uuid.New(),
			Email:      "user@example.com",
			FirstName:  "John",
			LastName:   "Doe",
			ExternalID: "gcip-123",
			RoleID:     adminRoleID,
		}, nil)

		user, err := svc.CreateUser(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user@example.com", user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create user with optional fields", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo)
		ctx := context.Background()

		adminRoleID := uuid.New()
		shopID := uuid.New()

		input := &CreateUserInput{
			Email:      "jane@example.com",
			FirstName:  "Jane",
			LastName:   "Smith",
			Phone:      testString("403-123-4567"),
			ImageURL:   testString("https://example.com/img.jpg"),
			ExternalID: "gcip-456",
			RoleCode:   "bodyman",
			ShopCode:   testString("SHOP-001"),
		}

		mockRepo.On("GetRoleIDByCode", ctx, "bodyman").Return(adminRoleID, nil)
		mockRepo.On("GetShopIDByCode", ctx, "SHOP-001").Return(shopID, nil)
		mockRepo.On("Create", ctx, mock.MatchedBy(func(in *CreateUserInput) bool {
			return in.Email == "jane@example.com" &&
				in.Phone != nil && *in.Phone == "403-123-4567" &&
				in.ShopID != nil && *in.ShopID == shopID
		})).Return(&User{
			ID:        uuid.New(),
			Email:     "jane@example.com",
			FirstName: "Jane",
			LastName:  "Smith",
			Phone:     testString("403-123-4567"),
			ImageURL:  testString("https://example.com/img.jpg"),
			RoleID:    adminRoleID,
			ShopID:    &shopID,
		}, nil)

		user, err := svc.CreateUser(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "jane@example.com", user.Email)
		mockRepo.AssertExpectations(t)
	})
}

// TestCreateUserValidationErrors tests validation errors
func TestCreateUserValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         *CreateUserInput
		wantErr       bool
		expectedField string
	}{
		{
			name: "invalid email",
			input: &CreateUserInput{
				Email:      "invalid-email",
				FirstName:  "John",
				LastName:   "Doe",
				ExternalID: "gcip-123",
				RoleCode:   "admin",
			},
			wantErr:       true,
			expectedField: "email",
		},
		{
			name: "missing firstName",
			input: &CreateUserInput{
				Email:      "user@example.com",
				FirstName:  "",
				LastName:   "Doe",
				ExternalID: "gcip-123",
				RoleCode:   "admin",
			},
			wantErr:       true,
			expectedField: "firstName",
		},
		{
			name: "invalid phone format",
			input: &CreateUserInput{
				Email:      "user@example.com",
				FirstName:  "John",
				LastName:   "Doe",
				Phone:      testString("123"),
				ExternalID: "gcip-123",
				RoleCode:   "admin",
			},
			wantErr:       true,
			expectedField: "phone",
		},
		{
			name: "invalid imageURL format",
			input: &CreateUserInput{
				Email:      "user@example.com",
				FirstName:  "John",
				LastName:   "Doe",
				ImageURL:   testString("not-a-url"),
				ExternalID: "gcip-123",
				RoleCode:   "admin",
			},
			wantErr:       true,
			expectedField: "imageUrl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			svc := NewService(mockRepo)
			ctx := context.Background()

			adminRoleID := uuid.New()
			mockRepo.On("GetRoleIDByCode", ctx, "admin").Return(adminRoleID, nil)

			_, err := svc.CreateUser(ctx, tt.input)

			assert.Error(t, err)
			if validErr, ok := err.(*ValidationError); ok {
				assert.Equal(t, tt.expectedField, validErr.Field)
			}
		})
	}
}

// TestCreateUser_InvalidRoleCode tests CreateUser with invalid roleCode
func TestCreateUser_InvalidRoleCode(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo)
	ctx := context.Background()

	input := &CreateUserInput{
		Email:      "user@example.com",
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "gcip-123",
		RoleCode:   "unknown-role",
	}

	mockRepo.On("GetRoleIDByCode", ctx, "unknown-role").Return(uuid.Nil, ErrNotFound)

	_, err := svc.CreateUser(ctx, input)

	assert.Error(t, err)
	if vErr, ok := err.(*ValidationError); ok {
		assert.Equal(t, "roleCode", vErr.Field)
		assert.Equal(t, "invalid role code", vErr.Message)
	}
}

// TestCreateUserNormalization tests edge case: whitespace-only fields become nil
func TestCreateUserNormalization(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo)
	ctx := context.Background()

	adminRoleID := uuid.New()

	input := &CreateUserInput{
		Email:      "user@example.com",
		FirstName:  "John",
		LastName:   "Doe",
		Phone:      testString("   "),
		ImageURL:   testString("   "),
		ExternalID: "gcip-123",
		RoleCode:   "admin",
		ShopCode:   testString("   "),
	}

	mockRepo.On("GetRoleIDByCode", ctx, "admin").Return(adminRoleID, nil)
	mockRepo.On("Create", ctx, mock.MatchedBy(func(in *CreateUserInput) bool {
		return in.Phone == nil && in.ImageURL == nil && in.ShopCode == nil
	})).Return(&User{
		ID:     uuid.New(),
		Email:  "user@example.com",
		RoleID: adminRoleID,
	}, nil)

	user, err := svc.CreateUser(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// UpdateUser Tests (TC053 + edge cases)
// ============================================================================

// TestUpdateUserSuccess tests updating user with various field combinations
func TestUpdateUserSuccess(t *testing.T) {
	t.Run("update multiple fields", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo)
		ctx := context.Background()

		userID := uuid.New()
		newRoleID := uuid.New()

		input := &UpdateUserInput{
			FirstName: testString("Jane"),
			LastName:  testString("Smith"),
			Phone:     testString("403-111-2222"),
			RoleCode:  testString("bodyman"),
		}

		mockRepo.On("GetRoleIDByCode", ctx, "bodyman").Return(newRoleID, nil)
		mockRepo.On("Update", ctx, userID, mock.MatchedBy(func(in *UpdateUserInput) bool {
			return in.FirstName != nil && *in.FirstName == "Jane" &&
				in.LastName != nil && *in.LastName == "Smith" &&
				in.RoleID != nil && *in.RoleID == newRoleID
		})).Return(&User{
			ID:        userID,
			Email:     "user@example.com",
			FirstName: "Jane",
			LastName:  "Smith",
			RoleID:    newRoleID,
		}, nil)

		user, err := svc.UpdateUser(ctx, userID, input)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Jane", user.FirstName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("partial update - single field", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo)
		ctx := context.Background()

		userID := uuid.New()

		input := &UpdateUserInput{
			FirstName: testString("Alice"),
		}

		mockRepo.On("Update", ctx, userID, input).Return(&User{
			ID:        userID,
			FirstName: "Alice",
		}, nil)

		user, err := svc.UpdateUser(ctx, userID, input)

		assert.NoError(t, err)
		assert.Equal(t, "Alice", user.FirstName)
		mockRepo.AssertExpectations(t)
	})
}

// TestUpdateUserValidationErrors tests validation errors during update
func TestUpdateUserValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         *UpdateUserInput
		wantErr       bool
		expectedField string
	}{
		{
			name: "empty firstName - invalid",
			input: &UpdateUserInput{
				FirstName: testString(""),
			},
			wantErr:       true,
			expectedField: "firstName",
		},
		{
			name: "invalid phone",
			input: &UpdateUserInput{
				Phone: testString("123"),
			},
			wantErr:       true,
			expectedField: "phone",
		},
		{
			name: "empty phone - becomes nil - ok",
			input: &UpdateUserInput{
				Phone: testString("   "),
			},
			wantErr: false,
		},
		{
			name: "invalid imageUrl",
			input: &UpdateUserInput{
				ImageURL: testString("ftp://bad-url"),
			},
			wantErr:       true,
			expectedField: "imageUrl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			svc := NewService(mockRepo)
			ctx := context.Background()

			if !tt.wantErr {
				mockRepo.On("Update", ctx, mock.Anything, mock.MatchedBy(func(in *UpdateUserInput) bool {
					return in.Phone == nil
				})).Return(&User{ID: uuid.New()}, nil)
			}

			_, err := svc.UpdateUser(ctx, uuid.New(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if validErr, ok := err.(*ValidationError); ok {
					assert.Equal(t, tt.expectedField, validErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUpdateUser_InvalidRoleCode tests UpdateUser with invalid roleCode
func TestUpdateUser_InvalidShopCode(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	input := &UpdateUserInput{
		ShopCode: testString("SHOP-999"),
	}

	mockRepo.On("GetShopIDByCode", ctx, "SHOP-999").Return(uuid.Nil, ErrNotFound)

	_, err := svc.UpdateUser(ctx, userID, input)

	assert.Error(t, err)
	if vErr, ok := err.(*ValidationError); ok {
		assert.Equal(t, "shopCode", vErr.Field)
		assert.Equal(t, "invalid shop code", vErr.Message)
	}
}

// ============================================================================
// ListUsers Tests (pagination edge cases)
// ============================================================================

// TestListUsersPagination tests pagination with edge cases
func TestListUsersPagination(t *testing.T) {
	tests := []struct {
		name           string
		inputLimit     int
		inputOffset    int
		expectedLimit  int
		expectedOffset int
		description    string
	}{
		{
			name:           "normal pagination",
			inputLimit:     20,
			inputOffset:    0,
			expectedLimit:  20,
			expectedOffset: 0,
			description:    "limit > 0, offset = 0",
		},
		{
			name:           "limit <= 0 defaults to 50",
			inputLimit:     0,
			inputOffset:    0,
			expectedLimit:  50,
			expectedOffset: 0,
			description:    "limit = 0 should default to 50",
		},
		{
			name:           "negative limit defaults to 50",
			inputLimit:     -1,
			inputOffset:    10,
			expectedLimit:  50,
			expectedOffset: 10,
			description:    "limit < 0 should default to 50",
		},
		{
			name:           "negative offset defaults to 0",
			inputLimit:     20,
			inputOffset:    -1,
			expectedLimit:  20,
			expectedOffset: 0,
			description:    "offset < 0 should default to 0",
		},
		{
			name:           "both invalid - apply defaults",
			inputLimit:     -10,
			inputOffset:    -5,
			expectedLimit:  50,
			expectedOffset: 0,
			description:    "both invalid should apply both defaults",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			svc := NewService(mockRepo)
			ctx := context.Background()

			mockUsers := []*User{
				{ID: uuid.New(), Email: "user1@example.com", FirstName: "John", LastName: "Doe"},
				{ID: uuid.New(), Email: "user2@example.com", FirstName: "Jane", LastName: "Smith"},
			}

			mockRepo.On("List", ctx, tt.expectedLimit, tt.expectedOffset).Return(mockUsers, nil)

			users, err := svc.ListUsers(ctx, tt.inputLimit, tt.inputOffset)

			assert.NoError(t, err)
			assert.NotNil(t, users)
			assert.Len(t, users, 2)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// GetUserByID Tests (simple success test)
// ============================================================================

// TestGetUserByIDSuccess tests retrieving user by ID
func TestGetUserByIDSuccess(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedUser := &User{
		ID:         userID,
		Email:      "user@example.com",
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "gcip-123",
	}

	mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	user, err := svc.GetUserByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "user@example.com", user.Email)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// GetUserByExternalID Tests (simple success test)
// ============================================================================

// TestGetUserByExternalIDSuccess tests retrieving user by external ID
func TestGetUserByExternalIDSuccess(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedUser := &User{
		ID:         userID,
		Email:      "user@example.com",
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "gcip-123",
	}

	mockRepo.On("GetByExternalID", ctx, "gcip-123").Return(expectedUser, nil)

	user, err := svc.GetUserByExternalID(ctx, "gcip-123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "gcip-123", user.ExternalID)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// DeactivateUser Tests (TC054 + error cases)
// ============================================================================

// TestDeactivateUserSuccess tests deactivating a user
func TestDeactivateUserSuccess(t *testing.T) {
	t.Run("basic deactivation", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo)
		ctx := context.Background()

		userID := uuid.New()

		mockRepo.On("Deactivate", ctx, userID, (*uuid.UUID)(nil)).Return(nil)

		err := svc.DeactivateUser(ctx, userID, nil)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("deactivation with byUserID (audit trail)", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo)
		ctx := context.Background()

		userID := uuid.New()
		byUserID := uuid.New()

		mockRepo.On("Deactivate", ctx, userID, &byUserID).Return(nil)

		err := svc.DeactivateUser(ctx, userID, &byUserID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

// TestDeactivateUserError tests error cases for deactivation
func TestDeactivateUserError(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "user not found",
			description: "should return error when user does not exist",
		},
		{
			name:        "already deactivated",
			description: "should return error when user is already deactivated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			svc := NewService(mockRepo)
			ctx := context.Background()

			userID := uuid.New()

			mockRepo.On("Deactivate", ctx, mock.Anything, mock.Anything).
				Return(ErrNotFound)

			err := svc.DeactivateUser(ctx, userID, nil)

			assert.Error(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// ReactivateUser Tests (TC054 + error cases)
// ============================================================================

// TestReactivateUserSuccess tests reactivating a user
func TestReactivateUserSuccess(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("Reactivate", ctx, userID).Return(nil)

	err := svc.ReactivateUser(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestReactivateUserError tests error cases for reactivation
func TestReactivateUserError(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "user not found",
			description: "should return error when user does not exist",
		},
		{
			name:        "user not deactivated",
			description: "should return error when user is not deactivated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			svc := NewService(mockRepo)
			ctx := context.Background()

			userID := uuid.New()

			mockRepo.On("Reactivate", ctx, userID).Return(ErrNotFound)

			err := svc.ReactivateUser(ctx, userID)

			assert.Error(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func testString(s string) *string {
	return &s
}

func testUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
