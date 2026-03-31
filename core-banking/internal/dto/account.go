package dto

type CreateAccountRequest struct {
	CustomerID     string `json:"customer_id"`
	AccountType    string `json:"account_type"`
	Currency       string `json:"currency"`
	OverdraftLimit int64  `json:"overdraft_limit"`
}

type UpdateAccountStatusRequest struct {
	Status string `json:"status"`
}
