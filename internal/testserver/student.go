package testserver

import (
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/checker"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/composer"
	"github.com/LaPingvino/recuerdo/internal/richtext"
)

// StudentItem is a question as a student sees it in a test.
type StudentItem struct {
	ID           int    `json:"id"`
	Question     string `json:"question"`     // composed, as typed (OpenTeacher's notation)
	QuestionHTML string `json:"questionHtml"` // with its markup and formulas, safe to show
	Comment      string `json:"comment,omitempty"`
	// Math: the answer is a formula, so the page offers the formula
	// builder (it tells the kind of answer, not the answer)
	Math bool `json:"math,omitempty"`
}

// StudentTest is a test as a student gets it: the questions, not the
// answers, and how far the student is with it.
type StudentTest struct {
	ID               int64         `json:"id"`
	Title            string        `json:"title"`
	Open             bool          `json:"open"`
	QuestionLanguage string        `json:"questionLanguage,omitempty"`
	AnswerLanguage   string        `json:"answerLanguage,omitempty"`
	Items            []StudentItem `json:"items"`
	HandedIn         bool          `json:"handedIn"`
	Published        bool          `json:"published"`
}

// StudentView is the test for one student.
func (s *Store) StudentView(t Test, studentID int64) StudentTest {
	v := StudentTest{ID: t.ID, Title: t.Title, Open: t.Open, QuestionLanguage: t.List.QuestionLanguage,
		AnswerLanguage: t.List.AnswerLanguage, Items: []StudentItem{}}
	for _, it := range t.List.Items {
		q := composer.Compose(checker.StoredAnswers(it.Questions))
		math := false
		for _, a := range it.Answers {
			math = math || richtext.HasMath(a)
		}
		v.Items = append(v.Items, StudentItem{ID: it.ID, Question: richtext.Plain(q),
			QuestionHTML: richtext.Sanitize(q), Comment: it.Comment, Math: math})
	}
	if r, err := s.ResultOf(t.ID, studentID); err == nil {
		v.HandedIn, v.Published = true, r.Published
	}
	return v
}
