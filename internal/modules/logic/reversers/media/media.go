// Package media reverses a media list: each item's question becomes its
// answer and the other way round. Port of OpenTeacher's
// logic/reversers/media, which swaps "question" and "answer"; Recuerdo
// keeps those as an item's Questions and Answers, as for words.
package media

import (
	"context"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// Reverse swaps the question and answer of every item in place; the
// media files, names and results stay as they are.
func Reverse(list *lesson.WordList) {
	for i := range list.Items {
		item := &list.Items[i]
		item.Questions, item.Answers = item.Answers, item.Questions
	}
}

// MediaReverserModule offers Reverse as an OpenTeacher "reverser" module.
type MediaReverserModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewMediaReverserModule creates the module.
func NewMediaReverserModule() *MediaReverserModule {
	return &MediaReverserModule{BaseModule: core.NewBaseModule("reverser", "media-reverser")}
}

// DataType is the kind of list this reverser handles.
func (mod *MediaReverserModule) DataType() string { return "media" }

// Reverse reverses the list in place.
func (mod *MediaReverserModule) Reverse(list *lesson.WordList) { Reverse(list) }

func (mod *MediaReverserModule) Enable(ctx context.Context) error { return mod.BaseModule.Enable(ctx) }
func (mod *MediaReverserModule) Disable(ctx context.Context) error {
	return mod.BaseModule.Disable(ctx)
}

// SetManager sets the module manager reference.
func (mod *MediaReverserModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitMediaReverserModule creates and returns the module.
func InitMediaReverserModule() core.Module { return NewMediaReverserModule() }
