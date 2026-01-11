package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JulianDominic/GatheringTheBulk/internal/api/common"
	"github.com/JulianDominic/GatheringTheBulk/internal/api/inventory"
	"github.com/JulianDominic/GatheringTheBulk/internal/api/jobs"
	"github.com/JulianDominic/GatheringTheBulk/internal/api/pages"
	"github.com/JulianDominic/GatheringTheBulk/internal/api/review"
	"github.com/JulianDominic/GatheringTheBulk/internal/database"
	"github.com/JulianDominic/GatheringTheBulk/internal/store"
	"github.com/JulianDominic/GatheringTheBulk/internal/worker"
)

func main() {
	// 1. Initialize DB
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "inventory.db"
	}
	if err := database.InitDB(dsn); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer database.Close()

	// 2. Initialize Dependencies
	s := store.NewSQLiteStore(database.DB)
	renderer := &common.Renderer{Store: s}
	dispatcher := worker.NewDispatcher(s, 100)
	dispatcher.Start(3)

	// 3. Initialize Handlers
	pagesHandler := &pages.Handler{Store: s, Renderer: renderer}
	inventoryHandler := &inventory.Handler{Store: s, Renderer: renderer}
	reviewHandler := &review.Handler{Store: s, Renderer: renderer}
	jobsHandler := &jobs.Handler{Store: s, Dispatcher: dispatcher}

	// 4. Setup Routes
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Pages
	mux.HandleFunc("GET /", pagesHandler.HandleDashboard)
	mux.HandleFunc("GET /settings", pagesHandler.HandleSettings)
	mux.HandleFunc("GET /import", pagesHandler.HandleImportHub)

	// API / HTMX
	mux.HandleFunc("GET /api/jobs/{id}", jobsHandler.HandleStatus)
	mux.HandleFunc("POST /api/jobs/sync", jobsHandler.HandleSync)
	mux.HandleFunc("POST /api/jobs/import", jobsHandler.HandleImport)
	mux.HandleFunc("GET /api/search", inventoryHandler.HandleSearch)
	mux.HandleFunc("GET /api/inventory/autocomplete", inventoryHandler.HandleAutocomplete)

	// Inventory
	mux.HandleFunc("GET /inventory/edit/{id}", inventoryHandler.HandleEditModal)
	mux.HandleFunc("GET /inventory/add-details/{scryfall_id}", inventoryHandler.HandleAddDetails)
	mux.HandleFunc("POST /inventory", inventoryHandler.HandleAdd)
	mux.HandleFunc("PUT /inventory/{id}", inventoryHandler.HandleEdit)
	mux.HandleFunc("DELETE /inventory/{id}", inventoryHandler.HandleDelete)

	// Review
	mux.HandleFunc("GET /review/content", reviewHandler.HandleContent)
	mux.HandleFunc("GET /review/resolve/{id}", reviewHandler.HandleResolveModal)
	mux.HandleFunc("GET /review/resolve-select/{scryfall_id}", reviewHandler.HandleResolveSelect)
	mux.HandleFunc("DELETE /review/{id}", reviewHandler.HandleDelete)
	mux.HandleFunc("POST /review/resolve", reviewHandler.HandleResolve)
	mux.HandleFunc("GET /api/review/badge", reviewHandler.HandleBadge)

	// 5. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: loggingMiddleware(mux),
	}

	log.Printf("Server starting on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// Logging Middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		log.Printf("%d %s %s %v", wrapped.statusCode, r.Method, r.URL.RequestURI(), time.Since(start))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
