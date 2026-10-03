// Package dutch grades on the Dutch 1-10 scale ("7,3", and "10" for a
// perfect score). Port of OpenTeacher's logic/noteCalculators/dutch.
package dutch

import (
	"strconv"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	notecalculators "github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators"
	percentscalculator "github.com/LaPingvino/recuerdo/internal/modules/logic/percentsCalculator"
)

func grade(test lessontypes.Test) float64 { return percentscalculator.Fraction(test)*9 + 1 }

// format writes a grade with one decimal and a comma; 10 is written "10"
// ("10,0" would suggest 10,8 exists).
func format(g float64) string {
	if g == 10 {
		return "10"
	}
	return strings.Replace(strconv.FormatFloat(g, 'f', 1, 64), ".", ",", 1)
}

// Note is the Dutch grade for a test.
func Note(test lessontypes.Test) string { return format(grade(test)) }

// AverageNote is the average Dutch grade of the tests.
func AverageNote(tests []lessontypes.Test) string {
	sum := 0.0
	for _, t := range tests {
		sum += grade(t)
	}
	return format(sum / float64(len(tests)))
}

type DutchNoteCalculatorModule = notecalculators.Module

// NewDutchNoteCalculatorModule creates the module.
func NewDutchNoteCalculatorModule() *DutchNoteCalculatorModule {
	return notecalculators.NewModule("dutch", "Dutch", 935, Note, AverageNote)
}

// InitDutchNoteCalculatorModule creates and returns the module.
func InitDutchNoteCalculatorModule() core.Module { return NewDutchNoteCalculatorModule() }
