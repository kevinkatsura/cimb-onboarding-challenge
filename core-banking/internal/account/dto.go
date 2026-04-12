package account

import "core-banking/internal/snap"

type CreateAccountRequest struct {
	CustomerID     string `json:"customer_id"`
	AccountType    string `json:"account_type"`
	Currency       string `json:"currency"`
	OverdraftLimit int64  `json:"overdraft_limit"`
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
