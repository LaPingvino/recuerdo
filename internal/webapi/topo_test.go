package webapi

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

// A map lesson: a picture, places put on it and moved, saved as .ottp and
// opened again; clicking near a place answers with its name.
func TestTopo(t *testing.T) {
	var pic bytes.Buffer
	_ = png.Encode(&pic, image.NewRGBA(image.Rect(0, 0, 400, 300)))
	var a App
	l, err := a.NewTopo("Benelux", pic.Bytes())
	if err != nil || l.Kind != "topo" || l.Map == "" {
		t.Fatalf("new: %v %+v", err, l.Kind)
	}
	ams, _ := a.AddPlace("Amsterdam", 100, 80)
	if _, err := a.AddPlace("Brussel", 120, 200); err != nil {
		t.Fatal(err)
	}
	if err := a.SetPlace(ams.ID, "Amsterdam", 110, 70); err != nil {
		t.Fatal(err)
	}
	if got := a.NearestPlace(115, 75, 20); got != "Amsterdam" {
		t.Errorf("nearest %q", got)
	}
	if got := a.NearestPlace(300, 10, 20); got != "" {
		t.Errorf("far away: %q", got)
	}
	saved, err := a.Save("Benelux.ottp")
	if err != nil {
		t.Fatal(err)
	}
	var b App
	l, err = b.Open("Benelux.ottp", saved)
	if err != nil || l.Kind != "topo" || len(l.Items) != 2 || *l.Items[0].X != 110 || *l.Items[0].Y != 70 {
		t.Fatalf("opened again: %v %+v", err, l.Items)
	}
	st, err := b.Start(Options{})
	if err != nil || st.ItemID == 0 {
		t.Errorf("practice: %v %+v", err, st)
	}
}
