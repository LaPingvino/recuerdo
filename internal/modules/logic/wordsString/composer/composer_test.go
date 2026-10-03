package composer

import (
	"testing"

	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/parser"
)

// Ported from OpenTeacher's logic/wordsString/composerTest.
func TestCompose(t *testing.T) {
	cases := []struct {
		name string
		in   [][]string
		want string
	}{
		{"Empty", [][]string{}, ""},
		{"SingleWord", [][]string{{"one"}}, "one"},
		{"Number", [][]string{{"100"}}, "100"},
		{"MultipleWords", [][]string{{"one", "two"}}, "one, two"},
		{"ObligatoryWords", [][]string{{"one"}, {"two"}}, "1. one 2. two"},
		{"ObligatoryAndMultipleWords", [][]string{{"one", "uno"}, {"two"}}, "1. one, uno 2. two"},
		{"NonASCIILetters", [][]string{{"être"}}, "être"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Compose(c.in); got != c.want {
				t.Errorf("Compose(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// Composing and parsing again gives the same item.
func TestRoundTrip(t *testing.T) {
	for _, text := range []string{"one", "one, two", "1. one, uno 2. two", "être, être 2"} {
		if got := Compose(parser.Parse(text)); got != text {
			t.Errorf("Compose(Parse(%q)) = %q", text, got)
		}
	}
}
