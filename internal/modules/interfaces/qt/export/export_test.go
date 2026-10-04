package export

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/mappu/miqt/qt"
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
	results map[string]error
	paths   = map[string]string{}
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
