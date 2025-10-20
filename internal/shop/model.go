//Citation:
// GPT-4.0
// "According to my table structure in this file(20250915235050_create_shop_table), help me quick with struct?"
// Response:
// package entities
// import "time"

// ShopStatus mirrors the SQL enum app.shop_status ('active','inactive')
// type ShopStatus string

// const (
// 	ShopStatusActive   ShopStatus = "active"
// 	ShopStatusInactive ShopStatus = "inactive"
// )

//	type Shop struct {
//		ID          int64       `db:"id"           json:"id"`
//		Code        string      `db:"code"         json:"code"`
//		ShopName    string      `db:"shop_name"    json:"shopName"`
//		Status      ShopStatus  `db:"status"       json:"status"`
//		Address     string      `db:"address"      json:"address"`
//		City        string      `db:"city"         json:"city"`
//		Province    string      `db:"province"     json:"province"`   // e.g., AB/BC/ON
//		PostalCode  string      `db:"postal_code"  json:"postalCode"` // e.g., T2P 2B5
//		ContactName string      `db:"contact_name" json:"contactName"`
//		Phone       string      `db:"phone"        json:"phone"`       // 403-555-1234
//		Email       string      `db:"email"        json:"email"`
//		CreatedAt   time.Time   `db:"created_at"   json:"createdAt"`
//		UpdatedAt   time.Time   `db:"updated_at"   json:"updatedAt"`
//	}

package shop

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	Active   Status = "active"
	Inactive Status = "inactive"
)

type Shop struct {
	ID          uuid.UUID `db:"id"           json:"id"`
	Code        string    `db:"code"         json:"code"`
	ShopName    string    `db:"shop_name"    json:"shopName"`
	Status      Status    `db:"status"       json:"status"`
	Address     string    `db:"address"      json:"address"`
	City        string    `db:"city"         json:"city"`
	Province    string    `db:"province"     json:"province"`
	PostalCode  string    `db:"postal_code"  json:"postalCode"`
	ContactName string    `db:"contact_name" json:"contactName"`
	Phone       string    `db:"phone"        json:"phone"`
	Email       string    `db:"email"        json:"email"`
	CreatedAt   time.Time `db:"created_at"   json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updatedAt"`
}
