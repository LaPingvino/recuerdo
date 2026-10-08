package osm

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"golang.org/x/image/vector"
)

// A map without names, drawn from OpenFreeMap's vector tiles of
// OpenStreetMap (free, no key): land, water and rivers, the borders of
// countries and provinces; no labels, which would give a topography test
// away.

const (
	openFreeMap    = "https://tiles.openfreemap.org/planet"
	vectorMaxZoom  = 14
	vectorStyleURL = "vector:" // Style.URL of the vector map
)

var (
	tileURLOnce sync.Once
	tileURL     string
	tileURLErr  error
)

// vectorTiles is the URL template of the current tiles (OpenFreeMap
// changes it with each update of the data).
func vectorTiles(ctx context.Context) (string, error) {
	tileURLOnce.Do(func() {
		b, err := get(ctx, openFreeMap)
		if err != nil {
			tileURLErr = err
			return
		}
		var tj struct {
			Tiles []string `json:"tiles"`
		}
		if err := json.Unmarshal(b, &tj); err != nil || len(tj.Tiles) == 0 {
			tileURLErr = fmt.Errorf("no tiles in OpenFreeMap's description")
			return
		}
		tileURL = tj.Tiles[0]
	})
	return tileURL, tileURLErr
}

var (
	landColour     = color.RGBA{0xf4, 0xef, 0xe1, 0xff}
	waterColour    = color.RGBA{0xb9, 0xd3, 0xe8, 0xff}
	riverColour    = color.RGBA{0x9c, 0xc0, 0xde, 0xff}
	provinceColour = color.RGBA{0x9a, 0x9a, 0xa8, 0xff}
	countryColour  = color.RGBA{0x55, 0x55, 0x60, 0xff}
)

// vectorMap draws g from vector tiles.
func vectorMap(ctx context.Context, g Geo, progress func(done, all int)) (*image.RGBA, error) {
	tmpl, err := vectorTiles(ctx)
	if err != nil {
		return nil, err
	}
	img := image.NewRGBA(image.Rect(0, 0, g.Width, g.Height))
	draw.Draw(img, img.Bounds(), image.NewUniform(landColour), image.Point{}, draw.Src)

	// the tiles' zoom (they stop at 14; deeper maps scale them up)
	tz := min(g.Zoom, vectorMaxZoom)
	scale := math.Exp2(float64(g.Zoom - tz)) // map pixels per tile-zoom pixel
	left, top := g.Left/scale, g.Top/scale
	right, bottom := (g.Left+float64(g.Width))/scale, (g.Top+float64(g.Height))/scale
	tx0, ty0 := int(left/tileSize), int(top/tileSize)
	tx1, ty1 := int((right-1)/tileSize), int((bottom-1)/tileSize)
	all := (tx1 - tx0 + 1) * (ty1 - ty0 + 1)
	if all > 100 {
		return nil, fmt.Errorf("the map would need %d tiles", all)
	}

	water := vector.NewRasterizer(g.Width, g.Height)
	rivers := vector.NewRasterizer(g.Width, g.Height)
	provinces := vector.NewRasterizer(g.Width, g.Height)
	countries := vector.NewRasterizer(g.Width, g.Height)
	for _, r := range []*vector.Rasterizer{water, rivers, provinces, countries} {
		r.DrawOp = draw.Over
	}
	n := 1 << tz
	done := 0
	for ty := ty0; ty <= ty1; ty++ {
		for tx := tx0; tx <= tx1; tx++ {
			if ty < 0 || ty >= n {
				continue
			}
			u := strings.NewReplacer("{z}", strconv.Itoa(tz), "{x}", strconv.Itoa(((tx%n)+n)%n), "{y}", strconv.Itoa(ty)).Replace(tmpl)
			b, err := get(ctx, u)
			if err != nil {
				return nil, err
			}
			layers, err := mvt.Unmarshal(b)
			if err != nil {
				if layers, err = mvt.UnmarshalGzipped(b); err != nil {
					return nil, fmt.Errorf("tile %s: %w", u, err)
				}
			}
			// tile coordinates (0..extent) to map pixels
			for _, l := range layers {
				ext := float64(l.Extent)
				if ext == 0 {
					ext = 4096
				}
				pt := func(p orb.Point) (float32, float32) {
					x := (float64(tx)*tileSize + p[0]/ext*tileSize) * scale
					y := (float64(ty)*tileSize + p[1]/ext*tileSize) * scale
					return float32(x - g.Left), float32(y - g.Top)
				}
				for _, f := range l.Features {
					switch l.Name {
					case "water":
						fill(water, f.Geometry, pt)
					case "waterway":
						if c, _ := f.Properties["class"].(string); c == "river" || c == "canal" {
							stroke(rivers, f.Geometry, pt, 0.7*float32(scale))
						}
					case "boundary":
						if m, _ := f.Properties["maritime"].(float64); m == 1 {
							continue
						}
						switch level, _ := f.Properties["admin_level"].(float64); {
						case level == 2:
							stroke(countries, f.Geometry, pt, 1.1)
						case level == 4:
							stroke(provinces, f.Geometry, pt, 0.6)
						}
					}
				}
			}
			done++
			if progress != nil {
				progress(done, all)
			}
		}
	}
	b := img.Bounds()
	water.Draw(img, b, image.NewUniform(waterColour), image.Point{})
	rivers.Draw(img, b, image.NewUniform(riverColour), image.Point{})
	provinces.Draw(img, b, image.NewUniform(provinceColour), image.Point{})
	countries.Draw(img, b, image.NewUniform(countryColour), image.Point{})
	return img, nil
}

func ring(r *vector.Rasterizer, ps []orb.Point, pt func(orb.Point) (float32, float32)) {
	for i, p := range ps {
		x, y := pt(p)
		if i == 0 {
			r.MoveTo(x, y)
		} else {
			r.LineTo(x, y)
		}
	}
	r.ClosePath()
}

// fill adds a polygon geometry to r.
func fill(r *vector.Rasterizer, g orb.Geometry, pt func(orb.Point) (float32, float32)) {
	switch g := g.(type) {
	case orb.Polygon:
		for _, rg := range g {
			ring(r, rg, pt)
		}
	case orb.MultiPolygon:
		for _, p := range g {
			fill(r, p, pt)
		}
	}
}

// stroke adds a line geometry to r as a band w pixels either side.
func stroke(r *vector.Rasterizer, g orb.Geometry, pt func(orb.Point) (float32, float32), w float32) {
	switch g := g.(type) {
	case orb.LineString:
		for i := 1; i < len(g); i++ {
			x0, y0 := pt(g[i-1])
			x1, y1 := pt(g[i])
			dx, dy := x1-x0, y1-y0
			l := float32(math.Hypot(float64(dx), float64(dy)))
			if l == 0 {
				continue
			}
			nx, ny := -dy/l*w, dx/l*w
			r.MoveTo(x0+nx, y0+ny)
			r.LineTo(x1+nx, y1+ny)
			r.LineTo(x1-nx, y1-ny)
			r.LineTo(x0-nx, y0-ny)
			r.ClosePath()
		}
	case orb.MultiLineString:
		for _, ls := range g {
			stroke(r, ls, pt, w)
		}
	}
}
