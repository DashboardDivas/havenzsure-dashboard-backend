package workorder

import (
	"context"
	"database/sql"

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
		wo.code,
		wo.status,
		wo.created_at,
		wo.updated_at,
		c.first_name || ' ' || c.last_name AS full_name,
		c.email
	FROM app.work_orders AS wo
	JOIN app.customers AS c
		ON wo.customer_id = c.id
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
			&wl.CustomerFullName,
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
			wo.code,
			wo.status,
			wo.created_at AS date_received,
			wo.updated_at AS date_updated,

			c.first_name || ' ' || c.last_name AS full_name,
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
			v.color,

			i.insurance_company,
			i.agent_first_name || ' ' || i.agent_last_name AS agent_full_name,
			i.agent_phone,
			i.policy_number,
			i.claim_number

		FROM app.work_orders wo
		JOIN app.customers c ON wo.customer_id = c.id
		JOIN app.vehicles  v ON wo.vehicle_id  = v.id
		LEFT JOIN app.insurance i ON wo.id = i.work_order_id
		WHERE wo.code = $1
	`, code)

	var (
		insCompany    sql.NullString
		agentFullName sql.NullString
		agentPhone    sql.NullString
		policyNumber  sql.NullString
		claimNumber   sql.NullString
	)

	err := row.Scan(
		&detail.Code,
		&detail.Status,
		&detail.DateReceived,
		&detail.DateUpdated,

		&detail.Customer.FullName,
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

		&insCompany,
		&agentFullName,
		&agentPhone,
		&policyNumber,
		&claimNumber,
	)
	if err != nil {
		return detail, err
	}

	// if any of the insurance fields is not null, set insurance info as non-nil
	if insCompany.Valid || agentFullName.Valid ||
		agentPhone.Valid || policyNumber.Valid || claimNumber.Valid {

		detail.Insurance = &dto.Insurance{
			InsuranceCompany: insCompany.String,
			AgentFullName:    agentFullName.String,
			AgentPhone:       agentPhone.String,
			PolicyNumber:     policyNumber.String,
			ClaimNumber:      claimNumber.String,
		}
	} else {
		detail.Insurance = nil
	}
	return detail, nil
}
