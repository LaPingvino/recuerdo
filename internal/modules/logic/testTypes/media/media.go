// Package media provides functionality ported from Python module
//
// MediaTestTypeModule handles display and formatting of media lesson test results
package media

import (
	"context"
	"fmt"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// Column constants for media test results table
const (
	NAME = iota
	QUESTION
	ANSWER
	GIVEN_ANSWER
	CORRECT
)

// MediaTestTypeModule is a Go port of the Python MediaTestTypeModule class
type MediaTestTypeModule struct {
	*core.BaseModule
	manager *core.Manager
	list    map[string]interface{}
	test    map[string]interface{}
	active  bool
}

// NewMediaTestTypeModule creates a new MediaTestTypeModule instance
func NewMediaTestTypeModule() *MediaTestTypeModule {
	base := core.NewBaseModule("logic", "media-module")

	return &MediaTestTypeModule{
		BaseModule: base,
		active:     false,
	}
}

// Enable activates the module
func (mod *MediaTestTypeModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	mod.active = true
	fmt.Println("MediaTestTypeModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *MediaTestTypeModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	mod.active = false
	fmt.Println("MediaTestTypeModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *MediaTestTypeModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// GetType returns the module type
func (mod *MediaTestTypeModule) GetType() string {
	return "testType"
}

// GetDataType returns the data type this module handles
func (mod *MediaTestTypeModule) GetDataType() string {
	return "media"
}

// UpdateList updates the list and test data for result display
func (mod *MediaTestTypeModule) UpdateList(list map[string]interface{}, test map[string]interface{}) {
	mod.list = list
	mod.test = test
}

// Header returns the column headers for the media test results table
func (mod *MediaTestTypeModule) Header() []string {
	return []string{
		"Name",
		"Question",
		"Answer",
		"Given answer",
		"Correct",
	}
}

// itemForResult finds the item corresponding to a test result
func (mod *MediaTestTypeModule) itemForResult(result map[string]interface{}) map[string]interface{} {
	if mod.list == nil {
		return nil
	}

	items, ok := mod.list["items"].([]interface{})
	if !ok {
		return nil
	}

	resultItemID, ok := result["itemId"]
	if !ok {
		return nil
	}

	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if itemID, exists := itemMap["id"]; exists && itemID == resultItemID {
			return itemMap
		}
	}

	return nil
}

// Data returns the data for a specific cell in the results table
func (mod *MediaTestTypeModule) Data(row, column int) interface{} {
	if mod.test == nil {
		return nil
	}

	results, ok := mod.test["results"].([]interface{})
	if !ok || row >= len(results) {
		return nil
	}

	result, ok := results[row].(map[string]interface{})
	if !ok {
		return nil
	}

	item := mod.itemForResult(result)
	if item == nil {
		return nil
	}

	switch column {
	case NAME:
		if name, exists := item["name"]; exists {
			return name
		}
		return ""
	case QUESTION:
		if question, exists := item["question"]; exists {
			return question
		}
		return ""
	case ANSWER:
		if answer, exists := item["answer"]; exists {
			return answer
		}
		return ""
	case GIVEN_ANSWER:
		if givenAnswer, exists := result["givenAnswer"]; exists {
			return givenAnswer
		}
		return ""
	case CORRECT:
		if resultStatus, exists := result["result"]; exists {
			return resultStatus == "right"
		}
		return false
	}

	return nil
}

// RowCount returns the number of rows in the results table
func (mod *MediaTestTypeModule) RowCount() int {
	if mod.test == nil {
		return 0
	}

	results, ok := mod.test["results"].([]interface{})
	if !ok {
		return 0
	}

	return len(results)
}

// ColumnCount returns the number of columns in the results table
func (mod *MediaTestTypeModule) ColumnCount() int {
	return 5 // NAME, QUESTION, ANSWER, GIVEN_ANSWER, CORRECT
}

// GetDisplayName returns a user-friendly name for this test type
func (mod *MediaTestTypeModule) GetDisplayName() string {
	return "Media Test"
}

// SupportsLessonType checks if this module supports the given lesson type
func (mod *MediaTestTypeModule) SupportsLessonType(lessonType string) bool {
	return lessonType == "media"
}

// FormatResult formats a single test result for display
func (mod *MediaTestTypeModule) FormatResult(result map[string]interface{}) string {
	item := mod.itemForResult(result)
	if item == nil {
		return "Unknown item"
	}

	name := ""
	if n, exists := item["name"]; exists {
		name = fmt.Sprintf("%v", n)
	}

	givenAnswer := ""
	if ga, exists := result["givenAnswer"]; exists {
		givenAnswer = fmt.Sprintf("%v", ga)
	}

	correct := false
	if res, exists := result["result"]; exists {
		correct = res == "right"
	}

	status := "❌"
	if correct {
		status = "✅"
	}

	return fmt.Sprintf("%s %s: %s", status, name, givenAnswer)
}

// GetStatistics returns statistics about the test results
func (mod *MediaTestTypeModule) GetStatistics() map[string]interface{} {
	if mod.test == nil {
		return map[string]interface{}{
			"total":   0,
			"correct": 0,
			"wrong":   0,
		}
	}

	results, ok := mod.test["results"].([]interface{})
	if !ok {
		return map[string]interface{}{
			"total":   0,
			"correct": 0,
			"wrong":   0,
		}
	}

	correct := 0
	total := len(results)

	for _, result := range results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		if res, exists := resultMap["result"]; exists && res == "right" {
			correct++
		}
	}

	return map[string]interface{}{
		"total":   total,
		"correct": correct,
		"wrong":   total - correct,
		"percentage": func() float64 {
			if total == 0 {
				return 0.0
			}
			return float64(correct) / float64(total) * 100.0
		}(),
	}
}

// InitMediaTestTypeModule creates and returns a new MediaTestTypeModule instance
func InitMediaTestTypeModule() core.Module {
	return NewMediaTestTypeModule()
}
