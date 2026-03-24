package account

import (
	"core-banking/internal/pkg/pagination"
	"core-banking/internal/pkg/response"
	"encoding/json"
	"fmt"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	acc, err := h.service.CreateAccount(r.Context(), req)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: acc})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")

	acc, err := h.service.GetAccount(r.Context(), accountID)
	if err != nil {
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
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: "invalid cursor"})
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
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

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

	var req UpdateAccountStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	err := h.service.UpdateStatus(r.Context(), accountID, req.Status)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: "updated"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")

	err := h.service.DeleteAccount(r.Context(), accountID)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: "deleted"})
}
