package dto

import (
	"time"
)

type WorkOrderListItem struct {
	Code      string          `db:"work_order_code" json:"code"`
	Status    WorkOrderStatus `json:"status"        db:"status"`
	CreatedAt time.Time       `json:"createdAt"     db:"created_at"`
	UpdatedAt time.Time       `json:"updatedAt"     db:"updated_at"`

	// Joined from customers
	CustomerFirstName string `json:"customerFirstName" db:"first_name"`
	CustomerLastName  string `json:"customerLastName" db:"last_name"`
	CustomerEmail     string `json:"customerEmail" db:"email"`
}
