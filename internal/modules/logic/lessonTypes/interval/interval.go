// Package interval asks items in small groups until each one is known:
// an item comes back a few questions later until it has been asked often
// enough and answered right often enough. Port of OpenTeacher's
// logic/lessonTypes/interval.
package interval

import (
	"context"
	"math/rand"

	"github.com/LaPingvino/recuerdo/internal/core"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
)

// Settings are the interval lesson's settings, as in OpenTeacher.
type Settings struct {
	GroupSize    int // how close together repetitions come (at least 2)
	MinQuestions int // how often an item is asked at least (at least 1)
	WhenKnown    int // percentage right that counts as known (0-99)
}

// DefaultSettings are OpenTeacher's defaults.
var DefaultSettings = Settings{GroupSize: 4, MinQuestions: 2, WhenKnown: 80}

func (s Settings) groupSize() int { return max(s.GroupSize, 2) }

func (s Settings) minQuestions() int {
	if s.MinQuestions < 1 {
		return 2
	}
	return s.MinQuestions
}

func (s Settings) whenKnown() int {
	if s.WhenKnown < 0 || s.WhenKnown > 99 {
		return 80
	}
	return s.WhenKnown
}

// LessonType is the interval lesson.
type LessonType struct {
	lessontypes.Base
	settings Settings
	queue    []int
	current  int
	// the item index each recorded result is for
	resultIndexes []int
	// Intn picks the position to ask an item again; math/rand by default.
	Intn func(n int) int
}

// New creates an interval lesson over the given item indexes.
func New(indexes []int, settings Settings) *LessonType {
	return &LessonType{
		settings: settings,
		queue:    append([]int(nil), indexes...),
		Intn:     rand.Intn,
	}
}

func (l *LessonType) Start() { l.sendNext() }

func (l *LessonType) SetResult(r lessontypes.Result) {
	l.Record(r)
	l.resultIndexes = append(l.resultIndexes, l.current)

	right, total := 0, 0
	for i, res := range l.Test().Results {
		if l.resultIndexes[i] == l.current {
			total++
			if res.Right {
				right++
			}
		}
	}

	// not known well enough yet: ask again a little later (never right
	// away, unless nothing else is left)
	if total < l.settings.minQuestions() || right*100 < l.settings.whenKnown()*total {
		pos := 0
		if limit := min(len(l.queue), l.settings.groupSize()-1); limit >= 1 {
			pos = 1 + l.Intn(limit)
		}
		l.queue = append(l.queue[:pos], append([]int{l.current}, l.queue[pos:]...)...)
	}
	l.sendNext()
}

// Skip asks the current item again at the end.
func (l *LessonType) Skip() {
	l.queue = append(l.queue, l.current)
	l.sendNext()
}

// CorrectLastAnswer replaces the last result; the item is asked again
// anyway, so the queue stays as it is.
func (l *LessonType) CorrectLastAnswer(r lessontypes.Result) { l.ReplaceLast(r) }

// TotalItems is the number of items asked so far plus those still queued
// (the best case).
func (l *LessonType) TotalItems() int { return len(l.queue) + l.AskedItems() }

// AskedItems is the number of different items answered so far.
func (l *LessonType) AskedItems() int {
	seen := map[int]bool{}
	for _, i := range l.resultIndexes {
		seen[i] = true
	}
	return len(seen)
}

func (l *LessonType) sendNext() {
	if len(l.queue) == 0 {
		l.Done()
		return
	}
	l.current, l.queue = l.queue[0], l.queue[1:]
	l.Ask(l.current)
}

// IntervalModule offers the lesson type as an OpenTeacher "lessonType" module.
type IntervalModule struct {
	*core.BaseModule
	manager *core.Manager
	// Settings used for new lessons; OpenTeacher's defaults unless changed.
	Settings Settings
}

// NewIntervalModule creates the module.
func NewIntervalModule() *IntervalModule {
	base := core.NewBaseModule("lessonType", "interval")
	base.SetPriority(170)
	return &IntervalModule{BaseModule: base, Settings: DefaultSettings}
}

// Name is the lesson type's name as shown to users.
func (mod *IntervalModule) Name() string { return "Interval" }

// CreateLessonType starts a lesson over the given item indexes.
func (mod *IntervalModule) CreateLessonType(indexes []int) lessontypes.LessonType {
	return New(indexes, mod.Settings)
}

func (mod *IntervalModule) Enable(ctx context.Context) error  { return mod.BaseModule.Enable(ctx) }
func (mod *IntervalModule) Disable(ctx context.Context) error { return mod.BaseModule.Disable(ctx) }

// SetManager sets the module manager reference.
func (mod *IntervalModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitIntervalModule creates and returns the module.
func InitIntervalModule() core.Module { return NewIntervalModule() }
