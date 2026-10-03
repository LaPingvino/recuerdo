package teaching

import (
	"math/rand"

	"strings"
	"time"
)

// Teach types (practice modes), as offered to the user. All three are
// OpenTeacher's typing mode; the other two add something on screen.
const (
	// Typing: type the answer.
	TeachTyping = "Typing"
	// ShuffleAnswer: typing, with the answer's letters shuffled as a hint.
	ShuffleAnswer = "Shuffle answer"
	// RepeatAnswer: the answer is shown first and fades away, then typed
	// from memory.
	RepeatAnswer = "Repeat answer"
	// InMind: think of the answer, look at it, and say whether you knew it.
	InMind = "In mind"
)

// TeachTypes lists the practice modes, default first.
var TeachTypes = []string{TeachTyping, ShuffleAnswer, RepeatAnswer, InMind, Hangman}

// RepeatFadeDuration is how long Repeat answer shows the answer
// (OpenTeacher's default).
const RepeatFadeDuration = 3000 * time.Millisecond

// ShuffleHint is Shuffle answer's hint for an answer: its characters
// shuffled so that each one moves (OpenTeacher's algorithm), or dots when
// the answer is too short or the shuffle gives the answer itself. random
// returns numbers in [0, 1); math/rand when nil.
func ShuffleHint(answer string, random func() float64) string {
	if random == nil {
		random = rand.Float64
	}
	old := []rune(answer)
	shuffled := append([]rune(nil), old...)
	for i := 0; i < len(shuffled)-1; i++ {
		j := i + 1 + int(random()*float64(len(shuffled)-i-1))
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	if len(shuffled) <= 2 || string(shuffled) == answer {
		return "Hint: " + strings.Repeat(".", len(shuffled))
	}
	return "Hint: " + string(shuffled)
}

// CurrentAnswer is the right answer to the current question, in
// OpenTeacher's notation ("" when nothing is asked).
func (s *Session) CurrentAnswer() string {
	item, _, ok := s.Current()
	if !ok {
		return ""
	}
	return compose(item.Answers)
}

// ViewAnswer is In mind's "View answer": the thinking time for the
// current question ends now. It returns the answer to show.
func (s *Session) ViewAnswer() string {
	if !s.hasItem || s.viewed {
		return s.CurrentAnswer()
	}
	s.viewedAt, s.viewed = s.opts.Now(), true
	return s.CurrentAnswer()
}

// Judge is In mind's "I was right" / "I was wrong": it records the user's
// own verdict on the current question and moves to the next one.
func (s *Session) Judge(right bool) { s.Record(right, "") }
