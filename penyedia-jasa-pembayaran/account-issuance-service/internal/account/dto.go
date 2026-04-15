package account

// RegistrationRequest maps the SNAP registration-account-creation payload.
type RegistrationRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo" validate:"required"`
	CustomerID         string `json:"customerId" validate:"omitempty,uuid4"`
	Name               string `json:"name" validate:"required,min=2,max=255"`
	Email              string `json:"email" validate:"required,email"`
	PhoneNo            string `json:"phoneNo" validate:"required,min=8,max=20"`
	CountryCode        string `json:"countryCode" validate:"omitempty,len=2"`
	DeviceID           string `json:"deviceId"`
	DeviceType         string `json:"deviceType"`
	DeviceModel        string `json:"deviceModel"`
	DeviceOS           string `json:"deviceOs"`
	OnboardingPartner  string `json:"onboardingPartner"`
	Lang               string `json:"lang"`
	Locale             string `json:"locale"`
}

// RegistrationResponse is the SNAP-compliant response for account creation.
type RegistrationResponse struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	AccountNumber      string `json:"accountNumber"`
	AccountID          string `json:"accountId"`
	CustomerID         string `json:"customerId"`
	ProductCode        string `json:"productCode"`
	Currency           string `json:"currency"`
	Status             string `json:"status"`
}
