package teaching

import (
	"fmt"
	"math"
	"time"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/checker"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/composer"
)

// Row is one answer in a results report.
type Row struct {
	Question, Answer, Given string
	Right                   bool
}

// Report is what OpenTeacher's test viewer shows about a test of a word
// list: every answer, the word most often answered wrong, whether the
// test was finished and how long the user spent thinking.
type Report struct {
	Rows          []Row
	MostDoneWrong string // "" if nothing was answered wrong
	Finished      bool
	ThinkingTime  time.Duration
}

// NewReport builds the report for a test of list; result ItemIDs are
// indexes into list.Items.
func NewReport(list lesson.WordList, test lessontypes.Test) Report {
	r := Report{Finished: test.Finished}
	wrong := make(map[int]int)
	for _, res := range test.Results {
		row := Row{Given: res.GivenAnswer, Right: res.Right}
		if res.ItemID >= 0 && res.ItemID < len(list.Items) {
			item := list.Items[res.ItemID]
			row.Question = compose(item.Questions)
			row.Answer = compose(item.Answers)
		}
		r.Rows = append(r.Rows, row)
		if !res.Right {
			wrong[res.ItemID]++
		}
		if !res.Start.IsZero() && res.End.After(res.Start) {
			r.ThinkingTime += res.End.Sub(res.Start)
		}
	}

	// in list order, the first of the items wrong most often
	most := 0
	for i, item := range list.Items {
		if wrong[i] > most {
			most = wrong[i]
			r.MostDoneWrong = compose(item.Questions)
		}
	}
	return r
}

// Report builds the results report for this session.
func (s *Session) Report() Report { return NewReport(s.list, s.Test()) }

// ThinkingTimeText describes the thinking time as OpenTeacher does: in
// seconds below three minutes, else in (rounded) minutes.
func (r Report) ThinkingTimeText() string {
	seconds := int(math.Round(r.ThinkingTime.Seconds()))
	switch {
	case seconds == 1:
		return "Total thinking time: 1 second"
	case seconds < 180:
		return fmt.Sprintf("Total thinking time: %d seconds", seconds)
	}
	return fmt.Sprintf("Total thinking time: %d minutes", int(math.Round(float64(seconds)/60)))
}

func compose(stored []string) string {
	return composer.Compose(checker.StoredAnswers(stored))
}
