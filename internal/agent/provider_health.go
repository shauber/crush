package agent

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/charmbracelet/crush/internal/config"
)

// ProviderHealthChecker validates provider connectivity and manages health status
type ProviderHealthChecker struct {
	client    *http.Client
	providers *config.Config
}

// NewProviderHealthChecker creates a new health checker
func NewProviderHealthChecker(providers *config.Config) *ProviderHealthChecker {
	return &ProviderHealthChecker{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		providers: providers,
	}
}

// CheckProviderConnectivity validates that we can connect to the specified provider
func (h *ProviderHealthChecker) CheckProviderConnectivity(ctx context.Context, providerID string) error {
	providerConfig, exists := h.providers.Providers.Get(providerID)
	if !exists {
		return fmt.Errorf("provider %s not found", providerID)
	}
	
	// Skip disabled providers
	if providerConfig.Disable {
		return fmt.Errorf("provider %s is disabled", providerID)
	}
	
	return providerConfig.TestConnection(h.providers.Resolver())
}

// GetHealthyModels returns a list of models from providers that are actually reachable
func (h *ProviderHealthChecker) GetHealthyModels(modelType config.SelectedModelType) ([]config.SelectedModel, error) {
	modelCfg, exists := h.providers.Models[modelType]
	if !exists {
		return nil, fmt.Errorf("no model configuration for type %s", modelType)
	}
	
	// Check if the provider for this model is healthy
	err := h.CheckProviderConnectivity(context.Background(), modelCfg.Provider)
	if err != nil {
		slog.Error("provider connectivity check failed", "provider", modelCfg.Provider, "error", err)
		return nil, fmt.Errorf("provider %s unreachable: %w", modelCfg.Provider, err)
	}
	
	return []config.SelectedModel{modelCfg}, nil
}

// ValidateProviders checks all configured providers and returns healthy ones
func (h *ProviderHealthChecker) ValidateProviders() map[string]error {
	results := make(map[string]error)
	
	for providerConfig := range h.providers.Providers.Seq() {
		if providerConfig.Disable {
			continue
		}
		
		err := providerConfig.TestConnection(h.providers.Resolver())
		if err != nil {
			slog.Error("provider validation failed", "provider", providerConfig.ID, "error", err)
			results[providerConfig.ID] = err
		}
	}
	
	return results
}