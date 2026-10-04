package teaching

import (
	"reflect"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

var list = lesson.WordList{
	QuestionLanguage: "Dutch",
	AnswerLanguage:   "English",
	Items: []lesson.WordItem{
		{ID: 0, Questions: []string{"twee"}, Answers: []string{"two"}},
		{ID: 1, Questions: []string{"een"}, Answers: []string{"one", "a"}},
		{ID: 2, Questions: []string{"drie"}, Answers: []string{"three"}},
	},
}

// practise answers every question with answer(item) and returns the
// questions asked.
func practise(t *testing.T, s *Session, answer func(lesson.WordItem) string) []string {
	t.Helper()
	var asked []string
	s.Start()
	for !s.Done() {
		item, _, ok := s.Current()
		if !ok || len(asked) > 30 {
			t.Fatalf("no current item / no end; asked %v", asked)
		}
		asked = append(asked, item.Questions[0])
		s.Answer(answer(item))
		s.Next()
	}
	return asked
}

func rightAnswer(item lesson.WordItem) string { return item.Answers[0] }

func TestOrders(t *testing.T) {
	cases := map[string][]string{
		AsEntered: {"twee", "een", "drie"},
		Reversed:  {"drie", "een", "twee"},
		Sorted:    {"drie", "een", "twee"},
		Random:    {"drie", "een", "twee"}, // with the fixed shuffle below
	}
	swapEnds := func(n int, swap func(i, j int)) { swap(0, n-1) }
	for order, want := range cases {
		s := New(list, Options{Order: order, Shuffle: swapEnds})
		if got := practise(t, s, rightAnswer); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: asked %v, want %v", order, got, want)
		}
	}
}

func TestAskAnswersAndListUntouched(t *testing.T) {
	s := New(list, Options{AskAnswers: true})
	asked := practise(t, s, rightAnswer)
	if want := []string{"two", "one", "three"}; !reflect.DeepEqual(asked, want) {
		t.Errorf("asked %v, want %v", asked, want)
	}
	if list.Items[0].Questions[0] != "twee" || list.QuestionLanguage != "Dutch" {
		t.Error("New changed the caller's list")
	}
}

func TestSmartRepeatsWrongAnswerAndScore(t *testing.T) {
	s := New(list, Options{LessonType: Smart})
	wrongOnce := true
	asked := practise(t, s, func(item lesson.WordItem) string {
		if item.Questions[0] == "een" && wrongOnce {
			wrongOnce = false
			return "uno"
		}
		return rightAnswer(item)
	})
	if want := []string{"twee", "een", "drie", "een"}; !reflect.DeepEqual(asked, want) {
		t.Errorf("asked %v, want %v", asked, want)
	}
	if right, answered := s.Score(); right != 3 || answered != 4 {
		t.Errorf("score %d/%d, want 3/4", right, answered)
	}
}

func TestIntervalAsksUntilKnown(t *testing.T) {
	s := New(list, Options{LessonType: Interval, Intn: func(int) int { return 0 }})
	asked := practise(t, s, rightAnswer)
	if len(asked) != 6 {
		t.Errorf("interval asked %v; every word twice expected", asked)
	}
}

func TestAnswerUsesOpenTeacherRules(t *testing.T) {
	s := New(list, Options{})
	s.Start()
	s.Answer("x")
	s.Next() // twee: wrong
	if a := s.Answer("A"); !a.Right || a.Correct != "one, a" {
		t.Errorf("answer for een = %+v", a)
	}
	asked, total := s.Progress()
	if asked != 1 || total != 3 {
		t.Errorf("progress %d/%d before Next, want 1/3", asked, total)
	}
}

func TestEmptyList(t *testing.T) {
	s := New(lesson.WordList{}, Options{LessonType: Smart})
	s.Start()
	if !s.Done() {
		t.Error("an empty list should be done at once")
	}
	if _, _, ok := s.Current(); ok {
		t.Error("no current item expected")
	}
}

func TestWordChoices(t *testing.T) {
	list := lesson.WordList{
		Items: []lesson.WordItem{
			{ID: 0, Questions: []string{"een"}, Answers: []string{"one"}},
			{ID: 1, Questions: []string{"twee"}, Answers: []string{"two"}},
			{ID: 2, Questions: []string{"drie"}, Answers: []string{"three"}},
		},
		Tests: []lesson.Test{{Results: []lesson.TestResult{
			{ItemID: 0, Result: "right"}, {ItemID: 1, Result: "wrong"}, {ItemID: 1, Result: "wrong"},
			{ItemID: 2, Result: "wrong"}, {ItemID: 2, Result: "right"},
		}}},
	}
	for choice, want := range map[string]int{AllWords: 3, HardWords: 1, NeverRight: 1, "": 3} {
		s := New(list, Options{Words: choice})
		if _, total := s.Progress(); total != want {
			t.Errorf("%q: %d words, want %d", choice, total, want)
		}
	}
}
