// using Chatgpt to help with writing comments and check query errors

package shop

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the persistence contract for shops.
// The service layer depends on this interface, not on concrete DB details.
type Repository interface {
	CreateShop(ctx context.Context, shop *Shop) error
	GetShopByCode(ctx context.Context, code string) (*Shop, error)
	UpdateShop(ctx context.Context, shop *Shop) (*Shop, error)
	ListShops(ctx context.Context, limit, offset int) ([]*Shop, error)
}

// PGRepository is a Postgres implementation of Repository using pgxpool.
type PGRepository struct {
	db *pgxpool.Pool
}

// NewShopRepository constructs a Postgres-backed repository.
func NewShopRepository(db *pgxpool.Pool) *PGRepository {
	return &PGRepository{db: db}
}

// Implementing the Repository interface methods
func (r *PGRepository) CreateShop(ctx context.Context, shop *Shop) error {
	const q = `INSERT INTO app.shop (code, shop_name, status, address, city, province, postal_code, contact_name, phone, email)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING id, code, shop_name;`
	if err := r.db.QueryRow(ctx, q,
		shop.Code, shop.ShopName, shop.Status, shop.Address,
		shop.City, shop.Province, shop.PostalCode,
		shop.ContactName, shop.Phone, shop.Email,
	).Scan(&shop.ID, &shop.Code, &shop.ShopName); err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) {
			switch pe.Code {
			case pgerrcode.UniqueViolation:
				return ErrConflict
			case pgerrcode.NotNullViolation, pgerrcode.CheckViolation:
				return ErrInvalidInput
			}
		}
		return fmt.Errorf("failed to create shop: %w", err)
	}
	return nil
}

func (r *PGRepository) GetShopByCode(ctx context.Context, code string) (*Shop, error) {
	const q = `
SELECT id, code, shop_name, status, address, city, province, postal_code, contact_name, phone, email, created_at, updated_at
FROM app.shop
WHERE code=$1;`
	var s Shop
	if err := r.db.QueryRow(ctx, q, code).Scan(&s.ID, &s.Code, &s.ShopName, &s.Status, &s.Address, &s.City, &s.Province, &s.PostalCode,
		&s.ContactName, &s.Phone, &s.Email, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get the shop by code: %w", err)
	}
	return &s, nil
}

func (r *PGRepository) ListShops(ctx context.Context, limit, offset int) ([]*Shop, error) {
	const q = `
SELECT id, code, shop_name, status, city, province, contact_name, phone
FROM app.shop
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;`
	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute list shop query: %w", err)
	}
	defer rows.Close()

	var shops []*Shop
	for rows.Next() {
		shop := new(Shop)
		if err := rows.Scan(&shop.ID, &shop.Code, &shop.ShopName, &shop.Status, &shop.City, &shop.Province, &shop.ContactName, &shop.Phone); err != nil {
			return nil, fmt.Errorf("failed to scan shop row: %w", err)
		}
		shops = append(shops, shop)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to list shops: %w", err)
	}
	return shops, nil
}

func (r *PGRepository) UpdateShop(ctx context.Context, s *Shop) (*Shop, error) {
	const q = `
UPDATE app.shop
SET code=$2, shop_name=$3, status=$4, address=$5, city=$6, province=$7, postal_code=$8,
    contact_name=$9, phone=$10, email=$11
WHERE id=$1
RETURNING id, code, shop_name, status, address, city, province, postal_code,
          contact_name, phone, email, created_at, updated_at;`
	row := r.db.QueryRow(ctx, q,
		s.ID, s.Code, s.ShopName, s.Status, s.Address, s.City, s.Province, s.PostalCode,
		s.ContactName, s.Phone, s.Email,
	)
	var updatedShop Shop
	if err := row.Scan(
		&updatedShop.ID, &updatedShop.Code, &updatedShop.ShopName, &updatedShop.Status, &updatedShop.Address, &updatedShop.City, &updatedShop.Province, &updatedShop.PostalCode,
		&updatedShop.ContactName, &updatedShop.Phone, &updatedShop.Email, &updatedShop.CreatedAt, &updatedShop.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		var pe *pgconn.PgError
		if errors.As(err, &pe) {
			switch pe.Code {
			case pgerrcode.UniqueViolation:
				return nil, ErrConflict
			case pgerrcode.NotNullViolation, pgerrcode.CheckViolation:
				return nil, ErrInvalidInput
			}
		}
		return nil, fmt.Errorf("failed to update shop: %w", err)
	}
	return &updatedShop, nil
}
