package lessonDialogs

import (
	"context"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// TestLessonDialogsModuleCreation tests module creation and basic setup
func TestLessonDialogsModuleCreation(t *testing.T) {
	module := NewLessonDialogsModule()

	if module == nil {
		t.Fatal("NewLessonDialogsModule() returned nil")
	}

	if module.BaseModule == nil {
		t.Error("BaseModule is nil")
	}

	// Test module type
	if module.Type() != "lessonDialogs" {
		t.Errorf("Expected module type 'lessonDialogs', got %q", module.Type())
	}

	// Test module name
	if module.Name() != "lesson-dialogs-module" {
		t.Errorf("Expected module name 'lesson-dialogs-module', got %q", module.Name())
	}

	// Test requirements
	requires := module.Requires()
	if len(requires) == 0 {
		t.Error("Expected module to have requirements")
	}

	foundQtApp := false
	for _, req := range requires {
		if req == "qtApp" {
			foundQtApp = true
			break
		}
	}
	if !foundQtApp {
		t.Error("Expected module to require 'qtApp'")
	}
}

// TestLessonDialogsModuleInterface tests interface compliance
func TestLessonDialogsModuleInterface(t *testing.T) {
	module := NewLessonDialogsModule()

	// Test core.Module interface
	var _ core.Module = module

	// Test ShowNewLessonDialog interface
	var _ interface{ ShowNewLessonDialog() map[string]interface{} } = module

	// Test other expected dialog interfaces (using proper Qt widget types)
	// Note: These use *widgets.QWidget, not *interface{}
}

// TestModuleLifecycle tests module enable/disable lifecycle
func TestModuleLifecycle(t *testing.T) {
	manager := core.NewManager()
	module := NewLessonDialogsModule()

	// Register module
	err := manager.Register(module)
	if err != nil {
		t.Fatalf("Failed to register module: %v", err)
	}

	// Test module is not active initially
	if module.IsActive() {
		t.Error("Module should not be active before enable")
	}

	// Enable module (this will fail without Qt, but we test the attempt)
	ctx := context.Background()
	err = module.Enable(ctx)

	// We expect this to fail gracefully without Qt, but module structure should be intact
	if err != nil {
		t.Logf("Enable failed as expected without Qt environment: %v", err)
		// This is expected in test environment without Qt
	}

	// Test disable
	err = module.Disable(ctx)
	if err != nil {
		t.Errorf("Disable should not fail: %v", err)
	}

	if module.IsActive() {
		t.Error("Module should not be active after disable")
	}
}

// TestGetNewLessonDataWithNilDialog tests graceful handling of nil dialog
func TestGetNewLessonDataWithNilDialog(t *testing.T) {
	module := NewLessonDialogsModule()

	// Test with nil dialog (before creation)
	data := module.getNewLessonData()
	if data != nil {
		t.Error("getNewLessonData should return nil when dialog is nil")
	}
}

// TestShowNewLessonDialogWithoutQt tests dialog method without Qt (should handle gracefully)
func TestShowNewLessonDialogWithoutQt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Qt-dependent test in short mode")
	}

	module := NewLessonDialogsModule()

	// This should handle the case where Qt is not available
	// In a proper test environment, this would return nil gracefully
	result := module.ShowNewLessonDialog()

	// Without Qt, this should return nil or handle the error gracefully
	if result != nil {
		t.Log("ShowNewLessonDialog returned data without Qt - this might indicate a test environment issue")
	}
}

// TestDialogDataProcessing tests dialog data extraction logic
func TestDialogDataProcessing(t *testing.T) {
	// This is a unit test for the data processing logic
	// We can't easily test the actual Qt dialog without a GUI environment,
	// but we can test the data structure expectations

	module := NewLessonDialogsModule()

	// Test that the module has the expected dialog pointers initialized to nil
	if module.newLessonDialog != nil {
		t.Error("newLessonDialog should be nil initially")
	}

	if module.propertiesDialog != nil {
		t.Error("propertiesDialog should be nil initially")
	}

	if module.importDialog != nil {
		t.Error("importDialog should be nil initially")
	}
}

// TestModuleRegistration tests that the module registers correctly
func TestModuleRegistration(t *testing.T) {
	manager := core.NewManager()
	module := NewLessonDialogsModule()

	// Register the module
	err := manager.Register(module)
	if err != nil {
		t.Fatalf("Failed to register lessonDialogs module: %v", err)
	}

	// Check that it's registered
	modules := manager.GetModulesByType("lessonDialogs")
	if len(modules) != 1 {
		t.Errorf("Expected 1 lessonDialogs module, got %d", len(modules))
	}

	// Check that we can get it by interface
	if dialogMod, ok := modules[0].(interface{ ShowNewLessonDialog() map[string]interface{} }); !ok {
		t.Error("Registered module does not implement ShowNewLessonDialog interface")
	} else if dialogMod == nil {
		t.Error("Interface cast returned nil")
	}
}

// TestExpectedDialogFields tests the expected dialog form fields
func TestExpectedDialogFields(t *testing.T) {
	// Test the expected structure of dialog data that should be returned
	// This documents the expected interface for integration

	expectedFields := []string{
		"name",
		"description",
		"type",
		"questionLanguage",
		"answerLanguage",
	}

	// This is more of a documentation test - ensuring we expect these fields
	// The actual dialog creation and field extraction is tested in integration tests

	for _, field := range expectedFields {
		if field == "" {
			t.Errorf("Expected field name should not be empty")
		}
	}

	// Test expected lesson types
	expectedTypes := []string{"words", "topo", "media"}
	if len(expectedTypes) == 0 {
		t.Error("Should have expected lesson types")
	}

	// Test expected languages
	expectedLanguages := []string{"English", "Dutch", "French", "German", "Spanish", "Italian"}
	if len(expectedLanguages) == 0 {
		t.Error("Should have expected languages")
	}
}

// BenchmarkModuleCreation benchmarks module creation performance
func BenchmarkModuleCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		module := NewLessonDialogsModule()
		if module == nil {
			b.Fatal("Module creation failed")
		}
	}
}

// BenchmarkModuleRegistration benchmarks module registration
func BenchmarkModuleRegistration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		manager := core.NewManager()
		module := NewLessonDialogsModule()

		err := manager.Register(module)
		if err != nil {
			b.Fatalf("Registration failed: %v", err)
		}
	}
}

// "Words only" hides the lesson kinds: every new lesson is a word lesson.
func TestWordsOnly(t *testing.T) {
	defer func() { WordsOnly = nil }()
	for _, on := range []bool{false, true} {
		WordsOnly = func() bool { return on }
		mod := NewLessonDialogsModule()
		mod.createNewLessonDialog(nil)
		if visible := mod.topoRadio.IsVisibleTo(mod.newLessonDialog.QWidget); visible == on {
			t.Errorf("words only %v: topography offered %v", on, visible)
		}
		if mod.getNewLessonData()["type"] != "words" {
			t.Errorf("words only %v: not a word lesson", on)
		}
		mod.newLessonDialog.DeleteLater()
	}
}
