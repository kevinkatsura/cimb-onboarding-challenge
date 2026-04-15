package server

import (
	"net/http"
	"payment-initiation-acquiring-service/internal/transfer"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewRouter(transferH *transfer.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1.0/transfer-intrabank", transferH.Transfer)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	return otelhttp.NewHandler(mux, "payment-initiation-service")
}
