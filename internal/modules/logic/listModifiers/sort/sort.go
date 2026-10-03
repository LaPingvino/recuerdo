// Package sort asks a word list in alphabetical order of the questions.
// Port of OpenTeacher's logic/listModifiers/sort.
package sort

import (
	"context"
	"slices"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// ModifyList returns the item indexes ordered by each item's first
// question (items without questions first; equal ones keep their order).
func ModifyList(indexes []int, items []lesson.WordItem) []int {
	out := append([]int(nil), indexes...)
	first := func(i int) (string, bool) {
		if q := items[i].Questions; len(q) > 0 {
			return q[0], true
		}
		return "", false
	}
	slices.SortStableFunc(out, func(a, b int) int {
		qa, oka := first(a)
		qb, okb := first(b)
		switch {
		case !oka && !okb:
			return 0
		case !oka:
			return -1
		case !okb:
			return 1
		}
		switch {
		case qa < qb:
			return -1
		case qa > qb:
			return 1
		}
		return 0
	})
	return out
}

// SortModule offers ModifyList as an OpenTeacher "listModifier" module.
type SortModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewSortModule creates the module.
func NewSortModule() *SortModule {
	base := core.NewBaseModule("listModifier", "sort")
	base.SetPriority(911)
	return &SortModule{BaseModule: base}
}

// DisplayName is the modifier's name as shown to users (Name is the module identifier).
func (mod *SortModule) DisplayName() string { return "Sort" }

// ModifyList returns the indexes in alphabetical order of the questions.
func (mod *SortModule) ModifyList(indexes []int, items []lesson.WordItem) []int {
	return ModifyList(indexes, items)
}

func (mod *SortModule) Enable(ctx context.Context) error  { return mod.BaseModule.Enable(ctx) }
func (mod *SortModule) Disable(ctx context.Context) error { return mod.BaseModule.Disable(ctx) }

// SetManager sets the module manager reference.
func (mod *SortModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitSortModule creates and returns the module.
func InitSortModule() core.Module { return NewSortModule() }
