package richtext

import (
	"strings"
	"unicode/utf16"
)

// The formula builder's palette, shared by the desktop and the web version:
// buttons that insert TeX, so formulas can be made without knowing TeX.

// PaletteItem is one button. In Template, § is where the selected text
// goes (or the cursor, when nothing is selected); {} and [] are the other
// places to fill in.
type PaletteItem struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Template string `json:"template"`
	Tip      string `json:"tip"` // English; the page translates it
}

// PaletteGroup is a row of buttons.
type PaletteGroup struct {
	Name  string        `json:"name"` // English; the page translates it
	Items []PaletteItem `json:"items"`
}

// Palette is the formula builder's buttons.
var Palette = []PaletteGroup{
	{"Structure", []PaletteItem{
		{"sup", "x²", "^{§}", "Power"},
		{"sub", "xₙ", "_{§}", "Index"},
		{"frac", "a⁄b", `\frac{§}{}`, "Fraction"},
		{"sqrt", "√", `\sqrt{§}`, "Square root"},
		{"root", "ⁿ√", `\sqrt[]{§}`, "Root"},
		{"paren", "( )", `\left(§\right)`, "Brackets"},
		{"text", "abc", `\text{§}`, "Text"},
	}},
	{"Operators", []PaletteItem{
		{"pm", "±", `\pm `, ""}, {"times", "×", `\times `, ""}, {"div", "÷", `\div `, ""},
		{"cdot", "·", `\cdot `, ""}, {"le", "≤", `\le `, ""}, {"ge", "≥", `\ge `, ""},
		{"ne", "≠", `\ne `, ""}, {"approx", "≈", `\approx `, ""}, {"infty", "∞", `\infty `, ""},
		{"to", "→", `\to `, ""}, {"implies", "⇒", `\Rightarrow `, ""},
		{"degree", "°", `^{\circ}`, "Degrees"},
	}},
	{"Functions", []PaletteItem{
		{"sum", "∑", `\sum_{§}^{}`, "Sum"},
		{"int", "∫", `\int_{§}^{}`, "Integral"},
		{"lim", "lim", `\lim_{§}`, "Limit"},
		{"sin", "sin", `\sin `, ""}, {"cos", "cos", `\cos `, ""}, {"tan", "tan", `\tan `, ""},
		{"log", "log", `\log `, ""}, {"ln", "ln", `\ln `, ""},
	}},
	{"Greek", greek()},
}

func greek() []PaletteItem {
	var items []PaletteItem
	for _, name := range []string{"alpha", "beta", "gamma", "delta", "epsilon", "theta", "lambda", "mu", "pi",
		"rho", "sigma", "tau", "phi", "omega", "Delta", "Sigma", "Omega"} {
		items = append(items, PaletteItem{name, texSymbols[name], `\` + name + " ", ""})
	}
	return items
}

// PaletteItemByID finds a button by its ID.
func PaletteItemByID(id string) (PaletteItem, bool) {
	for _, g := range Palette {
		for _, it := range g.Items {
			if it.ID == id {
				return it, true
			}
		}
	}
	return PaletteItem{}, false
}

// Expand is what the button inserts when selected is selected (often
// ""), and where the cursor goes in it, counted in UTF-16 code units as
// Qt's and JavaScript's text fields count: at the first place still to
// fill in, else after the inserted text.
func Expand(it PaletteItem, selected string) (text string, cursor int) {
	before, after, hasSlot := strings.Cut(it.Template, "§")
	if !hasSlot {
		return it.Template, utf16Len(it.Template)
	}
	text = before + selected + after
	pos := -1
	for _, slot := range []string{"{}", "[]"} {
		if i := strings.Index(text, slot); i >= 0 && (pos < 0 || i+1 < pos) {
			pos = i + 1
		}
	}
	if selected == "" && (pos < 0 || len(before) < pos) {
		// nothing selected: the § place comes first unless a [] or {}
		// before it does (the n of ⁿ√)
		emptyBefore := strings.Contains(before, "{}") || strings.Contains(before, "[]")
		if !emptyBefore {
			pos = len(before)
		}
	}
	if pos < 0 {
		return text, utf16Len(text)
	}
	return text, utf16Len(text[:pos])
}

func utf16Len(s string) int { return len(utf16.Encode([]rune(s))) }
