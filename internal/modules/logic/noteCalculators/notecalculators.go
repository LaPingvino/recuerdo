// Package notecalculators holds what OpenTeacher's note calculators share:
// each turns a test's results into a grade in some national notation
// (implementations in the dutch, american, ects, french, german and
// percents packages).
package notecalculators

import (
	"context"
	"sort"

	"github.com/LaPingvino/recuerdo/internal/core"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	percentscalculator "github.com/LaPingvino/recuerdo/internal/modules/logic/percentsCalculator"
)

// Calculator gives grades for tests.
type Calculator interface {
	DisplayName() string
	CalculateNote(test lessontypes.Test) string
	CalculateAverageNote(tests []lessontypes.Test) string
}

// Module is an OpenTeacher "noteCalculator" module around a grading scheme.
type Module struct {
	*core.BaseModule
	manager *core.Manager
	display string
	note    func(lessontypes.Test) string
	average func([]lessontypes.Test) string
}

// NewModule creates a note calculator module. average may be nil to grade
// the average percentage with the same scheme.
func NewModule(name, display string, priority int, note func(lessontypes.Test) string, average func([]lessontypes.Test) string) *Module {
	base := core.NewBaseModule("noteCalculator", name)
	base.SetPriority(priority)
	return &Module{BaseModule: base, display: display, note: note, average: average}
}

func (m *Module) DisplayName() string                        { return m.display }
func (m *Module) CalculateNote(test lessontypes.Test) string { return m.note(test) }
func (m *Module) CalculateAverageNote(tests []lessontypes.Test) string {
	if len(tests) == 0 {
		return ""
	}
	return m.average(tests)
}
func (m *Module) Enable(ctx context.Context) error  { return m.BaseModule.Enable(ctx) }
func (m *Module) Disable(ctx context.Context) error { return m.BaseModule.Disable(ctx) }

// SetManager sets the module manager reference.
func (m *Module) SetManager(manager *core.Manager) { m.manager = manager }

// FromPercents builds a scheme that grades a percentage: the grade for one
// test from its percents, the average from the average percents.
func FromPercents(convert func(percents int) string) (func(lessontypes.Test) string, func([]lessontypes.Test) string) {
	return func(t lessontypes.Test) string { return convert(percentscalculator.Percents(t)) },
		func(ts []lessontypes.Test) string { return convert(percentscalculator.AveragePercents(ts)) }
}

// Bisect is Python's bisect.bisect: the number of bounds <= x.
func Bisect(bounds []int, x int) int {
	return sort.Search(len(bounds), func(i int) bool { return bounds[i] > x })
}
