package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/modules"
	"github.com/LaPingvino/recuerdo/internal/modules/system"
)

func main() {
	fmt.Println("Recuerdo Core Test - Starting...")

	// Create a new manager
	manager := core.NewManager()

	// Register some core modules
	fmt.Println("Registering core modules...")

	// Register event module
	// the hand-written core modules the app uses, not the generated
	// skeletons under internal/modules/logic
	eventModule := modules.NewEventModule()
	if err := manager.Register(eventModule); err != nil {
		log.Fatalf("Failed to register event module: %v", err)
	}
	fmt.Println("  ✓ Registered event module")

	// Register settings module
	settingsModule := modules.NewSettingsModule()
	// a throwaway settings file, so testing never touches the user's own
	settingsDir, err := os.MkdirTemp("", "recuerdo-test-core-")
	if err != nil {
		log.Fatalf("Failed to create temporary settings directory: %v", err)
	}
	defer os.RemoveAll(settingsDir)
	if err := settingsModule.SetSettingsPath(filepath.Join(settingsDir, "settings.json")); err != nil {
		log.Fatalf("Failed to set settings path: %v", err)
	}
	if err := manager.Register(settingsModule); err != nil {
		log.Fatalf("Failed to register settings module: %v", err)
	}
	fmt.Println("  ✓ Registered settings module")

	// Register systeminfo module
	systeminfoModule := system.NewSystemInfoModule()
	if err := manager.Register(systeminfoModule); err != nil {
		log.Fatalf("Failed to register systeminfo module: %v", err)
	}
	fmt.Println("  ✓ Registered systeminfo module")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Enable modules
	fmt.Println("Enabling modules...")
	if err := manager.EnableAll(ctx); err != nil {
		log.Fatalf("Failed to enable modules: %v", err)
	}
	fmt.Println("  ✓ All modules enabled successfully")

	// Test event system
	fmt.Println("Testing event system...")
	eventMod, exists := manager.GetDefaultModule("event")
	if !exists {
		log.Fatal("Event module not found")
	}

	received := make(chan interface{}, 1)
	ev := eventMod.(*modules.EventModule).CreateEvent("test-core")
	if err := ev.Subscribe(func(data interface{}) error { received <- data; return nil }); err != nil {
		log.Fatalf("Failed to subscribe to event: %v", err)
	}
	if err := ev.Trigger("ping"); err != nil {
		log.Fatalf("Failed to trigger event: %v", err)
	}
	if got := <-received; got != "ping" {
		log.Fatalf("Event handler received %v, want ping", got)
	}
	fmt.Println("  ✓ Event triggered and handled")

	// Test settings system
	fmt.Println("Testing settings system...")
	settingsMod, exists := manager.GetDefaultModule("settings")
	if !exists {
		log.Fatal("Settings module not found")
	}
	sm := settingsMod.(*modules.SettingsModule)
	if err := sm.SetSetting("test-core.key", "value"); err != nil {
		log.Fatalf("Failed to set setting: %v", err)
	}
	if err := sm.SaveSettings(); err != nil {
		log.Fatalf("Failed to save settings: %v", err)
	}
	if err := sm.LoadSettings(); err != nil {
		log.Fatalf("Failed to load settings: %v", err)
	}
	if v, err := sm.GetSetting("test-core.key"); err != nil || v != "value" {
		log.Fatalf("Setting read back as %v (%v), want value", v, err)
	}
	fmt.Println("  ✓ Setting saved and read back")

	// Show module statistics
	fmt.Printf("Module Statistics:\n")
	fmt.Printf("  Total registered: %d\n", manager.ModuleCount())

	fmt.Println("\n🎉 SUCCESS: Recuerdo core system is working!")
	fmt.Println("   - Module registration: ✓")
	fmt.Println("   - Module enabling: ✓")
	fmt.Println("   - Module discovery: ✓")
	fmt.Println("   - Dependency resolution: ✓")
	fmt.Println("   - Event system: ✓")
	fmt.Println("   - Settings system: ✓")
	fmt.Println("   - System info system: ✓")

	// Disable modules cleanly
	fmt.Println("Shutting down...")
	if err := manager.DisableAll(ctx); err != nil {
		log.Printf("Warning: Failed to disable some modules: %v", err)
	}
	fmt.Println("  ✓ Clean shutdown complete")
}
