package maps

import (
	"os"
	"path/filepath"
	"testing"
)

// All six of OpenTeacher's maps (data/maps) load with their places.
func TestOpenTeacherMapsLoad(t *testing.T) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, "data", "maps")); err != nil {
		t.Skip("data/maps not found")
	}
	mm := NewMapManager(root)
	if err := mm.LoadAvailableMaps(); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"africa", "asia", "europe", "latinamerica", "usa", "world"} {
		m, err := mm.GetMap(id)
		if err != nil {
			t.Errorf("%s: %v", id, err)
			continue
		}
		if len(m.Places) < 10 {
			t.Errorf("%s: %d places", id, len(m.Places))
		}
		if _, err := os.Stat(m.ImagePath); err != nil {
			t.Errorf("%s: map picture: %v", id, err)
		}
	}
}
