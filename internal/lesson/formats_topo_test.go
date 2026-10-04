package lesson

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestTopoFileRoundTrip(t *testing.T) {
	sample := filepath.Join("..", "..", "testdata", "legacy_files", "application_x-openteachingtopography.openteacher3x.ottp")
	data, err := NewFileLoader().LoadFile(sample)
	if err != nil {
		t.Fatal(err)
	}
	img, _ := data.Resources[MapImageResource].([]byte)
	if !bytes.HasPrefix(img, []byte("GIF8")) && !bytes.HasPrefix(img, []byte("\x89PNG")) {
		t.Fatalf("map image not kept: %d bytes", len(img))
	}
	if len(data.List.Items) != 1 || data.List.Items[0].Name != "Test" || *data.List.Items[0].X != 399 {
		t.Fatalf("items %+v", data.List.Items)
	}
	if len(data.List.Tests) != 2 || data.List.Tests[1].Results[0].Result != "right" || data.List.Tests[0].Date == nil {
		t.Fatalf("tests %+v", data.List.Tests)
	}

	out := filepath.Join(t.TempDir(), "copy.ottp")
	if err := NewFileSaver().SaveFile(data, out); err != nil {
		t.Fatal(err)
	}
	again, err := NewFileLoader().LoadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := again.Resources[MapImageResource].([]byte); !bytes.Equal(got, img) {
		t.Errorf("map changed: %d bytes, was %d", len(got), len(img))
	}
	if len(again.List.Items) != 1 || *again.List.Items[0].Y != 364 || len(again.List.Tests) != 2 ||
		!again.List.Tests[0].Results[0].Time.Equal(*data.List.Tests[0].Results[0].Time) {
		t.Errorf("list changed: %+v", again.List)
	}
}
