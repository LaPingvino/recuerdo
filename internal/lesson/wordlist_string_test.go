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
