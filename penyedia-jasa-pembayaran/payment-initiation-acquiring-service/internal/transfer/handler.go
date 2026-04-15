package transfer

import (
	"encoding/json"
	"net/http"

	"payment-initiation-acquiring-service/pkg/response"
	"payment-initiation-acquiring-service/pkg/validator"
)

const serviceCode = "17"

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Transfer godoc
// @Summary      Intrabank Transfer
// @Description  Execute an intrabank fund transfer (SNAP Service Code 17)
// @Tags         Transfer
// @Accept       json
// @Produce      json
// @Param        X-TIMESTAMP     header string true  "Request timestamp"       default(2026-04-15T10:00:00+07:00)
// @Param        X-SIGNATURE     header string true  "Digital signature"       default(PLACEHOLDER)
// @Param        X-PARTNER-ID    header string true  "Partner ID"              default(PJP-001)
// @Param        X-EXTERNAL-ID   header string true  "External ID (idempotency)" default(ext-transfer-001)
// @Param        CHANNEL-ID      header string true  "Channel ID"             default(CH001)
// @Param        body body TransferRequest true "Transfer payload"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /v1.0/transfer-intrabank [post]
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(ctx, w, serviceCode, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		response.WriteError(ctx, w, serviceCode, err)
		return
	}

	resp, err := h.svc.Transfer(ctx, req)
	if err != nil {
		response.WriteError(ctx, w, serviceCode, err)
		return
	}

	response.WriteSuccess(ctx, w, serviceCode, resp)
}
