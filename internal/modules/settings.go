package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/LaPingvino/openteacher/internal/core"
)

// SettingsModule provides configuration management for the application
type SettingsModule struct {
	*core.BaseModule
	settings map[string]interface{}
	filePath string
	mu       sync.RWMutex
}

// NewSettingsModule creates a new settings module
func NewSettingsModule() *SettingsModule {
	base := core.NewBaseModule("settings", "settings-module")
	base.SetPriority(1500) // High priority - many modules depend on settings

	// Default settings file path
	homeDir, _ := os.UserHomeDir()
	settingsPath := filepath.Join(homeDir, ".openteacher", "settings.json")

	return &SettingsModule{
		BaseModule: base,
		settings:   make(map[string]interface{}),
		filePath:   settingsPath,
	}
}

// Enable initializes the settings module
func (s *SettingsModule) Enable(ctx context.Context) error {
	if err := s.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// Ensure settings directory exists
	if err := s.ensureSettingsDir(); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Load existing settings
	if err := s.LoadSettings(); err != nil {
		// If settings don't exist, create defaults
		if os.IsNotExist(err) {
			s.setDefaultSettings()
			if saveErr := s.SaveSettings(); saveErr != nil {
				fmt.Printf("Warning: failed to save default settings: %v\n", saveErr)
			}
		} else {
			return fmt.Errorf("failed to load settings: %w", err)
		}
	}

	fmt.Printf("Settings module enabled - loaded from: %s\n", s.filePath)
	return nil
}

// Disable shuts down the settings module
func (s *SettingsModule) Disable(ctx context.Context) error {
	// Save settings before shutdown
	if err := s.SaveSettings(); err != nil {
		fmt.Printf("Warning: failed to save settings during shutdown: %v\n", err)
	}

	fmt.Println("Settings module disabled")
	return s.BaseModule.Disable(ctx)
}

// GetSetting retrieves a configuration value
func (s *SettingsModule) GetSetting(key string) (interface{}, error) {
	if key == "" {
		return nil, fmt.Errorf("setting key cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.settings[key]
	if !exists {
		return nil, fmt.Errorf("setting %q not found", key)
	}

	return value, nil
}

// SetSetting stores a configuration value
func (s *SettingsModule) SetSetting(key string, value interface{}) error {
	if key == "" {
		return fmt.Errorf("setting key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings[key] = value
	return nil
}

// GetSettingWithDefault retrieves a setting or returns a default value
func (s *SettingsModule) GetSettingWithDefault(key string, defaultValue interface{}) interface{} {
	value, err := s.GetSetting(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetString retrieves a string setting
func (s *SettingsModule) GetString(key string) (string, error) {
	value, err := s.GetSetting(key)
	if err != nil {
		return "", err
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("setting %q is not a string", key)
	}

	return str, nil
}

// GetBool retrieves a boolean setting
func (s *SettingsModule) GetBool(key string) (bool, error) {
	value, err := s.GetSetting(key)
	if err != nil {
		return false, err
	}

	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("setting %q is not a boolean", key)
	}

	return b, nil
}

// GetInt retrieves an integer setting
func (s *SettingsModule) GetInt(key string) (int, error) {
	value, err := s.GetSetting(key)
	if err != nil {
		return 0, err
	}

	// JSON unmarshaling creates float64 for numbers
	if f, ok := value.(float64); ok {
		return int(f), nil
	}

	if i, ok := value.(int); ok {
		return i, nil
	}

	return 0, fmt.Errorf("setting %q is not an integer", key)
}

// LoadSettings loads settings from storage
func (s *SettingsModule) LoadSettings() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &s.settings); err != nil {
		return fmt.Errorf("failed to parse settings file: %w", err)
	}

	return nil
}

// SaveSettings persists settings to storage
func (s *SettingsModule) SaveSettings() error {
	s.mu.RLock()
	settingsCopy := make(map[string]interface{})
	for k, v := range s.settings {
		settingsCopy[k] = v
	}
	s.mu.RUnlock()

	data, err := json.MarshalIndent(settingsCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// SetSettingsPath changes the settings file path
func (s *SettingsModule) SetSettingsPath(path string) error {
	if path == "" {
		return fmt.Errorf("settings path cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.filePath = path
	return nil
}

// GetSettingsPath returns the current settings file path
func (s *SettingsModule) GetSettingsPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filePath
}

// ListSettings returns all setting keys
func (s *SettingsModule) ListSettings() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.settings))
	for key := range s.settings {
		keys = append(keys, key)
	}

	return keys
}

// SettingCount returns the number of stored settings
func (s *SettingsModule) SettingCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.settings)
}

// ClearSettings removes all settings
func (s *SettingsModule) ClearSettings() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings = make(map[string]interface{})
}

// ensureSettingsDir creates the settings directory if it doesn't exist
func (s *SettingsModule) ensureSettingsDir() error {
	dir := filepath.Dir(s.filePath)
	return os.MkdirAll(dir, 0755)
}

// setDefaultSettings initializes the settings with default values
func (s *SettingsModule) setDefaultSettings() {
	s.settings = map[string]interface{}{
		"app.name":          "OpenTeacher",
		"app.version":       "4.0.0-alpha",
		"app.profile":       "all",
		"ui.language":       "en",
		"ui.theme":          "default",
		"app.autoSave":      true,
		"app.autoSaveDelay": 30,
		"debug.enabled":     false,
		"debug.logLevel":    "info",
		"window.width":      800,
		"window.height":     600,
		"window.maximized":  false,
	}
}
