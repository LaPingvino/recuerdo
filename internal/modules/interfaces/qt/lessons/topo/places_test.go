package topo

import (
	"path/filepath"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

func TestBundledMaps(t *testing.T) {
	maps := BundledMaps(filepath.Join("..", "..", "..", "..", "..", "..", "data", "maps"))
	if len(maps) != 6 {
		t.Fatalf("%d maps", len(maps))
	}
	var europe Map
	for _, m := range maps {
		if m.Name == "Europe" {
			europe = m
		}
	}
	if len(europe.Places) != 31 {
		t.Fatalf("Europe: %+v", europe)
	}
	if p, ok := europe.Find(" amsterdam"); !ok || p.X != 335 || p.Y != 563 {
		t.Errorf("Amsterdam: %+v %v", p, ok)
	}
	if p, ok := europe.Find("brussels"); !ok || p.Name != "Brussels" {
		t.Errorf("other names: %+v %v", p, ok)
	}
}

func TestFit(t *testing.T) {
	f := NewFit(1000, 500, 500, 500) // half size, centred vertically
	if f.Scale != 0.5 || f.DX != 0 || f.DY != 125 {
		t.Fatalf("%+v", f)
	}
	if x, y := f.ToView(200, 100); x != 100 || y != 175 {
		t.Errorf("to view %v,%v", x, y)
	}
	if x, y := f.ToPicture(100, 175); x != 200 || y != 100 {
		t.Errorf("to picture %v,%v", x, y)
	}
	if NewFit(10, 10, 100, 100).Scale != 2 {
		t.Error("small pictures are enlarged at most twice")
	}
}

func TestNearestAndItems(t *testing.T) {
	places := []Place{{ID: 1, Name: "A", X: 10, Y: 10}, {ID: 2, Name: "B", X: 100, Y: 10}}
	if p, ok := Nearest(places, 90, 12, 20); !ok || p.Name != "B" {
		t.Errorf("%+v %v", p, ok)
	}
	if _, ok := Nearest(places, 55, 60, 20); ok {
		t.Error("no place near")
	}
	items := []lesson.WordItem{Item(places[0]), Item(places[1])}
	if got := Places(items); len(got) != 2 || got[1] != places[1] {
		t.Errorf("%+v", got)
	}
	if NextID(items) != 3 {
		t.Error("next id")
	}
}
