package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image/png"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/osm"
)

// Maps from OpenStreetMap in the web version, as the desktop makes them
// (lessons/topo/osmdialog.go): search an area, draw it, and put places
// on it by name.

// osmSide is the drawn map's longer side in pixels (the desktop's).
const osmSide = 1400

// OSMStyles are the kinds of map that can be drawn, by name.
func (a *App) OSMStyles() []string {
	var out []string
	for _, s := range osm.Styles {
		out = append(out, s.Name)
	}
	return out
}

// OSMSearch finds areas by name (countries, regions): their names.
func (a *App) OSMSearch(ctx context.Context, query string) ([]string, error) {
	found, err := osm.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	a.osmAreas = nil
	names := []string{}
	for _, ar := range found {
		if len(ar.Border) > 0 { // areas, not points
			a.osmAreas = append(a.osmAreas, ar)
			names = append(names, ar.Name)
		}
	}
	return names, nil
}

// OSMMap draws the found area i in style and starts a map lesson on it.
func (a *App) OSMMap(ctx context.Context, i, style int, title string) (Lesson, error) {
	if i < 0 || i >= len(a.osmAreas) || style < 0 || style >= len(osm.Styles) {
		return Lesson{}, errors.New("search for an area first")
	}
	ar := a.osmAreas[i].Main()
	g := osm.Fit(ar, osmSide)
	img, err := osm.Map(ctx, g, ar, osm.Styles[style], nil)
	if err != nil {
		return Lesson{}, err
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return Lesson{}, err
	}
	l, err := a.NewTopo(title, b.Bytes())
	if err != nil {
		return l, err
	}
	geo, _ := json.Marshal(g)
	a.data.Resources[lesson.MapGeoResource] = geo
	return a.Lesson()
}

// OSMPlace looks a place up on the map (one drawn from OpenStreetMap) and
// adds it where it is.
func (a *App) OSMPlace(ctx context.Context, name string) (Item, error) {
	if a.data == nil {
		return Item{}, ErrNoLesson
	}
	raw, ok := a.data.Resources[lesson.MapGeoResource].([]byte)
	var g osm.Geo
	if !ok || json.Unmarshal(raw, &g) != nil || g.Width == 0 {
		return Item{}, errors.New("this map is not from OpenStreetMap: click where the place is")
	}
	found, err := osm.SearchIn(ctx, name, g)
	if err != nil {
		return Item{}, err
	}
	for _, f := range found {
		if x, y, on := g.ToPicture(f.Lat, f.Lon); on {
			return a.AddPlace(name, x, y)
		}
	}
	return Item{}, errors.New("not found on this map: click where it is")
}
