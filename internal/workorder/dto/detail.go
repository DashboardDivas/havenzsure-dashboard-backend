package dto

import (
	"time"
)

type WorkOrderDetail struct {
	Code         string           `json:"code"`
	Status       WorkOrderStatus  `json:"status"`
	DateReceived time.Time        `json:"date_received"`
	DateUpdated  time.Time        `json:"date_updated"`
	Customer     CustomerDetail   `json:"customer"`
	Vehicle      VehicleDetail    `json:"vehicle"`
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
	InsuranceCompany string `json:"insurance_company"`
	AgentFullName    string `json:"agent_full_name"`
	AgentPhone       string `json:"agent_phone"`
	PolicyNumber     string `json:"policy_number"`
	ClaimNumber      string `json:"claim_number"`
}
