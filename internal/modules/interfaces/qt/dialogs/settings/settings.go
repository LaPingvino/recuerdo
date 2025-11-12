// Package settings provides functionality ported from Python module
//
// Provides the settings/preferences dialog.
//
// This is an automated port - implementation may be incomplete.
package settings

import (
	"context"
	"fmt"

	"github.com/LaPingvino/openteacher/internal/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// SettingsDialogModule is a Go port of the Python SettingsDialogModule class
type SettingsDialogModule struct {
	*core.BaseModule
	manager      *core.Manager
	dialog       *widgets.QDialog
	tabWidget    *widgets.QTabWidget
	settingsData map[string]interface{}
}

// NewSettingsDialogModule creates a new SettingsDialogModule instance
func NewSettingsDialogModule() *SettingsDialogModule {
	base := core.NewBaseModule("settingsDialog", "settings-dialog-module")
	base.SetRequires("qtApp", "settings")

	return &SettingsDialogModule{
		BaseModule:   base,
		settingsData: make(map[string]interface{}),
	}
}

// Show displays the settings dialog
func (mod *SettingsDialogModule) Show() {
	if mod.dialog == nil {
		mod.createDialog()
	}

	if mod.dialog != nil {
		mod.loadSettings()
		mod.dialog.Show()
		mod.dialog.Raise()
		mod.dialog.ActivateWindow()
	}
}

// createDialog creates and configures the settings dialog
func (mod *SettingsDialogModule) createDialog() {
	mod.dialog = widgets.NewQDialog(nil, 0)
	mod.dialog.SetWindowTitle("OpenTeacher Settings")
	mod.dialog.SetFixedSize2(500, 400)
	mod.dialog.SetWindowModality(qtcore.Qt__ApplicationModal)

	// Create main layout
	layout := widgets.NewQVBoxLayout()
	mod.dialog.SetLayout(layout)

	// Create tab widget
	mod.tabWidget = widgets.NewQTabWidget(nil)
	layout.AddWidget(mod.tabWidget, 1, 0)

	// Add tabs
	mod.createGeneralTab()
	mod.createLanguageTab()
	mod.createInterfaceTab()

	// Add button box
	buttonBox := widgets.NewQDialogButtonBox3(
		widgets.QDialogButtonBox__Ok|widgets.QDialogButtonBox__Cancel|widgets.QDialogButtonBox__Apply,
		nil)
	layout.AddWidget(buttonBox, 0, 0)

	// Connect buttons
	buttonBox.ConnectAccepted(func() {
		mod.saveSettings()
		mod.dialog.Accept()
	})

	buttonBox.ConnectRejected(func() {
		mod.dialog.Reject()
	})

	buttonBox.Button(widgets.QDialogButtonBox__Apply).ConnectClicked(func(checked bool) {
		mod.saveSettings()
	})

	mod.retranslate()
}

// createGeneralTab creates the general settings tab
func (mod *SettingsDialogModule) createGeneralTab() {
	generalWidget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQFormLayout(nil)
	generalWidget.SetLayout(layout)

	// Auto-save checkbox
	autoSaveCheck := widgets.NewQCheckBox2("Enable auto-save", nil)
	layout.AddRow3("Auto-save:", autoSaveCheck)

	// Save interval
	saveIntervalSpin := widgets.NewQSpinBox(nil)
	saveIntervalSpin.SetRange(1, 60)
	saveIntervalSpin.SetValue(5)
	saveIntervalSpin.SetSuffix(" minutes")
	layout.AddRow3("Save interval:", saveIntervalSpin)

	// Check for updates
	updateCheck := widgets.NewQCheckBox2("Check for updates at startup", nil)
	layout.AddRow3("Updates:", updateCheck)

	// Recent files count
	recentFilesSpin := widgets.NewQSpinBox(nil)
	recentFilesSpin.SetRange(0, 20)
	recentFilesSpin.SetValue(10)
	layout.AddRow3("Recent files:", recentFilesSpin)

	mod.tabWidget.AddTab(generalWidget, "General")
}

// createLanguageTab creates the language settings tab
func (mod *SettingsDialogModule) createLanguageTab() {
	languageWidget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQFormLayout(nil)
	languageWidget.SetLayout(layout)

	// Interface language
	languageCombo := widgets.NewQComboBox(nil)
	languageCombo.AddItems([]string{
		"English",
		"Dutch",
		"French",
		"German",
		"Spanish",
	})
	layout.AddRow3("Interface language:", languageCombo)

	// Default question language
	questionLangCombo := widgets.NewQComboBox(nil)
	questionLangCombo.AddItems([]string{
		"Auto-detect",
		"English",
		"Dutch",
		"French",
		"German",
		"Spanish",
	})
	layout.AddRow3("Default question language:", questionLangCombo)

	// Default answer language
	answerLangCombo := widgets.NewQComboBox(nil)
	answerLangCombo.AddItems([]string{
		"Auto-detect",
		"English",
		"Dutch",
		"French",
		"German",
		"Spanish",
	})
	layout.AddRow3("Default answer language:", answerLangCombo)

	mod.tabWidget.AddTab(languageWidget, "Language")
}

// createInterfaceTab creates the interface settings tab
func (mod *SettingsDialogModule) createInterfaceTab() {
	interfaceWidget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQFormLayout(nil)
	interfaceWidget.SetLayout(layout)

	// Theme selection
	themeCombo := widgets.NewQComboBox(nil)
	themeCombo.AddItems([]string{
		"System Default",
		"Light",
		"Dark",
	})
	layout.AddRow3("Theme:", themeCombo)

	// Show toolbar
	showToolbarCheck := widgets.NewQCheckBox2("Show toolbar", nil)
	showToolbarCheck.SetChecked(true)
	layout.AddRow3("Toolbar:", showToolbarCheck)

	// Show status bar
	showStatusCheck := widgets.NewQCheckBox2("Show status bar", nil)
	showStatusCheck.SetChecked(true)
	layout.AddRow3("Status bar:", showStatusCheck)

	// Window opacity
	opacitySlider := widgets.NewQSlider(nil)
	opacitySlider.SetRange(50, 100)
	opacitySlider.SetValue(100)
	opacityLabel := widgets.NewQLabel2("100%", nil, 0)

	opacitySlider.ConnectValueChanged(func(value int) {
		opacityLabel.SetText(fmt.Sprintf("%d%%", value))
	})

	opacityWidget := widgets.NewQWidget(nil, 0)
	opacityLayout := widgets.NewQHBoxLayout()
	opacityWidget.SetLayout(opacityLayout)
	opacityLayout.AddWidget(opacitySlider, 1, 0)
	opacityLayout.AddWidget(opacityLabel, 0, 0)

	layout.AddRow3("Window opacity:", opacityWidget)

	mod.tabWidget.AddTab(interfaceWidget, "Interface")
}

// loadSettings loads current settings into the dialog
func (mod *SettingsDialogModule) loadSettings() {
	// TODO: Load actual settings from settings module
	fmt.Println("Loading settings...")
}

// saveSettings saves the dialog settings
func (mod *SettingsDialogModule) saveSettings() {
	// TODO: Save settings to settings module
	fmt.Println("Saving settings...")
}

// retranslate updates dialog text for localization
func (mod *SettingsDialogModule) retranslate() {
	if mod.dialog != nil {
		mod.dialog.SetWindowTitle("OpenTeacher Settings")
	}
}

// Enable activates the module
func (mod *SettingsDialogModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("SettingsDialogModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *SettingsDialogModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up dialog
	if mod.dialog != nil {
		mod.dialog.Close()
		mod.dialog = nil
	}

	fmt.Println("SettingsDialogModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *SettingsDialogModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitSettingsDialogModule creates and returns a new SettingsDialogModule instance
func InitSettingsDialogModule() core.Module {
	return NewSettingsDialogModule()
}
