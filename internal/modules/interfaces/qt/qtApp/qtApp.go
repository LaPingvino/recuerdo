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

	"github.com/LaPingvino/openteacher/internal/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// QtAppModule is a Go port of the Python QtAppModule class
type QtAppModule struct {
	*core.BaseModule
	manager *core.Manager
	app     *widgets.QApplication
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
		mod.app = widgets.NewQApplication(len(os.Args), os.Args)

		// Set application properties
		mod.app.SetApplicationName("OpenTeacher")
		mod.app.SetApplicationVersion("4.0.0")
		mod.app.SetOrganizationName("OpenTeacher")
		mod.app.SetOrganizationDomain("openteacher.org")

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
		mod.app.Quit()
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
func (mod *QtAppModule) GetApplication() *widgets.QApplication {
	return mod.app
}

// ProcessEvents processes pending Qt events
func (mod *QtAppModule) ProcessEvents() {
	if mod.app != nil {
		mod.app.ProcessEvents2(qtcore.QEventLoop__AllEvents, 0)
	}
}

// Exec runs the Qt event loop (blocking)
func (mod *QtAppModule) Exec() int {
	if mod.app != nil {
		return mod.app.Exec()
	}
	return 0
}

// InitQtAppModule creates and returns a new QtAppModule instance
// This is the Go equivalent of the Python init function
func InitQtAppModule() core.Module {
	return NewQtAppModule()
}
