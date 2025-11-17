// Package topo provides functionality ported from Python module
//
// TopoTestTypeModule handles display and formatting of topography lesson test results
package topo

import (
	"context"
	"fmt"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// Column constants for topo test results table
const (
	PLACE_NAME = iota
	CORRECT
)

// TopoTestTypeModule is a Go port of the Python TopoTestTypeModule class
type TopoTestTypeModule struct {
	*core.BaseModule
	manager *core.Manager
	list    map[string]interface{}
	test    map[string]interface{}
	active  bool
}

// NewTopoTestTypeModule creates a new TopoTestTypeModule instance
func NewTopoTestTypeModule() *TopoTestTypeModule {
	base := core.NewBaseModule("logic", "topo-module")

	return &TopoTestTypeModule{
		BaseModule: base,
		active:     false,
	}
}

// Enable activates the module
func (mod *TopoTestTypeModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	mod.active = true
	fmt.Println("TopoTestTypeModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *TopoTestTypeModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	mod.active = false
	fmt.Println("TopoTestTypeModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *TopoTestTypeModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// GetType returns the module type
func (mod *TopoTestTypeModule) GetType() string {
	return "testType"
}

// GetDataType returns the data type this module handles
func (mod *TopoTestTypeModule) GetDataType() string {
	return "topo"
}

// UpdateList updates the list and test data for result display
func (mod *TopoTestTypeModule) UpdateList(list map[string]interface{}, test map[string]interface{}) {
	mod.list = list
	mod.test = test
}

// Header returns the column headers for the topo test results table
func (mod *TopoTestTypeModule) Header() []string {
	return []string{
		"Place name",
		"Correct",
	}
}

// itemForResult finds the item corresponding to a test result
func (mod *TopoTestTypeModule) itemForResult(result map[string]interface{}) map[string]interface{} {
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
func (mod *TopoTestTypeModule) Data(row, column int) interface{} {
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
	case PLACE_NAME:
		if name, exists := item["name"]; exists {
			return name
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
func (mod *TopoTestTypeModule) RowCount() int {
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
func (mod *TopoTestTypeModule) ColumnCount() int {
	return 2 // PLACE_NAME, CORRECT
}

// GetDisplayName returns a user-friendly name for this test type
func (mod *TopoTestTypeModule) GetDisplayName() string {
	return "Topography Test"
}

// SupportsLessonType checks if this module supports the given lesson type
func (mod *TopoTestTypeModule) SupportsLessonType(lessonType string) bool {
	return lessonType == "topo"
}

// FormatResult formats a single test result for display
func (mod *TopoTestTypeModule) FormatResult(result map[string]interface{}) string {
	item := mod.itemForResult(result)
	if item == nil {
		return "Unknown location"
	}

	name := ""
	if n, exists := item["name"]; exists {
		name = fmt.Sprintf("%v", n)
	}

	correct := false
	if res, exists := result["result"]; exists {
		correct = res == "right"
	}

	status := "❌"
	if correct {
		status = "✅"
	}

	return fmt.Sprintf("%s %s", status, name)
}

// GetStatistics returns statistics about the test results
func (mod *TopoTestTypeModule) GetStatistics() map[string]interface{} {
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

// GetPlaces returns a list of all places in the current lesson
func (mod *TopoTestTypeModule) GetPlaces() []string {
	if mod.list == nil {
		return nil
	}

	items, ok := mod.list["items"].([]interface{})
	if !ok {
		return nil
	}

	var places []string
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if name, exists := itemMap["name"]; exists {
			if nameStr, ok := name.(string); ok {
				places = append(places, nameStr)
			}
		}
	}

	return places
}

// GetMapInfo returns map information if available
func (mod *TopoTestTypeModule) GetMapInfo() map[string]interface{} {
	if mod.list == nil {
		return nil
	}

	// Try to extract map information from the list
	mapInfo := make(map[string]interface{})

	if mapData, exists := mod.list["mapInfo"]; exists {
		if mapDataMap, ok := mapData.(map[string]interface{}); ok {
			return mapDataMap
		}
	}

	// Return basic map info if no specific data is available
	mapInfo["type"] = "topo"
	mapInfo["places"] = mod.GetPlaces()

	return mapInfo
}

// InitTopoTestTypeModule creates and returns a new TopoTestTypeModule instance
func InitTopoTestTypeModule() core.Module {
	return NewTopoTestTypeModule()
}
