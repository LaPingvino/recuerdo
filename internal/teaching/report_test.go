package teaching

import (
	"testing"
	"time"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

func TestReport(t *testing.T) {
	clock := time.Date(2026, 10, 3, 9, 0, 0, 0, time.UTC)
	now := func() time.Time { clock = clock.Add(4 * time.Second); return clock }

	s := New(list, Options{LessonType: Smart, Now: now})
	wrong := map[string]int{"een": 2}
	practise(t, s, func(item lesson.WordItem) string {
		if wrong[item.Questions[0]] > 0 {
			wrong[item.Questions[0]]--
			return "uno"
		}
		return rightAnswer(item)
	})

	r := s.Report()
	if len(r.Rows) != len(s.Test().Results) || !r.Finished {
		t.Fatalf("report %+v", r)
	}
	if got := r.Rows[1]; got.Question != "een" || got.Answer != "one, a" || got.Given != "uno" || got.Right {
		t.Errorf("row 1 = %+v", got)
	}
	if r.MostDoneWrong != "een" {
		t.Errorf("most done wrong = %q", r.MostDoneWrong)
	}
	// each answer took 4 seconds on the fake clock
	if want := time.Duration(len(r.Rows)) * 4 * time.Second; r.ThinkingTime != want {
		t.Errorf("thinking time %v, want %v", r.ThinkingTime, want)
	}
}

func TestThinkingTimeText(t *testing.T) {
	for d, want := range map[time.Duration]string{
		time.Second:                     "Total thinking time: 1 second",
		45 * time.Second:                "Total thinking time: 45 seconds",
		179 * time.Second:               "Total thinking time: 179 seconds",
		3 * time.Minute:                 "Total thinking time: 3 minutes",
		10*time.Minute + 31*time.Second: "Total thinking time: 11 minutes",
	} {
		if got := (Report{ThinkingTime: d}).ThinkingTimeText(); got != want {
			t.Errorf("%v: %q, want %q", d, got, want)
		}
	}
}

func TestReportNothingWrong(t *testing.T) {
	s := New(list, Options{})
	practise(t, s, rightAnswer)
	if r := s.Report(); r.MostDoneWrong != "" {
		t.Errorf("most done wrong = %q with everything right", r.MostDoneWrong)
	}
}
