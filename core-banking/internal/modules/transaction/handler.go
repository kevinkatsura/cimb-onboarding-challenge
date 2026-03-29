package transaction

import (
	"core-banking/internal/pkg/pagination"
	"core-banking/internal/pkg/response"
	"encoding/json"
	"fmt"
	"net/http"
)

type Handler struct {
	service *TransactionServiceInterface
}

func NewHandler(service *TransactionServiceInterface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	err := h.service.Transfer(r.Context(), req)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: "success"})
}

func (h *Handler) TransferWithLock(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	data, err := h.service.TransferWithLock(r.Context(), req)
	if err != nil {
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

	cursor, err := pagination.DecodeCursor(q.Get("cursor"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: "invalid cursor"})
		return
	}

	limit := 20
	fmt.Sscanf(q.Get("limit"), "%d", &limit)

	filter := ListFilter{
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

func (h *Handler) ListByAccount(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	h.handleList(w, r, &accountID)
}

func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	h.handleList(w, r, nil)
}
