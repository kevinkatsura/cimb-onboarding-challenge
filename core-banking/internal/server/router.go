package server

import (
	"net/http"

	accountHandler "core-banking/internal/handler/account"
	txHandler "core-banking/internal/handler/transaction"
	"core-banking/pkg/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewRouter builds the HTTP handler with all routes, middleware, Swagger, and Prometheus.
func NewRouter(accountH *accountHandler.Handler, txH *txHandler.Handler) http.Handler {
	mux := http.NewServeMux()

	// ---- Swagger ----
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// ---- Prometheus ----
	mux.Handle("GET /metrics", promhttp.Handler())

	// ---- Transaction Endpoints ----
	mux.Handle("POST /v1/transfer", otelhttp.NewHandler(http.HandlerFunc(txH.Transfer), "TransactionHandler.Transfer"))
	mux.Handle("POST /v2/transfer", otelhttp.NewHandler(http.HandlerFunc(txH.TransferWithLock), "TransactionHandler.TransferWithLock"))
	mux.Handle("GET /transactions", otelhttp.NewHandler(http.HandlerFunc(txH.ListAll), "TransactionHandler.ListAll"))
	mux.Handle("GET /accounts/{id}/transactions", otelhttp.NewHandler(http.HandlerFunc(txH.ListByAccount), "TransactionHandler.ListByAccount"))

	// ---- Account Endpoints ----
	mux.Handle("GET /accounts", otelhttp.NewHandler(http.HandlerFunc(accountH.List), "AccountHandler.ListAccounts"))
	mux.Handle("GET /accounts/{id}", otelhttp.NewHandler(http.HandlerFunc(accountH.Get), "AccountHandler.GetAccountByID"))
	mux.Handle("POST /accounts", otelhttp.NewHandler(http.HandlerFunc(accountH.Create), "AccountHandler.CreateAccount"))
	mux.Handle("PATCH /accounts/{id}", otelhttp.NewHandler(http.HandlerFunc(accountH.UpdateStatus), "AccountHandler.UpdateStatus"))
	mux.Handle("DELETE /accounts/{id}", otelhttp.NewHandler(http.HandlerFunc(accountH.Delete), "AccountHandler.DeleteAccount"))

	return middleware.ForceHTTPS(middleware.CORS(mux))
}
