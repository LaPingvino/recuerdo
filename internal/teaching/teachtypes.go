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
)

// TeachTypes lists the practice modes, default first.
var TeachTypes = []string{TeachTyping, ShuffleAnswer, RepeatAnswer}

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
