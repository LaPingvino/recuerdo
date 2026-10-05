// Package spellcheck checks the spelling of words being entered, in the
// lesson's question or answer language, with Hunspell dictionaries (as
// installed for LibreOffice and Firefox), through the hunspell program.
// OpenTeacher used Enchant; without hunspell or a dictionary for the
// language nothing is checked.
package spellcheck

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"unicode"

	"github.com/LaPingvino/recuerdo/internal/langcode"
	"github.com/LaPingvino/recuerdo/internal/richtext"
)

// Dirs are where Hunspell dictionaries (NAME.dic with NAME.aff) are
// looked for, besides DICPATH.
func Dirs() []string {
	home, _ := os.UserHomeDir()
	dirs := filepath.SplitList(os.Getenv("DICPATH"))
	dirs = append(dirs, filepath.Join(home, ".local", "share", "hunspell"),
		"/usr/share/hunspell", "/usr/share/myspell", "/usr/share/myspell/dicts", "/usr/local/share/hunspell",
		"/opt/homebrew/share/hunspell", filepath.Join(home, "Library", "Spelling"), "/Library/Spelling")
	if runtime.GOOS == "windows" {
		dirs = append(dirs, filepath.Join(os.Getenv("LOCALAPPDATA"), "hunspell"))
	}
	return dirs
}

// Dictionaries are the names of the installed dictionaries ("en_US",
// "nl_NL"), mapped to their directory.
func Dictionaries() map[string]string {
	found := map[string]string{}
	for _, dir := range Dirs() {
		files, _ := filepath.Glob(filepath.Join(dir, "*.dic"))
		for _, f := range files {
			name := strings.TrimSuffix(filepath.Base(f), ".dic")
			if _, err := os.Stat(filepath.Join(dir, name+".aff")); err == nil && found[name] == "" {
				found[name] = dir
			}
		}
	}
	return found
}

// DictionaryFor is the dictionary for a language: a name ("Dutch",
// "Nederlands"), a code ("nl") or a dictionary name ("en_GB"); for a
// language with several, the main one (nl_NL, de_DE, en_US) first.
func DictionaryFor(language string, dicts map[string]string) string {
	language = strings.TrimSpace(language)
	if language == "" {
		return ""
	}
	if _, ok := dicts[language]; ok {
		return language
	}
	code := langcode.Guess(language)
	if code == "" && len(language) <= 3 {
		code = strings.ToLower(language)
	}
	if code == "" {
		return ""
	}
	main := map[string]string{"en": "en_US", "pt": "pt_PT", "zh": "zh_CN", "sv": "sv_SE", "da": "da_DK",
		"el": "el_GR", "cs": "cs_CZ", "uk": "uk_UA", "ca": "ca_ES", "sr": "sr_RS"}[code]
	if main == "" {
		main = code + "_" + strings.ToUpper(code)
	}
	for _, name := range []string{main, code} {
		if _, ok := dicts[name]; ok {
			return name
		}
	}
	var names []string
	for name := range dicts {
		if strings.HasPrefix(name, code+"_") || strings.HasPrefix(name, code+"-") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) > 0 {
		return names[0]
	}
	return ""
}

// Words are the words of an entered text worth checking: markup and
// formulas left out, split at anything that is not a letter or digit (an
// apostrophe or hyphen inside a word belongs to it); words with a digit
// are left out.
func Words(text string) []string {
	text = richtext.Plain(withoutFormulas(text))
	var words []string
	var cur []rune
	flush := func() {
		w := strings.Trim(string(cur), "'’-")
		if w != "" && !hasDigit(w) {
			words = append(words, w)
		}
		cur = cur[:0]
	}
	for _, r := range text {
		// digits stay in the word, so that H2O is left out as a whole
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r) ||
			((r == '\'' || r == '’' || r == '-') && len(cur) > 0) {
			cur = append(cur, r)
			continue
		}
		flush()
	}
	flush()
	return words
}

func withoutFormulas(s string) string {
	if !richtext.HasMath(s) {
		return s
	}
	var b strings.Builder
	richtext.EachSpan(s, func(string) { b.WriteString(" ") }, func(t string) { b.WriteString(t) })
	return b.String()
}

func hasDigit(s string) bool {
	return strings.IndexFunc(s, unicode.IsDigit) >= 0
}

// Checker checks words against one dictionary.
type Checker struct {
	Dictionary string
	dir        string
	program    string
	cache      map[string][]string // word: nil if right, else suggestions (maybe empty)
}

// New is a checker for a language, or nil when hunspell or a dictionary
// for it is not installed.
func New(language string) *Checker {
	program, err := exec.LookPath("hunspell")
	if err != nil {
		return nil
	}
	dicts := Dictionaries()
	name := DictionaryFor(language, dicts)
	if name == "" {
		return nil
	}
	return &Checker{Dictionary: name, dir: dicts[name], program: program, cache: map[string][]string{}}
}

// Wrong are the words of text that are not in the dictionary, each with
// Hunspell's suggestions.
func (c *Checker) Wrong(texts ...string) map[string][]string {
	wrong := map[string][]string{}
	if c == nil {
		return wrong
	}
	var ask []string
	seen := map[string]bool{}
	for _, t := range texts {
		for _, w := range Words(t) {
			if s, ok := c.cache[w]; ok {
				if s != nil {
					wrong[w] = s
				}
			} else if !seen[w] {
				seen[w] = true
				ask = append(ask, w)
			}
		}
	}
	if len(ask) > 0 {
		for w, s := range c.run(ask) {
			c.cache[w] = s
		}
		for _, w := range ask {
			if _, ok := c.cache[w]; !ok {
				c.cache[w] = nil // right
			}
			if s := c.cache[w]; s != nil {
				wrong[w] = s
			}
		}
	}
	return wrong
}

// run asks hunspell in its ispell mode (-a): one word per line; a line
// starting with & (with suggestions) or # (none) is a misspelled word.
func (c *Checker) run(words []string) map[string][]string {
	var in bytes.Buffer
	for _, w := range words {
		in.WriteString("^" + w + "\n") // ^: text, not a command
	}
	cmd := exec.Command(c.program, "-a", "-i", "utf-8", "-d", filepath.Join(c.dir, c.Dictionary))
	cmd.Stdin = &in
	out, err := cmd.Output()
	wrong := map[string][]string{}
	if err != nil {
		return wrong
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "& "):
			head, list, _ := strings.Cut(line, ": ")
			fields := strings.Fields(head)
			if len(fields) >= 2 {
				wrong[fields[1]] = strings.Split(list, ", ")
			}
		case strings.HasPrefix(line, "# "):
			if fields := strings.Fields(line); len(fields) >= 2 {
				wrong[fields[1]] = []string{}
			}
		}
	}
	return wrong
}
