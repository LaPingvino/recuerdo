package webapi

import (
	"encoding/json"
	"strings"
	"testing"
)

// A media lesson: a picture and a video address added, given answers,
// saved as .otmd and opened again; only the items with an answer are asked.
func TestMedia(t *testing.T) {
	var a App
	if l, err := a.NewMedia("Animals"); err != nil || l.Kind != "media" {
		t.Fatalf("new: %v %+v", err, l)
	}
	dog, err := a.AddMediaFile("dog.png", []byte("\x89PNG\r\n\x1a\nfake"))
	if err != nil || dog.File != "resources/dog.png" || dog.MediaKind != "picture" || dog.Name != "dog" {
		t.Fatalf("file: %v %+v", err, dog)
	}
	vid, err := a.AddMediaAddress("https://www.youtube.com/watch?v=abc")
	if err != nil || vid.MediaKind != "YouTube video" || vid.Media == "" {
		t.Fatalf("address: %v %+v", err, vid)
	}
	if _, err := a.AddMediaAddress("dog"); err == nil {
		t.Error("not an address, added")
	}
	if err := a.SetMedia(dog.ID, "Dog", "Which animal?", "dog"); err != nil {
		t.Fatal(err)
	}
	if b, typ := a.MediaFile(dog.File); len(b) == 0 || typ != "image/png" {
		t.Errorf("media file: %d bytes, %q", len(b), typ)
	}
	st, err := a.Start(Options{})
	if err != nil || st.Total != 1 || st.ItemID != dog.ID {
		t.Fatalf("start: %v %+v", err, st)
	}
	saved, err := a.Save("Animals.otmd")
	if err != nil {
		t.Fatal(err)
	}
	var b App
	l, err := b.Open("Animals.otmd", saved)
	if err != nil || l.Kind != "media" || len(l.Items) != 2 || l.Items[0].Answer != "dog" || l.Items[0].Name != "Dog" {
		t.Fatalf("opened again: %v %+v", err, l)
	}
	if got, _ := b.MediaFile(l.Items[0].File); string(got) != "\x89PNG\r\n\x1a\nfake" {
		t.Errorf("the picture: %q", got)
	}
}

// The first item's id is 0: the page still gets it, to show the item.
func TestMediaStateJSON(t *testing.T) {
	var a App
	a.NewMedia("x")
	it, _ := a.AddMediaFile("dog.png", []byte("x"))
	a.SetMedia(it.ID, "", "", "dog")
	st, _ := a.Start(Options{})
	b, _ := json.Marshal(st)
	if !strings.Contains(string(b), `"itemId":0`) {
		t.Errorf("no itemId: %s", b)
	}
}
