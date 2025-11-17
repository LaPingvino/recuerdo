package core

import (
	"context"
	"fmt"
)

// Module represents the basic interface that all OpenTeacher modules must implement.
// This mirrors the Python module system where each module defines its type,
// dependencies, and lifecycle methods.
type Module interface {
	// Type returns the module type identifier (e.g., "execute", "gui", "loader")
	Type() string

	// Name returns a human-readable name for the module
	Name() string

	// Requires returns a list of module types that this module depends on.
	// These dependencies must be loaded and enabled before this module.
	Requires() []string

	// Uses returns a list of module types that this module can optionally use.
	// These are soft dependencies - the module works without them but gains
	// functionality when they're available.
	Uses() []string

	// Priority returns the loading priority for this module type.
	// Higher numbers load first. Default is 0.
	Priority() int

	// Enable activates the module. Called after all dependencies are loaded.
	Enable(ctx context.Context) error

	// Disable deactivates the module. Called during shutdown.
	Disable(ctx context.Context) error

	// IsActive returns true if the module is currently enabled
	IsActive() bool
}

// EventModule represents modules that provide event handling capabilities
type EventModule interface {
	Module

	// CreateEvent creates a new event that other modules can listen to
	CreateEvent(name string) Event

	// Subscribe allows modules to listen for specific events
	Subscribe(eventName string, handler EventHandler) error

	// Unsubscribe removes an event handler
	Unsubscribe(eventName string, handler EventHandler) error
}

// Event represents an event that can be triggered and handled
type Event interface {
	// Name returns the event name
	Name() string

	// Trigger sends the event to all registered handlers
	Trigger(data interface{}) error

	// Subscribe adds a handler for this event
	Subscribe(handler EventHandler) error

	// Unsubscribe removes a handler from this event
	Unsubscribe(handler EventHandler) error
}

// EventHandler is a function that handles events
type EventHandler func(data interface{}) error

// ExecuteModule represents modules that control application execution
type ExecuteModule interface {
	Module

	// StartRunning begins the main application loop
	StartRunning(ctx context.Context) error

	// SetProfile sets the execution profile (e.g., "all", "cli", "gui")
	SetProfile(profile string) error

	// GetProfile returns the current execution profile
	GetProfile() string
}

// ResourceProvider provides access to module resources
type ResourceProvider interface {
	// ResourcePath returns the full path to a resource file relative to the module
	ResourcePath(resource string) (string, error)

	// ModulePath returns the base path of the module
	ModulePath() string
}

// SettingsModule represents modules that handle configuration
type SettingsModule interface {
	Module

	// GetSetting retrieves a configuration value
	GetSetting(key string) (interface{}, error)

	// SetSetting stores a configuration value
	SetSetting(key string, value interface{}) error

	// LoadSettings loads settings from storage
	LoadSettings() error

	// SaveSettings persists settings to storage
	SaveSettings() error
}

// ModuleError represents errors that occur during module operations
type ModuleError struct {
	ModuleType string
	ModuleName string
	Operation  string
	Err        error
}

func (e *ModuleError) Error() string {
	return fmt.Sprintf("module %s (%s) failed during %s: %v",
		e.ModuleName, e.ModuleType, e.Operation, e.Err)
}

func (e *ModuleError) Unwrap() error {
	return e.Err
}

// NewModuleError creates a new module error
func NewModuleError(moduleType, moduleName, operation string, err error) *ModuleError {
	return &ModuleError{
		ModuleType: moduleType,
		ModuleName: moduleName,
		Operation:  operation,
		Err:        err,
	}
}

// BaseModule provides a basic implementation that other modules can embed
type BaseModule struct {
	moduleType string
	moduleName string
	requires   []string
	uses       []string
	priority   int
	active     bool
	manager    *Manager
}

// NewBaseModule creates a new base module
func NewBaseModule(moduleType, moduleName string) *BaseModule {
	return &BaseModule{
		moduleType: moduleType,
		moduleName: moduleName,
		requires:   make([]string, 0),
		uses:       make([]string, 0),
		priority:   0,
		active:     false,
	}
}

// Type returns the module type
func (b *BaseModule) Type() string {
	return b.moduleType
}

// Name returns the module name
func (b *BaseModule) Name() string {
	return b.moduleName
}

// GetType returns the module type (alias for Type for test compatibility)
func (b *BaseModule) GetType() string {
	return b.moduleType
}

// GetName returns the module name (alias for Name for test compatibility)
func (b *BaseModule) GetName() string {
	return b.moduleName
}

// GetRequires returns required dependencies (alias for Requires for test compatibility)
func (b *BaseModule) GetRequires() []string {
	return append([]string(nil), b.requires...)
}

// Requires returns required dependencies
func (b *BaseModule) Requires() []string {
	return append([]string(nil), b.requires...)
}

// Uses returns optional dependencies
func (b *BaseModule) Uses() []string {
	return append([]string(nil), b.uses...)
}

// Priority returns the loading priority
func (b *BaseModule) Priority() int {
	return b.priority
}

// IsActive returns the activation status
func (b *BaseModule) IsActive() bool {
	return b.active
}

// SetRequires sets the required dependencies
func (b *BaseModule) SetRequires(requires ...string) {
	b.requires = append([]string(nil), requires...)
}

// SetUses sets the optional dependencies
func (b *BaseModule) SetUses(uses ...string) {
	b.uses = append([]string(nil), uses...)
}

// SetPriority sets the loading priority
func (b *BaseModule) SetPriority(priority int) {
	b.priority = priority
}

// SetActive sets the activation status
func (b *BaseModule) SetActive(active bool) {
	b.active = active
}

// Enable provides a default enable implementation
func (b *BaseModule) Enable(ctx context.Context) error {
	b.active = true
	return nil
}

// Disable provides a default disable implementation
func (b *BaseModule) Disable(ctx context.Context) error {
	b.active = false
	return nil
}

// SetManager sets the module manager for this module
func (b *BaseModule) SetManager(manager *Manager) {
	b.manager = manager
}

// GetManager returns the module manager
func (b *BaseModule) GetManager() *Manager {
	return b.manager
}
