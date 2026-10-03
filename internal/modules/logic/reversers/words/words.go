// Package words reverses a word list: questions become answers and the
// other way round, so a list can be practised in both directions. Port of
// OpenTeacher's logic/reversers/words.
package words

import (
	"context"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// Reverse swaps the questions and answers of every item, and the question
// and answer languages, in place.
func Reverse(list *lesson.WordList) {
	for i := range list.Items {
		item := &list.Items[i]
		item.Questions, item.Answers = item.Answers, item.Questions
	}
	list.QuestionLanguage, list.AnswerLanguage = list.AnswerLanguage, list.QuestionLanguage
}

// WordsReverserModule offers Reverse as an OpenTeacher "reverser" module.
type WordsReverserModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewWordsReverserModule creates the module.
func NewWordsReverserModule() *WordsReverserModule {
	return &WordsReverserModule{BaseModule: core.NewBaseModule("reverser", "words-reverser")}
}

// DataType is the kind of list this reverser handles.
func (mod *WordsReverserModule) DataType() string { return "words" }

// Reverse reverses the list in place.
func (mod *WordsReverserModule) Reverse(list *lesson.WordList) { Reverse(list) }

func (mod *WordsReverserModule) Enable(ctx context.Context) error { return mod.BaseModule.Enable(ctx) }
func (mod *WordsReverserModule) Disable(ctx context.Context) error {
	return mod.BaseModule.Disable(ctx)
}

// SetManager sets the module manager reference.
func (mod *WordsReverserModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitWordsReverserModule creates and returns the module.
func InitWordsReverserModule() core.Module { return NewWordsReverserModule() }
