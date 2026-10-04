package userdocumentation

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// The guide exists and every picture it shows is there.
func TestGettingStarted(t *testing.T) {
	t.Setenv("RECUERDO_DATA", filepath.Join("..", "..", "..", ".."))
	html, err := GettingStarted()
	if err != nil {
		t.Fatal(err)
	}
	imgs := regexp.MustCompile(`src="([^"]+)"`).FindAllStringSubmatch(html, -1)
	if len(imgs) < 4 {
		t.Errorf("%d pictures", len(imgs))
	}
	for _, m := range imgs {
		if _, err := os.Stat(filepath.Join(Dir(), m[1])); err != nil {
			t.Errorf("picture %s: %v", m[1], err)
		}
	}
}
