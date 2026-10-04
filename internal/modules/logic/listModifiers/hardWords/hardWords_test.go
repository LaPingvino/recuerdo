package hardwords

import (
	"reflect"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

func results(rs ...string) lesson.Test {
	var t lesson.Test
	for _, r := range rs {
		t.Results = append(t.Results, lesson.TestResult{ItemID: 0, Result: r})
	}
	return t
}

// OpenTeacher's hardWordsTest cases.
func TestModifyList(t *testing.T) {
	word := []lesson.WordItem{{ID: 0}}
	for name, c := range map[string]struct {
		tests []lesson.Test
		want  []int
	}{
		"without results":     {nil, []int{0}},
		"multiple tests":      {[]lesson.Test{results("right", "wrong", "right"), results("wrong", "wrong")}, []int{0}},
		"mostly right":        {[]lesson.Test{results("right", "right", "wrong")}, nil},
		"exactly 50% is fine": {[]lesson.Test{results("right", "wrong")}, nil},
	} {
		got := ModifyList([]int{0}, lesson.WordList{Items: word, Tests: c.tests})
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: %v, want %v", name, got, c.want)
		}
	}
	// several words: only the hard ones stay, in order
	list := lesson.WordList{
		Items: []lesson.WordItem{{ID: 0}, {ID: 1}, {ID: 2}},
		Tests: []lesson.Test{{Results: []lesson.TestResult{
			{ItemID: 0, Result: "right"}, {ItemID: 1, Result: "wrong"}, {ItemID: 2, Result: "wrong"}, {ItemID: 2, Result: "right"},
		}}},
	}
	if got := ModifyList([]int{0, 1, 2}, list); !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("several words: %v", got)
	}
}
