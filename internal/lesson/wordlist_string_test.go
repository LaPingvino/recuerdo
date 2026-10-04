package lesson

import (
	"reflect"
	"testing"
)

func TestWordListString(t *testing.T) {
	items, err := ParseWordList("een = one\ntwee\ttwo, deux\n\n1 \\= 1 = one equals one\n", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 || !reflect.DeepEqual(items[1].Answers, []string{"two", "deux"}) ||
		items[2].Questions[0] != "1 = 1" || items[2].Answers[0] != "one equals one" {
		t.Errorf("parsed %+v", items)
	}
	if _, err := ParseWordList("no separator here", false); err == nil {
		t.Error("a line without separator should be an error")
	}
	if items, _ := ParseWordList("skip me\na = b", true); len(items) != 1 {
		t.Errorf("lenient: %+v", items)
	}
	// composing and parsing again gives the same words
	again, err := ParseWordList(ComposeWordList(items), false)
	if err != nil || !reflect.DeepEqual(again, items) {
		t.Errorf("round trip: %+v, %v", again, err)
	}
}

// Formulas and HTML tags keep their commas, semicolons and equals signs.
func TestParseWordListFormulasAndTags(t *testing.T) {
	items, err := ParseWordList("a function = $f(x, y)$, $g$\n"+
		"$a = b$ = an equation\n"+
		`a dog = <img src="data:image/png;base64,AA,BB" alt="dog">`+"\n"+
		`price = \$5, five dollars`+"\n", false)
	if err != nil || len(items) != 4 {
		t.Fatalf("%+v %v", items, err)
	}
	if got := items[0].Answers; len(got) != 2 || got[0] != "$f(x, y)$" || got[1] != "$g$" {
		t.Errorf("formula with a comma: %q", got)
	}
	if items[1].Questions[0] != "$a = b$" || items[1].Answers[0] != "an equation" {
		t.Errorf("formula with an equals sign: %q = %q", items[1].Questions, items[1].Answers)
	}
	if got := items[2].Answers; len(got) != 1 || got[0] != `<img src="data:image/png;base64,AA,BB" alt="dog">` {
		t.Errorf("tag with commas: %q", got)
	}
	if got := items[3].Answers; len(got) != 2 || got[0] != `\$5` {
		t.Errorf("escaped dollar: %q", got)
	}
}
