package transaction

type TransferRequest struct {
	ReferenceID string `json:"reference_id"`
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
}
