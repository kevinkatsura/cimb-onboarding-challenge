package domain

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	FullName      string     `db:"full_name" json:"full_name"`
	DateOfBirth   time.Time  `db:"data_of_birth" json:"date_of_birth"`
	Nationality   string     `db:"nationality" json:"nationality"`
	Email         string     `db:"email" json:"email"`
	PhoneNumber   string     `db:"phone_number" json:"phone_number"`
	KYCStatus     string     `db:"kyc_status" json:"kyc_status"`
	KYCVerifiedAt *time.Time `db:"kyc_verified_at" json:"kyc_verified_at,omitempty"`
	RiskLevel     string     `db:"risk_level" json:"risk_level"`
	PEPFlag       bool       `db:"pep_flag" json:"pep_flag"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

type CustomerDocument struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	CustomerID     uuid.UUID  `db:"customer_id" json:"customer_id"`
	DocumentType   string     `db:"document_type" json:"document_type"`
	DocumentNumber string     `db:"document_number" json:"document_number"`
	IssuingCountry string     `db:"issuing_country" json:"issuing_country"`
	ExpiresAt      *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}

type Account struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	CustomerID       uuid.UUID  `db:"customer_id" json:"customer_id"`
	AccountNumber    string     `db:"account_number" json:"account_number"`
	AccountType      string     `db:"account_type" json:"account_type"`
	Currency         string     `db:"currency" json:"currency"`
	Status           string     `db:"status" json:"status"`
	AvailableBalance int64      `db:"available_balance" json:"available_balance"`
	PendingBalance   int64      `db:"pending_balance" json:"pending_balance"`
	OverdraftLimit   int64      `db:"overdraft_limit" json:"overdraft_limit"`
	OpenedAt         time.Time  `db:"opened_at" json:"opened_at"`
	ClosedAt         *time.Time `db:"closed_at" json:"closed_at,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt        *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

type Transaction struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	ReferenceID       string     `db:"reference_id" json:"reference_id"`
	ExternalReference *string    `db:"external_reference" json:"external_reference,omitempty"`
	TransactionType   string     `db:"transaction_type" json:"transaction_type"`
	Status            string     `db:"status" json:"status"`
	Amount            int64      `db:"amount" json:"amount"`
	Currency          string     `db:"currency" json:"currency"`
	InitiatedBy       *uuid.UUID `db:"initiated_by" json:"initiated_by,omitempty"`
	Description       *string    `db:"description" json:"description,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	CompletedAt       *time.Time `db:"completed_at" json:"completed_at,omitempty"`
}

type JournalEntry struct {
	ID            uuid.UUID `db:"id"`
	TransactionID uuid.UUID `db:"transaction_id"`
	JournalType   string    `db:"journal_type"`
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

type AuditLog struct {
	ID         uuid.UUID  `db:"id"`
	ActorID    *uuid.UUID `db:"actor_id"`
	EntityType string     `db:"entity_type"`
	EntityID   *uuid.UUID `db:"entity_id"`
	Action     string     `db:"action"`
	OldValue   *[]byte    `db:"old_value"`
	NewValue   *[]byte    `db:"new_value"`
	IPAddress  *string    `db:"ip_address"`
	CreatedAt  time.Time  `db:"created_at"`
}

type IdempotencyKey struct {
	ID          uuid.UUID `db:"id"`
	Key         string    `db:"key"`
	RequestHash string    `db:"request_hash"`
	Response    *[]byte   `db:"response"`
	CreatedAt   time.Time `db:"created_at"`
}

type FXRate struct {
	ID            uuid.UUID `db:"id"`
	BaseCurrency  string    `db:"base_currency"`
	QuoteCurrency string    `db:"quote_currency"`
	Rate          float64   `db:"rate"`
	EffectiveAt   time.Time `db:"effective_at"`
}
