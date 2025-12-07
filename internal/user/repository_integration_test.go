package user

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB   *pgxpool.Pool
	testRepo Repository
)

// TestMain sets up the test database connection and runs all tests
// Usage:
//  1. Start database: docker-compose -f compose.test.yml up -d and run migrations
//  2. Set environment variables in .env:
//  3. Run tests: go test -v ./internal/user
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Load environment variables from .env file
	if err := godotenv.Load("../../.env"); err != nil {
		fmt.Println("warning: could not load ../../.env:", err)
	}

	// Build connection string from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_TEST_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_APP_USER")
	dbPassword := os.Getenv("DB_APP_PASSWORD")
	dbSchema := os.Getenv("DB_SCHEMA")

	if dbHost == "" || dbPort == "" || dbName == "" || dbUser == "" || dbPassword == "" {
		fmt.Fprintf(os.Stderr, "Missing required environment variables\n")
		fmt.Fprintf(os.Stderr, "Required: DB_HOST, DB_TEST_PORT, DB_NAME, DB_APP_USER, DB_APP_PASSWORD\n")
		fmt.Fprintf(os.Stderr, "Optional: DB_SCHEMA (default: app)\n")
		os.Exit(1)
	}

	if dbSchema == "" {
		dbSchema = "app" // Default schema
	}

	// Construct connection string
	databaseURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSchema)

	// Connect to database
	var err error
	testDB, err = pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		fmt.Fprintf(os.Stderr, "Connection string: postgresql://%s:***@%s:%s/%s?search_path=%s\n",
			dbUser, dbHost, dbPort, dbName, dbSchema)
		os.Exit(1)
	}
	defer testDB.Close()

	// Ping database to verify connection
	if err := testDB.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully connected to test database")
	fmt.Printf("   Schema: %s\n", dbSchema)

	// Create repository
	testRepo = NewUserRepository(testDB)

	fmt.Println("Repository initialized")
	fmt.Println("Running tests...")

	// Run tests
	code := m.Run()

	fmt.Println("\nTests completed")
	os.Exit(code)
}

// cleanupTestData removes all test data between tests
func cleanupTestData(t *testing.T) {
	ctx := context.Background()

	// Delete all test users (avoid deleting seeded roles)
	_, err := testDB.Exec(ctx, `DELETE FROM app.users WHERE email LIKE '%@example.com'`)
	require.NoError(t, err)
}

// getRoleIDByCode returns the ID of a role by code
func getRoleIDByCode(t *testing.T, code string) uuid.UUID {
	ctx := context.Background()
	var roleID uuid.UUID
	err := testDB.QueryRow(ctx, "SELECT id FROM app.roles WHERE code = $1", code).Scan(&roleID)
	require.NoError(t, err, "Role %s should exist", code)
	return roleID
}

// Test: Create user with all fields
func TestCreateUserWithAllFields(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	// Get shop if exists
	var shopID *uuid.UUID
	var sid uuid.UUID
	err := testDB.QueryRow(ctx, "SELECT id FROM app.shop LIMIT 1").Scan(&sid)
	if err == nil {
		shopID = &sid
	}

	phone := "123-456-7890" // Format: 123-456-7890
	imageURL := "https://example.com/image.jpg"

	input := &CreateUserInput{
		Email:         "test@example.com",
		FirstName:     "John",
		LastName:      "Doe",
		Phone:         &phone,
		ImageURL:      &imageURL,
		ExternalID:    "firebase|test123",
		EmailVerified: true,
		ShopID:        shopID,
		RoleID:        roleID,
	}

	user, err := testRepo.Create(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEqual(t, uuid.Nil, user.ID)

	// Code should be auto-generated like "U-0001"
	assert.NotEmpty(t, user.Code)
	assert.Regexp(t, `^U-\d{4}$`, user.Code)

	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.NotNil(t, user.Phone)
	assert.Equal(t, phone, *user.Phone)
	assert.NotNil(t, user.ImageURL)
	assert.Equal(t, imageURL, *user.ImageURL)
	assert.Equal(t, "firebase|test123", user.ExternalID)
	assert.True(t, user.EmailVerified)
	assert.True(t, user.IsActive)
	assert.Equal(t, roleID, user.RoleID)
	assert.Equal(t, "admin", user.Role.Code)
}

// Test: Create user with minimum required fields
func TestCreateUserWithMinimalFields(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "bodyman")

	input := &CreateUserInput{
		Email:         "minimal@example.com",
		FirstName:     "Jane",
		LastName:      "Smith",
		ExternalID:    "firebase|minimal",
		EmailVerified: false,
		RoleID:        roleID,
	}

	user, err := testRepo.Create(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "minimal@example.com", user.Email)
	assert.Equal(t, "Jane", user.FirstName)
	assert.Equal(t, "Smith", user.LastName)
	assert.Nil(t, user.Phone)
	assert.Nil(t, user.ImageURL)
	assert.Nil(t, user.Shop)
	assert.False(t, user.EmailVerified)
	assert.True(t, user.IsActive)
	assert.Equal(t, "bodyman", user.Role.Code)
}

// Test: Attempt to create user with duplicate email
func TestCreateUserWithDuplicateEmail(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	// Create first user
	input1 := &CreateUserInput{
		Email:      "duplicate@example.com",
		FirstName:  "First",
		LastName:   "User",
		ExternalID: "firebase|first",
		RoleID:     roleID,
	}
	_, err := testRepo.Create(ctx, input1)
	require.NoError(t, err)

	// Attempt to create second user with same email
	input2 := &CreateUserInput{
		Email:      "duplicate@example.com",
		FirstName:  "Second",
		LastName:   "User",
		ExternalID: "firebase|second",
		RoleID:     roleID,
	}
	user, err := testRepo.Create(ctx, input2)

	assert.Error(t, err)
	assert.Equal(t, ErrConflict, err)
	assert.Nil(t, user)
}

// Test: Attempt to create user with duplicate external_id
func TestCreateUserWithDuplicateExternalID(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	// Create first user
	input1 := &CreateUserInput{
		Email:      "user1@example.com",
		FirstName:  "First",
		LastName:   "User",
		ExternalID: "firebase|same-id",
		RoleID:     roleID,
	}
	_, err := testRepo.Create(ctx, input1)
	require.NoError(t, err)

	// Attempt to create second user with same external_id
	input2 := &CreateUserInput{
		Email:      "user2@example.com",
		FirstName:  "Second",
		LastName:   "User",
		ExternalID: "firebase|same-id",
		RoleID:     roleID,
	}
	user, err := testRepo.Create(ctx, input2)

	assert.Error(t, err)
	assert.Equal(t, ErrConflict, err)
	assert.Nil(t, user)
}

// Test: Attempt to create user with invalid role ID
func TestCreateUserWithInvalidRoleID(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()

	input := &CreateUserInput{
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		ExternalID: "firebase|test",
		RoleID:     uuid.New(),
	}

	user, err := testRepo.Create(ctx, input)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
	assert.Nil(t, user)
}

// Test: Attempt to create user with missing email
func TestCreateUserWithMissingEmail(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	input := &CreateUserInput{
		Email:      "",
		FirstName:  "Test",
		LastName:   "User",
		ExternalID: "firebase|test",
		RoleID:     roleID,
	}

	user, err := testRepo.Create(ctx, input)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
	assert.Nil(t, user)
}

// Test: Attempt to create user with invalid phone format
func TestCreateUserWithInvalidPhoneFormat(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	invalidPhone := "1234567890" // Wrong format, should be 123-456-7890
	input := &CreateUserInput{
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Phone:      &invalidPhone,
		ExternalID: "firebase|test",
		RoleID:     roleID,
	}

	user, err := testRepo.Create(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, user)
}

// Test: Retrieve user by ID
func TestGetUserByID(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	input := &CreateUserInput{
		Email:      "getbyid@example.com",
		FirstName:  "Get",
		LastName:   "ByID",
		ExternalID: "firebase|getbyid",
		RoleID:     roleID,
	}
	created, err := testRepo.Create(ctx, input)
	require.NoError(t, err)

	user, err := testRepo.GetByID(ctx, created.ID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, created.ID, user.ID)
	assert.Equal(t, "getbyid@example.com", user.Email)
	assert.Equal(t, "admin", user.Role.Code)
}

// Test: Attempt to retrieve non-existent user by ID
func TestGetUserByIDNotFound(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()

	user, err := testRepo.GetByID(ctx, uuid.New())

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, user)
}

// Test: Retrieve user by unique code
func TestGetUserByCode(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	input := &CreateUserInput{
		Email:      "getbycode@example.com",
		FirstName:  "Get",
		LastName:   "ByCode",
		ExternalID: "firebase|getbycode",
		RoleID:     roleID,
	}
	created, err := testRepo.Create(ctx, input)
	require.NoError(t, err)

	user, err := testRepo.GetByCode(ctx, created.Code)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, created.ID, user.ID)
	assert.Equal(t, created.Code, user.Code)
}

// Test: Attempt to retrieve user with non-existent code
func TestGetUserByCodeNotFound(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()

	user, err := testRepo.GetByCode(ctx, "U-9999")

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, user)
}

// Test: Retrieve user by external ID
func TestGetUserByExternalID(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	input := &CreateUserInput{
		Email:      "getbyext@example.com",
		FirstName:  "Get",
		LastName:   "ByExternal",
		ExternalID: "firebase|getbyext",
		RoleID:     roleID,
	}
	created, err := testRepo.Create(ctx, input)
	require.NoError(t, err)

	user, err := testRepo.GetByExternalID(ctx, "firebase|getbyext")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, created.ID, user.ID)
	assert.Equal(t, "firebase|getbyext", user.ExternalID)
}

// Test: Attempt to retrieve user with non-existent external ID
func TestGetUserByExternalIDNotFound(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()

	user, err := testRepo.GetByExternalID(ctx, "firebase|nonexistent")

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, user)
}

// Test: List users with pagination (first page)
func TestListUsersFirstPage(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	// Create 15 test users
	for i := 1; i <= 15; i++ {
		input := &CreateUserInput{
			Email:      fmt.Sprintf("user%d@example.com", i),
			FirstName:  fmt.Sprintf("User%d", i),
			LastName:   "Test",
			ExternalID: fmt.Sprintf("firebase|user%d", i),
			RoleID:     roleID,
		}
		_, err := testRepo.Create(ctx, input)
		require.NoError(t, err)
	}

	users, err := testRepo.List(ctx, 10, 0)

	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Len(t, users, 10)
}

// Test: List users with pagination (second page)
func TestListUsersSecondPage(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	// Create 15 test users
	for i := 1; i <= 15; i++ {
		input := &CreateUserInput{
			Email:      fmt.Sprintf("user%d@example.com", i),
			FirstName:  fmt.Sprintf("User%d", i),
			LastName:   "Test",
			ExternalID: fmt.Sprintf("firebase|user%d", i),
			RoleID:     roleID,
		}
		_, err := testRepo.Create(ctx, input)
		require.NoError(t, err)
	}

	users, err := testRepo.List(ctx, 10, 10)

	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Len(t, users, 5)
}

// Test: List users when database is empty
func TestListUsersEmptyList(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()

	users, err := testRepo.List(ctx, 10, 0)

	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Len(t, users, 0)
}

// Test: Update user information
func TestUpdateUser(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	input := &CreateUserInput{
		Email:      "update@example.com",
		FirstName:  "Original",
		LastName:   "Name",
		ExternalID: "firebase|update",
		RoleID:     roleID,
	}
	created, err := testRepo.Create(ctx, input)
	require.NoError(t, err)

	newFirstName := "Updated"
	newLastName := "User"
	phone := "555-123-4567"
	updateInput := &UpdateUserInput{
		FirstName: &newFirstName,
		LastName:  &newLastName,
		Phone:     &phone,
	}

	updated, err := testRepo.Update(ctx, created.ID, updateInput)

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "Updated", updated.FirstName)
	assert.Equal(t, "User", updated.LastName)
	assert.NotNil(t, updated.Phone)
	assert.Equal(t, phone, *updated.Phone)
}

// Test: Attempt to update non-existent user
func TestUpdateUserNotFound(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()

	firstName := "Test"
	updateInput := &UpdateUserInput{
		FirstName: &firstName,
	}

	user, err := testRepo.Update(ctx, uuid.New(), updateInput)

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, user)
}

// Test: Deactivate user
func TestDeactivateUser(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	input := &CreateUserInput{
		Email:      "deactivate@example.com",
		FirstName:  "To",
		LastName:   "Deactivate",
		ExternalID: "firebase|deactivate",
		RoleID:     roleID,
	}
	created, err := testRepo.Create(ctx, input)
	require.NoError(t, err)

	// use the same user as the one performing deactivation since rbac is not implemented yet
	byUserID := created.ID

	err = testRepo.Deactivate(ctx, created.ID, &byUserID)
	assert.NoError(t, err)

	user, err := testRepo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.False(t, user.IsActive)
	assert.NotNil(t, user.DeactivatedAt)
	assert.NotNil(t, user.DeactivatedBy)
	assert.Equal(t, byUserID, *user.DeactivatedBy)
}

// Test: Deactivate already deactivated user (idempotency)
func TestDeactivateUserIdempotent(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	input := &CreateUserInput{
		Email:      "idempotent@example.com",
		FirstName:  "Idempotent",
		LastName:   "Test",
		ExternalID: "firebase|idempotent",
		RoleID:     roleID,
	}
	created, err := testRepo.Create(ctx, input)
	require.NoError(t, err)

	// use the same user as the one performing deactivation since rbac is not implemented yet
	byUserID := created.ID
	err = testRepo.Deactivate(ctx, created.ID, &byUserID)
	require.NoError(t, err)

	err = testRepo.Deactivate(ctx, created.ID, &byUserID)
	assert.NoError(t, err)
}

// Test: Reactivate deactivated user
func TestReactivateUser(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	input := &CreateUserInput{
		Email:      "reactivate@example.com",
		FirstName:  "To",
		LastName:   "Reactivate",
		ExternalID: "firebase|reactivate",
		RoleID:     roleID,
	}
	created, err := testRepo.Create(ctx, input)
	require.NoError(t, err)

	// use the same user as the one performing deactivation since rbac is not implemented yet
	byUserID := created.ID
	err = testRepo.Deactivate(ctx, created.ID, &byUserID)
	require.NoError(t, err)

	err = testRepo.Reactivate(ctx, created.ID)
	assert.NoError(t, err)
}

// Test: Increment token version
func TestIncrementTokenVersion(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	input := &CreateUserInput{
		Email:      "token@example.com",
		FirstName:  "Token",
		LastName:   "Test",
		ExternalID: "firebase|token",
		RoleID:     roleID,
	}
	created, err := testRepo.Create(ctx, input)
	require.NoError(t, err)
	initialVersion := created.TokenVersion

	err = testRepo.IncrementTokenVersion(ctx, created.ID)
	assert.NoError(t, err)

	user, err := testRepo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, initialVersion+1, user.TokenVersion)
}

// Test: Touch last sign-in timestamp
func TestTouchLastSignIn(t *testing.T) {
	cleanupTestData(t)
	ctx := context.Background()
	roleID := getRoleIDByCode(t, "admin")

	input := &CreateUserInput{
		Email:      "signin@example.com",
		FirstName:  "SignIn",
		LastName:   "Test",
		ExternalID: "firebase|signin",
		RoleID:     roleID,
	}
	created, err := testRepo.Create(ctx, input)
	require.NoError(t, err)
	require.Nil(t, created.LastSignInAt)

	beforeSignIn := time.Now()
	time.Sleep(100 * time.Millisecond)

	err = testRepo.TouchLastSignInNowByExternalID(ctx, "firebase|signin")
	assert.NoError(t, err)

	user, err := testRepo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.NotNil(t, user.LastSignInAt)
	assert.True(t, user.LastSignInAt.After(beforeSignIn))
}

// Test: Get role ID by role code
func TestGetRoleIDByCode(t *testing.T) {
	ctx := context.Background()

	roleID, err := testRepo.GetRoleIDByCode(ctx, "admin")

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, roleID)
}

// Test: Attempt to get role ID with non-existent code
func TestGetRoleIDByCodeNotFound(t *testing.T) {
	ctx := context.Background()

	roleID, err := testRepo.GetRoleIDByCode(ctx, "NONEXISTENT")

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Equal(t, uuid.Nil, roleID)
}
