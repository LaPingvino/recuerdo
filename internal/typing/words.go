package typing

import (
	"embed"
	"strings"
	"sync"
	"unicode"
)

// The word levels' words: common words per language, written for
// Recuerdo (OpenTeacher's list came from its own source code).
//
//go:embed words/*.txt
var wordFiles embed.FS

var (
	wordMu    sync.Mutex
	wordCache = map[string][]string{}
)

// WordLanguages are the languages with a word list.
func WordLanguages() []string {
	entries, _ := wordFiles.ReadDir("words")
	var langs []string
	for _, e := range entries {
		langs = append(langs, strings.TrimSuffix(e.Name(), ".txt"))
	}
	return langs
}

// Words is the word list of a language (a code such as "nl" or "pt_BR";
// English when there is none): lower-case words of letters only.
func Words(lang string) []string {
	wordMu.Lock()
	defer wordMu.Unlock()
	code := strings.ToLower(strings.SplitN(strings.SplitN(lang, "_", 2)[0], "-", 2)[0])
	if w, ok := wordCache[code]; ok {
		return w
	}
	data, err := wordFiles.ReadFile("words/" + code + ".txt")
	if err != nil {
		data, _ = wordFiles.ReadFile("words/en.txt")
	}
	seen := map[string]bool{}
	var words []string
	for _, w := range strings.Fields(string(data)) {
		w = strings.ToLower(w)
		if !seen[w] && strings.IndexFunc(w, func(r rune) bool { return !unicode.IsLetter(r) }) < 0 {
			seen[w] = true
			words = append(words, w)
		}
	}
	wordCache[code] = words
	return words
}
