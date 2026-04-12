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
	TransactionType    string     `db:"transaction_type" json:"transactionType"`
	Status             string     `db:"status" json:"status"`
	Amount             int64      `db:"amount" json:"amount"`
	Currency           string     `db:"currency" json:"currency"`
	Description        *string    `db:"description" json:"description,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updatedAt"`
	CompletedAt        *time.Time `db:"completed_at" json:"completedAt,omitempty"`
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
	OriginatorInfos        *[]byte    `db:"originator_infos" json:"originatorInfos,omitempty"`
	AdditionalInfo         *[]byte    `db:"additional_info" json:"additionalInfo,omitempty"`
	CreatedAt              time.Time  `db:"created_at" json:"createdAt"`
}

type Journal struct {
	ID            uuid.UUID `db:"id"`
	TransactionID uuid.UUID `db:"transaction_id"`
	JournalType   string    `db:"journal_type"`
	Status        string    `db:"status"`
	PostedAt      time.Time `db:"posted_at"`
	CreatedAt     time.Time `db:"created_at"`
}

type LedgerEntry struct {
	ID           uuid.UUID `db:"id"`
	JournalID    uuid.UUID `db:"journal_id"`
	AccountID    uuid.UUID `db:"account_id"`
	EntryType    string    `db:"entry_type"`
	Amount       int64     `db:"amount"`
	Currency     string    `db:"currency"`
	BalanceAfter *int64    `db:"balance_after"`
	CreatedAt    time.Time `db:"created_at"`
}

type Payment struct {
	ID            uuid.UUID `db:"id"`
	TransactionID uuid.UUID `db:"transaction_id"`
	PaymentMethod string    `db:"payment_method"`
	Provider      string    `db:"provider"`
	Status        string    `db:"status"`
	FeeAmount     int64     `db:"fee_amount"`
	Metadata      *[]byte   `db:"metadata"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type IdempotencyKey struct {
	ID              uuid.UUID `db:"id"`
	Key             string    `db:"key"`
	RequestHash     string    `db:"request_hash"`
	ResponseCode    string    `db:"response_code"`
	ResponseMessage string    `db:"response_message"`
	ResponseBody    []byte    `db:"response_body"`
	CreatedAt       time.Time `db:"created_at"`
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
	JournalID uuid.UUID
	Entries   []LedgerEntryParam
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
	Balance    int64  `db:"balance"`
	CustomerID string `db:"customer_id"`
	AccountNo  string `db:"account_number"`
}
