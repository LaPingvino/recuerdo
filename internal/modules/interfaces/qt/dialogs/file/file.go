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

	"github.com/LaPingvino/openteacher/internal/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
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
		fileFilter: "All Files (*.*)",
	}
}

// OpenFile shows an open file dialog and returns the selected file path
func (mod *FileDialogModule) OpenFile(parent *widgets.QWidget, title string, filter string) string {
	if filter == "" {
		filter = mod.fileFilter
	}

	dialog := widgets.NewQFileDialog2(parent, title, mod.lastDir, filter)
	dialog.SetFileMode(widgets.QFileDialog__ExistingFile)
	dialog.SetAcceptMode(widgets.QFileDialog__AcceptOpen)

	if dialog.Exec() == int(widgets.QDialog__Accepted) {
		selectedFiles := dialog.SelectedFiles()
		if len(selectedFiles) > 0 {
			filePath := selectedFiles[0]
			mod.lastDir = filepath.Dir(filePath)
			return filePath
		}
	}

	return ""
}

// OpenFiles shows an open files dialog and returns the selected file paths
func (mod *FileDialogModule) OpenFiles(parent *widgets.QWidget, title string, filter string) []string {
	if filter == "" {
		filter = mod.fileFilter
	}

	dialog := widgets.NewQFileDialog2(parent, title, mod.lastDir, filter)
	dialog.SetFileMode(widgets.QFileDialog__ExistingFiles)
	dialog.SetAcceptMode(widgets.QFileDialog__AcceptOpen)

	if dialog.Exec() == int(widgets.QDialog__Accepted) {
		selectedFiles := dialog.SelectedFiles()
		if len(selectedFiles) > 0 {
			mod.lastDir = filepath.Dir(selectedFiles[0])
			return selectedFiles
		}
	}

	return []string{}
}

// SaveFile shows a save file dialog and returns the selected file path
func (mod *FileDialogModule) SaveFile(parent *widgets.QWidget, title string, filter string, defaultName string) string {
	if filter == "" {
		filter = mod.fileFilter
	}

	startPath := mod.lastDir
	if defaultName != "" {
		startPath = filepath.Join(mod.lastDir, defaultName)
	}

	dialog := widgets.NewQFileDialog2(parent, title, startPath, filter)
	dialog.SetFileMode(widgets.QFileDialog__AnyFile)
	dialog.SetAcceptMode(widgets.QFileDialog__AcceptSave)
	dialog.SetDefaultSuffix("ot")

	if dialog.Exec() == int(widgets.QDialog__Accepted) {
		selectedFiles := dialog.SelectedFiles()
		if len(selectedFiles) > 0 {
			filePath := selectedFiles[0]
			mod.lastDir = filepath.Dir(filePath)
			return filePath
		}
	}

	return ""
}

// SelectDirectory shows a directory selection dialog
func (mod *FileDialogModule) SelectDirectory(parent *widgets.QWidget, title string) string {
	dialog := widgets.NewQFileDialog2(parent, title, mod.lastDir, "")
	dialog.SetFileMode(widgets.QFileDialog__Directory)
	dialog.SetOption(widgets.QFileDialog__ShowDirsOnly, true)

	if dialog.Exec() == int(widgets.QDialog__Accepted) {
		selectedFiles := dialog.SelectedFiles()
		if len(selectedFiles) > 0 {
			dirPath := selectedFiles[0]
			mod.lastDir = dirPath
			return dirPath
		}
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
		"OpenTeacher Files": "*.ot",
		"Text Files":        "*.txt",
		"All Files":         "*.*",
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
	homeDir := qtcore.QDir_HomePath()
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
func InitFileDialogModule() core.Module {
	return NewFileDialogModule()
}
