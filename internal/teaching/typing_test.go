package teaching

import (
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// fakeUI records what the controller told it.
type fakeUI struct {
	input, check, skip, correctAnyway bool
	correction                        string
	showing                           bool
	cleared, done                     int
}

func (u *fakeUI) ClearInput()                    { u.cleared++ }
func (u *fakeUI) FocusInput()                    {}
func (u *fakeUI) SetInputEnabled(b bool)         { u.input = b }
func (u *fakeUI) SetCheckEnabled(b bool)         { u.check = b }
func (u *fakeUI) SetSkipEnabled(b bool)          { u.skip = b }
func (u *fakeUI) SetCorrectAnywayEnabled(b bool) { u.correctAnyway = b }
func (u *fakeUI) ShowCorrection(a string)        { u.correction, u.showing = a, true }
func (u *fakeUI) HideCorrection()                { u.showing = false }
func (u *fakeUI) LessonDone()                    { u.done++ }
func (u *fakeUI) uiEnabled() bool                { return u.input && u.check && u.skip }
func (u *fakeUI) uiDisabled() bool               { return !u.input && !u.check && !u.skip }

// the two-word list OpenTeacher's inputTypingLogicTest uses
var twoWords = lesson.WordList{Items: []lesson.WordItem{
	{ID: 0, Questions: []string{"een"}, Answers: []string{"one"}},
	{ID: 1, Questions: []string{"twee"}, Answers: []string{"two"}},
}}

func showingACorrection(t *testing.T) (*Typing, *fakeUI) {
	t.Helper()
	ui := &fakeUI{}
	ty := NewTyping(New(twoWords, Options{}), ui)
	if err := ty.Check("a wrong answer"); err != nil {
		t.Fatal(err)
	}
	if !ui.showing || ui.correction != "one" || !ui.uiDisabled() || !ui.correctAnyway {
		t.Fatalf("after a wrong answer: %+v", ui)
	}
	return ty, ui
}

// Ported from OpenTeacher's inputTypingLogicTest.
func TestMethodsWhileShowingCorrection(t *testing.T) {
	ty, _ := showingACorrection(t)
	if err := ty.Check("whatever"); err != ErrShowingCorrection {
		t.Errorf("Check: %v", err)
	}
	if err := ty.Skip(); err != ErrShowingCorrection {
		t.Errorf("Skip: %v", err)
	}
}

func TestCorrectAnywayWhileShowingCorrection(t *testing.T) {
	ty, ui := showingACorrection(t)
	if err := ty.CorrectAnyway(); err != nil {
		t.Fatal(err)
	}
	if ui.showing || !ui.uiEnabled() || ui.correctAnyway {
		t.Errorf("after correct anyway: %+v", ui)
	}
	res := ty.session.Test().Results
	if len(res) != 1 || !res[0].Right || res[0].GivenAnswer != "Corrected: a wrong answer" {
		t.Errorf("results %+v", res)
	}
	if right, answered := ty.session.Score(); right != 1 || answered != 1 {
		t.Errorf("score %d/%d", right, answered)
	}
}

func TestSkip(t *testing.T) {
	ui := &fakeUI{}
	ty := NewTyping(New(twoWords, Options{}), ui)
	if err := ty.Skip(); err != nil {
		t.Fatal(err)
	}
	if item, _, _ := ty.session.Current(); item.Questions[0] != "twee" {
		t.Errorf("after skip asking %q", item.Questions[0])
	}
}

func TestCompleteLesson(t *testing.T) {
	ui := &fakeUI{}
	ty := NewTyping(New(twoWords, Options{}), ui)
	for _, a := range []string{"one", "two"} {
		if err := ty.Check(a); err != nil {
			t.Fatal(err)
		}
	}
	if ui.done != 1 || !ui.uiDisabled() {
		t.Errorf("lesson end: %+v", ui)
	}
	if err := ty.Check("more"); err != ErrNoLesson {
		t.Errorf("Check after the end: %v", err)
	}
}

func TestCallingCorrectionShowingDoneWhileNoCorrectionIsShown(t *testing.T) {
	ty := NewTyping(New(twoWords, Options{}), &fakeUI{})
	if err := ty.CorrectionShowingDone(); err != ErrNoCorrectionShown {
		t.Errorf("got %v", err)
	}
	if err := ty.CorrectAnyway(); err != ErrNothingToCorrect {
		t.Errorf("correct anyway without a wrong answer: %v", err)
	}
}

func TestRightAnswerMovesOnAtOnce(t *testing.T) {
	ui := &fakeUI{}
	ty := NewTyping(New(twoWords, Options{}), ui)
	before := ui.cleared
	if err := ty.Check("ONE"); err != nil {
		t.Fatal(err)
	}
	if item, _, _ := ty.session.Current(); item.Questions[0] != "twee" || ui.cleared != before+1 || ui.showing {
		t.Errorf("after a right answer: asking %q, ui %+v", item.Questions[0], ui)
	}
}

func TestContinueAfterCorrection(t *testing.T) {
	ty, ui := showingACorrection(t)
	if err := ty.CorrectionShowingDone(); err != nil {
		t.Fatal(err)
	}
	if ui.showing || !ui.uiEnabled() || !ui.correctAnyway {
		t.Errorf("after continuing: %+v (correct anyway should stay possible)", ui)
	}
	if item, _, _ := ty.session.Current(); item.Questions[0] != "twee" {
		t.Errorf("asking %q", item.Questions[0])
	}
	// and it can still be corrected afterwards
	if err := ty.CorrectAnyway(); err != nil || !ty.session.Test().Results[0].Right {
		t.Errorf("late correct anyway: %v %+v", err, ty.session.Test().Results)
	}
}
