package topo

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// Map is one of the maps that come with Recuerdo (data/maps, from
// OpenTeacher's topoMaps): a picture and places known on it.
type Map struct {
	Name   string
	Image  string // path of the picture
	Places []Place
	names  [][]string // all names of each place ("Brussel", "Bruxelles", "Brussels")
}

// Place is a named point on a map, in picture pixels.
type Place struct {
	ID   int
	Name string
	X, Y int
}

var mapName = regexp.MustCompile(`self\.mapName\s*=\s*"([^"]+)"`)

// BundledMaps reads the maps in dir (each a directory with
// resources/map.gif, resources/places.json and the module naming it),
// sorted by name.
func BundledMaps(dir string) []Map {
	entries, _ := os.ReadDir(dir)
	var maps []Map
	for _, e := range entries {
		res := filepath.Join(dir, e.Name(), "resources")
		image := filepath.Join(res, "map.gif")
		if _, err := os.Stat(image); err != nil {
			continue
		}
		m := Map{Name: e.Name(), Image: image}
		if src, err := os.ReadFile(filepath.Join(dir, e.Name(), e.Name()+".py")); err == nil {
			if n := mapName.FindSubmatch(src); n != nil {
				m.Name = string(n[1])
			}
		}
		var known []struct {
			X, Y  int
			Names []string
		}
		if b, err := os.ReadFile(filepath.Join(res, "places.json")); err == nil && json.Unmarshal(b, &known) == nil {
			for i, k := range known {
				if len(k.Names) > 0 {
					m.Places = append(m.Places, Place{ID: i, Name: k.Names[0], X: k.X, Y: k.Y})
					m.names = append(m.names, k.Names)
				}
			}
		}
		maps = append(maps, m)
	}
	sort.Slice(maps, func(i, j int) bool { return maps[i].Name < maps[j].Name })
	return maps
}

// Find is the known place with name (any of its names, ignoring case),
// named as asked for.
func (m Map) Find(name string) (Place, bool) {
	name = strings.TrimSpace(name)
	for i, p := range m.Places {
		for _, n := range m.names[i] {
			if strings.EqualFold(n, name) {
				p.Name = n
				return p, true
			}
		}
	}
	return Place{}, false
}

// Places are a lesson's places.
func Places(items []lesson.WordItem) []Place {
	var out []Place
	for _, it := range items {
		if x, y, ok := it.GetTopoCoordinates(); ok {
			name := it.Name
			if name == "" && len(it.Questions) > 0 {
				name = it.Questions[0]
			}
			out = append(out, Place{ID: it.ID, Name: name, X: x, Y: y})
		}
	}
	return out
}

// Item is a place as a lesson item: its name is the question and answer.
func Item(p Place) lesson.WordItem {
	x, y := p.X, p.Y
	return lesson.WordItem{ID: p.ID, Name: p.Name, Questions: []string{p.Name}, Answers: []string{p.Name}, X: &x, Y: &y}
}

// Fit is how a picture of w×h is shown in a view of vw×vh: scaled to fit
// (never enlarged beyond twice) and centred.
type Fit struct{ Scale, DX, DY float64 }

// NewFit fits a w×h picture in a vw×vh view.
func NewFit(w, h, vw, vh int) Fit {
	if w <= 0 || h <= 0 || vw <= 0 || vh <= 0 {
		return Fit{Scale: 1}
	}
	s := math.Min(math.Min(float64(vw)/float64(w), float64(vh)/float64(h)), 2)
	return Fit{Scale: s, DX: (float64(vw) - float64(w)*s) / 2, DY: (float64(vh) - float64(h)*s) / 2}
}

// ToView converts picture coordinates to the view.
func (f Fit) ToView(x, y int) (float64, float64) {
	return f.DX + float64(x)*f.Scale, f.DY + float64(y)*f.Scale
}

// ToPicture converts view coordinates to the picture.
func (f Fit) ToPicture(vx, vy float64) (int, int) {
	return int(math.Round((vx - f.DX) / f.Scale)), int(math.Round((vy - f.DY) / f.Scale))
}

// Nearest is the place closest to (x, y), if one is within radius
// (all in picture pixels).
func Nearest(places []Place, x, y int, radius float64) (Place, bool) {
	best, bestD := Place{}, math.Inf(1)
	for _, p := range places {
		if d := math.Hypot(float64(p.X-x), float64(p.Y-y)); d < bestD {
			best, bestD = p, d
		}
	}
	return best, bestD <= radius
}

// NextID is an ID not used by any of items.
func NextID(items []lesson.WordItem) int {
	id := 0
	for _, it := range items {
		if it.ID >= id {
			id = it.ID + 1
		}
	}
	return id
}
