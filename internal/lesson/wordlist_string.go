package lesson

import (
	"fmt"
	"regexp"
	"strings"
)

// The word list string format: one word per line, "question = answer" or
// question<tab>answer, with "\=" and "\t" for an equals sign or tab in a
// word. OpenTeacher uses it for pasting lists and for several file
// formats (logic/wordListString).

// ParseWordList reads a word list string. Lines without a separator are
// an error, unless lenient is set, which skips them.
func ParseWordList(text string, lenient bool) ([]WordItem, error) {
	var items []WordItem
	for n, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		q, a, ok := splitUnescaped(line)
		if !ok {
			if lenient {
				continue
			}
			return nil, fmt.Errorf("line %d: missing equals sign or tab", n+1)
		}
		items = append(items, WordItem{
			ID:        len(items),
			Questions: splitWords(unescapeWordList(q)),
			Answers:   splitWords(unescapeWordList(a)),
		})
	}
	return items, nil
}

// ComposeWordList writes items as a word list string.
func ComposeWordList(items []WordItem) string {
	var b strings.Builder
	for _, it := range items {
		fmt.Fprintf(&b, "%s = %s\n", escapeWordList(composeWords(it.Questions)), escapeWordList(composeWords(it.Answers)))
	}
	return b.String()
}

// splitUnescaped splits a line at the first "=" or tab not preceded by a
// backslash.
func splitUnescaped(line string) (string, string, bool) {
	inside := protected(line)
	for i := 0; i < len(line); i++ {
		if (line[i] == '=' || line[i] == '\t') && (i == 0 || line[i-1] != '\\') && !inside[i] {
			return line[:i], line[i+1:], true
		}
	}
	return "", "", false
}

// protected marks the bytes of s inside a formula ($...$, $$...$$,
// \(...\), \[...\]) or an HTML tag (<...>): an equals sign, comma or
// semicolon there belongs to the formula or tag (f(x, y), a data: URL),
// not to the word list notation.
func protected(s string) []bool {
	in := make([]bool, len(s))
	mark := func(from, to int) {
		for k := from; k < to && k < len(s); k++ {
			in[k] = true
		}
	}
	for i := 0; i < len(s); {
		switch {
		case strings.HasPrefix(s[i:], `\$`):
			i += 2
			continue
		case s[i] == '<' && i+1 < len(s) && (s[i+1] == '/' || s[i+1] >= 'a' && s[i+1] <= 'z' || s[i+1] >= 'A' && s[i+1] <= 'Z'):
			if j := strings.IndexByte(s[i:], '>'); j > 0 {
				mark(i, i+j+1)
				i += j + 1
				continue
			}
		case s[i] == '$' || strings.HasPrefix(s[i:], `\(`) || strings.HasPrefix(s[i:], `\[`):
			open, close := "$", "$"
			switch {
			case strings.HasPrefix(s[i:], "$$"):
				open, close = "$$", "$$"
			case strings.HasPrefix(s[i:], `\(`):
				open, close = `\(`, `\)`
			case strings.HasPrefix(s[i:], `\[`):
				open, close = `\[`, `\]`
			}
			if j := strings.Index(s[i+len(open):], close); j > 0 {
				end := i + len(open) + j + len(close)
				mark(i, end)
				i = end
				continue
			}
		}
		i++
	}
	return in
}

var wordListUnescaper = strings.NewReplacer(`\=`, "=", `\t`, "\t")

func unescapeWordList(s string) string { return wordListUnescaper.Replace(strings.TrimSpace(s)) }

var wordListEscaper = regexp.MustCompile("[=\t]")

func escapeWordList(s string) string {
	return wordListEscaper.ReplaceAllStringFunc(s, func(m string) string {
		if m == "=" {
			return `\=`
		}
		return `\t`
	})
}

// splitWords splits alternatives as OpenTeacher's words string parser
// does: by commas and semicolons.
func splitWords(s string) []string {
	var out []string
	inside := protected(s)
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || (s[i] == ',' || s[i] == ';') && !inside[i] {
			if w := strings.TrimSpace(s[start:i]); w != "" {
				out = append(out, w)
			}
			start = i + 1
		}
	}
	return out
}
