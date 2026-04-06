package transaction

import (
	"core-banking/internal/dto"
	"core-banking/internal/service/transaction"
	"core-banking/pkg/logging"
	"core-banking/pkg/pagination"
	"core-banking/pkg/response"
	"encoding/json"
	"fmt"
	"net/http"

	"core-banking/internal/domain"
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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Ctx(ctx).Warnw("transfer_request_invalid_body",
			"error", err,
		)
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	logging.Ctx(ctx).Infow("transfer_handler_called",
		"reference_id", req.ReferenceID,
		"from_account", req.FromAccount,
		"to_account", req.ToAccount,
	)

	data, err := h.service.Transfer(r.Context(), req)
	if err != nil {
		logging.Ctx(ctx).Errorw("transfer_handler_error",
			"reference_id", req.ReferenceID,
			"error", err,
		)
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{
		Data:    data,
		Success: true,
		Message: "success",
	})
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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Ctx(ctx).Warnw("transfer_with_lock_request_invalid_body",
			"error", err,
		)
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	logging.Ctx(ctx).Infow("transfer_with_lock_handler_called",
		"reference_id", req.ReferenceID,
		"from_account", req.FromAccount,
		"to_account", req.ToAccount,
	)

	data, err := h.service.TransferWithLock(r.Context(), req)
	if err != nil {
		logging.Ctx(ctx).Errorw("transfer_with_lock_handler_error",
			"reference_id", req.ReferenceID,
			"error", err,
		)
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{
		Success: true,
		Data:    data,
		Message: "success",
	})
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request, accountID *string) {
	q := r.URL.Query()
	ctx := r.Context()

	cursor, err := pagination.DecodeCursor(q.Get("cursor"))
	if err != nil {
		logging.Ctx(ctx).Debugw("transaction_list_invalid_cursor",
			"account_id", accountID,
			"error", err,
		)
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: "invalid cursor"})
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

	data, total, nextCursor, prevCursor, err := h.service.List(r.Context(), filter)
	if err != nil {
		logging.Ctx(ctx).Errorw("transaction_list_handler_error",
			"account_id", accountID,
			"limit", limit,
			"error", err,
		)
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	logging.Ctx(ctx).Debugw("transaction_list_retrieved",
		"account_id", accountID,
		"limit", limit,
		"total", total,
	)

	response.JSON(w, http.StatusOK, response.APIResponse{
		Data: data,
		Meta: map[string]interface{}{
			"limit":       limit,
			"next_cursor": nextCursor,
			"prev_cursor": prevCursor,
			"total":       total,
		},
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
	h.handleList(w, r, &accountID)
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
	h.handleList(w, r, nil)
}
