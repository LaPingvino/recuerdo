package modules

import (
	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/modules/data/profiledescriptions"
	"github.com/LaPingvino/recuerdo/internal/modules/profilerunners/backgroundimage"
	"github.com/LaPingvino/recuerdo/internal/modules/system"
)

// ModuleRegistry holds all available module initializers
type ModuleRegistry struct {
	initializers map[string]func() core.Module
}

// NewModuleRegistry creates a new module registry with all available modules
func NewModuleRegistry() *ModuleRegistry {
	registry := &ModuleRegistry{
		initializers: make(map[string]func() core.Module),
	}

	// Register core modules
	registry.RegisterModule("execute", func() core.Module {
		return NewExecuteModule()
	})

	// Register business card related modules
	// Temporarily disable business card generator to avoid UI dependency issues
	// registry.RegisterModule("businessCardGenerator", businesscard.Init)
	registry.RegisterModule("backgroundImageGenerator", backgroundimage.Init)
	registry.RegisterModule("profileDescription-generateBusinessCard", profiledescriptions.Init)

	// Register system modules
	registry.RegisterModule("systeminfo", system.InitSystemInfoModule)

	// TODO: Register other modules as they are ported from Python
	// registry.RegisterModule("metadata", metadata.Init)
	// registry.RegisterModule("ui", ui.Init)
	// registry.RegisterModule("event", event.Init)
	// registry.RegisterModule("settings", settings.Init)

	return registry
}

// RegisterModule registers a module initializer function
func (r *ModuleRegistry) RegisterModule(moduleType string, initializer func() core.Module) {
	r.initializers[moduleType] = initializer
}

// GetInitializer returns the initializer function for a module type
func (r *ModuleRegistry) GetInitializer(moduleType string) (func() core.Module, bool) {
	initializer, exists := r.initializers[moduleType]
	return initializer, exists
}

// ListModuleTypes returns all registered module types
func (r *ModuleRegistry) ListModuleTypes() []string {
	types := make([]string, 0, len(r.initializers))
	for moduleType := range r.initializers {
		types = append(types, moduleType)
	}
	return types
}

// CreateModule creates a new instance of the specified module type
func (r *ModuleRegistry) CreateModule(moduleType string) (core.Module, bool) {
	initializer, exists := r.GetInitializer(moduleType)
	if !exists {
		return nil, false
	}

	module := initializer()
	return module, true
}
