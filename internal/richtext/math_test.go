package richtext

import "testing"

func TestMath(t *testing.T) {
	for in, want := range map[string]string{
		"x squared":           "x squared",
		"$x^2 + 1$":           "x^2+1",
		`area: $\pi r^2$`:     `area: \pir^2`,
		`$$\frac{a}{b}$$`:     `\frac{a}{b}`,
		`\(a + b\) and \[c\]`: "a+b and c",
		`costs \$5`:           "costs $5",
		"one $ dollar":        "one $ dollar",
		"$x$ <sub>2</sub>":    "x 2",
	} {
		if got := Plain(in); got != want {
			t.Errorf("Plain(%q) = %q, want %q", in, got, want)
		}
	}
	if !HasMath("$x$") || HasMath("5 $ only") || HasMath(`\$5 and \$6`) {
		t.Error("HasMath")
	}
	if NormalizeAnswer(" x^2 + 1 ") != "x^2+1" {
		t.Error("NormalizeAnswer")
	}
}

func TestNormalizeScripts(t *testing.T) {
	for _, c := range [][2]string{
		{`x^{2}+1`, `x^2+1`}, {`x_{1}`, `x_1`}, {`e^{\pi}`, `e^\pi`}, {`\pi r^{2}`, `\pir^2`},
	} {
		if a, b := NormalizeAnswer(c[0]), NormalizeAnswer(c[1]); a != b {
			t.Errorf("%q and %q: %q, %q", c[0], c[1], a, b)
		}
	}
	if NormalizeAnswer(`x^{10}`) == NormalizeAnswer(`x^10`) {
		t.Error("x^{10} is not x^10")
	}
}
