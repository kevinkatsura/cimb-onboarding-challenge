package middleware

import "net/http"

func Idempotency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "Missing Idempotency-Key", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
