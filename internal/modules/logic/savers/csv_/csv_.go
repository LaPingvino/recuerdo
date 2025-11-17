// Package csv provides CSV export functionality using the centralized FileSaver
package csv

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// CsvSaverModule provides CSV export functionality
type CsvSaverModule struct {
	*core.BaseModule
	manager   *core.Manager
	fileSaver *lesson.FileSaver
	active    bool
}

// NewCsvSaverModule creates a new CsvSaverModule instance
func NewCsvSaverModule() *CsvSaverModule {
	base := core.NewBaseModule("logic", "csv-saver-module")

	return &CsvSaverModule{
		BaseModule: base,
		fileSaver:  lesson.NewFileSaver(),
		active:     false,
	}
}

// Enable activates the module
func (mod *CsvSaverModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	mod.active = true
	fmt.Println("CsvSaverModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *CsvSaverModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	mod.active = false
	fmt.Println("CsvSaverModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *CsvSaverModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// GetType returns the module type
func (mod *CsvSaverModule) GetType() string {
	return "save"
}

// GetSaveFormats returns the formats this module can save
func (mod *CsvSaverModule) GetSaveFormats() map[string]string {
	return map[string]string{
		"csv": "Comma-Separated Values (Spreadsheet)",
	}
}

// CanSave checks if this module can save the given lesson type to the specified format
func (mod *CsvSaverModule) CanSave(lessonType string, format string) bool {
	if !mod.active {
		return false
	}

	// CSV format supports words lesson type
	return lessonType == "words" && format == "csv"
}

// Save saves the lesson data to the specified path in CSV format
func (mod *CsvSaverModule) Save(lessonData *lesson.LessonData, filePath string) error {
	if !mod.active {
		return fmt.Errorf("CSV saver module is not active")
	}

	// Validate file extension
	ext := filepath.Ext(filePath)
	if ext != ".csv" {
		return fmt.Errorf("CSV saver can only save .csv files, got %s", ext)
	}

	// Use centralized file saver
	return mod.fileSaver.SaveWithValidation(lessonData, filePath)
}

// GetDefaultExtension returns the default file extension for this saver
func (mod *CsvSaverModule) GetDefaultExtension() string {
	return ".csv"
}

// GetFileFilter returns Qt-style file filter for this format
func (mod *CsvSaverModule) GetFileFilter() string {
	return "CSV Files (*.csv)"
}

// GetDescription returns a description of the CSV format
func (mod *CsvSaverModule) GetDescription() string {
	return "Exports lesson data as comma-separated values suitable for spreadsheet applications like Excel, LibreOffice Calc, and Google Sheets."
}

// ValidateBeforeSave performs format-specific validation before saving
func (mod *CsvSaverModule) ValidateBeforeSave(lessonData *lesson.LessonData) error {
	// Use the centralized validation
	return mod.fileSaver.ValidateLessonData(lessonData)
}

// GetSuggestedFilename returns a suggested filename for the lesson
func (mod *CsvSaverModule) GetSuggestedFilename(lessonData *lesson.LessonData) string {
	return mod.fileSaver.GetDefaultFilename(lessonData, ".csv")
}

// IsActive returns whether the module is currently active
func (mod *CsvSaverModule) IsActive() bool {
	return mod.active
}

// GetPriority returns the priority of this saver (higher = preferred)
func (mod *CsvSaverModule) GetPriority() int {
	return 925 // Same as original Python implementation
}

// InitCsvSaverModule creates and returns a new CsvSaverModule instance
func InitCsvSaverModule() core.Module {
	return NewCsvSaverModule()
}
