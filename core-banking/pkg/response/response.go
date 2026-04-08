package response

import (
	"context"
	"core-banking/pkg/apperror"
	"core-banking/pkg/logging"
	"core-banking/pkg/telemetry"
	"encoding/json"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type APIResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Data    interface{}            `json:"data,omitempty"`
	Error   interface{}            `json:"error,omitempty"`
	Meta    map[string]interface{} `json:"meta"`
}

func JSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func Success(w http.ResponseWriter, status int, data interface{}, message string) {
	JSON(w, status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, status int, err interface{}, message string) {
	JSON(w, status, APIResponse{
		Success: false,
		Message: message,
		Error:   err,
	})
}

// httpStatusFromAppError maps domain error types to HTTP status codes.
func httpStatusFromAppError(err error) int {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Type {
		case apperror.NotFound:
			return http.StatusNotFound
		case apperror.BadRequest:
			return http.StatusBadRequest
		case apperror.Conflict:
			return http.StatusConflict
		case apperror.Forbidden:
			return http.StatusForbidden
		case apperror.Unavailable:
			return http.StatusServiceUnavailable
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// RespondError logs the error, records it on the active span with Error status,
// and writes an error JSON response. HTTP status is auto-resolved from AppError type
// if the error is a typed domain error; otherwise falls back to the provided fallbackStatus.
func RespondError(ctx context.Context, w http.ResponseWriter, fallbackStatus int, err error, operation string, fields ...interface{}) {
	// Resolve HTTP status from domain error type
	status := fallbackStatus
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		status = httpStatusFromAppError(err)
	}

	l := logging.Ctx(ctx).With("error", err, "http_status", status)
	if len(fields) > 0 {
		l = l.With(fields...)
	}
	l.Errorw(operation)

	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, operation)
		span.SetAttributes(telemetry.ResponseAttrs(status, 0)...)
		span.SetAttributes(attribute.String("error.operation", operation))
	}

	JSON(w, status, APIResponse{
		Success: false,
		Error:   err.Error(),
	})
}

// RespondOK sets the span status to Ok and writes a success JSON response.
func RespondOK(ctx context.Context, w http.ResponseWriter, status int, data interface{}) {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.SetStatus(codes.Ok, "success")
		span.SetAttributes(telemetry.ResponseAttrs(status, 0)...)
	}

	JSON(w, status, APIResponse{
		Success: true,
		Data:    data,
	})
}

// RespondOKWithMeta sets the span status to Ok and writes a success JSON response with metadata.
func RespondOKWithMeta(ctx context.Context, w http.ResponseWriter, status int, data interface{}, meta map[string]interface{}) {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.SetStatus(codes.Ok, "success")
		span.SetAttributes(telemetry.ResponseAttrs(status, 0)...)
	}

	JSON(w, status, APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}
