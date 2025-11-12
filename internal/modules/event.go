package modules

import (
	"context"
	"fmt"
	"sync"

	"github.com/LaPingvino/openteacher/internal/core"
)

// Event represents a concrete implementation of the Event interface
type Event struct {
	name     string
	handlers []core.EventHandler
	mu       sync.RWMutex
}

// NewEvent creates a new event
func NewEvent(name string) *Event {
	return &Event{
		name:     name,
		handlers: make([]core.EventHandler, 0),
	}
}

// Name returns the event name
func (e *Event) Name() string {
	return e.name
}

// Trigger sends the event to all registered handlers
func (e *Event) Trigger(data interface{}) error {
	e.mu.RLock()
	handlers := make([]core.EventHandler, len(e.handlers))
	copy(handlers, e.handlers)
	e.mu.RUnlock()

	// Call all handlers
	for _, handler := range handlers {
		if err := handler(data); err != nil {
			return fmt.Errorf("event handler failed for event %s: %w", e.name, err)
		}
	}

	return nil
}

// Subscribe adds a handler for this event
func (e *Event) Subscribe(handler core.EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.handlers = append(e.handlers, handler)
	return nil
}

// Unsubscribe removes a handler from this event
func (e *Event) Unsubscribe(handler core.EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Find and remove the handler
	// Note: This is a simple implementation that removes the first matching handler
	// In practice, you might want to use a more sophisticated approach
	for i, h := range e.handlers {
		// Compare function pointers (this is a limitation - in real code you'd want handler IDs)
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			e.handlers = append(e.handlers[:i], e.handlers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("handler not found for event %s", e.name)
}

// EventModule provides event handling capabilities to other modules
type EventModule struct {
	*core.BaseModule
	events map[string]core.Event
	mu     sync.RWMutex
}

// NewEventModule creates a new event module
func NewEventModule() *EventModule {
	base := core.NewBaseModule("event", "event-module")
	base.SetPriority(2000) // Very high priority - other modules depend on events

	return &EventModule{
		BaseModule: base,
		events:     make(map[string]core.Event),
	}
}

// Enable initializes the event module
func (e *EventModule) Enable(ctx context.Context) error {
	if err := e.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("Event module enabled - ready to handle events")
	return nil
}

// Disable shuts down the event module
func (e *EventModule) Disable(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear all events
	e.events = make(map[string]core.Event)

	fmt.Println("Event module disabled - all events cleared")
	return e.BaseModule.Disable(ctx)
}

// CreateEvent creates a new event that other modules can listen to
func (e *EventModule) CreateEvent(name string) core.Event {
	if name == "" {
		panic("event name cannot be empty")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if event already exists
	if existingEvent, exists := e.events[name]; exists {
		return existingEvent
	}

	// Create new event
	event := NewEvent(name)
	e.events[name] = event

	fmt.Printf("Created event: %s\n", name)
	return event
}

// Subscribe allows modules to listen for specific events
func (e *EventModule) Subscribe(eventName string, handler core.EventHandler) error {
	if eventName == "" {
		return fmt.Errorf("event name cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	e.mu.RLock()
	event, exists := e.events[eventName]
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("event %s does not exist", eventName)
	}

	return event.Subscribe(handler)
}

// Unsubscribe removes an event handler
func (e *EventModule) Unsubscribe(eventName string, handler core.EventHandler) error {
	if eventName == "" {
		return fmt.Errorf("event name cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	e.mu.RLock()
	event, exists := e.events[eventName]
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("event %s does not exist", eventName)
	}

	return event.Unsubscribe(handler)
}

// GetEvent returns an existing event by name
func (e *EventModule) GetEvent(name string) (core.Event, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	event, exists := e.events[name]
	return event, exists
}

// ListEvents returns all event names
func (e *EventModule) ListEvents() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	names := make([]string, 0, len(e.events))
	for name := range e.events {
		names = append(names, name)
	}

	return names
}

// EventCount returns the number of registered events
func (e *EventModule) EventCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.events)
}
