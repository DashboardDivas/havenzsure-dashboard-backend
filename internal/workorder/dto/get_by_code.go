package dto

import (
	"time"
)

type WorkOrderDetail struct {
	Code         string          `json:"code"`
	Status       WorkOrderStatus `json:"status"`
	DateReceived time.Time       `json:"date_received"`
	DateUpdated  time.Time       `json:"date_updated"`
	Customer     Customer        `json:"customer"`
	Vehicle      Vehicle         `json:"vehicle"`
	Insurance    *Insurance      `json:"insurance,omitempty"`
}
