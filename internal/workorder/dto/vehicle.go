package dto

type Vehicle struct {
	PlateNo   string `json:"plateNo"`
	Make      string `json:"make"`
	Model     string `json:"model"`
	BodyStyle string `json:"bodyStyle"`
	ModelYear int    `json:"modelYear"`
	VIN       string `json:"vin"`
	Color     string `json:"color"`
}
