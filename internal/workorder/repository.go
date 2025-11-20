package workorder

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/workorder/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListWorkOrder(ctx context.Context) ([]dto.WorkOrderListItem, error)
	GetWorkOrderByCode(ctx context.Context, code string) (dto.WorkOrderDetail, error)
	CreateWorkOrder(ctx context.Context, payload dto.IntakePayload) (dto.WorkOrderDetail, error)
	//EditorIntake(ctx context.Context, code string, payload dto.IntakeEditPayload) (dto.WorkOrderDetail, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) ListWorkOrder(ctx context.Context) ([]dto.WorkOrderListItem, error) {
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

func (r *repository) GetWorkOrderByCode(ctx context.Context, code string) (dto.WorkOrderDetail, error) {
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

		detail.Insurance = &dto.InsuranceDetail{
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

func (r *repository) CreateWorkOrder(ctx context.Context, payload dto.IntakePayload) (dto.WorkOrderDetail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return dto.WorkOrderDetail{}, err
	}

	defer tx.Rollback(ctx)

	//define customerID to capture inserted customer ID
	var customerID uuid.UUID
	//Execute insert statement and capture the returned ID
	//1. Insert customer and capture customer ID
	err = tx.QueryRow(ctx, `
		INSERT INTO app.customers
		(first_name, last_name, address, city, postal_code, province, email, phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
		`, payload.Customer.FirstName, payload.Customer.LastName, payload.Customer.Address, payload.Customer.City, payload.Customer.PostalCode, payload.Customer.Province, payload.Customer.Email,
		payload.Customer.Phone,
	).Scan(&customerID)
	if err != nil {
		return dto.WorkOrderDetail{}, fmt.Errorf("insert customer: %w", err)
	}
	//2. Insert vehicle and capture vehicle ID
	var vehicleID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO app.vehicles
		(plate_number, make, model, body_style, model_year, vin, color)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
		`, payload.Vehicle.PlateNo, payload.Vehicle.Make, payload.Vehicle.Model, payload.Vehicle.BodyStyle, payload.Vehicle.ModelYear, payload.Vehicle.VIN,
		payload.Vehicle.Color,
	).Scan(&vehicleID)
	if err != nil {
		return dto.WorkOrderDetail{}, fmt.Errorf("insert vehicle: %w", err)
	}

	//3. Insert work order and capture work order ID and code

	//真的没招了！！！这里！！现在是！！ shop_id  hard code 一个值进去！！！
	//COMMIT 之前一定要改！！！！！！
	//但是要怎么改！！！没有auth CONTEXT！！我也改不了啊！！！！
	//只能先这样了！！！！！
	var shopID uuid.UUID = uuid.MustParse("f4975c12-4fc4-4fe0-9c33-92bd2cc85883")
	//！！！！！！！！！！！！！！！！！！！！！！！！！！

	var workOrderID uuid.UUID
	var workOrderCode string
	err = tx.QueryRow(ctx, `
		INSERT INTO app.work_orders
		(customer_id, vehicle_id, shop_id)
		VALUES ($1, $2, $3)
		RETURNING id, code
		`, customerID,
		vehicleID,
		shopID,
	).Scan(&workOrderID, &workOrderCode)
	if err != nil {
		return dto.WorkOrderDetail{}, fmt.Errorf("insert work_order: %w", err)
	}
	//4. Insert insurance if provided
	ins := payload.Insurance
	// Scenarios:
	// 1) ins == nil               -> no insurance, skip
	// 2) ins != nil && IsEmpty()  -> treat as no insurance, skip
	// 3) ins != nil && !IsEmpty() but fail validation -> error
	// 4) ins != nil && !IsEmpty() && pass validation -> upsert (insert now, update future)

	if ins == nil {
		// 1. ins == nil               -> no insurance, skip
	} else if ins.IsEmpty() {
		// 2. ins != nil && IsEmpty()  -> treat as no insurance, skip
	} else {
		// 3. ins != nil && !IsEmpty() but fail validation -> error
		if err := ins.Validate(); err != nil {
			return dto.WorkOrderDetail{}, fmt.Errorf("invalid insurance info: %w", err)
		}

		// 4. ins != nil && !IsEmpty() && pass validation -> upsert (insert now, update future)
		if err := r.UpsertInsurance(ctx, tx, workOrderID, *ins); err != nil {
			return dto.WorkOrderDetail{}, fmt.Errorf("failed to upsert insurance: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.WorkOrderDetail{}, err
	}
	return r.GetWorkOrderByCode(ctx, workOrderCode)
}

type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (r *repository) UpsertInsurance(
	ctx context.Context,
	ex execer,
	workOrderID uuid.UUID,
	ins dto.InsuranceIntake) error {
	_, err := ex.Exec(ctx, `
		INSERT INTO app.insurance
		(work_order_id, insurance_company, agent_first_name, agent_last_name, agent_phone, policy_number, claim_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (work_order_id) DO UPDATE SET
			insurance_company = EXCLUDED.insurance_company,
			agent_first_name = EXCLUDED.agent_first_name,
			agent_last_name = EXCLUDED.agent_last_name,
			agent_phone = EXCLUDED.agent_phone,
			policy_number = EXCLUDED.policy_number,
			claim_number = EXCLUDED.claim_number
		`, workOrderID, nullIfEmpty(ins.InsuranceCompany), nullIfEmpty(ins.AgentFirstName), nullIfEmpty(ins.AgentLastName), nullIfEmpty(ins.AgentPhone), nullIfEmpty(ins.PolicyNumber), nullIfEmpty(ins.ClaimNumber))
	return err
}
