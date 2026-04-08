package transaction

import (
	"core-banking/internal/dto"
	"core-banking/internal/service/transaction"
	"core-banking/pkg/logging"
	"core-banking/pkg/pagination"
	"core-banking/pkg/response"
	"core-banking/pkg/telemetry"
	"encoding/json"
	"fmt"
	"net/http"

	"core-banking/internal/domain"

	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	service transaction.Interface
}

func NewHandler(service transaction.Interface) *Handler {
	return &Handler{service: service}
}

// Transfer performs a standard multi-leg P2P transfer between two accounts.
// @Summary Account Transfer (Optimistic)
// @Description Safely processes an inter-account fund transfer linearly.
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body dto.TransferRequest true "Transfer Payload"
// @Success 200 {object} response.APIResponse{data=dto.TransferResponse}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /v1/transfer [post]
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req dto.TransferRequest
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "POST /v1/transfer")...)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(ctx, w, http.StatusBadRequest, err, "transfer_request_invalid_body")
		return
	}

	logging.Ctx(ctx).Infow("transfer_handler_called",
		"reference_id", req.ReferenceID,
		"from_account", req.FromAccount,
		"to_account", req.ToAccount,
	)

	data, err := h.service.Transfer(ctx, req)
	if err != nil {
		response.RespondError(ctx, w, http.StatusInternalServerError, err, "transfer_handler_error",
			"reference_id", req.ReferenceID,
		)
		return
	}

	response.RespondOK(ctx, w, http.StatusOK, data)
}

// TransferWithLock performs an aggressive distributed-lock based transfer.
// @Summary Account Transfer (Pessimistic Redis Lock)
// @Description Safely processes a transfer explicitly restricting concurrency on the accounts.
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body dto.TransferRequest true "Transfer Payload"
// @Success 200 {object} response.APIResponse{data=dto.TransferResponse}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /v2/transfer [post]
func (h *Handler) TransferWithLock(w http.ResponseWriter, r *http.Request) {
	var req dto.TransferRequest
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "POST /v2/transfer")...)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(ctx, w, http.StatusBadRequest, err, "transfer_with_lock_request_invalid_body")
		return
	}

	logging.Ctx(ctx).Infow("transfer_with_lock_handler_called",
		"reference_id", req.ReferenceID,
		"from_account", req.FromAccount,
		"to_account", req.ToAccount,
	)

	data, err := h.service.TransferWithLock(ctx, req)
	if err != nil {
		response.RespondError(ctx, w, http.StatusInternalServerError, err, "transfer_with_lock_handler_error",
			"reference_id", req.ReferenceID,
		)
		return
	}

	response.RespondOK(ctx, w, http.StatusOK, data)
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request, accountID *string, route string) {
	q := r.URL.Query()
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, route)...)

	cursor, err := pagination.DecodeCursor(q.Get("cursor"))
	if err != nil {
		response.RespondError(ctx, w, http.StatusBadRequest, err, "transaction_list_invalid_cursor",
			"account_id", accountID,
		)
		return
	}

	limit := 20
	fmt.Sscanf(q.Get("limit"), "%d", &limit)

	filter := domain.TransactionListFilter{
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
		response.RespondError(ctx, w, http.StatusInternalServerError, err, "transaction_list_handler_error",
			"account_id", accountID,
			"limit", limit,
		)
		return
	}

	response.RespondOKWithMeta(ctx, w, http.StatusOK, data, map[string]interface{}{
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
// @Success 200 {object} response.APIResponse{data=[]dto.TransactionHistoryResponse,meta=object}
// @Failure 400 {object} response.APIResponse{error=string}
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
// @Success 200 {object} response.APIResponse{data=[]dto.TransactionHistoryResponse,meta=object}
// @Failure 400 {object} response.APIResponse{error=string}
// @Router /transactions [get]
func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	h.handleList(w, r, nil, "GET /transactions")
}
