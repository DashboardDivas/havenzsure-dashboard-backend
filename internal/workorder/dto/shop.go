package dto

import "github.com/google/uuid"

// For Request: Intake, Transfer
type ShopRef struct {
	ShopID   uuid.UUID `json:"shopId"`
	ShopCode string    `json:"shopCode"`
}

// For Response: WorkOrderDetail and WorkOrderListItem
type ShopSummary struct {
	ShopID   uuid.UUID `json:"shopId"`
	ShopCode string    `json:"shopCode"`
	ShopName string    `json:"shopName"`
}
