package richtext

import (
	"strings"
	"unicode"
)

// Formulas are TeX between $...$ or \(...\) (inline) and $$...$$ or
// \[...\] (on their own); \$ is a dollar sign. The web version shows them
// with KaTeX; answers are checked against their TeX source without
// spaces, so x^2+1 matches $x^2 + 1$.

// HasMath reports whether s contains a formula.
func HasMath(s string) bool {
	found := false
	mathSpans(s, func(string) { found = true }, nil)
	return found
}

// mathSpans calls onMath with each formula's TeX and onText with the
// text between them (\$ as $), in order.
func mathSpans(s string, onMath, onText func(string)) {
	var text strings.Builder
	flush := func() {
		if text.Len() > 0 && onText != nil {
			onText(text.String())
		}
		text.Reset()
	}
	for i := 0; i < len(s); {
		switch {
		case strings.HasPrefix(s[i:], `\$`):
			text.WriteByte('$')
			i += 2
			continue
		case strings.HasPrefix(s[i:], "$$"), strings.HasPrefix(s[i:], `\[`), strings.HasPrefix(s[i:], `\(`), s[i] == '$':
			open, close := s[i:i+1], "$"
			switch {
			case strings.HasPrefix(s[i:], "$$"):
				open, close = "$$", "$$"
			case strings.HasPrefix(s[i:], `\[`):
				open, close = `\[`, `\]`
			case strings.HasPrefix(s[i:], `\(`):
				open, close = `\(`, `\)`
			}
			if j := strings.Index(s[i+len(open):], close); j > 0 {
				flush()
				if onMath != nil {
					onMath(s[i+len(open) : i+len(open)+j])
				}
				i += len(open) + j + len(close)
				continue
			}
		}
		text.WriteByte(s[i])
		i++
	}
	flush()
}

// mathPlain is a formula's TeX without spaces, for checking.
func mathPlain(tex string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, tex)
}

// PlainMath is s with its formulas as their TeX without spaces and
// without delimiters ("area: $\pi r^2$" -> "area: \pi r^2" without the
// space in the formula: "area: \pir^2").
func PlainMath(s string) string {
	if !HasMath(s) && !strings.Contains(s, `\$`) {
		return s
	}
	var b strings.Builder
	mathSpans(s, func(tex string) { b.WriteString(mathPlain(tex)) }, func(t string) { b.WriteString(t) })
	return b.String()
}

// NormalizeAnswer makes a typed answer comparable with a formula's
// plain form: without spaces.
func NormalizeAnswer(typed string) string { return mathPlain(typed) }
