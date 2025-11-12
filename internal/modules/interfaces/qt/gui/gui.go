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
	"log"

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
		log.Printf("[ERROR] GuiModule.Enable() failed - qtApp module not found")
		return fmt.Errorf("qtApp module not found")
	}

	// Access the QApplication through interface
	if qtMod, ok := qtAppModule.(interface{ GetApplication() *widgets.QApplication }); ok {
		mod.app = qtMod.GetApplication()
		log.Printf("[SUCCESS] GuiModule got QApplication from qtApp module")
	} else {
		log.Printf("[ERROR] GuiModule.Enable() failed - qtApp module does not provide GetApplication method")
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

	log.Printf("[SUCCESS] GuiModule enabled - Qt main window created and shown")
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
		log.Printf("[SUCCESS] GuiModule.ShowMainWindow() - showing main window")
		mod.mainWindow.Show()
		mod.mainWindow.Raise()
		mod.mainWindow.ActivateWindow()
	} else {
		log.Printf("[ERROR] GuiModule.ShowMainWindow() - main window is nil")
	}
}

// GetMainWindow returns the main window widget
func (mod *GuiModule) GetMainWindow() *widgets.QMainWindow {
	return mod.mainWindow
}

// RunEventLoop starts the Qt event loop (blocking call)
func (mod *GuiModule) RunEventLoop() int {
	if mod.app != nil {
		log.Printf("[SUCCESS] GuiModule.RunEventLoop() - starting Qt event loop")
		exitCode := mod.app.Exec()
		log.Printf("[SUCCESS] GuiModule.RunEventLoop() - Qt event loop finished with code %d", exitCode)
		return exitCode
	}
	log.Printf("[ERROR] GuiModule.RunEventLoop() - QApplication is nil")
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
		log.Printf("[EVENT] New Lesson menu action triggered")
		mod.showNewLessonDialog()
	})

	openAction := fileMenu.AddAction("&Open...")
	openAction.SetShortcut(gui.NewQKeySequence2("Ctrl+O", gui.QKeySequence__NativeText))
	openAction.ConnectTriggered(func(checked bool) {
		log.Printf("[EVENT] Open Lesson menu action triggered")
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
		log.Printf("[EVENT] Exit menu action triggered")
		mod.mainWindow.Close()
	})

	// Edit menu
	editMenu := mod.menuBar.AddMenu2("&Edit")

	propertiesAction := editMenu.AddAction("&Properties...")
	propertiesAction.ConnectTriggered(func(checked bool) {
		log.Printf("[EVENT] Properties menu action triggered")
		mod.showPropertiesDialog()
	})

	// Tools menu
	toolsMenu := mod.menuBar.AddMenu2("&Tools")

	settingsAction := toolsMenu.AddAction("&Settings...")
	settingsAction.ConnectTriggered(func(checked bool) {
		log.Printf("[EVENT] Settings menu action triggered")
		mod.showSettingsDialog()
	})

	// Help menu
	helpMenu := mod.menuBar.AddMenu2("&Help")

	aboutAction := helpMenu.AddAction("&About...")
	aboutAction.ConnectTriggered(func(checked bool) {
		log.Printf("[EVENT] About menu action triggered")
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
		log.Printf("[EVENT] Create New Lesson button clicked")
		mod.showNewLessonDialog()
	})
	buttonsLayout.AddWidget(newLessonBtn, 0, 0)

	buttonsLayout.AddSpacing(20)

	// Open lesson button
	openLessonBtn := widgets.NewQPushButton2("Open Lesson", nil)
	openLessonBtn.SetFixedSize2(200, 50)
	openLessonBtn.ConnectClicked(func(checked bool) {
		log.Printf("[EVENT] Open Lesson button clicked")
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
	log.Printf("[ACTION] GuiModule.showNewLessonDialog() - attempting to show lesson dialog")

	// Try to find lesson dialog module
	lessonDialogModules := mod.manager.GetModulesByType("lessonDialogs")
	if len(lessonDialogModules) > 0 {
		log.Printf("[SUCCESS] Found %d lessonDialogs modules, using first one", len(lessonDialogModules))

		// Try to call ShowNewLessonDialog method on the module
		if lessonMod, ok := lessonDialogModules[0].(interface{ ShowNewLessonDialog() map[string]interface{} }); ok {
			log.Printf("[SUCCESS] Calling ShowNewLessonDialog() on lessonDialogs module")
			lessonData := lessonMod.ShowNewLessonDialog()
			if lessonData != nil {
				log.Printf("[SUCCESS] New lesson dialog returned data: %v", lessonData)
				mod.statusBar.ShowMessage("New lesson created successfully", 3000)
				// TODO: Create actual lesson from returned data
			} else {
				log.Printf("[INFO] New lesson dialog was cancelled")
				mod.statusBar.ShowMessage("New lesson creation cancelled", 3000)
			}
		} else {
			log.Printf("[ERROR] lessonDialogs module does not implement ShowNewLessonDialog() method")
			mod.statusBar.ShowMessage("Error: Lesson dialog not available", 3000)
		}
	} else {
		log.Printf("[ERROR] No lessonDialogs modules found")
		mod.statusBar.ShowMessage("Error: No lesson dialog modules available", 3000)
	}
}

func (mod *GuiModule) showOpenDialog() {
	log.Printf("[ACTION] GuiModule.showOpenDialog() - attempting to show file dialog")

	// Try to find file dialog module
	fileDialogModules := mod.manager.GetModulesByType("fileDialog")
	if len(fileDialogModules) > 0 {
		log.Printf("[SUCCESS] Found %d fileDialog modules, using first one", len(fileDialogModules))

		// Try to call OpenFile method on the module
		if fileMod, ok := fileDialogModules[0].(interface {
			OpenFile(parent interface{}, title string, filter string) string
		}); ok {
			log.Printf("[SUCCESS] Calling OpenFile() on fileDialog module")
			fileName := fileMod.OpenFile(nil, "Open Lesson File", "Lesson Files (*.ot *.xml);;All Files (*.*)")
			if fileName != "" {
				log.Printf("[SUCCESS] File dialog returned: %s", fileName)
				mod.statusBar.ShowMessage(fmt.Sprintf("Selected file: %s", fileName), 5000)
				// TODO: Load the selected file
			} else {
				log.Printf("[INFO] File dialog was cancelled")
				mod.statusBar.ShowMessage("File selection cancelled", 3000)
			}
		} else {
			log.Printf("[ERROR] fileDialog module does not implement OpenFile() method")
			mod.statusBar.ShowMessage("Error: File dialog not available", 3000)
		}
	} else {
		log.Printf("[ERROR] No fileDialog modules found")
		mod.statusBar.ShowMessage("Error: No file dialog modules available", 3000)
	}
}

func (mod *GuiModule) showPropertiesDialog() {
	log.Printf("[STUB] GuiModule.showPropertiesDialog() - properties dialog not implemented")

	// Try to find properties dialog module
	propertiesDialogModules := mod.manager.GetModulesByType("propertiesDialog")
	if len(propertiesDialogModules) > 0 {
		log.Printf("[DEBUG] Found %d propertiesDialog modules", len(propertiesDialogModules))
	} else {
		log.Printf("[ERROR] No propertiesDialog modules found")
	}

	// TODO: Show lesson properties dialog
	mod.statusBar.ShowMessage("Properties dialog requested", 3000)
}

func (mod *GuiModule) showSettingsDialog() {
	log.Printf("[ACTION] GuiModule.showSettingsDialog() - attempting to show settings dialog")

	// Try to find settings dialog module
	settingsDialogModules := mod.manager.GetModulesByType("settingsDialog")
	if len(settingsDialogModules) > 0 {
		log.Printf("[SUCCESS] Found %d settingsDialog modules, using first one", len(settingsDialogModules))

		// Try to call ShowSettingsDialog method on the module
		if settingsMod, ok := settingsDialogModules[0].(interface{ ShowSettingsDialog() bool }); ok {
			log.Printf("[SUCCESS] Calling ShowSettingsDialog() on settingsDialog module")
			applied := settingsMod.ShowSettingsDialog()
			if applied {
				log.Printf("[SUCCESS] Settings dialog applied changes")
				mod.statusBar.ShowMessage("Settings updated successfully", 3000)
			} else {
				log.Printf("[INFO] Settings dialog was cancelled or no changes made")
				mod.statusBar.ShowMessage("Settings dialog cancelled", 3000)
			}
		} else {
			log.Printf("[ERROR] settingsDialog module does not implement ShowSettingsDialog() method")
			mod.statusBar.ShowMessage("Error: Settings dialog not available", 3000)
		}
	} else {
		log.Printf("[ERROR] No settingsDialog modules found")
		mod.statusBar.ShowMessage("Error: No settings dialog modules available", 3000)
	}
}

func (mod *GuiModule) showAboutDialog() {
	log.Printf("[ACTION] GuiModule.showAboutDialog() - attempting to show about dialog")

	// Try to find about dialog module
	aboutDialogModules := mod.manager.GetModulesByType("aboutDialog")
	if len(aboutDialogModules) > 0 {
		log.Printf("[SUCCESS] Found %d aboutDialog modules, using first one", len(aboutDialogModules))

		// Try to call ShowAboutDialog method on the module
		if aboutMod, ok := aboutDialogModules[0].(interface{ ShowAboutDialog() }); ok {
			log.Printf("[SUCCESS] Calling ShowAboutDialog() on aboutDialog module")
			aboutMod.ShowAboutDialog()
			log.Printf("[SUCCESS] About dialog was shown")
			mod.statusBar.ShowMessage("About dialog displayed", 2000)
		} else {
			log.Printf("[ERROR] aboutDialog module does not implement ShowAboutDialog() method")
			mod.statusBar.ShowMessage("Error: About dialog not available", 3000)
		}
	} else {
		log.Printf("[ERROR] No aboutDialog modules found")
		mod.statusBar.ShowMessage("Error: No about dialog modules available", 3000)
	}
}

// InitGuiModule creates and returns a new GuiModule instance
func InitGuiModule() core.Module {
	return NewGuiModule()
}
