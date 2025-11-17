// Package media provides media lesson widget functionality
//
// MediaLessonWidget handles the display and interaction for media lessons
package media

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/mappu/miqt/qt"
)

// MediaLessonWidget handles media lesson display and interaction
type MediaLessonWidget struct {
	*qt.QWidget
	lesson *lesson.Lesson

	// Main layout
	mainLayout *qt.QVBoxLayout

	// Tab widget for different views
	tabWidget *qt.QTabWidget

	// Tabs
	enterTab   *qt.QWidget
	teachTab   *qt.QWidget
	resultsTab *qt.QWidget

	// Enter tab components
	enterLayout  *qt.QVBoxLayout
	mediaList    *qt.QListWidget
	addButton    *qt.QPushButton
	removeButton *qt.QPushButton
	clearButton  *qt.QPushButton
	importButton *qt.QPushButton
	saveButton   *qt.QPushButton

	// Teach tab components
	teachLayout   *qt.QVBoxLayout
	mediaDisplay  *qt.QLabel
	mediaPreview  *qt.QWidget
	mediaImage    *qt.QLabel
	questionLabel *qt.QLabel
	answerInput   *qt.QLineEdit
	submitButton  *qt.QPushButton
	nextButton    *qt.QPushButton
	scoreLabel    *qt.QLabel
	progressLabel *qt.QLabel

	playButton        *qt.QPushButton
	openBrowserButton *qt.QPushButton
	mediaTypeLabel    *qt.QLabel

	// Results tab components
	resultsLayout *qt.QVBoxLayout
	resultsTable  *qt.QTableWidget
	statsLabel    *qt.QLabel

	// Teaching state
	currentIndex int
	score        int
	totalAnswers int
}

// NewMediaLessonWidget creates a new media lesson widget
func NewMediaLessonWidget(lesson *lesson.Lesson, parent *qt.QWidget) *MediaLessonWidget {
	widget := &MediaLessonWidget{
		QWidget: qt.NewQWidget(parent),
		lesson:  lesson,
	}

	widget.setupUI()
	widget.updateData()
	widget.connectSignals()

	return widget
}

// setupUI initializes the user interface
func (w *MediaLessonWidget) setupUI() {
	// Main layout
	w.mainLayout = qt.NewQVBoxLayout(w.QWidget)
	w.SetLayout(w.mainLayout.QLayout)

	// Create tab widget
	w.tabWidget = qt.NewQTabWidget(w.QWidget)
	w.mainLayout.AddWidget(w.tabWidget.QWidget)

	// Setup tabs
	w.setupEnterTab()
	w.setupTeachTab()
	w.setupResultsTab()
}

// setupEnterTab creates the enter/edit tab
func (w *MediaLessonWidget) setupEnterTab() {
	w.enterTab = qt.NewQWidget(w.QWidget)
	w.enterLayout = qt.NewQVBoxLayout(w.enterTab)
	w.enterTab.SetLayout(w.enterLayout.QLayout)

	// Title
	titleLabel := qt.NewQLabel(w.enterTab)
	titleLabel.SetText("Media Items to Learn")
	titleLabel.SetStyleSheet("font-size: 16px; font-weight: bold; margin: 10px;")
	w.enterLayout.AddWidget(titleLabel.QWidget)

	// Media list
	w.mediaList = qt.NewQListWidget(w.enterTab)
	w.mediaList.SetMinimumHeight(300)
	w.enterLayout.AddWidget(w.mediaList.QWidget)

	// Button layout
	buttonLayout := qt.NewQHBoxLayout2()

	w.addButton = qt.NewQPushButton(w.enterTab)
	w.addButton.SetText("Add Media")
	buttonLayout.AddWidget(w.addButton.QWidget)

	w.removeButton = qt.NewQPushButton(w.enterTab)
	w.removeButton.SetText("Remove Item")
	buttonLayout.AddWidget(w.removeButton.QWidget)

	w.clearButton = qt.NewQPushButton(w.enterTab)
	w.clearButton.SetText("Clear All")
	buttonLayout.AddWidget(w.clearButton.QWidget)

	w.importButton = qt.NewQPushButton(w.enterTab)
	w.importButton.SetText("Import...")
	buttonLayout.AddWidget(w.importButton.QWidget)

	w.saveButton = qt.NewQPushButton(w.enterTab)
	w.saveButton.SetText("Save")
	buttonLayout.AddWidget(w.saveButton.QWidget)

	buttonLayout.AddStretch()
	w.enterLayout.AddLayout(buttonLayout.QLayout)

	// Add tab
	w.tabWidget.AddTab(w.enterTab, "Enter Media")
}

// setupTeachTab creates the teaching tab
func (w *MediaLessonWidget) setupTeachTab() {
	w.teachTab = qt.NewQWidget(w.QWidget)
	w.teachLayout = qt.NewQVBoxLayout(w.teachTab)
	w.teachTab.SetLayout(w.teachLayout.QLayout)

	// Progress info
	w.progressLabel = qt.NewQLabel(w.teachTab)
	w.progressLabel.SetText("Progress: 0/0")
	w.progressLabel.SetStyleSheet("font-size: 14px; margin: 10px;")
	w.teachLayout.AddWidget(w.progressLabel.QWidget)

	// Score info
	w.scoreLabel = qt.NewQLabel(w.teachTab)
	w.scoreLabel.SetText("Score: 0/0 (0%)")
	w.scoreLabel.SetStyleSheet("font-size: 14px; margin: 10px;")
	w.teachLayout.AddWidget(w.scoreLabel.QWidget)

	// Media display area
	mediaFrame := qt.NewQFrame(w.teachTab)
	mediaFrame.SetFrameStyle(int(qt.QFrame__Box))
	mediaFrame.SetMinimumHeight(300)
	mediaLayout := qt.NewQVBoxLayout2()
	mediaFrame.SetLayout(mediaLayout.QLayout)

	// Media type indicator
	w.mediaTypeLabel = qt.NewQLabel(w.teachTab)
	w.mediaTypeLabel.SetText("Media Type: Unknown")
	w.mediaTypeLabel.SetStyleSheet("font-size: 12px; color: gray;")
	mediaLayout.AddWidget(w.mediaTypeLabel.QWidget)

	// Setup media preview widget
	w.setupMediaPreview()
	mediaLayout.AddWidget(w.mediaPreview)

	// Media display info
	w.mediaDisplay = qt.NewQLabel(w.teachTab)
	w.mediaDisplay.SetText("Media information will appear here")
	w.mediaDisplay.SetStyleSheet("font-size: 14px; background-color: #f8f9fa; border: 1px solid #ddd; padding: 10px;")
	w.mediaDisplay.SetAlignment(qt.AlignCenter)
	w.mediaDisplay.SetMinimumHeight(80)
	w.mediaDisplay.SetWordWrap(true)
	mediaLayout.AddWidget(w.mediaDisplay.QWidget)

	// Media control buttons
	buttonLayout := qt.NewQHBoxLayout2()

	w.playButton = qt.NewQPushButton(w.teachTab)
	w.playButton.SetText("▶ Play Media")
	w.playButton.SetStyleSheet("font-size: 12px; padding: 8px 16px; background-color: #4CAF50; color: white; border: none; border-radius: 4px;")
	w.playButton.SetEnabled(false)
	buttonLayout.AddWidget(w.playButton.QWidget)

	w.openBrowserButton = qt.NewQPushButton(w.teachTab)
	w.openBrowserButton.SetText("🌐 Open in Browser")
	w.openBrowserButton.SetStyleSheet("font-size: 12px; padding: 8px 16px; background-color: #2196F3; color: white; border: none; border-radius: 4px;")
	w.openBrowserButton.SetEnabled(false)
	buttonLayout.AddWidget(w.openBrowserButton.QWidget)

	buttonLayout.AddStretch()
	mediaLayout.AddLayout(buttonLayout.QLayout)

	w.teachLayout.AddWidget(mediaFrame.QWidget)

	// Question area
	questionFrame := qt.NewQFrame(w.teachTab)
	questionFrame.SetFrameStyle(int(qt.QFrame__Box))
	questionLayout := qt.NewQVBoxLayout2()
	questionFrame.SetLayout(questionLayout.QLayout)

	instructionLabel := qt.NewQLabel(w.teachTab)
	instructionLabel.SetText("What do you see/hear in this media?")
	instructionLabel.SetStyleSheet("font-size: 12px; color: gray;")
	questionLayout.AddWidget(instructionLabel.QWidget)

	w.questionLabel = qt.NewQLabel(w.teachTab)
	w.questionLabel.SetText("Click 'Next' to start")
	w.questionLabel.SetStyleSheet("font-size: 18px; font-weight: bold; margin: 20px; text-align: center;")
	w.questionLabel.SetAlignment(qt.AlignCenter)
	w.questionLabel.SetWordWrap(true)
	questionLayout.AddWidget(w.questionLabel.QWidget)

	w.teachLayout.AddWidget(questionFrame.QWidget)

	// Answer area
	answerFrame := qt.NewQFrame(w.teachTab)
	answerFrame.SetFrameStyle(int(qt.QFrame__Box))
	answerLayout := qt.NewQVBoxLayout2()
	answerFrame.SetLayout(answerLayout.QLayout)

	answerLabel := qt.NewQLabel(w.teachTab)
	answerLabel.SetText("Your Answer:")
	answerLabel.SetStyleSheet("font-size: 12px; color: gray;")
	answerLayout.AddWidget(answerLabel.QWidget)

	w.answerInput = qt.NewQLineEdit(w.teachTab)
	w.answerInput.SetPlaceholderText("Type what you see/hear...")
	w.answerInput.SetStyleSheet("font-size: 16px; padding: 8px; margin: 5px;")
	answerLayout.AddWidget(w.answerInput.QWidget)

	w.teachLayout.AddWidget(answerFrame.QWidget)

	// Control buttons
	controlLayout := qt.NewQHBoxLayout2()

	w.submitButton = qt.NewQPushButton(w.teachTab)
	w.submitButton.SetText("Submit Answer")
	w.submitButton.SetStyleSheet("font-size: 14px; padding: 8px; background-color: #4CAF50; color: white;")
	w.submitButton.SetEnabled(false)
	controlLayout.AddWidget(w.submitButton.QWidget)

	w.nextButton = qt.NewQPushButton(w.teachTab)
	w.nextButton.SetText("Next Item")
	w.nextButton.SetStyleSheet("font-size: 14px; padding: 8px; background-color: #2196F3; color: white;")
	controlLayout.AddWidget(w.nextButton.QWidget)

	controlLayout.AddStretch()
	w.teachLayout.AddLayout(controlLayout.QLayout)

	w.teachLayout.AddStretch()

	// Add tab
	w.tabWidget.AddTab(w.teachTab, "Practice")
}

// setupResultsTab creates the results tab
func (w *MediaLessonWidget) setupResultsTab() {
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
	w.resultsTable.SetColumnCount(4)

	headers := []string{"Media Item", "Question", "Your Answer", "Result"}
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

// connectSignals connects UI signals to slots
func (w *MediaLessonWidget) connectSignals() {
	// Answer input - enable submit button when text is entered
	w.answerInput.OnTextChanged(func(text string) {
		w.submitButton.SetEnabled(len(text) > 0)
	})

	// Submit button
	w.submitButton.OnClicked(func() {
		w.handleSubmitAnswer()
	})

	// Next button
	w.nextButton.OnClicked(func() {
		w.handleNextQuestion()
	})

	// Play button
	w.playButton.OnClicked(func() {
		w.handlePlayMedia()
	})

	// Enter key in answer input
	w.answerInput.OnReturnPressed(func() {
		if w.submitButton.IsEnabled() {
			w.handleSubmitAnswer()
		}
	})

	// List management buttons
	w.addButton.OnClicked(func() {
		w.handleAddMedia()
	})

	w.removeButton.OnClicked(func() {
		w.handleRemoveMedia()
	})

	w.clearButton.OnClicked(func() {
		w.handleClearMedia()
	})

	w.importButton.OnClicked(func() {
		w.handleImportMedia()
	})

	w.saveButton.OnClicked(func() {
		w.handleSave()
	})

	w.playButton.OnClicked(func() {
		w.handlePlayMedia()
	})

	w.openBrowserButton.OnClicked(func() {
		w.handleOpenInBrowser()
	})
}

// updateData updates the widget with lesson data
func (w *MediaLessonWidget) updateData() {
	if w.lesson == nil {
		return
	}

	// Update media list
	w.mediaList.Clear()
	for _, item := range w.lesson.Data.List.Items {
		var displayText string
		if item.Name != "" {
			displayText = item.Name
		} else if len(item.Questions) > 0 {
			displayText = item.Questions[0]
		} else if len(item.Answers) > 0 {
			displayText = item.Answers[0]
		} else {
			displayText = fmt.Sprintf("Media Item %d", item.ID)
		}

		// Add media file information if available
		if filename, remote, hasMedia := item.GetMediaInfo(); hasMedia {
			if remote {
				displayText = fmt.Sprintf("%s [Remote: %s]", displayText, filename)
			} else {
				displayText = fmt.Sprintf("%s [File: %s]", displayText, filename)
			}
		}

		w.mediaList.AddItem(displayText)
	}

	// Reset teaching state
	w.currentIndex = 0
	w.score = 0
	w.totalAnswers = 0
	w.updateProgress()
	w.updateTeachDisplay()
}

// updateProgress updates the progress display
func (w *MediaLessonWidget) updateProgress() {
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
func (w *MediaLessonWidget) updateTeachDisplay() {
	if w.lesson == nil || len(w.lesson.Data.List.Items) == 0 {
		w.questionLabel.SetText("No media items to practice")
		w.mediaDisplay.SetText("No media available")
		w.mediaTypeLabel.SetText("Media Type: None")
		return
	}

	if w.currentIndex >= len(w.lesson.Data.List.Items) {
		w.questionLabel.SetText("Practice complete!")
		w.mediaDisplay.SetText("All media items completed!")
		w.submitButton.SetEnabled(false)
		w.nextButton.SetText("Start Over")
		return
	}

	item := w.lesson.Data.List.Items[w.currentIndex]

	// Set question
	if len(item.Questions) > 0 {
		w.questionLabel.SetText(fmt.Sprintf("Question: %s", item.Questions[0]))
	} else {
		w.questionLabel.SetText("What do you see/hear in this media?")
	}

	// Set media display with actual file information
	mediaName := "Unknown Media"
	if item.Name != "" {
		mediaName = item.Name
	} else if len(item.Questions) > 0 {
		mediaName = item.Questions[0]
	} else if len(item.Answers) > 0 {
		mediaName = item.Answers[0]
	}

	// Update media preview and display
	w.updateMediaDisplay(item, mediaName)
	w.playButton.SetEnabled(true)
	w.openBrowserButton.SetEnabled(true)

	w.answerInput.Clear()
	w.answerInput.SetFocus()
}

// handleSubmitAnswer processes the submitted answer
func (w *MediaLessonWidget) handleSubmitAnswer() {
	if w.lesson == nil || w.currentIndex >= len(w.lesson.Data.List.Items) {
		return
	}

	userAnswer := w.answerInput.Text()
	item := w.lesson.Data.List.Items[w.currentIndex]

	// Check if answer is correct (case-insensitive)
	correct := false
	expectedAnswer := "No answer provided"
	if len(item.Answers) > 0 {
		expectedAnswer = item.Answers[0]
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
		w.questionLabel.SetText(fmt.Sprintf("✅ Correct! The answer is: %s", expectedAnswer))
		w.questionLabel.SetStyleSheet("font-size: 18px; font-weight: bold; margin: 20px; color: green;")
	} else {
		w.questionLabel.SetText(fmt.Sprintf("❌ Incorrect. The correct answer is: %s", expectedAnswer))
		w.questionLabel.SetStyleSheet("font-size: 18px; font-weight: bold; margin: 20px; color: red;")
	}

	// Add to results table
	questionText := "Media question"
	if len(item.Questions) > 0 {
		questionText = item.Questions[0]
	}
	w.addResultToTable(fmt.Sprintf("Item %d", item.ID), questionText, userAnswer, correct)

	w.updateProgress()
	w.submitButton.SetEnabled(false)
	w.nextButton.SetText("Next Item")
}

// handleNextQuestion moves to the next question
func (w *MediaLessonWidget) handleNextQuestion() {
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
func (w *MediaLessonWidget) addResultToTable(mediaItem, question, userAnswer string, correct bool) {
	row := w.resultsTable.RowCount()
	w.resultsTable.SetRowCount(row + 1)

	mediaItem2 := qt.NewQTableWidgetItem2(mediaItem)
	w.resultsTable.SetItem(row, 0, mediaItem2)

	questionItem := qt.NewQTableWidgetItem2(question)
	w.resultsTable.SetItem(row, 1, questionItem)

	answerItem := qt.NewQTableWidgetItem2(userAnswer)
	w.resultsTable.SetItem(row, 2, answerItem)

	resultText := "❌ Wrong"
	if correct {
		resultText = "✅ Correct"
	}
	resultItem := qt.NewQTableWidgetItem2(resultText)
	w.resultsTable.SetItem(row, 3, resultItem)

	// Update statistics
	w.updateStatistics()
}

// updateStatistics updates the statistics display
func (w *MediaLessonWidget) updateStatistics() {
	if w.totalAnswers == 0 {
		w.statsLabel.SetText("No test results yet")
		return
	}

	percentage := float64(w.score) / float64(w.totalAnswers) * 100
	w.statsLabel.SetText(fmt.Sprintf("Total: %d questions, Correct: %d, Wrong: %d, Score: %.1f%%",
		w.totalAnswers, w.score, w.totalAnswers-w.score, percentage))
}

// handleAddMedia adds new media with input dialog
func (w *MediaLessonWidget) handleAddMedia() {
	if w.lesson == nil {
		return
	}

	// Create input dialog
	dialog := qt.NewQDialog(w.QWidget)
	dialog.SetWindowTitle("Add New Media Item")
	dialog.SetModal(true)
	dialog.Resize(450, 300)

	// Main layout
	layout := qt.NewQVBoxLayout(dialog.QWidget)
	dialog.SetLayout(layout.QLayout)

	// Name input
	nameLabel := qt.NewQLabel(dialog.QWidget)
	nameLabel.SetText("Name:")
	layout.AddWidget(nameLabel.QWidget)
	nameInput := qt.NewQLineEdit(dialog.QWidget)
	nameInput.SetPlaceholderText("Enter media item name")
	layout.AddWidget(nameInput.QWidget)

	// Question input
	questionLabel := qt.NewQLabel(dialog.QWidget)
	questionLabel.SetText("Question:")
	layout.AddWidget(questionLabel.QWidget)
	questionInput := qt.NewQLineEdit(dialog.QWidget)
	questionInput.SetPlaceholderText("Enter question")
	layout.AddWidget(questionInput.QWidget)

	// Answer input
	answerLabel := qt.NewQLabel(dialog.QWidget)
	answerLabel.SetText("Answer:")
	layout.AddWidget(answerLabel.QWidget)
	answerInput := qt.NewQLineEdit(dialog.QWidget)
	answerInput.SetPlaceholderText("Enter answer")
	layout.AddWidget(answerInput.QWidget)

	// Filename input
	filenameLabel := qt.NewQLabel(dialog.QWidget)
	filenameLabel.SetText("Filename:")
	layout.AddWidget(filenameLabel.QWidget)
	filenameInput := qt.NewQLineEdit(dialog.QWidget)
	filenameInput.SetPlaceholderText("Enter filename or URL")
	layout.AddWidget(filenameInput.QWidget)

	// Remote checkbox
	typeLabel := qt.NewQLabel(dialog.QWidget)
	typeLabel.SetText("Type:")
	layout.AddWidget(typeLabel.QWidget)
	remoteCheck := qt.NewQCheckBox(dialog.QWidget)
	remoteCheck.SetText("Remote media (URL)")
	layout.AddWidget(remoteCheck.QWidget)

	// Button layout
	buttonLayout := qt.NewQHBoxLayout(dialog.QWidget)

	okButton := qt.NewQPushButton(dialog.QWidget)
	okButton.SetText("OK")
	cancelButton := qt.NewQPushButton(dialog.QWidget)
	cancelButton.SetText("Cancel")

	buttonLayout.AddWidget(okButton.QWidget)
	buttonLayout.AddWidget(cancelButton.QWidget)

	layout.AddLayout(buttonLayout.QLayout)

	// Connect signals
	okButton.OnClicked(func() {
		name := nameInput.Text()
		question := questionInput.Text()
		answer := answerInput.Text()
		filename := filenameInput.Text()
		remote := remoteCheck.IsChecked()

		if name != "" && question != "" && answer != "" {
			// Add new media item to lesson
			w.lesson.Data.List.AddMediaItem(name, []string{question}, []string{answer}, filename, remote)
			w.lesson.Data.Changed = true

			// Update display
			w.updateData()

			dialog.Accept()
		}
	})

	cancelButton.OnClicked(func() {
		dialog.Reject()
	})

	dialog.Exec()
}

// handleRemoveMedia removes selected media
func (w *MediaLessonWidget) handleRemoveMedia() {
	if w.lesson == nil {
		return
	}

	currentRow := w.mediaList.CurrentRow()
	if currentRow < 0 || currentRow >= len(w.lesson.Data.List.Items) {
		return
	}

	// Get item name for confirmation
	itemName := "Media Item"
	item := w.lesson.Data.List.Items[currentRow]
	if item.Name != "" {
		itemName = item.Name
	} else if len(item.Questions) > 0 {
		itemName = item.Questions[0]
	}

	// Confirm deletion
	msgBox := qt.NewQMessageBox(w.QWidget)
	msgBox.SetWindowTitle("Remove Media Item")
	msgBox.SetText(fmt.Sprintf("Are you sure you want to remove '%s'?", itemName))
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
	}
}

// handleClearMedia clears all media
func (w *MediaLessonWidget) handleClearMedia() {
	if w.lesson == nil {
		return
	}

	if len(w.lesson.Data.List.Items) == 0 {
		return
	}

	// Confirm clearing
	msgBox := qt.NewQMessageBox(w.QWidget)
	msgBox.SetWindowTitle("Clear All Media Items")
	msgBox.SetText(fmt.Sprintf("Are you sure you want to remove all %d media items?", len(w.lesson.Data.List.Items)))
	msgBox.SetIcon(qt.QMessageBox__Question)
	msgBox.SetStandardButtons(qt.QMessageBox__Yes | qt.QMessageBox__No)
	reply := msgBox.Exec()

	if reply == int(qt.QMessageBox__Yes) {
		w.lesson.Data.List.Items = make([]lesson.WordItem, 0)
		w.lesson.Data.Changed = true
		w.updateData()
	}
}

// handleImportMedia imports media items from another lesson file
func (w *MediaLessonWidget) handleImportMedia() {
	if w.lesson == nil {
		return
	}

	// Create file dialog to select file to import from
	fileDialog := qt.NewQFileDialog(w.QWidget)
	fileDialog.SetWindowTitle("Import Media From")
	fileDialog.SetNameFilter("OpenTeaching Files (*.otmd);;All Files (*.*)")
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

			// Count how many media items we're importing
			importCount := 0
			for _, item := range importedLesson.List.Items {
				if item.IsMediaItem() || (item.Name != "" && len(item.Questions) > 0) {
					importCount++
				}
			}

			if importCount == 0 {
				msgBox := qt.NewQMessageBox(w.QWidget)
				msgBox.SetWindowTitle("Import Result")
				msgBox.SetText("No media items found in the selected file.")
				msgBox.SetIcon(qt.QMessageBox__Information)
				msgBox.SetStandardButtons(qt.QMessageBox__Ok)
				msgBox.Exec()
				return
			}

			// Confirm import
			msgBox := qt.NewQMessageBox(w.QWidget)
			msgBox.SetWindowTitle("Import Media Items")
			msgBox.SetText(fmt.Sprintf("Found %d media items to import. Continue?", importCount))
			msgBox.SetIcon(qt.QMessageBox__Question)
			msgBox.SetStandardButtons(qt.QMessageBox__Yes | qt.QMessageBox__No)
			reply := msgBox.Exec()

			if reply == int(qt.QMessageBox__Yes) {
				imported := 0
				startID := len(w.lesson.Data.List.Items)

				for _, item := range importedLesson.List.Items {
					if item.IsMediaItem() || (item.Name != "" && len(item.Questions) > 0) {
						// Create new item with proper ID
						newItem := lesson.WordItem{
							ID:        startID + imported,
							Name:      item.Name,
							Questions: item.Questions,
							Answers:   item.Answers,
							Filename:  item.Filename,
							Remote:    item.Remote,
						}

						w.lesson.Data.List.Items = append(w.lesson.Data.List.Items, newItem)
						imported++
					}
				}

				w.lesson.Data.Changed = true
				w.updateData()

				msgBox := qt.NewQMessageBox(w.QWidget)
				msgBox.SetWindowTitle("Import Complete")
				msgBox.SetText(fmt.Sprintf("Successfully imported %d media items.", imported))
				msgBox.SetIcon(qt.QMessageBox__Information)
				msgBox.SetStandardButtons(qt.QMessageBox__Ok)
				msgBox.Exec()
			}
		}
	}
}

// handleSave saves the current lesson to file
func (w *MediaLessonWidget) handleSave() {
	if w.lesson == nil {
		return
	}

	var filePath string

	// If lesson has no path or is unsaved (starts with *), show Save As dialog
	if w.lesson.Path == "" || strings.HasPrefix(w.lesson.Path, "*") {
		fileDialog := qt.NewQFileDialog(w.QWidget)
		fileDialog.SetWindowTitle("Save Media Lesson")
		fileDialog.SetNameFilter("OpenTeaching Media (*.otmd);;All Files (*.*)")
		fileDialog.SetAcceptMode(qt.QFileDialog__AcceptSave)
		fileDialog.SetDefaultSuffix("otmd")

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
	msgBox.SetText(fmt.Sprintf("Successfully saved media lesson to %s", filepath.Base(filePath)))
	msgBox.SetIcon(qt.QMessageBox__Information)
	msgBox.SetStandardButtons(qt.QMessageBox__Ok)
	msgBox.Exec()
}

// setupMediaPreview creates the media preview widget
func (w *MediaLessonWidget) setupMediaPreview() {
	w.mediaPreview = qt.NewQWidget(w.teachTab)
	w.mediaPreview.SetMinimumSize2(400, 250)
	w.mediaPreview.SetStyleSheet("border: 2px solid #ddd; background-color: #fafafa; border-radius: 8px;")

	layout := qt.NewQVBoxLayout(w.mediaPreview)
	w.mediaPreview.SetLayout(layout.QLayout)

	// Image display for image files
	w.mediaImage = qt.NewQLabel(w.mediaPreview)
	w.mediaImage.SetAlignment(qt.AlignCenter)
	w.mediaImage.SetMinimumSize2(380, 200)
	w.mediaImage.SetStyleSheet("border: 1px solid #ccc; background-color: white; margin: 5px;")
	w.mediaImage.SetScaledContents(true)
	layout.AddWidget(w.mediaImage.QWidget)

	// Default display
	w.mediaImage.SetText("Media Preview\n\n📁 No media loaded")
}

// updateMediaDisplay updates the media display based on the current item
func (w *MediaLessonWidget) updateMediaDisplay(item lesson.WordItem, mediaName string) {
	filename, remote, hasMedia := item.GetMediaInfo()

	if hasMedia {
		if remote {
			// Remote media - show info and enable browser button
			w.mediaDisplay.SetText(fmt.Sprintf("🌐 Remote Media: %s\n\nURL: %s\n\nClick 'Open in Browser' to view", mediaName, filename))
			w.mediaTypeLabel.SetText("Media Type: Remote Content")
			w.mediaImage.SetText("🌐 Remote Media\n\nClick 'Open in Browser'\nto view online content")
		} else {
			// Local file - try to display or show info
			w.displayLocalMedia(filename, mediaName)
		}
	} else {
		// No media file info
		w.mediaDisplay.SetText(fmt.Sprintf("📝 Media Item: %s\n\nNo media file specified", mediaName))
		w.mediaTypeLabel.SetText("Media Type: Text Only")
		w.mediaImage.SetText("📝 Text-only Media\n\nNo file to display")
	}
}

// displayLocalMedia attempts to display local media files
func (w *MediaLessonWidget) displayLocalMedia(filename, mediaName string) {
	if !filepath.IsAbs(filename) {
		// Relative path - might need to resolve
		w.mediaDisplay.SetText(fmt.Sprintf("📁 Local Media: %s\n\nFile: %s\n\n⚠️ File path may need resolution", mediaName, filename))
		w.mediaTypeLabel.SetText("Media Type: Local File (Unresolved)")
		w.mediaImage.SetText("📁 Local Media\n\nFile path needs resolution")
		return
	}

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		w.mediaDisplay.SetText(fmt.Sprintf("❌ Local Media: %s\n\nFile: %s\n\n⚠️ File not found", mediaName, filename))
		w.mediaTypeLabel.SetText("Media Type: Missing File")
		w.mediaImage.SetText("❌ Media File\nNot Found\n\nClick 'Play Media' to try\nopening anyway")
		return
	}

	// Determine file type by extension
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		// Image file - try to display
		w.displayImage(filename, mediaName)
	case ".mp4", ".avi", ".mov", ".mkv":
		// Video file
		w.mediaDisplay.SetText(fmt.Sprintf("🎬 Video: %s\n\nFile: %s\n\nClick 'Play Media' to open in system player", mediaName, filepath.Base(filename)))
		w.mediaTypeLabel.SetText("Media Type: Video File")
		w.mediaImage.SetText("🎬 Video File\n\n" + filepath.Base(filename) + "\n\nClick 'Play Media'\nto open in system player")
	case ".mp3", ".wav", ".flac", ".ogg":
		// Audio file
		w.mediaDisplay.SetText(fmt.Sprintf("🎵 Audio: %s\n\nFile: %s\n\nClick 'Play Media' to open in system player", mediaName, filepath.Base(filename)))
		w.mediaTypeLabel.SetText("Media Type: Audio File")
		w.mediaImage.SetText("🎵 Audio File\n\n" + filepath.Base(filename) + "\n\nClick 'Play Media'\nto open in system player")
	default:
		// Unknown file type
		w.mediaDisplay.SetText(fmt.Sprintf("📄 Media File: %s\n\nFile: %s\n\nClick 'Play Media' to open with system default", mediaName, filepath.Base(filename)))
		w.mediaTypeLabel.SetText("Media Type: " + strings.ToUpper(ext[1:]) + " File")
		w.mediaImage.SetText("📄 Media File\n\n" + filepath.Base(filename) + "\n\nClick 'Play Media'\nto open with system default")
	}
}

// displayImage attempts to display an image file
func (w *MediaLessonWidget) displayImage(filename, mediaName string) {
	pixmap := qt.NewQPixmap()
	if pixmap.Load(filename) {
		// Successfully loaded image
		w.mediaImage.SetPixmap(pixmap)
		w.mediaDisplay.SetText(fmt.Sprintf("🖼️ Image: %s\n\nFile: %s\n\nImage displayed above", mediaName, filepath.Base(filename)))
		w.mediaTypeLabel.SetText("Media Type: Image File")
	} else {
		// Failed to load image
		w.mediaDisplay.SetText(fmt.Sprintf("❌ Image Load Error: %s\n\nFile: %s\n\n⚠️ Could not load image", mediaName, filepath.Base(filename)))
		w.mediaTypeLabel.SetText("Media Type: Invalid Image")
		w.mediaImage.SetText("❌ Image Load Error\n\nCould not display\n" + filepath.Base(filename))
	}
}

// handlePlayMedia handles the play media button click
func (w *MediaLessonWidget) handlePlayMedia() {
	if w.lesson == nil || w.currentIndex >= len(w.lesson.Data.List.Items) {
		return
	}

	item := w.lesson.Data.List.Items[w.currentIndex]
	filename, remote, hasMedia := item.GetMediaInfo()

	if !hasMedia {
		// No media file specified
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("No Media")
		msgBox.SetText("This item does not have a media file specified.")
		msgBox.SetIcon(qt.QMessageBox__Information)
		msgBox.SetStandardButtons(qt.QMessageBox__Ok)
		msgBox.Exec()
		return
	}

	var targetPath string
	if remote {
		targetPath = filename // URL
	} else {
		targetPath = filename // File path
	}

	// Open with system default application
	w.openWithSystem(targetPath)
}

// handleOpenInBrowser handles the open in browser button click
func (w *MediaLessonWidget) handleOpenInBrowser() {
	if w.lesson == nil || w.currentIndex >= len(w.lesson.Data.List.Items) {
		return
	}

	item := w.lesson.Data.List.Items[w.currentIndex]
	filename, remote, hasMedia := item.GetMediaInfo()

	var targetUrl string
	if hasMedia {
		if remote {
			targetUrl = filename // Already a URL
		} else {
			// Local file - convert to file:// URL
			absPath, err := filepath.Abs(filename)
			if err != nil {
				absPath = filename
			}
			targetUrl = "file://" + absPath
		}
	} else {
		// No media file - show a message
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("No Media URL")
		msgBox.SetText("This item does not have a media file or URL to open in browser.")
		msgBox.SetIcon(qt.QMessageBox__Information)
		msgBox.SetStandardButtons(qt.QMessageBox__Ok)
		msgBox.Exec()
		return
	}

	w.openUrlInBrowser(targetUrl)
}

// openWithSystem opens a file or URL with the system's default application
func (w *MediaLessonWidget) openWithSystem(path string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	case "darwin": // macOS
		cmd = exec.Command("open", path)
	default: // Linux and others
		cmd = exec.Command("xdg-open", path)
	}

	err := cmd.Start()
	if err != nil {
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("Open Error")
		msgBox.SetText(fmt.Sprintf("Could not open media with system default:\n\n%v\n\nPath: %s", err, path))
		msgBox.SetIcon(qt.QMessageBox__Warning)
		msgBox.SetStandardButtons(qt.QMessageBox__Ok)
		msgBox.Exec()
	}
}

// openUrlInBrowser opens a URL in the system's default browser
func (w *MediaLessonWidget) openUrlInBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	default: // Linux and others
		cmd = exec.Command("xdg-open", url)
	}

	err := cmd.Start()
	if err != nil {
		msgBox := qt.NewQMessageBox(w.QWidget)
		msgBox.SetWindowTitle("Browser Error")
		msgBox.SetText(fmt.Sprintf("Could not open URL in browser:\n\n%v\n\nURL: %s", err, url))
		msgBox.SetIcon(qt.QMessageBox__Warning)
		msgBox.SetStandardButtons(qt.QMessageBox__Ok)
		msgBox.Exec()
	}
}

// GetLesson returns the lesson associated with this widget
func (w *MediaLessonWidget) GetLesson() *lesson.Lesson {
	return w.lesson
}

// SetLesson sets a new lesson for this widget
func (w *MediaLessonWidget) SetLesson(lesson *lesson.Lesson) {
	w.lesson = lesson
	w.updateData()
}
