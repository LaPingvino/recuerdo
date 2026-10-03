package modules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsModule(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		module := NewSettingsModule()

		assert.Equal(t, "settings", module.Type())
		assert.Equal(t, "settings-module", module.Name())
		assert.Equal(t, 1500, module.Priority())
		assert.False(t, module.IsActive())
		assert.NotEmpty(t, module.GetSettingsPath())
	})

	t.Run("lifecycle", func(t *testing.T) {
		module := NewSettingsModule()

		// Use temporary directory for testing
		tempDir := t.TempDir()
		settingsPath := filepath.Join(tempDir, "test_settings.json")
		err := module.SetSettingsPath(settingsPath)
		require.NoError(t, err)

		ctx := context.Background()

		// Enable module
		err = module.Enable(ctx)
		require.NoError(t, err)
		assert.True(t, module.IsActive())
		assert.Greater(t, module.SettingCount(), 0) // Should have default settings

		// Disable module
		err = module.Disable(ctx)
		require.NoError(t, err)
		assert.False(t, module.IsActive())
	})

	t.Run("default_settings", func(t *testing.T) {
		module := NewSettingsModule()

		tempDir := t.TempDir()
		settingsPath := filepath.Join(tempDir, "test_settings.json")
		err := module.SetSettingsPath(settingsPath)
		require.NoError(t, err)

		ctx := context.Background()
		err = module.Enable(ctx)
		require.NoError(t, err)

		// Check some default settings
		appName, err := module.GetString("app.name")
		require.NoError(t, err)
		assert.Equal(t, "OpenTeacher", appName)

		autoSave, err := module.GetBool("app.autoSave")
		require.NoError(t, err)
		assert.True(t, autoSave)

		windowWidth, err := module.GetInt("window.width")
		require.NoError(t, err)
		assert.Equal(t, 800, windowWidth)
	})

	t.Run("set_and_get_setting", func(t *testing.T) {
		module := NewSettingsModule()

		// Set string setting
		err := module.SetSetting("test.string", "test value")
		require.NoError(t, err)

		value, err := module.GetSetting("test.string")
		require.NoError(t, err)
		assert.Equal(t, "test value", value)

		// Set bool setting
		err = module.SetSetting("test.bool", true)
		require.NoError(t, err)

		value, err = module.GetSetting("test.bool")
		require.NoError(t, err)
		assert.Equal(t, true, value)

		// Set int setting
		err = module.SetSetting("test.int", 42)
		require.NoError(t, err)

		value, err = module.GetSetting("test.int")
		require.NoError(t, err)
		assert.Equal(t, 42, value)
	})

	t.Run("set_empty_key", func(t *testing.T) {
		module := NewSettingsModule()

		err := module.SetSetting("", "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "setting key cannot be empty")
	})

	t.Run("get_nonexistent_setting", func(t *testing.T) {
		module := NewSettingsModule()

		value, err := module.GetSetting("nonexistent.setting")
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "setting \"nonexistent.setting\" not found")
	})

	t.Run("get_empty_key", func(t *testing.T) {
		module := NewSettingsModule()

		value, err := module.GetSetting("")
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "setting key cannot be empty")
	})

	t.Run("get_setting_with_default", func(t *testing.T) {
		module := NewSettingsModule()

		// Existing setting
		err := module.SetSetting("existing.setting", "existing value")
		require.NoError(t, err)

		value := module.GetSettingWithDefault("existing.setting", "default value")
		assert.Equal(t, "existing value", value)

		// Non-existing setting
		value = module.GetSettingWithDefault("nonexistent.setting", "default value")
		assert.Equal(t, "default value", value)
	})

	t.Run("get_string", func(t *testing.T) {
		module := NewSettingsModule()

		// Valid string
		err := module.SetSetting("test.string", "test value")
		require.NoError(t, err)

		str, err := module.GetString("test.string")
		require.NoError(t, err)
		assert.Equal(t, "test value", str)

		// Non-string value
		err = module.SetSetting("test.int", 42)
		require.NoError(t, err)

		str, err = module.GetString("test.int")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a string")
	})

	t.Run("get_bool", func(t *testing.T) {
		module := NewSettingsModule()

		// Valid bool
		err := module.SetSetting("test.bool", true)
		require.NoError(t, err)

		b, err := module.GetBool("test.bool")
		require.NoError(t, err)
		assert.True(t, b)

		// Non-bool value
		err = module.SetSetting("test.string", "not a bool")
		require.NoError(t, err)

		b, err = module.GetBool("test.string")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a boolean")
	})

	t.Run("get_int", func(t *testing.T) {
		module := NewSettingsModule()

		// Valid int
		err := module.SetSetting("test.int", 42)
		require.NoError(t, err)

		i, err := module.GetInt("test.int")
		require.NoError(t, err)
		assert.Equal(t, 42, i)

		// Valid float64 (JSON unmarshaling creates float64 for numbers)
		err = module.SetSetting("test.float", 42.0)
		require.NoError(t, err)

		i, err = module.GetInt("test.float")
		require.NoError(t, err)
		assert.Equal(t, 42, i)

		// Non-numeric value
		err = module.SetSetting("test.string", "not a number")
		require.NoError(t, err)

		i, err = module.GetInt("test.string")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not an integer")
	})

	t.Run("save_and_load_settings", func(t *testing.T) {
		tempDir := t.TempDir()
		settingsPath := filepath.Join(tempDir, "test_settings.json")

		// Create first module and save settings
		module1 := NewSettingsModule()
		err := module1.SetSettingsPath(settingsPath)
		require.NoError(t, err)

		err = module1.SetSetting("test.key1", "value1")
		require.NoError(t, err)
		err = module1.SetSetting("test.key2", 42)
		require.NoError(t, err)

		err = module1.SaveSettings()
		require.NoError(t, err)

		// Create second module and load settings
		module2 := NewSettingsModule()
		err = module2.SetSettingsPath(settingsPath)
		require.NoError(t, err)

		err = module2.LoadSettings()
		require.NoError(t, err)

		// Check loaded values
		value1, err := module2.GetString("test.key1")
		require.NoError(t, err)
		assert.Equal(t, "value1", value1)

		value2, err := module2.GetInt("test.key2")
		require.NoError(t, err)
		assert.Equal(t, 42, value2)
	})

	t.Run("load_nonexistent_file", func(t *testing.T) {
		module := NewSettingsModule()
		err := module.SetSettingsPath("/nonexistent/path/settings.json")
		require.NoError(t, err)

		err = module.LoadSettings()
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("set_empty_settings_path", func(t *testing.T) {
		module := NewSettingsModule()

		err := module.SetSettingsPath("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "settings path cannot be empty")
	})

	t.Run("list_settings", func(t *testing.T) {
		module := NewSettingsModule()

		// Initially empty
		settings := module.ListSettings()
		assert.Empty(t, settings)

		// Add some settings
		err := module.SetSetting("key1", "value1")
		require.NoError(t, err)
		err = module.SetSetting("key2", "value2")
		require.NoError(t, err)
		err = module.SetSetting("key3", "value3")
		require.NoError(t, err)

		settings = module.ListSettings()
		assert.Len(t, settings, 3)
		assert.Contains(t, settings, "key1")
		assert.Contains(t, settings, "key2")
		assert.Contains(t, settings, "key3")
	})

	t.Run("setting_count", func(t *testing.T) {
		module := NewSettingsModule()

		assert.Equal(t, 0, module.SettingCount())

		err := module.SetSetting("key1", "value1")
		require.NoError(t, err)
		assert.Equal(t, 1, module.SettingCount())

		err = module.SetSetting("key2", "value2")
		require.NoError(t, err)
		assert.Equal(t, 2, module.SettingCount())

		// Setting existing key should not increase count
		err = module.SetSetting("key1", "new value")
		require.NoError(t, err)
		assert.Equal(t, 2, module.SettingCount())
	})

	t.Run("clear_settings", func(t *testing.T) {
		module := NewSettingsModule()

		// Add some settings
		err := module.SetSetting("key1", "value1")
		require.NoError(t, err)
		err = module.SetSetting("key2", "value2")
		require.NoError(t, err)
		assert.Equal(t, 2, module.SettingCount())

		// Clear settings
		module.ClearSettings()
		assert.Equal(t, 0, module.SettingCount())

		// Getting settings should now fail
		_, err = module.GetSetting("key1")
		assert.Error(t, err)
	})

	t.Run("interface_compliance", func(t *testing.T) {
		module := NewSettingsModule()

		// Should implement Module interface
		var _ core.Module = module
	})

	t.Run("enable_creates_settings_dir", func(t *testing.T) {
		tempDir := t.TempDir()
		settingsDir := filepath.Join(tempDir, "subdir", "openteacher")
		settingsPath := filepath.Join(settingsDir, "settings.json")

		module := NewSettingsModule()
		err := module.SetSettingsPath(settingsPath)
		require.NoError(t, err)

		// Directory should not exist initially
		_, err = os.Stat(settingsDir)
		assert.True(t, os.IsNotExist(err))

		// Enable should create the directory
		ctx := context.Background()
		err = module.Enable(ctx)
		require.NoError(t, err)

		// Directory should now exist
		stat, err := os.Stat(settingsDir)
		require.NoError(t, err)
		assert.True(t, stat.IsDir())
	})

	t.Run("enable_with_invalid_json", func(t *testing.T) {
		tempDir := t.TempDir()
		settingsPath := filepath.Join(tempDir, "invalid_settings.json")

		// Create invalid JSON file
		err := os.WriteFile(settingsPath, []byte("{invalid json"), 0644)
		require.NoError(t, err)

		module := NewSettingsModule()
		err = module.SetSettingsPath(settingsPath)
		require.NoError(t, err)

		ctx := context.Background()
		err = module.Enable(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load settings")
	})
}

func TestSettingsModuleConcurrency(t *testing.T) {
	t.Run("concurrent_set_and_get", func(t *testing.T) {
		module := NewSettingsModule()
		const numGoroutines = 10
		const operationsPerGoroutine = 20

		done := make(chan bool, numGoroutines*2)

		// Concurrent setters
		for i := 0; i < numGoroutines; i++ {
			go func(routineID int) {
				for j := 0; j < operationsPerGoroutine; j++ {
					key := "test.concurrent"
					value := routineID*100 + j
					err := module.SetSetting(key, value)
					assert.NoError(t, err)
				}
				done <- true
			}(i)
		}

		// Concurrent getters
		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < operationsPerGoroutine; j++ {
					_, err := module.GetSetting("test.concurrent")
					// May error if key doesn't exist yet, which is fine
					if err != nil {
						assert.Contains(t, err.Error(), "not found")
					}
					time.Sleep(1 * time.Millisecond)
				}
				done <- true
			}()
		}

		// Wait for all operations to complete
		for i := 0; i < numGoroutines*2; i++ {
			<-done
		}

		// Final state should be consistent
		value, err := module.GetSetting("test.concurrent")
		assert.NoError(t, err)
		assert.NotNil(t, value)
	})

	t.Run("concurrent_list_and_modify", func(t *testing.T) {
		module := NewSettingsModule()
		const numModifiers = 3
		const numListers = 2
		const modificationsPerModifier = 10

		done := make(chan bool, numModifiers+numListers)

		// Concurrent modifiers
		for i := 0; i < numModifiers; i++ {
			go func(modifierID int) {
				for j := 0; j < modificationsPerModifier; j++ {
					key := fmt.Sprintf("concurrent.key.%d.%d", modifierID, j)
					value := fmt.Sprintf("value-%d-%d", modifierID, j)
					err := module.SetSetting(key, value)
					assert.NoError(t, err)
				}
				done <- true
			}(i)
		}

		// Concurrent listers (should not panic or deadlock)
		for i := 0; i < numListers; i++ {
			go func() {
				for k := 0; k < 20; k++ {
					settings := module.ListSettings()
					assert.NotNil(t, settings)
					count := module.SettingCount()
					assert.GreaterOrEqual(t, count, 0)
					time.Sleep(1 * time.Millisecond)
				}
				done <- true
			}()
		}

		// Wait for all operations
		for i := 0; i < numModifiers+numListers; i++ {
			<-done
		}

		// Final count should be correct
		expectedCount := numModifiers * modificationsPerModifier
		assert.Equal(t, expectedCount, module.SettingCount())
	})

	t.Run("concurrent_save_load", func(t *testing.T) {
		tempDir := t.TempDir()
		settingsPath := filepath.Join(tempDir, "concurrent_settings.json")

		module := NewSettingsModule()
		err := module.SetSettingsPath(settingsPath)
		require.NoError(t, err)

		const numOperations = 5
		done := make(chan bool, numOperations*2)

		// Add initial data
		err = module.SetSetting("initial.key", "initial.value")
		require.NoError(t, err)

		// Concurrent savers
		for i := 0; i < numOperations; i++ {
			go func(saveID int) {
				key := fmt.Sprintf("save.key.%d", saveID)
				value := fmt.Sprintf("save.value.%d", saveID)
				err := module.SetSetting(key, value)
				assert.NoError(t, err)

				err = module.SaveSettings()
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Concurrent loaders (should not corrupt data)
		for i := 0; i < numOperations; i++ {
			go func() {
				time.Sleep(10 * time.Millisecond) // Let some saves happen first
				err := module.LoadSettings()
				// May succeed or fail depending on timing, but should not panic
				_ = err
				done <- true
			}()
		}

		// Wait for all operations
		for i := 0; i < numOperations*2; i++ {
			<-done
		}

		// Final load should succeed
		err = module.LoadSettings()
		assert.NoError(t, err)

		// Initial key should still exist
		value, err := module.GetSetting("initial.key")
		assert.NoError(t, err)
		assert.Equal(t, "initial.value", value)
	})
}

func TestDefaultSettingsPath(t *testing.T) {
	p := DefaultSettingsPath()
	if filepath.Base(p) != "settings.json" || filepath.Base(filepath.Dir(p)) != "recuerdo" {
		t.Errorf("DefaultSettingsPath() = %q, want .../recuerdo/settings.json", p)
	}
}

func TestMigrateSettings(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "old", "settings.json")
	path := filepath.Join(dir, "new", "recuerdo", "settings.json")

	// nothing to copy
	if copied, err := migrateSettings(path, legacy); copied || err != nil {
		t.Fatalf("without old file: copied=%v err=%v", copied, err)
	}

	os.MkdirAll(filepath.Dir(legacy), 0o755)
	os.WriteFile(legacy, []byte(`{"a":1}`), 0o644)
	if copied, err := migrateSettings(path, legacy); !copied || err != nil {
		t.Fatalf("first run: copied=%v err=%v", copied, err)
	}
	if got, _ := os.ReadFile(path); string(got) != `{"a":1}` {
		t.Errorf("copied %q", got)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Errorf("old file should stay: %v", err)
	}

	// an existing new file is never overwritten
	os.WriteFile(path, []byte(`{"b":2}`), 0o644)
	if copied, err := migrateSettings(path, legacy); copied || err != nil {
		t.Fatalf("second run: copied=%v err=%v", copied, err)
	}
	if got, _ := os.ReadFile(path); string(got) != `{"b":2}` {
		t.Errorf("new file overwritten: %q", got)
	}
}
