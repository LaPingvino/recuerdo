// Package reverse asks a word list from last to first. Port of
// OpenTeacher's logic/listModifiers/reverse.
package reverse

import (
	"context"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// ModifyList returns the item indexes in reverse order.
func ModifyList(indexes []int) []int {
	out := make([]int, len(indexes))
	for i, index := range indexes {
		out[len(indexes)-1-i] = index
	}
	return out
}

// ReverseModule offers ModifyList as an OpenTeacher "listModifier" module.
type ReverseModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewReverseModule creates the module.
func NewReverseModule() *ReverseModule {
	base := core.NewBaseModule("listModifier", "reverse")
	base.SetPriority(848)
	return &ReverseModule{BaseModule: base}
}

// DisplayName is the modifier's name as shown to users (Name is the module identifier).
func (mod *ReverseModule) DisplayName() string { return "Reverse" }

// ModifyList returns the indexes reversed.
func (mod *ReverseModule) ModifyList(indexes []int) []int { return ModifyList(indexes) }

func (mod *ReverseModule) Enable(ctx context.Context) error  { return mod.BaseModule.Enable(ctx) }
func (mod *ReverseModule) Disable(ctx context.Context) error { return mod.BaseModule.Disable(ctx) }

// SetManager sets the module manager reference.
func (mod *ReverseModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitReverseModule creates and returns the module.
func InitReverseModule() core.Module { return NewReverseModule() }
