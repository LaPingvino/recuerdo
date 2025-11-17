// Package topo provides topography lesson widget functionality
package topo

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/maps"
	"github.com/mappu/miqt/qt"
)

// SimplePage represents a dedicated page for specific functionality
type SimplePage struct {
	name   string
	widget *qt.QWidget
	layout *qt.QVBoxLayout
	scroll *qt.QScrollArea
}

// NewSimplePage creates a new paginated section
func NewSimplePage(name string, parent *qt.QWidget) *SimplePage {
	// Create scroll area for the page
	scrollArea := qt.NewQScrollArea(parent)
	scrollArea.SetWidgetResizable(true)
	scrollArea.SetVerticalScrollBarPolicy(qt.ScrollBarAsNeeded)
	scrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAsNeeded)

	// Create content widget
	contentWidget := qt.NewQWidget(nil)
	layout := qt.NewQVBoxLayout(contentWidget)
	layout.SetSpacing(15)
	layout.SetContentsMargins(20, 20, 20, 20)
	contentWidget.SetLayout(layout.QLayout)

	scrollArea.SetWidget(contentWidget)

	log.Printf("[SIMPLE-LAYOUT] Created page '%s' with scroll area", name)

	return &SimplePage{
		name:   name,
		widget: contentWidget,
		layout: layout,
		scroll: scrollArea,
	}
}

// AddWidget adds a widget to the page with proper spacing
func (sp *SimplePage) AddWidget(widget *qt.QWidget, fixedHeight int) {
	if fixedHeight > 0 {
		widget.SetFixedHeight(fixedHeight)
	}

	// Use stored layout reference directly
	sp.layout.AddWidget(widget)
}

// AddStretch adds flexible space to push content up
func (sp *SimplePage) AddStretch() {
	// Use stored layout reference directly
	sp.layout.AddStretch()
}

// GetScrollArea returns the scroll area widget for adding to parent layouts
func (sp *SimplePage) GetScrollArea() *qt.QScrollArea {
	return sp.scroll
}

// TopoLessonWidget handles topography lesson display and interaction
type TopoLessonWidget struct {
	*qt.QWidget
	lesson *lesson.Lesson

	// Tab widget for different views
	tabWidget *qt.QTabWidget

	// Enter tab (edit mode)
	enterTab   *qt.QWidget
	teachTab   *qt.QWidget
	resultsTab *qt.QWidget
	tilesTab   *qt.QWidget

	// Enter tab components
	enterLayout  *qt.QVBoxLayout
	placesList   *qt.QListWidget
	addButton    *qt.QPushButton
	removeButton *qt.QPushButton
	clearButton  *qt.QPushButton
	importButton *qt.QPushButton
	saveButton   *qt.QPushButton

	// Map selection components
	mapSelectorGroup *qt.QGroupBox
	mapComboBox      *qt.QComboBox
	loadMapButton    *qt.QPushButton

	// Tile maps selection
	tileMapGroup        *qt.QGroupBox
	tileMapComboBox     *qt.QComboBox
	zoomSpinBox         *qt.QSpinBox
	northSpinBox        *qt.QDoubleSpinBox
	southSpinBox        *qt.QDoubleSpinBox
	eastSpinBox         *qt.QDoubleSpinBox
	westSpinBox         *qt.QDoubleSpinBox
	downloadTilesButton *qt.QPushButton
	createTileMapButton *qt.QPushButton

	// Teach tab components
	teachLayout   *qt.QVBoxLayout
	questionLabel *qt.QLabel
	answerInput   *qt.QLineEdit
	clickToAnswer *qt.QCheckBox
	submitButton  *qt.QPushButton
	nextButton    *qt.QPushButton
	scoreLabel    *qt.QLabel
	progressLabel *qt.QLabel

	// Results tab components
	resultsLayout *qt.QVBoxLayout
	resultsTable  *qt.QTableWidget
	statsLabel    *qt.QLabel

	// Map components
	mapScrollArea  *qt.QScrollArea
	mapWidget      *qt.QWidget
	mapLabel       *qt.QLabel
	mapOverlay     *qt.QWidget
	mapMarkers     []*qt.QPushButton
	mapPixmap      *qt.QPixmap
	teachMapWidget *qt.QWidget
	teachMapLabel  *qt.QLabel

	// Map management
	mapManager  *maps.MapManager
	currentMap  *maps.BaseMap
	addingPlace bool

	// Simplified page management
	editPage     *SimplePage
	practicePage *SimplePage

	// Teaching state
	currentIndex int
	score        int
	totalAnswers int
}

// NewTopoLessonWidget creates a new topography lesson widget
func NewTopoLessonWidget(lesson *lesson.Lesson, parent *qt.QWidget) *TopoLessonWidget {
	widget := &TopoLessonWidget{
		QWidget: qt.NewQWidget(parent),
		lesson:  lesson,
	}

	// Initialize map manager
	widget.mapManager = maps.NewMapManager("./")
	if err := widget.mapManager.LoadAvailableMaps(); err != nil {
		log.Printf("Warning: Failed to load available maps: %v", err)
	}

	widget.setupUI()
	widget.setupMapWidget()
	widget.updateData()
	widget.connectSignals()

	return widget
}

// ValidateLayoutAfterShow validates the simplified layout
func (w *TopoLessonWidget) ValidateLayoutAfterShow() {
	strictMode := os.Getenv("RECUERDO_STRICT_LAYOUT") == "1"
	if !strictMode {
		return
	}

	log.Printf("[SIMPLE-LAYOUT] Layout validation passed - paginated design prevents overlaps")
}

// setupUI initializes the user interface with relative layout
func (w *TopoLessonWidget) setupUI() {
	// Main layout
	mainLayout := qt.NewQVBoxLayout(w.QWidget)
	w.SetLayout(mainLayout.QLayout)

	// Tab widget
	w.tabWidget = qt.NewQTabWidget(w.QWidget)
	mainLayout.AddWidget(w.tabWidget.QWidget)

	// Setup tabs
	w.setupEnterTab()
	w.setupTeachTab()
	w.setupResultsTab()
	w.setupTilesTab()
}

// setupEnterTab creates the places entry tab with simplified pagination
func (w *TopoLessonWidget) setupEnterTab() {
	w.enterTab = qt.NewQWidget(w.QWidget)
	mainLayout := qt.NewQVBoxLayout(w.enterTab)
	w.enterTab.SetLayout(mainLayout.QLayout)

	// Create a sub-tab widget for different editing sections
	editTabWidget := qt.NewQTabWidget(w.enterTab)
	mainLayout.AddWidget(editTabWidget.QWidget)

	// Create editing sub-pages
	w.setupPlacesListPage(editTabWidget)
	w.setupMapEditPage(editTabWidget)

	// Add tab to main widget
	w.tabWidget.AddTab(w.enterTab, "Enter Places")
}

// setupPlacesListPage creates a dedicated page for place list management
func (w *TopoLessonWidget) setupPlacesListPage(parent *qt.QTabWidget) {
	w.editPage = NewSimplePage("PlacesList", parent.QWidget)

	// Title
	titleLabel := qt.NewQLabel(nil)
	titleLabel.SetText("Places to Learn")
	titleLabel.SetStyleSheet("font-size: 16px; font-weight: bold; color: #333;")
	w.editPage.AddWidget(titleLabel.QWidget, 40)

	// Places list
	w.placesList = qt.NewQListWidget(nil)
	w.editPage.AddWidget(w.placesList.QWidget, 200)

	// Button toolbar
	buttonContainer := qt.NewQWidget(nil)
	buttonLayout := qt.NewQHBoxLayout(buttonContainer)
	buttonContainer.SetLayout(buttonLayout.QLayout)

	w.addButton = qt.NewQPushButton(nil)
	w.addButton.SetText("Add Place")
	w.addButton.SetStyleSheet("padding: 8px 16px; background: #4CAF50; color: white; border-radius: 4px;")
	buttonLayout.AddWidget(w.addButton.QWidget)

	w.removeButton = qt.NewQPushButton(nil)
	w.removeButton.SetText("Remove")
	w.removeButton.SetStyleSheet("padding: 8px 16px; background: #f44336; color: white; border-radius: 4px;")
	buttonLayout.AddWidget(w.removeButton.QWidget)

	w.clearButton = qt.NewQPushButton(nil)
	w.clearButton.SetText("Clear All")
	w.clearButton.SetStyleSheet("padding: 8px 16px; background: #FF9800; color: white; border-radius: 4px;")
	buttonLayout.AddWidget(w.clearButton.QWidget)

	w.importButton = qt.NewQPushButton(nil)
	w.importButton.SetText("Import...")
	w.importButton.SetStyleSheet("padding: 8px 16px; background: #2196F3; color: white; border-radius: 4px;")
	buttonLayout.AddWidget(w.importButton.QWidget)

	w.saveButton = qt.NewQPushButton(nil)
	w.saveButton.SetText("Save")
	w.saveButton.SetStyleSheet("padding: 8px 16px; background: #9C27B0; color: white; border-radius: 4px;")
	buttonLayout.AddWidget(w.saveButton.QWidget)

	buttonLayout.AddStretch()
	w.editPage.AddWidget(buttonContainer, 60)

	// Add flexible space
	w.editPage.AddStretch()

	parent.AddTab(w.editPage.GetScrollArea().QWidget, "Places List")
}

// setupMapEditPage creates a dedicated page for map-based editing
func (w *TopoLessonWidget) setupMapEditPage(parent *qt.QTabWidget) {
	mapPage := NewSimplePage("MapEdit", parent.QWidget)

	// Map selector section
	w.setupMapSelector()
	mapPage.AddWidget(w.mapSelectorGroup.QWidget, 80)

	// Map display
	w.setupMapWidget()
	mapPage.AddWidget(w.mapScrollArea.QWidget, 350)

	// Instructions
	instructionsLabel := qt.NewQLabel(nil)
	instructionsLabel.SetText("Select a base map above, then click 'Add Place' from the Places List tab to add places by clicking on the map.")
	instructionsLabel.SetWordWrap(true)
	instructionsLabel.SetStyleSheet("color: #666; font-style: italic; padding: 10px; background: #f5f5f5; border-radius: 4px;")
	mapPage.AddWidget(instructionsLabel.QWidget, 60)

	mapPage.AddStretch()

	parent.AddTab(mapPage.GetScrollArea().QWidget, "Map Editor")
}

// setupTeachTab creates the teaching tab
func (w *TopoLessonWidget) setupTeachTab() {
	w.teachTab = qt.NewQWidget(w.QWidget)
	mainLayout := qt.NewQVBoxLayout(w.teachTab)
	w.teachTab.SetLayout(mainLayout.QLayout)

	// Create paginated practice layout
	w.practicePage = NewSimplePage("Practice", w.teachTab)

	// Progress section
	progressContainer := qt.NewQWidget(nil)
	progressLayout := qt.NewQHBoxLayout(progressContainer)
	progressContainer.SetLayout(progressLayout.QLayout)

	w.progressLabel = qt.NewQLabel(nil)
	w.progressLabel.SetText("Progress: 0/0")
	w.progressLabel.SetStyleSheet("font-size: 14px; font-weight: bold; color: #333;")
	progressLayout.AddWidget(w.progressLabel.QWidget)

	w.scoreLabel = qt.NewQLabel(nil)
	w.scoreLabel.SetText("Score: 0/0 (0%)")
	w.scoreLabel.SetStyleSheet("font-size: 14px; font-weight: bold; color: #2196F3;")
	progressLayout.AddWidget(w.scoreLabel.QWidget)

	progressLayout.AddStretch()
	w.practicePage.AddWidget(progressContainer, 50)

	// Question section
	questionContainer := qt.NewQWidget(nil)
	questionContainer.SetStyleSheet("background: #f9f9f9; border: 1px solid #ddd; border-radius: 8px; padding: 15px;")
	questionLayout := qt.NewQVBoxLayout(questionContainer)
	questionContainer.SetLayout(questionLayout.QLayout)

	instructionLabel := qt.NewQLabel(nil)
	instructionLabel.SetText("What is the name of this place?")
	instructionLabel.SetStyleSheet("font-size: 12px; color: #666; margin-bottom: 10px;")
	questionLayout.AddWidget(instructionLabel.QWidget)

	w.questionLabel = qt.NewQLabel(nil)
	w.questionLabel.SetText("Click 'Next' to start practice")
	w.questionLabel.SetStyleSheet("font-size: 18px; font-weight: bold; color: #333; text-align: center; padding: 20px;")
	w.questionLabel.SetAlignment(qt.AlignCenter)
	w.questionLabel.SetWordWrap(true)
	questionLayout.AddWidget(w.questionLabel.QWidget)

	w.practicePage.AddWidget(questionContainer, 120)

	// Answer section
	answerContainer := qt.NewQWidget(nil)
	answerContainer.SetStyleSheet("background: #fff; border: 1px solid #ddd; border-radius: 8px; padding: 15px;")
	answerLayout := qt.NewQVBoxLayout(answerContainer)
	answerContainer.SetLayout(answerLayout.QLayout)

	answerLabel := qt.NewQLabel(nil)
	answerLabel.SetText("Your Answer:")
	answerLabel.SetStyleSheet("font-size: 12px; color: #666; margin-bottom: 5px;")
	answerLayout.AddWidget(answerLabel.QWidget)

	w.answerInput = qt.NewQLineEdit(nil)
	w.answerInput.SetPlaceholderText("Type the name of the place...")
	w.answerInput.SetStyleSheet("font-size: 16px; padding: 10px; border: 2px solid #ddd; border-radius: 4px;")
	answerLayout.AddWidget(w.answerInput.QWidget)

	w.clickToAnswer = qt.NewQCheckBox(nil)
	w.clickToAnswer.SetText("Click on map to answer (disable text input)")
	w.clickToAnswer.SetStyleSheet("margin-top: 10px; color: #666;")
	answerLayout.AddWidget(w.clickToAnswer.QWidget)

	w.practicePage.AddWidget(answerContainer, 120)

	// Control buttons
	buttonContainer := qt.NewQWidget(nil)
	buttonLayout := qt.NewQHBoxLayout(buttonContainer)
	buttonContainer.SetLayout(buttonLayout.QLayout)

	w.submitButton = qt.NewQPushButton(nil)
	w.submitButton.SetText("Submit Answer")
	w.submitButton.SetStyleSheet("font-size: 14px; padding: 10px 20px; background-color: #4CAF50; color: white; border-radius: 4px; font-weight: bold;")
	w.submitButton.SetEnabled(false)
	buttonLayout.AddWidget(w.submitButton.QWidget)

	w.nextButton = qt.NewQPushButton(nil)
	w.nextButton.SetText("Next Place")
	w.nextButton.SetStyleSheet("font-size: 14px; padding: 10px 20px; background-color: #2196F3; color: white; border-radius: 4px; font-weight: bold;")
	buttonLayout.AddWidget(w.nextButton.QWidget)

	buttonLayout.AddStretch()
	w.practicePage.AddWidget(buttonContainer, 60)

	// Teaching map section
	w.setupTeachingMapWidget()
	w.practicePage.AddWidget(w.teachMapWidget, 280)

	// Add flexible space
	w.practicePage.AddStretch()

	// Add the practice page to the main layout
	mainLayout.AddWidget(w.practicePage.GetScrollArea().QWidget)

	// Add tab
	w.tabWidget.AddTab(w.teachTab, "Practice")
}

// setupResultsTab creates the results tab
func (w *TopoLessonWidget) setupResultsTab() {
	w.resultsTab = qt.NewQWidget(w.QWidget)
	w.resultsLayout = qt.NewQVBoxLayout(w.resultsTab)
	w.resultsTab.SetLayout(w.resultsLayout.QLayout)

	// Statistics
	w.statsLabel = qt.NewQLabel(w.resultsTab)
	w.statsLabel.SetText("No test results yet")
	w.statsLabel.SetStyleSheet("font-size: 14px; margin: 10px;")
	w.resultsLayout.AddWidget(w.statsLabel.QWidget)

	// Results table
	w.resultsTable = qt.NewQTableWidget2()
	w.resultsTable.SetRowCount(0)
	w.resultsTable.SetColumnCount(2)

	headers := []string{"Place Name", "Result"}
	for i, header := range headers {
		item := qt.NewQTableWidgetItem2(header)
		w.resultsTable.SetHorizontalHeaderItem(i, item)
	}

	w.resultsTable.SetAlternatingRowColors(true)
	w.resultsTable.SetSelectionBehavior(qt.QAbstractItemView__SelectRows)
	w.resultsLayout.AddWidget(w.resultsTable.QWidget)

	// Add tab
	w.tabWidget.AddTab(w.resultsTab, "Results")
}

// setupTilesTab creates the tile maps tab
func (w *TopoLessonWidget) setupTilesTab() {
	w.tilesTab = qt.NewQWidget(w.QWidget)
	mainLayout := qt.NewQVBoxLayout(w.tilesTab)
	w.tilesTab.SetLayout(mainLayout.QLayout)

	// Create paginated tile selection layout
	tilePage := NewSimplePage("TileSelection", w.tilesTab)

	// Title
	titleLabel := qt.NewQLabel(nil)
	titleLabel.SetText("Online Map Tiles")
	titleLabel.SetStyleSheet("font-size: 18px; font-weight: bold; color: #333; margin-bottom: 10px;")
	tilePage.AddWidget(titleLabel.QWidget, 30)

	// Description
	descLabel := qt.NewQLabel(nil)
	descLabel.SetText("Download real-world map tiles for topography training. CartoDB Light (No Labels) is recommended for learning place names without hints.")
	descLabel.SetWordWrap(true)
	descLabel.SetStyleSheet("color: #666; font-size: 13px; margin-bottom: 15px; padding: 10px; background: #f0f8ff; border-radius: 4px;")
	tilePage.AddWidget(descLabel.QWidget, 60)

	// Tile map selection
	w.setupTileMapSelector()
	tilePage.AddWidget(w.tileMapGroup.QWidget, 200)

	// Geographic bounds configuration
	boundsGroup := qt.NewQGroupBox(nil)
	boundsGroup.SetTitle("Geographic Bounds")
	boundsGroup.SetStyleSheet("QGroupBox { font-weight: bold; margin-top: 10px; } QGroupBox::title { subcontrol-origin: margin; left: 10px; padding: 0 5px; }")
	boundsLayout := qt.NewQGridLayout2()
	boundsGroup.SetLayout(boundsLayout.QLayout)

	// North
	northLabel := qt.NewQLabel2()
	northLabel.SetText("North Latitude:")
	boundsLayout.AddWidget2(northLabel.QWidget, 0, 0)
	w.northSpinBox = qt.NewQDoubleSpinBox(nil)
	w.northSpinBox.SetRange(-90.0, 90.0)
	w.northSpinBox.SetValue(60.0)
	w.northSpinBox.SetDecimals(4)
	w.northSpinBox.SetSingleStep(0.1)
	boundsLayout.AddWidget2(w.northSpinBox.QWidget, 0, 1)

	// South
	southLabel := qt.NewQLabel2()
	southLabel.SetText("South Latitude:")
	boundsLayout.AddWidget2(southLabel.QWidget, 1, 0)
	w.southSpinBox = qt.NewQDoubleSpinBox(nil)
	w.southSpinBox.SetRange(-90.0, 90.0)
	w.southSpinBox.SetValue(40.0)
	w.southSpinBox.SetDecimals(4)
	w.southSpinBox.SetSingleStep(0.1)
	boundsLayout.AddWidget2(w.southSpinBox.QWidget, 1, 1)

	// East
	eastLabel := qt.NewQLabel2()
	eastLabel.SetText("East Longitude:")
	boundsLayout.AddWidget2(eastLabel.QWidget, 0, 2)
	w.eastSpinBox = qt.NewQDoubleSpinBox(nil)
	w.eastSpinBox.SetRange(-180.0, 180.0)
	w.eastSpinBox.SetValue(20.0)
	w.eastSpinBox.SetDecimals(4)
	w.eastSpinBox.SetSingleStep(0.1)
	boundsLayout.AddWidget2(w.eastSpinBox.QWidget, 0, 3)

	// West
	westLabel := qt.NewQLabel2()
	westLabel.SetText("West Longitude:")
	boundsLayout.AddWidget2(westLabel.QWidget, 1, 2)
	w.westSpinBox = qt.NewQDoubleSpinBox(nil)
	w.westSpinBox.SetRange(-180.0, 180.0)
	w.westSpinBox.SetValue(-10.0)
	w.westSpinBox.SetDecimals(4)
	w.westSpinBox.SetSingleStep(0.1)
	boundsLayout.AddWidget2(w.westSpinBox.QWidget, 1, 3)

	// Zoom level
	zoomLabel := qt.NewQLabel2()
	zoomLabel.SetText("Zoom Level:")
	boundsLayout.AddWidget2(zoomLabel.QWidget, 2, 0)
	w.zoomSpinBox = qt.NewQSpinBox(nil)
	w.zoomSpinBox.SetRange(0, 18)
	w.zoomSpinBox.SetValue(6)
	boundsLayout.AddWidget2(w.zoomSpinBox.QWidget, 2, 1)

	tilePage.AddWidget(boundsGroup.QWidget, 140)

	// Preset buttons
	presetGroup := qt.NewQGroupBox(nil)
	presetGroup.SetTitle("Geographic Presets")
	presetLayout := qt.NewQHBoxLayout(presetGroup.QWidget)
	presetGroup.SetLayout(presetLayout.QLayout)

	europeBtn := qt.NewQPushButton(nil)
	europeBtn.SetText("Europe")
	europeBtn.SetStyleSheet("padding: 8px 16px; background: #2196F3; color: white; border-radius: 4px;")
	europeBtn.OnClicked(func() {
		w.northSpinBox.SetValue(71.0)
		w.southSpinBox.SetValue(35.0)
		w.eastSpinBox.SetValue(40.0)
		w.westSpinBox.SetValue(-25.0)
	})
	presetLayout.AddWidget(europeBtn.QWidget)

	usaBtn := qt.NewQPushButton(nil)
	usaBtn.SetText("USA")
	usaBtn.SetStyleSheet("padding: 8px 16px; background: #2196F3; color: white; border-radius: 4px;")
	usaBtn.OnClicked(func() {
		w.northSpinBox.SetValue(50.0)
		w.southSpinBox.SetValue(25.0)
		w.eastSpinBox.SetValue(-65.0)
		w.westSpinBox.SetValue(-125.0)
	})
	presetLayout.AddWidget(usaBtn.QWidget)

	worldBtn := qt.NewQPushButton(nil)
	worldBtn.SetText("World")
	worldBtn.SetStyleSheet("padding: 8px 16px; background: #2196F3; color: white; border-radius: 4px;")
	worldBtn.OnClicked(func() {
		w.northSpinBox.SetValue(80.0)
		w.southSpinBox.SetValue(-60.0)
		w.eastSpinBox.SetValue(180.0)
		w.westSpinBox.SetValue(-180.0)
	})
	presetLayout.AddWidget(worldBtn.QWidget)

	presetLayout.AddStretch()
	tilePage.AddWidget(presetGroup.QWidget, 60)

	// Action buttons
	buttonContainer := qt.NewQWidget(nil)
	buttonLayout := qt.NewQHBoxLayout(buttonContainer)
	buttonContainer.SetLayout(buttonLayout.QLayout)

	w.downloadTilesButton = qt.NewQPushButton(nil)
	w.downloadTilesButton.SetText("Download & Cache Tiles")
	w.downloadTilesButton.SetStyleSheet("padding: 10px 20px; background: #FF9800; color: white; border-radius: 4px; font-weight: bold;")
	buttonLayout.AddWidget(w.downloadTilesButton.QWidget)

	w.createTileMapButton = qt.NewQPushButton(nil)
	w.createTileMapButton.SetText("Create Tile-Based Map")
	w.createTileMapButton.SetStyleSheet("padding: 10px 20px; background: #4CAF50; color: white; border-radius: 4px; font-weight: bold;")
	buttonLayout.AddWidget(w.createTileMapButton.QWidget)

	buttonLayout.AddStretch()
	tilePage.AddWidget(buttonContainer, 60)

	// Add flexible space
	tilePage.AddStretch()

	// Add the tile page to the main layout
	mainLayout.AddWidget(tilePage.GetScrollArea().QWidget)

	// Add tab
	w.tabWidget.AddTab(w.tilesTab, "Online Tiles")
}

// validateLayout is no longer needed with simplified pagination
func (w *TopoLessonWidget) validateLayout() {
	// Paginated design prevents overlaps by design
	log.Printf("[SIMPLE-LAYOUT] Using paginated layout - no validation needed")
}

// connectSignals connects UI signals to slots
func (w *TopoLessonWidget) connectSignals() {
	// Map selection
	w.loadMapButton.OnClicked(func() {
		w.handleLoadMap()
	})

	// Answer input - enable submit button when text is entered
	w.answerInput.OnTextChanged(func(text string) {
		if !w.clickToAnswer.IsChecked() {
			w.submitButton.SetEnabled(len(text) > 0)
		}
	})

	// Click mode checkbox
	w.clickToAnswer.OnToggled(func(checked bool) {
		w.answerInput.SetEnabled(!checked)
		if checked {
			w.answerInput.Clear()
			w.submitButton.SetEnabled(false)
		}
	})

	// Submit button
	w.submitButton.OnClicked(func() {
		w.handleSubmitAnswer()
	})

	// Next button
	w.nextButton.OnClicked(func() {
		w.handleNextQuestion()
	})

	// Enter key in answer input
	w.answerInput.OnReturnPressed(func() {
		if w.submitButton.IsEnabled() {
			w.handleSubmitAnswer()
		}
	})

	// Map click handling for adding places and teaching
	w.mapOverlay.SetMouseTracking(true)

	// Install event filter for map clicks
	w.mapOverlay.InstallEventFilter(w.mapOverlay.QObject)

	// Install event filter for teaching map clicks
	if w.teachMapWidget != nil {
		w.teachMapWidget.SetMouseTracking(true)
		w.teachMapWidget.InstallEventFilter(w.teachMapWidget.QObject)
	}

	// List management buttons
	w.addButton.OnClicked(func() {
		w.handleAddPlace()
	})

	w.removeButton.OnClicked(func() {
		w.handleRemovePlace()
	})

	w.clearButton.OnClicked(func() {
		w.handleClearPlaces()
	})

	w.importButton.OnClicked(func() {
		w.handleImportPlaces()
	})

	w.saveButton.OnClicked(func() {
		w.handleSave()
	})

	// Tile map buttons
	if w.downloadTilesButton != nil {
		w.downloadTilesButton.OnClicked(func() {
			w.handleDownloadTiles()
		})
	}

	if w.createTileMapButton != nil {
		w.createTileMapButton.OnClicked(func() {
			w.handleCreateTileMap()
		})
	}
}

// updateData updates the widget with lesson data
func (w *TopoLessonWidget) updateData() {
	if w.lesson == nil {
		return
	}

	// Update places list
	w.placesList.Clear()
	for _, item := range w.lesson.Data.List.Items {
		if len(item.Questions) > 0 {
			displayText := item.Questions[0]
			// Add coordinate information if available
			if x, y, hasCoords := item.GetTopoCoordinates(); hasCoords {
				displayText = fmt.Sprintf("%s (x:%d, y:%d)", item.Questions[0], x, y)
			}
			w.placesList.AddItem(displayText)
		}
	}

	// Reset teaching state
	w.currentIndex = 0
	w.score = 0
	w.totalAnswers = 0
	w.updateProgress()
	w.updateTeachDisplay()
	w.updateMapDisplay()
}

// updateProgress updates the progress display
func (w *TopoLessonWidget) updateProgress() {
	if w.lesson == nil {
		return
	}

	total := len(w.lesson.Data.List.Items)
	w.progressLabel.SetText(fmt.Sprintf("Progress: %d/%d", w.currentIndex+1, total))

	percentage := 0.0
	if w.totalAnswers > 0 {
		percentage = float64(w.score) / float64(w.totalAnswers) * 100
	}
	w.scoreLabel.SetText(fmt.Sprintf("Score: %d/%d (%.0f%%)", w.score, w.totalAnswers, percentage))
}

// updateTeachDisplay updates the teaching interface
func (w *TopoLessonWidget) updateTeachDisplay() {
	if w.lesson == nil || len(w.lesson.Data.List.Items) == 0 {
		w.questionLabel.SetText("No places to practice")
		return
	}

	if w.currentIndex >= len(w.lesson.Data.List.Items) {
		w.questionLabel.SetText("Practice complete!")
		w.submitButton.SetEnabled(false)
		w.nextButton.SetText("Start Over")
		return
	}

	item := w.lesson.Data.List.Items[w.currentIndex]
	if len(item.Questions) > 0 {
		questionText := fmt.Sprintf("Place: %s", item.Questions[0])
		// Add coordinate information if available
		if x, y, hasCoords := item.GetTopoCoordinates(); hasCoords {
			questionText = fmt.Sprintf("Place: %s\nCoordinates: (%d, %d)", item.Questions[0], x, y)
		}
		w.questionLabel.SetText(questionText)
	}

	w.answerInput.Clear()
	w.answerInput.SetFocus()
}

// handleSubmitAnswer processes the submitted answer
func (w *TopoLessonWidget) handleSubmitAnswer() {
	if w.lesson == nil || w.currentIndex >= len(w.lesson.Data.List.Items) {
		return
	}

	userAnswer := w.answerInput.Text()
	item := w.lesson.Data.List.Items[w.currentIndex]

	// Check if answer is correct (case-insensitive)
	correct := false
	if len(item.Answers) > 0 {
		for _, correctAnswer := range item.Answers {
			if len(userAnswer) > 0 && len(correctAnswer) > 0 {
				if userAnswer == correctAnswer {
					correct = true
					break
				}
			}
		}
	}

	w.totalAnswers++
	if correct {
		w.score++
		w.questionLabel.SetText(fmt.Sprintf("✅ Correct! The answer is: %s", item.Answers[0]))
		w.questionLabel.SetStyleSheet("font-size: 18px; font-weight: bold; margin: 20px; color: green;")
	} else {
		expectedAnswer := "Unknown"
		if len(item.Answers) > 0 {
			expectedAnswer = item.Answers[0]
		}
		w.questionLabel.SetText(fmt.Sprintf("❌ Incorrect. The correct answer is: %s", expectedAnswer))
		w.questionLabel.SetStyleSheet("font-size: 18px; font-weight: bold; margin: 20px; color: red;")
	}

	// Add to results table
	w.addResultToTable(item.Questions[0], correct)

	w.updateProgress()
	w.submitButton.SetEnabled(false)
	w.nextButton.SetText("Next Place")
}

// handleNextQuestion moves to the next question
func (w *TopoLessonWidget) handleNextQuestion() {
	if w.lesson == nil {
		return
	}

	w.currentIndex++
	if w.currentIndex >= len(w.lesson.Data.List.Items) {
		// Practice complete, restart
		w.currentIndex = 0
	}

	w.questionLabel.SetStyleSheet("font-size: 18px; font-weight: bold; margin: 20px; text-align: center;")
	w.updateTeachDisplay()
	w.updateProgress()
}

// addResultToTable adds a result to the results table
func (w *TopoLessonWidget) addResultToTable(placeName string, correct bool) {
	row := w.resultsTable.RowCount()
	w.resultsTable.SetRowCount(row + 1)

	placeItem := qt.NewQTableWidgetItem2(placeName)
	w.resultsTable.SetItem(row, 0, placeItem)

	resultText := "❌ Wrong"
	if correct {
		resultText = "✅ Correct"
	}
	resultItem := qt.NewQTableWidgetItem2(resultText)
	w.resultsTable.SetItem(row, 1, resultItem)

	// Update statistics
	w.updateStatistics()
}

// updateStatistics updates the statistics display
func (w *TopoLessonWidget) updateStatistics() {
	if w.totalAnswers == 0 {
		w.statsLabel.SetText("No test results yet")
		return
	}

	percentage := float64(w.score) / float64(w.totalAnswers) * 100
	w.statsLabel.SetText(fmt.Sprintf("Total: %d questions, Correct: %d, Wrong: %d, Score: %.1f%%",
		w.totalAnswers, w.score, w.totalAnswers-w.score, percentage))
}

// setupMapSelector creates the map selection interface
func (w *TopoLessonWidget) setupMapSelector() {
	w.mapSelectorGroup = qt.NewQGroupBox(nil)
	w.mapSelectorGroup.SetTitle("Base Map")
	w.mapSelectorGroup.SetMinimumWidth(500)

	selectorLayout := qt.NewQHBoxLayout(w.mapSelectorGroup.QWidget)
	w.mapSelectorGroup.SetLayout(selectorLayout.QLayout)

	w.mapComboBox = qt.NewQComboBox(w.mapSelectorGroup.QWidget)
	w.mapComboBox.AddItem("Select a base map...")

	// Add available maps
	availableMaps := w.mapManager.GetAvailableMaps()
	for _, baseMap := range availableMaps {
		w.mapComboBox.AddItem(baseMap.Name)
		w.mapComboBox.SetItemData(w.mapComboBox.Count()-1, qt.NewQVariant17(baseMap.ID))
	}

	selectorLayout.AddWidget(w.mapComboBox.QWidget)

	w.loadMapButton = qt.NewQPushButton(w.mapSelectorGroup.QWidget)
	w.loadMapButton.SetText("Load")
	w.loadMapButton.SetStyleSheet("background-color: #4CAF50; color: white; padding: 6px 12px; font-weight: bold;")
	w.loadMapButton.SetFixedWidth(80)
	selectorLayout.AddWidget(w.loadMapButton.QWidget)
}

// setupTileMapSelector creates the tile map selection interface
func (w *TopoLessonWidget) setupTileMapSelector() {
	w.tileMapGroup = qt.NewQGroupBox(nil)
	w.tileMapGroup.SetTitle("Tile Map Source")
	w.tileMapGroup.SetStyleSheet("QGroupBox { font-weight: bold; margin-top: 10px; } QGroupBox::title { subcontrol-origin: margin; left: 10px; padding: 0 5px; }")

	selectorLayout := qt.NewQVBoxLayout(w.tileMapGroup.QWidget)
	w.tileMapGroup.SetLayout(selectorLayout.QLayout)

	w.tileMapComboBox = qt.NewQComboBox(w.tileMapGroup.QWidget)
	w.tileMapComboBox.AddItem("Select a tile map source...")

	// Add available tile maps
	tileManager := w.mapManager.GetTileManager()
	if tileManager != nil {
		availableTileMaps := tileManager.GetAvailableTileMaps()
		for _, tileMap := range availableTileMaps {
			w.tileMapComboBox.AddItem(tileMap.Config.Name)
			w.tileMapComboBox.SetItemData(w.tileMapComboBox.Count()-1, qt.NewQVariant17(tileMap.Config.ID))
		}
	}

	selectorLayout.AddWidget(w.tileMapComboBox.QWidget)

	// Add description label
	descLabel := qt.NewQLabel(nil)
	descLabel.SetText("CartoDB Light (No Labels) is recommended for topography training as it shows terrain without revealing place names.")
	descLabel.SetWordWrap(true)
	descLabel.SetStyleSheet("font-size: 11px; color: #666; margin-top: 5px;")
	selectorLayout.AddWidget(descLabel.QWidget)
}

// setupMapWidget creates the interactive map display widget
func (w *TopoLessonWidget) setupMapWidget() {
	// Create scroll area for the map
	w.mapScrollArea = qt.NewQScrollArea(nil)
	w.mapScrollArea.SetWidgetResizable(false)
	w.mapScrollArea.SetMinimumSize2(580, 320)
	w.mapScrollArea.SetMaximumSize2(580, 320)

	w.mapWidget = qt.NewQWidget(nil)
	w.mapWidget.SetMinimumSize2(580, 320)
	w.mapWidget.SetStyleSheet("background-color: #f0f8ff; border: 2px solid #ddd; border-radius: 8px;")

	// Create map label for background image
	w.mapLabel = qt.NewQLabel(w.mapWidget)
	w.mapLabel.SetGeometry(10, 10, 560, 300)
	w.mapLabel.SetAlignment(qt.AlignCenter)
	w.mapLabel.SetText("Load a base map to start placing locations")
	w.mapLabel.SetStyleSheet("color: #666; font-size: 14px; padding: 20px; background: rgba(255,255,255,0.8); border-radius: 4px;")

	// Create overlay for markers and interactions
	w.mapOverlay = qt.NewQWidget(w.mapWidget)
	w.mapOverlay.SetGeometry(10, 10, 560, 300)
	w.mapOverlay.SetStyleSheet("background: transparent;")

	w.mapScrollArea.SetWidget(w.mapWidget)
	w.mapMarkers = make([]*qt.QPushButton, 0)
}

// setupTeachingMapWidget creates the map widget for teaching mode
func (w *TopoLessonWidget) setupTeachingMapWidget() {
	w.teachMapWidget = qt.NewQWidget(nil)
	w.teachMapWidget.SetFixedSize2(450, 280)
	w.teachMapWidget.SetStyleSheet("border: 2px solid #ddd; background-color: #f8f8f8; border-radius: 8px;")

	w.teachMapLabel = qt.NewQLabel(w.teachMapWidget)
	w.teachMapLabel.SetGeometry(10, 10, 430, 260)
	w.teachMapLabel.SetAlignment(qt.AlignCenter)
	w.teachMapLabel.SetText("Practice Map\n\nLoad a base map from the 'Map Editor' tab to begin practicing.\nThe map will appear here with highlighted places to identify.")
	w.teachMapLabel.SetStyleSheet("color: #666; font-size: 13px; padding: 20px; background: rgba(255,255,255,0.9); border-radius: 6px;")
	w.teachMapLabel.SetWordWrap(true)
	w.teachMapLabel.SetScaledContents(true)
}

// updateMapDisplay updates the map with current places
func (w *TopoLessonWidget) updateMapDisplay() {
	if w.lesson == nil || w.mapOverlay == nil {
		return
	}

	// Clear existing markers
	for _, marker := range w.mapMarkers {
		marker.Close()
	}
	w.mapMarkers = make([]*qt.QPushButton, 0)

	// Add markers for each place
	for i, item := range w.lesson.Data.List.Items {
		if x, y, hasCoords := item.GetTopoCoordinates(); hasCoords {
			// Scale coordinates based on actual map dimensions
			scaledX := x
			scaledY := y

			if w.currentMap != nil && w.mapPixmap != nil {
				// Use actual map dimensions for scaling
				originalWidth := w.mapPixmap.Width()
				originalHeight := w.mapPixmap.Height()

				if originalWidth > 0 && originalHeight > 0 {
					scaledX = (x * 600) / originalWidth
					scaledY = (y * 300) / originalHeight
				}
			}

			// Ensure coordinates are within bounds
			if scaledX < 10 {
				scaledX = 10
			}
			if scaledX > 590 {
				scaledX = 590
			}
			if scaledY < 10 {
				scaledY = 10
			}
			if scaledY > 290 {
				scaledY = 290
			}

			marker := qt.NewQPushButton(w.mapOverlay)
			marker.SetGeometry(scaledX-10, scaledY-10, 20, 20)
			marker.SetText(fmt.Sprintf("%d", i+1))
			marker.SetStyleSheet("QPushButton { background-color: #ff6b6b; color: white; border: 2px solid #ee5a5a; border-radius: 10px; font-weight: bold; font-size: 10px; } QPushButton:hover { background-color: #ff5252; }")
			marker.SetToolTip(fmt.Sprintf("%s\nCoordinates: (%d, %d)", item.Name, x, y))

			// Connect marker click to show place details
			itemIndex := i
			marker.OnClicked(func() {
				w.showPlaceDetails(itemIndex)
			})

			marker.SetVisible(true)
			w.mapMarkers = append(w.mapMarkers, marker)
		}
	}
}

// showPlaceDetails shows details for a clicked place
func (w *TopoLessonWidget) showPlaceDetails(itemIndex int) {
	if w.lesson == nil || itemIndex >= len(w.lesson.Data.List.Items) {
		return
	}

	item := w.lesson.Data.List.Items[itemIndex]
	x, y, hasCoords := item.GetTopoCoordinates()

	var message string
	if hasCoords {
		message = fmt.Sprintf("Place: %s\nCoordinates: (%d, %d)", item.Name, x, y)
	} else {
		message = fmt.Sprintf("Place: %s\nNo coordinates available", item.Name)
	}

	msgBox := qt.NewQMessageBox(w.QWidget)
	msgBox.SetWindowTitle("Place Details")
	msgBox.SetText(message)
	msgBox.SetIcon(qt.QMessageBox__Information)
	msgBox.SetStandardButtons(qt.QMessageBox__Ok)
	msgBox.Exec()
}

// handleLoadMap loads the selected base map
func (w *TopoLessonWidget) handleLoadMap() {
	currentIndex := w.mapComboBox.CurrentIndex()
	if currentIndex == 0 {
		return // "No map selected"
	}

	variant := w.mapComboBox.ItemData(currentIndex)
	if variant.IsNull() {
		return
	}

	mapID := variant.ToString()
	baseMap, err := w.mapManager.GetMap(mapID)
	if err != nil {
		log.Printf("Failed to load map %s: %v", mapID, err)
		return
	}

	// Load the map image (handle both embedded and external maps)
	w.mapPixmap = qt.NewQPixmap()
	if baseMap.IsEmbedded {
		// Load from embedded data
		imageData, err := w.mapManager.GetEmbeddedMapData(mapID)
		if err != nil {
			log.Printf("Failed to get embedded map data for %s: %v", mapID, err)
			return
		}
		if !w.mapPixmap.LoadFromData2(&imageData[0], uint(len(imageData)), "") {
			log.Printf("Failed to load embedded map image for: %s", mapID)
			return
		}
	} else {
		// Load from file system
		if !w.mapPixmap.Load(baseMap.ImagePath) {
			log.Printf("Failed to load external map image: %s", baseMap.ImagePath)
			return
		}
	}

	w.currentMap = baseMap

	// Update map display
	w.mapLabel.SetPixmap(w.mapPixmap)
	w.mapLabel.SetScaledContents(false)

	// Scale image to fit fixed widget size
	scaledPixmap := w.mapPixmap.Scaled2(600, 300, qt.KeepAspectRatio)
	w.mapLabel.SetPixmap(scaledPixmap)
	w.mapLabel.SetScaledContents(true)

	// Also update teaching map with improved sizing
	if w.teachMapLabel != nil {
		teachScaledPixmap := w.mapPixmap.Scaled2(430, 260, qt.KeepAspectRatio)
		w.teachMapLabel.SetPixmap(teachScaledPixmap)
		w.teachMapLabel.SetScaledContents(true)
		w.teachMapLabel.SetStyleSheet("border-radius: 6px;") // Remove text styling since we now have a map
	}

	w.updateMapDisplay()

	// Provide user feedback
	mapType := "external"
	if baseMap.IsEmbedded {
		mapType = "embedded"
	}
	log.Printf("Successfully loaded %s map: %s (%dx%d) with %d places",
		mapType, baseMap.Name, w.mapPixmap.Width(), w.mapPixmap.Height(), len(baseMap.Places))

	// Update map selector to show loaded status
	w.loadMapButton.SetText("Loaded ✓")
	w.loadMapButton.SetStyleSheet("background-color: #4CAF50; color: white; padding: 6px 12px; font-weight: bold;")

	// Button feedback - simplified without timer
	w.loadMapButton.SetText("Load")
	w.loadMapButton.SetStyleSheet("background-color: #2196F3; color: white; padding: 6px 12px; font-weight: bold;")
}

// handleAddPlace adds a new place with enhanced input dialog
func (w *TopoLessonWidget) handleAddPlace() {
	if w.lesson == nil {
		return
	}

	if w.currentMap == nil {
		// Show message to select a base map first
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("No Base Map")
		msgBox.SetText("Please select and load a base map first before adding places.")
		msgBox.SetIcon(qt.QMessageBox__Information)
		msgBox.SetStandardButtons(qt.QMessageBox__Ok)
		msgBox.Exec()
		return
	}

	// Start adding mode - user can click on map or use dialog
	w.addingPlace = true

	msgBox := qt.NewQMessageBox(w.QWidget)
	msgBox.SetWindowTitle("Add Place")
	msgBox.SetText("Click on the map to add a place at that location, or use the manual entry dialog below.")
	msgBox.SetInformativeText("Map-based adding is recommended for accuracy.")
	msgBox.SetIcon(qt.QMessageBox__Information)

	// Add buttons
	msgBox.SetStandardButtons(qt.QMessageBox__Yes | qt.QMessageBox__No | qt.QMessageBox__Cancel)

	// Get button references and set text
	yesBtn := msgBox.Button(qt.QMessageBox__Yes)
	noBtn := msgBox.Button(qt.QMessageBox__No)
	cancelBtn := msgBox.Button(qt.QMessageBox__Cancel)

	if yesBtn != nil {
		yesBtn.SetText("Click on Map")
	}
	if noBtn != nil {
		noBtn.SetText("Manual Entry Dialog")
	}
	if cancelBtn != nil {
		cancelBtn.SetText("Cancel")
	}

	result := msgBox.Exec()

	if result == int(qt.QMessageBox__No) {
		w.showManualAddDialog()
	}

	// Note: Click-to-add is handled by mouse events on the map
}

// showManualAddDialog shows the manual coordinate entry dialog
func (w *TopoLessonWidget) showManualAddDialog() {
	dialog := qt.NewQDialog(w.QWidget)
	dialog.SetWindowTitle("Add New Place - Manual Entry")
	dialog.SetModal(true)
	dialog.Resize(450, 350)

	layout := qt.NewQVBoxLayout(dialog.QWidget)
	dialog.SetLayout(layout.QLayout)

	// Instructions
	instructLabel := qt.NewQLabel(dialog.QWidget)
	instructLabel.SetText("Enter place details manually. Click-to-add on map is recommended for accuracy.")
	instructLabel.SetStyleSheet("color: #666; margin-bottom: 10px;")
	instructLabel.SetWordWrap(true)
	layout.AddWidget(instructLabel.QWidget)

	// Name input
	nameLabel := qt.NewQLabel(dialog.QWidget)
	nameLabel.SetText("Place Name:")
	layout.AddWidget(nameLabel.QWidget)
	nameInput := qt.NewQLineEdit(dialog.QWidget)
	nameInput.SetPlaceholderText("Enter place name")
	layout.AddWidget(nameInput.QWidget)

	// Coordinate inputs
	coordFrame := qt.NewQFrame(dialog.QWidget)
	coordFrame.SetFrameStyle(int(qt.QFrame__Box))
	coordFrame.SetStyleSheet("padding: 10px; margin: 5px;")
	coordLayout := qt.NewQGridLayout(coordFrame.QWidget)
	coordFrame.SetLayout(coordLayout.QLayout)

	// X coordinate
	xLabel := qt.NewQLabel(coordFrame.QWidget)
	xLabel.SetText("X Coordinate:")
	coordLayout.AddWidget2(xLabel.QWidget, 0, 0)
	xInput := qt.NewQSpinBox(coordFrame.QWidget)
	xInput.SetMinimum(0)
	xInput.SetMaximum(9999)
	xInput.SetValue(100)
	coordLayout.AddWidget2(xInput.QWidget, 0, 1)

	// Y coordinate
	yLabel := qt.NewQLabel(coordFrame.QWidget)
	yLabel.SetText("Y Coordinate:")
	coordLayout.AddWidget2(yLabel.QWidget, 1, 0)
	yInput := qt.NewQSpinBox(coordFrame.QWidget)
	yInput.SetMinimum(0)
	yInput.SetMaximum(9999)
	yInput.SetValue(100)
	coordLayout.AddWidget2(yInput.QWidget, 1, 1)

	layout.AddWidget(coordFrame.QWidget)

	// Plus Code input
	plusFrame := qt.NewQFrame(dialog.QWidget)
	plusFrame.SetFrameStyle(int(qt.QFrame__Box))
	plusFrame.SetStyleSheet("background-color: #f8f9fa; padding: 10px; margin: 5px;")
	plusLayout := qt.NewQVBoxLayout(plusFrame.QWidget)
	plusFrame.SetLayout(plusLayout.QLayout)

	plusLabel := qt.NewQLabel(plusFrame.QWidget)
	plusLabel.SetText("Or enter Plus Code:")
	plusLabel.SetStyleSheet("font-weight: bold;")
	plusLayout.AddWidget(plusLabel.QWidget)

	plusInput := qt.NewQLineEdit(plusFrame.QWidget)
	plusInput.SetPlaceholderText("e.g. 52.37+4.90 (will convert to coordinates)")
	plusLayout.AddWidget(plusInput.QWidget)

	convertButton := qt.NewQPushButton(plusFrame.QWidget)
	convertButton.SetText("Convert Plus Code to Coordinates")
	plusLayout.AddWidget(convertButton.QWidget)

	layout.AddWidget(plusFrame.QWidget)

	// Button layout
	buttonLayout := qt.NewQHBoxLayout(dialog.QWidget)

	okButton := qt.NewQPushButton(dialog.QWidget)
	okButton.SetText("Add Place")
	okButton.SetStyleSheet("background-color: #4CAF50; color: white; padding: 8px;")
	cancelButton := qt.NewQPushButton(dialog.QWidget)
	cancelButton.SetText("Cancel")

	buttonLayout.AddWidget(okButton.QWidget)
	buttonLayout.AddWidget(cancelButton.QWidget)
	layout.AddLayout(buttonLayout.QLayout)

	// Connect signals
	convertButton.OnClicked(func() {
		if w.currentMap != nil && plusInput.Text() != "" {
			if x, y, err := maps.PlusCodeToCoordinate(plusInput.Text(), w.currentMap); err == nil {
				xInput.SetValue(x)
				yInput.SetValue(y)
			} else {
				msgBox := qt.NewQMessageBox(dialog.QWidget)
				msgBox.SetWindowTitle("Plus Code Error")
				msgBox.SetText(fmt.Sprintf("Failed to convert Plus Code: %v", err))
				msgBox.SetIcon(qt.QMessageBox__Warning)
				msgBox.Exec()
			}
		}
	})

	okButton.OnClicked(func() {
		name := nameInput.Text()
		if name != "" {
			x := xInput.Value()
			y := yInput.Value()

			w.lesson.Data.List.AddTopoItem(name, x, y, []string{name}, []string{name})
			w.lesson.Data.Changed = true

			w.updateData()
			w.updateMapDisplay()

			dialog.Accept()
		}
	})

	cancelButton.OnClicked(func() {
		dialog.Reject()
	})

	dialog.Exec()
}

// eventFilter handles mouse events for map interactions
func (w *TopoLessonWidget) eventFilter(watched *qt.QObject, event *qt.QEvent) bool {
	if event.Type() == qt.QEvent__MouseButtonPress {
		// Simple click handling without detailed mouse event parsing
		// For now, we'll handle clicks through widget click events instead
		return false
	}

	return false
}

// handleMapClick processes clicks on the editing map for adding places
func (w *TopoLessonWidget) handleMapClick(x, y int) {
	if w.lesson == nil || w.currentMap == nil || !w.addingPlace {
		return
	}

	// Get place name from user
	name := qt.QInputDialog_GetText(w.QWidget, "Place Name",
		fmt.Sprintf("Enter name for place at coordinates (%d, %d):", x, y))

	if name != "" {
		// Add the place
		w.lesson.Data.List.AddTopoItem(name, x, y, []string{name}, []string{name})
		w.lesson.Data.Changed = true

		// Update display
		w.updateData()
		w.updateMapDisplay()

		// Show confirmation with Plus Code
		plusCode := maps.CoordinateToPlusCode(x, y, w.currentMap)

		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("Place Added")
		msgBox.SetText(fmt.Sprintf("Added '%s' at coordinates (%d, %d)", name, x, y))
		msgBox.SetInformativeText(fmt.Sprintf("Plus Code: %s", plusCode))
		msgBox.SetIcon(qt.QMessageBox__Information)
		msgBox.SetStandardButtons(qt.QMessageBox__Ok)
		msgBox.Exec()
	}

	w.addingPlace = false
}

// handleTeachMapClick processes clicks on the teaching map during practice
func (w *TopoLessonWidget) handleTeachMapClick(x, y int) {
	if w.lesson == nil || w.currentMap == nil || w.currentIndex >= len(w.lesson.Data.List.Items) {
		return
	}

	// Scale coordinates from teaching map (400x250) to full map size
	if w.mapPixmap != nil {
		fullX := (x * w.mapPixmap.Width()) / 400
		fullY := (y * w.mapPixmap.Height()) / 250

		// Find the nearest place within a reasonable threshold
		nearestPlace, err := w.mapManager.FindNearestPlace(w.currentMap.ID, fullX, fullY, 50)
		if err != nil {
			// No place found near click
			w.submitWrongAnswer("No place found at clicked location")
			return
		}

		// Check if this matches the current question
		currentItem := w.lesson.Data.List.Items[w.currentIndex]
		placeName := ""
		if len(currentItem.Questions) > 0 {
			placeName = currentItem.Questions[0]
		}

		// Check if clicked place matches the question
		found := false
		for _, name := range nearestPlace.Names {
			if strings.EqualFold(name, placeName) || strings.EqualFold(name, currentItem.Name) {
				found = true
				break
			}
		}

		if found {
			w.submitCorrectAnswer(placeName)
		} else {
			w.submitWrongAnswer(fmt.Sprintf("Clicked on '%s', but looking for '%s'",
				nearestPlace.Names[0], placeName))
		}
	}
}

// submitCorrectAnswer handles a correct answer submission
func (w *TopoLessonWidget) submitCorrectAnswer(answer string) {
	w.score++
	w.totalAnswers++

	// Add to results table
	w.addResultToTable(answer, true)

	// Update progress
	w.updateProgress()

	// Show feedback
	w.questionLabel.SetText(fmt.Sprintf("✅ Correct! The answer is '%s'", answer))
	w.questionLabel.SetStyleSheet("font-size: 18px; font-weight: bold; margin: 20px; color: #4CAF50;")

	w.submitButton.SetEnabled(false)
	w.nextButton.SetText("Next Place")
}

// submitWrongAnswer handles a wrong answer submission
func (w *TopoLessonWidget) submitWrongAnswer(feedback string) {
	w.totalAnswers++

	currentItem := w.lesson.Data.List.Items[w.currentIndex]
	correctAnswer := ""
	if len(currentItem.Questions) > 0 {
		correctAnswer = currentItem.Questions[0]
	}

	// Add to results table
	w.addResultToTable(correctAnswer, false)

	// Update progress
	w.updateProgress()

	// Show feedback
	w.questionLabel.SetText(fmt.Sprintf("❌ Wrong! %s\nThe correct answer is '%s'",
		feedback, correctAnswer))
	w.questionLabel.SetStyleSheet("font-size: 18px; font-weight: bold; margin: 20px; color: #f44336;")

	w.submitButton.SetEnabled(false)
	w.nextButton.SetText("Next Place")
}

// handleRemovePlace removes selected place
func (w *TopoLessonWidget) handleRemovePlace() {
	if w.lesson == nil {
		return
	}

	currentRow := w.placesList.CurrentRow()
	if currentRow < 0 || currentRow >= len(w.lesson.Data.List.Items) {
		return
	}

	// Confirm deletion
	msgBox := qt.NewQMessageBox(w.QWidget)
	msgBox.SetWindowTitle("Remove Place")
	msgBox.SetText(fmt.Sprintf("Are you sure you want to remove '%s'?", w.lesson.Data.List.Items[currentRow].Questions[0]))
	msgBox.SetIcon(qt.QMessageBox__Question)
	msgBox.SetStandardButtons(qt.QMessageBox__Yes | qt.QMessageBox__No)
	reply := msgBox.Exec()

	if reply == int(qt.QMessageBox__Yes) {
		// Remove item from slice
		items := w.lesson.Data.List.Items
		items = append(items[:currentRow], items[currentRow+1:]...)
		w.lesson.Data.List.Items = items

		// Update IDs for remaining items
		for i := range w.lesson.Data.List.Items {
			w.lesson.Data.List.Items[i].ID = i
		}

		w.lesson.Data.Changed = true
		w.updateData()
		w.updateMapDisplay()
	}
}

// handleClearPlaces clears all places
func (w *TopoLessonWidget) handleClearPlaces() {
	if w.lesson == nil {
		return
	}

	if len(w.lesson.Data.List.Items) == 0 {
		return
	}

	// Confirm clearing
	msgBox := qt.NewQMessageBox(w.QWidget)
	msgBox.SetWindowTitle("Clear All Places")
	msgBox.SetText(fmt.Sprintf("Are you sure you want to remove all %d places?", len(w.lesson.Data.List.Items)))
	msgBox.SetIcon(qt.QMessageBox__Question)
	msgBox.SetStandardButtons(qt.QMessageBox__Yes | qt.QMessageBox__No)
	reply := msgBox.Exec()

	if reply == int(qt.QMessageBox__Yes) {
		w.lesson.Data.List.Items = make([]lesson.WordItem, 0)
		w.lesson.Data.Changed = true
		w.updateData()
		w.updateMapDisplay()
	}
}

// handleImportPlaces imports places from another lesson file
func (w *TopoLessonWidget) handleImportPlaces() {
	if w.lesson == nil {
		return
	}

	// Create file dialog to select file to import from
	fileDialog := qt.NewQFileDialog(w.QWidget)
	fileDialog.SetWindowTitle("Import Places From")
	fileDialog.SetNameFilter("OpenTeaching Files (*.ottp *.kgm);;All Files (*.*)")
	fileDialog.SetAcceptMode(qt.QFileDialog__AcceptOpen)
	fileDialog.SetFileMode(qt.QFileDialog__ExistingFile)

	if fileDialog.Exec() == int(qt.QDialog__Accepted) {
		selectedFiles := fileDialog.SelectedFiles()
		if len(selectedFiles) > 0 {
			filePath := selectedFiles[0]

			// Load the file using the file loader
			loader := lesson.NewFileLoader()
			importedLesson, err := loader.LoadFile(filePath)
			if err != nil {
				msgBox := qt.NewQMessageBox(w.QWidget)
				msgBox.SetWindowTitle("Import Error")
				msgBox.SetText(fmt.Sprintf("Failed to load file: %v", err))
				msgBox.SetIcon(qt.QMessageBox__Warning)
				msgBox.SetStandardButtons(qt.QMessageBox__Ok)
				msgBox.Exec()
				return
			}

			// Count how many topo items we're importing
			importCount := 0
			for _, item := range importedLesson.List.Items {
				if item.IsTopoItem() || (item.Name != "" && len(item.Questions) > 0) {
					importCount++
				}
			}

			if importCount == 0 {
				msgBox := qt.NewQMessageBox(w.QWidget)
				msgBox.SetWindowTitle("Import Result")
				msgBox.SetText("No topography items found in the selected file.")
				msgBox.SetIcon(qt.QMessageBox__Information)
				msgBox.SetStandardButtons(qt.QMessageBox__Ok)
				msgBox.Exec()
				return
			}

			// Confirm import
			msgBox := qt.NewQMessageBox(w.QWidget)
			msgBox.SetWindowTitle("Import Places")
			msgBox.SetText(fmt.Sprintf("Found %d topography items to import. Continue?", importCount))
			msgBox.SetIcon(qt.QMessageBox__Question)
			msgBox.SetStandardButtons(qt.QMessageBox__Yes | qt.QMessageBox__No)
			reply := msgBox.Exec()

			if reply == int(qt.QMessageBox__Yes) {
				imported := 0
				startID := len(w.lesson.Data.List.Items)

				for _, item := range importedLesson.List.Items {
					if item.IsTopoItem() || (item.Name != "" && len(item.Questions) > 0) {
						// Create new item with proper ID
						newItem := lesson.WordItem{
							ID:        startID + imported,
							Name:      item.Name,
							Questions: item.Questions,
							Answers:   item.Answers,
							X:         item.X,
							Y:         item.Y,
						}

						// If no coordinates, set defaults
						if newItem.X == nil && newItem.Y == nil {
							defaultX := 100 + imported*10
							defaultY := 100 + imported*10
							newItem.X = &defaultX
							newItem.Y = &defaultY
						}

						w.lesson.Data.List.Items = append(w.lesson.Data.List.Items, newItem)
						imported++
					}
				}

				w.lesson.Data.Changed = true
				w.updateData()
				w.updateMapDisplay()

				msgBox := qt.NewQMessageBox(w.QWidget)
				msgBox.SetWindowTitle("Import Complete")
				msgBox.SetText(fmt.Sprintf("Successfully imported %d places.", imported))
				msgBox.SetIcon(qt.QMessageBox__Information)
				msgBox.SetStandardButtons(qt.QMessageBox__Ok)
				msgBox.Exec()
			}
		}
	}
}

// handleSave saves the current lesson to file
func (w *TopoLessonWidget) handleSave() {
	if w.lesson == nil {
		return
	}

	var filePath string

	// If lesson has no path or is unsaved (starts with *), show Save As dialog
	if w.lesson.Path == "" || strings.HasPrefix(w.lesson.Path, "*") {
		fileDialog := qt.NewQFileDialog(w.QWidget)
		fileDialog.SetWindowTitle("Save Topography Lesson")
		fileDialog.SetNameFilter("OpenTeaching Topography (*.ottp);;All Files (*.*)")
		fileDialog.SetAcceptMode(qt.QFileDialog__AcceptSave)
		fileDialog.SetDefaultSuffix("ottp")

		if fileDialog.Exec() == int(qt.QDialog__Accepted) {
			selectedFiles := fileDialog.SelectedFiles()
			if len(selectedFiles) > 0 {
				filePath = selectedFiles[0]
				w.lesson.Path = filePath
			} else {
				return
			}
		} else {
			return
		}
	} else {
		filePath = w.lesson.Path
	}

	// Save using FileSaver
	saver := lesson.NewFileSaver()
	err := saver.SaveFile(&w.lesson.Data, filePath)
	if err != nil {
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("Save Error")
		msgBox.SetText(fmt.Sprintf("Failed to save file: %v", err))
		msgBox.SetIcon(qt.QMessageBox__Critical)
		msgBox.SetStandardButtons(qt.QMessageBox__Ok)
		msgBox.Exec()
		return
	}

	// Mark as saved
	w.lesson.Data.Changed = false
	msgBox := qt.NewQMessageBox(w.QWidget)
	msgBox.SetWindowTitle("Save Complete")
	msgBox.SetText(fmt.Sprintf("Successfully saved topography lesson to %s", filepath.Base(filePath)))
	msgBox.SetIcon(qt.QMessageBox__Information)
	msgBox.SetStandardButtons(qt.QMessageBox__Ok)
	msgBox.Exec()
}

// GetLesson returns the lesson associated with this widget
func (w *TopoLessonWidget) GetLesson() *lesson.Lesson {
	return w.lesson
}

// handleDownloadTiles downloads and caches tiles for the selected region
func (w *TopoLessonWidget) handleDownloadTiles() {
	currentIndex := w.tileMapComboBox.CurrentIndex()
	if currentIndex == 0 {
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("No Tile Map Selected")
		msgBox.SetText("Please select a tile map source first.")
		msgBox.SetIcon(qt.QMessageBox__Warning)
		msgBox.Exec()
		return
	}

	variant := w.tileMapComboBox.ItemData(currentIndex)
	if variant.IsNull() {
		return
	}

	tileMapID := variant.ToString()
	north := w.northSpinBox.Value()
	south := w.southSpinBox.Value()
	east := w.eastSpinBox.Value()
	west := w.westSpinBox.Value()
	zoom := w.zoomSpinBox.Value()

	// Validate bounds
	if north <= south {
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("Invalid Bounds")
		msgBox.SetText("North latitude must be greater than south latitude.")
		msgBox.SetIcon(qt.QMessageBox__Warning)
		msgBox.Exec()
		return
	}

	if east <= west {
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("Invalid Bounds")
		msgBox.SetText("East longitude must be greater than west longitude.")
		msgBox.SetIcon(qt.QMessageBox__Warning)
		msgBox.Exec()
		return
	}

	// Show progress dialog
	w.downloadTilesButton.SetText("Downloading...")
	w.downloadTilesButton.SetEnabled(false)

	go func() {
		err := w.mapManager.DownloadTilesForRegion(tileMapID, north, south, east, west, zoom)

		// Update UI - simplified without timer
		w.downloadTilesButton.SetText("Download & Cache Tiles")
		w.downloadTilesButton.SetEnabled(true)

		if err != nil {
			log.Printf("Failed to download tiles: %v", err)
		} else {
			log.Printf("Tiles downloaded and cached successfully")
		}
	}()
}

// handleCreateTileMap creates a new map from tiles
func (w *TopoLessonWidget) handleCreateTileMap() {
	currentIndex := w.tileMapComboBox.CurrentIndex()
	if currentIndex == 0 {
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("No Tile Map Selected")
		msgBox.SetText("Please select a tile map source first.")
		msgBox.SetIcon(qt.QMessageBox__Warning)
		msgBox.Exec()
		return
	}

	variant := w.tileMapComboBox.ItemData(currentIndex)
	if variant.IsNull() {
		return
	}

	tileMapID := variant.ToString()
	north := w.northSpinBox.Value()
	south := w.southSpinBox.Value()
	east := w.eastSpinBox.Value()
	west := w.westSpinBox.Value()
	zoom := w.zoomSpinBox.Value()

	// Validate bounds
	if north <= south || east <= west {
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("Invalid Bounds")
		msgBox.SetText("Please check that north > south and east > west.")
		msgBox.SetIcon(qt.QMessageBox__Warning)
		msgBox.Exec()
		return
	}

	// Create the tile-based map
	baseMap, err := w.mapManager.CreateTileBasedMap(tileMapID, north, south, east, west, zoom)
	if err != nil {
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("Creation Error")
		msgBox.SetText(fmt.Sprintf("Failed to create tile map: %v", err))
		msgBox.SetIcon(qt.QMessageBox__Critical)
		msgBox.Exec()
		return
	}

	// Add to regular map selector
	w.mapComboBox.AddItem(baseMap.Name)
	w.mapComboBox.SetItemData(w.mapComboBox.Count()-1, qt.NewQVariant17(baseMap.ID))

	// Show success message
	msgBox := qt.NewQMessageBox(w.QWidget)
	msgBox.SetWindowTitle("Map Created")
	msgBox.SetText(fmt.Sprintf("Tile-based map '%s' has been created and added to your map selection.", baseMap.Name))
	msgBox.SetIcon(qt.QMessageBox__Information)
	msgBox.Exec()

	log.Printf("Created tile-based map: %s", baseMap.Name)
}

// SetLesson sets a new lesson for this widget
func (w *TopoLessonWidget) SetLesson(lesson *lesson.Lesson) {
	w.lesson = lesson
	w.updateData()
}
