package richtext

import "testing"

func TestExpand(t *testing.T) {
	for _, c := range []struct {
		id, selected, text string
		cursor             int
	}{
		{"sqrt", "", `\sqrt{}`, 6},
		{"sqrt", "2", `\sqrt{2}`, 8},
		{"frac", "", `\frac{}{}`, 6},
		{"frac", "a", `\frac{a}{}`, 9},
		{"root", "", `\sqrt[]{}`, 6},
		{"root", "x", `\sqrt[]{x}`, 6},
		{"sup", "", `^{}`, 2},
		{"sup", "n+1", `^{n+1}`, 6},
		{"sum", "", `\sum_{}^{}`, 6},
		{"sum", "i=1", `\sum_{i=1}^{}`, 12},
		{"pi", "", `\pi `, 4},
		{"pi", "x", `\pi `, 4},
		{"paren", "", `\left(\right)`, 6},
		{"text", "één", `\text{één}`, 10},
	} {
		it, ok := PaletteItemByID(c.id)
		if !ok {
			t.Fatalf("no item %q", c.id)
		}
		text, cursor := Expand(it, c.selected)
		if text != c.text || cursor != c.cursor {
			t.Errorf("Expand(%s, %q) = %q, %d; want %q, %d", c.id, c.selected, text, cursor, c.text, c.cursor)
		}
	}
}

func TestPaletteRenders(t *testing.T) {
	// every button's TeX renders without being left as a TeX command
	for _, g := range Palette {
		for _, it := range g.Items {
			text, _ := Expand(it, "x")
			if r := TeXToRich(text); len(r) > 0 && r[0] == '\\' {
				t.Errorf("%s: %q renders as %q", it.ID, text, r)
			}
		}
	}
}
