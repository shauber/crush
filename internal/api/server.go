package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/charmbracelet/crush/internal/app"
)

func Serve(app *app.App, port int) error {
	mux := http.NewServeMux()

	// Create handlers with app instance
	h := &handlers{app: app}

	// Register routes
	mux.HandleFunc("POST /sessions", h.createSession)
	mux.HandleFunc("GET /sessions", h.listSessions)
	mux.HandleFunc("GET /sessions/{id}", h.getSession)
	mux.HandleFunc("DELETE /sessions/{id}", h.deleteSession)

	mux.HandleFunc("POST /sessions/{id}/messages", h.sendMessage)
	mux.HandleFunc("GET /sessions/{id}/messages", h.listMessages)
	mux.HandleFunc("GET /sessions/{id}/stream", h.streamSession)

	mux.HandleFunc("POST /sessions/{id}/cancel", h.cancelSession)
	mux.HandleFunc("GET /sessions/{id}/status", h.sessionStatus)

	// Model picker endpoints
	mux.HandleFunc("GET /models", h.listModels)
	mux.HandleFunc("GET /providers", h.getProviders)
	mux.HandleFunc("POST /models", h.setModel)

	// Wrap with CORS and logging
	handler := corsMiddleware(loggingMiddleware(mux))

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	slog.Info("Starting API server", "addr", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server.ListenAndServe()
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
