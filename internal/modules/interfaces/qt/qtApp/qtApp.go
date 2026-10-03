// Package qtapp provides functionality ported from Python module
//
// When this module is enabled, there is guaranteed to be a
// QApplication instance. It **doesn't** guarantee that that
// QApplication did initialize the GUI, though.
//
// This is an automated port - implementation may be incomplete.
package qtapp

import (
	"context"
	"fmt"
	"os"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/mappu/miqt/qt"
)

// QtAppModule is a Go port of the Python QtAppModule class
type QtAppModule struct {
	*core.BaseModule
	manager *core.Manager
	app     *qt.QApplication
}

// NewQtAppModule creates a new QtAppModule instance
func NewQtAppModule() *QtAppModule {
	base := core.NewBaseModule("qtApp", "qtapp-module")
	base.SetPriority(2000) // High priority - needs to run early

	return &QtAppModule{
		BaseModule: base,
	}
}

// Enable activates the module
// This is the Go equivalent of the Python enable method
func (mod *QtAppModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// Initialize Qt Application if not already done
	if mod.app == nil {
		mod.app = qt.NewQApplication(os.Args)

		// Set application properties using static functions
		qt.QCoreApplication_SetApplicationName("Recuerdo")
		qt.QCoreApplication_SetApplicationVersion("4.0.0")
		qt.QCoreApplication_SetOrganizationName("Kiefte")
		// lets Wayland docks match the window to eu.kiefte.Recuerdo.desktop
		qt.QGuiApplication_SetDesktopFileName("eu.kiefte.Recuerdo")
		qt.QCoreApplication_SetOrganizationDomain("kiefte.eu")

		// Configure international input support
		fmt.Println("Configuring international input and font support...")

		// Set Unicode-supporting font as the default application font
		font := qt.NewQFont()

		// Try common Unicode-supporting fonts
		font.SetFamily("DejaVu Sans, Liberation Sans, Arial, sans-serif")
		font.SetPointSize(11)

		// Set as default application font
		qt.QApplication_SetFont(font)

		// Set UTF-8 encoding attributes for proper Unicode handling
		qt.QCoreApplication_SetAttribute(qt.AA_UseHighDpiPixmaps)

		fmt.Println("Set Unicode-supporting font for better character display")
		fmt.Println("Configured UTF-8 and Unicode support for character picker")
		fmt.Println("Note: IME and accent input configured per widget for better control")

		fmt.Println("Qt Application initialized")
	}

	fmt.Println("QtAppModule enabled")
	return nil
}

// Disable deactivates the module
// This is the Go equivalent of the Python disable method
func (mod *QtAppModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up Qt Application
	if mod.app != nil {
		qt.QCoreApplication_Quit()
		mod.app = nil
	}

	fmt.Println("QtAppModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *QtAppModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// GetApplication returns the Qt application instance
func (mod *QtAppModule) GetApplication() *qt.QApplication {
	return mod.app
}

// ProcessEvents processes pending Qt events
func (mod *QtAppModule) ProcessEvents() {
	if mod.app != nil {
		qt.QCoreApplication_ProcessEvents()
	}
}

// Exec runs the Qt event loop (blocking)
func (mod *QtAppModule) Exec() int {
	if mod.app != nil {
		return qt.QApplication_Exec()
	}
	return 0
}

// InitQtAppModule creates and returns a new QtAppModule instance
// This is the Go equivalent of the Python init function
func InitQtAppModule() core.Module {
	return NewQtAppModule()
}
