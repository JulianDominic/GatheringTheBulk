package middleware

import (
	"log/slog"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("REQUEST RECEIVED", "protocol", r.Proto, "method", r.Method, "url", r.URL)
		next.ServeHTTP(w, r)
	})
}
