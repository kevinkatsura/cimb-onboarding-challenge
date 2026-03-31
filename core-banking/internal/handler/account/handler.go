package account

import (
	"core-banking/internal/dto"
	"core-banking/internal/service/account"
	"core-banking/pkg/logging"
	"core-banking/pkg/pagination"
	"core-banking/pkg/response"
	"encoding/json"
	"fmt"
	"net/http"

	"core-banking/internal/domain"
)

type Handler struct {
	service account.Interface
}

func NewHandler(service account.Interface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Logger().Warnw("account_create_invalid_request_body",
			"error", err,
		)
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	acc, err := h.service.CreateAccount(r.Context(), req)
	if err != nil {
		logging.Logger().Errorw("account_create_failed",
			"customer_id", req.CustomerID,
			"account_type", req.AccountType,
			"error", err,
		)
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	logging.Logger().Infow("account_created_via_handler",
		"account_id", acc.ID,
		"account_number", acc.AccountNumber,
	)
	response.JSON(w, http.StatusOK, response.APIResponse{Data: acc})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")

	logging.Logger().Debugw("account_get_request",
		"account_id", accountID,
	)

	acc, err := h.service.GetAccount(r.Context(), accountID)
	if err != nil {
		logging.Logger().Warnw("account_get_not_found",
			"account_id", accountID,
			"error", err,
		)
		response.JSON(w, http.StatusNotFound, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: acc})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := 20
	fmt.Sscanf(q.Get("limit"), "%d", &limit)

	cursor, err := pagination.DecodeCursor(q.Get("cursor"))
	if err != nil {
		logging.Logger().Debugw("account_list_invalid_cursor", "error", err)
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: "invalid cursor"})
		return
	}

	filter := domain.ListFilter{
		Limit:     limit,
		Cursor:    cursor,
		Direction: q.Get("direction"),
	}

	if v := q.Get("customer_id"); v != "" {
		filter.CustomerID = &v
	}
	if v := q.Get("account_type"); v != "" {
		filter.AccountType = &v
	}
	if v := q.Get("status"); v != "" {
		filter.Status = &v
	}
	if v := q.Get("currency"); v != "" {
		filter.Currency = &v
	}

	data, total, nextCursor, prevCursor, err := h.service.ListAccounts(r.Context(), filter)
	if err != nil {
		logging.Logger().Errorw("account_list_failed",
			"limit", limit,
			"customer_id", filter.CustomerID,
			"error", err,
		)
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	logging.Logger().Debugw("account_list_retrieved",
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

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")

	var req dto.UpdateAccountStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Logger().Warnw("account_update_status_invalid_request",
			"account_id", accountID,
			"error", err,
		)
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	logging.Logger().Infow("account_update_status_requested",
		"account_id", accountID,
		"new_status", req.Status,
	)

	err := h.service.UpdateStatus(r.Context(), accountID, req.Status)
	if err != nil {
		logging.Logger().Errorw("account_update_status_failed",
			"account_id", accountID,
			"requested_status", req.Status,
			"error", err,
		)
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: "updated"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")

	logging.Logger().Infow("account_deletion_requested",
		"account_id", accountID,
	)

	err := h.service.DeleteAccount(r.Context(), accountID)
	if err != nil {
		logging.Logger().Errorw("account_deletion_failed",
			"account_id", accountID,
			"error", err,
		)
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	logging.Logger().Infow("account_deleted",
		"account_id", accountID,
	)
	response.JSON(w, http.StatusOK, response.APIResponse{Data: "deleted"})
}
