package teaching

import (
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
)

func TestSessionGrade(t *testing.T) {
	s := New(list, Options{LessonType: Smart})
	wrongOnce := true
	practise(t, s, func(item lesson.WordItem) string {
		if item.Questions[0] == "een" && wrongOnce {
			wrongOnce = false
			return "uno"
		}
		return rightAnswer(item)
	})
	// 3 of 4 answers right, as OpenTeacher grades the repeated item too
	test := s.Test()
	for notation, want := range map[string]string{"Dutch": "7,8", "Percents": "75%", "American": "C", "": "7,8"} {
		if got := Grade(notation, test); got != want {
			t.Errorf("Grade(%q) = %q, want %q", notation, got, want)
		}
	}
	if got := AverageGrade("Percents", []lessontypes.Test{test, {}, test}); got != "75%" {
		t.Errorf("average = %q (tests without answers must not count)", got)
	}
	if Grade("Dutch", lessontypes.Test{}) != "" || AverageGrade("Dutch", nil) != "" {
		t.Error("no answers should give no grade")
	}
}
