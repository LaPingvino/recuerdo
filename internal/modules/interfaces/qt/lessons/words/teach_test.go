package words

import (
	"fmt"
	"github.com/LaPingvino/recuerdo/internal/spellcheck"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	qt "github.com/mappu/miqt/qt6"
)

func init() { runtime.LockOSThread() }

// Qt widgets need a QApplication on the main thread, which is where
// TestMain runs, so the GUI checks run here and are reported by tests.
func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"words-test"})
	teachErr = checkTeachTab()
	modesErr = checkModes()
	inMindErr = checkInMind()
	hangmanErr = checkHangman()
	settingsErr = checkSettings()
	editErr = checkEditing()
	wordsErr = checkWordChoice()
	storedErr = checkSessionStored()
	pronounceErr = checkPronounce()
	formulaErr = checkFormulas()
	spellErr = checkSpellingMarks()
	os.Exit(m.Run())
}

var pronounceErr, formulaErr, spellErr error

func TestSpellingMarks(t *testing.T) {
	if spellErr == errNoHunspell {
		t.Skip("hunspell with nl and en dictionaries is not installed")
	}
	if spellErr != nil {
		t.Fatal(spellErr)
	}
}

var errNoHunspell = fmt.Errorf("no hunspell")

// Words not in the dictionary of their column's language are marked.
func checkSpellingMarks() error {
	if spellcheck.New("Dutch") == nil || spellcheck.New("English") == nil {
		return errNoHunspell
	}
	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{
		QuestionLanguage: "Dutch", AnswerLanguage: "English", Items: []lesson.WordItem{
			{ID: 0, Questions: []string{"hond"}, Answers: []string{"dog"}},
			{ID: 1, Questions: []string{"huiss"}, Answers: []string{"house"}},
			{ID: 2, Questions: []string{"fiets"}, Answers: []string{"bicylce"}},
		}}}}
	w := NewWordsLessonWidget(l, nil)
	e := w.enterWidget
	e.checkSpelling(0, 1)
	marked := func(r, c int) bool { return e.wordsTable.Item(r, c).Font().Underline() }
	if marked(0, 0) || marked(0, 1) || !marked(1, 0) || marked(1, 1) || marked(2, 0) || !marked(2, 1) {
		return fmt.Errorf("marks wrong")
	}
	if tip := e.wordsTable.Item(2, 1).ToolTip(); !strings.Contains(tip, "bicylce: bicycle") {
		return fmt.Errorf("tooltip %q", tip)
	}
	e.wordsTable.Item(1, 0).SetText("huis") // fixed: no longer marked
	if marked(1, 0) || l.Data.List.Items[1].Questions[0] != "huis" {
		return fmt.Errorf("after fixing: %v %q", marked(1, 0), l.Data.List.Items[1].Questions)
	}
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		e.wordsTable.Item(1, 0).SetText("huiss")
		w.Resize(900, 400)
		w.Show()
		qt.QCoreApplication_ProcessEvents()
		w.Grab().Save(dir + "/spelling.png")
	}
	e.noSpelling = true // the setting off: no marks
	e.checkSpelling(0, 1)
	if marked(2, 1) {
		return fmt.Errorf("marked with the setting off")
	}
	return nil
}

func TestFormulasInTeach(t *testing.T) {
	if formulaErr != nil {
		t.Fatal(formulaErr)
	}
}

// Words with formulas and markup are shown rendered and checked by their
// plain form (teaching.Correct).
func checkFormulas() error {
	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{Items: []lesson.WordItem{
		{ID: 1, Questions: []string{"$x^2 + 1$ when $x = 3$"}, Answers: []string{"10"}},
		{ID: 2, Questions: []string{"the area of a circle"}, Answers: []string{"$\\pi r^2$"}},
	}}}}
	w := NewWordsLessonWidget(l, nil)
	t := w.teachWidget
	w.Resize(900, 600)
	w.tabWidget.SetCurrentIndex(1) // Teach
	t.startButton.Click()
	if q := t.questionLabel.Text(); q != "x<sup>2</sup> + 1 when x = 3" {
		return fmt.Errorf("question shown as %q", q)
	}
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		w.Show()
		qt.QCoreApplication_ProcessEvents()
		w.Grab().Save(dir + "/formula-question.png")
	}
	t.answerEdit.SetText("10")
	t.submitButton.Click()
	t.answerEdit.SetText("pi r^2") // wrong: \pi
	t.submitButton.Click()
	if r := t.resultLabel.Text(); !strings.Contains(r, "πr<sup>2</sup>") {
		return fmt.Errorf("correction shown as %q", r)
	}
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		qt.QCoreApplication_ProcessEvents()
		w.Grab().Save(dir + "/formula-correction.png")
	}
	// the second answer is a formula: the builder is there
	if !t.formulaPad.IsVisibleTo(t.QWidget) {
		return fmt.Errorf("no formula builder for a formula answer")
	}
	t.nextButton.Click()
	if !t.formulaPad.IsHidden() {
		return fmt.Errorf("formula builder left after the lesson")
	}

	// the Enter tab: the formula builder edits the current cell
	e := w.enterWidget
	e.wordsTable.SetCurrentCell(0, 1)
	if !e.formulaButton.IsEnabled() {
		return fmt.Errorf("formula button off on an answer cell")
	}
	e.editWithFormulaPad(func(_ *qt.QWidget, old string) (string, bool) { return "$" + old + "^2$", true })
	if got := l.Data.List.Items[0].Answers; len(got) != 1 || got[0] != "$10^2$" {
		return fmt.Errorf("edited answer: %q", got)
	}
	e.wordsTable.SetCurrentCell(0, 2)
	if e.formulaButton.IsEnabled() {
		return fmt.Errorf("formula button on for a comment cell")
	}
	return nil
}

func TestPronounceQuestions(t *testing.T) {
	if pronounceErr != nil {
		t.Fatal(pronounceErr)
	}
}

// fakeSpeaker records what is said.
type fakeSpeaker struct{ said []string }

func (f *fakeSpeaker) Available() bool { return true }
func (f *fakeSpeaker) Speak(text, language string) error {
	f.said = append(f.said, language+":"+text)
	return nil
}

func checkPronounce() error {
	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{
		QuestionLanguage: "Dutch", AnswerLanguage: "English",
		Items: []lesson.WordItem{{ID: 1, Questions: []string{"hond"}, Answers: []string{"dog"}}},
	}}}
	w := NewWordsLessonWidget(l, nil)
	t := w.teachWidget
	speaker := &fakeSpeaker{}
	t.speaker = speaker
	t.startButton.Click()
	t.finishTeaching()
	if len(speaker.said) != 0 {
		return fmt.Errorf("spoke without being asked to: %v", speaker.said)
	}
	t.pronounceCheck.SetEnabled(true)
	t.pronounceCheck.SetChecked(true)
	t.startButton.Click()
	t.finishTeaching()
	t.askAnswersCheck.SetChecked(true)
	t.startButton.Click()
	if fmt.Sprint(speaker.said) != "[Dutch:hond English:dog]" {
		return fmt.Errorf("said %v, want the question in its language both ways", speaker.said)
	}
	return nil
}

var teachErr, modesErr, inMindErr, hangmanErr, settingsErr, editErr, wordsErr, storedErr error

func TestFinishedSessionIsStored(t *testing.T) {
	if storedErr != nil {
		t.Fatal(storedErr)
	}
}

func TestTeachTabWordChoice(t *testing.T) {
	if wordsErr != nil {
		t.Fatal(wordsErr)
	}
}

func TestEnterTabEditsTheLesson(t *testing.T) {
	if editErr != nil {
		t.Fatal(editErr)
	}
}

func TestLessonSettings(t *testing.T) {
	if settingsErr != nil {
		t.Fatal(settingsErr)
	}
}

func TestTeachTabHangman(t *testing.T) {
	if hangmanErr != nil {
		t.Fatal(hangmanErr)
	}
}

func TestTeachTabInMind(t *testing.T) {
	if inMindErr != nil {
		t.Fatal(inMindErr)
	}
}

func TestTeachTabModes(t *testing.T) {
	if modesErr != nil {
		t.Fatal(modesErr)
	}
}

func TestTeachTabTypingMode(t *testing.T) {
	if teachErr != nil {
		t.Fatal(teachErr)
	}
}

func checkTeachTab() error {
	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{Items: []lesson.WordItem{
		{ID: 0, Questions: []string{"een"}, Answers: []string{"one"}},
		{ID: 1, Questions: []string{"twee"}, Answers: []string{"two"}},
		{ID: 2, Questions: []string{"drie"}, Answers: []string{"three"}},
	}}}}
	w := NewTeachTabWidget(l, nil)
	var finished *TeachingSession
	w.SetSessionCompletedCallback(func(s *TeachingSession) { finished = s })

	answer := func(text string) {
		w.answerEdit.SetText(text)
		w.submitButton.Click()
	}
	question := func() string { return w.questionLabel.Text() }

	w.startButton.Click()
	if !strings.Contains(question(), "een") || !w.submitButton.IsEnabled() || w.nextButton.IsEnabled() {
		return fmt.Errorf("start: question %q", question())
	}

	// right: moves on at once
	answer("one")
	if !strings.Contains(question(), "twee") || w.answerEdit.Text() != "" {
		return fmt.Errorf("after a right answer: question %q, input %q", question(), w.answerEdit.Text())
	}

	// wrong: correction shown, input locked, Continue and Correct anyway possible
	answer("deux")
	if !strings.Contains(w.resultLabel.Text(), "two") || w.answerEdit.IsEnabled() ||
		w.submitButton.IsEnabled() || !w.nextButton.IsEnabled() || !w.correctButton.IsEnabled() {
		return fmt.Errorf("after a wrong answer: %q", w.resultLabel.Text())
	}
	w.correctButton.Click() // "I was right"
	if !strings.Contains(question(), "drie") || !w.answerEdit.IsEnabled() || w.correctButton.IsEnabled() {
		return fmt.Errorf("after correct anyway: question %q", question())
	}

	// skip, then the lesson asks drie again at the end
	w.skipButton.Click()
	if !strings.Contains(question(), "drie") {
		return fmt.Errorf("after skip: question %q", question())
	}
	answer("three")

	if finished == nil || !finished.Completed {
		return fmt.Errorf("session not completed")
	}
	if finished.CorrectCount != 3 || finished.TotalQuestions != 3 || len(finished.Results) != 3 {
		return fmt.Errorf("session %d/%d, %d results", finished.CorrectCount, finished.TotalQuestions, len(finished.Results))
	}
	if r := finished.Results[1]; !r.IsCorrect || r.UserAnswer != "Corrected: deux" {
		return fmt.Errorf("corrected result %+v", r)
	}
	if !w.startButton.IsEnabled() || w.skipButton.IsEnabled() {
		return fmt.Errorf("buttons after the end")
	}
	return nil
}

func checkModes() error {
	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{Items: []lesson.WordItem{
		{ID: 0, Questions: []string{"drie"}, Answers: []string{"three"}},
		{ID: 1, Questions: []string{"vier"}, Answers: []string{"four"}},
	}}}}

	// Shuffle answer: a hint with the answer's letters
	w := NewTeachTabWidget(l, nil)
	w.modeCombo.SetCurrentText("Shuffle answer")
	w.startButton.Click()
	hint := w.hintLabel.Text()
	if !w.hintLabel.IsVisibleTo(w.QWidget) || !strings.HasPrefix(hint, "Hint: ") || len(hint) != len("Hint: three") || hint == "Hint: three" {
		return fmt.Errorf("shuffle hint %q", hint)
	}

	// Repeat answer: the answer first, input locked; then typing from memory
	w = NewTeachTabWidget(l, nil)
	w.modeCombo.SetCurrentText("Repeat answer")
	w.startButton.Click()
	if w.hintLabel.Text() != "three" || w.answerEdit.IsEnabled() || w.submitButton.IsEnabled() {
		return fmt.Errorf("repeat: showing %q, input enabled %v", w.hintLabel.Text(), w.answerEdit.IsEnabled())
	}
	// the answer fades out evenly over the duration
	if o := w.hintOpacity.Opacity(); o != 1 {
		return fmt.Errorf("repeat: opacity %v at the start", o)
	}
	w.fadeStart = time.Now().Add(-w.RepeatDuration / 2)
	w.fadeStep()
	if o := w.hintOpacity.Opacity(); o < 0.4 || o > 0.6 {
		return fmt.Errorf("repeat: opacity %v half-way", o)
	}
	w.fadeStart = time.Now().Add(-2 * w.RepeatDuration)
	w.fadeStep()
	if o := w.hintOpacity.Opacity(); o != 0 || w.fadeTimer.IsActive() {
		return fmt.Errorf("repeat: opacity %v at the end, fading %v", o, w.fadeTimer.IsActive())
	}
	w.repeatShown() // what the timer does when the answer has been shown long enough
	if o := w.hintOpacity.Opacity(); o != 1 {
		return fmt.Errorf("repeat: opacity %v after showing (hint label is reused)", o)
	}
	if w.hintLabel.IsVisibleTo(w.QWidget) || !w.answerEdit.IsEnabled() || !w.submitButton.IsEnabled() {
		return fmt.Errorf("repeat: after showing, input enabled %v", w.answerEdit.IsEnabled())
	}
	w.answerEdit.SetText("three")
	w.submitButton.Click()
	if w.hintLabel.Text() != "four" || w.answerEdit.IsEnabled() {
		return fmt.Errorf("repeat: next word shows %q", w.hintLabel.Text())
	}
	if w.modeCombo.IsEnabled() {
		return fmt.Errorf("mode can be changed during a session")
	}
	return nil
}

func checkInMind() error {
	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{Items: []lesson.WordItem{
		{ID: 0, Questions: []string{"een"}, Answers: []string{"one"}},
		{ID: 1, Questions: []string{"twee"}, Answers: []string{"two"}},
	}}}}
	w := NewTeachTabWidget(l, nil)
	var finished *TeachingSession
	w.SetSessionCompletedCallback(func(s *TeachingSession) { finished = s })
	visible := func(b *qt.QPushButton) bool { return b.IsVisibleTo(w.QWidget) }

	w.modeCombo.SetCurrentText("In mind")
	w.startButton.Click()
	if w.answerEdit.IsVisibleTo(w.QWidget) || !visible(w.viewButton) || visible(w.rightButton) || !strings.Contains(w.questionLabel.Text(), "een") {
		return fmt.Errorf("in mind start: question %q", w.questionLabel.Text())
	}
	w.viewButton.Click()
	if w.hintLabel.Text() != "Translation: one" || !visible(w.rightButton) || !visible(w.wrongButton) || visible(w.viewButton) {
		return fmt.Errorf("after view answer: %q", w.hintLabel.Text())
	}
	w.rightButton.Click()
	if !strings.Contains(w.questionLabel.Text(), "twee") || visible(w.rightButton) {
		return fmt.Errorf("after I was right: %q", w.questionLabel.Text())
	}
	w.viewButton.Click()
	w.wrongButton.Click()

	if finished == nil || finished.CorrectCount != 1 || finished.TotalQuestions != 2 {
		return fmt.Errorf("finished session %+v", finished)
	}
	if !w.answerEdit.IsVisibleTo(w.QWidget) || visible(w.viewButton) {
		return fmt.Errorf("typing layout not restored after In mind")
	}
	return nil
}

func checkHangman() error {
	// the head's outline: left edge of the circle at (178, 40)-(202, 64)
	headPixel := func(pm *qt.QPixmap) uint { return pm.ToImage().Pixel(178, 52) & 0xffffff }
	if headPixel(drawHangman(0)) != 0xffffff || headPixel(drawHangman(1)) == 0xffffff {
		return fmt.Errorf("drawing: head pixel %x without mistakes, %x after one", headPixel(drawHangman(0)), headPixel(drawHangman(1)))
	}

	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{Items: []lesson.WordItem{
		{ID: 0, Questions: []string{"kat"}, Answers: []string{"cat"}},
		{ID: 1, Questions: []string{"hond"}, Answers: []string{"dog"}},
	}}}}
	w := NewTeachTabWidget(l, nil)
	var finished *TeachingSession
	w.SetSessionCompletedCallback(func(s *TeachingSession) { finished = s })
	guess := func(g string) { w.answerEdit.SetText(g); w.submitButton.Click() }

	w.modeCombo.SetCurrentText("Hangman")
	w.startButton.Click()
	if w.hintLabel.Text() != "---" || !w.hangmanLabel.IsVisibleTo(w.QWidget) || w.skipButton.IsVisibleTo(w.QWidget) {
		return fmt.Errorf("hangman start: %q", w.hintLabel.Text())
	}
	guess("a")
	if w.hintLabel.Text() != "-a-" {
		return fmt.Errorf("after a: %q", w.hintLabel.Text())
	}
	guess("a")
	if !strings.Contains(w.resultLabel.Text(), "already tried") {
		return fmt.Errorf("repeat guess: %q", w.resultLabel.Text())
	}
	guess("c")
	guess("t") // won: next word
	if !strings.Contains(w.questionLabel.Text(), "hond") || w.hintLabel.Text() != "---" {
		return fmt.Errorf("after winning: %q %q", w.questionLabel.Text(), w.hintLabel.Text())
	}

	for _, g := range []string{"x", "y", "z", "q", "w"} {
		guess(g)
	}
	if !strings.Contains(w.resultLabel.Text(), "x  |  y") {
		return fmt.Errorf("mistakes: %q", w.resultLabel.Text())
	}
	guess("v") // sixth mistake
	if !strings.Contains(w.resultLabel.Text(), "the answer was: dog") || w.answerEdit.IsEnabled() {
		return fmt.Errorf("lost: %q", w.resultLabel.Text())
	}
	w.hangmanLost() // what the timer does

	if finished == nil || finished.CorrectCount != 1 || finished.TotalQuestions != 2 {
		return fmt.Errorf("finished %+v", finished)
	}
	if r := finished.Results[1]; r.IsCorrect || r.UserAnswer != "hanged man" {
		return fmt.Errorf("lost round result %+v", r)
	}
	if w.hangmanLabel.IsVisibleTo(w.QWidget) || !w.skipButton.IsVisibleTo(w.QWidget) {
		return fmt.Errorf("layout not restored after hangman")
	}
	return nil
}

// mapSettings stands in for the settings module; values read back from
// its JSON file are float64, as here.
type mapSettings map[string]interface{}

func (m mapSettings) GetSettingWithDefault(key string, def interface{}) interface{} {
	if v, ok := m[key]; ok {
		return v
	}
	return def
}

func (m mapSettings) SetSetting(key string, value interface{}) error {
	m[key] = value
	return nil
}

func checkSettings() error {
	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{Items: []lesson.WordItem{
		{ID: 0, Questions: []string{"een"}, Answers: []string{"one"}},
	}}}}

	// remembered values are used
	s := mapSettings{NotationSetting: "American", RepeatDurationSetting: float64(1500)}
	w := NewWordsLessonWidget(l, nil)
	w.UseSettings(s)
	if got := w.resultsWidget.Notation(); got != "American" {
		return fmt.Errorf("notation %q, want American", got)
	}
	if got := w.teachWidget.RepeatDuration; got.Milliseconds() != 1500 {
		return fmt.Errorf("repeat duration %v, want 1.5s", got)
	}

	// the duration field only shows for Repeat answer
	t := w.teachWidget
	t.modeCombo.SetCurrentText("Typing")
	if t.repeatSpin.IsVisibleTo(t.QWidget) {
		return fmt.Errorf("duration shown in Typing mode")
	}
	t.modeCombo.SetCurrentText("Repeat answer")
	if !t.repeatSpin.IsVisibleTo(t.QWidget) {
		return fmt.Errorf("duration hidden in Repeat answer mode")
	}

	// changes are remembered
	w.resultsWidget.notationCombo.SetCurrentText("Percents")
	t.repeatSpin.SetValue(5)
	if s[NotationSetting] != "Percents" {
		return fmt.Errorf("notation saved as %v", s[NotationSetting])
	}
	if ms, _ := number(s[RepeatDurationSetting]); ms != 5000 {
		return fmt.Errorf("duration saved as %v", s[RepeatDurationSetting])
	}

	// unknown or missing values keep the defaults
	w = NewWordsLessonWidget(l, nil)
	w.UseSettings(mapSettings{NotationSetting: "Klingon"})
	if w.resultsWidget.Notation() == "Klingon" || w.teachWidget.RepeatDuration <= 0 {
		return fmt.Errorf("bad settings: notation %q, duration %v", w.resultsWidget.Notation(), w.teachWidget.RepeatDuration)
	}
	return nil
}

func checkEditing() error {
	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{Items: []lesson.WordItem{
		{ID: 0, Questions: []string{"hond"}, Answers: []string{"dog"}},
		{ID: 1, Questions: []string{"kat"}, Answers: []string{"cat"}},
	}}}}
	w := NewWordsLessonWidget(l, nil)
	edits := 0
	w.SetOnModified(func() { edits++ })
	table := w.enterWidget.wordsTable

	table.Item(1, 1).SetText("cat; puss")
	if got := l.Data.List.Items[1].Answers; len(got) != 2 || got[1] != "puss" {
		return fmt.Errorf("edited answer not in the lesson: %q", got)
	}
	table.Item(0, 2).SetText("a pet")
	if l.Data.List.Items[0].Comment != "a pet" {
		return fmt.Errorf("edited comment not in the lesson: %q", l.Data.List.Items[0].Comment)
	}
	w.enterWidget.titleEdit.SetText("Dieren")
	if l.Data.List.Title != "Dieren" || edits != 3 {
		return fmt.Errorf("title %q, %d edits reported, want 3", l.Data.List.Title, edits)
	}
	w.enterWidget.addWordButton.Click()
	if n := len(l.Data.List.Items); n != 3 || l.Data.List.Items[2].ID != 2 {
		return fmt.Errorf("added word: %d items, new ID %d", n, l.Data.List.Items[n-1].ID)
	}
	// a new word is an empty row, the cursor in its question
	if it := l.Data.List.Items[2]; len(it.Questions) != 0 || table.CurrentRow() != 2 || table.CurrentColumn() != 0 {
		return fmt.Errorf("new word %+v, cursor at %d,%d", it, table.CurrentRow(), table.CurrentColumn())
	}
	// filling the table from the lesson is not an edit
	before := edits
	w.enterWidget.updateWordsTable()
	if edits != before {
		return fmt.Errorf("refilling the table reported %d edits", edits-before)
	}
	// Delete removes the selected words, not just the current one
	table.ClearSelection()
	table.SelectRow(0)
	table.SetRangeSelected(qt.NewQTableWidgetSelectionRange2(0, 0, 1, 2), true)
	w.enterWidget.tableKeys(func(*qt.QKeyEvent) {}, qt.NewQKeyEvent(qt.QEvent__KeyPress, int(qt.Key_Delete), qt.NoModifier))
	if n := len(l.Data.List.Items); n != 1 || l.Data.List.Items[0].ID != 2 {
		return fmt.Errorf("after Delete on two rows: %+v", l.Data.List.Items)
	}
	// Tab in the last cell starts a new word
	table.SetCurrentCell(0, table.ColumnCount()-1)
	w.enterWidget.tableKeys(func(*qt.QKeyEvent) {}, qt.NewQKeyEvent(qt.QEvent__KeyPress, int(qt.Key_Tab), qt.NoModifier))
	if n := len(l.Data.List.Items); n != 2 {
		return fmt.Errorf("Tab in the last cell: %d words", n)
	}
	return nil
}

func checkWordChoice() error {
	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{
		Items: []lesson.WordItem{
			{ID: 0, Questions: []string{"een"}, Answers: []string{"one"}},
			{ID: 1, Questions: []string{"twee"}, Answers: []string{"two"}},
		},
		Tests: []lesson.Test{{Results: []lesson.TestResult{{ItemID: 0, Result: "right"}, {ItemID: 1, Result: "wrong"}}}},
	}}}
	w := NewTeachTabWidget(l, nil)
	w.wordsCombo.SetCurrentText("Hard words")
	w.startButton.Click()
	if w.totalQuestions != 1 || w.questionLabel.Text() != "twee" {
		return fmt.Errorf("hard words: %d words, asking %q", w.totalQuestions, w.questionLabel.Text())
	}
	if w.wordsCombo.IsEnabled() {
		return fmt.Errorf("the word choice can be changed during a session")
	}

	// nothing to practise: a message, no session
	l.Data.List.Tests[0].Results[1].Result = "right"
	w = NewTeachTabWidget(l, nil)
	w.wordsCombo.SetCurrentText("Never answered correctly")
	w.startButton.Click()
	if w.isTeaching || !strings.Contains(w.statusLabel.Text(), "No words to practise") {
		return fmt.Errorf("empty choice: teaching %v, status %q", w.isTeaching, w.statusLabel.Text())
	}
	return nil
}

func checkSessionStored() error {
	l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{Items: []lesson.WordItem{
		{ID: 7, Questions: []string{"een"}, Answers: []string{"one"}},
		{ID: 9, Questions: []string{"twee"}, Answers: []string{"two"}},
	}}}}
	w := NewWordsLessonWidget(l, nil)
	modified := 0
	w.SetOnModified(func() { modified++ })
	t := w.teachWidget
	t.startButton.Click()
	for _, answer := range []string{"one", "wrong"} {
		t.answerEdit.SetText(answer)
		t.submitButton.Click()
		if t.nextButton.IsEnabled() {
			t.nextButton.Click() // after a wrong answer: continue
		}
	}
	tests := l.Data.List.Tests
	if len(tests) != 1 || len(tests[0].Results) != 2 || tests[0].Results[0].ItemID != 7 ||
		tests[0].Results[0].Result != "right" || tests[0].Results[1].Result != "wrong" {
		return fmt.Errorf("stored tests %+v", tests)
	}
	if modified != 1 {
		return fmt.Errorf("lesson reported modified %d times, want 1", modified)
	}
	return nil
}
