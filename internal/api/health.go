package api

import (
	"encoding/json"
	"net/http"

	"github.com/charmbracelet/crush/internal/agent"
)

// HealthHandler provides health check endpoints
func HealthHandler(agent agent.SessionAgent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Create a reporter to check provider health
		reporter := agent.NewConnectionReporter(agent)
		ctx := r.Context()
		
		// Validate providers and get status
		status := reporter.ValidateAndReport(ctx)
		
		response := map[string]interface{}{
			"status":   "ok",
			"providers": status,
		}
		
		if len(status) == 0 || allHealthy(status) {
			response["status"] = "ok"
			w.WriteHeader(http.StatusOK)
		} else {
			response["status"] = "degraded"
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		
		json.NewEncoder(w).Encode(response)
	}
}

func allHealthy(status map[string]string) bool {
	for _, s := range status {
		if s != "healthy" {
			return false
		}
	}
	return true
}