package words

import (
	"fmt"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/logging"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// WordsLessonWidget represents a complete lesson widget with Enter/Teach/Results tabs
type WordsLessonWidget struct {
	*widgets.QWidget

	lesson *lesson.Lesson
	logger *logging.Logger

	// Main tab widget
	tabWidget *widgets.QTabWidget

	// Tab widgets
	enterWidget   *EnterTabWidget
	teachWidget   *TeachTabWidget
	resultsWidget *ResultsTabWidget

	// Signals
	lessonChanged *core.QObject
	tabChanged    *core.QObject
}

// NewWordsLessonWidget creates a new words lesson widget
func NewWordsLessonWidget(lesson *lesson.Lesson, parent widgets.QWidget_ITF) *WordsLessonWidget {
	widget := &WordsLessonWidget{
		QWidget:       widgets.NewQWidget(parent, 0),
		lesson:        lesson,
		logger:        logging.NewLogger("WordsLessonWidget"),
		lessonChanged: core.NewQObject(nil),
		tabChanged:    core.NewQObject(nil),
	}

	widget.setupUI()
	widget.connectSignals()
	widget.updateLesson()

	return widget
}

// setupUI initializes the user interface
func (w *WordsLessonWidget) setupUI() {
	layout := widgets.NewQVBoxLayout()
	w.SetLayout(layout)

	// Create main tab widget
	w.tabWidget = widgets.NewQTabWidget(w)
	layout.AddWidget(w.tabWidget, 0, 0)

	// Create Enter tab
	w.enterWidget = NewEnterTabWidget(w.lesson, w)
	w.tabWidget.AddTab(w.enterWidget, "Enter")

	// Create Teach tab
	w.teachWidget = NewTeachTabWidget(w.lesson, w)
	w.tabWidget.AddTab(w.teachWidget, "Teach")

	// Create Results tab
	w.resultsWidget = NewResultsTabWidget(w.lesson, w)
	w.tabWidget.AddTab(w.resultsWidget, "Results")

	// Connect teach widget to results widget for session completion
	w.teachWidget.SetSessionCompletedCallback(func(session *TeachingSession) {
		w.logger.Event("Teaching session completed - adding to results")
		w.resultsWidget.AddSession(session)
		// Auto-switch to Results tab to show the results
		w.tabWidget.SetCurrentIndex(2)
	})

	w.logger.Success("Created lesson widget with 3 tabs")
}

// connectSignals connects widget signals
func (w *WordsLessonWidget) connectSignals() {
	// Connect tab change signal
	w.tabWidget.ConnectCurrentChanged(func(index int) {
		w.logger.Event("Tab changed to index: %d", index)
		// Qt signal handling - will be implemented with proper Qt bindings

		// Update widgets when switching tabs
		switch index {
		case 0: // Enter tab
			w.enterWidget.UpdateLesson(w.lesson)
		case 1: // Teach tab
			w.teachWidget.UpdateLesson(w.lesson)
		case 2: // Results tab
			w.resultsWidget.UpdateLesson(w.lesson)
		}
	})

	// Connect enter widget signals
	// Connect enter widget signals - will be implemented with proper Qt bindings
	// w.enterWidget.Connect("lessonChanged", func() {
	//     w.logger.Event("Lesson changed from Enter tab")
	// })
	w.logger.LegacyReminder("Qt signal connections", "legacy/modules/org/openteacher/interfaces/qt/lessons/words/words.py", "proper signal handling needed")

	// Connect teach widget signals - will be implemented with proper Qt bindings
	// w.teachWidget.Connect("lessonCompleted", func() {
	//     w.logger.Event("Lesson completed - switching to Results tab")
	//     w.tabWidget.SetCurrentIndex(2) // Switch to Results tab
	// })
	w.logger.LegacyReminder("Teaching session completion signals", "legacy/modules/org/openteacher/interfaces/qt/lessons/words/words.py", "Qt signal implementation needed")
}

// UpdateLesson updates the lesson data and refreshes all tabs
func (w *WordsLessonWidget) UpdateLesson(newLesson *lesson.Lesson) {
	w.lesson = newLesson
	w.updateLesson()
}

// updateLesson refreshes all tab widgets with current lesson data
func (w *WordsLessonWidget) updateLesson() {
	if w.lesson == nil {
		w.logger.Warning("No lesson data to update")
		return
	}

	w.logger.Info("Updating lesson widget with %d words", w.lesson.Data.List.GetWordCount())

	// Update all tab widgets
	w.enterWidget.UpdateLesson(w.lesson)
	w.teachWidget.UpdateLesson(w.lesson)
	w.resultsWidget.UpdateLesson(w.lesson)

	// Update window title
	title := w.lesson.Data.List.Title
	if title == "" {
		title = "Unnamed Lesson"
	}
	w.SetWindowTitle(fmt.Sprintf("Word Lesson: %s", title))
}

// GetCurrentTab returns the currently active tab index
func (w *WordsLessonWidget) GetCurrentTab() int {
	return w.tabWidget.CurrentIndex()
}

// SetCurrentTab sets the active tab
func (w *WordsLessonWidget) SetCurrentTab(index int) {
	if index >= 0 && index < w.tabWidget.Count() {
		w.tabWidget.SetCurrentIndex(index)
	}
}

// EnterTabWidget handles lesson editing and entry
type EnterTabWidget struct {
	*widgets.QWidget

	lesson *lesson.Lesson
	logger *logging.Logger

	// UI components
	titleEdit        *widgets.QLineEdit
	qLanguageEdit    *widgets.QLineEdit
	aLanguageEdit    *widgets.QLineEdit
	wordsTable       *widgets.QTableWidget
	addWordButton    *widgets.QPushButton
	removeWordButton *widgets.QPushButton
}

// NewEnterTabWidget creates a new Enter tab widget
func NewEnterTabWidget(lesson *lesson.Lesson, parent widgets.QWidget_ITF) *EnterTabWidget {
	widget := &EnterTabWidget{
		QWidget: widgets.NewQWidget(parent, 0),
		lesson:  lesson,
		logger:  logging.NewLogger("EnterTabWidget"),
	}

	widget.setupUI()
	widget.connectSignals()
	return widget
}

// setupUI initializes the Enter tab interface
func (w *EnterTabWidget) setupUI() {
	layout := widgets.NewQVBoxLayout()
	w.SetLayout(layout)

	// Lesson properties section
	propsGroup := widgets.NewQGroupBox2("Lesson Properties", w)
	propsLayout := widgets.NewQFormLayout(propsGroup)

	w.titleEdit = widgets.NewQLineEdit(w)
	w.configureInternationalInput(w.titleEdit)
	propsLayout.AddRow3("Title:", w.titleEdit)

	w.qLanguageEdit = widgets.NewQLineEdit(w)
	w.configureInternationalInput(w.qLanguageEdit)
	propsLayout.AddRow3("Question Language:", w.qLanguageEdit)

	w.aLanguageEdit = widgets.NewQLineEdit(w)
	w.configureInternationalInput(w.aLanguageEdit)
	propsLayout.AddRow3("Answer Language:", w.aLanguageEdit)

	layout.AddWidget(propsGroup, 0, 0)

	// Word pairs section
	wordsGroup := widgets.NewQGroupBox2("Word Pairs", w)
	wordsLayout := widgets.NewQVBoxLayout()
	wordsGroup.SetLayout(wordsLayout)

	// Buttons
	buttonLayout := widgets.NewQHBoxLayout()
	w.addWordButton = widgets.NewQPushButton2("Add Word", w)
	w.removeWordButton = widgets.NewQPushButton2("Remove Word", w)
	buttonLayout.AddWidget(w.addWordButton, 0, 0)
	buttonLayout.AddWidget(w.removeWordButton, 0, 0)
	buttonLayout.AddStretch(0)

	wordsLayout.AddLayout(buttonLayout, 0)

	// Words table
	w.wordsTable = widgets.NewQTableWidget2(0, 3, w)
	w.wordsTable.SetHorizontalHeaderLabels([]string{"Questions", "Answers", "Comment"})
	w.wordsTable.HorizontalHeader().SetStretchLastSection(true)
	wordsLayout.AddWidget(w.wordsTable, 0, 0)

	layout.AddWidget(wordsGroup, 0, 0)

	w.logger.Success("Enter tab UI created")
}

// connectSignals connects Enter tab signals
func (w *EnterTabWidget) connectSignals() {
	// Title changed
	w.titleEdit.ConnectTextChanged(func(text string) {
		if w.lesson != nil {
			w.lesson.Data.List.Title = text
			// Qt signal emission - will be implemented with proper Qt bindings
			w.logger.LegacyReminder("lessonChanged signal emission", "legacy/modules/org/openteacher/interfaces/qt/lessons/words/words.py", "proper Qt signal emission needed")
		}
	})

	// Language fields changed
	w.qLanguageEdit.ConnectTextChanged(func(text string) {
		if w.lesson != nil {
			w.lesson.Data.List.QuestionLanguage = text
			// Qt signal emission - will be implemented with proper Qt bindings
			w.logger.LegacyReminder("lessonChanged signal for question language", "legacy/modules/org/openteacher/interfaces/qt/lessons/words/words.py", "signal emission needed")
		}
	})

	w.aLanguageEdit.ConnectTextChanged(func(text string) {
		if w.lesson != nil {
			w.lesson.Data.List.AnswerLanguage = text
			// Qt signal emission - will be implemented with proper Qt bindings
			w.logger.LegacyReminder("lessonChanged signal for answer language", "legacy/modules/org/openteacher/interfaces/qt/lessons/words/words.py", "signal emission needed")
		}
	})

	// Button clicks
	w.addWordButton.ConnectClicked(func(checked bool) {
		w.addNewWord()
	})

	w.removeWordButton.ConnectClicked(func(checked bool) {
		w.removeSelectedWord()
	})
}

// UpdateLesson updates the Enter tab with lesson data
func (w *EnterTabWidget) UpdateLesson(lesson *lesson.Lesson) {
	w.lesson = lesson
	if lesson == nil {
		return
	}

	// Update fields
	w.titleEdit.SetText(lesson.Data.List.Title)
	w.qLanguageEdit.SetText(lesson.Data.List.QuestionLanguage)
	w.aLanguageEdit.SetText(lesson.Data.List.AnswerLanguage)

	// Update table
	w.updateWordsTable()
	w.logger.Info("Enter tab updated with lesson data")
}

// updateWordsTable refreshes the words table
func (w *EnterTabWidget) updateWordsTable() {
	if w.lesson == nil {
		return
	}

	items := w.lesson.Data.List.Items
	w.wordsTable.SetRowCount(len(items))

	for i, item := range items {
		questionsText := strings.Join(item.Questions, "; ")
		answersText := strings.Join(item.Answers, "; ")

		questionItem := widgets.NewQTableWidgetItem2(questionsText, 0)
		answerItem := widgets.NewQTableWidgetItem2(answersText, 0)
		commentItem := widgets.NewQTableWidgetItem2(item.Comment, 0)

		w.wordsTable.SetItem(i, 0, questionItem)
		w.wordsTable.SetItem(i, 1, answerItem)
		w.wordsTable.SetItem(i, 2, commentItem)
	}

	w.wordsTable.ResizeColumnsToContents()
}

// addNewWord adds a new word pair
func (w *EnterTabWidget) addNewWord() {
	if w.lesson == nil {
		return
	}

	newItem := lesson.WordItem{
		Questions: []string{"New Question"},
		Answers:   []string{"New Answer"},
		Comment:   "",
	}

	w.lesson.Data.List.Items = append(w.lesson.Data.List.Items, newItem)
	w.updateWordsTable()
	// Qt signal emission - will be implemented with proper Qt bindings
	w.logger.LegacyReminder("lessonChanged signal for adding word", "legacy/modules/org/openteacher/interfaces/qt/lessons/words/words.py", "signal emission needed")
	w.logger.Action("Added new word pair")
}

// removeSelectedWord removes the selected word pair
func (w *EnterTabWidget) removeSelectedWord() {
	if w.lesson == nil {
		return
	}

	currentRow := w.wordsTable.CurrentRow()
	if currentRow >= 0 && currentRow < len(w.lesson.Data.List.Items) {
		// Remove item from slice
		items := w.lesson.Data.List.Items
		w.lesson.Data.List.Items = append(items[:currentRow], items[currentRow+1:]...)

		w.updateWordsTable()
		// Qt signal emission - will be implemented with proper Qt bindings
		w.logger.LegacyReminder("lessonChanged signal for removing word", "legacy/modules/org/openteacher/interfaces/qt/lessons/words/words.py", "signal emission needed")
		w.logger.Action("Removed word pair at row %d", currentRow)
	}
}

// TeachingResult represents the result of answering a single question
type TeachingResult struct {
	Question      string
	CorrectAnswer string
	UserAnswer    string
	IsCorrect     bool
	ItemIndex     int
}

// TeachingSession represents a complete teaching session with all results
type TeachingSession struct {
	Results        []TeachingResult
	TotalQuestions int
	CorrectCount   int
	Score          int // percentage
	Completed      bool
}

// TeachTabWidget handles the teaching/quiz functionality
type TeachTabWidget struct {
	*widgets.QWidget

	lesson *lesson.Lesson
	logger *logging.Logger

	// UI components
	startButton   *widgets.QPushButton
	statusLabel   *widgets.QLabel
	progressBar   *widgets.QProgressBar
	questionLabel *widgets.QLabel
	answerEdit    *widgets.QLineEdit
	submitButton  *widgets.QPushButton
	nextButton    *widgets.QPushButton
	resultLabel   *widgets.QLabel
	unicodeButton *widgets.QPushButton

	// Unicode character picker
	unicodePicker *IntegratedUnicodePicker

	// Teaching state
	currentIndex   int
	correctAnswers int
	totalQuestions int
	isTeaching     bool

	// Session tracking
	currentSession   *TeachingSession
	sessionCompleted func(*TeachingSession) // Callback for when session completes
}

// NewTeachTabWidget creates a new Teach tab widget
func NewTeachTabWidget(lesson *lesson.Lesson, parent widgets.QWidget_ITF) *TeachTabWidget {
	widget := &TeachTabWidget{
		QWidget: widgets.NewQWidget(parent, 0),
		lesson:  lesson,
		logger:  logging.NewLogger("TeachTabWidget"),
	}

	// Create integrated Unicode picker
	widget.unicodePicker = NewIntegratedUnicodePicker("", widget)
	widget.logger.Debug("Created integrated Unicode picker widget")

	widget.setupUI()
	widget.connectSignals()
	return widget
}

// SetSessionCompletedCallback sets the callback function called when a session completes
func (w *TeachTabWidget) SetSessionCompletedCallback(callback func(*TeachingSession)) {
	w.sessionCompleted = callback
}

// setupUI initializes the Teach tab interface
func (w *TeachTabWidget) setupUI() {
	layout := widgets.NewQVBoxLayout()
	w.SetLayout(layout)

	// Status section
	statusGroup := widgets.NewQGroupBox2("Teaching Progress", w)
	statusLayout := widgets.NewQVBoxLayout()
	statusGroup.SetLayout(statusLayout)

	w.statusLabel = widgets.NewQLabel2("Ready to start teaching", w, 0)
	statusLayout.AddWidget(w.statusLabel, 0, 0)

	w.progressBar = widgets.NewQProgressBar(w)
	statusLayout.AddWidget(w.progressBar, 0, 0)

	layout.AddWidget(statusGroup, 0, 0)

	// Question section
	questionGroup := widgets.NewQGroupBox2("Current Question", w)
	questionLayout := widgets.NewQVBoxLayout()
	questionGroup.SetLayout(questionLayout)

	w.questionLabel = widgets.NewQLabel2("Click 'Start Teaching' to begin", w, 0)
	w.questionLabel.SetWordWrap(true)
	w.questionLabel.SetAlignment(core.Qt__AlignCenter)
	questionLayout.AddWidget(w.questionLabel, 0, 0)

	// Answer input with Unicode picker
	answerLayout := widgets.NewQHBoxLayout()
	answerLayout.AddWidget(widgets.NewQLabel2("Your Answer:", w, 0), 0, 0)
	w.answerEdit = widgets.NewQLineEdit(w)
	w.answerEdit.SetEnabled(false)

	// Minimal configuration for dead keys and AltGr to work properly
	// Don't override input method hints - let Qt use system defaults
	// w.answerEdit.SetInputMethodHints(core.Qt__ImhNone) // REMOVED - was blocking dead keys

	// Don't force input method attributes - let system handle it naturally
	// w.answerEdit.SetAttribute(core.Qt__WA_InputMethodEnabled, true) // REMOVED

	// Configure font to support international characters
	font := w.answerEdit.Font()
	font.SetFamily("DejaVu Sans, Liberation Sans, Arial, sans-serif")
	font.SetPointSize(12)
	w.answerEdit.SetFont(font)

	answerLayout.AddWidget(w.answerEdit, 0, 0)

	// Unicode character picker button
	w.unicodeButton = widgets.NewQPushButton2("⚿ Characters", w)
	w.unicodeButton.SetToolTip("Show/hide Unicode character picker for accented letters and special characters")
	w.unicodeButton.SetCheckable(true)
	w.unicodeButton.SetEnabled(false)
	w.unicodeButton.SetStyleSheet(`
		QPushButton {
			background-color: #e8f4fd;
			border: 1px solid #0078d4;
			border-radius: 4px;
			padding: 8px 12px;
			font-weight: bold;
			color: #0078d4;
		}
		QPushButton:hover {
			background-color: #deecf9;
		}
		QPushButton:pressed {
			background-color: #c7e0f4;
		}
		QPushButton:disabled {
			background-color: #f3f2f1;
			border-color: #d2d0ce;
			color: #a19f9d;
		}
	`)
	answerLayout.AddWidget(w.unicodeButton, 0, 0)

	questionLayout.AddLayout(answerLayout, 0)

	// Add integrated Unicode picker (initially hidden)
	w.unicodePicker.Hide()
	questionLayout.AddWidget(w.unicodePicker, 0, 0)

	w.logger.Info("Configured answer input field with integrated Unicode character picker")
	w.logger.Action("Dead keys may not work - use Unicode picker button for accented characters")

	// Result label
	w.resultLabel = widgets.NewQLabel(w, 0)
	w.resultLabel.SetAlignment(core.Qt__AlignCenter)
	w.resultLabel.SetVisible(false)
	questionLayout.AddWidget(w.resultLabel, 0, 0)

	layout.AddWidget(questionGroup, 0, 0)

	// Buttons
	buttonLayout := widgets.NewQHBoxLayout()
	w.startButton = widgets.NewQPushButton2("Start Teaching", w)
	w.submitButton = widgets.NewQPushButton2("Submit Answer", w)
	w.submitButton.SetEnabled(false)
	w.nextButton = widgets.NewQPushButton2("Next Question", w)
	w.nextButton.SetEnabled(false)

	buttonLayout.AddWidget(w.startButton, 0, 0)
	buttonLayout.AddWidget(w.submitButton, 0, 0)
	buttonLayout.AddWidget(w.nextButton, 0, 0)
	buttonLayout.AddStretch(0)

	layout.AddLayout(buttonLayout, 0)

	w.logger.Success("Teach tab UI created")
}

// connectSignals connects Teach tab signals
func (w *TeachTabWidget) connectSignals() {
	w.startButton.ConnectClicked(func(checked bool) {
		w.startTeaching()
	})

	w.submitButton.ConnectClicked(func(checked bool) {
		w.submitAnswer()
	})

	w.nextButton.ConnectClicked(func(checked bool) {
		w.nextQuestion()
	})

	w.unicodeButton.ConnectToggled(func(checked bool) {
		w.logger.Debug("Unicode picker button toggled: %v", checked)
		w.toggleUnicodePicker(checked)
	})

	w.answerEdit.ConnectReturnPressed(func() {
		if w.submitButton.IsEnabled() {
			w.submitAnswer()
		} else if w.nextButton.IsEnabled() {
			w.nextQuestion()
		}
	})
}

// UpdateLesson updates the Teach tab with lesson data
func (w *TeachTabWidget) UpdateLesson(lesson *lesson.Lesson) {
	w.lesson = lesson
	w.resetTeachingState()
}

// startTeaching begins the teaching session
func (w *TeachTabWidget) startTeaching() {
	if w.lesson == nil || len(w.lesson.Data.List.Items) == 0 {
		w.statusLabel.SetText("No words available for teaching")
		return
	}

	w.isTeaching = true
	w.currentIndex = 0
	w.correctAnswers = 0
	w.totalQuestions = len(w.lesson.Data.List.Items)

	// Initialize new teaching session
	w.currentSession = &TeachingSession{
		Results:        make([]TeachingResult, 0, w.totalQuestions),
		TotalQuestions: w.totalQuestions,
		CorrectCount:   0,
		Score:          0,
		Completed:      false,
	}

	w.startButton.SetEnabled(false)
	w.answerEdit.SetEnabled(true)
	w.answerEdit.SetFocus2()
	w.submitButton.SetEnabled(true)
	w.unicodeButton.SetEnabled(true)

	// Set Unicode picker target
	w.unicodePicker.SetTargetEdit(w.answerEdit)

	w.showCurrentQuestion()
	w.logger.Action("Started teaching session with %d words", w.totalQuestions)
}

// showCurrentQuestion displays the current question
func (w *TeachTabWidget) showCurrentQuestion() {
	if w.lesson == nil || w.currentIndex >= len(w.lesson.Data.List.Items) {
		w.finishTeaching()
		return
	}

	item := w.lesson.Data.List.Items[w.currentIndex]
	question := strings.Join(item.Questions, " / ")

	w.questionLabel.SetText(fmt.Sprintf("Question: %s", question))
	w.answerEdit.Clear()
	w.answerEdit.SetFocus2()
	w.resultLabel.SetVisible(false)

	// Update progress
	progress := int((float64(w.currentIndex) / float64(w.totalQuestions)) * 100)
	w.progressBar.SetValue(progress)
	w.statusLabel.SetText(fmt.Sprintf("Question %d of %d (Score: %d/%d correct)",
		w.currentIndex+1, w.totalQuestions, w.correctAnswers, w.currentIndex))
}

// submitAnswer checks the user's answer
func (w *TeachTabWidget) submitAnswer() {
	if w.lesson == nil || w.currentIndex >= len(w.lesson.Data.List.Items) || w.currentSession == nil {
		return
	}

	userAnswer := strings.TrimSpace(w.answerEdit.Text())
	if userAnswer == "" {
		return
	}

	item := w.lesson.Data.List.Items[w.currentIndex]
	correct := false

	// Check if answer matches any of the correct answers (case-insensitive)
	for _, answer := range item.Answers {
		if strings.EqualFold(userAnswer, strings.TrimSpace(answer)) {
			correct = true
			break
		}
	}

	// Create teaching result record
	result := TeachingResult{
		Question:      strings.Join(item.Questions, " / "),
		CorrectAnswer: strings.Join(item.Answers, " / "),
		UserAnswer:    userAnswer,
		IsCorrect:     correct,
		ItemIndex:     w.currentIndex,
	}

	// Add to session results
	w.currentSession.Results = append(w.currentSession.Results, result)

	// Update score and show result
	if correct {
		w.correctAnswers++
		w.currentSession.CorrectCount++
		w.resultLabel.SetText("[CORRECT!]")
		w.resultLabel.SetStyleSheet("color: green; font-weight: bold; background-color: lightgreen; padding: 5px; border-radius: 3px;")
	} else {
		w.resultLabel.SetText(fmt.Sprintf("[INCORRECT] Correct answer(s): %s", result.CorrectAnswer))
		w.resultLabel.SetStyleSheet("color: red; font-weight: bold; background-color: lightcoral; padding: 5px; border-radius: 3px;")
	}

	w.resultLabel.SetVisible(true)
	w.answerEdit.SetEnabled(false)
	w.submitButton.SetEnabled(false)
	w.nextButton.SetEnabled(true)
	w.nextButton.SetFocus2()

	w.logger.Info("Answer submitted: %s (correct: %v)", userAnswer, correct)
}

// nextQuestion moves to the next question
func (w *TeachTabWidget) nextQuestion() {
	w.currentIndex++

	if w.currentIndex >= len(w.lesson.Data.List.Items) {
		w.finishTeaching()
	} else {
		w.answerEdit.SetEnabled(true)
		w.submitButton.SetEnabled(true)
		w.nextButton.SetEnabled(false)
		w.showCurrentQuestion()
	}
}

// finishTeaching completes the teaching session
func (w *TeachTabWidget) finishTeaching() {
	w.isTeaching = false
	percentage := 0
	if w.totalQuestions > 0 {
		percentage = int((float64(w.correctAnswers) / float64(w.totalQuestions)) * 100)
	}

	// Complete the session record
	if w.currentSession != nil {
		w.currentSession.Score = percentage
		w.currentSession.Completed = true
	}

	w.questionLabel.SetText(fmt.Sprintf("Teaching completed! Final Score: %d/%d correct (%d%%)",
		w.correctAnswers, w.totalQuestions, percentage))

	w.answerEdit.SetEnabled(false)
	w.submitButton.SetEnabled(false)
	w.nextButton.SetEnabled(false)
	w.startButton.SetEnabled(true)
	w.startButton.SetText("Start Again")
	w.unicodeButton.SetEnabled(false)

	w.progressBar.SetValue(100)
	w.statusLabel.SetText("Teaching session completed")

	// Notify parent widget of session completion
	if w.sessionCompleted != nil && w.currentSession != nil {
		w.sessionCompleted(w.currentSession)
	}

	w.logger.Success("Teaching completed: %d/%d correct (%d%%)", w.correctAnswers, w.totalQuestions, percentage)
}

// resetTeachingState resets the teaching state
func (w *TeachTabWidget) resetTeachingState() {
	w.isTeaching = false
	w.currentIndex = 0
	w.correctAnswers = 0
	w.totalQuestions = 0

	w.startButton.SetEnabled(true)
	w.startButton.SetText("Start Teaching")
	w.answerEdit.SetEnabled(false)
	w.answerEdit.Clear()
	w.submitButton.SetEnabled(false)
	w.nextButton.SetEnabled(false)
	w.unicodeButton.SetEnabled(false)
	w.resultLabel.SetVisible(false)
	w.progressBar.SetValue(0)

	// Hide Unicode picker
	w.unicodePicker.Hide()
	w.unicodeButton.SetChecked(false)

	if w.lesson != nil && len(w.lesson.Data.List.Items) > 0 {
		w.statusLabel.SetText("Ready to start teaching")
		w.questionLabel.SetText("Click 'Start Teaching' to begin")
	} else {
		w.statusLabel.SetText("No words available for teaching")
		w.questionLabel.SetText("Please add words in the Enter tab")
	}
}

// toggleUnicodePicker toggles the Unicode character picker visibility
func (w *TeachTabWidget) toggleUnicodePicker(show bool) {
	w.logger.Debug("toggleUnicodePicker called - show: %v", show)

	if w.unicodePicker == nil {
		w.logger.Error("Unicode picker is nil!")
		return
	}

	if show {
		// Set the target edit widget first
		w.unicodePicker.SetTargetEdit(w.answerEdit)
		w.logger.Debug("Set target edit widget for Unicode picker")

		// Show the integrated picker
		w.unicodePicker.Show()
		w.logger.Info("Integrated Unicode picker shown")
	} else {
		// Hide the integrated picker
		w.unicodePicker.Hide()
		w.logger.Info("Integrated Unicode picker hidden")
	}
}

// configureInternationalInput sets up minimal Qt config to allow dead keys/AltGr
func (w *EnterTabWidget) configureInternationalInput(lineEdit *widgets.QLineEdit) {
	// Don't override input method hints - let system handle dead keys naturally
	// lineEdit.SetInputMethodHints(core.Qt__ImhNone) // REMOVED - was interfering
	// lineEdit.SetAttribute(core.Qt__WA_InputMethodEnabled, true) // REMOVED

	// Set Unicode-supporting font only
	font := lineEdit.Font()
	font.SetFamily("DejaVu Sans, Liberation Sans, Arial, sans-serif")
	font.SetPointSize(11)
	lineEdit.SetFont(font)

	w.logger.Debug("Configured minimal Qt setup for dead keys/AltGr support")
}

// ResultsTabWidget displays teaching results and statistics
type ResultsTabWidget struct {
	*widgets.QWidget

	lesson *lesson.Lesson
	logger *logging.Logger

	// UI components
	overviewLabel *widgets.QLabel
	resultsTable  *widgets.QTableWidget

	// Results data
	sessions []*TeachingSession
}

// NewResultsTabWidget creates a new Results tab widget
func NewResultsTabWidget(lesson *lesson.Lesson, parent widgets.QWidget_ITF) *ResultsTabWidget {
	widget := &ResultsTabWidget{
		QWidget: widgets.NewQWidget(parent, 0),
		lesson:  lesson,
		logger:  logging.NewLogger("ResultsTabWidget"),
	}

	widget.setupUI()
	return widget
}

// setupUI initializes the Results tab interface
func (w *ResultsTabWidget) setupUI() {
	layout := widgets.NewQVBoxLayout()
	w.SetLayout(layout)

	// Overview section
	overviewGroup := widgets.NewQGroupBox2("Results Overview", w)
	overviewLayout := widgets.NewQVBoxLayout()
	overviewGroup.SetLayout(overviewLayout)

	w.overviewLabel = widgets.NewQLabel2("No teaching results available yet", w, 0)
	w.overviewLabel.SetAlignment(core.Qt__AlignCenter)
	overviewLayout.AddWidget(w.overviewLabel, 0, 0)

	layout.AddWidget(overviewGroup, 0, 0)

	// Detailed results
	detailsGroup := widgets.NewQGroupBox2("Latest Session Results", w)
	detailsLayout := widgets.NewQVBoxLayout()
	detailsGroup.SetLayout(detailsLayout)

	w.resultsTable = widgets.NewQTableWidget2(0, 4, w)
	w.resultsTable.SetHorizontalHeaderLabels([]string{"Question", "Correct Answer", "Your Answer", "Result"})
	w.resultsTable.HorizontalHeader().SetStretchLastSection(true)
	w.resultsTable.SetAlternatingRowColors(true)
	detailsLayout.AddWidget(w.resultsTable, 0, 0)

	layout.AddWidget(detailsGroup, 0, 0)

	w.logger.Success("Results tab UI created")
}

// UpdateLesson updates the Results tab with lesson data
func (w *ResultsTabWidget) UpdateLesson(lesson *lesson.Lesson) {
	w.lesson = lesson
	w.updateResultsDisplay()
}

// AddSession adds a completed teaching session to the results
func (w *ResultsTabWidget) AddSession(session *TeachingSession) {
	if session == nil || !session.Completed {
		return
	}

	w.sessions = append(w.sessions, session)
	w.updateResultsDisplay()
	w.logger.Info("Added teaching session results: %d/%d correct (%d%%)",
		session.CorrectCount, session.TotalQuestions, session.Score)
}

// updateResultsDisplay updates the results display
func (w *ResultsTabWidget) updateResultsDisplay() {
	if w.lesson == nil {
		w.overviewLabel.SetText("No lesson data available")
		return
	}

	wordCount := len(w.lesson.Data.List.Items)

	if len(w.sessions) == 0 {
		w.overviewLabel.SetText(fmt.Sprintf("Lesson contains %d word pairs\n\nComplete a teaching session to see detailed results here.", wordCount))
		w.resultsTable.SetRowCount(0)
	} else {
		// Show statistics from latest session
		latestSession := w.sessions[len(w.sessions)-1]
		totalSessions := len(w.sessions)

		// Calculate average score across all sessions
		totalScore := 0
		for _, session := range w.sessions {
			totalScore += session.Score
		}
		avgScore := totalScore / totalSessions

		overviewText := fmt.Sprintf(`Lesson: %d word pairs | Sessions completed: %d

Latest Session: %d/%d correct (%d%%)
Average Score: %d%%

Detailed results from latest session:`,
			wordCount, totalSessions,
			latestSession.CorrectCount, latestSession.TotalQuestions, latestSession.Score,
			avgScore)

		w.overviewLabel.SetText(overviewText)

		// Populate results table with latest session details
		w.populateResultsTable(latestSession)
	}

	w.logger.Info("Results display updated - %d sessions available", len(w.sessions))
}

// populateResultsTable fills the results table with session data
func (w *ResultsTabWidget) populateResultsTable(session *TeachingSession) {
	if session == nil {
		return
	}

	w.resultsTable.SetRowCount(len(session.Results))

	for i, result := range session.Results {
		// Question
		questionItem := widgets.NewQTableWidgetItem2(result.Question, 0)
		w.resultsTable.SetItem(i, 0, questionItem)

		// Correct Answer
		correctItem := widgets.NewQTableWidgetItem2(result.CorrectAnswer, 0)
		w.resultsTable.SetItem(i, 1, correctItem)

		// User Answer
		userItem := widgets.NewQTableWidgetItem2(result.UserAnswer, 0)
		w.resultsTable.SetItem(i, 2, userItem)

		// Result (CORRECT/WRONG)
		var resultText string
		var resultItem *widgets.QTableWidgetItem
		if result.IsCorrect {
			resultText = "[CORRECT]"
			resultItem = widgets.NewQTableWidgetItem2(resultText, 0)
			brush := gui.NewQBrush3(gui.NewQColor3(200, 255, 200, 100), core.Qt__SolidPattern)
			resultItem.SetBackground(brush) // Light green background
		} else {
			resultText = "[WRONG]"
			resultItem = widgets.NewQTableWidgetItem2(resultText, 0)
			brush := gui.NewQBrush3(gui.NewQColor3(255, 200, 200, 100), core.Qt__SolidPattern)
			resultItem.SetBackground(brush) // Light red background
		}
		w.resultsTable.SetItem(i, 3, resultItem)
	}

	w.resultsTable.ResizeColumnsToContents()
}
