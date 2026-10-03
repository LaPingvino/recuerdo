package modules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// MetadataModule provides application metadata and configuration
type MetadataModule struct {
	*core.BaseModule
	manager  *core.Manager
	metadata map[string]interface{}
}

// NewMetadataModule creates a new metadata module
func NewMetadataModule() *MetadataModule {
	base := core.NewBaseModule("metadata", "metadata-module")
	base.SetPriority(2000) // Very high priority - other modules depend on metadata

	return &MetadataModule{
		BaseModule: base,
		metadata:   make(map[string]interface{}),
	}
}

// Enable initializes the metadata module
func (m *MetadataModule) Enable(ctx context.Context) error {
	if err := m.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// Initialize metadata
	m.initializeMetadata()

	fmt.Println("Metadata module enabled - application metadata available")
	return nil
}

// Disable shuts down the metadata module
func (m *MetadataModule) Disable(ctx context.Context) error {
	fmt.Println("Metadata module disabled")
	return m.BaseModule.Disable(ctx)
}

// SetManager sets the module manager
func (m *MetadataModule) SetManager(manager *core.Manager) {
	m.manager = manager
}

// GetMetadata returns a copy of all metadata
func (m *MetadataModule) GetMetadata() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m.metadata {
		result[k] = v
	}
	return result
}

// Get retrieves a specific metadata value
func (m *MetadataModule) Get(key string) (interface{}, bool) {
	value, exists := m.metadata[key]
	return value, exists
}

// GetString retrieves a string metadata value
func (m *MetadataModule) GetString(key string) (string, bool) {
	if value, exists := m.metadata[key]; exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetInt retrieves an integer metadata value
func (m *MetadataModule) GetInt(key string) (int, bool) {
	if value, exists := m.metadata[key]; exists {
		if i, ok := value.(int); ok {
			return i, true
		}
	}
	return 0, false
}

// Set sets a metadata value
func (m *MetadataModule) Set(key string, value interface{}) {
	m.metadata[key] = value
}

// initializeMetadata sets up the default application metadata
func (m *MetadataModule) initializeMetadata() {
	// Get executable path for icon path resolution
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	// Look for icons in common locations
	iconPath := m.findIconPath(execDir)
	logoPath := m.findLogoPath(execDir)

	m.metadata = map[string]interface{}{
		"name":             "Recuerdo",
		"version":          "4.0.0-alpha",
		"shortDescription": "Free open source learning software",
		"longDescription":  "Recuerdo is a free open source application that helps you learn a foreign language vocabulary, topography or any other subject that can be learned by heart quickly and effectively.",
		"author":           "Joop Kiefte (based on OpenTeacher by the OpenTeacher Team)",
		"website":          "https://github.com/LaPingvino/recuerdo",
		"email":            "",
		"iconPath":         iconPath,
		"logoPath":         logoPath,

		// Visual theme properties
		"mainColorHue":    200, // Blue theme
		"accentColor":     "#3B4D55",
		"backgroundColor": "#F5F5F5",

		// Application properties
		"applicationId": "eu.kiefte.Recuerdo",
		"category":      "Education",
		"keywords":      []string{"education", "learning", "vocabulary", "language", "study"},

		// License information
		"license":    "GPL-3.0+",
		"licenseUrl": "https://www.gnu.org/licenses/gpl-3.0.html",

		// Build information
		"buildDate":   "2024-01-01", // Would be set during build
		"buildNumber": "1",
		"gitCommit":   "unknown", // Would be set during build

		// Platform information
		"platform":     "Go",
		"minGoVersion": "1.19",

		// Feature flags
		"hasGUI":        true,
		"hasAudio":      true,
		"hasNetwork":    true,
		"hasFileImport": true,
		"hasFileExport": true,

		// Supported file formats
		"supportedImportFormats": []string{".ot", ".csv", ".tsv", ".xml"},
		"supportedExportFormats": []string{".ot", ".csv", ".tsv", ".xml", ".html", ".pdf"},

		// Default settings
		"defaultLanguage":     "en",
		"defaultTheme":        "default",
		"defaultWindowWidth":  640,
		"defaultWindowHeight": 520,

		// Resource paths
		"resourcePath":    filepath.Join(execDir, "resources"),
		"translationPath": filepath.Join(execDir, "translations"),
		"themePath":       filepath.Join(execDir, "themes"),
		"helpPath":        filepath.Join(execDir, "help"),
	}
}

// findIconPath attempts to locate the application icon
func (m *MetadataModule) findIconPath(execDir string) string {
	// Common icon locations to check
	locations := []string{
		filepath.Join(execDir, "icons", "openteacher.png"),
		filepath.Join(execDir, "resources", "icons", "openteacher.png"),
		filepath.Join(execDir, "..", "resources", "icons", "openteacher.png"),
		"/usr/share/icons/hicolor/48x48/apps/openteacher.png",
		"/usr/share/pixmaps/openteacher.png",
	}

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Return a default path even if file doesn't exist
	return filepath.Join(execDir, "icons", "openteacher.png")
}

// findLogoPath attempts to locate the application logo
func (m *MetadataModule) findLogoPath(execDir string) string {
	// Common logo locations to check
	locations := []string{
		filepath.Join(execDir, "icons", "logo.png"),
		filepath.Join(execDir, "resources", "icons", "logo.png"),
		filepath.Join(execDir, "..", "resources", "icons", "logo.png"),
		filepath.Join(execDir, "icons", "openteacher-logo.png"),
		filepath.Join(execDir, "resources", "icons", "openteacher-logo.png"),
	}

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Fallback to icon path
	return m.findIconPath(execDir)
}

// GetResourcePath returns the path to a resource file
func (m *MetadataModule) GetResourcePath(resource string) string {
	if resourcePath, ok := m.GetString("resourcePath"); ok {
		return filepath.Join(resourcePath, resource)
	}
	return resource
}

// GetTranslationPath returns the path to translation files
func (m *MetadataModule) GetTranslationPath(language string) string {
	if translationPath, ok := m.GetString("translationPath"); ok {
		return filepath.Join(translationPath, language+".json")
	}
	return filepath.Join("translations", language+".json")
}
