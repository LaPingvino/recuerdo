// Package hardwords keeps the words that went wrong more often than right
// in earlier tests, and words not tested yet. Port of OpenTeacher's
// logic/listModifiers/hardWords.
package hardwords

import (
	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// ModifyList returns the indexes of the hard words in list.
func ModifyList(indexes []int, list lesson.WordList) []int {
	right, wrong := map[int]int{}, map[int]int{}
	for _, t := range list.Tests {
		for _, r := range t.Results {
			if r.Result == "wrong" {
				wrong[r.ItemID]++
			} else {
				right[r.ItemID]++
			}
		}
	}
	var out []int
	for _, i := range indexes {
		id := list.Items[i].ID
		total := right[id] + wrong[id]
		if total == 0 || float64(wrong[id]) > float64(total)/2 {
			out = append(out, i)
		}
	}
	return out
}

// HardWordsModule offers ModifyList as an OpenTeacher list modifier.
type HardWordsModule struct {
	*core.BaseModule
}

// NewHardWordsModule creates the module.
func NewHardWordsModule() *HardWordsModule {
	return &HardWordsModule{BaseModule: core.NewBaseModule("listModifier", "hardwords-module")}
}

// DisplayName is the modifier's name.
func (mod *HardWordsModule) DisplayName() string { return "Hard words" }

// ModifyList keeps the hard words.
func (mod *HardWordsModule) ModifyList(indexes []int, list lesson.WordList) []int {
	return ModifyList(indexes, list)
}

// InitHardWordsModule creates the module.
func InitHardWordsModule() core.Module { return NewHardWordsModule() }
