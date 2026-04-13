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

type DeviceInfo struct {
	OS           string `json:"os"`
	OSVersion    string `json:"osVersion"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
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
