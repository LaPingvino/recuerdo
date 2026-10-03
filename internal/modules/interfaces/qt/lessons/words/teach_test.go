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
	os.Exit(m.Run())
}

var teachErr error

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
