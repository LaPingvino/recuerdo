// Package parser parses the text a user types for a word's questions or
// answers into OpenTeacher's internal form. Port of OpenTeacher's
// logic/wordsString/parser module.
package parser

import (
	"context"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// Parse splits text into obligatory parts, each a list of alternatives:
//
//	"one, two"          -> [[one two]]
//	"1. one, uno 2. two" -> [[one uno] [two]]
//
// Parts are numbered "1. ", "2. ", ...; alternatives are separated by ","
// or ";". A backslash before a number, comma or semicolon keeps it as
// text. Text before the first number means it is not a numbered list
// ("Am 20. Mai habe ich ..." stays one answer).
func Parse(text string) [][]string {
	segments := splitSegments(text)
	for i := range segments {
		segments[i] = strings.TrimSpace(segments[i])
	}
	if segments[0] != "" {
		// https://bugs.launchpad.net/openteacher/+bug/1233809
		segments = []string{text}
	}

	item := [][]string{}
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		var words []string
		for _, word := range splitAlternatives(segment) {
			if word = strings.TrimSpace(word); word != "" {
				words = append(words, word)
			}
		}
		if len(words) > 0 {
			item = append(item, words)
		}
	}
	return item
}

// splitSegments splits text at "<digits>. " not preceded by a backslash,
// as OpenTeacher's (?<!\\)[0-9]+\. regex does.
func splitSegments(text string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(text); i++ {
		if !isDigit(text[i]) || (i > 0 && text[i-1] == '\\') {
			continue
		}
		j := i
		for j < len(text) && isDigit(text[j]) {
			j++
		}
		if j+1 < len(text) && text[j] == '.' && text[j+1] == ' ' {
			parts = append(parts, text[start:i])
			start = j + 2
			i = j + 1
		}
	}
	return append(parts, text[start:])
}

// splitAlternatives splits at "," and ";" not preceded by a backslash.
func splitAlternatives(segment string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(segment); i++ {
		if (segment[i] == ',' || segment[i] == ';') && (i == 0 || segment[i-1] != '\\') {
			parts = append(parts, segment[start:i])
			start = i + 1
		}
	}
	return append(parts, segment[start:])
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// WordsStringParserModule offers Parse as an OpenTeacher
// "wordsStringParser" module.
type WordsStringParserModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewWordsStringParserModule creates the module.
func NewWordsStringParserModule() *WordsStringParserModule {
	base := core.NewBaseModule("wordsStringParser", "words-string-parser")
	base.SetPriority(10)
	return &WordsStringParserModule{BaseModule: base}
}

// Parse parses questions or answers text; see the package function.
func (mod *WordsStringParserModule) Parse(text string) [][]string {
	return Parse(text)
}

// Enable activates the module.
func (mod *WordsStringParserModule) Enable(ctx context.Context) error {
	return mod.BaseModule.Enable(ctx)
}

// Disable deactivates the module.
func (mod *WordsStringParserModule) Disable(ctx context.Context) error {
	return mod.BaseModule.Disable(ctx)
}

// SetManager sets the module manager reference.
func (mod *WordsStringParserModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitWordsStringParserModule creates and returns the module.
func InitWordsStringParserModule() core.Module {
	return NewWordsStringParserModule()
}
