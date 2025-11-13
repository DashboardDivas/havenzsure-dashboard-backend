package dto

import (
	"time"
)

type WorkOrderListItem struct {
	Code      string          `json:"code"`
	Status    WorkOrderStatus `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`

	// Joined from customers
	CustomerFullName string `json:"customerFullName"`
	CustomerEmail    string `json:"customerEmail"`
}
