package teaching

import (
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	notecalculatorchooser "github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculatorChooser"
)

// Notations lists the grade notations; DefaultNotation is used when none
// is chosen.
var (
	Notations       = notecalculatorchooser.Names()
	DefaultNotation = notecalculatorchooser.Default
)

// Grade is the grade for a test in the given notation ("" for a test
// without answers).
func Grade(notation string, test lessontypes.Test) string {
	if len(test.Results) == 0 {
		return ""
	}
	return notecalculatorchooser.Choose(notation).CalculateNote(test)
}

// AverageGrade is the average grade of the tests that have answers.
func AverageGrade(notation string, tests []lessontypes.Test) string {
	var answered []lessontypes.Test
	for _, t := range tests {
		if len(t.Results) > 0 {
			answered = append(answered, t)
		}
	}
	if len(answered) == 0 {
		return ""
	}
	return notecalculatorchooser.Choose(notation).CalculateAverageNote(answered)
}
