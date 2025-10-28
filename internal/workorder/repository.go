package workorder

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	List(ctx context.Context) ([]WorkOrder, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]WorkOrder, error) {
	rows, err := r.db.Query(ctx, `
		SELECT work_order_id, work_order_code, customer_id, shop_id, status, damage_date, created_at
		FROM app.work_orders
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []WorkOrder
	for rows.Next() {
		var w WorkOrder
		err := rows.Scan(
			&w.ID,
			&w.Code,
			&w.CustomerID,
			&w.ShopID,
			&w.Status,
			&w.DamageDate,
			&w.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, w)
	}
	return result, rows.Err()
}
