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
