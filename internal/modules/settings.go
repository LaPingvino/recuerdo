package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LaPingvino/recuerdo/internal/core"
	"os"
	"path/filepath"
	"sync"
)

// SettingsModule provides configuration management for the application
type SettingsModule struct {
	*core.BaseModule
	settings map[string]interface{}
	filePath string
	mu       sync.RWMutex // guards settings and filePath
	fileMu   sync.Mutex   // serialises reading and writing the settings file; taken before mu
}

// NewSettingsModule creates a new settings module
func NewSettingsModule() *SettingsModule {
	base := core.NewBaseModule("settings", "settings-module")
	base.SetPriority(1500) // High priority - many modules depend on settings

	return &SettingsModule{
		BaseModule: base,
		settings:   make(map[string]interface{}),
		filePath:   DefaultSettingsPath(),
	}
}

// DefaultSettingsPath is recuerdo/settings.json in the user's configuration
// directory (~/.config on Linux, ~/Library/Application Support on macOS,
// %AppData% on Windows).
func DefaultSettingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "recuerdo", "settings.json")
}

// legacySettingsPath is where earlier Recuerdo versions kept settings.
func legacySettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".openteacher", "settings.json")
}

// migrateSettings copies the settings file at legacy to path if path does
// not exist yet. The old file is left alone. It reports whether it copied.
func migrateSettings(path, legacy string) (bool, error) {
	if legacy == "" {
		return false, nil
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return false, err
	}
	data, err := os.ReadFile(legacy)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return false, err
	}
	return true, os.Rename(tmp, path)
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

	// First run with the new location: carry over the old settings.
	if s.GetSettingsPath() == DefaultSettingsPath() {
		if copied, err := migrateSettings(s.GetSettingsPath(), legacySettingsPath()); err != nil {
			fmt.Printf("Warning: could not copy settings from %s: %v\n", legacySettingsPath(), err)
		} else if copied {
			fmt.Printf("Copied settings from %s\n", legacySettingsPath())
		}
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

// LoadSettings loads settings from storage, merging them into the
// current settings. A file that cannot be parsed leaves them unchanged.
func (s *SettingsModule) LoadSettings() error {
	s.fileMu.Lock()
	defer s.fileMu.Unlock()

	data, err := os.ReadFile(s.GetSettingsPath())
	if err != nil {
		return err
	}

	loaded := make(map[string]interface{})
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("failed to parse settings file: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range loaded {
		s.settings[k] = v
	}
	return nil
}

// SaveSettings persists settings to storage. Saves are serialised and
// each one writes a complete file under a temporary name before renaming
// it into place, so concurrent saves and loads never see a partial file.
func (s *SettingsModule) SaveSettings() error {
	s.fileMu.Lock()
	defer s.fileMu.Unlock()

	s.mu.RLock()
	path := s.filePath
	data, err := json.MarshalIndent(s.settings, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := writeFileAtomic(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}
	return nil
}

// writeFileAtomic writes data to a temporary file next to path and renames
// it over path.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op once renamed

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), perm); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
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
		"app.name":          "Recuerdo",
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
