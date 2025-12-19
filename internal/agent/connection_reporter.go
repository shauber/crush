package agent

import (
	"context"
	"fmt"

	"github.com/charmbracelet/crush/internal/csync"
)

// ConnectionStatus tracks the health of various providers
const (
	ProviderStatusHealthy   = "healthy"
	ProviderStatusUnhealthy = "unhealthy"
	ProviderStatusUnknown   = "unknown"
)

// ConnectionReporter provides real-time connection status updates
type ConnectionReporter struct {
	status *csync.Map[string, string]
	agent  SessionAgent
}

// NewConnectionReporter creates a new reporter
func NewConnectionReporter(agent SessionAgent) *ConnectionReporter {
	return &ConnectionReporter{
		status: csync.NewMap[string, string](),
		agent:  agent,
	}
}

// GetProviderStatus returns the current status of a provider
func (r *ConnectionReporter) GetProviderStatus(providerID string) string {
	status, ok := r.status.Get(providerID)
	if !ok {
		return ProviderStatusUnknown
	}
	return status
}

// UpdateProviderStatus sets the status for a provider
func (r *ConnectionReporter) UpdateProviderStatus(providerID, status string) {
	r.status.Set(providerID, status)
}

// ValidateAndReport checks all providers and reports their status
func (r *ConnectionReporter) ValidateAndReport(ctx context.Context) map[string]string {
	report := make(map[string]string)
	
	// If agent has health checker, use it
	sa, ok := r.agent.(*sessionAgent)
	if ok && sa.healthChecker != nil {
		results := sa.healthChecker.ValidateProviders()
		for provider, err := range results {
			if err != nil {
				r.status.Set(provider, ProviderStatusUnhealthy)
				report[provider] = fmt.Sprintf("unhealthy: %v", err)
			} else {
				r.status.Set(provider, ProviderStatusHealthy)
				report[provider] = ProviderStatusHealthy
			}
		}
	}
	
	return report
}