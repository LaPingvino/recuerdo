// Package composer turns OpenTeacher's internal form of a word's questions
// or answers back into the text a user would type. Port of OpenTeacher's
// logic/wordsString/composer module.
package composer

import (
	"context"
	"strconv"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// Compose is the inverse of parser.Parse:
//
//	[[one two]]          -> "one, two"
//	[[one uno] [two]]    -> "1. one, uno 2. two"
func Compose(item [][]string) string {
	switch len(item) {
	case 0:
		return ""
	case 1:
		return strings.Join(item[0], ", ")
	}
	parts := make([]string, len(item))
	for i, alternatives := range item {
		parts[i] = strconv.Itoa(i+1) + ". " + strings.Join(alternatives, ", ")
	}
	return strings.Join(parts, " ")
}

// WordsStringComposerModule offers Compose as an OpenTeacher
// "wordsStringComposer" module.
type WordsStringComposerModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewWordsStringComposerModule creates the module.
func NewWordsStringComposerModule() *WordsStringComposerModule {
	base := core.NewBaseModule("wordsStringComposer", "words-string-composer")
	base.SetPriority(10)
	return &WordsStringComposerModule{BaseModule: base}
}

// Compose composes questions or answers text; see the package function.
func (mod *WordsStringComposerModule) Compose(item [][]string) string {
	return Compose(item)
}

// Enable activates the module.
func (mod *WordsStringComposerModule) Enable(ctx context.Context) error {
	return mod.BaseModule.Enable(ctx)
}

// Disable deactivates the module.
func (mod *WordsStringComposerModule) Disable(ctx context.Context) error {
	return mod.BaseModule.Disable(ctx)
}

// SetManager sets the module manager reference.
func (mod *WordsStringComposerModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitWordsStringComposerModule creates and returns the module.
func InitWordsStringComposerModule() core.Module {
	return NewWordsStringComposerModule()
}
