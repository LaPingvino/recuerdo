// Package osm makes topography maps from OpenStreetMap: it finds an area
// by name (Nominatim), downloads the map tiles that cover it and stitches
// them into one picture with OpenStreetMap's credit, and turns latitude
// and longitude into positions on that picture, so that places can be
// put where they really are.
//
// It follows the services' usage policies: a User-Agent that names
// Recuerdo, at most one search a second, and a map is a few dozen tiles
// fetched once.
package osm

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

// UserAgent identifies Recuerdo to the map services, as they ask.
var UserAgent = "Recuerdo (https://github.com/LaPingvino/recuerdo)"

// Style is a kind of map tiles.
type Style struct {
	Name        string
	URL         string // with {z}, {x}, {y}
	Attribution string
}

// Styles are the maps to choose from: without names first, for
// topography (names on the map would give the answers away).
var Styles = []Style{
	{"Map (no names)", vectorStyleURL, "© OpenStreetMap contributors, OpenFreeMap"},
	{"Outline (no names)", "", "© OpenStreetMap contributors"}, // drawn from the area's border
	{"OpenStreetMap", "https://tile.openstreetmap.org/{z}/{x}/{y}.png", "© OpenStreetMap contributors"},
}

const tileSize = 256

// Area is a part of the world: a bounding box in degrees.
type Area struct {
	Name                     string
	South, North, West, East float64
	Lat, Lon                 float64 // its middle, as the search gives it
	// Border is the area's outline: rings of [longitude, latitude]
	// (GeoJSON's order), empty for a point
	Border [][][2]float64
}

// Geo places a map picture on the Earth: the Web Mercator zoom and the
// world pixel of its top left corner. It is stored with a lesson.
type Geo struct {
	Zoom   int     `json:"zoom"`
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
}

// worldPixel is where latitude and longitude are on the Web Mercator
// world map of zoom z, in pixels.
func worldPixel(lat, lon float64, z int) (float64, float64) {
	n := float64(tileSize) * math.Exp2(float64(z))
	lat = math.Max(-85.05112878, math.Min(85.05112878, lat))
	x := (lon + 180) / 360 * n
	s := math.Sin(lat * math.Pi / 180)
	y := (0.5 - math.Log((1+s)/(1-s))/(4*math.Pi)) * n
	return x, y
}

// ToPicture is where latitude and longitude are on the map's picture,
// and whether that is on it.
func (g Geo) ToPicture(lat, lon float64) (int, int, bool) {
	x, y := worldPixel(lat, lon, g.Zoom)
	px, py := int(math.Round(x-g.Left)), int(math.Round(y-g.Top))
	return px, py, px >= 0 && py >= 0 && px < g.Width && py < g.Height
}

// Fit is the map of an area: the largest zoom at which the area fits in
// maxSide pixels, with a margin around it.
func Fit(a Area, maxSide int) Geo {
	for z := 18; z >= 0; z-- {
		x0, y0 := worldPixel(a.North, a.West, z)
		x1, y1 := worldPixel(a.South, a.East, z)
		w, h := x1-x0, y1-y0
		if w*1.1 <= float64(maxSide) && h*1.1 <= float64(maxSide) || z == 0 {
			mx, my := w*0.05, h*0.05
			return Geo{Zoom: z, Left: math.Floor(x0 - mx), Top: math.Floor(y0 - my),
				Width: int(math.Ceil(w + 2*mx)), Height: int(math.Ceil(h + 2*my))}
		}
	}
	return Geo{}
}

var (
	searchMu   sync.Mutex
	lastSearch time.Time
	client     = &http.Client{Timeout: 30 * time.Second}
)

func get(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: %s", u, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 8<<20))
}

// Search finds areas and places by name (Nominatim), at most one search
// a second.
func Search(ctx context.Context, query string) ([]Area, error) { return search(ctx, query, "") }

func search(ctx context.Context, query, extra string) ([]Area, error) {
	searchMu.Lock()
	if wait := time.Second - time.Since(lastSearch); wait > 0 {
		time.Sleep(wait)
	}
	lastSearch = time.Now()
	searchMu.Unlock()
	// with the outline, simplified to what a map of the area shows
	u := "https://nominatim.openstreetmap.org/search?format=jsonv2&limit=8&polygon_geojson=1&polygon_threshold=0.005&q=" +
		url.QueryEscape(query) + extra
	b, err := get(ctx, u)
	if err != nil {
		return nil, err
	}
	var rs []struct {
		Name        string   `json:"display_name"`
		Lat         string   `json:"lat"`
		Lon         string   `json:"lon"`
		BoundingBox []string `json:"boundingbox"` // south, north, west, east
		GeoJSON     struct {
			Type        string          `json:"type"`
			Coordinates json.RawMessage `json:"coordinates"`
		} `json:"geojson"`
	}
	if err := json.Unmarshal(b, &rs); err != nil {
		return nil, err
	}
	var out []Area
	for _, r := range rs {
		if len(r.BoundingBox) != 4 {
			continue
		}
		f := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
		a := Area{Name: r.Name, South: f(r.BoundingBox[0]), North: f(r.BoundingBox[1]),
			West: f(r.BoundingBox[2]), East: f(r.BoundingBox[3]), Lat: f(r.Lat), Lon: f(r.Lon)}
		switch r.GeoJSON.Type {
		case "Polygon":
			json.Unmarshal(r.GeoJSON.Coordinates, &a.Border)
		case "MultiPolygon":
			var polys [][][][2]float64
			if json.Unmarshal(r.GeoJSON.Coordinates, &polys) == nil {
				for _, p := range polys {
					a.Border = append(a.Border, p...)
				}
			}
		}
		out = append(out, a)
	}
	return out, nil
}

// Map downloads the tiles of style that cover g and stitches them into
// its picture, with the style's credit in a corner. progress (may be nil)
// is told the tiles fetched.
func Map(ctx context.Context, g Geo, a Area, style Style, progress func(done, all int)) (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, g.Width, g.Height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0xf2, 0xef, 0xe9, 0xff}), image.Point{}, draw.Src)
	switch style.URL {
	case "":
		outline(img, g, a)
		credit(img, style.Attribution)
		return img, nil
	case vectorStyleURL:
		vimg, err := vectorMap(ctx, g, progress)
		if err != nil {
			return nil, err
		}
		credit(vimg, style.Attribution)
		return vimg, nil
	}
	tx0, ty0 := int(math.Floor(g.Left/tileSize)), int(math.Floor(g.Top/tileSize))
	tx1, ty1 := int(math.Floor((g.Left+float64(g.Width)-1)/tileSize)), int(math.Floor((g.Top+float64(g.Height)-1)/tileSize))
	all := (tx1 - tx0 + 1) * (ty1 - ty0 + 1)
	if all > 100 {
		return nil, fmt.Errorf("the map would need %d tiles", all)
	}
	n := 1 << g.Zoom
	done := 0
	for ty := ty0; ty <= ty1; ty++ {
		for tx := tx0; tx <= tx1; tx++ {
			if ty < 0 || ty >= n {
				continue
			}
			u := strings.NewReplacer("{z}", strconv.Itoa(g.Zoom), "{x}", strconv.Itoa(((tx%n)+n)%n), "{y}", strconv.Itoa(ty)).Replace(style.URL)
			b, err := get(ctx, u)
			if err != nil {
				return nil, err
			}
			tile, _, err := image.Decode(strings.NewReader(string(b)))
			if err != nil {
				return nil, fmt.Errorf("tile %s: %w", u, err)
			}
			at := image.Pt(tx*tileSize-int(g.Left), ty*tileSize-int(g.Top))
			draw.Draw(img, image.Rectangle{at, at.Add(image.Pt(tileSize, tileSize))}, tile, image.Point{}, draw.Src)
			done++
			if progress != nil {
				progress(done, all)
			}
		}
	}
	credit(img, style.Attribution)
	return img, nil
}

// credit writes the map's attribution in its bottom right corner.
func credit(img *image.RGBA, text string) {
	face := basicfont.Face7x13
	w := font.MeasureString(face, text).Ceil()
	b := img.Bounds()
	box := image.Rect(b.Max.X-w-8, b.Max.Y-18, b.Max.X, b.Max.Y)
	draw.Draw(img, box, image.NewUniform(color.NRGBA{255, 255, 255, 200}), image.Point{}, draw.Over)
	d := font.Drawer{Dst: img, Src: image.NewUniform(color.RGBA{60, 60, 60, 255}), Face: face,
		Dot: fixed.P(box.Min.X+4, b.Max.Y-5)}
	d.DrawString(text)
}

// outline draws the area: filled, with its border, on a pale ground.
func outline(img *image.RGBA, g Geo, a Area) {
	b := img.Bounds()
	draw.Draw(img, b, image.NewUniform(color.RGBA{0xdc, 0xe9, 0xf2, 0xff}), image.Point{}, draw.Src)
	if len(a.Border) == 0 {
		return
	}
	pt := func(c [2]float64) (float32, float32) {
		x, y := worldPixel(c[1], c[0], g.Zoom)
		return float32(x - g.Left), float32(y - g.Top)
	}
	land := vector.NewRasterizer(b.Dx(), b.Dy())
	land.DrawOp = draw.Over
	for _, ring := range a.Border {
		for i, c := range ring {
			x, y := pt(c)
			if i == 0 {
				land.MoveTo(x, y)
			} else {
				land.LineTo(x, y)
			}
		}
		land.ClosePath()
	}
	land.Draw(img, b, image.NewUniform(color.RGBA{0xf3, 0xe9, 0xc6, 0xff}), image.Point{})
	// the border: a thin band along each ring
	edge := vector.NewRasterizer(b.Dx(), b.Dy())
	edge.DrawOp = draw.Over
	const w = 1.2
	for _, ring := range a.Border {
		for i := 1; i < len(ring); i++ {
			x0, y0 := pt(ring[i-1])
			x1, y1 := pt(ring[i])
			dx, dy := x1-x0, y1-y0
			l := float32(math.Hypot(float64(dx), float64(dy)))
			if l == 0 {
				continue
			}
			nx, ny := -dy/l*w, dx/l*w
			edge.MoveTo(x0+nx, y0+ny)
			edge.LineTo(x1+nx, y1+ny)
			edge.LineTo(x1-nx, y1-ny)
			edge.LineTo(x0-nx, y0-ny)
			edge.ClosePath()
		}
	}
	edge.Draw(img, b, image.NewUniform(color.RGBA{0x55, 0x55, 0x55, 0xff}), image.Point{})
}

// Main is the area's main part, for a map of it: its largest ring and
// the rings near it (the Wadden islands go with the Netherlands, Bonaire
// does not), with the box around them.
func (a Area) Main() Area {
	if len(a.Border) < 2 {
		return a
	}
	type box struct{ w, e, s, n float64 }
	boxOf := func(r [][2]float64) box {
		b := box{180, -180, 90, -90}
		for _, c := range r {
			b.w, b.e = math.Min(b.w, c[0]), math.Max(b.e, c[0])
			b.s, b.n = math.Min(b.s, c[1]), math.Max(b.n, c[1])
		}
		return b
	}
	big, size := 0, -1.0
	for i, r := range a.Border {
		b := boxOf(r)
		if s := (b.e - b.w) * (b.n - b.s); s > size {
			big, size = i, s
		}
	}
	mb := boxOf(a.Border[big])
	// near: within half the main part's size of it
	mx, my := (mb.e-mb.w)/2, (mb.n-mb.s)/2
	out := a
	out.Border = nil
	all := box{180, -180, 90, -90}
	for _, r := range a.Border {
		b := boxOf(r)
		if b.e < mb.w-mx || b.w > mb.e+mx || b.n < mb.s-my || b.s > mb.n+my {
			continue
		}
		out.Border = append(out.Border, r)
		all.w, all.e = math.Min(all.w, b.w), math.Max(all.e, b.e)
		all.s, all.n = math.Min(all.s, b.s), math.Max(all.n, b.n)
	}
	out.West, out.East, out.South, out.North = all.w, all.e, all.s, all.n
	return out
}

// latLon is the latitude and longitude of a world pixel at zoom z.
func latLon(x, y float64, z int) (float64, float64) {
	n := float64(tileSize) * math.Exp2(float64(z))
	lon := x/n*360 - 180
	lat := math.Atan(math.Sinh(math.Pi*(1-2*y/n))) * 180 / math.Pi
	return lat, lon
}

// Bounds is the part of the world the map shows.
func (g Geo) Bounds() Area {
	north, west := latLon(g.Left, g.Top, g.Zoom)
	south, east := latLon(g.Left+float64(g.Width), g.Top+float64(g.Height), g.Zoom)
	return Area{North: north, West: west, South: south, East: east}
}

// SearchIn finds places by name within a map's area (a city, a river, a
// mountain), the most important first.
func SearchIn(ctx context.Context, query string, g Geo) ([]Area, error) {
	b := g.Bounds()
	return search(ctx, query, fmt.Sprintf("&bounded=1&viewbox=%f,%f,%f,%f", b.West, b.North, b.East, b.South))
}
