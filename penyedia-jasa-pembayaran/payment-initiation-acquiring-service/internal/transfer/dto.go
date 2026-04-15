package transfer

// TransferRequest is the SNAP transfer-intrabank payload.
type TransferRequest struct {
	PartnerReferenceNo string       `json:"partnerReferenceNo" validate:"required"`
	Amount             AmountField  `json:"amount" validate:"required"`
	BeneficiaryAccount AccountField `json:"beneficiaryAccount" validate:"required"`
	SourceAccount      AccountField `json:"sourceAccount" validate:"required"`
	FeeType            string       `json:"feeType" validate:"omitempty,oneof=OUR BEN SHA"`
	Remark             string       `json:"remark"`
	AdditionalInfo     interface{}  `json:"additionalInfo,omitempty"`
}

type AmountField struct {
	Value    string `json:"value" validate:"required"`
	Currency string `json:"currency" validate:"required,len=3"`
}

type AccountField struct {
	AccountNo string `json:"accountNo" validate:"required"`
}

// TransferResponse is the SNAP-compliant response.
type TransferResponse struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ReferenceNo        string `json:"referenceNo"`
	Amount             string `json:"amount"`
	Currency           string `json:"currency"`
	FeeAmount          string `json:"feeAmount"`
	FeeType            string `json:"feeType"`
	SourceAccount      string `json:"sourceAccount"`
	BeneficiaryAccount string `json:"beneficiaryAccount"`
	Status             string `json:"status"`
}
