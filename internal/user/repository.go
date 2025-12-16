package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for user data operations
type Repository interface {
	Create(ctx context.Context, in *CreateUserInput) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByCode(ctx context.Context, code string) (*User, error)
	GetByExternalID(ctx context.Context, externalID string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Update(ctx context.Context, id uuid.UUID, in *UpdateUserInput) (*User, error)
	Deactivate(ctx context.Context, id uuid.UUID, byUserID *uuid.UUID) error
	Reactivate(ctx context.Context, id uuid.UUID) error
	IncrementTokenVersion(ctx context.Context, id uuid.UUID) error
	TouchLastSignInNowByExternalID(ctx context.Context, externalID string) error
	GetRoleIDByCode(ctx context.Context, code string) (uuid.UUID, error)
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error
}

type pgRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) Repository {
	return &pgRepo{db: db}
}

/* ---------- repo-private row (db tags) ---------- */

type userRow struct {
	ID            uuid.UUID  `db:"id"`
	Code          string     `db:"code"`
	Email         string     `db:"email"`
	FirstName     string     `db:"first_name"`
	LastName      string     `db:"last_name"`
	Phone         *string    `db:"phone"`
	ImageURL      *string    `db:"image_url"`
	ExternalID    string     `db:"external_id"`
	EmailVerified bool       `db:"email_verified"`
	IsActive      bool       `db:"is_active"`
	DeactivatedAt *time.Time `db:"deactivated_at"`
	DeactivatedBy *uuid.UUID `db:"deactivated_by"`
	TokenVersion  int        `db:"token_version"`
	ShopID        *uuid.UUID `db:"shop_id"`
	RoleID        uuid.UUID  `db:"role_id"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	LastSignInAt  *time.Time `db:"last_sign_in_at"`

	// Role join fields
	RoleCode     *string    `db:"role_code"`
	RoleName     *string    `db:"role_name"`
	RoleIsSystem *bool      `db:"role_is_system"`
	RoleCreated  *time.Time `db:"role_created_at"`
	RoleUpdated  *time.Time `db:"role_updated_at"`

	// Shop join fields
	ShopCode *string `db:"shop_code"`
	ShopName *string `db:"shop_name"`
}

func (r userRow) toDomain() *User {
	user := &User{
		ID:            r.ID,
		Code:          r.Code,
		Email:         r.Email,
		FirstName:     r.FirstName,
		LastName:      r.LastName,
		Phone:         r.Phone,
		ImageURL:      r.ImageURL,
		ExternalID:    r.ExternalID,
		EmailVerified: r.EmailVerified,
		IsActive:      r.IsActive,
		DeactivatedAt: r.DeactivatedAt,
		DeactivatedBy: r.DeactivatedBy,
		TokenVersion:  r.TokenVersion,
		ShopID:        r.ShopID,
		RoleID:        r.RoleID,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		LastSignInAt:  r.LastSignInAt,
		Role: Role{
			Code:      *r.RoleCode,
			Name:      *r.RoleName,
			IsSystem:  *r.RoleIsSystem,
			CreatedAt: *r.RoleCreated,
			UpdatedAt: *r.RoleUpdated,
		},
	}
	if r.ShopCode != nil && r.ShopName != nil {
		user.Shop = &Shop{
			Code: *r.ShopCode,
			Name: *r.ShopName,
		}
	}
	return user
}

/* ---------- error mapping ---------- */

func mapPgError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return ErrConflict
		case pgerrcode.ForeignKeyViolation, pgerrcode.CheckViolation, pgerrcode.NotNullViolation:
			return ErrInvalidInput
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

/* ---------- base SELECT with role join ---------- */

const baseSelect = `
SELECT 
  u.id, u.code, u.email, u.first_name, u.last_name, u.phone, u.image_url,
  u.external_id, u.email_verified, u.is_active, u.deactivated_at, u.deactivated_by,
  u.token_version, u.shop_id, u.role_id, u.created_at, u.updated_at, u.last_sign_in_at,
  r.code        AS role_code,
  r.name        AS role_name,
  r.is_system   AS role_is_system,
  r.created_at  AS role_created_at,
  r.updated_at  AS role_updated_at,
  s.code        AS shop_code,
  s.shop_name        AS shop_name
FROM app.users u
INNER JOIN app.roles r ON r.id = u.role_id
LEFT JOIN app.shop s ON s.id = u.shop_id
`

/* ---------- queries ---------- */

func (r *pgRepo) Create(ctx context.Context, in *CreateUserInput) (*User, error) {
	if in == nil || in.Email == "" || in.FirstName == "" || in.LastName == "" || in.ExternalID == "" || in.RoleID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	const q = `
INSERT INTO app.users (
  email, first_name, last_name, phone, image_url,
  external_id, email_verified, shop_id, role_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING id
`
	var id uuid.UUID
	if err := r.db.QueryRow(ctx, q,
		in.Email, in.FirstName, in.LastName, in.Phone, in.ImageURL,
		in.ExternalID, in.EmailVerified, in.ShopID, in.RoleID,
	).Scan(&id); err != nil {
		return nil, mapPgError(err)
	}
	return r.GetByID(ctx, id)
}

func (r *pgRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	rows, err := r.db.Query(ctx, baseSelect+` WHERE u.id=$1`, id)
	if err != nil {
		return nil, mapPgError(err)
	}
	defer rows.Close()

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		return nil, mapPgError(err)
	}
	return row.toDomain(), nil
}

func (r *pgRepo) GetByCode(ctx context.Context, code string) (*User, error) {
	rows, err := r.db.Query(ctx, baseSelect+` WHERE u.code=$1`, code)
	if err != nil {
		return nil, mapPgError(err)
	}
	defer rows.Close()

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		return nil, mapPgError(err)
	}
	return row.toDomain(), nil
}

func (r *pgRepo) GetByExternalID(ctx context.Context, externalID string) (*User, error) {
	rows, err := r.db.Query(ctx, baseSelect+` WHERE u.external_id=$1`, externalID)
	if err != nil {
		return nil, mapPgError(err)
	}
	defer rows.Close()

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		return nil, mapPgError(err)
	}
	return row.toDomain(), nil
}

func (r *pgRepo) List(ctx context.Context, limit, offset int) ([]*User, error) {

	rows, err := r.db.Query(ctx, baseSelect+` ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, mapPgError(err)
	}
	defer rows.Close()

	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		return nil, mapPgError(err)
	}

	out := make([]*User, 0, len(list))
	for _, user_row := range list {
		out = append(out, user_row.toDomain())
	}
	return out, nil
}

func (r *pgRepo) Update(ctx context.Context, id uuid.UUID, in *UpdateUserInput) (*User, error) {
	if in == nil {
		return nil, ErrInvalidInput
	}

	const q = `
UPDATE app.users AS u
SET
  first_name     = COALESCE($2, u.first_name),
  last_name      = COALESCE($3, u.last_name),
  phone          = COALESCE($4, u.phone),
  image_url      = COALESCE($5, u.image_url),
  email_verified = COALESCE($6, u.email_verified),
  shop_id        = COALESCE($7, u.shop_id),
  role_id        = COALESCE($8, u.role_id)
WHERE u.id = $1
RETURNING u.id
`
	var ret uuid.UUID
	if err := r.db.QueryRow(ctx, q,
		id,
		in.FirstName,
		in.LastName,
		in.Phone,
		in.ImageURL,
		in.EmailVerified,
		in.ShopID,
		in.RoleID,
	).Scan(&ret); err != nil {
		return nil, mapPgError(err)
	}
	return r.GetByID(ctx, ret)
}

func (r *pgRepo) Deactivate(ctx context.Context, id uuid.UUID, byUserID *uuid.UUID) error {
	const q = `
UPDATE app.users
SET is_active = FALSE,
    deactivated_by = $2
WHERE id = $1 AND is_active = TRUE
`
	ct, err := r.db.Exec(ctx, q, id, byUserID)
	if err != nil {
		return mapPgError(err)
	}
	if ct.RowsAffected() == 0 {
		if _, err := r.GetByID(ctx, id); err != nil {
			return ErrNotFound
		}
		return nil // idempotent
	}
	return nil
}

func (r *pgRepo) Reactivate(ctx context.Context, id uuid.UUID) error {
	const q = `
UPDATE app.users
SET is_active = TRUE
WHERE id = $1 AND is_active = FALSE
`
	ct, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return mapPgError(err)
	}
	if ct.RowsAffected() == 0 {
		if _, err := r.GetByID(ctx, id); err != nil {
			return ErrNotFound
		}
		return nil // idempotent
	}
	return nil
}

func (r *pgRepo) IncrementTokenVersion(ctx context.Context, id uuid.UUID) error {
	ct, err := r.db.Exec(ctx, `UPDATE app.users SET token_version = token_version + 1 WHERE id=$1`, id)
	if err != nil {
		return mapPgError(err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepo) TouchLastSignInNowByExternalID(ctx context.Context, externalID string) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE app.users SET last_sign_in_at = NOW() WHERE external_id = $1`, externalID)
	if err != nil {
		return mapPgError(err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepo) GetRoleIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		"SELECT id FROM app.roles WHERE code = $1", code).Scan(&id)
	if err != nil {
		return uuid.Nil, mapPgError(err)
	}
	return id, nil
}

func (r *pgRepo) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	const q = `
UPDATE app.users
SET email_verified = TRUE,
    updated_at     = NOW()
WHERE id = $1 AND email_verified = FALSE
`
	ct, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return mapPgError(err)
	}
	if ct.RowsAffected() == 0 {
		// Either not found or already verified
		// Return error only if not found
		if _, err := r.GetByID(ctx, id); err != nil {
			return ErrNotFound
		}
	}
	return nil
}
