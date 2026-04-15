package account

import (
	"account-issuance-service/pkg/response"
	"account-issuance-service/pkg/validator"
	"encoding/json"
	"net/http"
)

const serviceCode = "06"

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterAccount godoc
// @Summary      Register a new account
// @Description  Creates a customer and account atomically (SNAP Service Code 06)
// @Tags         Account
// @Accept       json
// @Produce      json
// @Param        X-TIMESTAMP     header string true  "Request timestamp" default(2026-04-15T10:00:00+07:00)
// @Param        X-SIGNATURE     header string true  "Digital signature" default(PLACEHOLDER)
// @Param        X-PARTNER-ID    header string true  "Partner ID"        default(PJP-001)
// @Param        X-EXTERNAL-ID   header string true  "External ID"      default(ext-001)
// @Param        CHANNEL-ID      header string true  "Channel ID"       default(CH001)
// @Param        body body RegistrationRequest true "Registration payload"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /v1.0/registration-account-creation [post]
func (h *Handler) RegisterAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(ctx, w, serviceCode, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		response.WriteError(ctx, w, serviceCode, err)
		return
	}

	resp, err := h.svc.RegisterAccount(ctx, req)
	if err != nil {
		response.WriteError(ctx, w, serviceCode, err)
		return
	}

	response.WriteSuccess(ctx, w, serviceCode, resp)
}
