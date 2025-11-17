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
	"path/filepath"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/logging"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessons/media"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessons/topo"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessons/words"
	"github.com/mappu/miqt/qt"
)

// GuiModule is a Go port of the Python GuiModule class
type GuiModule struct {
	*core.BaseModule
	manager        *core.Manager
	mainWindow     *qt.QMainWindow
	app            *qt.QApplication
	menuBar        *qt.QMenuBar
	statusBar      *qt.QStatusBar
	tabWidget      *qt.QTabWidget
	lastLoadedFile string
	lastLoadTime   int64
	logger         *logging.Logger
	addingTab      bool
	showingDialog  bool
}

// NewGuiModule creates a new GuiModule instance
func NewGuiModule() *GuiModule {
	base := core.NewBaseModule("ui", "gui-module")
	base.SetRequires("qtApp")

	return &GuiModule{
		BaseModule: base,
		logger:     logging.GetModuleLogger("GUI"),
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
	if qtMod, ok := qtAppModule.(interface{ GetApplication() *qt.QApplication }); ok {
		mod.app = qtMod.GetApplication()
		mod.logger.Success("Got QApplication from qtApp module")
	} else {
		log.Printf("[ERROR] GuiModule.Enable() failed - qtApp module does not provide GetApplication method")
		return fmt.Errorf("qtApp module does not provide GetApplication method")
	}

	// Create main window
	mod.mainWindow = qt.NewQMainWindow(nil)
	mod.mainWindow.SetWindowTitle("OpenTeacher 4.0")
	mod.mainWindow.Resize(1000, 700)
	mod.mainWindow.SetMinimumSize2(800, 600)

	// Create menu bar
	mod.createMenuBar()

	// Create status bar
	mod.statusBar = mod.mainWindow.StatusBar()
	mod.statusBar.ShowMessage("Ready")

	// Create central widget with basic layout
	centralWidget := qt.NewQWidget(nil)
	mod.mainWindow.SetCentralWidget(centralWidget)

	// Create main layout
	mainLayout := qt.NewQVBoxLayout(centralWidget)

	// Add welcome area
	welcomeWidget := mod.createWelcomeWidget()
	mainLayout.AddWidget(welcomeWidget)

	// Show the window
	mod.mainWindow.Show()

	mod.logger.Success("Qt main window created and shown")
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

	// Clean up tab widget
	mod.tabWidget = nil

	// Don't quit the app - that's managed by qtApp module
	mod.app = nil

	fmt.Println("GuiModule disabled")
	return nil
}

// SetManager sets the module manager
// Public methods for command-line interface

// ShowNewLessonDialog is a public wrapper for showNewLessonDialog
func (mod *GuiModule) ShowNewLessonDialog() {
	mod.showNewLessonDialog()
}

// ShowPropertiesDialog is a public wrapper for showPropertiesDialog
func (mod *GuiModule) ShowPropertiesDialog() {
	mod.showPropertiesDialog()
}

// ShowSettingsDialog is a public wrapper for showSettingsDialog
func (mod *GuiModule) ShowSettingsDialog() {
	mod.showSettingsDialog()
}

// ShowAboutDialog is a public wrapper for showAboutDialog
func (mod *GuiModule) ShowAboutDialog() {
	mod.showAboutDialog()
}

// ShowOpenDialog is a public wrapper for showOpenDialog
func (mod *GuiModule) ShowOpenDialog() {
	mod.showOpenDialog()
}

// LoadSelectedFile is a public wrapper for loadSelectedFile
func (mod *GuiModule) LoadSelectedFile(fileName string) error {
	mod.loadSelectedFile(fileName)
	return nil
}

// Exit closes the application
func (mod *GuiModule) Exit() {
	if mod.mainWindow != nil {
		mod.mainWindow.Close()
	}
}

func (mod *GuiModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// ShowMainWindow shows the main application window
func (mod *GuiModule) ShowMainWindow() {
	if mod.mainWindow != nil {
		mod.logger.Success("ShowMainWindow() - main window displayed")
		mod.mainWindow.Show()
		mod.mainWindow.Raise()
		mod.mainWindow.ActivateWindow()
	} else {
		log.Printf("[ERROR] GuiModule.ShowMainWindow() - main window is nil")
	}
}

// GetMainWindow returns the main window widget
func (mod *GuiModule) GetMainWindow() *qt.QMainWindow {
	return mod.mainWindow
}

// RunEventLoop starts the Qt event loop (blocking call)
func (mod *GuiModule) RunEventLoop() int {
	if mod.app != nil {
		mod.logger.Success("RunEventLoop() - Qt event loop started")
		exitCode := qt.QApplication_Exec()
		mod.logger.Success("RunEventLoop() - Qt event loop finished with code %d", exitCode)
		return exitCode
	}
	log.Printf("[ERROR] GuiModule.RunEventLoop() - QApplication is nil")
	return 0
}

// createMenuBar creates the main menu bar
func (mod *GuiModule) createMenuBar() {
	mod.menuBar = mod.mainWindow.MenuBar()

	// File menu
	fileMenu := qt.NewQMenu2()
	fileMenu.SetTitle("&File")
	mod.menuBar.AddMenu(fileMenu)

	newAction := fileMenu.AddAction("&New Lesson...")
	newAction.SetShortcut(qt.NewQKeySequence2("Ctrl+N"))
	newAction.OnTriggered(func() {
		mod.logger.Event("New Lesson menu action triggered")
		mod.showNewLessonDialog()
	})

	openAction := fileMenu.AddAction("&Open...")
	openAction.SetShortcut(qt.NewQKeySequence2("Ctrl+O"))
	openAction.OnTriggered(func() {
		mod.logger.Event("Open Lesson menu action triggered")
		mod.showOpenDialogFrom("MENU")
	})

	fileMenu.AddSeparator()

	saveAction := fileMenu.AddAction("&Save")
	saveAction.SetShortcut(qt.NewQKeySequence2("Ctrl+S"))
	saveAction.SetEnabled(false) // Enable when lesson is loaded
	saveAction.OnTriggered(func() {
		mod.logger.Event("Save menu action triggered")
	})

	saveAsAction := fileMenu.AddAction("Save &As...")
	saveAsAction.SetShortcut(qt.NewQKeySequence2("Ctrl+Shift+S"))
	saveAsAction.SetEnabled(false) // Enable when lesson is loaded
	saveAsAction.OnTriggered(func() {
		mod.logger.Event("Save As menu action triggered")
	})

	fileMenu.AddSeparator()

	exitAction := fileMenu.AddAction("E&xit")
	exitAction.SetShortcut(qt.NewQKeySequence2("Ctrl+Q"))
	exitAction.OnTriggered(func() {
		mod.logger.Event("Exit menu action triggered")
		mod.mainWindow.Close()
	})

	// Edit menu
	editMenu := qt.NewQMenu2()
	editMenu.SetTitle("&Edit")
	mod.menuBar.AddMenu(editMenu)

	propertiesAction := editMenu.AddAction("&Properties...")
	propertiesAction.OnTriggered(func() {
		mod.logger.Event("Properties menu action triggered")
		mod.showPropertiesDialog()
	})

	// Tools menu
	toolsMenu := qt.NewQMenu2()
	toolsMenu.SetTitle("&Tools")
	mod.menuBar.AddMenu(toolsMenu)

	settingsAction := toolsMenu.AddAction("&Settings...")
	settingsAction.OnTriggered(func() {
		mod.logger.Event("Settings menu action triggered")
		mod.showSettingsDialog()
	})

	toolsMenu.AddSeparator()

	importAction := toolsMenu.AddAction("&Import...")
	importAction.OnTriggered(func() {
		mod.logger.Event("Import menu action triggered")
		mod.logger.Warning("Import functionality not yet implemented")
	})

	// Help menu
	helpMenu := qt.NewQMenu2()
	helpMenu.SetTitle("&Help")
	mod.menuBar.AddMenu(helpMenu)

	aboutAction := helpMenu.AddAction("&About...")
	aboutAction.OnTriggered(func() {
		mod.logger.Event("About menu action triggered")
		mod.showAboutDialog()
	})
}

// createWelcomeWidget creates the welcome screen widget
func (mod *GuiModule) createWelcomeWidget() *qt.QWidget {
	widget := qt.NewQWidget(nil)
	layout := qt.NewQVBoxLayout(widget)

	// Add some spacing
	layout.AddStretch()

	// Main title
	titleLabel := qt.NewQLabel(nil)
	titleLabel.SetText("Welcome to OpenTeacher 4.0")
	titleFont := titleLabel.Font()
	titleFont.SetPointSize(24)
	titleFont.SetBold(true)
	titleLabel.SetFont(titleFont)
	titleLabel.SetAlignment(qt.AlignHCenter)
	layout.AddWidget(titleLabel.QWidget)

	// Subtitle
	subtitleLabel := qt.NewQLabel(nil)
	subtitleLabel.SetText("Learn whatever you want to learn")
	subtitleFont := subtitleLabel.Font()
	subtitleFont.SetPointSize(14)
	subtitleLabel.SetFont(subtitleFont)
	subtitleLabel.SetAlignment(qt.AlignHCenter)
	layout.AddWidget(subtitleLabel.QWidget)

	// Add some spacing
	layout.AddSpacing(20)

	// Quick action buttons
	buttonsWidget := qt.NewQWidget(nil)
	buttonsLayout := qt.NewQHBoxLayout(buttonsWidget)

	buttonsLayout.AddStretch()

	// New lesson button
	newLessonBtn := qt.NewQPushButton(nil)
	newLessonBtn.SetText("Create New Lesson")
	newLessonBtn.SetFixedSize2(200, 50)
	newLessonBtn.OnClicked(func() {
		mod.logger.Event("Create New Lesson button clicked")
		mod.showNewLessonDialog()
	})
	buttonsLayout.AddWidget(newLessonBtn.QWidget)

	buttonsLayout.AddSpacing(20)

	// Open lesson button
	openLessonBtn := qt.NewQPushButton(nil)
	openLessonBtn.SetText("Open Lesson")
	openLessonBtn.SetFixedSize2(200, 50)
	openLessonBtn.OnClicked(func() {
		mod.logger.Event("Open Lesson button clicked")
		mod.showOpenDialogFrom("BUTTON")
	})
	buttonsLayout.AddWidget(openLessonBtn.QWidget)

	buttonsLayout.AddStretch()

	layout.AddWidget(buttonsWidget)

	// Status info
	statusLabel := qt.NewQLabel(nil)
	statusLabel.SetText("Module system initialized successfully")
	statusLabel.SetAlignment(qt.AlignHCenter)
	statusLabel.SetStyleSheet("color: #888; font-size: 12px; margin: 20px;")
	layout.AddWidget(statusLabel.QWidget)

	layout.AddStretch()

	return widget
}

// Dialog helper methods
func (mod *GuiModule) showNewLessonDialog() {
	mod.logger.Action("showNewLessonDialog() - attempting to show lesson dialog")

	// Try to find lesson dialog module
	lessonDialogModules := mod.manager.GetModulesByType("lessonDialogs")
	if len(lessonDialogModules) > 0 {
		mod.logger.Success("Found %d lessonDialogs modules, using first one", len(lessonDialogModules))

		// Try to call ShowNewLessonDialog method on the module
		if lessonMod, ok := lessonDialogModules[0].(interface{ ShowNewLessonDialog() map[string]interface{} }); ok {
			mod.logger.Success("Calling ShowNewLessonDialog() on lessonDialogs module")
			lessonData := lessonMod.ShowNewLessonDialog()
			if lessonData != nil {
				log.Printf("[SUCCESS] New lesson dialog returned data: %v", lessonData)

				// Create actual lesson from returned data
				newLesson, err := mod.CreateLessonFromDialogData(lessonData)
				if err != nil {
					mod.logger.Error("Failed to create lesson from dialog data: %v", err)
					mod.statusBar.ShowMessage("Error creating lesson: " + err.Error())
					return
				}

				// Display lesson in new tab
				mod.displayLessonInTab(newLesson)
				mod.statusBar.ShowMessage("New lesson created successfully")
			} else {
				log.Printf("[INFO] New lesson dialog was cancelled")
				mod.statusBar.ShowMessage("New lesson dialog created")
			}
		} else {
			mod.logger.DeadEnd("lessonDialogs module", "does not implement ShowNewLessonDialog() method", "legacy/modules/org/openteacher/interfaces/qt/lessonDialogs/")
			mod.statusBar.ShowMessage("Error: Lesson dialog not available")
		}
	} else {
		mod.logger.DeadEnd("lessonDialogs system", "No lessonDialogs modules found", "legacy/modules/org/openteacher/interfaces/qt/lessonDialogs/")
		mod.statusBar.ShowMessage("Error: No lesson dialog modules available")
	}
}

func (mod *GuiModule) showOpenDialog() {
	mod.showOpenDialogFrom("UNKNOWN")
}

func (mod *GuiModule) showOpenDialogFrom(source string) {
	mod.logger.Action("showOpenDialogFrom(%s) - attempting to show file dialog", source)

	// Prevent double dialog calls (Qt signal issue)
	if mod.showingDialog {
		mod.logger.Warning("PREVENTED: showOpenDialog called while already showing dialog")
		return
	}
	mod.showingDialog = true
	defer func() {
		mod.showingDialog = false
	}()

	// Try to find file dialog module
	fileDialogModules := mod.manager.GetModulesByType("fileDialog")
	if len(fileDialogModules) > 0 {
		mod.logger.Success("Found %d fileDialog modules, using first one", len(fileDialogModules))

		// Try to call OpenFile method on the module
		if fileMod, ok := fileDialogModules[0].(interface {
			OpenFile(parent interface{}, title string, filter string) string
		}); ok {
			mod.logger.Success("Calling OpenFile() on fileDialog module")
			mod.logger.Debug("TRACKING: About to call OpenFile() - call stack marker A")
			fileName := fileMod.OpenFile(nil, "Open Lesson File", "")
			mod.logger.Debug("TRACKING: OpenFile() returned - call stack marker B")
			if fileName != "" {
				mod.logger.Success("File dialog returned: %s", fileName)
				mod.logger.Debug("TRACKING: About to call loadSelectedFile() - call stack marker C")
				mod.statusBar.ShowMessage(fmt.Sprintf("Selected file: %s", fileName))
				mod.loadSelectedFile(fileName)
				mod.logger.Debug("TRACKING: loadSelectedFile() completed - call stack marker D")
			} else {
				mod.logger.Info("File dialog was cancelled")
				mod.statusBar.ShowMessage("Open operation cancelled")
			}
		} else {
			mod.logger.DeadEnd("fileDialog module", "does not implement OpenFile() method", "legacy/modules/org/openteacher/interfaces/qt/dialogs/")
			mod.statusBar.ShowMessage("Error: File dialog not available")
		}
	} else {
		mod.logger.DeadEnd("fileDialog system", "No fileDialog modules found", "legacy/modules/org/openteacher/interfaces/qt/dialogs/")
		mod.statusBar.ShowMessage("Error: No file dialog modules available")
	}
}

// loadSelectedFile loads the file selected by the user
func (mod *GuiModule) loadSelectedFile(fileName string) {
	mod.logger.Action("loadSelectedFile() - loading file: %s", fileName)

	// Prevent duplicate loading of the same file within 2 seconds
	currentTime := qt.QDateTime_CurrentMSecsSinceEpoch()
	if mod.lastLoadedFile == fileName && (currentTime-mod.lastLoadTime) < 2000 {
		mod.logger.Warning("Ignoring duplicate load request for: %s (double-click protection)", fileName)
		return
	}
	mod.lastLoadedFile = fileName
	mod.lastLoadTime = currentTime

	// Create file loader
	fileLoader := lesson.NewFileLoader()

	// Load the lesson data
	lessonData, err := fileLoader.LoadFile(fileName)
	if err != nil {
		mod.logger.Error("Failed to load file '%s': %v", fileName, err)
		mod.statusBar.ShowMessage(fmt.Sprintf("Error loading file: %v", err))
		return
	}

	// Get file type
	fileType := fileLoader.GetFileType(fileName)
	mod.logger.Success("Loaded lesson file - Type: %s, Items: %d", fileType, len(lessonData.List.Items))

	// Create lesson instance
	newLesson := lesson.NewLesson(fileType)
	newLesson.Data = *lessonData
	newLesson.Path = fileName

	// Display lesson summary in status bar
	wordCount := newLesson.Data.List.GetWordCount()
	testCount := newLesson.Data.List.GetTestCount()
	title := newLesson.Data.List.Title
	if title == "" {
		title = filepath.Base(fileName)
	}

	statusMsg := fmt.Sprintf("Loaded '%s': %d words", title, wordCount)
	if testCount > 0 {
		statusMsg += fmt.Sprintf(", %d tests", testCount)
	}
	mod.statusBar.ShowMessage(statusMsg)

	// Log the lesson details
	mod.logger.Success("Lesson loaded successfully:")
	mod.logger.Info("  - Title: %s", title)
	mod.logger.Info("  - Question Language: %s", newLesson.Data.List.QuestionLanguage)
	mod.logger.Info("  - Answer Language: %s", newLesson.Data.List.AnswerLanguage)
	mod.logger.Info("  - Word pairs: %d", wordCount)
	mod.logger.Info("  - Test results: %d", testCount)

	// Sample the first few words for verification
	if len(newLesson.Data.List.Items) > 0 {
		mod.logger.Debug("Sample word pairs:")
		maxSamples := 3
		if len(newLesson.Data.List.Items) < maxSamples {
			maxSamples = len(newLesson.Data.List.Items)
		}
		for i := 0; i < maxSamples; i++ {
			item := newLesson.Data.List.Items[i]
			mod.logger.Debug("  - %v → %v", item.Questions, item.Answers)
		}
		if len(newLesson.Data.List.Items) > maxSamples {
			mod.logger.Debug("  - ... and %d more", len(newLesson.Data.List.Items)-maxSamples)
		}
	}

	// Create lesson tab and display in main window
	mod.displayLessonInTab(newLesson)
}

// CreateLessonFromDialogData creates a new lesson from dialog data
func (mod *GuiModule) CreateLessonFromDialogData(data map[string]interface{}) (*lesson.Lesson, error) {
	mod.logger.Action("CreateLessonFromDialogData() - creating lesson from dialog data")

	// Extract lesson information
	name, _ := data["name"].(string)
	description, _ := data["description"].(string)
	lessonType, _ := data["type"].(string)
	questionLang, _ := data["questionLanguage"].(string)
	answerLang, _ := data["answerLanguage"].(string)

	// Set defaults if missing
	if name == "" {
		name = "New Lesson"
	}
	if lessonType == "" {
		lessonType = "words"
	}
	if questionLang == "" {
		questionLang = "English"
	}
	if answerLang == "" {
		answerLang = "English"
	}

	// Create new lesson
	newLesson := lesson.NewLesson(lessonType)
	newLesson.Data.List.Title = name
	newLesson.Data.List.QuestionLanguage = questionLang
	newLesson.Data.List.AnswerLanguage = answerLang

	// Add description as metadata if provided
	if description != "" {
		if newLesson.Data.Resources == nil {
			newLesson.Data.Resources = make(map[string]interface{})
		}
		newLesson.Data.Resources["description"] = description
	}

	// Set path for new lesson (unsaved initially)
	newLesson.Path = fmt.Sprintf("*%s", name) // * indicates unsaved
	newLesson.Data.Changed = true

	mod.logger.Success("Created new lesson: %s (%s -> %s)", name, questionLang, answerLang)
	return newLesson, nil
}

// displayLessonInTab creates a new tab for the lesson
func (mod *GuiModule) displayLessonInTab(lesson *lesson.Lesson) {
	mod.logger.Action("displayLessonInTab() - creating lesson tab for: %s", lesson.Path)

	// Prevent double tab creation similar to Python _addingTab flag
	if mod.addingTab {
		mod.logger.Warning("PREVENTED: displayLessonInTab called while already adding tab for: %s", lesson.Path)
		return
	}
	mod.addingTab = true
	defer func() {
		mod.addingTab = false
	}()

	// Create tab widget if it doesn't exist
	if mod.tabWidget == nil {
		mod.tabWidget = qt.NewQTabWidget(nil)
		mod.mainWindow.SetCentralWidget(mod.tabWidget.QWidget)
		mod.logger.Success("Created central tab widget")
	} else {
		// Tab widget already exists, just update the central widget if needed
		if mod.mainWindow.CentralWidget() != mod.tabWidget.QWidget {
			mod.mainWindow.SetCentralWidget(mod.tabWidget.QWidget)
			mod.logger.Success("Replaced central widget with tab widget")
		}
	}

	// Create lesson content widget
	lessonWidget := mod.createLessonWidget(lesson)

	// Create tab title
	title := lesson.Data.List.Title
	if title == "" {
		title = filepath.Base(lesson.Path)
	}

	// Add the tab
	tabIndex := mod.tabWidget.AddTab(lessonWidget, title)
	mod.tabWidget.SetCurrentIndex(tabIndex)

	// Update status bar
	statusMsg := fmt.Sprintf("Opened '%s' - %d words", title, lesson.Data.List.GetWordCount())
	mod.statusBar.ShowMessage(statusMsg)

	mod.logger.Success("Lesson tab created: %s (%d words)", title, lesson.Data.List.GetWordCount())
}

// createLessonWidget creates a widget to display lesson content
func (mod *GuiModule) createLessonWidget(lesson *lesson.Lesson) *qt.QWidget {
	// Determine lesson type and create appropriate widget
	var lessonWidget *qt.QWidget

	switch lesson.DataType {
	case "topo":
		mod.logger.Info("Creating topography lesson widget for: %s", lesson.Path)
		topoWidget := topo.NewTopoLessonWidget(lesson, mod.mainWindow.QWidget)
		lessonWidget = topoWidget.QWidget

		// Validate layout after creation (will check for overlaps in strict mode)
		topoWidget.ValidateLayoutAfterShow()
	case "media":
		mod.logger.Info("Creating media lesson widget for: %s", lesson.Path)
		mediaWidget := media.NewMediaLessonWidget(lesson, mod.mainWindow.QWidget)
		lessonWidget = mediaWidget.QWidget
	case "words":
		fallthrough
	default:
		// Default to words widget for unknown types or actual words lessons
		mod.logger.Info("Creating words lesson widget for: %s (type: %s)", lesson.Path, lesson.DataType)
		wordsWidget := words.NewWordsLessonWidget(lesson, mod.mainWindow.QWidget)
		lessonWidget = wordsWidget.QWidget
	}

	// TODO: Connect lesson change signal to update window title and status
	// This will be implemented when proper Qt signal system is in place
	mod.logger.LegacyReminder("Qt signal connection for lesson changes", "legacy/modules/org/openteacher/interfaces/qt/gui/gui.py", "proper signal handling needed")
	mod.logger.Info("Created lesson widget for: %s", lesson.Path)

	mod.logger.Success("Created lesson widget with Enter/Teach/Results tabs")
	return lessonWidget
}

func (mod *GuiModule) showPropertiesDialog() {
	mod.logger.Action("showPropertiesDialog() - attempting to show lesson properties dialog")

	// Get the current lesson data
	currentLessonData := mod.getCurrentLessonData()
	if currentLessonData == nil {
		mod.logger.Error("No current lesson to show properties for")
		mod.statusBar.ShowMessage("No lesson open to show properties")
		return
	}

	// Try to find lesson dialog module which handles properties
	lessonDialogModules := mod.manager.GetModulesByType("lessonDialogs")
	if len(lessonDialogModules) == 0 {
		mod.logger.DeadEnd("lessonDialogs system", "No lessonDialogs modules found", "internal/modules/interfaces/qt/lessonDialogs/")
		mod.statusBar.ShowMessage("Properties dialog cancelled")
		return
	}

	mod.logger.Success("Found %d lessonDialogs modules, using first one", len(lessonDialogModules))

	// Cast to the interface we need
	if dialogMod, ok := lessonDialogModules[0].(interface {
		ShowPropertiesDialog(*qt.QWidget, map[string]interface{}) map[string]interface{}
	}); ok {
		mod.logger.Success("Calling ShowPropertiesDialog() on lessonDialogs module")

		// Show the properties dialog and get the result
		updatedData := dialogMod.ShowPropertiesDialog(mod.mainWindow.QWidget, currentLessonData)

		if updatedData != nil {
			mod.logger.Success("Properties dialog returned updated data")
			mod.updateCurrentLessonData(updatedData)
			mod.statusBar.ShowMessage("Lesson properties updated successfully")
		} else {
			mod.logger.Info("Properties dialog was cancelled or no changes made")
		}
	} else {
		mod.logger.Error("lessonDialogs module doesn't have ShowPropertiesDialog method")
		mod.statusBar.ShowMessage("Error: Properties dialog not available")
	}
}

func (mod *GuiModule) showSettingsDialog() {
	mod.logger.Action("showSettingsDialog() - attempting to show settings dialog")

	// Try to find settings dialog module
	settingsDialogModules := mod.manager.GetModulesByType("settingsDialog")
	if len(settingsDialogModules) > 0 {
		mod.logger.Success("Found %d settingsDialog modules, using first one", len(settingsDialogModules))

		// Try to call ShowSettingsDialog method on the module
		if settingsMod, ok := settingsDialogModules[0].(interface{ ShowSettingsDialog() bool }); ok {
			mod.logger.Success("Calling ShowSettingsDialog() on settingsDialog module")
			applied := settingsMod.ShowSettingsDialog()
			if applied {
				mod.logger.Success("Settings dialog applied changes")
				mod.statusBar.ShowMessage("File opened successfully")
			} else {
				mod.logger.Info("Settings dialog was cancelled or no changes made")
				mod.statusBar.ShowMessage("Settings dialog created")
			}
		} else {
			mod.logger.DeadEnd("settingsDialog module", "does not implement ShowSettingsDialog() method", "legacy/modules/org/openteacher/interfaces/qt/dialogs/settings/")
			mod.statusBar.ShowMessage("Error: Settings dialog not available")
		}
	} else {
		mod.logger.DeadEnd("settingsDialog system", "No settingsDialog modules found", "legacy/modules/org/openteacher/interfaces/qt/dialogs/settings/")
		mod.statusBar.ShowMessage("Error: No settings dialog modules available")
	}
}

func (mod *GuiModule) showAboutDialog() {
	mod.logger.Action("showAboutDialog() - attempting to show about dialog")

	// Try to find about dialog module
	aboutDialogModules := mod.manager.GetModulesByType("aboutDialog")
	if len(aboutDialogModules) > 0 {
		mod.logger.Success("Found %d aboutDialog modules, using first one", len(aboutDialogModules))

		// Try to call ShowAboutDialog method on the module
		if aboutMod, ok := aboutDialogModules[0].(interface{ ShowAboutDialog() }); ok {
			mod.logger.Success("Calling ShowAboutDialog() on aboutDialog module")
			aboutMod.ShowAboutDialog()
			mod.logger.Success("About dialog was shown")
			mod.statusBar.ShowMessage("Import dialog created")
		} else {
			mod.logger.DeadEnd("aboutDialog module", "does not implement ShowAboutDialog() method", "legacy/modules/org/openteacher/interfaces/qt/dialogs/about/")
			mod.statusBar.ShowMessage("About dialog not available")
		}
	} else {
		mod.logger.DeadEnd("aboutDialog system", "No aboutDialog modules found", "legacy/modules/org/openteacher/interfaces/qt/dialogs/about/")
		mod.statusBar.ShowMessage("Error: No about dialog modules available")
	}
}

// getCurrentLessonData gets the current lesson data for the properties dialog
func (mod *GuiModule) getCurrentLessonData() map[string]interface{} {
	if mod.tabWidget == nil {
		return nil
	}

	currentIndex := mod.tabWidget.CurrentIndex()
	if currentIndex < 0 {
		return nil
	}

	tabText := mod.tabWidget.TabText(currentIndex)

	// Extract lesson information from the current tab
	// This is a simplified implementation - in a full version you'd get the actual lesson data
	lessonData := make(map[string]interface{})
	lessonData["name"] = tabText
	lessonData["description"] = ""
	lessonData["author"] = ""
	lessonData["version"] = "1.0"
	lessonData["itemCount"] = 0 // TODO: Get actual count from lesson

	mod.logger.Debug("Retrieved current lesson data for: %s", tabText)

	return lessonData
}

// updateCurrentLessonData updates the current lesson with new property data
func (mod *GuiModule) updateCurrentLessonData(data map[string]interface{}) {
	if mod.tabWidget == nil || data == nil {
		return
	}

	currentIndex := mod.tabWidget.CurrentIndex()
	if currentIndex < 0 {
		return
	}

	// Update tab title if name changed
	if name, ok := data["name"].(string); ok && name != "" {
		mod.tabWidget.SetTabText(currentIndex, name)
		mod.logger.Info("Updated lesson name to: %s", name)
	}

	// TODO: Update actual lesson data in the lesson widget
	mod.logger.Info("Lesson properties updated successfully")
}

// InitGuiModule creates and returns a new GuiModule instance
func InitGuiModule() core.Module {
	return NewGuiModule()
}
