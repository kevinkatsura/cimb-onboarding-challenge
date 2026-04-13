package transaction

import (
	"time"

	"core-banking/pkg/pagination"

	"github.com/google/uuid"
)

type Transaction struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	PartnerReferenceNo string     `db:"partner_reference_no" json:"partnerReferenceNo"`
	ReferenceNo        *string    `db:"reference_no" json:"referenceNo,omitempty"`
	TransactionType    string     `db:"transaction_type" json:"transactionType"` // transfer, deposit, withdrawal...
	Status             string     `db:"status" json:"status"`                    // initiated, completed, failed, reversed
	Amount             int64      `db:"amount" json:"amount"`
	Currency           string     `db:"currency" json:"currency"`
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updatedAt"`
	CompletedAt        *time.Time `db:"completed_at" json:"completedAt,omitempty"`
}

type TransactionCompletedEvent struct {
	TransactionID        string `json:"transactionId"`
	PartnerReferenceNo   string `json:"partnerReferenceNo"`
	ReferenceNo          string `json:"referenceNo"`
	SourceAccountNo      string `json:"sourceAccountNo"`
	BeneficiaryAccountNo string `json:"beneficiaryAccountNo"`
	Amount               int64  `json:"amount"`
	Currency             string `json:"currency"`
	Status               string `json:"status"`
	CompletedAt          string `json:"completedAt"`
}

type TransferDetail struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	TransactionID          uuid.UUID  `db:"transaction_id" json:"transactionId"`
	SourceAccountNo        string     `db:"source_account_no" json:"sourceAccountNo"`
	BeneficiaryAccountNo   string     `db:"beneficiary_account_no" json:"beneficiaryAccountNo"`
	BeneficiaryAccountName string     `db:"beneficiary_account_name" json:"beneficiaryAccountName,omitempty"`
	BeneficiaryAddress     string     `db:"beneficiary_address" json:"beneficiaryAddress,omitempty"`
	BeneficiaryBankCode    string     `db:"beneficiary_bank_code" json:"beneficiaryBankCode,omitempty"`
	BeneficiaryBankName    string     `db:"beneficiary_bank_name" json:"beneficiaryBankName,omitempty"`
	BeneficiaryEmail       string     `db:"beneficiary_email" json:"beneficiaryEmail,omitempty"`
	CustomerReference      string     `db:"customer_reference" json:"customerReference,omitempty"`
	FeeType                string     `db:"fee_type" json:"feeType,omitempty"`
	TransactionDate        *time.Time `db:"transaction_date" json:"transactionDate,omitempty"`
	Remark                 string     `db:"remark" json:"remark,omitempty"`
	OriginatorInfos        []byte     `db:"originator_infos" json:"originatorInfos,omitempty"`
	AdditionalInfo         []byte     `db:"additional_info" json:"additionalInfo,omitempty"`
	CreatedAt              time.Time  `db:"created_at" json:"createdAt"`
}

type LedgerEntry struct {
	ID            uuid.UUID `db:"id" json:"id"`
	TransactionID uuid.UUID `db:"transaction_id" json:"transactionId"`
	AccountID     uuid.UUID `db:"account_id" json:"accountId"`
	EntryType     string    `db:"entry_type" json:"entryType"` // debit, credit
	Amount        int64     `db:"amount" json:"amount"`
	Currency      string    `db:"currency" json:"currency"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
}

type AccountBalance struct {
	AccountID        uuid.UUID `db:"account_id" json:"accountId"`
	AvailableBalance int64     `db:"available_balance" json:"availableBalance"`
	PendingBalance   int64     `db:"pending_balance" json:"pendingBalance"`
	LastUpdated      time.Time `db:"last_updated" json:"lastUpdated"`
}

type AccountTransaction struct {
	ID            uuid.UUID `db:"id" json:"id"`
	AccountID     uuid.UUID `db:"account_id" json:"accountId"`
	TransactionID uuid.UUID `db:"transaction_id" json:"transactionId"`
	Direction     string    `db:"direction" json:"direction"` // in, out
	Amount        int64     `db:"amount" json:"amount"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
}

type IdempotencyKey struct {
	ID              uuid.UUID `db:"id" json:"id"`
	Key             string    `db:"key" json:"key"`
	ResponseCode    string    `db:"response_code" json:"responseCode"`
	ResponseMessage string    `db:"response_message" json:"responseMessage"`
	ResponseBody    []byte    `db:"response_body" json:"responseBody"`
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
}

type TransactionListFilter struct {
	AccountID *string
	Status    *string
	Type      *string
	Limit     int
	Cursor    *pagination.Cursor
	Direction string
}

type InsertTransactionParams struct {
	PartnerReferenceNo string
	Amount             int64
	Currency           string

	// SNAP Metadata
	SourceAccountNo        string
	BeneficiaryAccountNo   string
	BeneficiaryAccountName string
	BeneficiaryAddress     string
	BeneficiaryBankCode    string
	BeneficiaryBankName    string
	BeneficiaryEmail       string
	CustomerReference      string
	FeeType                string
	TransactionDate        *time.Time
	Remark                 string
	OriginatorInfos        []byte
	AdditionalInfo         []byte
}

type LedgerEntryParam struct {
	AccountID string
	EntryType string
	Amount    int64
	Currency  string
}

type InsertLedgerParams struct {
	TransactionID uuid.UUID
	Entries       []LedgerEntryParam
}

type AuditLog struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	ActorID    *uuid.UUID `db:"actor_id" json:"actor_id"`
	EntityType string     `db:"entity_type" json:"entity_type"`
	EntityID   *uuid.UUID `db:"entity_id" json:"entity_id"`
	Action     string     `db:"action" json:"action"`
	OldValue   *[]byte    `db:"old_value" json:"old_value,omitempty"`
	NewValue   *[]byte    `db:"new_value" json:"new_value,omitempty"`
	IPAddress  *string    `db:"ip_address" json:"ip_address,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

type FXRate struct {
	ID            uuid.UUID `db:"id" json:"id"`
	BaseCurrency  string    `db:"base_currency" json:"base_currency"`
	QuoteCurrency string    `db:"quote_currency" json:"quote_currency"`
	Rate          float64   `db:"rate" json:"rate"`
	EffectiveAt   time.Time `db:"effective_at" json:"effective_at"`
}

type SenderAccount struct {
	ID         uuid.UUID `db:"id"`
	Balance    int64     `db:"balance"`
	CustomerID string    `db:"customer_id"`
	AccountNo  string    `db:"account_number"`
}
