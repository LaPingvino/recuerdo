package wordsneveransweredcorrectly

import (
	"reflect"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// OpenTeacher's wordsNeverAnsweredCorrectlyTest cases, on real tests.
func TestModifyList(t *testing.T) {
	list := func(results ...string) lesson.WordList {
		var test lesson.Test
		for _, r := range results {
			test.Results = append(test.Results, lesson.TestResult{ItemID: 0, Result: r})
		}
		return lesson.WordList{Items: []lesson.WordItem{{ID: 0}}, Tests: []lesson.Test{test}}
	}
	for name, c := range map[string]struct {
		list lesson.WordList
		want []int
	}{
		"no results":  {lesson.WordList{Items: []lesson.WordItem{{ID: 0}}}, []int{0}},
		"wrong word":  {list("wrong"), []int{0}},
		"right word":  {list("right"), nil},
		"right later": {list("wrong", "wrong", "right"), nil},
	} {
		if got := ModifyList([]int{0}, c.list); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: %v, want %v", name, got, c.want)
		}
	}
}
