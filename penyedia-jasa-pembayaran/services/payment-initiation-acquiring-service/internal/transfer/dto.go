package transfer

// TransferRequest is the SNAP transfer-intrabank payload.
type TransferRequest struct {
	PartnerReferenceNo string       `json:"partnerReferenceNo" validate:"required" example:"202311010000000000002"`
	Amount             AmountField  `json:"amount" validate:"required"`
	BeneficiaryAccount AccountField `json:"beneficiaryAccount" validate:"required"`
	SourceAccount      AccountField `json:"sourceAccount" validate:"required"`
	FeeType            string       `json:"feeType" validate:"omitempty,oneof=OUR BEN SHA" example:"OUR"`
	Remark             string       `json:"remark" example:"Lunch at Warung Ikan Bakar"`
	AdditionalInfo     interface{}  `json:"additionalInfo,omitempty"`
}

type AmountField struct {
	Value    string `json:"value" validate:"required" example:"150000.00"`
	Currency string `json:"currency" validate:"required,len=3" example:"IDR"`
}

type AccountField struct {
	AccountNo string `json:"accountNo" validate:"required" example:"8001234567890123"`
}

// FraudContext carries device, IP, and geolocation data for fraud evaluation.
type FraudContext struct {
	DeviceID         string  `json:"deviceId,omitempty"`
	UserAgent        string  `json:"userAgent,omitempty"`
	Platform         string  `json:"platform,omitempty"`
	ScreenResolution string  `json:"screenResolution,omitempty"`
	Timezone         string  `json:"timezone,omitempty"`
	SourceIP         string  `json:"sourceIp,omitempty"`
	Channel          string  `json:"channel,omitempty"`
	Latitude         float64 `json:"latitude,omitempty"`
	Longitude        float64 `json:"longitude,omitempty"`
}

// AdditionalInfo carries optional metadata alongside the transfer request.
type AdditionalInfo struct {
	FraudContext *FraudContext `json:"fraudContext,omitempty"`
}

// TransferResponse is the SNAP-compliant response.
type TransferResponse struct {
	PartnerReferenceNo string `json:"partnerReferenceNo" example:"202311010000000000002"`
	ReferenceNo        string `json:"referenceNo" example:"REF777000001"`
	Amount             string `json:"amount" example:"150000.00"`
	Currency           string `json:"currency" example:"IDR"`
	FeeAmount          string `json:"feeAmount" example:"2500.00"`
	FeeType            string `json:"feeType" example:"OUR"`
	SourceAccount      string `json:"sourceAccount" example:"8001234567890123"`
	BeneficiaryAccount string `json:"beneficiaryAccount" example:"8007776665554443"`
	Status             string `json:"status" example:"SUCCESS"`
	FraudDecision      string `json:"fraudDecision,omitempty"`
}
