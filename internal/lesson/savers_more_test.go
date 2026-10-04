package lesson

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func sampleLesson() *LessonData {
	d := NewLessonData()
	d.List.Title = "Dieren"
	d.List.QuestionLanguage, d.List.AnswerLanguage = "Nederlands", "English"
	d.List.Items = []WordItem{
		{ID: 0, Questions: []string{"hond"}, Answers: []string{"dog"}},
		{ID: 1, Questions: []string{"kat"}, Answers: []string{"cat", "puss"}, Comment: "a; b"},
	}
	d.List.Tests = []Test{{Results: []TestResult{{ItemID: 0, Result: "right"}}}}
	return d
}

// Saving in a format and loading it again keeps the words and languages.
func TestSaversRoundTrip(t *testing.T) {
	for _, ext := range []string{".otwd", ".wrts"} {
		t.Run(ext, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "dieren"+ext)
			in := sampleLesson()
			if err := NewFileSaver().SaveFile(in, path); err != nil {
				t.Fatal(err)
			}
			out, err := NewFileLoader().LoadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if out.List.Title != "Dieren" || out.List.QuestionLanguage != "Nederlands" || out.List.AnswerLanguage != "English" {
				t.Errorf("list: %q %q %q", out.List.Title, out.List.QuestionLanguage, out.List.AnswerLanguage)
			}
			if len(out.List.Items) != 2 || !reflect.DeepEqual(out.List.Items[1].Answers, []string{"cat", "puss"}) {
				t.Errorf("items: %+v", out.List.Items)
			}
			if ext == ".otwd" && (len(out.List.Tests) != 1 || out.List.Items[1].Comment != "a; b") {
				t.Errorf("otwd should keep comments and results: %+v", out.List)
			}
		})
	}
}

func TestSaveSYLK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dieren.slk")
	if err := NewFileSaver().SaveFile(sampleLesson(), path); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	s := string(b)
	for _, want := range []string{"ID;P", `C;X1;Y1;K"Dieren"`, `C;X1;Y3;K"Questions"`, `C;X2;Y5;K"cat, puss"`, `C;X3;Y5;K"a;; b"`, "\nE\n"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in:\n%s", want, s)
		}
	}
}
