package words

import (
	"fmt"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/logging"
	"github.com/mappu/miqt/qt"
)

// WordsLessonWidget represents a complete lesson widget with Enter/Teach/Results tabs
type WordsLessonWidget struct {
	*qt.QWidget

	lesson *lesson.Lesson
	logger *logging.Logger

	// Main tab widget
	tabWidget *qt.QTabWidget

	// Tab widgets
	enterWidget   *EnterTabWidget
	teachWidget   *TeachTabWidget
	resultsWidget *ResultsTabWidget

	// Signals
	lessonChanged *qt.QObject
	tabChanged    *qt.QObject
}

// NewWordsLessonWidget creates a new words lesson widget
func NewWordsLessonWidget(lesson *lesson.Lesson, parent *qt.QWidget) *WordsLessonWidget {
	widget := &WordsLessonWidget{
		QWidget:       qt.NewQWidget(parent),
		lesson:        lesson,
		logger:        logging.NewLogger("WordsLessonWidget"),
		lessonChanged: qt.NewQObject(),
		tabChanged:    qt.NewQObject(),
	}

	widget.setupUI()
	widget.connectSignals()
	widget.updateLesson()

	return widget
}

// setupUI initializes the user interface
func (w *WordsLessonWidget) setupUI() {
	layout := qt.NewQVBoxLayout(w.QWidget)

	// Create main tab widget
	w.tabWidget = qt.NewQTabWidget(w.QWidget)
	layout.AddWidget(w.tabWidget.QWidget)

	// Create Enter tab
	w.enterWidget = NewEnterTabWidget(w.lesson, w.QWidget)
	w.tabWidget.AddTab(w.enterWidget.QWidget, "Enter")

	// Create Teach tab
	w.teachWidget = NewTeachTabWidget(w.lesson, w.QWidget)
	w.tabWidget.AddTab(w.teachWidget.QWidget, "Teach")

	// Create Results tab
	w.resultsWidget = NewResultsTabWidget(w.lesson, w.QWidget)
	w.tabWidget.AddTab(w.resultsWidget.QWidget, "Results")

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
	w.tabWidget.OnCurrentChanged(func(index int) {
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
	*qt.QWidget

	lesson *lesson.Lesson
	logger *logging.Logger

	// UI components
	titleEdit        *qt.QLineEdit
	qLanguageEdit    *qt.QLineEdit
	aLanguageEdit    *qt.QLineEdit
	wordsTable       *qt.QTableWidget
	addWordButton    *qt.QPushButton
	removeWordButton *qt.QPushButton
}

// NewEnterTabWidget creates a new Enter tab widget
func NewEnterTabWidget(lesson *lesson.Lesson, parent *qt.QWidget) *EnterTabWidget {
	widget := &EnterTabWidget{
		QWidget: qt.NewQWidget(parent),
		lesson:  lesson,
		logger:  logging.NewLogger("EnterTabWidget"),
	}

	widget.setupUI()
	widget.connectSignals()
	return widget
}

// setupUI initializes the Enter tab interface
func (w *EnterTabWidget) setupUI() {
	layout := qt.NewQVBoxLayout(w.QWidget)

	// Lesson properties section
	propsGroup := qt.NewQGroupBox(w.QWidget)
	propsGroup.SetTitle("Lesson Properties")
	propsLayout := qt.NewQFormLayout(propsGroup.QWidget)

	w.titleEdit = qt.NewQLineEdit(w.QWidget)
	w.titleEdit.SetPlaceholderText("Enter lesson title")
	propsLayout.AddRow3("Title:", w.titleEdit.QWidget)

	w.qLanguageEdit = qt.NewQLineEdit(w.QWidget)
	w.qLanguageEdit.SetPlaceholderText("Question language")
	propsLayout.AddRow3("Question Language:", w.qLanguageEdit.QWidget)

	w.aLanguageEdit = qt.NewQLineEdit(w.QWidget)
	w.aLanguageEdit.SetPlaceholderText("Answer language")
	propsLayout.AddRow3("Answer Language:", w.aLanguageEdit.QWidget)

	layout.AddWidget(propsGroup.QWidget)

	// Word pairs section
	wordsGroup := qt.NewQGroupBox(w.QWidget)
	wordsGroup.SetTitle("Word Pairs")
	wordsLayout := qt.NewQVBoxLayout(wordsGroup.QWidget)

	// Buttons
	buttonLayout := qt.NewQHBoxLayout2()
	w.addWordButton = qt.NewQPushButton(w.QWidget)
	w.addWordButton.SetText("Add Word")
	w.removeWordButton = qt.NewQPushButton(w.QWidget)
	w.removeWordButton.SetText("Remove Word")
	buttonLayout.AddWidget(w.addWordButton.QWidget)
	buttonLayout.AddWidget(w.removeWordButton.QWidget)
	buttonLayout.AddStretch()

	wordsLayout.AddLayout2(buttonLayout.QLayout, 0)

	// Words table
	w.wordsTable = qt.NewQTableWidget2()
	w.wordsTable.SetRowCount(0)
	w.wordsTable.SetColumnCount(3)
	w.wordsTable.SetHorizontalHeaderLabels([]string{"Questions", "Answers", "Comment"})
	w.wordsTable.HorizontalHeader().SetStretchLastSection(true)
	wordsLayout.AddWidget(w.wordsTable.QWidget)

	layout.AddWidget(wordsGroup.QWidget)

	w.logger.Success("Enter tab UI created")
}

// connectSignals connects Enter tab signals
func (w *EnterTabWidget) connectSignals() {
	// Title changed
	w.titleEdit.OnTextChanged(func(text string) {
		if w.lesson != nil {
			w.lesson.Data.List.Title = text
			// Qt signal emission - will be implemented with proper Qt bindings
			w.logger.LegacyReminder("lessonChanged signal emission", "legacy/modules/org/openteacher/interfaces/qt/lessons/words/words.py", "proper Qt signal emission needed")
		}
	})

	// Language fields changed
	w.qLanguageEdit.OnTextChanged(func(text string) {
		if w.lesson != nil {
			w.lesson.Data.List.QuestionLanguage = text
			// Qt signal emission - will be implemented with proper Qt bindings
			w.logger.LegacyReminder("lessonChanged signal for question language", "legacy/modules/org/openteacher/interfaces/qt/lessons/words/words.py", "signal emission needed")
		}
	})

	w.aLanguageEdit.OnTextChanged(func(text string) {
		if w.lesson != nil {
			w.lesson.Data.List.AnswerLanguage = text
			// Qt signal emission - will be implemented with proper Qt bindings
			w.logger.LegacyReminder("lessonChanged signal for answer language", "legacy/modules/org/openteacher/interfaces/qt/lessons/words/words.py", "signal emission needed")
		}
	})

	// Button clicks
	w.addWordButton.OnClicked(func() {
		w.addNewWord()
	})

	w.removeWordButton.OnClicked(func() {
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

		questionItem := qt.NewQTableWidgetItem2(questionsText)
		answerItem := qt.NewQTableWidgetItem2(answersText)
		commentItem := qt.NewQTableWidgetItem2(item.Comment)

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
	*qt.QWidget

	lesson *lesson.Lesson
	logger *logging.Logger

	// UI components
	startButton   *qt.QPushButton
	statusLabel   *qt.QLabel
	progressBar   *qt.QProgressBar
	questionLabel *qt.QLabel
	answerEdit    *qt.QLineEdit
	submitButton  *qt.QPushButton
	nextButton    *qt.QPushButton
	resultLabel   *qt.QLabel
	unicodeButton *qt.QPushButton

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
func NewTeachTabWidget(lesson *lesson.Lesson, parent *qt.QWidget) *TeachTabWidget {
	widget := &TeachTabWidget{
		QWidget: qt.NewQWidget(parent),
		lesson:  lesson,
		logger:  logging.NewLogger("TeachTabWidget"),
	}

	// Create integrated Unicode picker
	widget.unicodePicker = NewIntegratedUnicodePicker("", widget.QWidget)
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
	layout := qt.NewQVBoxLayout(w.QWidget)

	// Status section
	statusGroup := qt.NewQGroupBox(w.QWidget)
	statusGroup.SetTitle("Teaching Status")
	statusLayout := qt.NewQVBoxLayout(statusGroup.QWidget)

	w.statusLabel = qt.NewQLabel(w.QWidget)
	w.statusLabel.SetText("Ready to start")
	statusLayout.AddWidget(w.statusLabel.QWidget)

	w.progressBar = qt.NewQProgressBar(w.QWidget)
	statusLayout.AddWidget(w.progressBar.QWidget)

	layout.AddWidget(statusGroup.QWidget)

	// Question section
	questionGroup := qt.NewQGroupBox(w.QWidget)
	questionGroup.SetTitle("Current Question")
	questionLayout := qt.NewQVBoxLayout(questionGroup.QWidget)

	w.questionLabel = qt.NewQLabel(w.QWidget)
	w.questionLabel.SetText("Click 'Start Teaching' to begin")
	w.questionLabel.SetWordWrap(true)
	w.questionLabel.SetAlignment(qt.AlignCenter)
	questionLayout.AddWidget(w.questionLabel.QWidget)

	// Answer input with Unicode picker
	answerLayout := qt.NewQHBoxLayout2()
	answerLabel := qt.NewQLabel(w.QWidget)
	answerLabel.SetText("Your Answer:")
	answerLayout.AddWidget(answerLabel.QWidget)
	w.answerEdit = qt.NewQLineEdit(w.QWidget)
	w.answerEdit.SetEnabled(false)
	answerLayout.AddWidget(w.answerEdit.QWidget)

	// Unicode picker button
	w.unicodeButton = qt.NewQPushButton(w.QWidget)
	w.unicodeButton.SetText("🔤")
	w.unicodeButton.SetToolTip("Special Characters")
	w.unicodeButton.SetEnabled(false)
	answerLayout.AddWidget(w.unicodeButton.QWidget)

	questionLayout.AddLayout2(answerLayout.QLayout, 0)

	// Result label
	w.resultLabel = qt.NewQLabel(w.QWidget)
	w.resultLabel.SetText("")
	w.resultLabel.SetWordWrap(true)
	questionLayout.AddWidget(w.resultLabel.QWidget)

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

	answerLayout.AddWidget(w.answerEdit.QWidget)

	// Unicode character picker button
	w.unicodeButton = qt.NewQPushButton(w.QWidget)
	w.unicodeButton.SetText("⚿ Characters")
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
	answerLayout.AddWidget(w.unicodeButton.QWidget)

	questionLayout.AddLayout2(answerLayout.QLayout, 0)

	// Add integrated Unicode picker (initially hidden)
	w.unicodePicker.Hide()
	questionLayout.AddWidget(w.unicodePicker.QWidget)

	w.logger.Info("Configured answer input field with integrated Unicode character picker")
	w.logger.Action("Dead keys may not work - use Unicode picker button for accented characters")

	// Result label
	// Result label was already created earlier, just set it up
	w.resultLabel.SetAlignment(qt.AlignCenter)
	w.resultLabel.SetVisible(false)

	layout.AddWidget(questionGroup.QWidget)

	// Buttons
	buttonLayout := qt.NewQHBoxLayout2()
	w.startButton = qt.NewQPushButton(w.QWidget)
	w.startButton.SetText("Start Teaching")
	w.submitButton = qt.NewQPushButton(w.QWidget)
	w.submitButton.SetText("Submit Answer")
	w.submitButton.SetEnabled(false)
	w.nextButton = qt.NewQPushButton(w.QWidget)
	w.nextButton.SetText("Next Question")
	w.nextButton.SetEnabled(false)

	buttonLayout.AddWidget(w.startButton.QWidget)
	buttonLayout.AddWidget(w.submitButton.QWidget)
	buttonLayout.AddWidget(w.nextButton.QWidget)
	buttonLayout.AddStretch()

	layout.AddLayout2(buttonLayout.QLayout, 0)

	w.logger.Success("Teach tab UI created")
}

// connectSignals connects Teach tab signals
func (w *TeachTabWidget) connectSignals() {
	w.startButton.OnClicked(func() {
		w.startTeaching()
	})

	w.submitButton.OnClicked(func() {
		w.submitAnswer()
	})

	w.nextButton.OnClicked(func() {
		w.nextQuestion()
	})

	w.unicodeButton.OnToggled(func(checked bool) {
		w.logger.Debug("Unicode picker button toggled: %v", checked)
		w.toggleUnicodePicker(checked)
	})

	w.answerEdit.OnReturnPressed(func() {
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
	w.answerEdit.SetFocus()
	w.submitButton.SetEnabled(true)
	w.unicodeButton.SetEnabled(true)

	// Set Unicode picker target
	// Unicode picker target will be set when needed

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
	w.answerEdit.SetFocus()
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
	w.nextButton.SetFocus()

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
		// Target edit widget will be handled internally by Unicode picker
		w.logger.Debug("Unicode picker will handle target edit widget internally")

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
func (w *EnterTabWidget) configureInternationalInput(lineEdit *qt.QLineEdit) {
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
	*qt.QWidget

	lesson *lesson.Lesson
	logger *logging.Logger

	// UI components
	overviewLabel *qt.QLabel
	resultsTable  *qt.QTableWidget

	// Results data
	sessions []*TeachingSession
}

// NewResultsTabWidget creates a new Results tab widget
func NewResultsTabWidget(lesson *lesson.Lesson, parent *qt.QWidget) *ResultsTabWidget {
	widget := &ResultsTabWidget{
		QWidget: qt.NewQWidget(parent),
		lesson:  lesson,
		logger:  logging.NewLogger("ResultsTabWidget"),
	}

	widget.setupUI()
	return widget
}

// setupUI initializes the Results tab interface
func (w *ResultsTabWidget) setupUI() {
	layout := qt.NewQVBoxLayout(w.QWidget)

	// Overview section
	overviewGroup := qt.NewQGroupBox(w.QWidget)
	overviewGroup.SetTitle("Results Overview")
	overviewLayout := qt.NewQVBoxLayout(overviewGroup.QWidget)

	w.overviewLabel = qt.NewQLabel(w.QWidget)
	w.overviewLabel.SetText("No teaching results available yet")
	w.overviewLabel.SetAlignment(qt.AlignCenter)
	overviewLayout.AddWidget(w.overviewLabel.QWidget)

	layout.AddWidget(overviewGroup.QWidget)

	// Detailed results
	detailsGroup := qt.NewQGroupBox(w.QWidget)
	detailsGroup.SetTitle("Latest Session Results")
	detailsLayout := qt.NewQVBoxLayout(detailsGroup.QWidget)

	w.resultsTable = qt.NewQTableWidget2()
	w.resultsTable.SetRowCount(0)
	w.resultsTable.SetColumnCount(4)
	w.resultsTable.SetHorizontalHeaderLabels([]string{"Question", "Correct Answer", "Your Answer", "Result"})
	w.resultsTable.HorizontalHeader().SetStretchLastSection(true)
	w.resultsTable.SetAlternatingRowColors(true)
	detailsLayout.AddWidget(w.resultsTable.QWidget)

	layout.AddWidget(detailsGroup.QWidget)

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
		questionItem := qt.NewQTableWidgetItem2(result.Question)
		w.resultsTable.SetItem(i, 0, questionItem)

		// Correct Answer
		correctItem := qt.NewQTableWidgetItem2(result.CorrectAnswer)
		w.resultsTable.SetItem(i, 1, correctItem)

		// User Answer
		userItem := qt.NewQTableWidgetItem2(result.UserAnswer)
		w.resultsTable.SetItem(i, 2, userItem)

		// Result (CORRECT/WRONG)
		var resultText string
		var resultItem *qt.QTableWidgetItem
		if result.IsCorrect {
			resultText = "[CORRECT]"
			resultItem = qt.NewQTableWidgetItem2(resultText)
			color := qt.NewQColor()
			color.SetRgb(200, 255, 200)
			brush := qt.NewQBrush3(color)
			resultItem.SetBackground(brush) // Light green background
		} else {
			resultText = "[WRONG]"
			resultItem = qt.NewQTableWidgetItem2(resultText)
			color := qt.NewQColor()
			color.SetRgb(255, 200, 200)
			brush := qt.NewQBrush3(color)
			resultItem.SetBackground(brush) // Light red background
		}
		w.resultsTable.SetItem(i, 3, resultItem)
	}

	w.resultsTable.ResizeColumnsToContents()
}
