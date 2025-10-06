// using Chatgpt to help with writing comments
package shop

import (
	"context"
	"fmt"
)

// TODO: Implement validateShopInput in service layer.

// ShopService defines business operations for the Shop domain.
// The service coordinates domain logic and delegates persistence to Repository.
// It should not know anything about HTTP, SQL, or other infrastructure details.
type ShopService interface {

	// CreateShop creates a new shop. On success, the input *Shop is updated in-place
	// with generated values (e.g., ID) by the repository. Returns a domain error
	// such as ErrInvalidInput or ErrConflict where appropriate.
	CreateShop(ctx context.Context, shop *Shop) error

	// GetShopByCode fetches a shop by its business code.
	// Returns (*Shop, nil) on success, or ErrNotFound if no record exists.
	GetShopByCode(ctx context.Context, code string) (*Shop, error)

	// UpdateShop updates an existing shop identified by its ID (already set on s).
	// Returns the freshly-updated copy from the database (useful for refreshing UI).
	UpdateShop(ctx context.Context, shop *Shop) (*Shop, error)

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

// CreateShop performs minimal validation, then delegates to the repository.
// The repository fills in generated fields via the pointer (e.g., shop.ID).
// Errors are wrapped with context ("service create shop") so logs show WHERE
// failures occurred; the original error is preserved with %w for errors.Is/As.
func (svc *service) CreateShop(ctx context.Context, s *Shop) error {
	if s == nil {
		return ErrInvalidInput
	}
	if err := svc.repo.CreateShop(ctx, s); err != nil {
		return fmt.Errorf("service create shop: %w", err)
	}
	return nil
}

// GetShopByCode validates the input, then calls the repository.
// Returns ErrInvalidInput when code is empty; ErrNotFound if no row matches.
func (svc *service) GetShopByCode(ctx context.Context, code string) (*Shop, error) {
	if code == "" {
		return nil, ErrInvalidInput
	}
	shop, err := svc.repo.GetShopByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("service get shop by code: %w", err)
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

// UpdateShop validates the input pointer and delegates to the repository,
// which returns a fully refreshed record from the database.
// Useful when the UI needs to reflect server-side canonical values (e.g., trimmed
// fields, normalized casing, updated timestamps).
func (svc *service) UpdateShop(ctx context.Context, s *Shop) (*Shop, error) {
	if s == nil {
		return nil, ErrInvalidInput
	}
	shop, err := svc.repo.UpdateShop(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("service update shop: %w", err)
	}
	return shop, nil
}
