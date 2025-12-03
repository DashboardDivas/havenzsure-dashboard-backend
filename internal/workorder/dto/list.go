package dto

import (
	"time"

	"github.com/google/uuid"
)

type WorkOrderListItem struct {
	ID        uuid.UUID       `json:"id"`
	Code      string          `json:"code"`
	Status    WorkOrderStatus `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`

	// Joined from customers
	CustomerFullName string `json:"customerFullName"`
	CustomerEmail    string `json:"customerEmail"`

	Shop ShopSummary `json:"shop,omitempty"`
}
