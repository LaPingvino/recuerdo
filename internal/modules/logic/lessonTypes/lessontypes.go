// Package lessontypes holds what OpenTeacher's lesson types share: a lesson
// type decides which item of a word list to ask next and records the
// results as a test. Implementations live in the allOnce, smart and
// interval packages.
package lessontypes

import "time"

// Result is the outcome of one question.
type Result struct {
	ItemID int
	Right  bool
	// GivenAnswer is what the user answered (optional).
	GivenAnswer string
	// Start and End are when the question was shown and answered
	// (optional; OpenTeacher's "active" period, for thinking time).
	Start, End time.Time
}

// Pause is a period the user paused the lesson.
type Pause struct {
	Start, End time.Time
}

// Test is the record of one run through a lesson.
type Test struct {
	Results  []Result
	Finished bool
	Pauses   []Pause
}

// LessonType asks the items of a list in some order. Items are referred to
// by their index in the list the caller holds.
type LessonType interface {
	// OnNewItem sets the function called with the index of each item to ask.
	OnNewItem(func(index int))
	// OnLessonDone sets the function called when there is nothing left to ask.
	OnLessonDone(func())

	// Start asks the first item.
	Start()
	// SetResult records the result for the item just asked and asks the next.
	SetResult(Result)
	// Skip puts the current item back for later and asks the next.
	Skip()
	// CorrectLastAnswer replaces the last result ("I was right").
	CorrectLastAnswer(Result)
	AddPause(Pause)

	TotalItems() int
	AskedItems() int
	// Test is the record of this run; nil until the first result.
	Test() *Test
}

// Base implements the bookkeeping all lesson types share.
type Base struct {
	newItem    func(int)
	lessonDone func()
	test       *Test
	current    Test
}

func (b *Base) OnNewItem(f func(int)) { b.newItem = f }
func (b *Base) OnLessonDone(f func()) { b.lessonDone = f }
func (b *Base) AddPause(p Pause)      { b.current.Pauses = append(b.current.Pauses, p) }
func (b *Base) Test() *Test           { return b.test }

// Record adds a result to the test, which starts existing with its first
// result (as OpenTeacher adds a test to the list on the first answer).
func (b *Base) Record(r Result) {
	b.current.Results = append(b.current.Results, r)
	b.test = &b.current
}

// ReplaceLast replaces the last recorded result.
func (b *Base) ReplaceLast(r Result) {
	if n := len(b.current.Results); n > 0 {
		b.current.Results[n-1] = r
	}
}

// Ask reports the next item to ask.
func (b *Base) Ask(index int) {
	if b.newItem != nil {
		b.newItem(index)
	}
}

// Done marks the test finished (if anything was answered) and reports the
// end of the lesson.
func (b *Base) Done() {
	if len(b.current.Results) > 0 {
		b.current.Finished = true
	}
	if b.lessonDone != nil {
		b.lessonDone()
	}
}
