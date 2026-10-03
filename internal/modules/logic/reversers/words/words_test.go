package words

import (
	"reflect"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

func TestReverse(t *testing.T) {
	list := lesson.WordList{
		QuestionLanguage: "Dutch",
		AnswerLanguage:   "English",
		Items: []lesson.WordItem{
			{ID: 0, Questions: []string{"een"}, Answers: []string{"one"}},
			{ID: 1, Questions: []string{"twee"}},
		},
	}
	Reverse(&list)
	want := lesson.WordList{
		QuestionLanguage: "English",
		AnswerLanguage:   "Dutch",
		Items: []lesson.WordItem{
			{ID: 0, Questions: []string{"one"}, Answers: []string{"een"}},
			{ID: 1, Answers: []string{"twee"}},
		},
	}
	if !reflect.DeepEqual(list, want) {
		t.Errorf("got %+v\nwant %+v", list, want)
	}
	if NewWordsReverserModule().Type() != "reverser" {
		t.Error("module type")
	}
}
