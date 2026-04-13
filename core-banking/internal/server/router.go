package server

import (
	"net/http"

	account "core-banking/internal/account"
	transaction "core-banking/internal/transaction"
	"core-banking/pkg/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewRouter(accountH *account.Handler, txH *transaction.Handler) http.Handler {
	mux := http.NewServeMux()

	// ---- Swagger ----
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// ---- Prometheus ----
	mux.Handle("GET /metrics", promhttp.Handler())

	// ---- Transaction Endpoints ----
	mux.Handle("POST /v1.0/transfer-intrabank", middleware.SNAP(otelhttp.NewHandler(http.HandlerFunc(txH.Transfer), "TransactionHandler.Transfer")))
	mux.Handle("POST /v1.0/transfer-intrabank-locked", middleware.SNAP(otelhttp.NewHandler(http.HandlerFunc(txH.TransferWithLock), "TransactionHandler.TransferWithLock")))
	mux.Handle("POST /v1.0/transfer/status", middleware.SNAP(otelhttp.NewHandler(http.HandlerFunc(txH.TransferStatusInquiry), "TransactionHandler.TransferStatusInquiry")))

	mux.Handle("GET /transactions", otelhttp.NewHandler(http.HandlerFunc(txH.ListAll), "TransactionHandler.ListAll"))
	mux.Handle("GET /accounts/{id}/transactions", otelhttp.NewHandler(http.HandlerFunc(txH.ListByAccount), "TransactionHandler.ListByAccount"))

	// ---- Account Endpoints ----
	mux.Handle("GET /accounts", otelhttp.NewHandler(http.HandlerFunc(accountH.List), "AccountHandler.ListAccounts"))
	mux.Handle("GET /accounts/{id}", otelhttp.NewHandler(http.HandlerFunc(accountH.Get), "AccountHandler.GetAccountByID"))
	mux.Handle("POST /accounts", otelhttp.NewHandler(http.HandlerFunc(accountH.Create), "AccountHandler.CreateAccount"))
	mux.Handle("POST /v1.0/registration-account-creation", middleware.SNAP(otelhttp.NewHandler(http.HandlerFunc(accountH.Register), "AccountHandler.Register")))
	mux.Handle("PATCH /accounts/{id}", otelhttp.NewHandler(http.HandlerFunc(accountH.UpdateStatus), "AccountHandler.UpdateStatus"))
	mux.Handle("DELETE /accounts/{id}", otelhttp.NewHandler(http.HandlerFunc(accountH.Delete), "AccountHandler.DeleteAccount"))

	return middleware.ForceHTTPS(middleware.CORS(mux))
}
