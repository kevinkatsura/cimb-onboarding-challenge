package account

import (
	"time"

	"core-banking/pkg/pagination"

	"github.com/google/uuid"
)

type Customer struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	FullName      string     `db:"full_name" json:"full_name"`
	DateOfBirth   time.Time  `db:"date_of_birth" json:"date_of_birth"`
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

type ListFilter struct {
	CustomerID  *string
	AccountType *string
	Status      *string
	Currency    *string
	Limit       int
	Cursor      *pagination.Cursor
	Direction   string
}
