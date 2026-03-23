package transaction

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
