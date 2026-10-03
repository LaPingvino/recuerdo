package teaching

import (
	"strings"
	"unicode"

	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/checker"
)

// Hangman practice mode: guess the answer letter by letter.
const Hangman = "Hangman"

// HangmanMaxMistakes is how many mistakes hang the man.
const HangmanMaxMistakes = 6

// Guess outcomes.
type GuessResult int

const (
	GuessAlreadyTried GuessResult = iota
	GuessRight
	GuessWrong
	GuessWon
	GuessLost
)

// HangmanWord is one round of hangman. Port of OpenTeacher's hangman teach
// type: a wrong letter is one mistake, a wrong guess of the whole word two,
// and six mistakes lose. Letters match whatever their capitals, and the
// guess is matched as text (OpenTeacher matched it as a regular
// expression, so "." matched every letter).
type HangmanWord struct {
	word     []rune
	shown    []rune
	Mistakes int
	// Wrong lists the wrongly guessed letters, in order.
	Wrong []string
	tried map[string]bool
}

// NewHangmanWord starts a round for an answer as stored in the lesson: the
// word is its first alternative. Spaces are shown from the start.
func NewHangmanWord(stored []string) *HangmanWord {
	word := ""
	if parts := checker.StoredAnswers(stored); len(parts) > 0 && len(parts[0]) > 0 {
		word = parts[0][0]
	}
	h := &HangmanWord{word: []rune(word), tried: map[string]bool{}}
	for _, r := range h.word {
		if r == ' ' {
			h.shown = append(h.shown, ' ')
		} else {
			h.shown = append(h.shown, '-')
		}
	}
	return h
}

// Word is the word to guess.
func (h *HangmanWord) Word() string { return string(h.word) }

// Shown is the word as guessed so far, with "-" for letters still hidden.
func (h *HangmanWord) Shown() string { return string(h.shown) }

// Guess guesses a letter (one character) or the whole word.
func (h *HangmanWord) Guess(guess string) GuessResult {
	guess = strings.TrimSpace(guess)
	key := strings.ToLower(guess)
	if guess == "" || h.tried[key] {
		return GuessAlreadyTried
	}
	h.tried[key] = true

	if g := []rune(guess); len(g) == 1 {
		found := false
		for i, r := range h.word {
			if unicode.ToLower(r) == unicode.ToLower(g[0]) {
				h.shown[i] = r
				found = true
			}
		}
		if found {
			if string(h.shown) == string(h.word) {
				return GuessWon
			}
			return GuessRight
		}
		h.Mistakes++
		h.Wrong = append(h.Wrong, guess)
	} else {
		if strings.EqualFold(guess, string(h.word)) {
			copy(h.shown, h.word)
			return GuessWon
		}
		h.Mistakes += 2
	}
	if h.Mistakes >= HangmanMaxMistakes {
		return GuessLost
	}
	return GuessWrong
}

// Record records the end of a round of a self-run mode (hangman, In mind)
// as the result for the current question, and moves to the next one.
func (s *Session) Record(right bool, given string) {
	if !s.hasItem {
		return
	}
	end := s.opts.Now()
	if s.viewed {
		end = s.viewedAt
	}
	_, index, _ := s.Current()
	s.pending = &lessontypes.Result{ItemID: index, Right: right, GivenAnswer: given, Start: s.asked, End: end}
	s.viewed = false
	s.Next()
}
