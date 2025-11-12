// Package gui provides functionality ported from Python module
//
// Package gui provides functionality ported from Python module
// legacy/modules/org/openteacher/interfaces/qt/gui/gui.py
//
// This is an automated port - implementation may be incomplete.
//
// This is an automated port - implementation may be incomplete.
package gui

import (
	"context"
	"fmt"

	"github.com/LaPingvino/openteacher/internal/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// GuiModule is a Go port of the Python GuiModule class
type GuiModule struct {
	*core.BaseModule
	manager    *core.Manager
	mainWindow *widgets.QMainWindow
	app        *widgets.QApplication
	menuBar    *widgets.QMenuBar
	statusBar  *widgets.QStatusBar
}

// NewGuiModule creates a new GuiModule instance
func NewGuiModule() *GuiModule {
	base := core.NewBaseModule("ui", "gui-module")
	base.SetRequires("event", "qtApp")

	return &GuiModule{
		BaseModule: base,
	}
}

// Enable activates the module
// This is the Go equivalent of the Python enable method
func (mod *GuiModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// Get Qt application from qtApp module (don't create our own)
	qtAppModule, exists := mod.manager.GetDefaultModule("qtApp")
	if !exists {
		return fmt.Errorf("qtApp module not found")
	}

	// Access the QApplication through interface
	if qtMod, ok := qtAppModule.(interface{ GetApplication() *widgets.QApplication }); ok {
		mod.app = qtMod.GetApplication()
	} else {
		return fmt.Errorf("qtApp module does not provide GetApplication method")
	}

	// Create main window
	mod.mainWindow = widgets.NewQMainWindow(nil, 0)
	mod.mainWindow.SetWindowTitle("OpenTeacher 4.0")
	mod.mainWindow.Resize2(1000, 700)
	mod.mainWindow.SetMinimumSize2(800, 600)

	// Create menu bar
	mod.createMenuBar()

	// Create status bar
	mod.statusBar = mod.mainWindow.StatusBar()
	mod.statusBar.ShowMessage("Ready", 0)

	// Create central widget with basic layout
	centralWidget := widgets.NewQWidget(nil, 0)
	mod.mainWindow.SetCentralWidget(centralWidget)

	// Create main layout
	mainLayout := widgets.NewQVBoxLayout()
	centralWidget.SetLayout(mainLayout)

	// Add welcome area
	welcomeWidget := mod.createWelcomeWidget()
	mainLayout.AddWidget(welcomeWidget, 1, 0)

	// Show the window
	mod.mainWindow.Show()

	fmt.Println("GuiModule enabled - Main window created")
	return nil
}

// Disable deactivates the module
// This is the Go equivalent of the Python disable method
func (mod *GuiModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up GUI resources
	if mod.mainWindow != nil {
		// Safely close the main window
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Warning: Error closing main window: %v\n", r)
			}
		}()
		mod.mainWindow.Close()
		mod.mainWindow = nil
	}

	// Don't quit the app - that's managed by qtApp module
	mod.app = nil

	fmt.Println("GuiModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *GuiModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// ShowMainWindow shows the main application window
func (mod *GuiModule) ShowMainWindow() {
	if mod.mainWindow != nil {
		mod.mainWindow.Show()
		mod.mainWindow.Raise()
		mod.mainWindow.ActivateWindow()
	}
}

// GetMainWindow returns the main window widget
func (mod *GuiModule) GetMainWindow() *widgets.QMainWindow {
	return mod.mainWindow
}

// RunEventLoop starts the Qt event loop (blocking call)
func (mod *GuiModule) RunEventLoop() int {
	if mod.app != nil {
		return mod.app.Exec()
	}
	return 0
}

// createMenuBar creates the main menu bar
func (mod *GuiModule) createMenuBar() {
	mod.menuBar = mod.mainWindow.MenuBar()

	// File menu
	fileMenu := mod.menuBar.AddMenu2("&File")

	newAction := fileMenu.AddAction("&New Lesson...")
	newAction.SetShortcut(gui.NewQKeySequence2("Ctrl+N", gui.QKeySequence__NativeText))
	newAction.ConnectTriggered(func(checked bool) {
		mod.showNewLessonDialog()
	})

	openAction := fileMenu.AddAction("&Open...")
	openAction.SetShortcut(gui.NewQKeySequence2("Ctrl+O", gui.QKeySequence__NativeText))
	openAction.ConnectTriggered(func(checked bool) {
		mod.showOpenDialog()
	})

	fileMenu.AddSeparator()

	saveAction := fileMenu.AddAction("&Save")
	saveAction.SetShortcut(gui.NewQKeySequence2("Ctrl+S", gui.QKeySequence__NativeText))
	saveAction.SetEnabled(false) // Enable when lesson is loaded

	saveAsAction := fileMenu.AddAction("Save &As...")
	saveAsAction.SetShortcut(gui.NewQKeySequence2("Ctrl+Shift+S", gui.QKeySequence__NativeText))
	saveAsAction.SetEnabled(false)

	fileMenu.AddSeparator()

	exitAction := fileMenu.AddAction("E&xit")
	exitAction.SetShortcut(gui.NewQKeySequence2("Ctrl+Q", gui.QKeySequence__NativeText))
	exitAction.ConnectTriggered(func(checked bool) {
		mod.mainWindow.Close()
	})

	// Edit menu
	editMenu := mod.menuBar.AddMenu2("&Edit")

	propertiesAction := editMenu.AddAction("&Properties...")
	propertiesAction.ConnectTriggered(func(checked bool) {
		mod.showPropertiesDialog()
	})

	// Tools menu
	toolsMenu := mod.menuBar.AddMenu2("&Tools")

	settingsAction := toolsMenu.AddAction("&Settings...")
	settingsAction.ConnectTriggered(func(checked bool) {
		mod.showSettingsDialog()
	})

	// Help menu
	helpMenu := mod.menuBar.AddMenu2("&Help")

	aboutAction := helpMenu.AddAction("&About...")
	aboutAction.ConnectTriggered(func(checked bool) {
		mod.showAboutDialog()
	})
}

// createWelcomeWidget creates the main welcome widget
func (mod *GuiModule) createWelcomeWidget() *widgets.QWidget {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	widget.SetLayout(layout)

	// Add some spacing
	layout.AddStretch(1)

	// Main title
	titleLabel := widgets.NewQLabel2("Welcome to OpenTeacher 4.0!", nil, 0)
	titleFont := titleLabel.Font()
	titleFont.SetPointSize(24)
	titleFont.SetBold(true)
	titleLabel.SetFont(titleFont)
	titleLabel.SetAlignment(qtcore.Qt__AlignHCenter)
	layout.AddWidget(titleLabel, 0, 0)

	// Subtitle
	subtitleLabel := widgets.NewQLabel2("Learn whatever you want to learn", nil, 0)
	subtitleFont := subtitleLabel.Font()
	subtitleFont.SetPointSize(14)
	subtitleLabel.SetFont(subtitleFont)
	subtitleLabel.SetAlignment(qtcore.Qt__AlignHCenter)
	layout.AddWidget(subtitleLabel, 0, 0)

	// Add some spacing
	layout.AddSpacing(40)

	// Quick action buttons
	buttonsWidget := widgets.NewQWidget(nil, 0)
	buttonsLayout := widgets.NewQHBoxLayout()
	buttonsWidget.SetLayout(buttonsLayout)

	buttonsLayout.AddStretch(1)

	// New lesson button
	newLessonBtn := widgets.NewQPushButton2("Create New Lesson", nil)
	newLessonBtn.SetFixedSize2(200, 50)
	newLessonBtn.ConnectClicked(func(checked bool) {
		mod.showNewLessonDialog()
	})
	buttonsLayout.AddWidget(newLessonBtn, 0, 0)

	buttonsLayout.AddSpacing(20)

	// Open lesson button
	openLessonBtn := widgets.NewQPushButton2("Open Lesson", nil)
	openLessonBtn.SetFixedSize2(200, 50)
	openLessonBtn.ConnectClicked(func(checked bool) {
		mod.showOpenDialog()
	})
	buttonsLayout.AddWidget(openLessonBtn, 0, 0)

	buttonsLayout.AddStretch(1)

	layout.AddWidget(buttonsWidget, 0, 0)

	// Status info
	statusLabel := widgets.NewQLabel2("Module system initialized successfully", nil, 0)
	statusLabel.SetAlignment(qtcore.Qt__AlignHCenter)
	statusLabel.SetStyleSheet("color: green; font-style: italic;")
	layout.AddWidget(statusLabel, 0, 0)

	layout.AddStretch(2)

	return widget
}

// Dialog helper methods
func (mod *GuiModule) showNewLessonDialog() {
	// TODO: Get lesson dialogs module and show new lesson dialog
	mod.statusBar.ShowMessage("New lesson dialog requested", 3000)
}

func (mod *GuiModule) showOpenDialog() {
	// TODO: Get file dialog module and show open dialog
	mod.statusBar.ShowMessage("Open dialog requested", 3000)
}

func (mod *GuiModule) showPropertiesDialog() {
	// TODO: Show lesson properties dialog
	mod.statusBar.ShowMessage("Properties dialog requested", 3000)
}

func (mod *GuiModule) showSettingsDialog() {
	// TODO: Get settings dialog module and show it
	mod.statusBar.ShowMessage("Settings dialog requested", 3000)
}

func (mod *GuiModule) showAboutDialog() {
	// TODO: Get about dialog module and show it
	mod.statusBar.ShowMessage("About dialog requested", 3000)
}

// InitGuiModule creates and returns a new GuiModule instance
func InitGuiModule() core.Module {
	return NewGuiModule()
}
