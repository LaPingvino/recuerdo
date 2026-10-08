package lesson

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
)

// KGeography maps (.kgm) are a picture in which every region has a colour
// of its own, and an XML file naming the regions with their colours. A
// topography lesson is made of them: the picture is the map, each region
// a place at its middle.

type kgmColor struct {
	Red   int `xml:"red"`
	Green int `xml:"green"`
	Blue  int `xml:"blue"`
}

type kgmDivision struct {
	Name    string   `xml:"name"`
	Capital string   `xml:"capital"`
	Ignore  string   `xml:"ignore"`
	Color   kgmColor `xml:"color"`
}

type kgmMap struct {
	XMLName   xml.Name      `xml:"map"`
	Name      string        `xml:"name"`
	MapFile   string        `xml:"mapFile"`
	File      string        `xml:"file"` // older maps
	Divisions []kgmDivision `xml:"division"`
}

// loadKGeographyMapFile reads a KGeography map as a topography lesson.
func (fl *FileLoader) loadKGeographyMapFile(path string) (*LessonData, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m kgmMap
	dec := xml.NewDecoder(bytes.NewReader(raw))
	dec.Strict = false // <!DOCTYPE kgeographyMap> and old files
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("reading the KGeography map: %w", err)
	}
	data := NewLessonData()
	data.List.Title = m.Name
	if data.List.Title == "" {
		data.List.Title = titleFromPath(path)
	}

	picture := m.MapFile
	if picture == "" {
		picture = m.File
	}
	img, imgBytes, err := kgmPicture(filepath.Dir(path), picture)
	if err != nil {
		return nil, fmt.Errorf("the map's picture %s: %w", picture, err)
	}
	data.Resources[MapImageResource] = imgBytes

	id := 0
	for _, d := range m.Divisions {
		if d.Name == "" || strings.EqualFold(strings.TrimSpace(d.Ignore), "yes") {
			continue // the frontier, the sea, the neighbours
		}
		x, y, ok := regionMiddle(img, d.Color)
		if !ok {
			continue // a region the picture does not have
		}
		data.List.Items = append(data.List.Items, WordItem{
			ID: id, Name: d.Name, Questions: []string{d.Name}, Answers: []string{d.Name}, X: &x, Y: &y,
		})
		id++
	}
	if len(data.List.Items) == 0 {
		return nil, fmt.Errorf("no regions of %s found in its picture", data.List.Title)
	}
	return data, nil
}

// kgmPicture loads the map's picture, next to the .kgm (also in another
// case: maps come from systems that do not mind).
func kgmPicture(dir, name string) (image.Image, []byte, error) {
	if name == "" {
		return nil, nil, fmt.Errorf("the map names no picture")
	}
	path := filepath.Join(dir, filepath.Base(name))
	b, err := os.ReadFile(path)
	if err != nil {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if strings.EqualFold(e.Name(), filepath.Base(name)) {
				b, err = os.ReadFile(filepath.Join(dir, e.Name()))
				break
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	return img, b, err
}

// regionMiddle is a place for the region of colour c: its pixels'
// average, or, when that falls outside it (a crescent, islands), the
// region's pixel nearest to it.
func regionMiddle(img image.Image, c kgmColor) (int, int, bool) {
	b := img.Bounds()
	is := func(x, y int) bool {
		r, g, bl, _ := img.At(x, y).RGBA()
		return int(r>>8) == c.Red && int(g>>8) == c.Green && int(bl>>8) == c.Blue
	}
	sx, sy, n := 0, 0, 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if is(x, y) {
				sx, sy, n = sx+x, sy+y, n+1
			}
		}
	}
	if n == 0 {
		return 0, 0, false
	}
	mx, my := sx/n, sy/n
	if is(mx, my) {
		return mx, my, true
	}
	best, bx, by := -1, mx, my
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if is(x, y) {
				if d := (x-mx)*(x-mx) + (y-my)*(y-my); best < 0 || d < best {
					best, bx, by = d, x, y
				}
			}
		}
	}
	return bx, by, true
}
