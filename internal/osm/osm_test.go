package osm

import (
	"context"
	"image"
	"image/png"
	"math"
	"os"
	"testing"
)

func TestWorldPixel(t *testing.T) {
	// zoom 0: the world is one 256 pixel tile, (0, 0) in its middle
	if x, y := worldPixel(0, 0, 0); math.Abs(x-128) > 1e-9 || math.Abs(y-128) > 1e-9 {
		t.Errorf("(0,0) at %v,%v", x, y)
	}
	// Amsterdam at zoom 10: tile (525, 336), as OpenStreetMap numbers them
	x, y := worldPixel(52.3676, 4.9041, 10)
	if int(x/256) != 525 || int(y/256) != 336 {
		t.Errorf("Amsterdam in tile %d, %d", int(x/256), int(y/256))
	}
}

func TestFitAndToPicture(t *testing.T) {
	nl := Area{South: 50.75, North: 53.55, West: 3.36, East: 7.23}
	g := Fit(nl, 1200)
	if g.Width > 1200 || g.Height > 1200 || g.Width < 400 || g.Zoom < 6 {
		t.Errorf("fit %+v", g)
	}
	// Amsterdam is on the map, in its upper half; Paris is not
	if x, y, on := g.ToPicture(52.3676, 4.9041); !on || y > g.Height/2 || x < g.Width/4 {
		t.Errorf("Amsterdam at %d,%d (%v) on %+v", x, y, on, g)
	}
	if _, _, on := g.ToPicture(48.8566, 2.3522); on {
		t.Error("Paris on the map of the Netherlands")
	}
}

// Against the real services: only when asked (RECUERDO_OSM_ONLINE=1).
func TestOnline(t *testing.T) {
	if os.Getenv("RECUERDO_OSM_ONLINE") == "" {
		t.Skip("set RECUERDO_OSM_ONLINE=1 to use OpenStreetMap")
	}
	ctx := context.Background()
	as, err := Search(ctx, "Netherlands")
	if err != nil || len(as) == 0 {
		t.Fatalf("search: %v %v", as, err)
	}
	as[0] = as[0].Main()
	g := Fit(as[0], 800)
	if len(as[0].Border) == 0 {
		t.Errorf("no border for %s", as[0].Name)
	}
	img, err := Map(ctx, g, as[0], Styles[0], nil)
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != g.Width || b.Dy() != g.Height {
		t.Errorf("picture %v for %+v", b, g)
	}
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		saveForLook(t, img, dir+"/osm-netherlands.png")
	}
}

func saveForLook(t *testing.T, img interface{ Bounds() image.Rectangle }, path string) {
	f, err := os.Create(path)
	if err != nil {
		t.Log(err)
		return
	}
	defer f.Close()
	png.Encode(f, img.(image.Image))
}

func TestMain_area(t *testing.T) {
	sq := func(x, y, d float64) [][2]float64 {
		return [][2]float64{{x, y}, {x + d, y}, {x + d, y + d}, {x, y + d}, {x, y}}
	}
	a := Area{Border: [][][2]float64{sq(3, 51, 3), sq(5, 53.5, 0.2), sq(-68, 12, 0.3)}} // mainland, island, Bonaire
	m := a.Main()
	if len(m.Border) != 2 || m.West != 3 || m.East != 6 || m.North != 54 || m.South != 51 {
		t.Errorf("main part %+v", m)
	}
}

func TestBounds(t *testing.T) {
	g := Fit(Area{South: 50.75, North: 53.55, West: 3.36, East: 7.23}, 1200)
	b := g.Bounds()
	if b.North < 53.55 || b.South > 50.75 || b.West > 3.36 || b.East < 7.23 || b.North > 54.5 {
		t.Errorf("bounds %+v", b)
	}
	x, y := worldPixel(52.37, 4.9, 9)
	lat, lon := latLon(x, y, 9)
	if math.Abs(lat-52.37) > 1e-9 || math.Abs(lon-4.9) > 1e-9 {
		t.Errorf("round trip %v %v", lat, lon)
	}
}
