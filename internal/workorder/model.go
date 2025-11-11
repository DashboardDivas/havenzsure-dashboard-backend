package workorder

import (
	"time"

	"github.com/google/uuid"
)

// Customer ↔ app.customers
type Customer struct {
	CustomerID uuid.UUID `db:"customer_id"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	Address    string    `db:"address"`
	City       string    `db:"city"`
	PostalCode string    `db:"postal_code"`
	Province   string    `db:"province"`
	Email      string    `db:"email"`
	Phone      string    `db:"phone"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// Vehicle ↔ app.vehicles
type Vehicle struct {
	VehicleID uuid.UUID `db:"vehicle_id"`
	Make      string    `db:"make"`
	Model     string    `db:"model"`
	BodyStyle string    `db:"body_style"`
	ModelYear int       `db:"model_year"`
	VIN       string    `db:"vin"`
	Color     string    `db:"color"`
	PlateNo   string    `db:"plate_number"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type InsuranceInfo struct {
	WorkOrderID      uuid.UUID `db:"work_order_id"`
	InsuranceCompany string    `db:"insurance_company"`
	AgentFirstName   string    `db:"agent_first_name"`
	AgentLastName    string    `db:"agent_last_name"`
	AgentPhone       string    `db:"agent_phone"`
	PolicyNumber     string    `db:"policy_number"`
	ClaimNumber      string    `db:"claim_number"`
}

// WorkOrder represents one row from app.work_orders
type WorkOrder struct {
	ID              uuid.UUID  `db:"work_order_id"`
	Code            string     `db:"work_order_code"`
	CustomerID      uuid.UUID  `db:"customer_id"`
	ShopID          uuid.UUID  `db:"shop_id"`
	VehicleID       uuid.UUID  `db:"vehicle_id"`
	CreatedByUserID uuid.UUID  `db:"created_by_user_id"`
	Status          string     `db:"status"`
	DamageDate      *time.Time `db:"damage_date"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}
