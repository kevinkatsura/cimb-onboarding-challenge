package response

import (
	"context"
	"core-banking/pkg/apperror"
	"core-banking/pkg/logging"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ErrorResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
}

// BuildCode safely produces the standard 7-length response code
func BuildCode(httpCode int, serviceCode, caseCode string) string {
	return fmt.Sprintf("%03d%s%s", httpCode, serviceCode, caseCode)
}

// WriteError extracts the AppError and writes a standardized error payload
func WriteError(ctx context.Context, w http.ResponseWriter, serviceCode string, operation string, err error, fields ...interface{}) {
	var appErr *apperror.AppError
	var ok bool
	if appErr, ok = err.(*apperror.AppError); !ok {
		// Fallback for wrapped or uncasted errors
		appErr = apperror.Wrap(apperror.ErrGeneralError, "", err)
	}

	httpCode := appErr.Type.HTTPCode
	caseCode := appErr.Type.CaseCode
	message := appErr.Type.Message

	responseCode := BuildCode(httpCode, serviceCode, caseCode)

	l := logging.Ctx(ctx).With("error", err, "http_status", httpCode, "category", string(appErr.Type.Category))
	if len(fields) > 0 {
		l = l.With(fields...)
	}
	l.Errorw(operation)

	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, message)
		span.SetAttributes(
			attribute.Int("http.status_code", httpCode),
			attribute.String("error.operation", operation),
			attribute.String("error.category", string(appErr.Type.Category)),
			attribute.String("snap.response_code", responseCode),
			attribute.String("snap.response_message", message),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		ResponseCode:    responseCode,
		ResponseMessage: message,
	})
}

// WriteSuccess iterates a flattened marshaler, conditionally embedding meta
func WriteSuccess(ctx context.Context, w http.ResponseWriter, serviceCode string, data interface{}, meta map[string]interface{}) {
	responseCode := BuildCode(http.StatusOK, serviceCode, "00")
	responseMessage := "Successful"

	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.SetStatus(codes.Ok, "success")
		span.SetAttributes(attribute.String("snap.response_code", responseCode))
	}

	payload := make(map[string]interface{})

	if data != nil {
		b, err := json.Marshal(data)
		if err == nil {
			var m map[string]interface{}
			if err := json.Unmarshal(b, &m); err == nil {
				// Safely unmarshaled into a map (flattening)
				for k, v := range m {
					payload[k] = v
				}
			} else {
				// Fallback, place into data node if it's an array or primitive
				payload["data"] = data
			}
		} else {
			payload["data"] = data
		}
	}

	// Ensure fields are forced at root level per SNAP specs
	payload["responseCode"] = responseCode
	payload["responseMessage"] = responseMessage

	if len(meta) > 0 {
		payload["meta"] = meta
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}
