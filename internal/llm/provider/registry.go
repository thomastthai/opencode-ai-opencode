package provider

import (
	"fmt"
	"sync"

	"github.com/opencode-ai/opencode/internal/llm/models"
)

// ProviderFactory is a function that creates a new provider instance with the given configuration.
type ProviderFactory func(config ProviderConfig) (Provider, error)

// ProviderInfo contains metadata about a registered provider.
type ProviderInfo struct {
	Name        models.ModelProvider `json:"name"`
	Description string               `json:"description"`
	Capabilities []string            `json:"capabilities"`
}

// ProviderRegistry manages the registration and creation of providers.
// It is thread-safe and supports both compile-time and runtime registration.
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[models.ModelProvider]ProviderFactory
	info      map[models.ModelProvider]ProviderInfo
}

// Global registry instance
var globalRegistry = &ProviderRegistry{
	providers: make(map[models.ModelProvider]ProviderFactory),
	info:      make(map[models.ModelProvider]ProviderInfo),
}

// RegisterProvider registers a new provider factory in the global registry.
// This function is typically called from provider init() functions.
func RegisterProvider(providerType models.ModelProvider, factory ProviderFactory, info ProviderInfo) error {
	return globalRegistry.RegisterProvider(providerType, factory, info)
}

// NewProvider creates a new provider instance using the global registry.
func NewProvider(providerType models.ModelProvider, config ProviderConfig) (Provider, error) {
	return globalRegistry.NewProvider(providerType, config)
}

// ListRegisteredProviders returns a list of all registered providers from the global registry.
func ListRegisteredProviders() []ProviderInfo {
	return globalRegistry.ListRegisteredProviders()
}

// GetProviderInfo returns information about a specific provider from the global registry.
func GetProviderInfo(providerType models.ModelProvider) (ProviderInfo, error) {
	return globalRegistry.GetProviderInfo(providerType)
}

// RegisterProvider registers a new provider factory in this registry.
func (r *ProviderRegistry) RegisterProvider(providerType models.ModelProvider, factory ProviderFactory, info ProviderInfo) error {
	if factory == nil {
		return fmt.Errorf("provider factory cannot be nil for provider '%s'", providerType)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.providers[providerType]; exists {
		return fmt.Errorf("provider '%s' is already registered", providerType)
	}
	
	r.providers[providerType] = factory
	r.info[providerType] = info
	
	return nil
}

// NewProvider creates a new provider instance of the specified type.
func (r *ProviderRegistry) NewProvider(providerType models.ModelProvider, config ProviderConfig) (Provider, error) {
	r.mu.RLock()
	factory, exists := r.providers[providerType]
	r.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("provider '%s' is not registered. Available providers: %v", 
			providerType, r.getProviderNames())
	}
	
	provider, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider '%s': %w", providerType, err)
	}
	
	return provider, nil
}

// ListRegisteredProviders returns a list of all registered providers.
func (r *ProviderRegistry) ListRegisteredProviders() []ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	providers := make([]ProviderInfo, 0, len(r.info))
	for _, info := range r.info {
		providers = append(providers, info)
	}
	
	return providers
}

// GetProviderInfo returns information about a specific provider.
func (r *ProviderRegistry) GetProviderInfo(providerType models.ModelProvider) (ProviderInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	info, exists := r.info[providerType]
	if !exists {
		return ProviderInfo{}, fmt.Errorf("provider '%s' is not registered", providerType)
	}
	
	return info, nil
}

// IsRegistered checks if a provider is registered.
func (r *ProviderRegistry) IsRegistered(providerType models.ModelProvider) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.providers[providerType]
	return exists
}

// UnregisterProvider removes a provider from the registry.
// This is primarily useful for testing scenarios.
func (r *ProviderRegistry) UnregisterProvider(providerType models.ModelProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.providers, providerType)
	delete(r.info, providerType)
}

// Clear removes all providers from the registry.
// This is primarily useful for testing scenarios.
func (r *ProviderRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.providers = make(map[models.ModelProvider]ProviderFactory)
	r.info = make(map[models.ModelProvider]ProviderInfo)
}

// getProviderNames returns a list of registered provider names (not thread-safe, caller must hold lock).
func (r *ProviderRegistry) getProviderNames() []models.ModelProvider {
	names := make([]models.ModelProvider, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// NewTestRegistry creates a new isolated registry for testing purposes.
func NewTestRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[models.ModelProvider]ProviderFactory),
		info:      make(map[models.ModelProvider]ProviderInfo),
	}
}