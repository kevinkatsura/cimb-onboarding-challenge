package server

import (
	"database-exercise/internal/response"
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

// Chain middleware
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// JSON Content-TYpe Enforcer
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Force response type
		w.Header().Set("Content-Type", "application/json")

		// Validate request content-type for POST/PUT
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if r.Header.Get("Content-Type") != "application/json" {
				response.Error(w, http.StatusUnsupportedMediaType, nil,
					"Content-Type must be application/json")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Panic Recovery Middleware
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("panic:", err)
				response.Error(w, http.StatusInternalServerError, err, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Logging Middleware (lightweight)
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
