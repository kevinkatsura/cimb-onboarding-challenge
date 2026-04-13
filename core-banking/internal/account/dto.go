package account

import "core-banking/internal/snap"

type CreateAccountRequest struct {
	PartnerReferenceNo string                   `json:"partnerReferenceNo"`
	CustomerID         string                   `json:"customerId"`
	CountryCode        string                   `json:"countryCode"`
	DeviceInfo         *DeviceInfo              `json:"deviceInfo,omitempty"`
	Name               string                   `json:"name"`
	Email              string                   `json:"email"`
	PhoneNo            string                   `json:"phoneNo"`
	OnboardingPartner  string                   `json:"onboardingPartner"`
	RedirectURL        string                   `json:"redirectUrl"`
	Scopes             string                   `json:"scopes"`
	SeamlessData       string                   `json:"seamlessData"`
	SeamlessSign       string                   `json:"seamlessSign"`
	State              string                   `json:"state"`
	Lang               string                   `json:"lang"`
	Locale             string                   `json:"locale"`
	MerchantID         string                   `json:"merchantId"`
	SubMerchantID      string                   `json:"subMerchantId"`
	TerminalType       string                   `json:"terminalType"`
	AdditionalInfo     *snap.SNAPAdditionalInfo `json:"additionalInfo,omitempty"`

	// Core banking fields
	ProductCode string `json:"product_code"`
	Currency    string `json:"currency"`
}

type RegistrationAccountCreationRequest struct {
	PartnerReferenceNo string                   `json:"partnerReferenceNo" example:"2020102900000000000001"`
	CustomerID         string                   `json:"customerId" example:"CUST001"`
	CountryCode        string                   `json:"countryCode" example:"ID"`
	DeviceInfo         *DeviceInfo              `json:"deviceInfo,omitempty"`
	Name               string                   `json:"name" example:"John Doe"`
	Email              string                   `json:"email" example:"john.doe@example.com"`
	PhoneNo            string                   `json:"phoneNo" example:"628123456789"`
	OnboardingPartner  string                   `json:"onboardingPartner" example:"PARTNER01"`
	RedirectURL        string                   `json:"redirectUrl" example:"https://partner.com/callback"`
	Scopes             string                   `json:"scopes" example:"read write"`
	SeamlessData       string                   `json:"seamlessData" example:"encrypted_data"`
	SeamlessSign       string                   `json:"seamlessSign" example:"signature_hash"`
	State              string                   `json:"state" example:"active"`
	Lang               string                   `json:"lang" example:"id-ID"`
	Locale             string                   `json:"locale" example:"JKT"`
	MerchantID         string                   `json:"merchantId" example:"MER001"`
	SubMerchantID      string                   `json:"subMerchantId" example:"SUB001"`
	TerminalType       string                   `json:"terminalType" example:"MOBILE"`
	AdditionalInfo     *snap.SNAPAdditionalInfo `json:"additionalInfo,omitempty"`
}

type RegistrationAccountCreationResponse struct {
	ReferenceNo        string                   `json:"referenceNo" example:"REF001"`
	PartnerReferenceNo string                   `json:"partnerReferenceNo" example:"2020102900000000000001"`
	AuthCode           string                   `json:"authCode,omitempty" example:"AUTH001"`
	ApiKey             string                   `json:"apiKey,omitempty" example:"CUST001"`
	AccountID          string                   `json:"accountId" example:"ACC001"`
	State              string                   `json:"state,omitempty" example:"active"`
	AdditionalInfo     *snap.SNAPAdditionalInfo `json:"additionalInfo,omitempty"`
}

type DeviceInfo struct {
	OS           string `json:"os" example:"Android"`
	OSVersion    string `json:"osVersion" example:"13.0"`
	Model        string `json:"model" example:"Pixel 7"`
	Manufacturer string `json:"manufacturer" example:"Google"`
}

type UpdateAccountStatusRequest struct {
	Status string `json:"status"`
}

// BalanceInquiryRequest complies with Standar Nasional Open API Pembayaran
type BalanceInquiryRequest struct {
	BeneficiaryAccountNo string                   `json:"beneficiaryAccountNo"`
	PartnerReferenceNo   *string                  `json:"partnerReferenceNo,omitempty"`
	AdditionalInfo       *snap.SNAPAdditionalInfo `json:"additionalInfo,omitempty"`
}

// BalanceInquiryResponse complies with Bank Indonesia specification
type BalanceInquiryResponse struct {
	ResponseCode             string                   `json:"responseCode"`
	ResponseMessage          string                   `json:"responseMessage"`
	ReferenceNo              *string                  `json:"referenceNo,omitempty"`
	PartnerReferenceNo       *string                  `json:"partnerReferenceNo,omitempty"`
	BeneficiaryAccountName   string                   `json:"beneficiaryAccountName"`
	BeneficiaryAccountNo     string                   `json:"beneficiaryAccountNo"`
	BeneficiaryAccountStatus *string                  `json:"beneficiaryAccountStatus,omitempty"`
	BeneficiaryAccountType   *string                  `json:"beneficiaryAccountType,omitempty"`
	Currency                 *string                  `json:"currency,omitempty"`
	Name                     string                   `json:"name,omitempty"`
	AdditionalInfo           *snap.SNAPAdditionalInfo `json:"additionalInfo,omitempty"`
}
