// Package about provides functionality ported from Python module
//
// Provides the about dialog.
//
// This is an automated port - implementation may be incomplete.
package about

import (
	"context"
	"fmt"

	"github.com/LaPingvino/openteacher/internal/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// AboutDialogModule is a Go port of the Python AboutDialogModule class
type AboutDialogModule struct {
	*core.BaseModule
	manager *core.Manager
	dialog  *widgets.QDialog
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
		mod.createDialog()
	}

	if mod.dialog != nil {
		mod.dialog.Show()
		mod.dialog.Raise()
		mod.dialog.ActivateWindow()
	}
}

// createDialog creates and configures the about dialog
func (mod *AboutDialogModule) createDialog() {
	mod.dialog = widgets.NewQDialog(nil, 0)
	mod.dialog.SetWindowTitle("About OpenTeacher")
	mod.dialog.SetFixedSize2(400, 300)
	mod.dialog.SetWindowModality(qtcore.Qt__ApplicationModal)

	// Create main layout
	layout := widgets.NewQVBoxLayout()
	mod.dialog.SetLayout(layout)

	// Add OpenTeacher logo/title
	titleLabel := widgets.NewQLabel2("OpenTeacher", nil, 0)
	titleFont := titleLabel.Font()
	titleFont.SetPointSize(18)
	titleFont.SetBold(true)
	titleLabel.SetFont(titleFont)
	titleLabel.SetAlignment(qtcore.Qt__AlignHCenter)
	layout.AddWidget(titleLabel, 0, 0)

	// Add version info
	versionLabel := widgets.NewQLabel2("Version 4.0.0-alpha", nil, 0)
	versionLabel.SetAlignment(qtcore.Qt__AlignHCenter)
	layout.AddWidget(versionLabel, 0, 0)

	// Add description
	descLabel := widgets.NewQLabel2("OpenTeacher helps you learn whatever you want to learn!\nIt's designed to help you learn a foreign language,\nbut can also be used for other subjects.", nil, 0)
	descLabel.SetAlignment(qtcore.Qt__AlignHCenter)
	descLabel.SetWordWrap(true)
	layout.AddWidget(descLabel, 0, 0)

	// Add copyright
	copyrightLabel := widgets.NewQLabel2("Copyright © 2010-2023 OpenTeacher Team", nil, 0)
	copyrightLabel.SetAlignment(qtcore.Qt__AlignHCenter)
	layout.AddWidget(copyrightLabel, 0, 0)

	// Add website link
	websiteLabel := widgets.NewQLabel2(`<a href="http://openteacher.org">http://openteacher.org</a>`, nil, 0)
	websiteLabel.SetAlignment(qtcore.Qt__AlignHCenter)
	websiteLabel.SetOpenExternalLinks(true)
	layout.AddWidget(websiteLabel, 0, 0)

	// Add spacer
	layout.AddStretch(1)

	// Add close button
	buttonBox := widgets.NewQDialogButtonBox3(widgets.QDialogButtonBox__Close, nil)
	layout.AddWidget(buttonBox, 0, 0)

	// Connect close button
	buttonBox.ConnectRejected(func() {
		mod.dialog.Close()
	})

	mod.retranslate()
}

// retranslate updates dialog text for localization
func (mod *AboutDialogModule) retranslate() {
	if mod.dialog != nil {
		mod.dialog.SetWindowTitle("About OpenTeacher")
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

// InitAboutDialogModule creates and returns a new AboutDialogModule instance
func InitAboutDialogModule() core.Module {
	return NewAboutDialogModule()
}
