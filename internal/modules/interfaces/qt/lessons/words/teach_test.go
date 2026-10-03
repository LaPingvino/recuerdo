package words

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/mappu/miqt/qt"
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
	os.Exit(m.Run())
}

var teachErr, modesErr, inMindErr error

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
	w.repeatShown() // what the timer does when the answer has been shown long enough
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
