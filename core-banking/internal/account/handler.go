package account

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

// Responses for Swagger tracking
type ErrorResponse response.ErrorResponse

type ResponseSuccessAccount struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	ID              string `json:"id"`
	AccountNumber   string `json:"account_number"`
	ProductCode     string `json:"product_code"`
}

type ResponseSuccessAccountList struct {
	ResponseCode    string                 `json:"responseCode"`
	ResponseMessage string                 `json:"responseMessage"`
	Data            []Account              `json:"data"`
	Meta            map[string]interface{} `json:"meta"`
}

type ResponseSuccessString struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	Data            string `json:"data"`
}

// Create creates a new bank account.
//
//	@Summary		Create Account
//	@Description	Creates a new bank account tied to a customer ID.
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateAccountRequest	true	"Account Payload"
//	@Success		200		{object}	ResponseSuccessAccount
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/accounts [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "POST /accounts")...)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(ctx, w, snap.AccountServiceCode, "account_create_invalid_request_body", apperror.Wrap(apperror.ErrInvalidFieldFormat, "Bad Request", err))
		return
	}

	acc, err := h.service.CreateAccount(ctx, req)
	if err != nil {
		response.WriteError(ctx, w, snap.AccountServiceCode, "account_create_failed", err,
			"customer_id", req.CustomerID,
			"product_code", req.ProductCode,
		)
		return
	}

	logging.Ctx(ctx).Infow("account_created_via_handler",
		"account_id", acc.ID,
		"account_number", acc.AccountNumber,
	)
	response.WriteSuccess(ctx, w, snap.AccountServiceCode, acc, nil)
}

// Register creates a new customer and default account for registration flow.
//
//	@Summary		Register Account
//	@Description	Creates a new customer profile and a default savings account. Service Code 06.
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			X-PARTNER-ID	header		string								true	"Partner ID provided by the bank"			default(PARTNER001)
//	@Param			X-TIMESTAMP		header		string								true	"ISO-8601 Timestamp"						default(2026-04-12T18:00:00Z)
//	@Param			X-SIGNATURE		header		string								true	"HMAC-SHA256 Signature"					default(valid-signature-for-testing)
//	@Param			X-EXTERNAL-ID	header		string								true	"Random unique ID representing the request"	default(1234567890)
//	@Param			request			body		RegistrationAccountCreationRequest	true	"Registration Payload"
//	@Success		200		{object}	ResponseSuccessAccount
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v1.0/registration-account-creation [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegistrationAccountCreationRequest
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "POST /v1.0/registration-account-creation")...)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(ctx, w, snap.RegistrationAccountCreationServiceCode, "registration_invalid_body", apperror.Wrap(apperror.ErrInvalidFieldFormat, "Bad Request", err))
		return
	}

	acc, err := h.service.RegisterAccount(ctx, req)
	if err != nil {
		response.WriteError(ctx, w, snap.RegistrationAccountCreationServiceCode, "registration_failed", err)
		return
	}

	logging.Ctx(ctx).Infow("registration_completed",
		"customer_id", req.CustomerID,
		"account_number", acc.AccountNumber,
	)

	res := RegistrationAccountCreationResponse{
		ReferenceNo:        acc.ID.String(),
		PartnerReferenceNo: req.PartnerReferenceNo,
		ApiKey:             req.CustomerID,
		AccountID:          acc.AccountNumber,
		State:              req.State,
		AdditionalInfo:     req.AdditionalInfo,
	}

	response.WriteSuccess(ctx, w, snap.RegistrationAccountCreationServiceCode, res, nil)
}

// Get fetches a single bank account uniquely.
//
//	@Summary		Get Account by ID
//	@Description	Fetches a bank account from the system given its UUID.
//	@Tags			accounts
//	@Produce		json
//	@Param			id	path		string	true	"Account ID"
//	@Success		200	{object}	ResponseSuccessAccount
//	@Failure		404	{object}	ErrorResponse
//	@Router			/accounts/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "GET /accounts/{id}")...)

	acc, err := h.service.GetAccount(ctx, accountID)
	if err != nil {
		response.WriteError(ctx, w, snap.AccountServiceCode, "account_get_not_found", err,
			"account_id", accountID,
		)
		return
	}

	response.WriteSuccess(ctx, w, snap.AccountServiceCode, acc, nil)
}

// List retrieves accounts under pagination scopes natively.
//
//	@Summary		List Accounts
//	@Description	Lists all bank accounts under specific cursors and filters.
//	@Tags			accounts
//	@Produce		json
//	@Param			limit		query		int		false	"Limits response objects"
//	@Param			cursor		query		string	false	"Pagination Base64 Cursor"
//	@Param			customer_id	query		string	false	"Filter by Customer ID"
//	@Param			status		query		string	false	"Filter by Status (active, suspended, closed)"
//	@Success		200			{object}	ResponseSuccessAccountList
//	@Failure		400			{object}	ErrorResponse
//	@Router			/accounts [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "GET /accounts")...)

	limit := 20
	fmt.Sscanf(q.Get("limit"), "%d", &limit)

	cursor, err := pagination.DecodeCursor(q.Get("cursor"))
	if err != nil {
		response.WriteError(ctx, w, snap.AccountServiceCode, "account_list_invalid_cursor", apperror.Wrap(apperror.ErrBadRequest, "Bad Request", err))
		return
	}

	filter := ListFilter{
		Limit:     limit,
		Cursor:    cursor,
		Direction: q.Get("direction"),
	}

	if v := q.Get("customer_id"); v != "" {
		filter.CustomerID = &v
	}
	if v := q.Get("product_code"); v != "" {
		filter.ProductCode = &v
	}
	if v := q.Get("status"); v != "" {
		filter.Status = &v
	}
	if v := q.Get("currency"); v != "" {
		filter.Currency = &v
	}

	data, total, nextCursor, prevCursor, err := h.service.ListAccounts(ctx, filter)
	if err != nil {
		response.WriteError(ctx, w, snap.AccountServiceCode, "account_list_failed", err,
			"limit", limit,
			"customer_id", filter.CustomerID,
		)
		return
	}

	response.WriteSuccess(ctx, w, snap.AccountServiceCode, data, map[string]interface{}{
		"limit":       limit,
		"next_cursor": nextCursor,
		"prev_cursor": prevCursor,
		"total":       total,
	})
}

// UpdateStatus transitions an account between active/suspended states.
//
//	@Summary		Update Account Status
//	@Description	Manually transitions the active lifecycle state of a bank account.
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Account ID"
//	@Param			request	body		UpdateAccountStatusRequest	true	"Update Payload"
//	@Success		200		{object}	ResponseSuccessString
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/accounts/{id} [patch]
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "PATCH /accounts/{id}")...)

	var req UpdateAccountStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(ctx, w, snap.AccountServiceCode, "account_update_status_invalid_request", apperror.Wrap(apperror.ErrInvalidFieldFormat, "Bad Request", err),
			"account_id", accountID,
		)
		return
	}

	err := h.service.UpdateStatus(ctx, accountID, req.Status)
	if err != nil {
		response.WriteError(ctx, w, snap.AccountServiceCode, "account_update_status_failed", err,
			"account_id", accountID,
			"requested_status", req.Status,
		)
		return
	}

	response.WriteSuccess(ctx, w, snap.AccountServiceCode, "updated", nil)
}

// Delete softly purges an account.
//
//	@Summary		Delete Account
//	@Description	Soft-deletes a bank account if allowed by constraints.
//	@Tags			accounts
//	@Produce		json
//	@Param			id	path		string	true	"Account ID"
//	@Success		200	{object}	ResponseSuccessString
//	@Failure		400	{object}	ErrorResponse
//	@Router			/accounts/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(telemetry.HandlerAttrs(r, "DELETE /accounts/{id}")...)

	err := h.service.DeleteAccount(ctx, accountID)
	if err != nil {
		response.WriteError(ctx, w, snap.AccountServiceCode, "account_deletion_failed", err,
			"account_id", accountID,
		)
		return
	}

	logging.Ctx(ctx).Infow("account_deleted", "account_id", accountID)
	response.WriteSuccess(ctx, w, snap.AccountServiceCode, "deleted", nil)
}
