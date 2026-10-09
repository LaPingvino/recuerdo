package lesson

import (
	"os"
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

// A map's place on Earth (OpenStreetMap maps) is kept in .ottp files.
func TestOTTPKeepsMapGeo(t *testing.T) {
	data, err := NewFileLoader().LoadFile("../../testdata/legacy_files/application_x-openteachingtopography.openteacher3x.ottp")
	if err != nil {
		t.Fatal(err)
	}
	data.Resources[MapGeoResource] = []byte(`{"zoom":7,"left":16700,"top":10600,"width":400,"height":300}`)
	path := filepath.Join(t.TempDir(), "map.ottp")
	if err := NewFileSaver().SaveFile(data, path); err != nil {
		t.Fatal(err)
	}
	back, err := NewFileLoader().LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if geo, _ := back.Resources[MapGeoResource].([]byte); string(geo) != `{"zoom":7,"left":16700,"top":10600,"width":400,"height":300}` {
		t.Errorf("geo %q", geo)
	}
	if img, _ := back.Resources[MapImageResource].([]byte); len(img) == 0 {
		t.Error("map lost")
	}
}

// A broken lesson file is an error, never a crash: a .wdl asking for an
// impossible length, a truncated one.
func TestBrokenFilesDoNotCrash(t *testing.T) {
	good, err := os.ReadFile("../../testdata/legacy_files/application_x-oriente-voca.voca4.0.wdl")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for name, data := range map[string][]byte{
		"huge.wdl":  append(append([]byte(nil), good[:20]...), 0x3f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff),
		"short.wdl": good[:len(good)/2],
		"empty.wdl": nil,
	} {
		p := filepath.Join(dir, name)
		os.WriteFile(p, data, 0o644)
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: panic %v", name, r)
				}
			}()
			NewFileLoader().LoadFile(p) // an error or a partial list, but no panic
		}()
	}
}

// Saving is atomic: a failed save is an error and leaves the old file;
// a symlink stays a link.
func TestSaveIsAtomic(t *testing.T) {
	data, err := NewFileLoader().LoadFile("../../testdata/lessons/sample.csv")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "words.csv")
	if err := NewFileSaver().SaveFile(data, path); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	if back, err := NewFileLoader().LoadFile(path); err != nil || len(back.List.Items) != len(data.List.Items) {
		t.Fatalf("round trip: %v", err)
	}
	// a directory the save cannot write in: an error, the file unchanged
	os.Chmod(dir, 0o500)
	defer os.Chmod(dir, 0o700)
	if os.Getuid() != 0 {
		if err := NewFileSaver().SaveFile(data, path); err == nil {
			t.Error("a save that could not be written reported success")
		}
		if after, _ := os.ReadFile(path); string(after) != string(before) {
			t.Error("a failed save changed the file")
		}
	}
	os.Chmod(dir, 0o700)
	link := filepath.Join(dir, "link.csv")
	if err := os.Symlink(path, link); err == nil {
		if err := NewFileSaver().SaveFile(data, link); err != nil {
			t.Fatal(err)
		}
		if fi, _ := os.Lstat(link); fi.Mode()&os.ModeSymlink == 0 {
			t.Error("saving through a symlink replaced it")
		}
	}
	if left, _ := filepath.Glob(filepath.Join(dir, ".*saving*")); len(left) > 0 {
		t.Errorf("temporary files left: %v", left)
	}
}

// A lesson saved as .txt reads back the same (Ctrl+S on a .txt lesson
// saves in place).
func TestTextRoundTrip(t *testing.T) {
	data := NewLessonData()
	data.List.Title = "Dieren"
	data.List.Items = []WordItem{
		{ID: 0, Questions: []string{"hond"}, Answers: []string{"dog"}},
		{ID: 1, Questions: []string{"kat", "poes"}, Answers: []string{"cat"}},
		{ID: 2, Questions: []string{"olifant"}, Answers: []string{"elephant", "jumbo"}},
		{ID: 3, Questions: []string{"één"}, Answers: []string{"one"}},
	}
	path := filepath.Join(t.TempDir(), "dieren.txt")
	if err := NewFileSaver().SaveFile(data, path); err != nil {
		t.Fatal(err)
	}
	back, err := NewFileLoader().LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if back.List.Title != "Dieren" || len(back.List.Items) != 4 {
		t.Fatalf("read back %q, %d items", back.List.Title, len(back.List.Items))
	}
	for i, it := range back.List.Items {
		want := data.List.Items[i]
		if strings.Join(it.Questions, "|") != strings.Join(want.Questions, "|") || strings.Join(it.Answers, "|") != strings.Join(want.Answers, "|") {
			t.Errorf("item %d: %v = %v, want %v = %v", i, it.Questions, it.Answers, want.Questions, want.Answers)
		}
	}
}
