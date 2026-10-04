// Package about provides functionality ported from Python module
//
// Provides the about dialog.
//
// This is an automated port - implementation may be incomplete.
package about

import (
	"strings"

	"context"
	"fmt"
	openteacherauthors "github.com/LaPingvino/recuerdo/internal/modules/data/openteacherAuthors"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/icon"
	"github.com/LaPingvino/recuerdo/internal/version"
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
	mod.dialog.SetWindowModality(qt.ApplicationModal)
	mod.dialog.SetMinimumWidth(440)

	layout := qt.NewQVBoxLayout(mod.dialog.QWidget)
	layout.SetSpacing(8)
	// grow to fit the wrapped text instead of clipping it
	layout.SetSizeConstraint(qt.QLayout__SetMinimumSize)
	layout.SetContentsMargins(24, 20, 24, 16)
	label := func(text string) *qt.QLabel {
		l := qt.NewQLabel(mod.dialog.QWidget)
		l.SetText(text)
		l.SetAlignment(qt.AlignHCenter)
		layout.AddWidget(l.QWidget)
		return l
	}

	logo := label("")
	logo.SetPixmap(icon.Pixmap(96))
	logo.SetMinimumHeight(100)

	title := label("Recuerdo")
	titleFont := title.Font()
	titleFont.SetPointSize(20)
	titleFont.SetBold(true)
	title.SetFont(titleFont)

	label("Version " + version.Version).SetStyleSheet("color: palette(dark);")
	layout.AddSpacing(6)
	label("Learn words, places and more by heart:\na foreign language's vocabulary, topography,\nor anything else you want to remember.")
	layout.AddSpacing(6)
	label("© 2025–2026 Joop Kiefte\nBased on OpenTeacher by the OpenTeacher Team, 2010–2023")
	license := label("Free software under the GNU General Public License,\nversion 3 or later")
	license.SetStyleSheet("color: palette(dark);")
	links := label(`<a href="https://github.com/LaPingvino/recuerdo">github.com/LaPingvino/recuerdo</a> · <a href="https://openteacher.org">openteacher.org</a>`)
	links.SetOpenExternalLinks(true)

	layout.AddSpacing(8)
	buttonBox := qt.NewQDialogButtonBox(mod.dialog.QWidget)
	buttonBox.SetStandardButtons(qt.QDialogButtonBox__Close)
	credits := buttonBox.AddButton2("Credits", qt.QDialogButtonBox__ActionRole)
	credits.OnClicked(func() {
		qt.QMessageBox_Information(mod.dialog.QWidget, "Credits", CreditsText())
	})
	layout.AddWidget(buttonBox.QWidget)
	buttonBox.OnRejected(func() {
		mod.dialog.Close()
	})
	mod.dialog.AdjustSize()

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

// CreditsText lists who made Recuerdo and OpenTeacher, by role.
func CreditsText() string {
	var b strings.Builder
	b.WriteString("Recuerdo: Joop Kiefte\n\nBased on OpenTeacher by:\n")
	roles, names := openteacherauthors.ByRole()
	for _, r := range roles {
		b.WriteString("\n" + r + ": " + strings.Join(names[r], ", "))
	}
	return b.String()
}
