package checker

import "testing"

// Ported from OpenTeacher's logic/wordsString/checkerTest.
var (
	// Latin "in" + ablative and + accusative
	word1     = [][]string{{"in", "op", "bij"}, {"tijdens"}}
	word2     = [][]string{{"naar(binnen)", "in"}, {"tot", "jegens"}}
	emptyWord = [][]string{}
)

func TestCorrect(t *testing.T) {
	cases := []struct {
		name    string
		given   [][]string
		answers [][]string
		right   bool
	}{
		{"SingleRightAnswer", [][]string{{"in"}}, word1, false},
		{"MultipleRightAnswers", [][]string{{"in", "tijdens", "bij"}}, word1, true},
		{"WrongAnswersNextToRightOnes", [][]string{{"in", "tijdens", "opp"}}, word1, false},
		{"SingleWrongAnswer", [][]string{{"opp"}}, word1, false},
		{"EmptyAnswer", [][]string{}, word1, false},
		{"EmptyAnswerWithEmptyWord", [][]string{}, emptyWord, true},
		{"FullAnswer", [][]string{{"in", "op", "bij"}, {"tijdens"}}, word1, true},
		{"FullAnswerWithExtraWrongWords", [][]string{{"in", "op", "bij"}, {"tijdens", "gelijktijdig met"}}, word1, false},
		{"WordsInWeirdOrder", [][]string{{"naar(binnen)", "jegens", "tot"}}, word2, true},
		{"FullNotationAndDontIncludeAnNonObligatoryWord", [][]string{{"in", "naar(binnen)"}, {"jegens"}}, word2, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Correct(c.given, c.answers); got != c.right {
				t.Errorf("Correct(%q, %q) = %v, want %v", c.given, c.answers, got, c.right)
			}
		})
	}
}

func TestForItemID(t *testing.T) {
	const theID = 342
	r := NewWordsStringCheckerModule().Check([][]string{{"a"}}, theID, [][]string{{"a"}})
	if r.ItemID != theID || !r.Right {
		t.Errorf("Check = %+v", r)
	}
}
