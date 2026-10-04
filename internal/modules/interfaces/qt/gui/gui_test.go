package gui

import (
	"testing"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessonDialogs"
)

// TestCreateLessonFromDialogData tests the lesson creation from dialog data
func TestCreateLessonFromDialogData(t *testing.T) {
	// Create GUI module instance
	gui := NewGuiModule()

	tests := []struct {
		name          string
		dialogData    map[string]interface{}
		expectedTitle string
		expectedType  string
		expectedQLang string
		expectedALang string
		shouldError   bool
	}{
		{
			name: "Complete lesson data",
			dialogData: map[string]interface{}{
				"name":             "Spanish Vocabulary",
				"description":      "Basic Spanish words for beginners",
				"type":             "words",
				"questionLanguage": "English",
				"answerLanguage":   "Spanish",
			},
			expectedTitle: "Spanish Vocabulary",
			expectedType:  "words",
			expectedQLang: "English",
			expectedALang: "Spanish",
			shouldError:   false,
		},
		{
			name: "Minimal lesson data with defaults",
			dialogData: map[string]interface{}{
				"name": "Test Lesson",
			},
			expectedTitle: "Test Lesson",
			expectedType:  "words",
			expectedQLang: "English",
			expectedALang: "English",
			shouldError:   false,
		},
		{
			name: "Empty name uses default",
			dialogData: map[string]interface{}{
				"name":             "",
				"type":             "topology",
				"questionLanguage": "French",
				"answerLanguage":   "German",
			},
			expectedTitle: "New Lesson",
			expectedType:  "topo",
			expectedQLang: "French",
			expectedALang: "German",
			shouldError:   false,
		},
		{
			name:          "Empty dialog data uses all defaults",
			dialogData:    map[string]interface{}{},
			expectedTitle: "New Lesson",
			expectedType:  "words",
			expectedQLang: "English",
			expectedALang: "English",
			shouldError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLesson, err := gui.CreateLessonFromDialogData(tt.dialogData)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if newLesson == nil {
				t.Fatal("CreateLessonFromDialogData returned nil lesson")
			}

			// Check lesson properties
			if newLesson.Data.List.Title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, newLesson.Data.List.Title)
			}

			if newLesson.DataType != tt.expectedType {
				t.Errorf("Expected type %q, got %q", tt.expectedType, newLesson.DataType)
			}

			if newLesson.Data.List.QuestionLanguage != tt.expectedQLang {
				t.Errorf("Expected question language %q, got %q", tt.expectedQLang, newLesson.Data.List.QuestionLanguage)
			}

			if newLesson.Data.List.AnswerLanguage != tt.expectedALang {
				t.Errorf("Expected answer language %q, got %q", tt.expectedALang, newLesson.Data.List.AnswerLanguage)
			}

			// Check that lesson is marked as changed (unsaved)
			if !newLesson.Data.Changed {
				t.Error("New lesson should be marked as changed")
			}

			// Check that path indicates unsaved status
			if newLesson.Path != "*"+tt.expectedTitle {
				t.Errorf("Expected path %q, got %q", "*"+tt.expectedTitle, newLesson.Path)
			}

			// Check description metadata if provided
			if desc, exists := tt.dialogData["description"]; exists && desc != "" {
				if newLesson.Data.Resources == nil {
					t.Error("Expected Resources map to be initialized")
				} else if storedDesc, ok := newLesson.Data.Resources["description"]; !ok || storedDesc != desc {
					t.Errorf("Expected description %q in resources, got %q", desc, storedDesc)
				}
			}
		})
	}
}

// TestNewLessonWorkflowIntegration tests the complete workflow
func TestNewLessonWorkflowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would be an integration test that requires Qt
	// For now, we'll test the non-Qt parts

	manager := core.NewManager()

	// Register required modules (without actually initializing Qt)
	lessonDialogsModule := lessonDialogs.NewLessonDialogsModule()
	if err := manager.Register(lessonDialogsModule); err != nil {
		t.Fatalf("Failed to register lessonDialogs module: %v", err)
	}

	guiModule := NewGuiModule()
	if err := manager.Register(guiModule); err != nil {
		t.Fatalf("Failed to register gui module: %v", err)
	}

	// Test module registration
	lessonDialogModules := manager.GetModulesByType("lessonDialogs")
	if len(lessonDialogModules) != 1 {
		t.Fatalf("Expected 1 lessonDialogs module, got %d", len(lessonDialogModules))
	}

	// Test interface compliance
	_, hasMethod := lessonDialogModules[0].(interface{ ShowNewLessonDialog() map[string]interface{} })
	if !hasMethod {
		t.Error("lessonDialogs module does not implement ShowNewLessonDialog() method")
	}

	// Test lesson creation workflow without Qt
	mockDialogData := map[string]interface{}{
		"name":             "Integration Test Lesson",
		"description":      "Test lesson for integration testing",
		"type":             "words",
		"questionLanguage": "English",
		"answerLanguage":   "Spanish",
	}

	newLesson, err := guiModule.CreateLessonFromDialogData(mockDialogData)
	if err != nil {
		t.Fatalf("Failed to create lesson from dialog data: %v", err)
	}

	// Verify lesson was created correctly
	if newLesson.Data.List.Title != "Integration Test Lesson" {
		t.Errorf("Expected title 'Integration Test Lesson', got %q", newLesson.Data.List.Title)
	}

	if newLesson.DataType != "words" {
		t.Errorf("Expected type 'words', got %q", newLesson.DataType)
	}

	// Test that lesson can accept new words
	if newLesson.Data.List.Items == nil {
		newLesson.Data.List.Items = []lesson.WordItem{}
	}

	// Add a test word pair
	testItem := lesson.WordItem{
		Questions: []string{"hello"},
		Answers:   []string{"hola"},
	}
	newLesson.Data.List.Items = append(newLesson.Data.List.Items, testItem)

	if len(newLesson.Data.List.Items) != 1 {
		t.Errorf("Expected 1 item in lesson, got %d", len(newLesson.Data.List.Items))
	}

	if newLesson.Data.List.GetWordCount() != 1 {
		t.Errorf("Expected word count 1, got %d", newLesson.Data.List.GetWordCount())
	}
}

// TestLessonDataValidation tests validation of lesson data
func TestLessonDataValidation(t *testing.T) {
	gui := NewGuiModule()

	// Test various invalid data scenarios
	invalidTests := []struct {
		name       string
		dialogData map[string]interface{}
	}{
		{
			name:       "nil dialog data",
			dialogData: nil,
		},
	}

	for _, tt := range invalidTests {
		t.Run(tt.name, func(t *testing.T) {
			newLesson, err := gui.CreateLessonFromDialogData(tt.dialogData)

			// Even with invalid/nil data, we should get a valid lesson with defaults
			if err != nil {
				t.Errorf("Unexpected error with %s: %v", tt.name, err)
			}

			if newLesson == nil {
				t.Errorf("CreateLessonFromDialogData should never return nil lesson")
			}

			// Should have sensible defaults
			if newLesson != nil {
				if newLesson.Data.List.Title != "New Lesson" {
					t.Errorf("Expected default title 'New Lesson', got %q", newLesson.Data.List.Title)
				}
				if newLesson.DataType != "words" {
					t.Errorf("Expected default type 'words', got %q", newLesson.DataType)
				}
			}
		})
	}
}

// TestLessonTypeSupport tests support for different lesson types
func TestLessonTypeSupport(t *testing.T) {
	gui := NewGuiModule()

	supportedTypes := []string{"words", "topo", "media"}

	for _, lessonType := range supportedTypes {
		t.Run("type_"+lessonType, func(t *testing.T) {
			dialogData := map[string]interface{}{
				"name": "Test " + lessonType + " Lesson",
				"type": lessonType,
			}

			newLesson, err := gui.CreateLessonFromDialogData(dialogData)
			if err != nil {
				t.Errorf("Failed to create %s lesson: %v", lessonType, err)
				return
			}

			if newLesson.DataType != lessonType {
				t.Errorf("Expected lesson type %q, got %q", lessonType, newLesson.DataType)
			}
		})
	}
}

// BenchmarkCreateLessonFromDialogData benchmarks lesson creation performance
func BenchmarkCreateLessonFromDialogData(b *testing.B) {
	gui := NewGuiModule()

	dialogData := map[string]interface{}{
		"name":             "Benchmark Lesson",
		"description":      "Performance test lesson",
		"type":             "words",
		"questionLanguage": "English",
		"answerLanguage":   "Spanish",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := gui.CreateLessonFromDialogData(dialogData)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
