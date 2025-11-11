package dto

type InsuranceInfo struct {
	InsuranceCompany string `json:"insurance_company"`
	AgentFirstName   string `json:"agent_first_name"`
	AgentLastName    string `json:"agent_last_name"`
	AgentPhone       string `json:"agent_phone"`
	PolicyNumber     string `json:"policy_number"`
	ClaimNumber      string `json:"claim_number"`
}
