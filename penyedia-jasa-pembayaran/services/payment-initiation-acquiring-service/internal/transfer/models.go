package transfer

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID                 uuid.UUID `db:"id" json:"id"`
	PartnerReferenceNo string    `db:"partner_reference_no" json:"partnerReferenceNo"`
	ReferenceNo        string    `db:"reference_no" json:"referenceNo"`
	Type               string    `db:"type" json:"type"`
	Status             string    `db:"status" json:"status"`
	Amount             int64     `db:"amount" json:"amount"`
	Currency           string    `db:"currency" json:"currency"`
	FeeAmount          int64     `db:"fee_amount" json:"feeAmount"`
	FeeType            string    `db:"fee_type" json:"feeType"`
	Remark             string    `db:"remark" json:"remark"`
	FraudDecision      string    `db:"fraud_decision" json:"fraudDecision"`
	FraudEventID       string    `db:"fraud_event_id" json:"fraudEventId"`
	CreatedAt          time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt          time.Time `db:"updated_at" json:"updatedAt"`
}

type TransferDetail struct {
	ID                     uuid.UUID `db:"id" json:"id"`
	TransactionID          uuid.UUID `db:"transaction_id" json:"transactionId"`
	SourceAccountNo        string    `db:"source_account_no" json:"sourceAccountNo"`
	SourceAccountName      string    `db:"source_account_name" json:"sourceAccountName"`
	BeneficiaryAccountNo   string    `db:"beneficiary_account_no" json:"beneficiaryAccountNo"`
	BeneficiaryAccountName string    `db:"beneficiary_account_name" json:"beneficiaryAccountName"`
	BeneficiaryEmail       string    `db:"beneficiary_email" json:"beneficiaryEmail"`
	CreatedAt              time.Time `db:"created_at" json:"createdAt"`
}

type TransferCompletedEvent struct {
	TransactionID      string `json:"transactionId"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ReferenceNo        string `json:"referenceNo"`
	Amount             int64  `json:"amount"`
	Currency           string `json:"currency"`
	FeeAmount          int64  `json:"feeAmount"`
	SourceAccount      string `json:"sourceAccount"`
	BeneficiaryAccount string `json:"beneficiaryAccount"`
	Status             string `json:"status"`
	CompletedAt        string `json:"completedAt"`
}
