package dto

import (
	"time"

	"github.com/google/uuid"
)

type WorkOrderDetail struct {
	ID           uuid.UUID        `json:"id"`
	Code         string           `json:"code"`
	Status       WorkOrderStatus  `json:"status"`
	DateReceived time.Time        `json:"dateReceived"`
	DateUpdated  time.Time        `json:"dateUpdated"`
	Customer     CustomerDetail   `json:"customer"`
	Vehicle      VehicleDetail    `json:"vehicle"`
	Shop         ShopSummary      `json:"shop"`
	Insurance    *InsuranceDetail `json:"insurance,omitempty"`
}

type CustomerDetail struct {
	FullName   string `json:"fullName"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCode string `json:"postalCode"`
	Province   string `json:"province"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}

type VehicleDetail struct {
	PlateNo   string `json:"plateNo"`
	Make      string `json:"make"`
	Model     string `json:"model"`
	BodyStyle string `json:"bodyStyle"`
	ModelYear int    `json:"modelYear"`
	VIN       string `json:"vin"`
	Color     string `json:"color"`
}

type InsuranceDetail struct {
	InsuranceCompany string `json:"insuranceCompany"`
	AgentFullName    string `json:"agentFullName"`
	AgentPhone       string `json:"agentPhone"`
	PolicyNumber     string `json:"policyNumber"`
	ClaimNumber      string `json:"claimNumber"`
}
