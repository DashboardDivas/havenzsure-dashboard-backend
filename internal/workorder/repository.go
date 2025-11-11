package workorder

import (
	"context"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/workorder/dto"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	List(ctx context.Context) ([]dto.WorkOrderListItem, error)
	GetByCode(ctx context.Context, code string) (dto.WorkOrderDetail, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]dto.WorkOrderListItem, error) {
	rows, err := r.db.Query(ctx, `
	SELECT 
		wo.work_order_code,
		wo.status,
		wo.created_at,
		wo.updated_at,
		c.first_name,
		c.last_name,
		c.email
	FROM app.work_orders AS wo
	JOIN app.customers AS c
		ON wo.customer_id = c.customer_id
	ORDER BY wo.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.WorkOrderListItem
	for rows.Next() {
		var wl dto.WorkOrderListItem
		err := rows.Scan(
			&wl.Code,
			&wl.Status,
			&wl.CreatedAt,
			&wl.UpdatedAt,
			&wl.CustomerFirstName,
			&wl.CustomerLastName,
			&wl.CustomerEmail,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, wl)
	}
	return result, rows.Err()
}

func (r *repository) GetByCode(ctx context.Context, code string) (dto.WorkOrderDetail, error) {
	var detail dto.WorkOrderDetail
	row := r.db.QueryRow(ctx, `
		SELECT
			wo.work_order_code,
			wo.status,
			wo.created_at AS date_received,
			wo.updated_at AS date_updated,

			c.first_name,
			c.last_name,
			c.phone,
			c.email,
			c.address,
			c.city,
			c.province,
			c.postal_code,

			v.plate_number,
			v.make,
			v.model,
			v.body_style,
			v.model_year,
			v.vin,
			v.color

		FROM app.work_orders wo
		JOIN app.customers c ON wo.customer_id = c.customer_id
		JOIN app.vehicles  v ON wo.vehicle_id  = v.vehicle_id
		WHERE wo.work_order_code = $1
	`, code)

	err := row.Scan(
		&detail.Code,
		&detail.Status,
		&detail.DateReceived,
		&detail.DateUpdated,

		&detail.Customer.FirstName,
		&detail.Customer.LastName,
		&detail.Customer.Phone,
		&detail.Customer.Email,
		&detail.Customer.Address,
		&detail.Customer.City,
		&detail.Customer.Province,
		&detail.Customer.PostalCode,

		&detail.Vehicle.PlateNo,
		&detail.Vehicle.Make,
		&detail.Vehicle.Model,
		&detail.Vehicle.BodyStyle,
		&detail.Vehicle.ModelYear,
		&detail.Vehicle.VIN,
		&detail.Vehicle.Color,

		// &detail.InsuranceInfo.InsuranceCompany,
		// &detail.InsuranceInfo.AgentFirstName,
		// &detail.InsuranceInfo.AgentLastName,
		// &detail.InsuranceInfo.AgentPhone,
		// &detail.InsuranceInfo.PolicyNumber,
		// &detail.InsuranceInfo.ClaimNumber,
	)
	if err != nil {
		return detail, err
	}
	return detail, nil
}
