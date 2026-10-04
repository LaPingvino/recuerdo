package export

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/printsupport"
)

func init() { runtime.LockOSThread() }

// Qt needs a QApplication on the main thread, where TestMain runs; the
// Qt-based exports run there and the tests report on them.
func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"export-test"})
	dir, _ := os.MkdirTemp("", "export-test")
	results = map[string]error{}
	for _, ext := range []string{".pdf", ".odt", ".docx", ".xlsx"} {
		p := filepath.Join(dir, "dieren"+ext)
		results[ext] = Save(sample(), p)
		paths[ext] = p
	}
	// printing, to a PDF printer
	printer := printsupport.NewQPrinter()
	printer.SetOutputFormat(printsupport.QPrinter__PdfFormat)
	paths["print"] = filepath.Join(dir, "printed.pdf")
	printer.SetOutputFileName(paths["print"])
	results["print"] = Print(sample(), printer)
	printer.Delete()

	// a topography lesson: its map with the places, as PNG, PDF and print
	if topo, err := lesson.NewFileLoader().LoadFile(filepath.Join("..", "..", "..", "..", "..", "testdata",
		"legacy_files", "application_x-openteachingtopography.openteacher3x.ottp")); err != nil {
		results["topo"] = err
	} else {
		topoData = topo
		for _, ext := range []string{".png", ".pdf"} {
			paths["topo"+ext] = filepath.Join(dir, "map"+ext)
			results["topo"+ext] = Save(topo, paths["topo"+ext])
		}
		if img := qt.NewQImage8(paths["topo.png"]); !img.IsNull() {
			topoSize = [2]int{img.Width(), img.Height()}
			x, y, _ := topo.List.Items[0].GetTopoCoordinates()
			topoDot = img.Pixel(x, y)
		}
		tp := printsupport.NewQPrinter()
		tp.SetOutputFormat(printsupport.QPrinter__PdfFormat)
		paths["topo-print"] = filepath.Join(dir, "map-printed.pdf")
		tp.SetOutputFileName(paths["topo-print"])
		results["topo-print"] = Print(topo, tp)
		tp.Delete()
	}

	// a media lesson: a table with a thumbnail, as PDF and printed
	if media, err := lesson.NewFileLoader().LoadFile(filepath.Join("..", "..", "..", "..", "..", "testdata",
		"legacy_files", "application_x-openteachingmedia.openteacher3x.otmd")); err != nil {
		results["media"] = err
	} else {
		mediaData = media
		media.List.Items[1].Questions = []string{"Whose logo?"}
		media.List.Items[1].Answers = []string{"OpenTeacher"}
		paths["media.pdf"] = filepath.Join(dir, "media.pdf")
		results["media.pdf"] = Save(media, paths["media.pdf"])
		mp := printsupport.NewQPrinter()
		mp.SetOutputFormat(printsupport.QPrinter__PdfFormat)
		paths["media-print"] = filepath.Join(dir, "media-printed.pdf")
		mp.SetOutputFileName(paths["media-print"])
		results["media-print"] = Print(media, mp)
		mp.Delete()
	}

	code := m.Run()
	if keep := os.Getenv("EXPORT_KEEP"); keep != "" {
		for ext, p := range paths {
			b, _ := os.ReadFile(p)
			os.WriteFile(filepath.Join(keep, "dieren"+ext), b, 0o644)
		}
	}
	os.RemoveAll(dir)
	os.Exit(code)
}

var (
	topoData  *lesson.LessonData
	mediaData *lesson.LessonData
	topoSize  [2]int
	topoDot   uint
	results   map[string]error
	paths     = map[string]string{}
)

func sample() *lesson.LessonData {
	d := lesson.NewLessonData()
	d.List.Title = "Dieren"
	d.List.Items = []lesson.WordItem{
		{ID: 0, Questions: []string{"hond"}, Answers: []string{"dog"}},
		{ID: 1, Questions: []string{"kat"}, Answers: []string{"cat"}},
	}
	return d
}

func TestPDF(t *testing.T) {
	if err := results[".pdf"]; err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(paths[".pdf"])
	if !bytes.HasPrefix(b, []byte("%PDF")) || len(b) < 1000 {
		t.Errorf("not a PDF: %d bytes, starts %q", len(b), b[:min(len(b), 8)])
	}
}

func TestODT(t *testing.T) {
	if err := results[".odt"]; err != nil {
		t.Fatal(err)
	}
	zr, err := zip.OpenReader(paths[".odt"])
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	f, err := zr.Open("content.xml")
	if err != nil {
		t.Fatal("no content.xml in the ODT")
	}
	content, _ := io.ReadAll(f)
	for _, w := range []string{"hond", "dog", "kat", "cat"} {
		if !bytes.Contains(content, []byte(w)) {
			t.Errorf("ODT lacks %q", w)
		}
	}
}

func TestLibreOfficeFormats(t *testing.T) {
	for _, ext := range []string{".docx", ".xlsx"} {
		err := results[ext]
		if errors.Is(err, ErrNoLibreOffice) {
			t.Skip("LibreOffice is not installed")
		}
		if err != nil {
			t.Fatalf("%s: %v", ext, err)
		}
		// both are zips (Office Open XML)
		if b, _ := os.ReadFile(paths[ext]); !bytes.HasPrefix(b, []byte("PK")) {
			t.Errorf("%s is not an Office Open XML file", ext)
		}
	}
}

func TestSaveFilterAndCanSave(t *testing.T) {
	f := SaveFilter()
	for _, want := range []string{"(*.otwd)", "(*.csv)", "(*.pdf)", "(*.docx)", "(*.xlsx)"} {
		if !strings.Contains(f, want) {
			t.Errorf("filter lacks %s: %s", want, f)
		}
	}
	if !strings.HasPrefix(f, "OpenTeaching Words") {
		t.Errorf("filter should start with OpenTeaching Words: %s", f)
	}
	for path, want := range map[string]bool{"a.otwd": true, "a.PDF": true, "a.xlsx": true, "a.csv": true, "a.apkg": false, "a": false} {
		if CanSave(path) != want {
			t.Errorf("CanSave(%q) = %v", path, !want)
		}
	}
	if got := WithExtension("/tmp/dieren", "PDF (*.pdf)"); got != "/tmp/dieren.pdf" {
		t.Errorf("WithExtension: %q", got)
	}
	if got := WithExtension("/tmp/dieren.csv", "PDF (*.pdf)"); got != "/tmp/dieren.csv" {
		t.Errorf("WithExtension kept extension: %q", got)
	}
}

func TestPrint(t *testing.T) {
	if err := results["print"]; err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(paths["print"])
	if !bytes.HasPrefix(b, []byte("%PDF")) {
		t.Fatalf("not a PDF: %d bytes", len(b))
	}
	// the text is in compressed font streams: read it with pdftotext
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed: PDF written, content not checked")
	}
	out, err := exec.Command("pdftotext", paths["print"], "-").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{"Dieren", "hond", "dog", "kat", "cat"} {
		if !bytes.Contains(out, []byte(w)) {
			t.Errorf("printed page lacks %q:\n%s", w, out)
		}
	}
}

func TestTopoExports(t *testing.T) {
	for _, k := range []string{"topo", "topo.png", "topo.pdf", "topo-print"} {
		if err := results[k]; err != nil {
			t.Fatalf("%s: %v", k, err)
		}
	}
	if topoSize != [2]int{757, 785} {
		t.Errorf("PNG is %v, want the map's 757×785", topoSize)
	}
	if r, g, b := topoDot>>16&0xff, topoDot>>8&0xff, topoDot&0xff; r != 33 || g != 102 || b != 172 {
		t.Errorf("no place dot at the place: pixel %06x", topoDot&0xffffff)
	}
	for _, k := range []string{"topo.pdf", "topo-print"} {
		b, _ := os.ReadFile(paths[k])
		if !bytes.HasPrefix(b, []byte("%PDF")) || len(b) < 10000 {
			t.Errorf("%s: %d bytes, not a PDF with the map", k, len(b))
		}
	}
	if f := SaveFilterFor(topoData); !strings.HasPrefix(f, "OpenTeaching Topography (*.ottp)") || DefaultExtension(topoData) != ".ottp" {
		t.Errorf("topo save filter %q", f)
	}
	if !strings.HasPrefix(SaveFilterFor(sample()), "OpenTeaching Words") || DefaultExtension(sample()) != ".otwd" {
		t.Errorf("words save filter %q", SaveFilterFor(sample()))
	}
}

func TestMediaExports(t *testing.T) {
	for _, k := range []string{"media", "media.pdf", "media-print"} {
		if err := results[k]; err != nil {
			t.Fatalf("%s: %v", k, err)
		}
	}
	if f := SaveFilterFor(mediaData); !strings.HasPrefix(f, "OpenTeaching Media (*.otmd)") || DefaultExtension(mediaData) != ".otmd" {
		t.Errorf("media save filter %q", f)
	}
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed: PDFs written, content not checked")
	}
	for _, k := range []string{"media.pdf", "media-print"} {
		out, err := exec.Command("pdftotext", paths[k], "-").Output()
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range []string{"openteacher-icon.png", "http://openteacher.org/", "Whose logo?", "OpenTeacher", "Question"} {
			if !bytes.Contains(out, []byte(w)) {
				t.Errorf("%s lacks %q:\n%s", k, w, out)
			}
		}
		// the picture is shown, so its file name is not printed in the Medium column
		if bytes.Count(out, []byte("openteacher-icon.png")) != 1 {
			t.Errorf("%s: the picture's name should appear once (as its name), not as its medium", k)
		}
	}
}
