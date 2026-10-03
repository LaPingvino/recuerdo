package notecalculators_test

import (
	"testing"

	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	notecalculators "github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/american"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/dutch"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/ects"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/french"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/german"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/noteCalculators/percents"
	percentscalculator "github.com/LaPingvino/recuerdo/internal/modules/logic/percentsCalculator"
)

func test(right, total int) lessontypes.Test {
	var t lessontypes.Test
	for i := 0; i < total; i++ {
		t.Results = append(t.Results, lessontypes.Result{ItemID: i, Right: i < right})
	}
	return t
}

var all = []notecalculators.Calculator{
	dutch.NewDutchNoteCalculatorModule(),
	american.NewAmericanNoteCalculatorModule(),
	ects.NewECTSNoteCalculatorModule(),
	german.NewGermanNoteCalculatorModule(),
	french.NewFrenchNoteCalculatorModule(),
	percents.NewPercentsNoteCalculatorModule(),
}

// Expected values computed with OpenTeacher's Python formulas (Python 2
// rounding: half away from zero).
var table = []struct {
	right, total, percents                            int
	dutch, american, ects, german, french, percentage string
}{
	{0, 5, 0, "1,0", "F", "F", "6", "0", "0%"},
	{1, 3, 33, "4,0", "F", "FX", "5", "7", "33%"},
	{2, 3, 67, "7,0", "D+", "B", "3", "13", "67%"},
	{5, 8, 63, "6,6", "D", "B", "4", "13", "63%"},
	{3, 4, 75, "7,8", "C", "A", "3", "15", "75%"},
	{7, 10, 70, "7,3", "C-", "A", "3", "14", "70%"},
	{13, 20, 65, "6,9", "D", "B", "4", "13", "65%"},
	{1, 1, 100, "10", "A+", "A", "1", "20", "100%"},
	{17, 20, 85, "8,6", "B", "A", "2", "17", "85%"},
	{9, 10, 90, "9,1", "A-", "A", "2", "18", "90%"},
	{19, 20, 95, "9,5", "A", "A", "1", "19", "95%"},
	{1, 8, 13, "2,1", "F", "F", "6", "3", "13%"},
	{1, 2, 50, "5,5", "F", "D", "4", "10", "50%"},
}

func TestGradesMatchOpenTeacher(t *testing.T) {
	for _, c := range table {
		tt := test(c.right, c.total)
		got := []string{dutch.Note(tt), american.Note(tt), ects.Note(tt), german.Note(tt), french.Note(tt), percents.Note(tt)}
		want := []string{c.dutch, c.american, c.ects, c.german, c.french, c.percentage}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("%d/%d with %s: %q, want %q", c.right, c.total, all[i].DisplayName(), got[i], want[i])
			}
		}
		if p := percentscalculator.Percents(tt); p != c.percents {
			t.Errorf("%d/%d: %d%%, want %d%%", c.right, c.total, p, c.percents)
		}
	}
}

func TestAverageNotesMatchOpenTeacher(t *testing.T) {
	tests := []lessontypes.Test{test(0, 5), test(5, 8), test(8, 8)}
	want := map[string]string{"Dutch": "5,9", "American": "F", "ECTS": "D", "German": "4", "French": "11", "Percents": "54%"}
	for _, c := range all {
		if got := c.CalculateAverageNote(tests); got != want[c.DisplayName()] {
			t.Errorf("%s average: %q, want %q", c.DisplayName(), got, want[c.DisplayName()])
		}
	}
	if got := percentscalculator.AveragePercents(tests); got != 54 {
		t.Errorf("average percents %d", got)
	}
}

// Ported from OpenTeacher's logic/noteCalculators/test: notes are
// non-empty strings and everything wrong differs from everything right.
func TestCalculateNote(t *testing.T) {
	worst, best := test(0, 5), test(4, 4)
	for _, c := range all {
		n1, n3 := c.CalculateNote(worst), c.CalculateNote(best)
		if n1 == "" || n3 == "" || n1 == n3 {
			t.Errorf("%s: worst %q, best %q", c.DisplayName(), n1, n3)
		}
		if c.CalculateAverageNote([]lessontypes.Test{worst, best}) == "" {
			t.Errorf("%s: empty average", c.DisplayName())
		}
	}
}
