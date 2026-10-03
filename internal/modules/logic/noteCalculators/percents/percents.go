// Package percents gives the score in percents as the grade ("73%"). Port
// of OpenTeacher's logic/noteCalculators/percents.
package percents

import (
	"strconv"

	"github.com/LaPingvino/recuerdo/internal/core"
	notecalculators "github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators"
)

// Convert writes a percentage.
func Convert(percents int) string { return strconv.Itoa(percents) + "%" }

// Note and AverageNote grade a test and the average of tests.
var Note, AverageNote = notecalculators.FromPercents(Convert)

type PercentsNoteCalculatorModule = notecalculators.Module

// NewPercentsNoteCalculatorModule creates the module.
func NewPercentsNoteCalculatorModule() *PercentsNoteCalculatorModule {
	return notecalculators.NewModule("percents", "Percents", 735, Note, AverageNote)
}

// InitPercentsNoteCalculatorModule creates and returns the module.
func InitPercentsNoteCalculatorModule() core.Module { return NewPercentsNoteCalculatorModule() }
