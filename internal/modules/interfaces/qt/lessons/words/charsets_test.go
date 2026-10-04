package words

import (
	"path/filepath"
	"testing"
)

// The character sets file in data/ parses (it had flattened quotes that
// broke the JSON) and adds its sets.
func TestCharacterSetsFile(t *testing.T) {
	sets, err := readCharacterSets(filepath.Join("..", "..", "..", "..", "..", "..", "data", "character_sets.json"))
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, s := range sets {
		names[s.Name] = len(s.Characters) > 0
	}
	for _, want := range []string{"Spanish Essentials", "Greek Letters", "Punctuation Extended"} {
		if !names[want] {
			t.Errorf("set %q missing or empty (have %v)", want, names)
		}
	}
}
