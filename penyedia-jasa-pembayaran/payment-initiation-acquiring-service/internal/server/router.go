package server

import (
	"net/http"
	_ "payment-initiation-acquiring-service/docs"
	"payment-initiation-acquiring-service/internal/transfer"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewRouter(transferH *transfer.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1.0/transfer-intrabank", transferH.Transfer)

	// Swagger
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	return otelhttp.NewHandler(mux, "payment-initiation-service")
}
