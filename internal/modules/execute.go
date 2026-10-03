package modules

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LaPingvino/recuerdo/internal/core"
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

	log.Printf("[SUCCESS] ExecuteModule enabled - ready to handle application lifecycle")
	return nil
}

// Disable shuts down the execute module
func (e *ExecuteModule) Disable(ctx context.Context) error {
	e.running = false
	log.Printf("[SUCCESS] ExecuteModule disabled - application lifecycle handling stopped")
	return e.BaseModule.Disable(ctx)
}

// StartRunning begins the main application loop
func (e *ExecuteModule) StartRunning(ctx context.Context) error {
	if e.running {
		log.Printf("[ERROR] ExecuteModule.StartRunning() failed - module is already running")
		return fmt.Errorf("execute module is already running")
	}

	e.running = true
	defer func() {
		e.running = false
		log.Printf("[SUCCESS] ExecuteModule.StartRunning() finished - execution stopped")
	}()

	log.Printf("[SUCCESS] ExecuteModule.StartRunning() - starting with profile: %s", e.profile)
	fmt.Printf("Execute module starting with profile: %s\n", e.profile)

	// Trigger start event if available
	if e.startEvent != nil {
		log.Printf("[SUCCESS] ExecuteModule triggering start event for profile: %s", e.profile)
		if err := e.startEvent.Trigger(e.profile); err != nil {
			log.Printf("[ERROR] ExecuteModule failed to trigger start event: %v", err)
			fmt.Printf("Warning: failed to trigger start event: %v\n", err)
		}
	} else {
		log.Printf("[STUB] ExecuteModule.StartRunning() - start event not available")
	}

	// Main application loop
	fmt.Println("Recuerdo is now running...")
	fmt.Println("Press Ctrl+C to exit")

	// Simple implementation: just wait for context cancellation
	// In a real application, this would start the GUI or other main functionality
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[SUCCESS] ExecuteModule received shutdown signal via context cancellation")
			fmt.Println("Execute module received shutdown signal")

			// Trigger stop event if available
			if e.stopEvent != nil {
				log.Printf("[SUCCESS] ExecuteModule triggering stop event")
				if err := e.stopEvent.Trigger(nil); err != nil {
					log.Printf("[ERROR] ExecuteModule failed to trigger stop event: %v", err)
					fmt.Printf("Warning: failed to trigger stop event: %v\n", err)
				}
			} else {
				log.Printf("[STUB] ExecuteModule.StartRunning() - stop event not available")
			}

			return ctx.Err()

		case <-ticker.C:
			fmt.Printf("Recuerdo heartbeat - profile: %s, active: %t\n", e.profile, e.running)
		}
	}
}

// SetProfile sets the execution profile
func (e *ExecuteModule) SetProfile(profile string) error {
	if profile == "" {
		log.Printf("[ERROR] ExecuteModule.SetProfile() failed - empty profile not allowed")
		return fmt.Errorf("profile cannot be empty")
	}

	oldProfile := e.profile
	e.profile = profile

	log.Printf("[SUCCESS] ExecuteModule.SetProfile() changed from '%s' to '%s'", oldProfile, profile)
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
	log.Printf("[SUCCESS] ExecuteModule.SetEventModule() - creating start/stop events")
	e.startEvent = eventModule.CreateEvent("execute.start")
	e.stopEvent = eventModule.CreateEvent("execute.stop")
	log.Printf("[SUCCESS] ExecuteModule.SetEventModule() - start/stop events created")
}
