package spellcheck

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWords(t *testing.T) {
	for in, want := range map[string][]string{
		"the area of a circle":      {"the", "area", "of", "a", "circle"},
		"dog, puppy; hound":         {"dog", "puppy", "hound"},
		"H<sub>2</sub>O":            nil, // H2O: a digit, not a word
		"area: $\\pi r^2$ (cm)":     {"area", "cm"},
		"don't well-known 'quoted'": {"don't", "well-known", "quoted"},
		"<ruby>水<rt>みず</rt></ruby>": {"水"},
		"café naïve":                {"café", "naïve"},
		"1. (a) to go":              {"a", "to", "go"},
	} {
		if got := Words(in); !reflect.DeepEqual(got, want) {
			t.Errorf("Words(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDictionaryFor(t *testing.T) {
	dicts := map[string]string{"nl_NL": "/d", "nl_BE": "/d", "en_GB": "/d", "en_US": "/d", "de_AT": "/d", "fy": "/d"}
	for lang, want := range map[string]string{
		"Dutch": "nl_NL", "Nederlands": "nl_NL", "nl": "nl_NL", "English": "en_US", "en_GB": "en_GB",
		"German": "de_AT", "Frysk": "fy", "Klingon": "", "": "",
	} {
		if got := DictionaryFor(lang, dicts); got != want {
			t.Errorf("DictionaryFor(%q) = %q, want %q", lang, got, want)
		}
	}
}

func TestDictionaries(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "xx_XX.dic"), []byte("1\nword\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "xx_XX.aff"), []byte("SET UTF-8\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "lonely.dic"), []byte("1\nword\n"), 0o644) // no .aff
	t.Setenv("DICPATH", dir)
	d := Dictionaries()
	if d["xx_XX"] != dir || d["lonely"] != "" {
		t.Errorf("dictionaries %v", d)
	}
}

// With hunspell and a Dutch and an English dictionary installed (CI may
// not have them).
func TestHunspell(t *testing.T) {
	nl, en := New("Dutch"), New("English")
	if nl == nil || en == nil {
		t.Skip("hunspell with nl and en dictionaries is not installed")
	}
	wrong := nl.Wrong("hond", "kat, huiss", "de fiets")
	if _, ok := wrong["huiss"]; !ok || len(wrong) != 1 {
		t.Errorf("Dutch: %v", wrong)
	}
	wrong = en.Wrong("dgo, puppy", "the area of a circle", "$x^2$")
	if s, ok := wrong["dgo"]; !ok || len(wrong) != 1 || len(s) == 0 || s[0] != "dog" {
		t.Errorf("English: %v", wrong)
	}
	if again := en.Wrong("dgo"); len(again["dgo"]) == 0 { // from the cache
		t.Errorf("cached: %v", again)
	}
	var none *Checker
	if len(none.Wrong("anything")) != 0 {
		t.Error("a nil checker found something")
	}
}
