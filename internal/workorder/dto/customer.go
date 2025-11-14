package dto

type Customer struct {
	FullName   string `json:"fullName"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCode string `json:"postalCode"`
	Province   string `json:"province"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}
