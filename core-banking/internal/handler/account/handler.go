package account

import (
	"core-banking/internal/dto"
	"core-banking/internal/service/account"
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
	service account.Interface
}

func NewHandler(service account.Interface) *Handler {
	return &Handler{service: service}
}

// Create creates a new bank account.
// @Summary Create Account
// @Description Creates a new bank account tied to a customer ID.
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body dto.CreateAccountRequest true "Account Payload"
// @Success 200 {object} response.APIResponse{data=domain.Account}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /accounts [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateAccountRequest
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "POST /accounts")...)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(ctx, w, http.StatusBadRequest, err, "account_create_invalid_request_body")
		return
	}

	acc, err := h.service.CreateAccount(ctx, req)
	if err != nil {
		response.RespondError(ctx, w, http.StatusInternalServerError, err, "account_create_failed",
			"customer_id", req.CustomerID,
			"account_type", req.AccountType,
		)
		return
	}

	logging.Ctx(ctx).Infow("account_created_via_handler",
		"account_id", acc.ID,
		"account_number", acc.AccountNumber,
	)
	response.RespondOK(ctx, w, http.StatusOK, acc)
}

// Get fetches a single bank account uniquely.
// @Summary Get Account by ID
// @Description Fetches a bank account from the system given its UUID.
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} response.APIResponse{data=domain.Account}
// @Failure 404 {object} response.APIResponse{error=string}
// @Router /accounts/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "GET /accounts/{id}")...)

	acc, err := h.service.GetAccount(ctx, accountID)
	if err != nil {
		response.RespondError(ctx, w, http.StatusNotFound, err, "account_get_not_found",
			"account_id", accountID,
		)
		return
	}

	response.RespondOK(ctx, w, http.StatusOK, acc)
}

// List retrieves accounts under pagination scopes natively.
// @Summary List Accounts
// @Description Lists all bank accounts under specific cursors and filters.
// @Tags accounts
// @Produce json
// @Param limit query int false "Limits response objects"
// @Param cursor query string false "Pagination Base64 Cursor"
// @Param customer_id query string false "Filter by Customer ID"
// @Param status query string false "Filter by Status (active, suspended, closed)"
// @Success 200 {object} response.APIResponse{data=[]domain.Account,meta=object}
// @Failure 400 {object} response.APIResponse{error=string}
// @Router /accounts [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "GET /accounts")...)

	limit := 20
	fmt.Sscanf(q.Get("limit"), "%d", &limit)

	cursor, err := pagination.DecodeCursor(q.Get("cursor"))
	if err != nil {
		response.RespondError(ctx, w, http.StatusBadRequest, err, "account_list_invalid_cursor")
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

	data, total, nextCursor, prevCursor, err := h.service.ListAccounts(ctx, filter)
	if err != nil {
		response.RespondError(ctx, w, http.StatusInternalServerError, err, "account_list_failed",
			"limit", limit,
			"customer_id", filter.CustomerID,
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

// UpdateStatus transitions an account between active/suspended states.
// @Summary Update Account Status
// @Description Manually transitions the active lifecycle state of a bank account.
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param request body dto.UpdateAccountStatusRequest true "Update Payload"
// @Success 200 {object} response.APIResponse{data=string}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /accounts/{id} [patch]
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "PATCH /accounts/{id}")...)

	var req dto.UpdateAccountStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(ctx, w, http.StatusBadRequest, err, "account_update_status_invalid_request",
			"account_id", accountID,
		)
		return
	}

	err := h.service.UpdateStatus(ctx, accountID, req.Status)
	if err != nil {
		response.RespondError(ctx, w, http.StatusInternalServerError, err, "account_update_status_failed",
			"account_id", accountID,
			"requested_status", req.Status,
		)
		return
	}

	response.RespondOK(ctx, w, http.StatusOK, "updated")
}

// Delete softly purges an account.
// @Summary Delete Account
// @Description Soft-deletes a bank account if allowed by constraints.
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} response.APIResponse{data=string}
// @Failure 400 {object} response.APIResponse{error=string}
// @Router /accounts/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "DELETE /accounts/{id}")...)

	err := h.service.DeleteAccount(ctx, accountID)
	if err != nil {
		response.RespondError(ctx, w, http.StatusBadRequest, err, "account_deletion_failed",
			"account_id", accountID,
		)
		return
	}

	logging.Ctx(ctx).Infow("account_deleted", "account_id", accountID)
	response.RespondOK(ctx, w, http.StatusOK, "deleted")
}
