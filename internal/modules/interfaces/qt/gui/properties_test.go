package gui

import (
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

func TestApplyProperties(t *testing.T) {
	list := lesson.WordList{Title: "Dieren", QuestionLanguage: "Dutch"}
	same := map[string]interface{}{"name": "Dieren", "questionLanguage": "Dutch", "answerLanguage": ""}
	if applyProperties(&list, same) {
		t.Error("nothing changed, but reported a change")
	}
	if !applyProperties(&list, map[string]interface{}{"name": "", "questionLanguage": "Dutch", "answerLanguage": "English"}) ||
		list.Title != "Dieren" || list.AnswerLanguage != "English" {
		t.Errorf("an empty title must not clear it, the language must be set: %+v", list)
	}
	if !applyProperties(&list, map[string]interface{}{"name": "Huisdieren", "questionLanguage": "Dutch", "answerLanguage": "English"}) ||
		list.Title != "Huisdieren" {
		t.Errorf("title: %+v", list)
	}
}
