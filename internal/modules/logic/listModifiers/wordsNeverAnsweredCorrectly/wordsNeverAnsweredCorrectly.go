// Package wordsneveransweredcorrectly keeps the words never answered
// right in earlier tests. Port of OpenTeacher's
// logic/listModifiers/wordsNeverAnsweredCorrectly (which read the tests
// as plain lists of results, as only its own test used them).
package wordsneveransweredcorrectly

import (
	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// ModifyList returns the indexes of words without a right answer in list's
// tests.
func ModifyList(indexes []int, list lesson.WordList) []int {
	right := map[int]bool{}
	for _, t := range list.Tests {
		for _, r := range t.Results {
			if r.Result == "right" {
				right[r.ItemID] = true
			}
		}
	}
	var out []int
	for _, i := range indexes {
		if !right[list.Items[i].ID] {
			out = append(out, i)
		}
	}
	return out
}

// WordsNeverAnsweredCorrectlyModule offers ModifyList as an OpenTeacher
// list modifier.
type WordsNeverAnsweredCorrectlyModule struct {
	*core.BaseModule
}

// NewWordsNeverAnsweredCorrectlyModule creates the module.
func NewWordsNeverAnsweredCorrectlyModule() *WordsNeverAnsweredCorrectlyModule {
	return &WordsNeverAnsweredCorrectlyModule{BaseModule: core.NewBaseModule("listModifier", "wordsneveransweredcorrectly-module")}
}

// DisplayName is the modifier's name.
func (mod *WordsNeverAnsweredCorrectlyModule) DisplayName() string {
	return "Words never answered correctly"
}

// ModifyList keeps the words never answered right.
func (mod *WordsNeverAnsweredCorrectlyModule) ModifyList(indexes []int, list lesson.WordList) []int {
	return ModifyList(indexes, list)
}

// InitWordsNeverAnsweredCorrectlyModule creates the module.
func InitWordsNeverAnsweredCorrectlyModule() core.Module {
	return NewWordsNeverAnsweredCorrectlyModule()
}
