// Package french grades on the French 0-20 scale. Port of OpenTeacher's
// logic/noteCalculators/french.
package french

import (
	"math"
	"strconv"

	"github.com/LaPingvino/recuerdo/internal/core"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	notecalculators "github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators"
	percentscalculator "github.com/LaPingvino/recuerdo/internal/modules/logic/percentsCalculator"
)

func grade(test lessontypes.Test) int {
	return int(math.Round(percentscalculator.Fraction(test) * 20))
}

// Note is the French grade for a test.
func Note(test lessontypes.Test) string { return strconv.Itoa(grade(test)) }

// AverageNote is the average of the tests' grades, rounded down (as
// OpenTeacher does).
func AverageNote(tests []lessontypes.Test) string {
	sum := 0
	for _, t := range tests {
		sum += grade(t)
	}
	return strconv.Itoa(int(float64(sum) / float64(len(tests))))
}

type FrenchNoteCalculatorModule = notecalculators.Module

// NewFrenchNoteCalculatorModule creates the module.
func NewFrenchNoteCalculatorModule() *FrenchNoteCalculatorModule {
	return notecalculators.NewModule("french", "French", 935, Note, AverageNote)
}

// InitFrenchNoteCalculatorModule creates and returns the module.
func InitFrenchNoteCalculatorModule() core.Module { return NewFrenchNoteCalculatorModule() }
