// Package file provides functionality ported from Python module
//
// Provides file dialogs for opening and saving files.
//
// This is an automated port - implementation may be incomplete.
package file

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/mappu/miqt/qt"
)

// FileDialogModule is a Go port of the Python FileDialogModule class
type FileDialogModule struct {
	*core.BaseModule
	manager    *core.Manager
	lastDir    string
	fileFilter string
}

// NewFileDialogModule creates a new FileDialogModule instance
func NewFileDialogModule() *FileDialogModule {
	base := core.NewBaseModule("fileDialog", "file-dialog-module")
	base.SetRequires("qtApp")

	return &FileDialogModule{
		BaseModule: base,
		lastDir:    "",
		fileFilter: "All Lesson Files (*.ot *.otwd *.csv *.tsv *.txt *.json *.kvtml *.anki *.anki2 *.apkg *.xml *.kgm *.ottp *.otmd);;OpenTeacher Files (*.ot *.otwd *.ottp *.otmd);;Text & CSV Files (*.txt *.csv *.tsv *.json);;Anki Files (*.anki *.anki2 *.apkg);;KDE/Educational Files (*.kvtml *.kgm);;Vocabulary Trainers (*.voc *.fq *.fmd *.dkf *.jml *.jvlt *.stp *.db);;Language Learning Apps (*.oh *.ohw *.oh4 *.ovr *.pau *.t2k *.vok2 *.wdl *.vtl3 *.wrts);;Other Formats (*.backpack *.wcu *.xml);;All Files (*.*)",
	}
}

// OpenFile shows an open file dialog and returns the selected file path
func (mod *FileDialogModule) OpenFile(parent interface{}, title string, filter string) string {
	// Convert parent to proper type
	var parentWidget *qt.QWidget
	if parent != nil {
		if pw, ok := parent.(*qt.QWidget); ok {
			parentWidget = pw
		}
	}
	if filter == "" {
		filter = mod.fileFilter
	}

	// Use GetOpenFileName with proper parameters
	selectedFile := qt.QFileDialog_GetOpenFileName4(parentWidget, title, mod.lastDir, filter)
	if selectedFile != "" {
		mod.lastDir = filepath.Dir(selectedFile)
		return selectedFile
	}

	return ""
}

// OpenFiles shows an open files dialog and returns the selected file paths
func (mod *FileDialogModule) OpenFiles(parent *qt.QWidget, title string, filter string) []string {
	if filter == "" {
		filter = mod.fileFilter
	}

	// Use GetOpenFileNames with proper parameters
	selectedFiles := qt.QFileDialog_GetOpenFileNames4(parent, title, mod.lastDir, filter)
	if len(selectedFiles) > 0 {
		mod.lastDir = filepath.Dir(selectedFiles[0])
		return selectedFiles
	}

	return []string{}
}

// SaveFile shows a save file dialog and returns the selected file path
func (mod *FileDialogModule) SaveFile(parent *qt.QWidget, title string, filter string, defaultName string) string {
	if filter == "" {
		filter = mod.fileFilter
	}

	startPath := mod.lastDir
	if defaultName != "" {
		startPath = filepath.Join(mod.lastDir, defaultName)
	}

	selectedFile := qt.QFileDialog_GetSaveFileName4(parent, title, startPath, filter)
	if selectedFile != "" {
		mod.lastDir = filepath.Dir(selectedFile)
		return selectedFile
	}

	return ""
}

// SelectDirectory shows a directory selection dialog
func (mod *FileDialogModule) SelectDirectory(parent *qt.QWidget, title string) string {
	selectedDir := qt.QFileDialog_GetExistingDirectory3(parent, title, mod.lastDir)
	if selectedDir != "" {
		mod.lastDir = selectedDir
		return selectedDir
	}

	return ""
}

// SetDefaultDirectory sets the default directory for file dialogs
func (mod *FileDialogModule) SetDefaultDirectory(dir string) {
	mod.lastDir = dir
}

// GetDefaultDirectory returns the current default directory
func (mod *FileDialogModule) GetDefaultDirectory() string {
	return mod.lastDir
}

// SetDefaultFilter sets the default file filter
func (mod *FileDialogModule) SetDefaultFilter(filter string) {
	mod.fileFilter = filter
}

// GetSupportedFormats returns supported file formats
func (mod *FileDialogModule) GetSupportedFormats() map[string]string {
	return map[string]string{
		"All Lesson Files":       "*.ot *.otwd *.csv *.tsv *.txt *.json *.kvtml *.anki *.anki2 *.apkg *.xml *.kgm *.ottp *.otmd",
		"OpenTeacher Files":      "*.ot *.otwd *.ottp *.otmd",
		"Text & CSV Files":       "*.txt *.csv *.tsv *.json",
		"Anki Files":             "*.anki *.anki2 *.apkg",
		"KDE/Educational Files":  "*.kvtml *.kgm",
		"Vocabulary Trainers":    "*.voc *.fq *.fmd *.dkf *.jml *.jvlt *.stp *.db",
		"Language Learning Apps": "*.oh *.ohw *.oh4 *.ovr *.pau *.t2k *.vok2 *.wdl *.vtl3 *.wrts",
		"Other Formats":          "*.backpack *.wcu *.xml",
		"All Files":              "*.*",
	}
}

// BuildFilterString builds a filter string from format map
func (mod *FileDialogModule) BuildFilterString(formats map[string]string) string {
	if len(formats) == 0 {
		return mod.fileFilter
	}

	var filters []string
	for name, pattern := range formats {
		filters = append(filters, fmt.Sprintf("%s (%s)", name, pattern))
	}

	result := ""
	for i, filter := range filters {
		if i > 0 {
			result += ";;"
		}
		result += filter
	}

	return result
}

// Enable activates the module
func (mod *FileDialogModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// Set default directory to user's home directory
	homeDir := qt.QDir_HomePath()
	if homeDir != "" {
		mod.lastDir = homeDir
	}

	fmt.Println("FileDialogModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *FileDialogModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	fmt.Println("FileDialogModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *FileDialogModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitFileDialogModule creates and returns a new FileDialogModule instance
// This is the Go equivalent of the Python init function
func InitFileDialogModule() core.Module {
	return NewFileDialogModule()
}
