package core

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// Manager handles module registration, dependency resolution, and lifecycle management
type Manager struct {
	modules       map[string]Module
	modulesByType map[string][]Module
	loadOrder     []string
	resourcePaths map[string]string
	mu            sync.RWMutex
	enabled       bool
}

// NewManager creates a new module manager
func NewManager() *Manager {
	return &Manager{
		modules:       make(map[string]Module),
		modulesByType: make(map[string][]Module),
		loadOrder:     make([]string, 0),
		resourcePaths: make(map[string]string),
		enabled:       false,
	}
}

// Register registers a module with the manager
func (m *Manager) Register(module Module) error {
	if module == nil {
		return fmt.Errorf("cannot register nil module")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	name := module.Name()
	if name == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	// Check for duplicate names
	if _, exists := m.modules[name]; exists {
		return fmt.Errorf("module with name %q already registered", name)
	}

	// Register the module
	m.modules[name] = module

	// Add to type mapping
	moduleType := module.Type()
	if moduleType == "" {
		return fmt.Errorf("module type cannot be empty for module %q", name)
	}

	m.modulesByType[moduleType] = append(m.modulesByType[moduleType], module)

	return nil
}

// Unregister removes a module from the manager
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	module, exists := m.modules[name]
	if !exists {
		return fmt.Errorf("module %q not found", name)
	}

	// Disable if active
	if module.IsActive() {
		if err := module.Disable(context.Background()); err != nil {
			return fmt.Errorf("failed to disable module %q: %w", name, err)
		}
	}

	// Remove from type mapping
	moduleType := module.Type()
	typeModules := m.modulesByType[moduleType]
	for i, mod := range typeModules {
		if mod.Name() == name {
			m.modulesByType[moduleType] = append(typeModules[:i], typeModules[i+1:]...)
			break
		}
	}

	// Clean up empty type entries
	if len(m.modulesByType[moduleType]) == 0 {
		delete(m.modulesByType, moduleType)
	}

	// Remove from main registry
	delete(m.modules, name)

	// Remove from load order
	for i, modName := range m.loadOrder {
		if modName == name {
			m.loadOrder = append(m.loadOrder[:i], m.loadOrder[i+1:]...)
			break
		}
	}

	return nil
}

// GetModule returns a module by name
func (m *Manager) GetModule(name string) (Module, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	module, exists := m.modules[name]
	return module, exists
}

// GetModulesByType returns all modules of a specific type
func (m *Manager) GetModulesByType(moduleType string) []Module {
	m.mu.RLock()
	defer m.mu.RUnlock()

	modules := m.modulesByType[moduleType]
	result := make([]Module, len(modules))
	copy(result, modules)
	return result
}

// GetDefaultModule returns the highest priority module of a specific type
func (m *Manager) GetDefaultModule(moduleType string) (Module, bool) {
	modules := m.GetModulesByType(moduleType)
	if len(modules) == 0 {
		return nil, false
	}

	// Sort by priority (higher first)
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Priority() > modules[j].Priority()
	})

	return modules[0], true
}

// ListModules returns all registered module names
func (m *Manager) ListModules() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.modules))
	for name := range m.modules {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListTypes returns all registered module types
func (m *Manager) ListTypes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]string, 0, len(m.modulesByType))
	for moduleType := range m.modulesByType {
		types = append(types, moduleType)
	}
	sort.Strings(types)
	return types
}

// ResolveLoadOrder computes the correct order to load modules based on dependencies
func (m *Manager) ResolveLoadOrder() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Build dependency graph
	graph := make(map[string][]string) // module -> dependencies
	inDegree := make(map[string]int)   // module -> number of dependencies

	// Initialize graph
	for name := range m.modules {
		graph[name] = make([]string, 0)
		inDegree[name] = 0
	}

	// Build edges based on module requirements
	for name, module := range m.modules {
		requires := module.Requires()

		for _, requiredType := range requires {
			// Find modules of required type
			requiredModules := m.modulesByType[requiredType]
			if len(requiredModules) == 0 {
				return nil, fmt.Errorf("module %q requires type %q but no modules of that type are registered", name, requiredType)
			}

			// For simplicity, depend on the default (highest priority) module of that type
			sort.Slice(requiredModules, func(i, j int) bool {
				return requiredModules[i].Priority() > requiredModules[j].Priority()
			})

			requiredName := requiredModules[0].Name()
			graph[requiredName] = append(graph[requiredName], name)
			inDegree[name]++
		}
	}

	// Topological sort using Kahn's algorithm
	var result []string
	var queue []string

	// Start with nodes that have no dependencies
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Sort by priority within each level
	for len(queue) > 0 {
		// Sort current level by priority
		sort.Slice(queue, func(i, j int) bool {
			return m.modules[queue[i]].Priority() > m.modules[queue[j]].Priority()
		})

		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Process dependent modules
		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for circular dependencies
	if len(result) != len(m.modules) {
		return nil, fmt.Errorf("circular dependency detected in module graph")
	}

	return result, nil
}

// EnableAll enables all registered modules in dependency order
func (m *Manager) EnableAll(ctx context.Context) error {
	loadOrder, err := m.ResolveLoadOrder()
	if err != nil {
		return fmt.Errorf("failed to resolve module load order: %w", err)
	}

	m.mu.Lock()
	m.loadOrder = loadOrder
	m.mu.Unlock()

	// Enable modules in order
	for _, name := range loadOrder {
		module, exists := m.GetModule(name)
		if !exists {
			return fmt.Errorf("module %q not found during enable", name)
		}

		if err := module.Enable(ctx); err != nil {
			return NewModuleError(module.Type(), name, "enable", err)
		}
	}

	m.mu.Lock()
	m.enabled = true
	m.mu.Unlock()

	return nil
}

// DisableAll disables all modules in reverse dependency order
func (m *Manager) DisableAll(ctx context.Context) error {
	m.mu.Lock()
	if !m.enabled {
		m.mu.Unlock()
		return nil
	}

	// Disable in reverse order
	loadOrder := make([]string, len(m.loadOrder))
	copy(loadOrder, m.loadOrder)
	m.mu.Unlock()

	// Reverse the order
	for i := len(loadOrder) - 1; i >= 0; i-- {
		name := loadOrder[i]
		module, exists := m.GetModule(name)
		if !exists {
			continue // Module may have been unregistered
		}

		if module.IsActive() {
			if err := module.Disable(ctx); err != nil {
				// Log error but continue disabling other modules
				fmt.Printf("Warning: failed to disable module %q: %v\n", name, err)
			}
		}
	}

	m.mu.Lock()
	m.enabled = false
	m.mu.Unlock()

	return nil
}

// IsEnabled returns true if all modules have been enabled
func (m *Manager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// SetResourcePath sets the resource path for a module
func (m *Manager) SetResourcePath(moduleName, path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resourcePaths[moduleName] = path
}

// GetResourcePath returns the resource path for a module
func (m *Manager) GetResourcePath(moduleName string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	path, exists := m.resourcePaths[moduleName]
	return path, exists
}

// ModuleCount returns the number of registered modules
func (m *Manager) ModuleCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.modules)
}

// TypeCount returns the number of registered module types
func (m *Manager) TypeCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.modulesByType)
}
