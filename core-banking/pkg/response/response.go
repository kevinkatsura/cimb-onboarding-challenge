package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response is the standard API response structure
type Response map[string]interface{}

type ResponseCodeBuilder struct {
	ServiceCode string
}

// BuildCode builds the response code: http_status_code + service_code + case_code
func (b ResponseCodeBuilder) BuildCode(httpStatus int, caseCode string) string {
	return fmt.Sprintf("%d%s%s", httpStatus, b.ServiceCode, caseCode)
}

// Success builds a success response with flattened data
func (b ResponseCodeBuilder) Success(httpStatus int, caseCode, message string, data map[string]interface{}) Response {
	resp := Response{
		"responseCode":    b.BuildCode(httpStatus, caseCode),
		"responseMessage": message,
	}
	for k, v := range data {
		resp[k] = v
	}
	return resp
}

// Error builds an error response
func (b ResponseCodeBuilder) Error(httpStatus int, caseCode, message string) Response {
	return Response{
		"responseCode":    b.BuildCode(httpStatus, caseCode),
		"responseMessage": message,
	}
}

// Predefined builders for each service
var (
	TransferResponse = ResponseCodeBuilder{ServiceCode: "17"}
	HistoryResponse  = ResponseCodeBuilder{ServiceCode: "12"}
	AccountResponse  = ResponseCodeBuilder{ServiceCode: "16"}
)

// WriteJSON writes the response as JSON
func WriteJSON(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
