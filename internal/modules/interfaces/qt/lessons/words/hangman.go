package words

import (
	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/LaPingvino/recuerdo/internal/teaching"
	qt "github.com/mappu/miqt/qt6"
)

// drawHangman draws the gallows with the man as far as the mistakes go
// (OpenTeacher's hangman graphics, line for line).
func drawHangman(mistakes int) *qt.QPixmap {
	pm := qt.NewQPixmap2(300, 200)
	pm.FillWithFillColor(qt.NewQColor3(255, 255, 255))
	p := qt.NewQPainter2(pm.QPaintDevice)
	p.SetRenderHint(qt.QPainter__Antialiasing)
	p.SetPenWithPen(qt.NewQPen4(qt.NewQBrush4(qt.Black), 2))
	line := p.DrawLine2

	// gallows
	line(20, 190, 280, 190)
	line(60, 190, 60, 20)
	line(20, 190, 60, 150)
	line(100, 190, 60, 150)
	line(60, 20, 190, 20)
	line(60, 60, 100, 20)
	line(190, 20, 190, 40)
	// the man
	if mistakes >= 1 {
		p.DrawEllipse2(178, 40, 24, 24)
	}
	if mistakes >= 2 {
		line(190, 63, 190, 120)
	}
	if mistakes >= 3 {
		line(190, 120, 144, 166)
	}
	if mistakes >= 4 {
		line(190, 120, 236, 166)
	}
	if mistakes >= 5 {
		line(190, 63, 144, 109)
	}
	if mistakes >= 6 {
		line(190, 63, 236, 109)
		// crossed-out eyes and a straight mouth
		line(185, 47, 189, 51)
		line(185, 51, 189, 47)
		line(191, 51, 195, 47)
		line(191, 47, 195, 51)
		line(185, 56, 195, 56)
	}
	p.End()
	return pm
}

// setHangmanLayout swaps the typing buttons for the hangman screen.
func (w *TeachTabWidget) setHangmanLayout(on bool) {
	w.hangmanLabel.SetVisible(on)
	w.nextButton.SetVisible(!on)
	w.correctButton.SetVisible(!on)
	w.skipButton.SetVisible(!on)
	if on {
		w.submitButton.SetText(i18n.T("Check!"))
	} else {
		w.submitButton.SetText(i18n.T("Check"))
	}
}

// hangmanNext starts a round for the next word, or ends the session.
func (w *TeachTabWidget) hangmanNext() {
	if w.session.Done() {
		w.hangman = nil
		w.finishTeaching()
		return
	}
	w.showCurrentQuestion()
	item, _, _ := w.session.Current()
	w.hangman = teaching.NewHangmanWord(item.Answers)
	w.resultLabel.SetVisible(false)
	w.answerEdit.SetEnabled(true)
	w.submitButton.SetEnabled(true)
	w.answerEdit.SetFocus()
	w.showHangman()
}

func (w *TeachTabWidget) showHangman() {
	w.hintLabel.SetStyleSheet("font-size: 24px; font-family: monospace; letter-spacing: 4px;")
	w.hintLabel.SetText(w.hangman.Shown())
	w.hintLabel.SetVisible(true)
	w.hangmanLabel.SetPixmap(drawHangman(w.hangman.Mistakes))
}

// hangmanGuess handles a guessed letter or word.
func (w *TeachTabWidget) hangmanGuess(guess string) {
	if w.hangman == nil {
		return
	}
	w.answerEdit.Clear()
	switch w.hangman.Guess(guess) {
	case teaching.GuessAlreadyTried:
		w.feedback("You have already tried this character / word", false)
	case teaching.GuessRight:
		w.resultLabel.SetVisible(len(w.hangman.Wrong) > 0)
	case teaching.GuessWrong:
		w.feedback("Mistakes:  "+joinMistakes(w.hangman.Wrong), false)
	case teaching.GuessWon:
		w.session.Record(true, w.hangman.Word())
		w.feedback("Right!", true)
		w.hangmanNext()
		return
	case teaching.GuessLost:
		// show the answer for a while, then go on
		w.feedback("You lose, the answer was: "+w.hangman.Word(), false)
		w.answerEdit.SetEnabled(false)
		w.submitButton.SetEnabled(false)
		w.hangmanTimer.Start(int(w.RepeatDuration.Milliseconds()))
	}
	w.showHangman()
}

// hangmanLost goes on after a lost round has been shown.
func (w *TeachTabWidget) hangmanLost() {
	if w.hangman == nil || w.session == nil || w.session.Done() {
		return
	}
	w.session.Record(false, "hanged man")
	w.hangmanNext()
}

func joinMistakes(wrong []string) string {
	s := ""
	for i, m := range wrong {
		if i > 0 {
			s += "  |  "
		}
		s += m
	}
	return s
}
