package dto

import (
	"errors"
	"strings"
)

type IntakePayload struct {
	Customer  CustomerIntake   `json:"customer"`
	Vehicle   VehicleIntake    `json:"vehicle"`
	Insurance *InsuranceIntake `json:"insurance,omitempty"`
	ShopCode  string           `json:"shopCode"`
}
type IntakeEditPayload struct {
	Customer  *CustomerIntake  `json:"customer,omitempty"`
	Vehicle   *VehicleIntake   `json:"vehicle,omitempty"`
	Insurance *InsuranceIntake `json:"insurance,omitempty"`
}

type CustomerIntake struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCode string `json:"postalCode"`
	Province   string `json:"province"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}
type InsuranceIntake struct {
	InsuranceCompany string `json:"insuranceCompany"`
	AgentFirstName   string `json:"agentFirstName"`
	AgentLastName    string `json:"agentLastName"`
	AgentPhone       string `json:"agentPhone"`
	PolicyNumber     string `json:"policyNumber"`
	ClaimNumber      string `json:"claimNumber"`
}
type VehicleIntake struct {
	PlateNo   string `json:"plateNo"`
	Make      string `json:"make"`
	Model     string `json:"model"`
	BodyStyle string `json:"bodyStyle"`
	ModelYear int    `json:"modelYear"`
	VIN       string `json:"vin"`
	Color     string `json:"color"`
}

func (i *InsuranceIntake) IsEmpty() bool {
	if i == nil {
		return true
	}

	return strings.TrimSpace(i.InsuranceCompany) == "" &&
		strings.TrimSpace(i.AgentFirstName) == "" &&
		strings.TrimSpace(i.AgentLastName) == "" &&
		strings.TrimSpace(i.AgentPhone) == "" &&
		strings.TrimSpace(i.PolicyNumber) == "" &&
		strings.TrimSpace(i.ClaimNumber) == ""
}

func (i *InsuranceIntake) Validate() error {
	if strings.TrimSpace(i.InsuranceCompany) == "" {
		return errors.New("insurance company is required when insurance information is provided")
	}
	return nil
}
