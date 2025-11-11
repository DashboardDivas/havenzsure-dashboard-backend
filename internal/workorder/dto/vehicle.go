package dto

// &detail.Vehicle.PlateNo,
// &detail.Vehicle.Make,
// &detail.Vehicle.Model,
// &detail.Vehicle.BodyStyle,
// &detail.Vehicle.ModelYear,
// &detail.Vehicle.VIN,
// &detail.Vehicle.Color,

type Vehicle struct {
	PlateNo   string `json:"plateNo"`
	Make      string `json:"make"`
	Model     string `json:"model"`
	BodyStyle string `json:"bodyStyle"`
	ModelYear int    `json:"modelYear"`
	VIN       string `json:"vin"`
	Color     string `json:"color"`
}
