// Package german grades in the German notation. Port of OpenTeacher's
// logic/noteCalculators/german.
package german

import (
	"github.com/LaPingvino/recuerdo/internal/core"
	notecalculators "github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators"
)

var (
	bounds = []int{30, 50, 67, 81, 92}
	grades = []string{"6", "5", "4", "3", "2", "1"}
)

// Convert turns a percentage into a grade.
func Convert(percents int) string { return grades[notecalculators.Bisect(bounds, percents)] }

// Note and AverageNote grade a test and the average of tests.
var Note, AverageNote = notecalculators.FromPercents(Convert)

type GermanNoteCalculatorModule = notecalculators.Module

// NewGermanNoteCalculatorModule creates the module.
func NewGermanNoteCalculatorModule() *GermanNoteCalculatorModule {
	return notecalculators.NewModule("german", "German", 935, Note, AverageNote)
}

// InitGermanNoteCalculatorModule creates and returns the module.
func InitGermanNoteCalculatorModule() core.Module { return NewGermanNoteCalculatorModule() }
