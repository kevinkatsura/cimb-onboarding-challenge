package account

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID                uuid.UUID `db:"id" json:"id"`
	Name              string    `db:"name" json:"name"`
	Email             string    `db:"email" json:"email"`
	PhoneNo           string    `db:"phone_no" json:"phoneNo"`
	CountryCode       string    `db:"country_code" json:"countryCode"`
	DeviceID          string    `db:"device_id" json:"deviceId"`
	DeviceType        string    `db:"device_type" json:"deviceType"`
	DeviceModel       string    `db:"device_model" json:"deviceModel"`
	DeviceOS          string    `db:"device_os" json:"deviceOs"`
	OnboardingPartner string    `db:"onboarding_partner" json:"onboardingPartner"`
	Lang              string    `db:"lang" json:"lang"`
	Locale            string    `db:"locale" json:"locale"`
	CreatedAt         time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time `db:"updated_at" json:"updatedAt"`
}

type Account struct {
	ID            uuid.UUID `db:"id" json:"id"`
	CustomerID    uuid.UUID `db:"customer_id" json:"customerId"`
	AccountNumber string    `db:"account_number" json:"accountNumber"`
	ProductCode   string    `db:"product_code" json:"productCode"`
	Currency      string    `db:"currency" json:"currency"`
	Status        string    `db:"status" json:"status"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time `db:"updated_at" json:"updatedAt"`
}

type AccountBalance struct {
	AccountID uuid.UUID `db:"account_id" json:"accountId"`
	Available int64     `db:"available" json:"available"`
	Pending   int64     `db:"pending" json:"pending"`
	Currency  string    `db:"currency" json:"currency"`
	Version   int       `db:"version" json:"version"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type AccountCreatedEvent struct {
	AccountID     string `json:"accountId"`
	CustomerID    string `json:"customerId"`
	AccountNumber string `json:"accountNumber"`
	ProductCode   string `json:"productCode"`
	Currency      string `json:"currency"`
	CreatedAt     string `json:"createdAt"`
}
