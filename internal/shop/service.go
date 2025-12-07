// Citation:　Use of Chatgpt for writing service layer validation, normalization and writing comments
// Edited and reviewed by AN-NI HUANG
// Date: 2025-10-19
package shop

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Importing regexp for validation
var (
	postalCodeRegex = regexp.MustCompile(`^[A-Z][0-9][A-Z][0-9][A-Z][0-9]$`)
	phoneRegex      = regexp.MustCompile(`^[0-9]{3}-[0-9]{3}-[0-9]{4}$`)
	emailRegex      = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)

	validProvinces = map[string]bool{
		"AB": true, "BC": true, "MB": true, "NB": true,
		"NL": true, "NT": true, "NS": true, "NU": true,
		"ON": true, "PE": true, "QC": true, "SK": true, "YT": true,
	}
)

// ShopService defines business operations for the Shop domain.
// The service coordinates domain logic and delegates persistence to Repository.
// It should not know anything about HTTP, SQL, or other infrastructure details.
type ShopService interface {

	// CreateShop creates a new shop. On success, the input *Shop is updated in-place
	// with generated values (e.g., ID) by the repository. Returns a domain error
	// such as ErrInvalidInput or ErrConflict where appropriate.
	CreateShop(ctx context.Context, shop *Shop) error

	// GetShopByID fetches a shop by its unique ID.
	// Returns (*Shop, nil) on success, or ErrNotFound if no record exists.
	GetShopByID(ctx context.Context, id uuid.UUID) (*Shop, error)

	// GetShopIDByCode fetches a shop's ID by its unique code.
	// Returns (uuid.UUID, nil) on success, or ErrNotFound if no record exists.
	GetShopIDByCode(ctx context.Context, code string) (uuid.UUID, error)

	// UpdateShop updates an existing shop identified by its ID (already set on s).
	// Returns the freshly-updated copy from the database (useful for refreshing UI).
	UpdateShop(ctx context.Context, id uuid.UUID, shop *Shop) (*Shop, error)

	// ListShops returns a page of shops. The service applies safe defaults for
	// limit/offset if the caller passes invalid values. An empty slice with nil error
	// means “no rows for this page”, not an error.
	ListShops(ctx context.Context, limit, offset int) ([]*Shop, error)
}

// service is the concrete implementation of ShopService.
// It depends only on the Repository interface (infrastructure-agnostic).
type service struct {
	repo Repository
}

// NewService constructs a Shop service that uses the given Repository.
// Typical wiring (in your server/app layer):
//
//	repo := shop.NewShopRepository(db)  // concrete PG repository
//	svc  := shop.NewService(repo)       // service depends on the interface
func NewService(repo Repository) *service {
	return &service{repo: repo}
}

// CreateShop performs normalization and validation before delegating to the repository.
// The repository fills in generated fields via the pointer (e.g., shop.ID).
// Errors are wrapped with context ("service create shop") so logs show WHERE
// failures occurred; the original error is preserved with %w for errors.Is/As.
func (svc *service) CreateShop(ctx context.Context, s *Shop) error {
	normalizeShop(s)

	if err := validateShop(s); err != nil {
		return err
	}

	if err := svc.repo.CreateShop(ctx, s); err != nil {
		return fmt.Errorf("service create shop: %w", err)
	}
	return nil
}

func (s *service) GetShopIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	if code == "" {
		return uuid.Nil, ErrInvalidInput
	}
	shopID, err := s.repo.GetShopIDByCode(ctx, code)
	if err != nil {
		return uuid.Nil, fmt.Errorf("service get shop id by code: %w", err)
	}
	return shopID, nil
}

// GetShopByID validates the input ID and delegates to the repository.
// Returns ErrInvalidInput if the ID is nil.
func (s *service) GetShopByID(ctx context.Context, id uuid.UUID) (*Shop, error) {
	if id == uuid.Nil {
		return nil, NewValidationError("id", "invalid shop ID")
	}

	shop, err := s.repo.GetShopByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service get shop by id: %w", err)
	}

	return shop, nil
}

// ListShops applies defensive defaults for pagination and delegates to the repository.
// Empty results are returned as an empty slice and nil error (not ErrNotFound).
func (svc *service) ListShops(ctx context.Context, limit, offset int) ([]*Shop, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	shops, err := svc.repo.ListShops(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service list shops: %w", err)
	}
	return shops, nil
}

// UpdateShop performs normalization and validation before delegating to the repository.
// Returns the updated shop on success. Errors are wrapped with context.
func (svc *service) UpdateShop(ctx context.Context, id uuid.UUID, s *Shop) (*Shop, error) {
	if s == nil {
		return nil, ErrInvalidInput
	}

	if id == uuid.Nil {
		return nil, ErrInvalidInput
	}

	normalizeShop(s)

	if err := validateShop(s); err != nil {
		return nil, err
	}
	shop, err := svc.repo.UpdateShop(ctx, id, s)
	if err != nil {
		return nil, fmt.Errorf("service update shop: %w", err)
	}
	return shop, nil
}

// normalizeShop trims whitespace and normalizes casing for certain fields.
func normalizeShop(s *Shop) {
	s.Code = strings.ToUpper(strings.TrimSpace(s.Code))
	s.ShopName = strings.TrimSpace(s.ShopName)
	s.Address = strings.TrimSpace(s.Address)
	s.City = strings.TrimSpace(s.City)
	s.Province = strings.ToUpper(strings.TrimSpace(s.Province))
	s.PostalCode = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s.PostalCode), " ", ""))
	s.ContactName = strings.TrimSpace(s.ContactName)
	s.Phone = strings.TrimSpace(s.Phone)
	s.Email = strings.ToLower(strings.TrimSpace(s.Email))
}

// validateShop checks the Shop fields for validity and returns appropriate ValidationError instances.
func validateShop(s *Shop) error {

	if s.Code == "" {
		return NewValidationError("code", "code is required")
	}
	if len(s.Code) < 2 {
		return NewValidationError("code", "code must be at least 2 characters")
	}
	if len(s.Code) > 10 {
		return NewValidationError("code", "code must not exceed 10 characters")
	}

	if s.ShopName == "" {
		return NewValidationError("shopName", "shop name is required")
	}

	if s.Status != Active && s.Status != Inactive {
		return NewValidationError("status", "status must be 'active' or 'inactive'")
	}

	if s.Address == "" {
		return NewValidationError("address", "address is required")
	}

	if s.City == "" {
		return NewValidationError("city", "city is required")
	}

	if s.Province == "" {
		return NewValidationError("province", "province is required")
	}
	if !validProvinces[s.Province] {
		return NewValidationError("province", "invalid province code")
	}

	if s.PostalCode == "" {
		return NewValidationError("postalCode", "postal code is required")
	}
	if !postalCodeRegex.MatchString(s.PostalCode) {
		return NewValidationError("postalCode", "postal code must be in format A1A1A1")
	}

	if s.ContactName == "" {
		return NewValidationError("contactName", "contact name is required")
	}

	if s.Phone == "" {
		return NewValidationError("phone", "phone is required")
	}
	if !phoneRegex.MatchString(s.Phone) {
		return NewValidationError("phone", "phone must be in format 403-555-1234")
	}

	if s.Email == "" {
		return NewValidationError("email", "email is required")
	}
	if !emailRegex.MatchString(s.Email) {
		return NewValidationError("email", "invalid email format")
	}

	return nil
}
