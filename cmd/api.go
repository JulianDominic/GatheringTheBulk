package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/JulianDominic/GatheringTheBulk/internal/healthcheck"
	"github.com/JulianDominic/GatheringTheBulk/internal/middleware"
)

// mount
func (app *api) mount() http.Handler {
	mux := http.NewServeMux()

	// handlers
	mux.HandleFunc("GET /health", healthcheck.HealthCheckHandler)

	// middleware
	middlewares := []func(http.Handler) http.Handler{
		middleware.LoggingMiddleware, // this comes first
		middleware.RecoveryMiddleware,
	}

	var wrappedHandler http.Handler = mux

	for i := len(middlewares) - 1; i >= 0; i-- {
		wrappedHandler = middlewares[i](wrappedHandler)
	}

	return wrappedHandler
}

// run
func (app *api) run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	slog.Info("Server started", "addr", app.config.addr)

	return srv.ListenAndServe()
}

type api struct {
	config config
}

type config struct {
	addr string // 10800
	db   dbConfig
}

type dbConfig struct {
	// data source name
	dsn string
}
