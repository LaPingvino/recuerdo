// Package langcode maps language names to ISO 639-1 codes and back, as
// OpenTeacher's languageCodeGuesser does (lessons store language names
// like "Dutch"; some formats and text to speech want codes like "nl").
// The names come from Unicode CLDR via golang.org/x/text.
package langcode

import (
	"strings"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

var (
	once   sync.Once
	byName map[string]string // lower-case English or native name -> code
	byCode map[string]string // code -> English name
)

func load() {
	byName, byCode = map[string]string{}, map[string]string{}
	english := display.English.Languages()
	for a := 'a'; a <= 'z'; a++ {
		for b := 'a'; b <= 'z'; b++ {
			code := string([]rune{a, b})
			base, err := language.ParseBase(code)
			if err != nil || base.String() != code {
				continue
			}
			name := english.Name(base)
			if name == "" || strings.EqualFold(name, code) {
				continue
			}
			byCode[code] = name
			byName[strings.ToLower(name)] = code
			if self := display.Self.Name(language.Make(code)); self != "" {
				if _, taken := byName[strings.ToLower(self)]; !taken {
					byName[strings.ToLower(self)] = code
				}
			}
		}
	}
}

// Guess returns the ISO 639-1 code of a language name in English or in the
// language itself ("Dutch", "Nederlands" -> "nl"), or "" if unknown.
func Guess(name string) string {
	once.Do(load)
	return byName[strings.ToLower(strings.TrimSpace(name))]
}

// Name returns the English name of a language code ("nl" -> "Dutch"), or
// "" if unknown.
func Name(code string) string {
	once.Do(load)
	return byCode[strings.ToLower(strings.TrimSpace(code))]
}
