package main

import (
	"testing"
	"time"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/gui"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessonDialogs"
)

// TestCompleteLessonCreationWorkflow tests the full end-to-end lesson creation
func TestCompleteLessonCreationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Step 1: Setup module system
	manager := core.NewManager()

	// Register required modules
	lessonDialogsModule := lessonDialogs.NewLessonDialogsModule()
	if err := manager.Register(lessonDialogsModule); err != nil {
		t.Fatalf("Failed to register lessonDialogs module: %v", err)
	}

	guiModule := gui.NewGuiModule()
	if err := manager.Register(guiModule); err != nil {
		t.Fatalf("Failed to register gui module: %v", err)
	}

	t.Log("✅ Modules registered successfully")

	// Step 2: Verify module registration and interfaces
	lessonDialogModules := manager.GetModulesByType("lessonDialogs")
	if len(lessonDialogModules) != 1 {
		t.Fatalf("Expected 1 lessonDialogs module, got %d", len(lessonDialogModules))
	}

	_, hasMethod := lessonDialogModules[0].(interface{ ShowNewLessonDialog() map[string]interface{} })
	if !hasMethod {
		t.Fatal("lessonDialogs module does not implement ShowNewLessonDialog() method")
	}

	t.Log("✅ Module interfaces verified")

	// Step 3: Simulate dialog data (what would come from actual dialog)
	testCases := []struct {
		name         string
		dialogData   map[string]interface{}
		expectedType string
		shouldWork   bool
	}{
		{
			name: "Spanish vocabulary lesson",
			dialogData: map[string]interface{}{
				"name":             "Spanish Basics",
				"description":      "Essential Spanish vocabulary for beginners",
				"type":             "words",
				"questionLanguage": "English",
				"answerLanguage":   "Spanish",
			},
			expectedType: "words",
			shouldWork:   true,
		},
		{
			name: "Geography lesson",
			dialogData: map[string]interface{}{
				"name":             "European Capitals",
				"description":      "Capital cities of European countries",
				"type":             "topology",
				"questionLanguage": "English",
				"answerLanguage":   "English",
			},
			expectedType: "topo",
			shouldWork:   true,
		},
		{
			name: "Minimal data with defaults",
			dialogData: map[string]interface{}{
				"name": "Quick Test",
			},
			expectedType: "words",
			shouldWork:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Step 4: Create lesson from dialog data
			newLesson, err := guiModule.CreateLessonFromDialogData(tc.dialogData)
			if !tc.shouldWork {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to create lesson: %v", err)
			}

			if newLesson == nil {
				t.Fatal("Created lesson is nil")
			}

			// Step 5: Verify lesson properties
			expectedName := tc.dialogData["name"].(string)
			if expectedName == "" {
				expectedName = "New Lesson"
			}

			if newLesson.Data.List.Title != expectedName {
				t.Errorf("Expected title %q, got %q", expectedName, newLesson.Data.List.Title)
			}

			if newLesson.DataType != tc.expectedType {
				t.Errorf("Expected type %q, got %q", tc.expectedType, newLesson.DataType)
			}

			t.Logf("✅ Created lesson: %s (type: %s)", newLesson.Data.List.Title, newLesson.DataType)

			// Step 6: Test lesson functionality (add content)
			if newLesson.DataType == "words" {
				// Test adding vocabulary words
				testWords := []struct {
					question string
					answer   string
				}{
					{"hello", "hola"},
					{"goodbye", "adiós"},
					{"thank you", "gracias"},
				}

				if newLesson.Data.List.Items == nil {
					newLesson.Data.List.Items = []lesson.WordItem{}
				}

				for _, word := range testWords {
					item := lesson.WordItem{
						Questions: []string{word.question},
						Answers:   []string{word.answer},
					}
					newLesson.Data.List.Items = append(newLesson.Data.List.Items, item)
				}

				// Verify word count
				expectedCount := len(testWords)
				actualCount := newLesson.Data.List.GetWordCount()
				if actualCount != expectedCount {
					t.Errorf("Expected %d words, got %d", expectedCount, actualCount)
				}

				t.Logf("✅ Added %d vocabulary pairs", actualCount)
			}

			// Step 7: Test lesson state management
			if !newLesson.Data.Changed {
				t.Error("New lesson should be marked as changed")
			}

			// Step 8: Verify lesson can be saved (simulate save path)
			if newLesson.Path == "" {
				t.Error("Lesson should have a path")
			}

			if newLesson.Path[0] != '*' {
				t.Error("New lesson path should start with * to indicate unsaved")
			}

			t.Log("✅ Lesson state management verified")
		})
	}

	t.Log("🎉 Complete lesson creation workflow test passed")
}

// TestLessonCreationPerformance tests the performance of lesson creation
func TestLessonCreationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	guiModule := gui.NewGuiModule()

	dialogData := map[string]interface{}{
		"name":             "Performance Test Lesson",
		"description":      "Testing lesson creation performance",
		"type":             "words",
		"questionLanguage": "English",
		"answerLanguage":   "Spanish",
	}

	// Measure lesson creation time
	start := time.Now()
	iterations := 100

	for i := 0; i < iterations; i++ {
		_, err := guiModule.CreateLessonFromDialogData(dialogData)
		if err != nil {
			t.Fatalf("Performance test failed on iteration %d: %v", i, err)
		}
	}

	duration := time.Since(start)
	avgTime := duration / time.Duration(iterations)

	t.Logf("📊 Performance results:")
	t.Logf("   Total time for %d lessons: %v", iterations, duration)
	t.Logf("   Average time per lesson: %v", avgTime)

	// Reasonable performance expectation (adjust as needed)
	maxAvgTime := 10 * time.Millisecond
	if avgTime > maxAvgTime {
		t.Errorf("Lesson creation too slow: %v > %v", avgTime, maxAvgTime)
	}
}

// TestLessonModificationWorkflow tests modifying lessons after creation
func TestLessonModificationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping modification workflow test in short mode")
	}

	guiModule := gui.NewGuiModule()

	// Create initial lesson
	dialogData := map[string]interface{}{
		"name":             "Modification Test",
		"description":      "Test lesson for modification",
		"type":             "words",
		"questionLanguage": "English",
		"answerLanguage":   "French",
	}

	newLesson, err := guiModule.CreateLessonFromDialogData(dialogData)
	if err != nil {
		t.Fatalf("Failed to create lesson: %v", err)
	}

	// Add initial content
	if newLesson.Data.List.Items == nil {
		newLesson.Data.List.Items = []lesson.WordItem{}
	}

	initialItem := lesson.WordItem{
		Questions: []string{"cat"},
		Answers:   []string{"chat"},
	}
	newLesson.Data.List.Items = append(newLesson.Data.List.Items, initialItem)

	initialCount := len(newLesson.Data.List.Items)
	t.Logf("✅ Created lesson with %d items", initialCount)

	// Test modification: add more content
	additionalItems := []lesson.WordItem{
		{Questions: []string{"dog"}, Answers: []string{"chien"}},
		{Questions: []string{"bird"}, Answers: []string{"oiseau"}},
		{Questions: []string{"fish"}, Answers: []string{"poisson"}},
	}

	for _, item := range additionalItems {
		newLesson.Data.List.Items = append(newLesson.Data.List.Items, item)
	}

	finalCount := len(newLesson.Data.List.Items)
	expectedCount := initialCount + len(additionalItems)

	if finalCount != expectedCount {
		t.Errorf("Expected %d items after modification, got %d", expectedCount, finalCount)
	}

	// Verify lesson is still marked as changed
	if !newLesson.Data.Changed {
		t.Error("Modified lesson should still be marked as changed")
	}

	// Test word count functionality
	wordCount := newLesson.Data.List.GetWordCount()
	if wordCount != finalCount {
		t.Errorf("GetWordCount() returned %d, expected %d", wordCount, finalCount)
	}

	t.Logf("✅ Modified lesson now has %d vocabulary pairs", wordCount)
	t.Log("🎉 Lesson modification workflow test passed")
}

// TestErrorHandlingInLessonCreation tests various error scenarios
func TestErrorHandlingInLessonCreation(t *testing.T) {
	guiModule := gui.NewGuiModule()

	// Test with nil dialog data
	t.Run("nil_dialog_data", func(t *testing.T) {
		lesson, err := guiModule.CreateLessonFromDialogData(nil)
		if err != nil {
			t.Errorf("Should handle nil data gracefully: %v", err)
		}
		if lesson == nil {
			t.Error("Should return valid lesson even with nil data")
		}
		if lesson != nil && lesson.Data.List.Title != "New Lesson" {
			t.Error("Should use default title with nil data")
		}
	})

	// Test with empty dialog data
	t.Run("empty_dialog_data", func(t *testing.T) {
		emptyData := map[string]interface{}{}
		lesson, err := guiModule.CreateLessonFromDialogData(emptyData)
		if err != nil {
			t.Errorf("Should handle empty data gracefully: %v", err)
		}
		if lesson == nil {
			t.Error("Should return valid lesson even with empty data")
		}
	})

	// Test with invalid data types
	t.Run("invalid_data_types", func(t *testing.T) {
		invalidData := map[string]interface{}{
			"name":        123,        // Should be string
			"type":        []string{}, // Should be string
			"description": struct{}{}, // Should be string
		}
		lesson, err := guiModule.CreateLessonFromDialogData(invalidData)
		if err != nil {
			t.Errorf("Should handle invalid types gracefully: %v", err)
		}
		if lesson == nil {
			t.Error("Should return valid lesson even with invalid types")
		}
		// Should use defaults when types are wrong
		if lesson != nil && lesson.Data.List.Title != "New Lesson" {
			t.Error("Should use default title when name is wrong type")
		}
	})

	t.Log("✅ Error handling tests passed")
}
