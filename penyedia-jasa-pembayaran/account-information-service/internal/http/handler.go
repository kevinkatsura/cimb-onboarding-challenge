package http

import (
	"encoding/json"
	"net/http"

	"github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/internal/repository"
)

type Handler struct {
	repo *repository.PostgresDatabase
}

func NewHandler(repo *repository.PostgresDatabase) *Handler {
	return &Handler{repo: repo}
}

// GetAccountInfo godoc
// @Summary      Get Account Information
// @Description  Returns basic account information
// @Tags         Account Information
// @Produce      json
// @Param        account_number path string true "Account Number"
// @Success      200 {object} repository.Account
// @Failure      404 {object} map[string]string
// @Router       /v1/accounts/{account_number} [get]
func (h *Handler) GetAccountInfo(w http.ResponseWriter, r *http.Request) {
	accNo := r.PathValue("account_number")
	acc, err := h.repo.GetAccountInfo(r.Context(), accNo)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "account not found"})
		return
	}
	json.NewEncoder(w).Encode(acc)
}

// GetBalance godoc
// @Summary      Get Account Balance
// @Description  Returns the current balance
// @Tags         Account Information
// @Produce      json
// @Param        account_number path string true "Account Number"
// @Success      200 {object} map[string]interface{}
// @Failure      404 {object} map[string]string
// @Router       /v1/accounts/{account_number}/balance [get]
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	accNo := r.PathValue("account_number")
	bal, cur, err := h.repo.GetBalance(r.Context(), accNo)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "account not found"})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"account_number": accNo, "balance": bal, "currency": cur})
}

// GetLastTransactionAsSource godoc
// @Summary      Get Last Transaction As Source
// @Description  Returns the last transaction where the account is the source
// @Tags         Account Information
// @Produce      json
// @Param        account_number path string true "Account Number"
// @Success      200 {object} repository.Transaction
// @Failure      404 {object} map[string]string
// @Router       /v1/accounts/{account_number}/transactions/source/last [get]
func (h *Handler) GetLastTransactionAsSource(w http.ResponseWriter, r *http.Request) {
	accNo := r.PathValue("account_number")
	tx, err := h.repo.GetLastTransactionAsSource(r.Context(), accNo)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "transaction not found"})
		return
	}
	json.NewEncoder(w).Encode(tx)
}

// GetLastTransactionAsBeneficiary godoc
// @Summary      Get Last Transaction As Beneficiary
// @Description  Returns the last transaction where the account is the beneficiary
// @Tags         Account Information
// @Produce      json
// @Param        account_number path string true "Account Number"
// @Success      200 {object} repository.Transaction
// @Failure      404 {object} map[string]string
// @Router       /v1/accounts/{account_number}/transactions/beneficiary/last [get]
func (h *Handler) GetLastTransactionAsBeneficiary(w http.ResponseWriter, r *http.Request) {
	accNo := r.PathValue("account_number")
	tx, err := h.repo.GetLastTransactionAsBeneficiary(r.Context(), accNo)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "transaction not found"})
		return
	}
	json.NewEncoder(w).Encode(tx)
}

// GetAverageAmountLast30Transactions godoc
// @Summary      Get Average Amount Last 30 Transactions
// @Description  Returns the average transaction amount of the last 30 transactions
// @Tags         Account Information
// @Produce      json
// @Param        account_number path string true "Account Number"
// @Success      200 {object} map[string]interface{}
// @Failure      500 {object} map[string]string
// @Router       /v1/accounts/{account_number}/transactions/average30 [get]
func (h *Handler) GetAverageAmountLast30Transactions(w http.ResponseWriter, r *http.Request) {
	accNo := r.PathValue("account_number")
	avg, cur, err := h.repo.GetAverageAmountLast30Transactions(r.Context(), accNo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to calculate average"})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"account_number": accNo, "average_amount": avg, "currency": cur})
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /v1/accounts/{account_number}", h.GetAccountInfo)
	mux.HandleFunc("GET /v1/accounts/{account_number}/balance", h.GetBalance)
	mux.HandleFunc("GET /v1/accounts/{account_number}/transactions/source/last", h.GetLastTransactionAsSource)
	mux.HandleFunc("GET /v1/accounts/{account_number}/transactions/beneficiary/last", h.GetLastTransactionAsBeneficiary)
	mux.HandleFunc("GET /v1/accounts/{account_number}/transactions/average30", h.GetAverageAmountLast30Transactions)
}
