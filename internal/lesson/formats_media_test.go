package lesson

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"
)

func TestMediaFileRoundTrip(t *testing.T) {
	sample := filepath.Join("..", "..", "testdata", "legacy_files", "application_x-openteachingmedia.openteacher3x.otmd")
	data, err := NewFileLoader().LoadFile(sample)
	if err != nil {
		t.Fatal(err)
	}
	items := data.List.Items
	if len(items) != 2 || !*items[0].Remote || *items[0].Filename != "http://openteacher.org/" ||
		items[0].Questions[0] != "a" || items[0].Answers[0] != "b" || *items[1].Remote || items[1].Questions != nil {
		t.Fatalf("items %+v", items)
	}
	icon := MediaFiles(data)["resources/openteacher-icon.png"]
	if !bytes.HasPrefix(icon, []byte("\x89PNG")) || len(icon) != 2123 {
		t.Fatalf("embedded picture: %d bytes", len(icon))
	}

	now := time.Date(2026, 10, 4, 18, 0, 0, 0, time.UTC)
	data.List.Tests = []Test{{Date: &now, Results: []TestResult{{ItemID: 0, Result: "right", Time: &now}}}}
	out := filepath.Join(t.TempDir(), "copy.otmd")
	if err := NewFileSaver().SaveFile(data, out); err != nil {
		t.Fatal(err)
	}
	again, err := NewFileLoader().LoadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(MediaFiles(again)["resources/openteacher-icon.png"], icon) {
		t.Error("embedded picture changed")
	}
	if len(again.List.Items) != 2 || again.List.Items[0].Answers[0] != "b" || *again.List.Items[1].Filename != "resources/openteacher-icon.png" ||
		len(again.List.Tests) != 1 || again.List.Tests[0].Results[0].Result != "right" {
		t.Errorf("list changed: %+v", again.List)
	}
}
