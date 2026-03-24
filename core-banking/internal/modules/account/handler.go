package account

import (
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

func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limit := 20
	offset := 0

	if l := query.Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := query.Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	accounts, err := h.service.ListAllAccounts(r.Context(), limit, offset)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: accounts})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("id")

	accs, err := h.service.ListAccounts(r.Context(), customerID)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.APIResponse{Error: err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{Data: accs})
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
