package transaction

import (
	"core-banking/internal/snap"

	"core-banking/pkg/apperror"
	"core-banking/pkg/logging"
	"core-banking/pkg/pagination"
	"core-banking/pkg/response"
	"core-banking/pkg/telemetry"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	service Interface
}

func NewHandler(service Interface) *Handler {
	return &Handler{service: service}
}

// Swagger helper structures
type ErrorResponse response.ErrorResponse

type ResponseSuccessTransactionList struct {
	ResponseCode    string                       `json:"responseCode"`
	ResponseMessage string                       `json:"responseMessage"`
	Data            []TransactionHistoryResponse `json:"data"`
	Meta            map[string]interface{}       `json:"meta"`
}

// Transfer performs a standard multi-leg P2P transfer between two accounts.
// @Summary Account Transfer (Optimistic)
// @Description Safely processes an inter-account fund transfer linearly.
// @Tags transactions
// @Accept json
// @Produce json
// @Param X-PARTNER-ID header string true "Partner ID provided by the bank" default(PARTNER001)
// @Param X-TIMESTAMP header string true "ISO-8601 Timestamp" default(2026-04-12T18:00:00Z)
// @Param X-SIGNATURE header string true "HMAC-SHA256 Signature" default(valid-signature-for-testing)
// @Param X-EXTERNAL-ID header string true "Random unique ID representing the request" default(1234567890)
// @Param request body IntrabankTransferRequest true "Transfer Payload"
// @Success 200 {object} IntrabankTransferResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1.0/transfer-intrabank [post]
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req IntrabankTransferRequest
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "POST /v1/transfer")...)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(ctx, w, snap.TransferIntrabankServiceCode, "transfer_invalid_request", apperror.Wrap(apperror.ErrInvalidFieldFormat, "Invalid Field Format", err))
		return
	}

	logging.Ctx(ctx).Infow("transfer_handler_called",
		"partner_reference_no", req.PartnerReferenceNo,
		"source_account", req.SourceAccountNo,
		"beneficiary_account", req.BeneficiaryAccountNo,
	)

	data, err := h.service.Transfer(ctx, req)
	if err != nil {
		response.WriteError(ctx, w, snap.TransferIntrabankServiceCode, "transfer_failed", err, "reference_id", req.PartnerReferenceNo)
		return
	}

	response.WriteSuccess(ctx, w, snap.TransferIntrabankServiceCode, data, nil)
}

// TransferWithLock performs an aggressive distributed-lock based transfer.
// @Summary Account Transfer (Pessimistic Redis Lock)
// @Description Safely processes a transfer explicitly restricting concurrency on the accounts.
// @Tags transactions
// @Accept json
// @Produce json
// @Param X-PARTNER-ID header string true "Partner ID provided by the bank" default(PARTNER001)
// @Param X-TIMESTAMP header string true "ISO-8601 Timestamp" default(2026-04-12T18:00:00Z)
// @Param X-SIGNATURE header string true "HMAC-SHA256 Signature" default(valid-signature-for-testing)
// @Param X-EXTERNAL-ID header string true "Random unique ID representing the request" default(1234567890)
// @Param request body IntrabankTransferRequest true "Transfer Payload"
// @Success 200 {object} IntrabankTransferResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1.0/transfer-intrabank-locked [post]
func (h *Handler) TransferWithLock(w http.ResponseWriter, r *http.Request) {
	var req IntrabankTransferRequest
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "POST /v1.0/transfer-intra-bank-locked")...)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(ctx, w, snap.TransferIntrabankServiceCode, "transfer_locked_invalid_request", apperror.Wrap(apperror.ErrInvalidFieldFormat, "Invalid Field Format", err))
		return
	}

	logging.Ctx(ctx).Infow("transfer_with_lock_handler_called",
		"partner_reference_no", req.PartnerReferenceNo,
		"source_account", req.SourceAccountNo,
		"beneficiary_account", req.BeneficiaryAccountNo,
	)

	data, err := h.service.TransferWithLock(ctx, req)
	if err != nil {
		response.WriteError(ctx, w, snap.TransferIntrabankServiceCode, "transfer_locked_failed", err, "reference_id", req.PartnerReferenceNo)
		return
	}

	response.WriteSuccess(ctx, w, snap.TransferIntrabankServiceCode, data, nil)
}

// TransferStatusInquiry performs an inquiry on a previous transfer using SNAP.
// @Summary Transfer Status Inquiry
// @Description Fetch transaction status conforming to SNAP BI using Service Code 36.
// @Tags transactions
// @Accept json
// @Produce json
// @Param X-PARTNER-ID header string true "Partner ID provided by the bank" default(PARTNER001)
// @Param X-TIMESTAMP header string true "ISO-8601 Timestamp" default(2026-04-12T18:00:00Z)
// @Param X-SIGNATURE header string true "HMAC-SHA256 Signature" default(valid-signature-for-testing)
// @Param X-EXTERNAL-ID header string true "Random unique ID representing the request" default(1234567890)
// @Param request body TransferStatusInquiryRequest true "Inquiry Payload"
// @Success 200 {object} TransferStatusInquiryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1.0/transfer/status [post]
func (h *Handler) TransferStatusInquiry(w http.ResponseWriter, r *http.Request) {
	var req TransferStatusInquiryRequest
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "POST /v1.0/transfer-status-inquiry")...)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(ctx, w, snap.TransferStatusInquiryServiceCode, "inquiry_invalid_request", apperror.Wrap(apperror.ErrInvalidFieldFormat, "Invalid Field Format", err))
		return
	}

	data, err := h.service.TransferStatusInquiry(ctx, req)
	if err != nil {
		response.WriteError(ctx, w, snap.TransferStatusInquiryServiceCode, "inquiry_failed", err, "reference_id", req.OriginalPartnerReferenceNo)
		return
	}

	response.WriteSuccess(ctx, w, snap.TransferStatusInquiryServiceCode, data, nil)
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request, accountID *string, route string) {
	q := r.URL.Query()
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, route)...)

	cursor, err := pagination.DecodeCursor(q.Get("cursor"))
	if err != nil {
		response.WriteError(ctx, w, snap.TransactionHistoryListServiceCode, "transaction_list_invalid_cursor", apperror.Wrap(apperror.ErrBadRequest, "Bad Request", err), "account_id", accountID)
		return
	}

	limit := 20
	fmt.Sscanf(q.Get("limit"), "%d", &limit)

	filter := TransactionListFilter{
		AccountID: accountID,
		Limit:     limit,
		Cursor:    cursor,
		Direction: q.Get("direction"),
	}

	if v := q.Get("type"); v != "" {
		filter.Type = &v
	}
	if v := q.Get("status"); v != "" {
		filter.Status = &v
	}

	data, total, nextCursor, prevCursor, err := h.service.List(ctx, filter)
	if err != nil {
		response.WriteError(ctx, w, snap.TransactionHistoryListServiceCode, "transaction_list_handler_error", err, "account_id", accountID, "limit", limit)
		return
	}

	response.WriteSuccess(ctx, w, snap.TransactionHistoryListServiceCode, data, map[string]interface{}{
		"limit":       limit,
		"next_cursor": nextCursor,
		"prev_cursor": prevCursor,
		"total":       total,
	})
}

// ListByAccount retrieves transaction ledger logs isolated natively to a single participant identity.
// @Summary List Transactions (By Account)
// @Description Iterates paginated lists isolated for an account.
// @Tags transactions
// @Produce json
// @Param id path string true "Account ID"
// @Param limit query int false "Limits response objects"
// @Param cursor query string false "Pagination Base64 Cursor"
// @Param direction query string false "Pagination Direction (next, prev)"
// @Success 200 {object} ResponseSuccessTransactionList
// @Failure 400 {object} ErrorResponse
// @Router /accounts/{id}/transactions [get]
func (h *Handler) ListByAccount(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	h.handleList(w, r, &accountID, "GET /accounts/{id}/transactions")
}

// ListAll retrieves all transaction ledger logs globally.
// @Summary List All Transactions
// @Description Aggregates the global banking journal history explicitly sorted.
// @Tags transactions
// @Produce json
// @Param limit query int false "Limits response objects"
// @Param cursor query string false "Pagination Base64 Cursor"
// @Success 200 {object} ResponseSuccessTransactionList
// @Failure 400 {object} ErrorResponse
// @Router /transactions [get]
func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	h.handleList(w, r, nil, "GET /transactions")
}
