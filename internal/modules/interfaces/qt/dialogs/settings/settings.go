// Package settings is the settings dialog. As OpenTeacher's, it is made
// from the settings the modules registered (internal/settingsdefs): a tab
// per category and a control per setting, saved when OK is pressed.
package settings

import (
	"context"
	"github.com/LaPingvino/recuerdo/internal/i18n"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/settingsdefs"
	qt "github.com/mappu/miqt/qt6"
)

// Dialog shows settings and saves them on Apply.
type Dialog struct {
	*qt.QDialog
	tabs  *qt.QTabWidget
	store settingsdefs.Store
	apply []func() // saves each control's value
	// the controls, by setting key
	checks  map[string]*qt.QCheckBox
	combos  map[string]*qt.QComboBox
	spins   map[string]*qt.QDoubleSpinBox
	buttons map[string]*qt.QPushButton
}

// NewDialog creates the dialog for defs, showing the values in store.
func NewDialog(parent *qt.QWidget, store settingsdefs.Store, defs []settingsdefs.Def) *Dialog {
	d := &Dialog{QDialog: qt.NewQDialog(parent), store: store, checks: map[string]*qt.QCheckBox{},
		combos: map[string]*qt.QComboBox{}, spins: map[string]*qt.QDoubleSpinBox{}, buttons: map[string]*qt.QPushButton{}}
	d.SetWindowTitle(i18n.T("Settings"))
	d.SetMinimumWidth(460)
	layout := qt.NewQVBoxLayout(d.QWidget)
	d.tabs = qt.NewQTabWidget(d.QWidget)
	layout.AddWidget(d.tabs.QWidget)

	forms := map[string]*qt.QFormLayout{}
	for _, def := range defs {
		form, ok := forms[def.Category]
		if !ok {
			page := qt.NewQWidget(nil)
			form = qt.NewQFormLayout(page)
			forms[def.Category] = form
			d.tabs.AddTab(page, i18n.T(def.Category))
		}
		d.addControl(form, def)
	}
	if len(defs) == 0 {
		layout.AddWidget(qt.NewQLabel3(i18n.T("There are no settings to change.")).QWidget)
	}

	buttons := qt.NewQDialogButtonBox(d.QWidget)
	buttons.SetStandardButtons(qt.QDialogButtonBox__Ok | qt.QDialogButtonBox__Cancel)
	buttons.OnAccepted(func() {
		d.Apply()
		d.Accept()
	})
	buttons.OnRejected(func() { d.Reject() })
	layout.AddWidget(buttons.QWidget)
	return d
}

func (d *Dialog) addControl(form *qt.QFormLayout, def settingsdefs.Def) {
	var w *qt.QWidget
	switch def.Kind {
	case settingsdefs.Bool:
		c := qt.NewQCheckBox3(i18n.T(def.Name))
		d.checks[def.Key] = c
		c.SetChecked(def.Value(d.store).(bool))
		d.apply = append(d.apply, func() { d.store.SetSetting(def.Key, c.IsChecked()) })
		form.AddRowWithWidget(c.QWidget)
		w = c.QWidget
	case settingsdefs.Choice:
		c := qt.NewQComboBox(nil)
		d.combos[def.Key] = c
		labels := def.Labels
		if len(labels) != len(def.Choices) {
			labels = def.Choices
		}
		labels = append([]string(nil), labels...)
		for i, l := range labels {
			labels[i] = i18n.T(l) // "System language"; values such as notations stay as they are
		}
		c.AddItems(labels)
		value := def.Value(d.store).(string)
		for i, v := range def.Choices {
			if v == value {
				c.SetCurrentIndex(i)
			}
		}
		d.apply = append(d.apply, func() {
			if i := c.CurrentIndex(); i >= 0 && i < len(def.Choices) {
				d.store.SetSetting(def.Key, def.Choices[i])
			}
		})
		form.AddRow3(i18n.T(def.Name)+":", c.QWidget)
		w = c.QWidget
	case settingsdefs.Seconds:
		s := qt.NewQDoubleSpinBox(nil)
		d.spins[def.Key] = s
		s.SetRange(def.Min, def.Max)
		s.SetSingleStep(0.5)
		s.SetDecimals(1)
		s.SetSuffix(" s")
		s.SetValue(def.Value(d.store).(float64) / 1000)
		d.apply = append(d.apply, func() { d.store.SetSetting(def.Key, int64(s.Value()*1000)) })
		form.AddRow3(i18n.T(def.Name)+":", s.QWidget)
		w = s.QWidget
	case settingsdefs.Action:
		b := qt.NewQPushButton3(i18n.T(def.Name))
		d.buttons[def.Key] = b
		b.OnClicked(func() {
			if def.Run != nil {
				def.Run(d.store)
			}
			b.SetEnabled(false)
			b.SetText(i18n.T(def.Name) + " ✔")
		})
		row := qt.NewQHBoxLayout2()
		row.AddWidget(b.QWidget)
		row.AddStretch()
		form.AddRowWithLayout(row.QLayout)
		w = b.QWidget
	default:
		return
	}
	if def.Help != "" {
		w.SetToolTip(i18n.T(def.Help))
		help := qt.NewQLabel3(i18n.T(def.Help))
		help.SetWordWrap(true)
		font := help.Font()
		font.SetPointSizeF(font.PointSizeF() * 0.9)
		help.SetFont(font)
		help.SetEnabled(false) // greyed: a description, not a control
		form.AddRowWithWidget(help.QWidget)
	}
}

// Apply saves the values of the controls (actions run when clicked).
func (d *Dialog) Apply() {
	for _, f := range d.apply {
		f()
	}
}

// SettingsDialogModule offers the dialog to the GUI.
type SettingsDialogModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewSettingsDialogModule creates the module.
func NewSettingsDialogModule() *SettingsDialogModule {
	base := core.NewBaseModule("settingsDialog", "settings-dialog-module")
	base.SetRequires("qtApp", "settings")
	return &SettingsDialogModule{BaseModule: base}
}

// ShowSettingsDialog shows the dialog over the main window and reports
// whether settings were saved.
func (mod *SettingsDialogModule) ShowSettingsDialog() bool {
	if mod.manager == nil {
		return false
	}
	sm, found := mod.manager.GetDefaultModule("settings")
	store, ok := sm.(settingsdefs.Store)
	if !found || !ok {
		return false
	}
	var parent *qt.QWidget
	for _, ui := range mod.manager.GetModulesByType("ui") {
		if g, ok := ui.(interface{ GetMainWindow() *qt.QMainWindow }); ok {
			parent = g.GetMainWindow().QWidget
		}
	}
	d := NewDialog(parent, store, settingsdefs.All())
	defer d.Delete()
	return d.Exec() == int(qt.QDialog__Accepted)
}

// Enable activates the module.
func (mod *SettingsDialogModule) Enable(ctx context.Context) error { return mod.BaseModule.Enable(ctx) }

// Disable deactivates the module.
func (mod *SettingsDialogModule) Disable(ctx context.Context) error {
	return mod.BaseModule.Disable(ctx)
}

// SetManager sets the module manager.
func (mod *SettingsDialogModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitSettingsDialogModule creates the module.
func InitSettingsDialogModule() core.Module { return NewSettingsDialogModule() }
