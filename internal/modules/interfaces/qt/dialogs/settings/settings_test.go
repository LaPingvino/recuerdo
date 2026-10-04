package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessons/words" // registers its settings
	recentlyopened "github.com/LaPingvino/recuerdo/internal/modules/logic/recentlyOpened"
	"github.com/LaPingvino/recuerdo/internal/settingsdefs"
	"github.com/mappu/miqt/qt"
)

type mapStore map[string]interface{}

func (m mapStore) GetSettingWithDefault(key string, def interface{}) interface{} {
	if v, ok := m[key]; ok {
		return v
	}
	return def
}
func (m mapStore) SetSetting(key string, v interface{}) error { m[key] = v; return nil }

var problem error

// Qt on the main thread: the dialog is driven in TestMain.
func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"settings-test"})
	problem = drive()
	os.Exit(m.Run())
}

const (
	notation  = "org.openteacher.noteCalculatorChooser.noteCalculator"
	pronounce = "org.openteacher.ttsProviders.words.pronounce"
	repeat    = "org.openteacher.teachTypes.repeatAnswer.fadeDuration"
)

func drive() error {
	defs := settingsdefs.All()
	if len(defs) != 4 {
		return fmt.Errorf("%d settings registered, want notation, pronounce, repeat and clear recent", len(defs))
	}
	store := mapStore{repeat: 2500.0, recentlyopened.SettingKey: []string{"/a.otwd"}}
	d := NewDialog(nil, store, defs)
	if d.tabs.Count() != 3 || d.tabs.TabText(0) != "Practice" || d.tabs.TabText(1) != "Results" || d.tabs.TabText(2) != "Files" {
		return fmt.Errorf("tabs: %d", d.tabs.Count())
	}
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		d.Show()
		for i := 0; i < d.tabs.Count(); i++ {
			d.tabs.SetCurrentIndex(i)
			qt.QCoreApplication_ProcessEvents()
			d.Grab().Save(filepath.Join(dir, fmt.Sprintf("after-%d.png", i)))
		}
	}
	// the controls show the stored values (and defaults)
	combo, spin, check := d.combos[notation], d.spins[repeat], d.checks[pronounce]
	if spin.Value() != 2.5 || check.IsChecked() || combo.CurrentText() == "" {
		return fmt.Errorf("loaded: %v s, pronounce %v, notation %q", spin.Value(), check.IsChecked(), combo.CurrentText())
	}
	// Cancel saves nothing
	spin.SetValue(4)
	d.Reject()
	if store[repeat] != 2500.0 || store[pronounce] != nil {
		return fmt.Errorf("cancel saved: %v", store)
	}
	// OK saves every control
	check.SetChecked(true)
	combo.SetCurrentIndex(combo.Count() - 1)
	want := combo.CurrentText()
	d.Apply()
	if store[repeat] != int64(4000) || store[pronounce] != true || store[notation] != want {
		return fmt.Errorf("saved: %v", store)
	}
	// the action runs at once
	clear := d.buttons[recentlyopened.SettingKey]
	clear.Click()
	if l, ok := store[recentlyopened.SettingKey].([]string); !ok || len(l) != 0 || clear.IsEnabled() {
		return fmt.Errorf("clear recent: %v", store[recentlyopened.SettingKey])
	}
	return nil
}

func TestSettingsDialog(t *testing.T) {
	if problem != nil {
		t.Fatal(problem)
	}
}
