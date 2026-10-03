// Package checker decides whether an answer a user gave for a word is
// right. Port of OpenTeacher's logic/wordsString/checker module.
package checker

import (
	"context"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/parser"
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

// StoredAnswers turns a word's answers as Recuerdo's lesson model keeps
// them (a flat list) back into obligatory parts: a single stored answer
// that is itself a numbered list ("1. one 2. two") gives its parts;
// otherwise the stored answers are the alternatives of one part.
func StoredAnswers(stored []string) [][]string {
	if len(stored) == 1 {
		if parts := parser.Parse(stored[0]); len(parts) > 1 {
			return parts
		}
	}
	var alternatives []string
	for _, a := range stored {
		if a = strings.TrimSpace(a); a != "" {
			alternatives = append(alternatives, a)
		}
	}
	if len(alternatives) == 0 {
		return [][]string{}
	}
	return [][]string{alternatives}
}

// CorrectText checks an answer as typed against a word's stored answers.
// Unless caseSensitive is set, capitals are ignored (OpenTeacher itself
// compares exactly).
func CorrectText(typed string, stored []string, caseSensitive bool) bool {
	given, answers := parser.Parse(typed), StoredAnswers(stored)
	if !caseSensitive {
		given, answers = lower(given), lower(answers)
	}
	return Correct(given, answers)
}

func lower(item [][]string) [][]string {
	out := make([][]string, len(item))
	for i, part := range item {
		out[i] = make([]string, len(part))
		for j, w := range part {
			out[i][j] = strings.ToLower(w)
		}
	}
	return out
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
