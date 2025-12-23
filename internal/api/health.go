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
		
		response := map[string]interface{}{
			"status": "ok",
		}
		
		w.WriteHeader(http.StatusOK)
		
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