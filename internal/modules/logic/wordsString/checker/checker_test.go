package checker

import (
	"reflect"
	"testing"
)

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

func TestStoredAnswers(t *testing.T) {
	cases := []struct {
		stored []string
		want   [][]string
	}{
		{[]string{"in", "op", "bij"}, [][]string{{"in", "op", "bij"}}},
		{[]string{"1. in, op 2. tijdens"}, [][]string{{"in", "op"}, {"tijdens"}}},
		{[]string{"Am 20. Mai"}, [][]string{{"Am 20. Mai"}}},
		{[]string{" ", ""}, [][]string{}},
	}
	for _, c := range cases {
		if got := StoredAnswers(c.stored); !reflect.DeepEqual(got, c.want) {
			t.Errorf("StoredAnswers(%q) = %q, want %q", c.stored, got, c.want)
		}
	}
}

func TestCorrectText(t *testing.T) {
	cases := []struct {
		typed         string
		stored        []string
		caseSensitive bool
		right         bool
	}{
		{"house", []string{"house"}, false, true},
		{"House", []string{"house"}, false, true},
		{"House", []string{"house"}, true, false},
		{"bij, op, in", []string{"in", "op", "bij"}, false, true},
		{"in", []string{"in", "op", "bij"}, false, true}, // one alternative is enough
		{"in, xyz", []string{"in", "op", "bij"}, false, false},
		{"1. in, op, bij 2. tijdens", []string{"1. in, op, bij 2. tijdens"}, false, true},
		{"in, tijdens, bij", []string{"1. in, op, bij 2. tijdens"}, false, true},
		{"in", []string{"1. in, op, bij 2. tijdens"}, false, false},
		{"", []string{"house"}, false, false},
	}
	for _, c := range cases {
		if got := CorrectText(c.typed, c.stored, c.caseSensitive); got != c.right {
			t.Errorf("CorrectText(%q, %q, %v) = %v, want %v", c.typed, c.stored, c.caseSensitive, got, c.right)
		}
	}
}
