// Package recentlyopened keeps the list of recently opened lessons for
// File > Open Recent, in the settings. Port of OpenTeacher's
// logic/recentlyOpened (which kept it in its data store).
package recentlyopened

import (
	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/settingsdefs"
)

// SettingKey is the setting with the list, newest first.
const SettingKey = "org.openteacher.recentlyOpened"

// Max is how many files the list keeps.
const Max = 10

// Settings is what the list needs from the settings module.
type Settings interface {
	GetSettingWithDefault(key string, defaultValue interface{}) interface{}
	SetSetting(key string, value interface{}) error
}

// List returns the recently opened files, newest first.
func List(s Settings) []string {
	var out []string
	switch v := s.GetSettingWithDefault(SettingKey, nil).(type) {
	case []string:
		out = append(out, v...)
	case []interface{}: // as read back from the JSON settings file
		for _, p := range v {
			if str, ok := p.(string); ok && str != "" {
				out = append(out, str)
			}
		}
	}
	return out
}

// Add puts path at the top of the list (once), keeping at most Max.
func Add(s Settings, path string) {
	list := []string{path}
	for _, p := range List(s) {
		if p != path && len(list) < Max {
			list = append(list, p)
		}
	}
	s.SetSetting(SettingKey, list)
}

// Clear empties the list.
func Clear(s Settings) { s.SetSetting(SettingKey, []string{}) }

// RecentlyOpenedModule is OpenTeacher's "recentlyOpened" module.
type RecentlyOpenedModule struct {
	*core.BaseModule
}

// NewRecentlyOpenedModule creates the module.
func NewRecentlyOpenedModule() *RecentlyOpenedModule {
	return &RecentlyOpenedModule{BaseModule: core.NewBaseModule("recentlyOpened", "recentlyopened-module")}
}

// InitRecentlyOpenedModule creates the module.
func InitRecentlyOpenedModule() core.Module { return NewRecentlyOpenedModule() }

// "Clear recent files" in the settings dialog.
func init() {
	settingsdefs.Register(settingsdefs.Def{
		Key: SettingKey, Category: "Files", Name: "Clear recent files",
		Help: "Empty the File > Open Recent list", Kind: settingsdefs.Action,
		Run: func(s settingsdefs.Store) { s.SetSetting(SettingKey, []string{}) },
	})
}
