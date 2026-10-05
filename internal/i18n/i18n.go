// Package i18n translates Recuerdo's interface with OpenTeacher's
// translations (data/translations/<lang>.po, made by
// scripts/merge_translations.py). Texts without a translation stay
// English.
package i18n

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// LanguageSetting is the setting with the chosen language ("" follows the
// system).
const LanguageSetting = "org.openteacher.translator.language"

var (
	mu      sync.RWMutex
	catalog map[string]string
	current string
)

// T is msgid in the current language (msgid itself without a
// translation). A text OpenTeacher has without the trailing "..." (a menu
// item that opens a dialog) is translated and gets it back; Recuerdo's
// own wording is looked up through aliases.
func T(msgid string) string {
	mu.RLock()
	defer mu.RUnlock()
	if catalog == nil {
		return msgid
	}
	if s, ok := lookup(msgid); ok {
		return s
	}
	for _, dots := range []string{"...", "…"} {
		if base, ok := strings.CutSuffix(msgid, dots); ok {
			if s, ok := lookup(base); ok {
				return s + dots
			}
		}
	}
	return msgid
}

func lookup(msgid string) (string, bool) {
	if s, ok := catalog[msgid]; ok {
		return s, true
	}
	if a, ok := aliases[msgid]; ok {
		s, ok := catalog[a]
		// OpenTeacher's label "Question:" serves Recuerdo's "Question"
		if strings.HasSuffix(a, ":") && !strings.HasSuffix(msgid, ":") {
			// (also the full-width "：" of Chinese and Japanese)
			s = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(s), ":："))
		}
		return s, ok
	}
	return "", false
}

// Tf is T with fmt.Sprintf arguments (OpenTeacher's "{0}" style is not used).
func Tf(msgid string, args ...any) string { return fmt.Sprintf(T(msgid), args...) }

// Current is the language in use ("" for English).
func Current() string {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// Use loads the translations for lang from dir ("" or English: none).
// A language with a region falls back to the language alone (nl_BE -> nl)
// and the other way round (pt -> pt_BR).
func Use(dir, lang string) error {
	file := Resolve(dir, lang)
	mu.Lock()
	defer mu.Unlock()
	catalog, current = nil, ""
	if file == "" {
		return nil
	}
	// OpenTeacher's translations, then Recuerdo's own over them
	c := map[string]string{}
	for _, name := range []string{file + ".po", "recuerdo-" + file + ".po"} {
		part, err := ReadPO(filepath.Join(dir, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		for k, v := range part {
			c[k] = v
		}
	}
	catalog, current = c, file
	return nil
}

// Resolve is the translation file (without .po) for lang in dir, or "".
func Resolve(dir, lang string) string {
	lang = normalize(lang)
	if lang == "" || lang == "en" || lang == "en_US" || lang == "C" || lang == "POSIX" {
		return ""
	}
	available := Available(dir)
	has := func(l string) bool {
		i := sort.SearchStrings(available, l)
		return i < len(available) && available[i] == l
	}
	if has(lang) {
		return lang
	}
	base := strings.SplitN(lang, "_", 2)[0]
	if has(base) {
		return base
	}
	for _, l := range available {
		if strings.HasPrefix(l, base+"_") {
			return l
		}
	}
	return ""
}

// normalize turns locale names (nl_NL.UTF-8, pt-BR, de@euro) into nl_NL,
// pt_BR, de.
func normalize(lang string) string {
	lang = strings.TrimSpace(lang)
	if i := strings.IndexAny(lang, ".@"); i >= 0 {
		lang = lang[:i]
	}
	lang = strings.ReplaceAll(lang, "-", "_")
	if i := strings.Index(lang, "_"); i >= 0 {
		return strings.ToLower(lang[:i]) + "_" + strings.ToUpper(lang[i+1:])
	}
	return strings.ToLower(lang)
}

// SystemLanguage is the language of the user's locale (LC_ALL,
// LC_MESSAGES, LANG; LANGUAGE's first entry first).
func SystemLanguage() string {
	if l := os.Getenv("LANGUAGE"); l != "" {
		return normalize(strings.SplitN(l, ":", 2)[0])
	}
	for _, v := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if l := os.Getenv(v); l != "" {
			return normalize(l)
		}
	}
	return ""
}

// Available are the languages with translations in dir (OpenTeacher's
// <lang>.po or Recuerdo's recuerdo-<lang>.po), sorted.
func Available(dir string) []string {
	files, _ := filepath.Glob(filepath.Join(dir, "*.po"))
	seen := map[string]bool{}
	var out []string
	for _, f := range files {
		l := strings.TrimPrefix(strings.TrimSuffix(filepath.Base(f), ".po"), "recuerdo-")
		if !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}
	sort.Strings(out)
	return out
}

// ReadPO reads the translations in a .po file (fuzzy ones left out).
func ReadPO(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	var msgid, msgstr *string
	field, fuzzy := "", false
	flush := func() {
		if msgid != nil && msgstr != nil && *msgid != "" && *msgstr != "" && !fuzzy {
			out[*msgid] = *msgstr
		}
		msgid, msgstr, field, fuzzy = nil, nil, "", false
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		var err error
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "#,"):
			fuzzy = fuzzy || strings.Contains(line, "fuzzy")
		case strings.HasPrefix(line, "#"):
		case strings.HasPrefix(line, "msgid "):
			if msgid != nil {
				flush()
			}
			var s string
			s, err = strconv.Unquote(line[6:])
			msgid, field = &s, "id"
		case strings.HasPrefix(line, "msgstr "):
			var s string
			s, err = strconv.Unquote(line[7:])
			msgstr, field = &s, "str"
		case strings.HasPrefix(line, `"`):
			var s string
			if s, err = strconv.Unquote(line); err == nil {
				switch field {
				case "id":
					*msgid += s
				case "str":
					*msgstr += s
				}
			}
		}
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %v", path, n, err)
		}
	}
	flush()
	return out, sc.Err()
}

// ReadPOT is the set of texts in a .pot template.
func ReadPOT(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]bool{}
	var cur *string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case strings.HasPrefix(line, "msgid "):
			s, err := strconv.Unquote(line[6:])
			if err != nil {
				return nil, err
			}
			cur = &s
		case strings.HasPrefix(line, `"`) && cur != nil:
			s, err := strconv.Unquote(line)
			if err != nil {
				return nil, err
			}
			*cur += s
		case strings.HasPrefix(line, "msgstr"):
			if cur != nil && *cur != "" {
				out[*cur] = true
			}
			cur = nil
		}
	}
	return out, sc.Err()
}
