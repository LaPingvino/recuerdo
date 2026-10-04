package ocrimport

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/ocr"
	"github.com/mappu/miqt/qt"
)

var picture = filepath.Join("..", "..", "..", "..", "ocr", "testdata", "list.png") // 900×420

// Qt runs on the main thread: the dialog checks run in TestMain.
var (
	sizes   = map[string][2]int{}
	read    = map[string][]lesson.WordItem{}
	readErr = map[string]error{}
	newErr  error
)

func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"ocrimport-test"})
	if _, err := New(nil, "no-such-picture.png"); err == nil {
		newErr = errors.New("a missing picture opened")
	}
	d, err := New(nil, picture)
	if err != nil {
		newErr = err
	} else {
		size := func(name string) {
			img := d.Prepared()
			sizes[name] = [2]int{img.Width(), img.Height()}
		}
		size("whole")
		read["whole"], readErr["whole"] = d.Words()
		// drag over the first two rows on the (unscaled-down) preview
		d.setCropFromPreview(int(890*d.scale), int(170*d.scale), int(20*d.scale), int(10*d.scale))
		size("cropped")
		read["cropped"], readErr["cropped"] = d.Words()
		if dir := os.Getenv("SHOT_DIR"); dir != "" {
			d.Show()
			d.Grab().Save(filepath.Join(dir, "import-from-picture.png"))
		}
		d.SetRotation(90)
		size("turned")
	}
	os.Exit(m.Run())
}

func TestToImage(t *testing.T) {
	if x, y := toImage(50, 30, 0.5, 900, 420); x != 100 || y != 60 {
		t.Errorf("%d,%d", x, y)
	}
	if x, y := toImage(-5, 900, 0.5, 900, 420); x != 0 || y != 420 {
		t.Errorf("clamped: %d,%d", x, y)
	}
	if r := spanning(30, 40, 10, 5); r != (Rect{10, 5, 20, 35}) {
		t.Errorf("%+v", r)
	}
}

func TestDialog(t *testing.T) {
	if newErr != nil {
		t.Fatal(newErr)
	}
	if sizes["whole"] != [2]int{900, 420} {
		t.Errorf("whole picture %v", sizes["whole"])
	}
	if c := sizes["cropped"]; c[0] < 860 || c[0] > 880 || c[1] < 155 || c[1] > 165 {
		t.Errorf("cropped to %v, want about 870×160", c)
	}
	if sizes["turned"] != [2]int{420, 900} {
		t.Errorf("turned %v", sizes["turned"])
	}
}

func TestReadsPicture(t *testing.T) {
	if errors.Is(readErr["whole"], ocr.ErrNoTesseract) {
		t.Skip("Tesseract is not installed")
	}
	for name, want := range map[string]int{"whole": 5, "cropped": 2} {
		if readErr[name] != nil {
			t.Fatalf("%s: %v", name, readErr[name])
		}
		if len(read[name]) != want || read[name][0].Answers[0] != "house" {
			t.Errorf("%s: %+v, want %d pairs from huis/house", name, read[name], want)
		}
	}
}
