package middleware

import (
	"net/http"
)

// ForceHTTPS permanently redirects proxy-handled HTTP payloads to HTTPS natively.
func ForceHTTPS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

		if !isHTTPS {
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}

		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}
