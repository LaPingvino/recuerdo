package richtext

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

// TeX to rich text: a readable form of a formula in the HTML subset that
// Qt's labels and documents show (no TeX engine needed): sub- and
// superscripts, Greek letters and operators as Unicode, \frac{a}{b} as
// a⁄b, \sqrt{x} as √x. Commands it does not know stay as TeX.

var texSymbols = map[string]string{
	"alpha": "α", "beta": "β", "gamma": "γ", "delta": "δ", "epsilon": "ε", "varepsilon": "ε", "zeta": "ζ",
	"eta": "η", "theta": "θ", "vartheta": "ϑ", "iota": "ι", "kappa": "κ", "lambda": "λ", "mu": "μ", "nu": "ν",
	"xi": "ξ", "pi": "π", "varpi": "ϖ", "rho": "ρ", "sigma": "σ", "varsigma": "ς", "tau": "τ", "upsilon": "υ",
	"phi": "φ", "varphi": "φ", "chi": "χ", "psi": "ψ", "omega": "ω",
	"Gamma": "Γ", "Delta": "Δ", "Theta": "Θ", "Lambda": "Λ", "Xi": "Ξ", "Pi": "Π", "Sigma": "Σ", "Upsilon": "Υ",
	"Phi": "Φ", "Psi": "Ψ", "Omega": "Ω",
	"pm": "±", "mp": "∓", "times": "×", "cdot": "·", "div": "÷", "ast": "∗", "star": "⋆", "circ": "∘",
	"le": "≤", "leq": "≤", "ge": "≥", "geq": "≥", "ne": "≠", "neq": "≠", "approx": "≈", "equiv": "≡", "sim": "∼",
	"simeq": "≃", "cong": "≅", "propto": "∝", "ll": "≪", "gg": "≫",
	"infty": "∞", "partial": "∂", "nabla": "∇", "sum": "∑", "prod": "∏", "int": "∫", "oint": "∮", "iint": "∬",
	"to": "→", "rightarrow": "→", "leftarrow": "←", "leftrightarrow": "↔", "Rightarrow": "⇒", "Leftarrow": "⇐",
	"Leftrightarrow": "⇔", "implies": "⇒", "iff": "⇔", "mapsto": "↦", "uparrow": "↑", "downarrow": "↓",
	"in": "∈", "notin": "∉", "ni": "∋", "subset": "⊂", "supset": "⊃", "subseteq": "⊆", "supseteq": "⊇",
	"cup": "∪", "cap": "∩", "emptyset": "∅", "varnothing": "∅", "setminus": "∖",
	"forall": "∀", "exists": "∃", "neg": "¬", "lnot": "¬", "wedge": "∧", "land": "∧", "vee": "∨", "lor": "∨",
	"oplus": "⊕", "otimes": "⊗", "perp": "⊥", "parallel": "∥", "angle": "∠", "triangle": "△",
	"degree": "°", "prime": "′", "ldots": "…", "cdots": "⋯", "dots": "…", "hbar": "ℏ", "ell": "ℓ",
	"Re": "ℜ", "Im": "ℑ", "aleph": "ℵ", "therefore": "∴", "because": "∵",
	"sin": "sin", "cos": "cos", "tan": "tan", "log": "log", "ln": "ln", "exp": "exp", "lim": "lim",
	"min": "min", "max": "max", "det": "det", "arcsin": "arcsin", "arccos": "arccos", "arctan": "arctan",
	"sinh": "sinh", "cosh": "cosh", "tanh": "tanh", "gcd": "gcd", "mod": "mod",
	"{": "{", "}": "}", "$": "$", "%": "%", "&": "&", "#": "#", "_": "_",
	",": " ", ";": " ", ":": " ", "!": "", "quad": " ", "qquad": "  ", " ": " ",
	"left": "", "right": "", "big": "", "Big": "", "displaystyle": "", "limits": "",
}

type texParser struct {
	s   string
	pos int
}

// TeXToRich turns TeX (without $ delimiters) into the rich text described
// above.
func TeXToRich(tex string) string {
	p := &texParser{s: tex}
	return p.parse(false)
}

// parse reads until the end, or until a closing } when inGroup.
func (p *texParser) parse(inGroup bool) string {
	var b strings.Builder
	for p.pos < len(p.s) {
		c := p.s[p.pos]
		switch {
		case c == '}' && inGroup:
			p.pos++
			return b.String()
		case c == '{':
			p.pos++
			b.WriteString(p.parse(true))
		case c == '^' || c == '_':
			p.pos++
			tag := "sup"
			if c == '_' {
				tag = "sub"
			}
			b.WriteString("<" + tag + ">" + p.argument() + "</" + tag + ">")
		case c == '\\':
			b.WriteString(p.command())
		case c == '~':
			p.pos++
			b.WriteString(" ")
		case c == ' ' || c == '\t' || c == '\n':
			p.pos++ // TeX ignores spaces in formulas; operators get their own
		case c == '-':
			p.pos++
			b.WriteString("−")
		case strings.IndexByte("+=<>", c) >= 0:
			p.pos++
			b.WriteString(" " + html.EscapeString(string(c)) + " ")
		default:
			r, size := decodeRune(p.s[p.pos:])
			p.pos += size
			b.WriteString(html.EscapeString(string(r)))
		}
	}
	return b.String()
}

func decodeRune(s string) (rune, int) {
	for i, r := range s {
		_ = i
		return r, len(string(r))
	}
	return 0, 1
}

// argument reads one argument: a {group}, a command or one character.
func (p *texParser) argument() string {
	for p.pos < len(p.s) && p.s[p.pos] == ' ' {
		p.pos++
	}
	if p.pos >= len(p.s) {
		return ""
	}
	switch p.s[p.pos] {
	case '{':
		p.pos++
		return p.parse(true)
	case '\\':
		return p.command()
	}
	r, size := decodeRune(p.s[p.pos:])
	p.pos += size
	return html.EscapeString(string(r))
}

// rawArgument reads a {group} as text (for \text and friends).
func (p *texParser) rawArgument() string {
	for p.pos < len(p.s) && p.s[p.pos] == ' ' {
		p.pos++
	}
	if p.pos >= len(p.s) || p.s[p.pos] != '{' {
		return p.argument()
	}
	depth, start := 0, p.pos+1
	for i := p.pos; i < len(p.s); i++ {
		switch p.s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				p.pos = i + 1
				return html.EscapeString(p.s[start:i])
			}
		}
	}
	p.pos = len(p.s)
	return html.EscapeString(p.s[start:])
}

// simple reports whether rich text is a single "word" that needs no
// brackets in a fraction or root.
func simple(rich string) bool {
	plain := Plain(rich)
	for _, r := range plain {
		if unicode.IsSpace(r) || strings.ContainsRune("+−-±×·÷=<>/⁄", r) {
			return false
		}
	}
	return true
}

func bracket(rich string) string {
	if simple(rich) {
		return rich
	}
	return "(" + rich + ")"
}

func (p *texParser) command() string {
	p.pos++ // the backslash
	if p.pos >= len(p.s) {
		return "\\"
	}
	start := p.pos
	if unicode.IsLetter(rune(p.s[p.pos])) {
		for p.pos < len(p.s) && unicode.IsLetter(rune(p.s[p.pos])) {
			p.pos++
		}
	} else {
		p.pos++
	}
	name := p.s[start:p.pos]
	switch name {
	case "frac", "dfrac", "tfrac":
		num, den := p.argument(), p.argument()
		return bracket(num) + "⁄" + bracket(den)
	case "sqrt":
		index := ""
		if p.pos < len(p.s) && p.s[p.pos] == '[' {
			end := strings.IndexByte(p.s[p.pos:], ']')
			if end > 0 {
				index = "<sup>" + html.EscapeString(p.s[p.pos+1:p.pos+end]) + "</sup>"
				p.pos += end + 1
			}
		}
		return index + "√" + bracket(p.argument())
	case "text", "textrm", "mathrm", "mbox", "operatorname", "textit", "mathit", "mathbf", "textbf", "mathsf", "mathtt":
		inner := p.rawArgument()
		switch name {
		case "textbf", "mathbf":
			return "<b>" + inner + "</b>"
		case "textit", "mathit":
			return "<i>" + inner + "</i>"
		}
		return inner
	case "vec", "hat", "bar", "overline", "tilde", "dot":
		marks := map[string]string{"vec": "⃗", "hat": "̂", "bar": "̄", "overline": "̅", "tilde": "̃", "dot": "̇"}
		return p.argument() + marks[name]
	}
	if sym, ok := texSymbols[name]; ok {
		switch name {
		case "pm", "mp", "times", "cdot", "div", "le", "leq", "ge", "geq", "ne", "neq", "approx", "equiv", "to",
			"rightarrow", "Rightarrow", "implies", "iff", "in", "notin", "subset", "subseteq", "cup", "cap", "propto":
			return " " + sym + " "
		}
		return sym
	}
	return html.EscapeString("\\" + name) // unknown: shown as TeX
}

// RichWithMath is s as rich text for Qt: its markup sanitized, and its
// formulas ($...$ and the like) as TeXToRich renders them.
func RichWithMath(s string) string {
	if !HasMath(s) {
		return rubyForQt(Sanitize(s))
	}
	var b strings.Builder
	mathSpans(s, func(tex string) { b.WriteString(TeXToRich(tex)) }, func(t string) { b.WriteString(Sanitize(t)) })
	return rubyForQt(b.String())
}

var rubyRT = regexp.MustCompile(`<rt>(.*?)</rt>`)

// rubyForQt shows furigana the way Qt's rich text can (it has no ruby):
// the reading small, in brackets, after the word: 水 (みず).
func rubyForQt(rich string) string {
	if !strings.Contains(rich, "<ruby>") {
		return rich
	}
	rich = rubyRT.ReplaceAllString(rich, "<span style=\"font-size: small; color: gray;\">\u00a0($1)</span>")
	rich = strings.NewReplacer("<ruby>", "", "</ruby>", "", "<rp>", "", "</rp>", "").Replace(rich)
	return rich
}
