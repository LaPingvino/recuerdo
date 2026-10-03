package parser

import (
	"reflect"
	"testing"
)

// Ported from OpenTeacher's logic/wordsString/parserTest.
func TestParse(t *testing.T) {
	cases := []struct {
		name, in string
		want     [][]string
	}{
		{"SingleWord", "one", [][]string{{"one"}}},
		{"MultipleWords", "one, two", [][]string{{"one", "two"}}},
		{"MultipleSemicolonWords", "one; two", [][]string{{"one", "two"}}},
		{"ObligatoryWords", "1. one 2. two", [][]string{{"one"}, {"two"}}},
		{"ObligatoryAndMultipleWords", "1. one, uno 2. two", [][]string{{"one", "uno"}, {"two"}}},
		{"WrongObligatoryNumbers", "1. one 3. two", [][]string{{"one"}, {"two"}}},
		{"NonASCIILetters", "être", [][]string{{"être"}}},
		{"NumbersWithSpacesOnly", "1. 2. 3. 4. ", [][]string{}},
		{"HundredWithSpace", "100 ", [][]string{{"100"}}},
		{"Number", "1.000.000", [][]string{{"1.000.000"}}},
		{"CommasOnly", ",,,", [][]string{}},
		{"SemicolonOnly", ";;;", [][]string{}},
		{"NumberEscaping", `I like to say \1. and \2. You too?`, [][]string{{`I like to say \1. and \2. You too?`}}},
		{"CommaEscaping", `one\, two`, [][]string{{`one\, two`}}},
		{"SemicolonEscaping", `one\; two`, [][]string{{`one\; two`}}},
		{"MultipleDigitObligatoryNumber", "9999999999999999999. one 2222222222222222222222222222. two", [][]string{{"one"}, {"two"}}},
		{"Bug1233809", "Am 20. Mai habe ich ...", [][]string{{"Am 20. Mai habe ich ..."}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Parse(c.in); !reflect.DeepEqual(got, c.want) {
				t.Errorf("Parse(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestModuleType(t *testing.T) {
	m := NewWordsStringParserModule()
	if m.Type() != "wordsStringParser" {
		t.Errorf("Type() = %q", m.Type())
	}
}
