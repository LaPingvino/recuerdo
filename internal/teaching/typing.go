package teaching

import "errors"

// TypingUI is the part of a GUI where the user types answers; the Typing
// controller tells it what to show. Port of the events of OpenTeacher's
// inputTypingLogic controller.
type TypingUI interface {
	ClearInput()
	FocusInput()
	SetInputEnabled(bool)
	SetCheckEnabled(bool)
	SetSkipEnabled(bool)
	SetCorrectAnywayEnabled(bool)
	// ShowCorrection shows the right answer after a wrong one.
	ShowCorrection(answer string)
	HideCorrection()
	// LessonDone is called when there is nothing left to ask.
	LessonDone()
}

// Errors for calls that make no sense in the controller's state.
var (
	ErrNoLesson          = errors.New("no lesson active")
	ErrShowingCorrection = errors.New("showing a correction")
	ErrNoCorrectionShown = errors.New("not showing a correction")
	ErrNothingToCorrect  = errors.New("no answer to correct")
)

// Typing is OpenTeacher's typing practice mode: the user types an answer
// and checks it. A right answer moves on at once; a wrong one shows the
// correction until the user continues, or says it was right after all
// ("Correct anyway"). Port of logic/interfaces/inputTypingLogic.
type Typing struct {
	session           *Session
	ui                TypingUI
	showingCorrection bool
	canCorrect        bool
}

// NewTyping starts practising session in ui.
func NewTyping(session *Session, ui TypingUI) *Typing {
	t := &Typing{session: session, ui: ui}
	session.Start()
	t.ui.SetCorrectAnywayEnabled(false)
	t.enableUI(true)
	t.next()
	return t
}

// ShowingCorrection reports whether a correction is on screen.
func (t *Typing) ShowingCorrection() bool { return t.showingCorrection }

// Check checks the typed answer.
func (t *Typing) Check(input string) error {
	if err := t.ready(); err != nil {
		return err
	}
	answer := t.session.Answer(input)
	if answer.Right {
		t.setCanCorrect(false)
		t.session.Next()
		t.next()
		return nil
	}
	t.enableUI(false)
	t.showingCorrection = true
	t.setCanCorrect(true)
	t.ui.ShowCorrection(answer.Correct)
	return nil
}

// CorrectionShowingDone continues after a correction.
func (t *Typing) CorrectionShowingDone() error {
	if !t.showingCorrection {
		return ErrNoCorrectionShown
	}
	t.showingCorrection = false
	t.ui.HideCorrection()
	t.session.Next()
	t.enableUI(true)
	t.next()
	return nil
}

// CorrectAnyway counts the last wrong answer as right.
func (t *Typing) CorrectAnyway() error {
	if !t.canCorrect {
		return ErrNothingToCorrect
	}
	if t.showingCorrection {
		if err := t.CorrectionShowingDone(); err != nil {
			return err
		}
	}
	t.session.CorrectLast()
	t.setCanCorrect(false)
	return nil
}

// Skip puts the current word back for later.
func (t *Typing) Skip() error {
	if err := t.ready(); err != nil {
		return err
	}
	t.session.Skip()
	t.next()
	return nil
}

func (t *Typing) ready() error {
	if t.session.Done() {
		return ErrNoLesson
	}
	if t.showingCorrection {
		return ErrShowingCorrection
	}
	return nil
}

// next shows the next question, or ends the lesson.
func (t *Typing) next() {
	if t.session.Done() {
		t.enableUI(false)
		t.ui.LessonDone()
		return
	}
	t.ui.ClearInput()
	t.ui.FocusInput()
}

func (t *Typing) enableUI(on bool) {
	t.ui.SetCheckEnabled(on)
	t.ui.SetSkipEnabled(on)
	t.ui.SetInputEnabled(on)
}

func (t *Typing) setCanCorrect(on bool) {
	t.canCorrect = on
	t.ui.SetCorrectAnywayEnabled(on)
}
