// Package american grades in the American notation. Port of OpenTeacher's
// logic/noteCalculators/american.
package american

import (
	"github.com/LaPingvino/recuerdo/internal/core"
	notecalculators "github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators"
)

var (
	bounds = []int{60, 63, 67, 70, 73, 77, 80, 83, 87, 90, 93, 97}
	grades = []string{"F", "D-", "D", "D+", "C-", "C", "C+", "B-", "B", "B+", "A-", "A", "A+"}
)

// Convert turns a percentage into a grade.
func Convert(percents int) string { return grades[notecalculators.Bisect(bounds, percents)] }

// Note and AverageNote grade a test and the average of tests.
var Note, AverageNote = notecalculators.FromPercents(Convert)

type AmericanNoteCalculatorModule = notecalculators.Module

// NewAmericanNoteCalculatorModule creates the module.
func NewAmericanNoteCalculatorModule() *AmericanNoteCalculatorModule {
	return notecalculators.NewModule("american", "American", 935, Note, AverageNote)
}

// InitAmericanNoteCalculatorModule creates and returns the module.
func InitAmericanNoteCalculatorModule() core.Module { return NewAmericanNoteCalculatorModule() }
