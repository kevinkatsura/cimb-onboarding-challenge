package response

import (
	"account-issuance-service/pkg/apperror"
	"account-issuance-service/pkg/logging"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func BuildCode(httpCode int, serviceCode, caseCode string) string {
	return fmt.Sprintf("%03d%s%s", httpCode, serviceCode, caseCode)
}

func WriteError(ctx context.Context, w http.ResponseWriter, serviceCode string, err error) {
	var appErr *apperror.AppError
	if ae, ok := err.(*apperror.AppError); ok {
		appErr = ae
	} else {
		appErr = apperror.Wrap(apperror.ErrGeneralError, "", err)
	}

	httpCode := appErr.Type.HTTPCode
	responseCode := BuildCode(httpCode, serviceCode, appErr.Type.CaseCode)
	message := appErr.Type.Message
	if appErr.Message != "" {
		message = appErr.Message
	}

	logging.Ctx(ctx).Errorw("request error", "response_code", responseCode, "message", message, "error", err)

	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, message)
		span.SetAttributes(attribute.String("snap.response_code", responseCode))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"responseCode":    responseCode,
		"responseMessage": message,
	})
}

func WriteSuccess(ctx context.Context, w http.ResponseWriter, serviceCode string, data interface{}) {
	responseCode := BuildCode(http.StatusOK, serviceCode, "00")

	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.SetStatus(codes.Ok, "success")
		span.SetAttributes(attribute.String("snap.response_code", responseCode))
	}

	payload := make(map[string]interface{})
	if data != nil {
		b, _ := json.Marshal(data)
		var m map[string]interface{}
		if json.Unmarshal(b, &m) == nil {
			for k, v := range m {
				payload[k] = v
			}
		} else {
			payload["data"] = data
		}
	}
	payload["responseCode"] = responseCode
	payload["responseMessage"] = "Successful"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}
