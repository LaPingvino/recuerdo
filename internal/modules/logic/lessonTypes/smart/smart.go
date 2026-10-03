// Package smart asks every item, and asks again the ones answered wrong:
// soon (two items later) and once more at the end. Port of OpenTeacher's
// logic/lessonTypes/smart.
package smart

import (
	"context"

	"github.com/LaPingvino/recuerdo/internal/core"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
)

// LessonType is the smart lesson: a queue of item indexes to ask.
type LessonType struct {
	lessontypes.Base
	queue             []int
	asked             int
	current, previous int
	hasCurrent        bool
}

// New creates a smart lesson over the given item indexes.
func New(indexes []int) *LessonType {
	return &LessonType{queue: append([]int(nil), indexes...)}
}

func (l *LessonType) Start() { l.sendNext() }

func (l *LessonType) SetResult(r lessontypes.Result) {
	l.Record(r)
	l.asked++
	if !r.Right {
		// ask it again at the end...
		if n := len(l.queue); n > 0 && l.queue[n-1] != l.current {
			l.queue = append(l.queue, l.current)
		}
		// ...and soon, unless it is coming up already
		if len(l.queue) > 2 && l.queue[1] != l.current && l.queue[2] != l.current {
			l.insert(2, l.current)
		}
	}
	l.sendNext()
}

func (l *LessonType) Skip() {
	l.insert(2, l.current)
	l.sendNext()
}

// CorrectLastAnswer turns the last (wrong) result into the given one and
// takes back the repetitions SetResult queued for it.
func (l *LessonType) CorrectLastAnswer(r lessontypes.Result) {
	l.ReplaceLast(r)
	if n := len(l.queue); n > 0 && l.queue[n-1] == l.previous {
		l.queue = l.queue[:n-1]
	}
	// the repetition at 2 is at 1 now that the next item was taken
	if len(l.queue) > 1 && l.queue[1] == l.previous {
		l.queue = append(l.queue[:1], l.queue[2:]...)
	}
}

func (l *LessonType) TotalItems() int { return len(l.queue) + l.asked }
func (l *LessonType) AskedItems() int { return l.asked }

// insert puts index at position i, or at the end if the queue is shorter
// (as Python's list.insert does).
func (l *LessonType) insert(i, index int) {
	if i > len(l.queue) {
		i = len(l.queue)
	}
	l.queue = append(l.queue[:i], append([]int{index}, l.queue[i:]...)...)
}

func (l *LessonType) sendNext() {
	if l.hasCurrent {
		l.previous = l.current
	}
	if len(l.queue) == 0 {
		l.Done()
		return
	}
	l.current, l.queue = l.queue[0], l.queue[1:]
	l.hasCurrent = true
	l.Ask(l.current)
}

// SmartModule offers the lesson type as an OpenTeacher "lessonType" module.
type SmartModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewSmartModule creates the module.
func NewSmartModule() *SmartModule {
	base := core.NewBaseModule("lessonType", "smart")
	base.SetPriority(130)
	return &SmartModule{BaseModule: base}
}

// DisplayName is the lesson type's name as shown to users (Name is the module identifier).
func (mod *SmartModule) DisplayName() string { return "Smart" }

// CreateLessonType starts a lesson over the given item indexes.
func (mod *SmartModule) CreateLessonType(indexes []int) lessontypes.LessonType {
	return New(indexes)
}

func (mod *SmartModule) Enable(ctx context.Context) error  { return mod.BaseModule.Enable(ctx) }
func (mod *SmartModule) Disable(ctx context.Context) error { return mod.BaseModule.Disable(ctx) }

// SetManager sets the module manager reference.
func (mod *SmartModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitSmartModule creates and returns the module.
func InitSmartModule() core.Module { return NewSmartModule() }
