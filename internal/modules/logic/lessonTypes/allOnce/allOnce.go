// Package allonce asks every item once, in the order given. Port of
// OpenTeacher's logic/lessonTypes/allOnce.
package allonce

import (
	"context"

	"github.com/LaPingvino/recuerdo/internal/core"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
)

// LessonType asks each index once; skipped items move to the end.
type LessonType struct {
	lessontypes.Base
	indexes []int
	asked   int
}

// New creates an "all once" lesson over the given item indexes.
func New(indexes []int) *LessonType {
	return &LessonType{indexes: append([]int(nil), indexes...)}
}

func (l *LessonType) Start() { l.sendNext() }

func (l *LessonType) SetResult(r lessontypes.Result) {
	l.Record(r)
	l.asked++
	l.sendNext()
}

func (l *LessonType) Skip() {
	if l.asked < len(l.indexes) {
		skipped := l.indexes[l.asked]
		l.indexes = append(append(l.indexes[:l.asked:l.asked], l.indexes[l.asked+1:]...), skipped)
	}
	l.sendNext()
}

func (l *LessonType) CorrectLastAnswer(r lessontypes.Result) { l.ReplaceLast(r) }
func (l *LessonType) TotalItems() int                        { return len(l.indexes) }
func (l *LessonType) AskedItems() int                        { return l.asked }

func (l *LessonType) sendNext() {
	if l.asked >= len(l.indexes) {
		l.Done()
		return
	}
	l.Ask(l.indexes[l.asked])
}

// AllOnceModule offers the lesson type as an OpenTeacher "lessonType" module.
type AllOnceModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewAllOnceModule creates the module.
func NewAllOnceModule() *AllOnceModule {
	base := core.NewBaseModule("lessonType", "all-once")
	base.SetPriority(140)
	return &AllOnceModule{BaseModule: base}
}

// DisplayName is the lesson type's name as shown to users (Name is the module identifier).
func (mod *AllOnceModule) DisplayName() string { return "All once" }

// CreateLessonType starts a lesson over the given item indexes.
func (mod *AllOnceModule) CreateLessonType(indexes []int) lessontypes.LessonType {
	return New(indexes)
}

func (mod *AllOnceModule) Enable(ctx context.Context) error  { return mod.BaseModule.Enable(ctx) }
func (mod *AllOnceModule) Disable(ctx context.Context) error { return mod.BaseModule.Disable(ctx) }

// SetManager sets the module manager reference.
func (mod *AllOnceModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitAllOnceModule creates and returns the module.
func InitAllOnceModule() core.Module { return NewAllOnceModule() }
