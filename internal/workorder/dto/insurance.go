package dto

type Insurance struct {
	InsuranceCompany string `json:"insurance_company"`
	AgentFullName    string `json:"agent_full_name"`
	AgentPhone       string `json:"agent_phone"`
	PolicyNumber     string `json:"policy_number"`
	ClaimNumber      string `json:"claim_number"`
}
