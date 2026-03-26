package transaction

type TransferRequest struct {
	ReferenceID string `json:"reference_id"`
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
}
type TransferResponse struct {
	Status string `json:"status"` // success | failed

	TransactionID *string `json:"transaction_id,omitempty"`

	SourceAccount      string `json:"source_account"`
	DestinationAccount string `json:"destination_account"`

	Amount int64 `json:"amount"`

	// Success fields
	SourceBalanceAfter      *int64 `json:"source_balance_after,omitempty"`
	DestinationBalanceAfter *int64 `json:"destination_balance_after,omitempty"`

	// Failure fields
	CurrentBalance *int64 `json:"current_balance,omitempty"`

	Message string `json:"message"`
}
