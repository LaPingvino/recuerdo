// Package teaching runs a practice session over a word list, the way
// OpenTeacher's words teacher does: a list modifier picks the order, a
// lesson type decides what to ask next (and what to ask again), and the
// words string checker judges the answers. It has no GUI of its own.
package teaching

import (
	"github.com/LaPingvino/recuerdo/internal/lesson"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	allonce "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/allOnce"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/interval"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/smart"
	random "github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/random_"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/reverse"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/sort"
	wordsreverser "github.com/LaPingvino/recuerdo/internal/modules/logic/reversers/words"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/checker"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/composer"
)

// Lesson types, as offered to the user.
const (
	AllOnce  = "All once"
	Smart    = "Smart"
	Interval = "Interval"
)

// List orders, as offered to the user.
const (
	AsEntered = "As entered"
	Random    = "Random"
	Reversed  = "Reversed"
	Sorted    = "Sorted"
)

// LessonTypes and Orders list the choices, default first.
var (
	LessonTypes = []string{AllOnce, Smart, Interval}
	Orders      = []string{AsEntered, Random, Reversed, Sorted}
)

// Options configure a session.
type Options struct {
	LessonType string // one of LessonTypes; AllOnce if unknown
	Order      string // one of Orders; AsEntered if unknown
	// AskAnswers practises the other way round: answers become questions.
	AskAnswers    bool
	CaseSensitive bool
	// Shuffle replaces the random shuffle (for tests).
	Shuffle func(n int, swap func(i, j int))
	// Intn replaces the interval lesson's random choice (for tests).
	Intn func(n int) int
}

// Answer is the outcome of answering the current question.
type Answer struct {
	Right bool
	// Correct is the right answer in OpenTeacher's notation.
	Correct string
}

// Session is one practice run.
type Session struct {
	list     lesson.WordList
	lt       lessontypes.LessonType
	opts     Options
	current  int
	hasItem  bool
	done     bool
	pending  *lessontypes.Result
	right    int
	answered int
}

// New prepares a session over list; call Start to ask the first question.
// The list is not changed.
func New(list lesson.WordList, opts Options) *Session {
	list.Items = append([]lesson.WordItem(nil), list.Items...)
	if opts.AskAnswers {
		wordsreverser.Reverse(&list)
	}

	indexes := make([]int, len(list.Items))
	for i := range indexes {
		indexes[i] = i
	}
	switch opts.Order {
	case Random:
		indexes = random.ModifyList(indexes, opts.Shuffle)
	case Reversed:
		indexes = reverse.ModifyList(indexes)
	case Sorted:
		indexes = sort.ModifyList(indexes, list.Items)
	}

	s := &Session{list: list, opts: opts}
	switch opts.LessonType {
	case Smart:
		s.lt = smart.New(indexes)
	case Interval:
		l := interval.New(indexes, interval.DefaultSettings)
		if opts.Intn != nil {
			l.Intn = opts.Intn
		}
		s.lt = l
	default:
		s.lt = allonce.New(indexes)
	}
	s.lt.OnNewItem(func(i int) { s.current, s.hasItem = i, true })
	s.lt.OnLessonDone(func() { s.done, s.hasItem = true, false })
	return s
}

// Start asks the first question.
func (s *Session) Start() { s.lt.Start() }

// Done reports whether the session is over.
func (s *Session) Done() bool { return s.done }

// Current is the item to ask (questions and answers already swapped when
// practising the other way round) and its index in the list.
func (s *Session) Current() (lesson.WordItem, int, bool) {
	if !s.hasItem {
		return lesson.WordItem{}, 0, false
	}
	return s.list.Items[s.current], s.current, true
}

// Answer checks the answer to the current question. The result counts
// once Next is called, so the lesson type can still change its plans.
func (s *Session) Answer(text string) Answer {
	item, index, ok := s.Current()
	if !ok {
		return Answer{}
	}
	right := checker.CorrectText(text, item.Answers, s.opts.CaseSensitive)
	s.pending = &lessontypes.Result{ItemID: index, Right: right}
	return Answer{Right: right, Correct: composer.Compose(checker.StoredAnswers(item.Answers))}
}

// Next records the last answer and moves to the next question.
func (s *Session) Next() {
	if s.pending == nil {
		return
	}
	r := *s.pending
	s.pending = nil
	s.answered++
	if r.Right {
		s.right++
	}
	s.lt.SetResult(r)
}

// Skip puts the current question back for later.
func (s *Session) Skip() {
	if s.pending == nil && s.hasItem {
		s.lt.Skip()
	}
}

// Progress is how many questions have been answered and how many there are
// in total so far (lesson types that repeat items add to the total).
func (s *Session) Progress() (asked, total int) {
	return s.answered, s.answered + s.remaining()
}

func (s *Session) remaining() int {
	r := s.lt.TotalItems() - s.lt.AskedItems()
	if r < 0 {
		return 0
	}
	return r
}

// Score is the number of right answers and answers given.
func (s *Session) Score() (right, answered int) { return s.right, s.answered }
