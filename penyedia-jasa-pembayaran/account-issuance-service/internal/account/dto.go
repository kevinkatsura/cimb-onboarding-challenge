package account

// RegistrationRequest maps the SNAP registration-account-creation payload.
type RegistrationRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo" validate:"required" example:"202311010000000000001"`
	CustomerID         string `json:"customerId" validate:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name               string `json:"name" validate:"required,min=2,max=255" example:"Kevin Katsura"`
	Email              string `json:"email" validate:"required,email" example:"kevin@gmail.com"`
	PhoneNo            string `json:"phoneNo" validate:"required,min=8,max=20" example:"6281234567890"`
	CountryCode        string `json:"countryCode" validate:"omitempty,len=2" example:"ID"`
	DeviceID           string `json:"deviceId" example:"device-abc-123"`
	DeviceType         string `json:"deviceType" example:"mobile"`
	DeviceModel        string `json:"deviceModel" example:"iPhone 15 Pro"`
	DeviceOS           string `json:"deviceOs" example:"iOS 17.0"`
	OnboardingPartner  string `json:"onboardingPartner" example:"CIMB-NIAGA"`
	Lang               string `json:"lang" example:"id-ID"`
	Locale             string `json:"locale" example:"Jakarta"`
}

// RegistrationResponse is the SNAP-compliant response for account creation.
type RegistrationResponse struct {
	PartnerReferenceNo string `json:"partnerReferenceNo" example:"202311010000000000001"`
	AccountNumber      string `json:"accountNumber" example:"8001234567890123"`
	AccountID          string `json:"accountId" example:"c8f2b3e4-f5a6-4b7c-8d9e-0f1e2d3c4b5a"`
	CustomerID         string `json:"customerId" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductCode        string `json:"productCode" example:"SAVINGS_01"`
	Currency           string `json:"currency" example:"IDR"`
	Status             string `json:"status" example:"SUCCESS"`
}
