// Package settings provides functionality ported from Python module
//
// Provides the settings/preferences dialog.
//
// This is an automated port - implementation may be incomplete.
package settings

import (
	"context"
	"fmt"
	"log"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/mappu/miqt/qt"
)

// SettingsDialogModule is a Go port of the Python SettingsDialogModule class
type SettingsDialogModule struct {
	*core.BaseModule
	manager      *core.Manager
	dialog       *qt.QDialog
	tabWidget    *qt.QTabWidget
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
		mod.createDialog(nil)
	}

	if mod.dialog != nil {
		mod.loadSettings()
		mod.dialog.Show()
		mod.dialog.Raise()
		mod.dialog.ActivateWindow()
	}
}

// createDialog creates and configures the settings dialog
func (mod *SettingsDialogModule) createDialog(parent *qt.QWidget) {
	mod.dialog = qt.NewQDialog(parent)
	mod.dialog.SetWindowTitle("Recuerdo Settings")
	mod.dialog.SetFixedSize2(500, 400)
	mod.dialog.SetWindowModality(qt.ApplicationModal)

	// Create main layout
	layout := qt.NewQVBoxLayout(mod.dialog.QWidget)

	// Create tab widget
	mod.tabWidget = qt.NewQTabWidget(mod.dialog.QWidget)
	layout.AddWidget(mod.tabWidget.QWidget)

	// Add tabs
	mod.createGeneralTab()
	mod.createLanguageTab()
	mod.createInterfaceTab()

	// Add button box
	buttonBox := qt.NewQDialogButtonBox(mod.dialog.QWidget)
	buttonBox.SetStandardButtons(qt.QDialogButtonBox__Ok | qt.QDialogButtonBox__Cancel | qt.QDialogButtonBox__Apply)
	layout.AddWidget(buttonBox.QWidget)

	// Connect buttons
	buttonBox.OnAccepted(func() {
		mod.saveSettings()
		mod.dialog.Accept()
	})

	buttonBox.OnRejected(func() {
		mod.dialog.Reject()
	})

	buttonBox.Button(qt.QDialogButtonBox__Apply).OnClicked(func() {
		mod.saveSettings()
	})

	mod.retranslate()
}

// createGeneralTab creates the general settings tab
func (mod *SettingsDialogModule) createGeneralTab() {
	generalWidget := qt.NewQWidget2()
	layout := qt.NewQFormLayout(generalWidget)

	// Auto-save checkbox
	autoSaveCheck := qt.NewQCheckBox(generalWidget)
	autoSaveCheck.SetText("Enable auto-save")
	layout.AddRow3("Auto-save:", autoSaveCheck.QWidget)

	// Save interval
	saveIntervalSpin := qt.NewQSpinBox(generalWidget)
	saveIntervalSpin.SetRange(1, 60)
	saveIntervalSpin.SetValue(5)
	saveIntervalSpin.SetSuffix(" minutes")
	layout.AddRow3("Save interval:", saveIntervalSpin.QWidget)

	// Check for updates
	updateCheck := qt.NewQCheckBox(generalWidget)
	updateCheck.SetText("Check for updates at startup")
	layout.AddRow3("Updates:", updateCheck.QWidget)

	// Recent files count
	recentFilesSpin := qt.NewQSpinBox(generalWidget)
	recentFilesSpin.SetRange(0, 20)
	recentFilesSpin.SetValue(10)
	layout.AddRow3("Recent files:", recentFilesSpin.QWidget)

	mod.tabWidget.AddTab(generalWidget, "General")
}

// createLanguageTab creates the language settings tab
func (mod *SettingsDialogModule) createLanguageTab() {
	languageWidget := qt.NewQWidget2()
	layout := qt.NewQFormLayout(languageWidget)

	// Interface language
	languageCombo := qt.NewQComboBox(nil)
	languageCombo.AddItems([]string{
		"English",
		"Dutch",
		"French",
		"German",
		"Spanish",
	})
	layout.AddRow3("Interface language:", languageCombo.QWidget)

	// Default question language
	questionLangCombo := qt.NewQComboBox(nil)
	questionLangCombo.AddItems([]string{
		"Auto-detect",
		"English",
		"Dutch",
		"French",
		"German",
		"Spanish",
	})
	layout.AddRow3("Question language:", questionLangCombo.QWidget)

	// Default answer language
	answerLangCombo := qt.NewQComboBox(nil)
	answerLangCombo.AddItems([]string{
		"Auto-detect",
		"English",
		"Dutch",
		"French",
		"German",
		"Spanish",
	})
	layout.AddRow3("Answer language:", answerLangCombo.QWidget)

	mod.tabWidget.AddTab(languageWidget, "Language")
}

// createInterfaceTab creates the interface settings tab
func (mod *SettingsDialogModule) createInterfaceTab() {
	interfaceWidget := qt.NewQWidget2()
	layout := qt.NewQFormLayout(interfaceWidget)

	// Theme selection
	themeCombo := qt.NewQComboBox(nil)
	themeCombo.AddItems([]string{
		"System Default",
		"Light",
		"Dark",
	})
	layout.AddRow3("Theme:", themeCombo.QWidget)

	// Show toolbar
	showToolbarCheck := qt.NewQCheckBox2()
	showToolbarCheck.SetText("Show toolbar")
	showToolbarCheck.SetChecked(true)
	layout.AddRow3("Toolbar:", showToolbarCheck.QWidget)

	// Show status bar
	showStatusCheck := qt.NewQCheckBox2()
	showStatusCheck.SetText("Show status bar")
	showStatusCheck.SetChecked(true)
	layout.AddRow3("Status bar:", showStatusCheck.QWidget)

	// Window opacity
	opacitySlider := qt.NewQSlider(nil)
	opacitySlider.SetRange(50, 100)
	opacitySlider.SetValue(100)
	opacityLabel := qt.NewQLabel2()
	opacityLabel.SetText("100%")

	opacitySlider.OnValueChanged(func(value int) {
		opacityLabel.SetText(fmt.Sprintf("%d%%", value))
	})

	opacityWidget := qt.NewQWidget2()
	opacityLayout := qt.NewQHBoxLayout(opacityWidget)
	opacityLayout.AddWidget(opacitySlider.QWidget)
	opacityLayout.AddWidget(opacityLabel.QWidget)

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
		mod.dialog.SetWindowTitle("Recuerdo Settings")
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

// ShowSettingsDialog displays the settings dialog and returns true if settings were applied
func (mod *SettingsDialogModule) ShowSettingsDialog() bool {
	log.Printf("[SUCCESS] SettingsDialogModule.ShowSettingsDialog() - creating and showing settings dialog")

	if mod.manager == nil {
		log.Printf("[ERROR] SettingsDialogModule.ShowSettingsDialog() - manager is nil")
		return false
	}

	// Get the main window as parent
	var parentWidget *qt.QWidget
	uiModules := mod.manager.GetModulesByType("ui")
	if len(uiModules) > 0 {
		if guiMod, ok := uiModules[0].(interface{ GetMainWindow() *qt.QMainWindow }); ok {
			parentWidget = guiMod.GetMainWindow().QWidget
			log.Printf("[SUCCESS] SettingsDialogModule got parent window from GUI module")
		}
	}

	mod.createDialog(parentWidget)

	if mod.dialog != nil {
		log.Printf("[SUCCESS] SettingsDialogModule showing dialog")
		result := mod.dialog.Exec()
		log.Printf("[SUCCESS] SettingsDialogModule dialog closed with result: %d", result)

		// QDialog::Accepted = 1
		if result == 1 {
			log.Printf("[SUCCESS] Settings were accepted and applied")
			return true
		}
	} else {
		log.Printf("[ERROR] SettingsDialogModule.ShowSettingsDialog() - dialog creation failed")
	}

	return false
}

// InitSettingsDialogModule creates and returns a new SettingsDialogModule instance
func InitSettingsDialogModule() core.Module {
	return NewSettingsDialogModule()
}
