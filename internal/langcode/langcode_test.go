package langcode

import "testing"

func TestGuessAndName(t *testing.T) {
	for name, want := range map[string]string{
		"Dutch": "nl", "nederlands": "nl", "English": "en", "French": "fr", "français": "fr",
		"German": "de", "Deutsch": "de", " Greek ": "el", "Klingon-ish": "",
	} {
		if got := Guess(name); got != want {
			t.Errorf("Guess(%q) = %q, want %q", name, got, want)
		}
	}
	if Name("nl") != "Dutch" || Name("xx") != "" {
		t.Errorf("Name: %q %q", Name("nl"), Name("xx"))
	}
}
