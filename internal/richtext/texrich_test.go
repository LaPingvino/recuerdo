package richtext

import "testing"

func TestTeXToRich(t *testing.T) {
	for tex, want := range map[string]string{
		`\pi r^2`:           "πr<sup>2</sup>",
		`x^{2} + 1`:         "x<sup>2</sup> + 1",
		`H_2O`:              "H<sub>2</sub>O",
		`\frac{a}{b}`:       "a⁄b",
		`\frac{a+1}{2}`:     "(a + 1)⁄2",
		`\sqrt{2}`:          "√2",
		`\sqrt{b^2-4ac}`:    "√(b<sup>2</sup>−4ac)",
		`\sqrt[3]{x}`:       "<sup>3</sup>√x",
		`\alpha \le \beta`:  "α ≤ β",
		`\text{speed} = v`:  "speed = v",
		`\mathbf{F}`:        "<b>F</b>",
		`a \cdot b`:         "a · b",
		`\unknowncommand x`: `\unknowncommandx`,
		`x < y`:             "x &lt; y",
		`\sum_{i=1}^{n} i`:  "∑<sub>i = 1</sub><sup>n</sup>i",
	} {
		if got := TeXToRich(tex); got != want {
			t.Errorf("TeXToRich(%q) = %q, want %q", tex, got, want)
		}
	}
}

func TestRichWithMath(t *testing.T) {
	for in, want := range map[string]string{
		"dog":                            "dog",
		"area: $\\pi r^2$":               "area: πr<sup>2</sup>",
		"H<sub>2</sub>O and $x^2$":       "H<sub>2</sub>O and x<sup>2</sup>",
		`<b>bold</b> <script>x</script>`: "<b>bold</b> ",
		`costs \$5`:                      `costs <span class="dollar">$</span>5`,
	} {
		if got := RichWithMath(in); got != want {
			t.Errorf("RichWithMath(%q) = %q, want %q", in, got, want)
		}
	}
}
