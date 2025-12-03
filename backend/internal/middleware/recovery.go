package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer tells Go to run this before it exits
		defer func() {
			if err := recover(); err != nil {
				slog.Error(
					"Recovery executed",
					"error", err,
					"stack trace", string(debug.Stack()),
				)
				http.Error(
					w,
					"Internal Server Error",
					http.StatusInternalServerError,
				)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
