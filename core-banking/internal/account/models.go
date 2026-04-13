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

	// SNAP / Compliance Identity
	PartnerReferenceNo string `db:"partner_reference_no" json:"partner_reference_no,omitempty"`
	CountryCode        string `db:"country_code" json:"country_code,omitempty"`
	ExternalCustomerID string `db:"external_customer_id" json:"external_customer_id,omitempty"`

	// Device Context
	DeviceOS           string `db:"device_os" json:"device_os,omitempty"`
	DeviceOSVersion    string `db:"device_os_version" json:"device_os_version,omitempty"`
	DeviceModel        string `db:"device_model" json:"device_model,omitempty"`
	DeviceManufacturer string `db:"device_manufacturer" json:"device_manufacturer,omitempty"`

	// Localization & Onboarding
	Lang              string `db:"lang" json:"lang,omitempty"`
	Locale            string `db:"locale" json:"locale,omitempty"`
	OnboardingPartner string `db:"onboarding_partner" json:"onboarding_partner,omitempty"`
	RedirectURL       string `db:"redirect_url" json:"redirect_url,omitempty"`

	// Auth & Flow
	Scopes       string `db:"scopes" json:"scopes,omitempty"`
	SeamlessData string `db:"seamless_data" json:"seamless_data,omitempty"`
	SeamlessSign string `db:"seamless_sign" json:"seamless_sign,omitempty"`
	State        string `db:"state" json:"state,omitempty"`

	// Merchant Identity
	MerchantID    string `db:"merchant_id" json:"merchant_id,omitempty"`
	SubMerchantID string `db:"sub_merchant_id" json:"sub_merchant_id,omitempty"`
	TerminalType  string `db:"terminal_type" json:"terminal_type,omitempty"`

	// Extensibility
	AdditionalInfo []byte `db:"additional_info" json:"additional_info,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
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

type Product struct {
	Code           string    `db:"code" json:"code"`
	Name           string    `db:"name" json:"name"`
	Currency       string    `db:"currency" json:"currency"`
	MinBalance     int64     `db:"min_balance" json:"min_balance"`
	OverdraftLimit int64     `db:"overdraft_limit" json:"overdraft_limit"`
	DailyLimit     *int64    `db:"daily_limit" json:"daily_limit,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type Account struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	CustomerID    uuid.UUID  `db:"customer_id" json:"customer_id"`
	AccountNumber string     `db:"account_number" json:"account_number"`
	ProductCode   string     `db:"product_code" json:"product_code"`
	Currency      string     `db:"currency" json:"currency"`
	Status        string     `db:"status" json:"status"`
	OpenedAt      time.Time  `db:"opened_at" json:"opened_at"`
	ClosedAt      *time.Time `db:"closed_at" json:"closed_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

type AccountBalance struct {
	AccountID        uuid.UUID `db:"account_id" json:"account_id"`
	AvailableBalance int64     `db:"available_balance" json:"available_balance"`
	PendingBalance   int64     `db:"pending_balance" json:"pending_balance"`
	LastUpdated      time.Time `db:"last_updated" json:"last_updated"`
}

type AccountTransaction struct {
	ID            uuid.UUID `db:"id" json:"id"`
	AccountID     uuid.UUID `db:"account_id" json:"account_id"`
	TransactionID uuid.UUID `db:"transaction_id" json:"transaction_id"`
	Direction     string    `db:"direction" json:"direction"` // in, out
	Amount        int64     `db:"amount" json:"amount"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type ListFilter struct {
	CustomerID  *string
	ProductCode *string
	Status      *string
	Currency    *string
	Limit       int
	Cursor      *pagination.Cursor
	Direction   string
}
