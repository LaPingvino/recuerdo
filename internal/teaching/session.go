// Package teaching runs a practice session over a word list, the way
// OpenTeacher's words teacher does: a list modifier picks the order, a
// lesson type decides what to ask next (and what to ask again), and the
// words string checker judges the answers. It has no GUI of its own.
package teaching

import (
	"github.com/LaPingvino/recuerdo/internal/richtext"
	"time"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	allonce "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/allOnce"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/interval"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/smart"
	hardwords "github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/hardWords"
	random "github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/random_"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/reverse"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/sort"
	neverright "github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/wordsNeverAnsweredCorrectly"
	wordsreverser "github.com/LaPingvino/recuerdo/internal/modules/logic/reversers/words"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/checker"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/composer"
)

// Lesson types, as offered to the user.
// i18n:values (shown translated, see valuecombo)
const (
	AllOnce  = "All once"
	Smart    = "Smart"
	Interval = "Interval"
)

// List orders, as offered to the user.
// i18n:values (shown translated, see valuecombo)
const (
	AsEntered = "As entered"
	Random    = "Random"
	Reversed  = "Reversed"
	Sorted    = "Sorted"
)

// Which words to practise, as offered to the user.
// i18n:values (shown translated, see valuecombo)
const (
	AllWords   = "All words"
	HardWords  = "Hard words"
	NeverRight = "Never answered correctly"
)

// LessonTypes, Orders and WordChoices list the choices, default first.
var (
	LessonTypes = []string{AllOnce, Smart, Interval}
	Orders      = []string{AsEntered, Random, Reversed, Sorted}
	WordChoices = []string{AllWords, HardWords, NeverRight}
)

// Options configure a session.
type Options struct {
	LessonType string // one of LessonTypes; AllOnce if unknown
	Order      string // one of Orders; AsEntered if unknown
	Words      string // one of WordChoices; AllWords if unknown
	// AskAnswers practises the other way round: answers become questions.
	AskAnswers    bool
	CaseSensitive bool
	// Shuffle replaces the random shuffle (for tests).
	Shuffle func(n int, swap func(i, j int))
	// Intn replaces the interval lesson's random choice (for tests).
	Intn func(n int) int
	// Now replaces the clock used to time answers (for tests).
	Now func() time.Time
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
	asked    time.Time // when the current question was shown
	last     *lessontypes.Result
	viewed   bool      // In mind: the answer was viewed
	viewedAt time.Time // ...at this time
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
	// which words, judged by the list's earlier tests
	switch opts.Words {
	case HardWords:
		indexes = hardwords.ModifyList(indexes, list)
	case NeverRight:
		indexes = neverright.ModifyList(indexes, list)
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
	if s.opts.Now == nil {
		s.opts.Now = time.Now
	}
	s.lt.OnNewItem(func(i int) { s.current, s.hasItem, s.asked, s.viewed = i, true, s.opts.Now(), false })
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
	right := Correct(text, item.Answers, s.opts.CaseSensitive)
	s.pending = &lessontypes.Result{ItemID: index, Right: right, GivenAnswer: text, Start: s.asked, End: s.opts.Now()}
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
	s.last = &r
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

// Test is the record of this run so far (empty before the first answer).
func (s *Session) Test() lessontypes.Test {
	if t := s.lt.Test(); t != nil {
		return *t
	}
	return lessontypes.Test{}
}

// LessonTest is the session's results as a test to store in the lesson,
// as OpenTeacher keeps every test in the list: each result refers to its
// word by ID, with the time it was answered; the test's date is when the
// first question was asked. Empty if nothing was answered.
func (s *Session) LessonTest() lesson.Test {
	var t lesson.Test
	for _, r := range s.Test().Results {
		if r.ItemID < 0 || r.ItemID >= len(s.list.Items) {
			continue
		}
		res := lesson.TestResult{ItemID: s.list.Items[r.ItemID].ID, Result: "wrong"}
		if r.Right {
			res.Result = "right"
		}
		if !r.End.IsZero() {
			end := r.End
			res.Time = &end
		}
		if t.Date == nil && !r.Start.IsZero() {
			start := r.Start
			t.Date = &start
		}
		t.Results = append(t.Results, res)
	}
	return t
}

// List is the list being practised (questions and answers swapped when
// practising the other way round); result ItemIDs index its items.
func (s *Session) List() lesson.WordList { return s.list }

// CorrectLast turns the last recorded answer into a right one ("I was
// right"), as OpenTeacher's "Correct anyway" does; the given answer is
// marked as corrected.
func (s *Session) CorrectLast() {
	if s.last == nil || s.last.Right {
		return
	}
	r := *s.last
	r.Right = true
	r.GivenAnswer = "Corrected: " + r.GivenAnswer
	s.last = &r
	s.right++
	s.lt.CorrectLastAnswer(r)
}

// Correct reports whether typed is a right answer to stored, as
// OpenTeacher checks it, for words with markup and formulas too (shared
// by the desktop and the web version):
//   - a formula answer ($x^2 + 1$) is right when typed is its TeX, spaces
//     not counting (x^2+1), not split at OpenTeacher's commas (f(x, y));
//   - other answers are compared by their plain text (H<sub>2</sub>O is
//     typed H2O; furigana left out), with OpenTeacher's notation.
func Correct(typed string, stored []string, caseSensitive bool) bool {
	plain := make([]string, len(stored))
	math := false
	for i, w := range stored {
		plain[i] = richtext.Plain(w)
		math = math || richtext.HasMath(w)
	}
	if math {
		for _, w := range plain {
			if richtext.NormalizeAnswer(w) == richtext.NormalizeAnswer(typed) {
				return true
			}
		}
		return false
	}
	return checker.CorrectText(typed, plain, caseSensitive)
}
