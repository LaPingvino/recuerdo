// Package checker decides whether an answer a user gave for a word is
// right. Port of OpenTeacher's logic/wordsString/checker module.
package checker

import (
	"context"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// Result is the outcome of checking one answer.
type Result struct {
	Right  bool
	ItemID int
}

// Correct reports whether given, an answer in the form parser.Parse
// produces, is right for a word whose answers are answers (same form:
// obligatory parts, each a list of alternatives). Words are compared
// exactly.
//
// When the user typed a plain list ("in, tijdens, bij"), every word they
// gave must belong to the answers, and every obligatory part must be
// covered by at least one of them. When they numbered the parts
// ("1. in 2. tijdens"), each given part must be part of an obligatory
// part, and there must be as many matches as obligatory parts.
func Correct(given, answers [][]string) bool {
	if len(given) == 1 {
		return singlePartCorrect(given[0], answers)
	}
	return multiplePartsCorrect(given, answers)
}

func singlePartCorrect(given []string, answers [][]string) bool {
	remaining := set(given)
	for _, part := range answers {
		covered := false
		for _, alternative := range part {
			if remaining[alternative] {
				delete(remaining, alternative)
				covered = true
			}
		}
		if !covered {
			return false
		}
	}
	return len(remaining) == 0
}

func multiplePartsCorrect(given, answers [][]string) bool {
	matches := 0
	for _, givenPart := range given {
		for _, part := range answers {
			if subset(givenPart, set(part)) {
				matches++
			}
		}
	}
	return matches == len(answers)
}

func set(words []string) map[string]bool {
	s := make(map[string]bool, len(words))
	for _, w := range words {
		s[w] = true
	}
	return s
}

func subset(words []string, of map[string]bool) bool {
	for _, w := range words {
		if !of[w] {
			return false
		}
	}
	return true
}

// WordsStringCheckerModule offers Check as an OpenTeacher
// "wordsStringChecker" module.
type WordsStringCheckerModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewWordsStringCheckerModule creates the module.
func NewWordsStringCheckerModule() *WordsStringCheckerModule {
	base := core.NewBaseModule("wordsStringChecker", "words-string-checker")
	base.SetPriority(10)
	return &WordsStringCheckerModule{BaseModule: base}
}

// Check checks the given answer for the word with the given id and answers.
func (mod *WordsStringCheckerModule) Check(given [][]string, itemID int, answers [][]string) Result {
	return Result{Right: Correct(given, answers), ItemID: itemID}
}

// Enable activates the module.
func (mod *WordsStringCheckerModule) Enable(ctx context.Context) error {
	return mod.BaseModule.Enable(ctx)
}

// Disable deactivates the module.
func (mod *WordsStringCheckerModule) Disable(ctx context.Context) error {
	return mod.BaseModule.Disable(ctx)
}

// SetManager sets the module manager reference.
func (mod *WordsStringCheckerModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitWordsStringCheckerModule creates and returns the module.
func InitWordsStringCheckerModule() core.Module {
	return NewWordsStringCheckerModule()
}
