// Package about provides functionality ported from Python module
//
// Provides the about dialog.
//
// This is an automated port - implementation may be incomplete.
package about

import (
	"context"
	"fmt"
	"log"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/mappu/miqt/qt"
)

// AboutDialogModule is a Go port of the Python AboutDialogModule class
type AboutDialogModule struct {
	*core.BaseModule
	manager *core.Manager
	dialog  *qt.QDialog
}

// NewAboutDialogModule creates a new AboutDialogModule instance
func NewAboutDialogModule() *AboutDialogModule {
	base := core.NewBaseModule("aboutDialog", "about-module")
	base.SetRequires("qtApp")

	return &AboutDialogModule{
		BaseModule: base,
	}
}

// Show displays the about dialog
func (mod *AboutDialogModule) Show() {
	if mod.dialog == nil {
		mod.createDialog(nil)
	}

	if mod.dialog != nil {
		mod.dialog.Show()
		mod.dialog.Raise()
		mod.dialog.ActivateWindow()
	}
}

// createDialog creates and configures the about dialog
func (mod *AboutDialogModule) createDialog(parent *qt.QWidget) {
	mod.dialog = qt.NewQDialog(parent)
	mod.dialog.SetWindowTitle("About Recuerdo")
	mod.dialog.SetFixedSize2(400, 300)
	mod.dialog.SetWindowModality(qt.ApplicationModal)

	// Create main layout
	layout := qt.NewQVBoxLayout(mod.dialog.QWidget)

	// Add Recuerdo logo/title
	titleLabel := qt.NewQLabel(mod.dialog.QWidget)
	titleLabel.SetText("Recuerdo")
	titleFont := titleLabel.Font()
	titleFont.SetPointSize(18)
	titleFont.SetBold(true)
	titleLabel.SetFont(titleFont)
	titleLabel.SetAlignment(qt.AlignHCenter)
	layout.AddWidget(titleLabel.QWidget)

	// Add version info
	versionLabel := qt.NewQLabel(mod.dialog.QWidget)
	versionLabel.SetText("Version 4.0.0-alpha")
	versionLabel.SetAlignment(qt.AlignHCenter)
	layout.AddWidget(versionLabel.QWidget)

	// Add description
	descLabel := qt.NewQLabel(mod.dialog.QWidget)
	descLabel.SetText("Recuerdo helps you learn whatever you want to learn!\nIt's designed to help you learn a foreign language,\nbut can also be used for other subjects.")
	descLabel.SetAlignment(qt.AlignHCenter)
	descLabel.SetWordWrap(true)
	layout.AddWidget(descLabel.QWidget)

	// Add copyright
	copyrightLabel := qt.NewQLabel(mod.dialog.QWidget)
	copyrightLabel.SetText("Copyright © 2025 Joop Kiefte\nBased on OpenTeacher © 2010-2023 OpenTeacher Team")
	copyrightLabel.SetAlignment(qt.AlignHCenter)
	layout.AddWidget(copyrightLabel.QWidget)

	// Add website link
	websiteLabel := qt.NewQLabel(mod.dialog.QWidget)
	websiteLabel.SetText(`<a href="http://openteacher.org">http://openteacher.org</a>`)
	websiteLabel.SetAlignment(qt.AlignHCenter)
	websiteLabel.SetOpenExternalLinks(true)
	layout.AddWidget(websiteLabel.QWidget)

	// Add spacer
	layout.AddStretch()

	// Add close button
	buttonBox := qt.NewQDialogButtonBox(mod.dialog.QWidget)
	buttonBox.SetStandardButtons(qt.QDialogButtonBox__Close)
	layout.AddWidget(buttonBox.QWidget)

	// Connect close button
	buttonBox.OnRejected(func() {
		mod.dialog.Close()
	})

	mod.retranslate()
}

// retranslate updates dialog text for localization
func (mod *AboutDialogModule) retranslate() {
	if mod.dialog != nil {
		mod.dialog.SetWindowTitle("About Recuerdo")
	}
}

// Enable activates the module
func (mod *AboutDialogModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("AboutDialogModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *AboutDialogModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up dialog
	if mod.dialog != nil {
		mod.dialog.Close()
		mod.dialog = nil
	}

	fmt.Println("AboutDialogModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *AboutDialogModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// ShowAboutDialog displays the about dialog
func (mod *AboutDialogModule) ShowAboutDialog() {
	log.Printf("[SUCCESS] AboutDialogModule.ShowAboutDialog() - creating and showing about dialog")

	if mod.manager == nil {
		log.Printf("[ERROR] AboutDialogModule.ShowAboutDialog() - manager is nil")
		return
	}

	// Get the main window as parent
	var parentWidget *qt.QWidget
	uiModules := mod.manager.GetModulesByType("ui")
	if len(uiModules) > 0 {
		if guiMod, ok := uiModules[0].(interface{ GetMainWindow() *qt.QMainWindow }); ok {
			parentWidget = guiMod.GetMainWindow().QWidget
			log.Printf("[SUCCESS] AboutDialogModule got parent window from GUI module")
		}
	}

	mod.createDialog(parentWidget)

	if mod.dialog != nil {
		log.Printf("[SUCCESS] AboutDialogModule showing dialog")
		mod.dialog.Exec()
		log.Printf("[SUCCESS] AboutDialogModule dialog closed")
	} else {
		log.Printf("[ERROR] AboutDialogModule.ShowAboutDialog() - dialog creation failed")
	}
}

// InitAboutDialogModule creates and returns a new AboutDialogModule instance
// This is the Go equivalent of the Python init function
func InitAboutDialogModule() core.Module {
	return NewAboutDialogModule()
}
