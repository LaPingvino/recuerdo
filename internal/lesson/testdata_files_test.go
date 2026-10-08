package lesson

import (
	"path/filepath"
	"strings"
	"testing"
)

// Every lesson file in testdata loads, with items; topography files with
// their map and every place on it.
func TestEveryTestdataFileLoads(t *testing.T) {
	files, _ := filepath.Glob("../../testdata/*/*")
	n := 0
	for _, f := range files {
		base := filepath.Base(f)
		if base == "COPYING" || strings.HasSuffix(base, ".png") { // the KGeography map's picture
			continue
		}
		n++
		data, err := NewFileLoader().LoadFile(f)
		if err != nil {
			t.Errorf("%s: %v", base, err)
			continue
		}
		if len(data.List.Items) == 0 {
			t.Errorf("%s: no items", base)
		}
		if NewFileLoader().GetFileType(f) == "topo" {
			img, _ := data.Resources[MapImageResource].([]byte)
			if len(img) == 0 {
				t.Errorf("%s: a topography lesson without its map", base)
			}
			for _, it := range data.List.Items {
				if it.X == nil || it.Y == nil {
					t.Errorf("%s: %s has no place on the map", base, it.Name)
				}
			}
		}
	}
	if n < 60 {
		t.Errorf("only %d files", n)
	}
}

// The KGeography map of the Netherlands: its twelve provinces, each in
// its own colour on the map, with the frontier, the sea and the
// neighbours left out.
func TestKGeographyMap(t *testing.T) {
	data, err := NewFileLoader().LoadFile("../../testdata/legacy_files/application_x-kgeographymap.kgeography.kgm")
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, it := range data.List.Items {
		names[it.Name] = true
	}
	if data.List.Title != "The Netherlands" || len(names) != 12 || !names["Friesland"] || !names["Zeeland"] || names["Water"] {
		t.Errorf("%q: %v", data.List.Title, names)
	}
}
