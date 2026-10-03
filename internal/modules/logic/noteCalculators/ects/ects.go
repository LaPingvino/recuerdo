// Package ects grades in the ECTS notation. Port of OpenTeacher's
// logic/noteCalculators/ects.
package ects

import (
	"github.com/LaPingvino/recuerdo/internal/core"
	notecalculators "github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators"
)

var (
	bounds = []int{30, 40, 50, 55, 60, 70}
	grades = []string{"F", "FX", "E", "D", "C", "B", "A"}
)

// Convert turns a percentage into a grade.
func Convert(percents int) string { return grades[notecalculators.Bisect(bounds, percents)] }

// Note and AverageNote grade a test and the average of tests.
var Note, AverageNote = notecalculators.FromPercents(Convert)

type ECTSNoteCalculatorModule = notecalculators.Module

// NewECTSNoteCalculatorModule creates the module.
func NewECTSNoteCalculatorModule() *ECTSNoteCalculatorModule {
	return notecalculators.NewModule("ects", "ECTS", 935, Note, AverageNote)
}

// InitECTSNoteCalculatorModule creates and returns the module.
func InitECTSNoteCalculatorModule() core.Module { return NewECTSNoteCalculatorModule() }
