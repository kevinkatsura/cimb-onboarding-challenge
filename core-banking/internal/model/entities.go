package model

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID            uuid.UUID  `db:"id"`
	FullName      string     `db:"full_name"`
	DateOfBirth   time.Time  `db:"data_of_birth"`
	Nationality   string     `db:"nationality"`
	Email         string     `db:"email"`
	PhoneNumber   string     `db:"phone_number"`
	KYCStatus     string     `db:"kyc_status"`
	KYCVerifiedAt *time.Time `db:"kyc_verified_at"`
	RiskLevel     string     `db:"risk_level"`
	PEPFlag       bool       `db:"pep_flag"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

type CustomerDocument struct {
	ID             uuid.UUID  `db:"id"`
	CustomerID     uuid.UUID  `db:"customer_id"`
	DocumentType   string     `db:"document_type"`
	DocumentNumber string     `db:"document_number"`
	IssuingCountry string     `db:"issuing_country"`
	ExpiresAt      *time.Time `db:"expires_at"`
	CreatedAt      time.Time  `db:"created_at"`
}

type Account struct {
	ID               uuid.UUID  `db:"id"`
	CustomerID       uuid.UUID  `db:"customer_id"`
	AccountNumber    string     `db:"account_number"`
	AccountType      string     `db:"account_type"`
	Currency         string     `db:"currency"`
	Status           string     `db:"status"`
	AvailableBalance int64      `db:"available_balance"`
	PendingBalance   int64      `db:"pending_balance"`
	OverdraftLimit   int64      `db:"overdraft_limit"`
	OpenedAt         time.Time  `db:"opened_at"`
	ClosedAt         *time.Time `db:"closed_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	DeletedAt        *time.Time `db:"deleted_at"`
}

type Transaction struct {
	ID                uuid.UUID  `db:"id"`
	ReferenceID       string     `db:"reference_id"`
	ExternalReference *string    `db:"external_reference"`
	TransactionType   string     `db:"transaction_type"`
	Status            string     `db:"status"`
	Amount            int64      `db:"amount"`
	Currency          string     `db:"currency"`
	InitiatedBy       *uuid.UUID `db:"initiated_by"`
	Description       *string    `db:"description"`
	CreatedAt         time.Time  `db:"created_at"`
	CompletedAt       *time.Time `db:"completed_at"`
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
