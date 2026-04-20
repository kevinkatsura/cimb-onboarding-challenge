// Package fraudpb provides hand-written Go types and gRPC client/server
// interfaces for the fraud detection service. These mirror the fraud.proto
// definition and will be replaced by protoc-generated code when protoc
// becomes available in the build toolchain.
package fraudpb

import (
	"encoding/json"
	"strings"
)

// Decision represents the fraud evaluation outcome.
type Decision int32

const (
	Decision_ALLOW     Decision = 0
	Decision_CHALLENGE Decision = 1
	Decision_BLOCK     Decision = 2
	Decision_REVIEW    Decision = 3
)

func (d Decision) String() string {
	switch d {
	case Decision_ALLOW:
		return "ALLOW"
	case Decision_CHALLENGE:
		return "CHALLENGE"
	case Decision_BLOCK:
		return "BLOCK"
	case Decision_REVIEW:
		return "REVIEW"
	default:
		return "UNKNOWN"
	}
}

func (d *Decision) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		// If it's not a string, try unmarshaling as int
		var i int32
		if err := json.Unmarshal(b, &i); err != nil {
			return err
		}
		*d = Decision(i)
		return nil
	}

	switch strings.ToUpper(s) {
	case "ALLOW":
		*d = Decision_ALLOW
	case "CHALLENGE":
		*d = Decision_CHALLENGE
	case "BLOCK":
		*d = Decision_BLOCK
	case "REVIEW":
		*d = Decision_REVIEW
	default:
		*d = Decision_ALLOW // Default to ALLOW for safety if unknown string
	}
	return nil
}

// DeviceFingerprint contains device identification data.
type DeviceFingerprint struct {
	UserAgent        string `json:"user_agent,omitempty"`
	Platform         string `json:"platform,omitempty"`
	ScreenResolution string `json:"screen_resolution,omitempty"`
	Timezone         string `json:"timezone,omitempty"`
}

// FraudEvaluationRequest is the input to EvaluateTransaction.
type FraudEvaluationRequest struct {
	TransactionId        string             `json:"transaction_id,omitempty"`
	PartnerReferenceNo   string             `json:"partner_reference_no,omitempty"`
	ReferenceNo          string             `json:"reference_no,omitempty"`
	SourceAccountNo      string             `json:"source_account_no,omitempty"`
	BeneficiaryAccountNo string             `json:"beneficiary_account_no,omitempty"`
	Amount               int64              `json:"amount,omitempty"`
	Currency             string             `json:"currency,omitempty"`
	SourceIp             string             `json:"source_ip,omitempty"`
	DeviceId             string             `json:"device_id,omitempty"`
	DeviceFingerprint    *DeviceFingerprint `json:"device_fingerprint,omitempty"`
	Channel              string             `json:"channel,omitempty"`
	Latitude             float64            `json:"latitude,omitempty"`
	Longitude            float64            `json:"longitude,omitempty"`
}

// FraudEvaluationResponse is the output of EvaluateTransaction.
type FraudEvaluationResponse struct {
	Decision       Decision `json:"decision,omitempty"`
	RiskScore      float64  `json:"risk_score,omitempty"`
	TriggeredRules []string `json:"triggered_rules,omitempty"`
	EventId        string   `json:"event_id,omitempty"`
	Message        string   `json:"message,omitempty"`
}
