// Package percentscalculator computes test scores in percents. Port of
// OpenTeacher's logic/percentsCalculator.
package percentscalculator

import (
	"context"
	"math"

	"github.com/LaPingvino/recuerdo/internal/core"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
)

// Fraction is the share of right answers in a test (0 for an empty test).
func Fraction(test lessontypes.Test) float64 {
	if len(test.Results) == 0 {
		return 0
	}
	right := 0
	for _, r := range test.Results {
		if r.Right {
			right++
		}
	}
	return float64(right) / float64(len(test.Results))
}

// Percents is the score of a test in whole percents (rounded half up).
func Percents(test lessontypes.Test) int {
	return round(Fraction(test) * 100)
}

// AveragePercents is the average score of the tests, rounded.
func AveragePercents(tests []lessontypes.Test) int {
	if len(tests) == 0 {
		return 0
	}
	sum := 0
	for _, t := range tests {
		sum += Percents(t)
	}
	return round(float64(sum) / float64(len(tests)))
}

// round rounds half away from zero, as Python 2's round does.
func round(x float64) int { return int(math.Round(x)) }

// PercentsCalculatorModule offers the calculations as an OpenTeacher
// "percentsCalculator" module.
type PercentsCalculatorModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewPercentsCalculatorModule creates the module.
func NewPercentsCalculatorModule() *PercentsCalculatorModule {
	return &PercentsCalculatorModule{BaseModule: core.NewBaseModule("percentsCalculator", "percents-calculator")}
}

func (mod *PercentsCalculatorModule) CalculatePercents(test lessontypes.Test) int {
	return Percents(test)
}
func (mod *PercentsCalculatorModule) CalculateAveragePercents(tests []lessontypes.Test) int {
	return AveragePercents(tests)
}
func (mod *PercentsCalculatorModule) Enable(ctx context.Context) error {
	return mod.BaseModule.Enable(ctx)
}
func (mod *PercentsCalculatorModule) Disable(ctx context.Context) error {
	return mod.BaseModule.Disable(ctx)
}

// SetManager sets the module manager reference.
func (mod *PercentsCalculatorModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitPercentsCalculatorModule creates and returns the module.
func InitPercentsCalculatorModule() core.Module { return NewPercentsCalculatorModule() }
