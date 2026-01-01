package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/crush/internal/api/streaming"
	"github.com/charmbracelet/crush/internal/app"
	crushstreaming "github.com/charmbracelet/crush/internal/streaming"
)

func Serve(app *app.App, port int) error {
	mux := http.NewServeMux()

	// Create handlers with app instance
	h := &handlers{app: app}

	// Create SSE handler for streaming
	sseHandler := streaming.NewSSEHandler()

	// Start message subscriber for tool streaming events.
	// This subscribes to message pubsub and emits SSE events when tools run.
	msgSubscriber := crushstreaming.NewMessageSubscriber(app.Messages, sseHandler)
	go msgSubscriber.Start(context.Background())

	// Register routes
	mux.HandleFunc("POST /sessions", h.createSession)
	mux.HandleFunc("GET /sessions", h.listSessions)
	mux.HandleFunc("GET /sessions/{id}", h.getSession)
	mux.HandleFunc("DELETE /sessions/{id}", h.deleteSession)

	mux.HandleFunc("POST /sessions/{id}/messages", h.sendMessage)
	mux.HandleFunc("GET /sessions/{id}/messages", h.listMessages)
	mux.HandleFunc("GET /sessions/{id}/stream", h.streamSession)

	// Streaming endpoints
	mux.HandleFunc("GET /stream", sseHandler.HandleSSE)
	mux.HandleFunc("GET /sessions/{id}/events", sseHandler.HandleSSE)

	mux.HandleFunc("POST /sessions/{id}/cancel", h.cancelSession)
	mux.HandleFunc("GET /sessions/{id}/status", h.sessionStatus)

	// Model picker endpoints
	mux.HandleFunc("GET /models", h.listModels)
	mux.HandleFunc("GET /providers", h.getProviders)
	mux.HandleFunc("POST /models", h.setModel)

	// Configuration endpoints
	mux.HandleFunc("GET /config", h.getConfig)

	// Health check - simple handler
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

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

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Server started successfully", "addr", addr)

	// Set up signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-quit
	slog.Info("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
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
