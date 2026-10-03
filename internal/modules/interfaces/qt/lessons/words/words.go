package words

import (
	"fmt"
	"strings"
	"time"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/logging"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/dialogs/results"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	percentscalculator "github.com/LaPingvino/recuerdo/internal/modules/logic/percentsCalculator"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/checker"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/composer"
	"github.com/LaPingvino/recuerdo/internal/teaching"
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

// Settings is what the lesson widget needs from the settings module.
type Settings interface {
	GetSettingWithDefault(key string, defaultValue interface{}) interface{}
	SetSetting(key string, value interface{}) error
}

// Setting keys, named as in OpenTeacher. The fade duration is in
// milliseconds.
const (
	NotationSetting       = "org.openteacher.noteCalculatorChooser.noteCalculator"
	RepeatDurationSetting = "org.openteacher.teachTypes.repeatAnswer.fadeDuration"
)

// UseSettings makes the lesson remember the grade notation and the Repeat
// answer duration in s, starting from the values already there.
func (w *WordsLessonWidget) UseSettings(s Settings) {
	if s == nil {
		return
	}
	w.teachWidget.useSettings(s)
	w.resultsWidget.useSettings(s)
}

// number converts a setting read from JSON (float64) or set in Go to a
// float64.
func number(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
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

	// The Teach tab shows grades in the notation chosen on the Results tab
	w.teachWidget.notation = w.resultsWidget.Notation

	// Connect teach widget to results widget for session completion
	w.teachWidget.SetSessionCompletedCallback(func(session *TeachingSession) {
		w.logger.Event("Teaching session completed - adding to results")
		w.resultsWidget.AddSession(session)
		// and show the results of this run, as OpenTeacher does
		notation := w.resultsWidget.Notation()
		results.Show(w.QWidget, session.Report, teaching.Grade(notation, session.Test), notation)
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
	// Test is the run as the lesson type recorded it, for grading
	Test lessontypes.Test
	// Report is what the results dialog shows about the run
	Report         teaching.Report
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

	// Practice options (OpenTeacher's lesson types and list modifiers)
	lessonTypeCombo *qt.QComboBox
	orderCombo      *qt.QComboBox
	askAnswersCheck *qt.QCheckBox
	session         *teaching.Session
	typing          *teaching.Typing
	modeCombo       *qt.QComboBox
	hintLabel       *qt.QLabel
	repeatTimer     *qt.QTimer
	// RepeatDuration is how long Repeat answer shows the answer
	RepeatDuration time.Duration
	repeatLabel    *qt.QLabel
	repeatSpin     *qt.QDoubleSpinBox
	settings       Settings
	skipButton     *qt.QPushButton
	correctButton  *qt.QPushButton
	// In mind
	viewButton, rightButton, wrongButton *qt.QPushButton
	// Hangman
	hangman      *teaching.HangmanWord
	hangmanLabel *qt.QLabel
	hangmanTimer *qt.QTimer
	// notation returns the grade notation to show; set by the lesson widget
	notation func() string

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

	// Practice options
	optionsGroup := qt.NewQGroupBox(w.QWidget)
	optionsGroup.SetTitle("Practice")
	optionsLayout := qt.NewQHBoxLayout(optionsGroup.QWidget)
	lessonTypeLabel := qt.NewQLabel(w.QWidget)
	lessonTypeLabel.SetText("Lesson type:")
	optionsLayout.AddWidget(lessonTypeLabel.QWidget)
	w.lessonTypeCombo = qt.NewQComboBox(w.QWidget)
	w.lessonTypeCombo.AddItems(teaching.LessonTypes)
	w.lessonTypeCombo.SetToolTip("All once: every word once. Smart: wrong words come back soon and at the end. Interval: words come back until you know them.")
	optionsLayout.AddWidget(w.lessonTypeCombo.QWidget)
	orderLabel := qt.NewQLabel(w.QWidget)
	orderLabel.SetText("Order:")
	optionsLayout.AddWidget(orderLabel.QWidget)
	w.orderCombo = qt.NewQComboBox(w.QWidget)
	w.orderCombo.AddItems(teaching.Orders)
	optionsLayout.AddWidget(w.orderCombo.QWidget)
	modeLabel := qt.NewQLabel(w.QWidget)
	modeLabel.SetText("Mode:")
	optionsLayout.AddWidget(modeLabel.QWidget)
	w.modeCombo = qt.NewQComboBox(w.QWidget)
	w.modeCombo.AddItems(teaching.TeachTypes)
	w.modeCombo.SetToolTip("Typing: type the answer. Shuffle answer: with the letters of the answer as a hint. Repeat answer: the answer is shown first, then typed from memory.")
	optionsLayout.AddWidget(w.modeCombo.QWidget)
	w.repeatLabel = qt.NewQLabel(w.QWidget)
	w.repeatLabel.SetText("Show answer for:")
	optionsLayout.AddWidget(w.repeatLabel.QWidget)
	w.repeatSpin = qt.NewQDoubleSpinBox(w.QWidget)
	w.repeatSpin.SetRange(0.5, 30)
	w.repeatSpin.SetSingleStep(0.5)
	w.repeatSpin.SetDecimals(1)
	w.repeatSpin.SetSuffix(" s")
	w.repeatSpin.SetValue(teaching.RepeatFadeDuration.Seconds())
	w.repeatSpin.SetToolTip("How long Repeat answer shows the answer before you type it")
	w.repeatSpin.OnValueChanged(w.setRepeatSeconds)
	optionsLayout.AddWidget(w.repeatSpin.QWidget)
	showRepeat := func() {
		on := w.modeCombo.CurrentText() == teaching.RepeatAnswer
		w.repeatLabel.SetVisible(on)
		w.repeatSpin.SetVisible(on)
	}
	w.modeCombo.OnCurrentTextChanged(func(string) { showRepeat() })
	showRepeat()
	w.askAnswersCheck = qt.NewQCheckBox(w.QWidget)
	w.askAnswersCheck.SetText("Ask the answers")
	w.askAnswersCheck.SetToolTip("Practise the other way round: the answers are asked and the questions are the answers")
	optionsLayout.AddWidget(w.askAnswersCheck.QWidget)
	optionsLayout.AddStretch()
	layout.AddWidget(optionsGroup.QWidget)

	// Question section
	questionGroup := qt.NewQGroupBox(w.QWidget)
	questionGroup.SetTitle("Current Question")
	questionLayout := qt.NewQVBoxLayout(questionGroup.QWidget)

	w.questionLabel = qt.NewQLabel(w.QWidget)
	w.questionLabel.SetText("Click 'Start Teaching' to begin")
	w.questionLabel.SetWordWrap(true)
	w.questionLabel.SetAlignment(qt.AlignCenter)
	questionLayout.AddWidget(w.questionLabel.QWidget)

	// Shuffle answer's hint / Repeat answer's answer
	w.hintLabel = qt.NewQLabel(w.QWidget)
	w.hintLabel.SetAlignment(qt.AlignCenter)
	w.hintLabel.SetVisible(false)
	questionLayout.AddWidget(w.hintLabel.QWidget)
	w.RepeatDuration = teaching.RepeatFadeDuration
	w.repeatTimer = qt.NewQTimer()
	w.repeatTimer.SetSingleShot(true)
	w.repeatTimer.OnTimeout(w.repeatShown)

	w.hangmanLabel = qt.NewQLabel(w.QWidget)
	w.hangmanLabel.SetAlignment(qt.AlignCenter)
	w.hangmanLabel.SetVisible(false)
	questionLayout.AddWidget(w.hangmanLabel.QWidget)
	w.hangmanTimer = qt.NewQTimer()
	w.hangmanTimer.SetSingleShot(true)
	w.hangmanTimer.OnTimeout(w.hangmanLost)

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
	w.submitButton.SetText("Check")
	w.submitButton.SetEnabled(false)
	w.nextButton = qt.NewQPushButton(w.QWidget)
	w.nextButton.SetText("Continue")
	w.nextButton.SetToolTip("Go on after seeing the right answer")
	w.nextButton.SetEnabled(false)
	w.skipButton = qt.NewQPushButton(w.QWidget)
	w.skipButton.SetText("Skip")
	w.skipButton.SetToolTip("Ask this word again later")
	w.skipButton.SetEnabled(false)
	w.correctButton = qt.NewQPushButton(w.QWidget)
	w.correctButton.SetText("Correct anyway")
	w.correctButton.SetToolTip("Count your last answer as right after all")
	w.correctButton.SetEnabled(false)

	buttonLayout.AddWidget(w.startButton.QWidget)
	buttonLayout.AddWidget(w.submitButton.QWidget)
	buttonLayout.AddWidget(w.nextButton.QWidget)
	buttonLayout.AddWidget(w.skipButton.QWidget)
	buttonLayout.AddWidget(w.correctButton.QWidget)
	w.viewButton = qt.NewQPushButton(w.QWidget)
	w.viewButton.SetText("View answer")
	w.rightButton = qt.NewQPushButton(w.QWidget)
	w.rightButton.SetText("I was right")
	w.wrongButton = qt.NewQPushButton(w.QWidget)
	w.wrongButton.SetText("I was wrong")
	for _, b := range []*qt.QPushButton{w.viewButton, w.rightButton, w.wrongButton} {
		b.SetVisible(false)
		buttonLayout.AddWidget(b.QWidget)
	}
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

	w.skipButton.OnClicked(func() {
		switch {
		case w.typing != nil:
			w.report(w.typing.Skip())
		case w.session != nil && !w.session.Done():
			w.session.Skip()
			w.inMindNext()
		}
	})

	w.viewButton.OnClicked(w.inMindView)
	w.rightButton.OnClicked(func() { w.inMindJudge(true) })
	w.wrongButton.OnClicked(func() { w.inMindJudge(false) })

	w.correctButton.OnClicked(func() {
		if w.typing != nil {
			w.report(w.typing.CorrectAnyway())
			w.feedback("Your last answer counts as right.", true)
		}
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

	w.session = teaching.New(w.lesson.Data.List, teaching.Options{
		LessonType: w.lessonTypeCombo.CurrentText(),
		Order:      w.orderCombo.CurrentText(),
		AskAnswers: w.askAnswersCheck.IsChecked(),
	})
	w.isTeaching = true
	w.correctAnswers = 0
	_, w.totalQuestions = w.session.Progress()

	// Initialize new teaching session
	w.currentSession = &TeachingSession{
		Results:        make([]TeachingResult, 0, w.totalQuestions),
		TotalQuestions: w.totalQuestions,
		CorrectCount:   0,
		Score:          0,
		Completed:      false,
	}

	w.setOptionsEnabled(false)
	w.startButton.SetEnabled(false)
	w.unicodeButton.SetEnabled(true)
	w.resultLabel.SetVisible(false)

	w.typing = nil
	w.hangman = nil
	if w.modeCombo.CurrentText() == teaching.Hangman {
		w.setHangmanLayout(true)
		w.session.Start()
		w.hangmanNext()
		return
	}
	if w.modeCombo.CurrentText() == teaching.InMind {
		// no typing: think, look, and say whether you knew it
		w.setInMindLayout(true)
		w.session.Start()
		w.inMindNext()
		return
	}

	// OpenTeacher's typing mode drives the tab from here (TypingUI below)
	w.typing = teaching.NewTyping(w.session, w)
	w.logger.Action("Started teaching session with %d words (%s, %s)", w.totalQuestions,
		w.lessonTypeCombo.CurrentText(), w.orderCombo.CurrentText())
}

func (w *TeachTabWidget) setOptionsEnabled(enabled bool) {
	w.lessonTypeCombo.SetEnabled(enabled)
	w.orderCombo.SetEnabled(enabled)
	w.modeCombo.SetEnabled(enabled)
	w.askAnswersCheck.SetEnabled(enabled)
}

// showCurrentQuestion displays the current question
func (w *TeachTabWidget) showCurrentQuestion() {
	if w.session == nil || w.session.Done() {
		w.finishTeaching()
		return
	}
	item, index, _ := w.session.Current()
	w.currentIndex = index

	question := composer.Compose(checker.StoredAnswers(item.Questions))
	w.questionLabel.SetText(fmt.Sprintf("Question: %s", question))
	w.showModeExtras()
	w.answerEdit.Clear()
	w.answerEdit.SetFocus()

	// Update progress (lesson types that repeat words add to the total)
	asked, total := w.session.Progress()
	w.totalQuestions = total
	if total > 0 {
		w.progressBar.SetValue(asked * 100 / total)
	}
	w.correctAnswers, _ = w.session.Score()
	w.statusLabel.SetText(fmt.Sprintf("Question %d of %d (Score: %d/%d correct)",
		asked+1, total, w.correctAnswers, asked))
}

// showModeExtras shows what the practice mode adds to a new question.
func (w *TeachTabWidget) showModeExtras() {
	w.repeatTimer.Stop()
	switch w.modeCombo.CurrentText() {
	case teaching.ShuffleAnswer:
		w.hintLabel.SetStyleSheet("")
		w.hintLabel.SetText(teaching.ShuffleHint(w.session.CurrentAnswer(), nil))
		w.hintLabel.SetVisible(true)
	case teaching.InMind:
		w.hintLabel.SetStyleSheet("")
		w.hintLabel.SetText("Think about the answer, and press 'View answer' when you're done.")
		w.hintLabel.SetVisible(true)
	case teaching.RepeatAnswer:
		// show the answer first; typing starts when it is gone
		w.hintLabel.SetStyleSheet("font-size: 20px; font-weight: bold;")
		w.hintLabel.SetText(w.session.CurrentAnswer())
		w.hintLabel.SetVisible(true)
		w.SetInputEnabled(false)
		w.SetCheckEnabled(false)
		w.SetSkipEnabled(false)
		w.repeatTimer.Start(int(w.RepeatDuration.Milliseconds()))
	default:
		w.hintLabel.SetVisible(false)
	}
}

// setInMindLayout swaps the answer field and typing buttons for In mind's.
func (w *TeachTabWidget) setInMindLayout(on bool) {
	w.answerEdit.SetVisible(!on)
	w.unicodeButton.SetVisible(!on)
	w.submitButton.SetVisible(!on)
	w.nextButton.SetVisible(!on)
	w.correctButton.SetVisible(!on)
	w.viewButton.SetVisible(on)
	w.rightButton.SetVisible(false)
	w.wrongButton.SetVisible(false)
}

// inMindNext asks the next question in In mind, or ends the session.
func (w *TeachTabWidget) inMindNext() {
	if w.session.Done() {
		w.finishTeaching()
		return
	}
	w.showCurrentQuestion()
	w.viewButton.SetVisible(true)
	w.viewButton.SetEnabled(true)
	w.viewButton.SetFocus()
	w.skipButton.SetEnabled(true)
	w.rightButton.SetVisible(false)
	w.wrongButton.SetVisible(false)
}

func (w *TeachTabWidget) inMindView() {
	if w.session == nil || w.session.Done() {
		return
	}
	w.hintLabel.SetText("Translation: " + w.session.ViewAnswer())
	w.viewButton.SetVisible(false)
	w.skipButton.SetEnabled(false)
	w.rightButton.SetVisible(true)
	w.wrongButton.SetVisible(true)
	w.rightButton.SetFocus()
}

func (w *TeachTabWidget) inMindJudge(right bool) {
	if w.session == nil || w.session.Done() {
		return
	}
	w.session.Judge(right)
	w.inMindNext()
}

// setRepeatSeconds sets how long Repeat answer shows the answer, and
// remembers it.
func (w *TeachTabWidget) setRepeatSeconds(seconds float64) {
	w.RepeatDuration = time.Duration(seconds * float64(time.Second))
	if w.settings != nil {
		w.settings.SetSetting(RepeatDurationSetting, w.RepeatDuration.Milliseconds())
	}
}

// useSettings loads the remembered Repeat answer duration and keeps it
// up to date in s.
func (w *TeachTabWidget) useSettings(s Settings) {
	if ms, ok := number(s.GetSettingWithDefault(RepeatDurationSetting, nil)); ok && ms > 0 {
		w.repeatSpin.SetValue(ms / 1000) // calls setRepeatSeconds
		w.RepeatDuration = time.Duration(ms) * time.Millisecond
	}
	w.settings = s
}

// repeatShown ends Repeat answer's showing of the answer.
func (w *TeachTabWidget) repeatShown() {
	if w.typing == nil || w.session == nil || w.session.Done() || w.typing.ShowingCorrection() {
		return
	}
	w.hintLabel.SetVisible(false)
	w.SetInputEnabled(true)
	w.SetCheckEnabled(true)
	w.SetSkipEnabled(true)
	w.FocusInput()
}

// submitAnswer checks the typed answer (Check button or Enter)
func (w *TeachTabWidget) submitAnswer() {
	if w.hangman != nil {
		w.hangmanGuess(w.answerEdit.Text())
		return
	}
	if w.typing == nil {
		return
	}
	userAnswer := strings.TrimSpace(w.answerEdit.Text())
	if userAnswer == "" {
		return
	}
	w.report(w.typing.Check(userAnswer))
	if !w.typing.ShowingCorrection() {
		w.feedback("Right!", true)
	}
	w.logger.Info("Answer checked: %s", userAnswer)
}

// nextQuestion continues after a correction (Continue button or Enter)
func (w *TeachTabWidget) nextQuestion() {
	if w.typing != nil {
		w.report(w.typing.CorrectionShowingDone())
	}
}

func (w *TeachTabWidget) report(err error) {
	if err != nil {
		w.logger.Debug("typing: %v", err)
	}
}

func (w *TeachTabWidget) feedback(text string, right bool) {
	style := "color: green; font-weight: bold; background-color: lightgreen; padding: 5px; border-radius: 3px;"
	if !right {
		style = "color: red; font-weight: bold; background-color: lightcoral; padding: 5px; border-radius: 3px;"
	}
	w.resultLabel.SetText(text)
	w.resultLabel.SetStyleSheet(style)
	w.resultLabel.SetVisible(true)
}

// TypingUI, called by the typing controller.

func (w *TeachTabWidget) ClearInput()                     { w.showCurrentQuestion() }
func (w *TeachTabWidget) FocusInput()                     { w.answerEdit.SetFocus() }
func (w *TeachTabWidget) SetInputEnabled(on bool)         { w.answerEdit.SetEnabled(on) }
func (w *TeachTabWidget) SetCheckEnabled(on bool)         { w.submitButton.SetEnabled(on) }
func (w *TeachTabWidget) SetSkipEnabled(on bool)          { w.skipButton.SetEnabled(on) }
func (w *TeachTabWidget) SetCorrectAnywayEnabled(on bool) { w.correctButton.SetEnabled(on) }
func (w *TeachTabWidget) LessonDone()                     { w.finishTeaching() }

func (w *TeachTabWidget) ShowCorrection(answer string) {
	w.feedback(fmt.Sprintf("Wrong. The right answer is: %s", answer), false)
	w.nextButton.SetEnabled(true)
	w.nextButton.SetFocus()
}

func (w *TeachTabWidget) HideCorrection() {
	w.resultLabel.SetVisible(false)
	w.nextButton.SetEnabled(false)
}

// finishTeaching completes the teaching session
func (w *TeachTabWidget) finishTeaching() {
	w.isTeaching = false
	w.repeatTimer.Stop()
	w.hintLabel.SetVisible(false)
	w.setInMindLayout(false)
	w.hangmanTimer.Stop()
	w.setHangmanLayout(false)
	w.skipButton.SetEnabled(false)
	if w.session != nil {
		w.correctAnswers, w.totalQuestions = w.session.Score()
	}
	percentage := 0
	if w.totalQuestions > 0 {
		percentage = int((float64(w.correctAnswers) / float64(w.totalQuestions)) * 100)
	}

	notation := teaching.DefaultNotation
	if w.notation != nil {
		notation = w.notation()
	}
	grade := ""

	// Complete the session record
	if w.currentSession != nil {
		if w.session != nil {
			w.currentSession.Test = w.session.Test()
			w.currentSession.Report = w.session.Report()
			w.currentSession.Results = w.currentSession.Results[:0]
			for i, row := range w.currentSession.Report.Rows {
				w.currentSession.Results = append(w.currentSession.Results, TeachingResult{
					Question:      row.Question,
					CorrectAnswer: row.Answer,
					UserAnswer:    row.Given,
					IsCorrect:     row.Right,
					ItemIndex:     w.currentSession.Test.Results[i].ItemID,
				})
			}
			percentage = percentscalculator.Percents(w.currentSession.Test)
			grade = teaching.Grade(notation, w.currentSession.Test)
		}
		w.currentSession.Score = percentage
		w.currentSession.CorrectCount = w.correctAnswers
		w.currentSession.TotalQuestions = w.totalQuestions
		w.currentSession.Completed = true
	}
	w.setOptionsEnabled(true)

	text := fmt.Sprintf("Teaching completed! Final Score: %d/%d correct (%d%%)",
		w.correctAnswers, w.totalQuestions, percentage)
	if grade != "" && notation != "Percents" {
		text += fmt.Sprintf("\nGrade (%s): %s", notation, grade)
	}
	w.questionLabel.SetText(text)

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
	w.repeatTimer.Stop()
	w.hangmanTimer.Stop()
	w.hangman = nil
	w.setInMindLayout(false)
	w.setHangmanLayout(false)
	w.hintLabel.SetVisible(false)
	w.isTeaching = false
	w.session = nil
	w.typing = nil
	w.skipButton.SetEnabled(false)
	w.correctButton.SetEnabled(false)
	w.setOptionsEnabled(true)
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
	notationCombo *qt.QComboBox
	settings      Settings
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

	notationLayout := qt.NewQHBoxLayout2()
	notationLabel := qt.NewQLabel(w.QWidget)
	notationLabel.SetText("Grades in:")
	notationLayout.AddWidget(notationLabel.QWidget)
	w.notationCombo = qt.NewQComboBox(w.QWidget)
	w.notationCombo.AddItems(teaching.Notations)
	w.notationCombo.SetCurrentText(teaching.DefaultNotation)
	w.notationCombo.OnCurrentTextChanged(func(name string) {
		if w.settings != nil {
			w.settings.SetSetting(NotationSetting, name)
		}
		w.updateResultsDisplay()
	})
	notationLayout.AddWidget(w.notationCombo.QWidget)
	notationLayout.AddStretch()
	overviewLayout.AddLayout2(notationLayout.QLayout, 0)

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

// useSettings shows grades in the remembered notation and remembers the
// one the user picks.
func (w *ResultsTabWidget) useSettings(s Settings) {
	if name, ok := s.GetSettingWithDefault(NotationSetting, nil).(string); ok {
		for _, n := range teaching.Notations {
			if n == name {
				w.notationCombo.SetCurrentText(name)
			}
		}
	}
	w.settings = s
}

// Notation is the grade notation chosen on the Results tab.
func (w *ResultsTabWidget) Notation() string {
	if w.notationCombo == nil {
		return teaching.DefaultNotation
	}
	return w.notationCombo.CurrentText()
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

		// Grades as OpenTeacher computes them, in the chosen notation
		notation := w.Notation()
		tests := make([]lessontypes.Test, 0, totalSessions)
		for _, session := range w.sessions {
			tests = append(tests, session.Test)
		}

		overviewText := fmt.Sprintf(`Lesson: %d word pairs | Sessions completed: %d

Latest Session: %d/%d correct (%d%%), grade %s
Average grade: %s (%d%%)

Detailed results from latest session:`,
			wordCount, totalSessions,
			latestSession.CorrectCount, latestSession.TotalQuestions, latestSession.Score,
			teaching.Grade(notation, latestSession.Test),
			teaching.AverageGrade(notation, tests), percentscalculator.AveragePercents(tests))

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
