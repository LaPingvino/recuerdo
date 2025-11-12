package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/LaPingvino/openteacher/internal/core"
)

// ExecuteModule controls the main application lifecycle and execution flow.
// It mirrors the Python execute module functionality.
type ExecuteModule struct {
	*core.BaseModule
	manager     *core.Manager
	profile     string
	running     bool
	startEvent  core.Event
	stopEvent   core.Event
	eventModule core.EventModule
}

// NewExecuteModule creates a new execute module
func NewExecuteModule() *ExecuteModule {
	base := core.NewBaseModule("execute", "execute-module")
	base.SetRequires("event")
	base.SetPriority(1000) // High priority - needs to start first

	return &ExecuteModule{
		BaseModule: base,
		profile:    "all", // Default profile
		running:    false,
	}
}

// Enable initializes the execute module
func (e *ExecuteModule) Enable(ctx context.Context) error {
	if err := e.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// TODO: Get event module from manager when we have dependency injection
	// For now, we'll create events manually when the event system is ready

	return nil
}

// Disable shuts down the execute module
func (e *ExecuteModule) Disable(ctx context.Context) error {
	e.running = false
	return e.BaseModule.Disable(ctx)
}

// StartRunning begins the main application loop
func (e *ExecuteModule) StartRunning(ctx context.Context) error {
	if e.running {
		return fmt.Errorf("execute module is already running")
	}

	e.running = true
	defer func() {
		e.running = false
	}()

	fmt.Printf("Execute module starting with profile: %s\n", e.profile)

	// Trigger start event if available
	if e.startEvent != nil {
		if err := e.startEvent.Trigger(e.profile); err != nil {
			fmt.Printf("Warning: failed to trigger start event: %v\n", err)
		}
	}

	// Main application loop
	fmt.Println("OpenTeacher is now running...")
	fmt.Println("Press Ctrl+C to exit")

	// Simple implementation: just wait for context cancellation
	// In a real application, this would start the GUI or other main functionality
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Execute module received shutdown signal")

			// Trigger stop event if available
			if e.stopEvent != nil {
				if err := e.stopEvent.Trigger(nil); err != nil {
					fmt.Printf("Warning: failed to trigger stop event: %v\n", err)
				}
			}

			return ctx.Err()

		case <-ticker.C:
			fmt.Printf("OpenTeacher heartbeat - profile: %s, active: %t\n", e.profile, e.running)
		}
	}
}

// SetProfile sets the execution profile
func (e *ExecuteModule) SetProfile(profile string) error {
	if profile == "" {
		return fmt.Errorf("profile cannot be empty")
	}

	oldProfile := e.profile
	e.profile = profile

	fmt.Printf("Profile changed from %s to %s\n", oldProfile, profile)
	return nil
}

// GetProfile returns the current execution profile
func (e *ExecuteModule) GetProfile() string {
	return e.profile
}

// IsRunning returns true if the module is currently running
func (e *ExecuteModule) IsRunning() bool {
	return e.running
}

// SetEventModule sets the event module for creating events
func (e *ExecuteModule) SetEventModule(eventModule core.EventModule) {
	e.eventModule = eventModule

	// Create start/stop events
	e.startEvent = eventModule.CreateEvent("execute.start")
	e.stopEvent = eventModule.CreateEvent("execute.stop")
}
