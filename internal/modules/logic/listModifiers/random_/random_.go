// Package random asks a word list in random order. Port of OpenTeacher's
// logic/listModifiers/random_.
package random

import (
	"context"
	"math/rand"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// ModifyList returns the item indexes in random order, using shuffle
// (rand.Shuffle when nil). The input is not changed.
func ModifyList(indexes []int, shuffle func(n int, swap func(i, j int))) []int {
	out := append([]int(nil), indexes...)
	if shuffle == nil {
		shuffle = rand.Shuffle
	}
	shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// RandomModule offers ModifyList as an OpenTeacher "listModifier" module.
type RandomModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewRandomModule creates the module.
func NewRandomModule() *RandomModule {
	base := core.NewBaseModule("listModifier", "random")
	base.SetPriority(811)
	return &RandomModule{BaseModule: base}
}

// DisplayName is the modifier's name as shown to users (Name is the module identifier).
func (mod *RandomModule) DisplayName() string { return "Random" }

// ModifyList returns the indexes shuffled.
func (mod *RandomModule) ModifyList(indexes []int) []int { return ModifyList(indexes, nil) }

func (mod *RandomModule) Enable(ctx context.Context) error  { return mod.BaseModule.Enable(ctx) }
func (mod *RandomModule) Disable(ctx context.Context) error { return mod.BaseModule.Disable(ctx) }

// SetManager sets the module manager reference.
func (mod *RandomModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitRandomModule creates and returns the module.
func InitRandomModule() core.Module { return NewRandomModule() }
