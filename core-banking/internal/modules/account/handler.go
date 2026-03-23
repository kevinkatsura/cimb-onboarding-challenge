package account

import (
	"core-banking/internal/pkg/response"
	"encoding/json"
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
	id := r.URL.Query().Get("id")

	acc, err := h.service.GetAccount(r.Context(), id)
	if err != nil {
		response.JSON(w, http.StatusNotFound, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: acc})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")

	accs, err := h.service.ListAccounts(r.Context(), customerID)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: accs})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	var req UpdateAccountStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.APIResponse{Error: err.Error()})
		return
	}

	err := h.service.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: "updated"})
}
