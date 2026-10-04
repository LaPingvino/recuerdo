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
	"errors"
	"github.com/LaPingvino/recuerdo/internal/i18n"
	datatypeicons "github.com/LaPingvino/recuerdo/internal/modules/data/dataTypeIcons"
	userdocumentation "github.com/LaPingvino/recuerdo/internal/modules/data/userDocumentation"
	plaintextwords "github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/enterers/plainTextWords"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/ocrimport"
	recentlyopened "github.com/LaPingvino/recuerdo/internal/modules/logic/recentlyOpened"
	"github.com/LaPingvino/recuerdo/internal/ocr"
	"github.com/LaPingvino/recuerdo/internal/settingsdefs"
	"unsafe"

	"context"
	"fmt"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/export"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/icon"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/logging"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessons/media"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessons/topo"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessons/words"
	"github.com/mappu/miqt/qt"
	"github.com/mappu/miqt/qt/printsupport"
)

// GuiModule is a Go port of the Python GuiModule class
type GuiModule struct {
	*core.BaseModule
	manager        *core.Manager
	lastWords      *words.WordsLessonWidget // the most recently opened word lesson
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

	saveAction, saveAsAction, printAction *qt.QAction
	// tabLessons is the lesson shown in each lesson tab, by tab widget
	tabLessons map[unsafe.Pointer]*lesson.Lesson
	// tabWords is the word lesson widget of each word lesson tab
	tabWords map[unsafe.Pointer]*words.WordsLessonWidget
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

	// The interface language (setting, or the system's), before any text
	if settings, ok := mod.manager.GetDefaultModule("settings"); ok {
		if st, ok := settings.(settingsdefs.Store); ok {
			i18n.Start(st)
		}
	}

	// Create main window
	mod.mainWindow = qt.NewQMainWindow(nil)
	mod.mainWindow.SetWindowTitle("Recuerdo")
	mod.mainWindow.Resize(1000, 700)
	mod.mainWindow.SetMinimumSize2(800, 600)

	// Create menu bar
	mod.createMenuBar()

	// Create status bar
	mod.statusBar = mod.mainWindow.StatusBar()
	mod.statusBar.ShowMessage(i18n.T("Ready"))

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
		mod.startScreenshots()
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
	fileMenu.SetTitle(i18n.T("&File"))
	mod.menuBar.AddMenu(fileMenu)

	newAction := fileMenu.AddAction(i18n.T("&New Lesson..."))
	newAction.SetShortcut(qt.NewQKeySequence2("Ctrl+N"))
	newAction.OnTriggered(func() {
		mod.logger.Event("New Lesson menu action triggered")
		mod.showNewLessonDialog()
	})

	textAction := fileMenu.AddAction(i18n.T("New from &Text..."))
	textAction.SetToolTip(i18n.T("Type or paste a word list as \"question = answer\" lines"))
	textAction.OnTriggered(mod.newLessonFromText)

	openAction := fileMenu.AddAction(i18n.T("&Open..."))
	openAction.SetShortcut(qt.NewQKeySequence2("Ctrl+O"))
	openAction.OnTriggered(func() {
		mod.logger.Event("Open Lesson menu action triggered")
		mod.showOpenDialogFrom("MENU")
	})

	recentMenu := fileMenu.AddMenuWithTitle(i18n.T("Open &Recent"))
	recentMenu.OnAboutToShow(func() { mod.fillRecentMenu(recentMenu) })

	mergeAction := fileMenu.AddAction(i18n.T("&Merge Lesson..."))
	mergeAction.SetToolTip(i18n.T("Add the words and results of another lesson to this one"))
	mergeAction.OnTriggered(mod.mergeIntoCurrentLesson)

	pictureAction := fileMenu.AddAction(i18n.T("Import from &Picture..."))
	pictureAction.SetToolTip(i18n.T("Read a word list from a scan or photo of a printed list (needs Tesseract)"))
	pictureAction.OnTriggered(mod.importFromPicture)

	fileMenu.AddSeparator()

	saveAction := fileMenu.AddAction(i18n.T("&Save"))
	saveAction.SetShortcut(qt.NewQKeySequence2("Ctrl+S"))
	saveAction.SetEnabled(false) // enabled when a lesson is open
	saveAction.OnTriggered(func() { mod.saveCurrentLesson(false) })
	mod.saveAction = saveAction

	saveAsAction := fileMenu.AddAction(i18n.T("Save &As..."))
	saveAsAction.SetShortcut(qt.NewQKeySequence2("Ctrl+Shift+S"))
	saveAsAction.SetEnabled(false) // enabled when a lesson is open
	saveAsAction.OnTriggered(func() { mod.saveCurrentLesson(true) })
	mod.saveAsAction = saveAsAction

	printAction := fileMenu.AddAction(i18n.T("&Print..."))
	printAction.SetShortcut(qt.NewQKeySequence2("Ctrl+P"))
	printAction.SetEnabled(false) // enabled when a lesson is open
	printAction.OnTriggered(mod.printCurrentLesson)
	mod.printAction = printAction

	fileMenu.AddSeparator()

	exitAction := fileMenu.AddAction(i18n.T("E&xit"))
	exitAction.SetShortcut(qt.NewQKeySequence2("Ctrl+Q"))
	exitAction.OnTriggered(func() {
		mod.logger.Event("Exit menu action triggered")
		mod.mainWindow.Close()
	})

	// Edit menu
	editMenu := qt.NewQMenu2()
	editMenu.SetTitle(i18n.T("&Edit"))
	mod.menuBar.AddMenu(editMenu)

	propertiesAction := editMenu.AddAction(i18n.T("&Properties..."))
	propertiesAction.OnTriggered(func() {
		mod.logger.Event("Properties menu action triggered")
		mod.showPropertiesDialog()
	})

	// Tools menu
	toolsMenu := qt.NewQMenu2()
	toolsMenu.SetTitle(i18n.T("&Tools"))
	mod.menuBar.AddMenu(toolsMenu)

	settingsAction := toolsMenu.AddAction(i18n.T("&Settings..."))
	settingsAction.OnTriggered(func() {
		mod.logger.Event("Settings menu action triggered")
		mod.showSettingsDialog()
	})

	// Help menu
	helpMenu := qt.NewQMenu2()
	helpMenu.SetTitle(i18n.T("&Help"))
	mod.menuBar.AddMenu(helpMenu)

	guideAction := helpMenu.AddAction(i18n.T("&Getting Started"))
	guideAction.SetShortcut(qt.NewQKeySequence2("F1"))
	guideAction.OnTriggered(mod.showGettingStarted)
	helpMenu.AddSeparator()

	aboutAction := helpMenu.AddAction(i18n.T("&About..."))
	aboutAction.OnTriggered(func() {
		mod.logger.Event("About menu action triggered")
		mod.showAboutDialog()
	})
}

// createWelcomeWidget creates the welcome screen widget
func (mod *GuiModule) createWelcomeWidget() *qt.QWidget {
	widget := qt.NewQWidget(nil)
	layout := qt.NewQVBoxLayout(widget)
	layout.AddStretch()

	logo := qt.NewQLabel(nil)
	logo.SetPixmap(icon.Pixmap(128))
	logo.SetAlignment(qt.AlignHCenter)
	layout.AddWidget(logo.QWidget)
	layout.AddSpacing(12)

	titleLabel := qt.NewQLabel(nil)
	titleLabel.SetText("Recuerdo")
	titleFont := titleLabel.Font()
	titleFont.SetPointSize(26)
	titleFont.SetBold(true)
	titleLabel.SetFont(titleFont)
	titleLabel.SetAlignment(qt.AlignHCenter)
	layout.AddWidget(titleLabel.QWidget)

	subtitleLabel := qt.NewQLabel(nil)
	subtitleLabel.SetText(i18n.T("Learn words, places and more by heart"))
	subtitleFont := subtitleLabel.Font()
	subtitleFont.SetPointSize(13)
	subtitleLabel.SetFont(subtitleFont)
	subtitleLabel.SetAlignment(qt.AlignHCenter)
	subtitleLabel.SetStyleSheet("color: palette(dark);")
	layout.AddWidget(subtitleLabel.QWidget)
	layout.AddSpacing(28)

	buttonsLayout := qt.NewQHBoxLayout2()
	buttonsLayout.AddStretch()
	newLessonBtn := qt.NewQPushButton(nil)
	newLessonBtn.SetText(i18n.T("New Lesson"))
	newLessonBtn.SetMinimumSize2(180, 44)
	newLessonBtn.SetDefault(true)
	newLessonBtn.OnClicked(func() {
		mod.logger.Event("Create New Lesson button clicked")
		mod.showNewLessonDialog()
	})
	buttonsLayout.AddWidget(newLessonBtn.QWidget)

	typeListBtn := qt.NewQPushButton3(i18n.T("Type a List"))
	typeListBtn.SetMinimumSize2(180, 44)
	typeListBtn.SetToolTip(i18n.T("Type or paste words as \"question = answer\" lines"))
	typeListBtn.OnClicked(mod.newLessonFromText)
	buttonsLayout.AddWidget(typeListBtn.QWidget)
	buttonsLayout.AddSpacing(16)
	openLessonBtn := qt.NewQPushButton(nil)
	openLessonBtn.SetText(i18n.T("Open Lesson…"))
	openLessonBtn.SetMinimumSize2(180, 44)
	openLessonBtn.OnClicked(func() {
		mod.logger.Event("Open Lesson button clicked")
		mod.showOpenDialogFrom("BUTTON")
	})
	buttonsLayout.AddWidget(openLessonBtn.QWidget)
	buttonsLayout.AddStretch()
	layout.AddLayout(buttonsLayout.QLayout)
	layout.AddSpacing(20)

	hint := qt.NewQLabel(nil)
	hint.SetText(i18n.T("Opens OpenTeacher lessons (.ot, .otwd), word lists (.csv, .txt) and KWordQuiz (.kvtml) files"))
	hint.SetAlignment(qt.AlignHCenter)
	hint.SetWordWrap(true)
	hint.SetStyleSheet("color: palette(dark);")
	layout.AddWidget(hint.QWidget)

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
					mod.statusBar.ShowMessage(i18n.Tf("Error creating lesson: %v", err))
					return
				}

				// Display lesson in new tab
				mod.displayLessonInTab(newLesson)
				mod.statusBar.ShowMessage(i18n.T("New lesson created successfully"))
			} else {
				log.Printf("[INFO] New lesson dialog was cancelled")
				mod.statusBar.ShowMessage(i18n.T("New lesson dialog created"))
			}
		} else {
			mod.logger.DeadEnd("lessonDialogs module", "does not implement ShowNewLessonDialog() method", "legacy/modules/org/openteacher/interfaces/qt/lessonDialogs/")
			mod.statusBar.ShowMessage(i18n.T("Error: Lesson dialog not available"))
		}
	} else {
		mod.logger.DeadEnd("lessonDialogs system", "No lessonDialogs modules found", "legacy/modules/org/openteacher/interfaces/qt/lessonDialogs/")
		mod.statusBar.ShowMessage(i18n.T("Error: No lesson dialog modules available"))
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
				mod.statusBar.ShowMessage(fmt.Sprintf(i18n.T("Selected file: %s"), fileName))
				mod.loadSelectedFile(fileName)
				mod.logger.Debug("TRACKING: loadSelectedFile() completed - call stack marker D")
			} else {
				mod.logger.Info("File dialog was cancelled")
				mod.statusBar.ShowMessage(i18n.T("Open operation cancelled"))
			}
		} else {
			mod.logger.DeadEnd("fileDialog module", "does not implement OpenFile() method", "legacy/modules/org/openteacher/interfaces/qt/dialogs/")
			mod.statusBar.ShowMessage(i18n.T("Error: File dialog not available"))
		}
	} else {
		mod.logger.DeadEnd("fileDialog system", "No fileDialog modules found", "legacy/modules/org/openteacher/interfaces/qt/dialogs/")
		mod.statusBar.ShowMessage(i18n.T("Error: No file dialog modules available"))
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
		mod.statusBar.ShowMessage(fmt.Sprintf(i18n.T("Error loading file: %v"), err))
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

	statusMsg := fmt.Sprintf(i18n.T("Loaded '%s': %d words"), title, wordCount)
	if testCount > 0 {
		statusMsg += fmt.Sprintf(i18n.T(", %d tests"), testCount)
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
	mod.rememberRecent(fileName)
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
	switch lessonType {
	case "":
		lessonType = "words"
	case "topology", "topography":
		lessonType = "topo" // as the lesson files and widgets call it
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
	if png := datatypeicons.Icon(lesson.DataType); png != nil {
		mod.tabWidget.SetTabIcon(tabIndex, icon.FromPNG(png))
	}
	mod.tabWidget.SetCurrentIndex(tabIndex)
	mod.rememberLesson(lessonWidget, lesson)
	if mod.lastWords != nil && mod.lastWords.QWidget.UnsafePointer() == lessonWidget.UnsafePointer() {
		mod.lastWords.SetOnModified(func() { mod.markModified(lessonWidget) })
	}
	if mod.saveAction != nil {
		mod.saveAction.SetEnabled(true)
		mod.saveAsAction.SetEnabled(true)
		mod.printAction.SetEnabled(true)
	}

	// Update status bar
	statusMsg := fmt.Sprintf(i18n.T("Opened '%s' - %d words"), title, lesson.Data.List.GetWordCount())
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
		topoWidget.SetOnModified(func() { mod.markModified(lessonWidget) })
	case "media":
		mod.logger.Info("Creating media lesson widget for: %s", lesson.Path)
		mediaWidget := media.NewMediaLessonWidget(lesson, mod.mainWindow.QWidget)
		lessonWidget = mediaWidget.QWidget
		mediaWidget.SetOnModified(func() { mod.markModified(lessonWidget) })
	case "words":
		fallthrough
	default:
		// Default to words widget for unknown types or actual words lessons
		mod.logger.Info("Creating words lesson widget for: %s (type: %s)", lesson.Path, lesson.DataType)
		wordsWidget := words.NewWordsLessonWidget(lesson, mod.mainWindow.QWidget)
		if settings, ok := mod.manager.GetDefaultModule("settings"); ok {
			if st, ok := settings.(words.Settings); ok {
				wordsWidget.UseSettings(st)
			}
		}
		lessonWidget = wordsWidget.QWidget
		mod.lastWords = wordsWidget
		if mod.tabWords == nil {
			mod.tabWords = map[unsafe.Pointer]*words.WordsLessonWidget{}
		}
		mod.tabWords[wordsWidget.QWidget.UnsafePointer()] = wordsWidget
	}

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
		mod.statusBar.ShowMessage(i18n.T("No lesson open to show properties"))
		return
	}

	// Try to find lesson dialog module which handles properties
	lessonDialogModules := mod.manager.GetModulesByType("lessonDialogs")
	if len(lessonDialogModules) == 0 {
		mod.logger.DeadEnd("lessonDialogs system", "No lessonDialogs modules found", "internal/modules/interfaces/qt/lessonDialogs/")
		mod.statusBar.ShowMessage(i18n.T("Properties dialog cancelled"))
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
		} else {
			mod.logger.Info("Properties dialog was cancelled or no changes made")
		}
	} else {
		mod.logger.Error("lessonDialogs module doesn't have ShowPropertiesDialog method")
		mod.statusBar.ShowMessage(i18n.T("Error: Properties dialog not available"))
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
			if settingsMod.ShowSettingsDialog() {
				mod.applySettingsToLessons()
				mod.statusBar.ShowMessage(i18n.T("Settings saved"))
			}
		} else {
			mod.logger.DeadEnd("settingsDialog module", "does not implement ShowSettingsDialog() method", "legacy/modules/org/openteacher/interfaces/qt/dialogs/settings/")
			mod.statusBar.ShowMessage(i18n.T("Error: Settings dialog not available"))
		}
	} else {
		mod.logger.DeadEnd("settingsDialog system", "No settingsDialog modules found", "legacy/modules/org/openteacher/interfaces/qt/dialogs/settings/")
		mod.statusBar.ShowMessage(i18n.T("Error: No settings dialog modules available"))
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
			mod.statusBar.ShowMessage(i18n.T("About dialog shown"))
		} else {
			mod.logger.DeadEnd("aboutDialog module", "does not implement ShowAboutDialog() method", "legacy/modules/org/openteacher/interfaces/qt/dialogs/about/")
			mod.statusBar.ShowMessage(i18n.T("About dialog not available"))
		}
	} else {
		mod.logger.DeadEnd("aboutDialog system", "No aboutDialog modules found", "legacy/modules/org/openteacher/interfaces/qt/dialogs/about/")
		mod.statusBar.ShowMessage(i18n.T("Error: No about dialog modules available"))
	}
}

// getCurrentLessonData is the current lesson's properties for the
// properties dialog.
func (mod *GuiModule) getCurrentLessonData() map[string]interface{} {
	l, _ := mod.currentLesson()
	if l == nil {
		return nil
	}
	list := l.Data.List
	return map[string]interface{}{
		"name":             list.Title,
		"questionLanguage": list.QuestionLanguage,
		"answerLanguage":   list.AnswerLanguage,
		"hasLanguages":     l.DataType != "topo" && l.DataType != "media",
		"itemCount":        len(list.Items),
		"sessionCount":     len(list.Tests),
	}
}

// applyProperties puts the properties dialog's title (unless empty) and
// languages in list, and reports whether that changed anything.
func applyProperties(list *lesson.WordList, data map[string]interface{}) bool {
	str := func(k string) string { v, _ := data[k].(string); return v }
	changed := false
	set := func(field *string, v string) {
		if *field != v {
			*field, changed = v, true
		}
	}
	if name := str("name"); name != "" {
		set(&list.Title, name)
	}
	set(&list.QuestionLanguage, str("questionLanguage"))
	set(&list.AnswerLanguage, str("answerLanguage"))
	return changed
}

// updateCurrentLessonData puts the properties dialog's title and
// languages in the current lesson, which then needs saving.
func (mod *GuiModule) updateCurrentLessonData(data map[string]interface{}) {
	l, _ := mod.currentLesson()
	if l == nil || data == nil {
		return
	}
	if !applyProperties(&l.Data.List, data) {
		return
	}
	l.Data.Changed = true
	tab := mod.tabWidget.CurrentWidget()
	mod.tabWidget.SetTabText(mod.tabWidget.CurrentIndex(), l.Data.List.Title)
	mod.markModified(tab)
}

// InitGuiModule creates and returns a new GuiModule instance
func InitGuiModule() core.Module {
	return NewGuiModule()
}

// currentLesson is the lesson in the current tab and its tab index.
func (mod *GuiModule) currentLesson() (*lesson.Lesson, int) {
	if mod.tabWidget == nil {
		return nil, -1
	}
	i := mod.tabWidget.CurrentIndex()
	if i < 0 {
		return nil, -1
	}
	return mod.tabLessons[mod.tabWidget.Widget(i).UnsafePointer()], i
}

// markModified puts a "*" before the title of the tab showing widget.
func (mod *GuiModule) markModified(widget *qt.QWidget) {
	i := mod.tabWidget.IndexOf(widget)
	if i >= 0 && !strings.HasPrefix(mod.tabWidget.TabText(i), "*") {
		mod.tabWidget.SetTabText(i, "*"+mod.tabWidget.TabText(i))
	}
}

// saveCurrentLesson saves the lesson in the current tab: to its own file
// if that is in a format Recuerdo writes, otherwise (or for Save As) to a
// file the user chooses.
func (mod *GuiModule) saveCurrentLesson(as bool) {
	l, tab := mod.currentLesson()
	if l == nil {
		return
	}
	path := l.Path
	if as || path == "" || !export.CanSave(path) {
		fd, ok := mod.manager.GetDefaultModule("fileDialog")
		chooser, ok2 := fd.(interface {
			SaveFile(parent *qt.QWidget, title, filter, defaultName string) string
		})
		if !ok || !ok2 {
			mod.logger.Error("no file dialog to choose where to save")
			return
		}
		name := l.Data.List.Title
		if name == "" {
			name = "lesson"
		}
		filter := export.SaveFilterFor(&l.Data)
		path = chooser.SaveFile(mod.mainWindow.QWidget, "Save Lesson", filter, name+export.DefaultExtension(&l.Data))
		if path == "" {
			return
		}
		path = export.WithExtension(path, filter)
	}
	if err := mod.SaveCurrentLessonTo(path); err != nil {
		qt.QMessageBox_Warning(mod.mainWindow.QWidget, "Save Lesson", "Could not save the lesson:\n"+err.Error())
	}
	_ = tab
}

// SaveCurrentLessonTo saves the lesson in the current tab to path, in the
// format its extension names, and makes path the lesson's file.
func (mod *GuiModule) SaveCurrentLessonTo(path string) error {
	l, tab := mod.currentLesson()
	if l == nil {
		return fmt.Errorf("no lesson is open")
	}
	if err := export.Save(&l.Data, path); err != nil {
		return err
	}
	if !readable(path) {
		// an export (PDF, Word, ...): the lesson keeps its own file
		mod.statusBar.ShowMessage(i18n.Tf("Exported %s", path))
		return nil
	}
	l.Path = path
	mod.rememberRecent(path)
	title := l.Data.List.Title
	if title == "" {
		title = filepath.Base(path)
	}
	mod.tabWidget.SetTabText(tab, title)
	mod.statusBar.ShowMessage(i18n.Tf("Saved %s", path))
	mod.logger.Success("Saved lesson to %s", path)
	return nil
}

// rememberLesson records which lesson a tab shows.
func (mod *GuiModule) rememberLesson(tab *qt.QWidget, l *lesson.Lesson) {
	if mod.tabLessons == nil {
		mod.tabLessons = map[unsafe.Pointer]*lesson.Lesson{}
	}
	mod.tabLessons[tab.UnsafePointer()] = l
}

// showGettingStarted shows the getting started guide (Help > Getting
// Started).
func (mod *GuiModule) showGettingStarted() {
	dialog, err := mod.gettingStartedDialog()
	if err != nil {
		qt.QMessageBox_Warning(mod.mainWindow.QWidget, "Getting Started", "The guide could not be found:\n"+err.Error())
		return
	}
	dialog.Show()
}

// gettingStartedDialog builds the dialog showing the getting started guide.
func (mod *GuiModule) gettingStartedDialog() (*qt.QDialog, error) {
	html, err := userdocumentation.GettingStarted()
	if err != nil {
		return nil, err
	}
	dialog := qt.NewQDialog(mod.mainWindow.QWidget)
	dialog.SetWindowTitle(i18n.T("Getting Started with Recuerdo"))
	dialog.Resize(640, 720)
	layout := qt.NewQVBoxLayout(dialog.QWidget)
	browser := qt.NewQTextBrowser(dialog.QWidget)
	browser.SetSearchPaths([]string{userdocumentation.Dir()})
	browser.SetOpenExternalLinks(true)
	browser.SetHtml(html)
	layout.AddWidget(browser.QWidget)
	buttons := qt.NewQDialogButtonBox(dialog.QWidget)
	buttons.SetStandardButtons(qt.QDialogButtonBox__Close)
	buttons.OnRejected(func() { dialog.Close() })
	layout.AddWidget(buttons.QWidget)
	return dialog, nil
}

// settings is the settings module, for the recently opened list.
func (mod *GuiModule) settings() recentlyopened.Settings {
	m, ok := mod.manager.GetDefaultModule("settings")
	if !ok {
		return nil
	}
	s, _ := m.(recentlyopened.Settings)
	return s
}

// rememberRecent puts path at the top of File > Open Recent.
func (mod *GuiModule) rememberRecent(path string) {
	if s := mod.settings(); s != nil {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		recentlyopened.Add(s, path)
	}
}

// fillRecentMenu lists the recently opened lessons that still exist.
func (mod *GuiModule) fillRecentMenu(menu *qt.QMenu) {
	menu.Clear()
	s := mod.settings()
	var paths []string
	if s != nil {
		for _, p := range recentlyopened.List(s) {
			if _, err := os.Stat(p); err == nil {
				paths = append(paths, p)
			}
		}
	}
	if len(paths) == 0 {
		menu.AddAction(i18n.T("No recent lessons")).SetEnabled(false)
		return
	}
	for i, p := range paths {
		path := p
		label := filepath.Base(p)
		if i < 9 {
			label = fmt.Sprintf("&%d  %s", i+1, label)
		}
		a := menu.AddAction(label)
		a.SetToolTip(p)
		a.OnTriggered(func() { mod.loadSelectedFile(path) })
	}
	menu.AddSeparator()
	menu.AddAction(i18n.T("Clear List")).OnTriggered(func() { recentlyopened.Clear(s) })
}

// mergeIntoCurrentLesson adds the words and results of a lesson the user
// chooses to the current word lesson (File > Merge Lesson).
func (mod *GuiModule) mergeIntoCurrentLesson() {
	l, tab := mod.currentLesson()
	if l == nil || l.DataType == "topo" || l.DataType == "media" {
		qt.QMessageBox_Information(mod.mainWindow.QWidget, "Merge Lesson", "Open a word lesson to merge another one into.")
		return
	}
	fd, ok := mod.manager.GetDefaultModule("fileDialog")
	chooser, ok2 := fd.(interface {
		OpenFile(parent interface{}, title, filter string) string
	})
	if !ok || !ok2 {
		return
	}
	path := chooser.OpenFile(mod.mainWindow.QWidget, "Merge Lesson", "")
	if path == "" {
		return
	}
	other, err := lesson.NewFileLoader().LoadFile(path)
	if err != nil {
		qt.QMessageBox_Warning(mod.mainWindow.QWidget, "Merge Lesson", "Could not open "+filepath.Base(path)+":\n"+err.Error())
		return
	}
	lesson.Merge(&l.Data.List, other.List)
	widget := mod.tabWidget.Widget(tab)
	if w := mod.tabWords[widget.UnsafePointer()]; w != nil {
		w.UpdateLesson(l)
	}
	mod.markModified(widget)
	mod.statusBar.ShowMessage(fmt.Sprintf(i18n.T("Merged %d words from %s"), len(other.List.Items), filepath.Base(path)))
}

// readable reports whether Recuerdo can open files like path again, so it
// can become the lesson's file (a PDF, say, is only an export).
func readable(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range lesson.NewFileLoader().GetSupportedExtensions() {
		if e == ext {
			return true
		}
	}
	return false
}

// printCurrentLesson prints the current lesson (File > Print).
func (mod *GuiModule) printCurrentLesson() {
	l, _ := mod.currentLesson()
	if l == nil {
		return
	}
	printer := printsupport.NewQPrinter()
	defer printer.Delete()
	dialog := printsupport.NewQPrintDialog4(printer, mod.mainWindow.QWidget)
	if dialog.Exec() != int(qt.QDialog__Accepted) {
		return
	}
	if err := export.Print(&l.Data, printer); err != nil {
		qt.QMessageBox_Warning(mod.mainWindow.QWidget, "Print", "Could not print the lesson:\n"+err.Error())
		return
	}
	mod.statusBar.ShowMessage(i18n.Tf("Printed %s", l.Data.List.Title))
}

// importFromPicture reads a word list from a picture into a new lesson
// (File > Import from Picture), as OpenTeacher's OCR wizard did.
func (mod *GuiModule) importFromPicture() {
	path := qt.QFileDialog_GetOpenFileName4(mod.mainWindow.QWidget, "Import from Picture", "",
		"Pictures (*.png *.jpg *.jpeg *.tif *.tiff *.bmp *.gif *.webp);;All files (*)")
	if path == "" {
		return
	}
	dialog, err := ocrimport.New(mod.mainWindow.QWidget, path)
	if err != nil {
		qt.QMessageBox_Warning(mod.mainWindow.QWidget, "Import from Picture", "Could not open the picture "+path)
		return
	}
	defer dialog.Delete()
	if dialog.Exec() != int(qt.QDialog__Accepted) {
		return
	}
	items, err := dialog.Words()
	switch {
	case errors.Is(err, ocr.ErrNoTesseract):
		qt.QMessageBox_Information(mod.mainWindow.QWidget, "Import from Picture",
			"Reading pictures needs Tesseract, a free text recognition program. Install it "+
				"(on Arch: pacman -S tesseract tesseract-data-eng) and try again.")
		return
	case err != nil:
		qt.QMessageBox_Warning(mod.mainWindow.QWidget, "Import from Picture", "Could not read the picture:\n"+err.Error())
		return
	case len(items) == 0:
		qt.QMessageBox_Information(mod.mainWindow.QWidget, "Import from Picture",
			"No word pairs were found. Straighten the picture or crop it to the list, and try again.")
		return
	}
	l := lesson.NewLesson("words")
	l.Data.List.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	l.Data.List.Items = items
	mod.displayLessonInTab(l)
	if tab := mod.tabWidget.CurrentWidget(); tab != nil {
		mod.markModified(tab)
	}
	mod.statusBar.ShowMessage(fmt.Sprintf(i18n.T("Read %d word pairs from the picture: check them on the Enter tab"), len(items)))
}

// applySettingsToLessons gives the open word lessons the current settings
// (after the settings dialog changed them).
func (mod *GuiModule) applySettingsToLessons() {
	settings, ok := mod.manager.GetDefaultModule("settings")
	if !ok {
		return
	}
	st, ok := settings.(words.Settings)
	if !ok {
		return
	}
	for _, w := range mod.tabWords {
		w.UseSettings(st)
	}
}

// newLessonFromText makes a word lesson from a typed or pasted list
// (OpenTeacher's plain text enterer).
func (mod *GuiModule) newLessonFromText() {
	d := plaintextwords.New(mod.mainWindow.QWidget)
	defer d.Delete()
	if d.Exec() != int(qt.QDialog__Accepted) {
		return
	}
	if l := d.Lesson(); l != nil {
		mod.displayLessonInTab(l)
		if tab := mod.tabWidget.CurrentWidget(); tab != nil {
			mod.markModified(tab)
		}
		mod.statusBar.ShowMessage(fmt.Sprintf(i18n.T("Made a lesson of %d words"), len(l.Data.List.Items)))
	}
}
