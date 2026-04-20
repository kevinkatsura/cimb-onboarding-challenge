package server

import (
	_ "account-issuance-service/docs"
	"account-issuance-service/internal/account"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewRouter(accountH *account.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1.0/registration-account-creation", accountH.RegisterAccount)

	// Swagger
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return otelhttp.NewHandler(mux, "account-issuance-service")
}
