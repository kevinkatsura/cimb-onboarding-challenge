package snap

// SNAPAmount represents the complex amount object dictated by BI SNAP
type SNAPAmount struct {
	Value    string `json:"value"` // e.g., "10000.00"
	Currency string `json:"currency"`
}

// SNAPResponse is a generic structure wrapper commonly expected in SNAP error structures
// where fields are output at root flat
type SNAPResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
}

// SNAPAdditionalInfo holds bank-specific extra details requested optionally
type SNAPAdditionalInfo struct {
	DeviceId string `json:"deviceId,omitempty"`
	Channel  string `json:"channel,omitempty"`
}
