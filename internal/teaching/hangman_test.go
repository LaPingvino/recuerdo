package teaching

import (
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

func TestHangmanWin(t *testing.T) {
	h := NewHangmanWord([]string{"Tree", "boom"})
	if h.Word() != "Tree" || h.Shown() != "----" {
		t.Fatalf("start: %q %q", h.Word(), h.Shown())
	}
	steps := []struct {
		guess string
		want  GuessResult
		shown string
	}{
		{"e", GuessRight, "--ee"},
		{"E", GuessAlreadyTried, "--ee"},
		{"x", GuessWrong, "--ee"},
		{"t", GuessRight, "T-ee"},
		{"r", GuessWon, "Tree"},
	}
	for _, s := range steps {
		if got := h.Guess(s.guess); got != s.want || h.Shown() != s.shown {
			t.Errorf("Guess(%q) = %v, shown %q; want %v, %q", s.guess, got, h.Shown(), s.want, s.shown)
		}
	}
	if h.Mistakes != 1 || len(h.Wrong) != 1 || h.Wrong[0] != "x" {
		t.Errorf("mistakes %d %v", h.Mistakes, h.Wrong)
	}
}

func TestHangmanLose(t *testing.T) {
	h := NewHangmanWord([]string{"cat"})
	for _, g := range []string{"x", "y", "dog"} { // 1 + 1 + 2
		if got := h.Guess(g); got != GuessWrong {
			t.Fatalf("Guess(%q) = %v", g, got)
		}
	}
	h.Guess("z")
	if got := h.Guess("w"); got != GuessLost || h.Mistakes != 6 {
		t.Errorf("sixth mistake: %v, %d mistakes", got, h.Mistakes)
	}
}

func TestHangmanWholeWordAndRegexCharacters(t *testing.T) {
	h := NewHangmanWord([]string{"ice cream"})
	if h.Shown() != "--- -----" {
		t.Errorf("spaces should show: %q", h.Shown())
	}
	if got := h.Guess("."); got != GuessWrong || h.Shown() != "--- -----" {
		t.Errorf(`"." matched as a pattern: %v %q`, got, h.Shown())
	}
	if got := h.Guess("ICE CREAM"); got != GuessWon || h.Shown() != "ice cream" {
		t.Errorf("whole word: %v %q", got, h.Shown())
	}
}

func TestHangmanSession(t *testing.T) {
	s := New(list, Options{})
	s.Start()
	item, _, _ := s.Current()
	h := NewHangmanWord(item.Answers)
	for _, r := range "two" {
		h.Guess(string(r))
	}
	s.Record(true, h.Word())
	s.Record(false, "hanged man")
	res := s.Test().Results
	if len(res) != 2 || !res[0].Right || res[0].GivenAnswer != "two" || res[1].Right || res[1].GivenAnswer != "hanged man" {
		t.Errorf("results %+v", res)
	}
	if NewHangmanWord(nil).Word() != "" || (lesson.WordItem{}).Answers != nil {
		t.Error("empty answers")
	}
}
