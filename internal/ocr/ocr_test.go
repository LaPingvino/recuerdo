package ocr

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// hOCR as Tesseract writes it for a printed two-column list.
const sampleHOCR = `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml"><body><div class='ocr_page' title='bbox 0 0 800 600'>
<span class='ocr_line' title="bbox 100 100 200 120; baseline 0 0"><span class='ocrx_word'>huis</span></span>
<span class='ocr_line' title="bbox 400 102 520 122"><span class='ocrx_word'>house,</span> <span class='ocrx_word'>home</span></span>
<span class='ocr_line' title="bbox 100 150 200 170"><span class='ocrx_word'>boom</span></span>
<span class='ocr_line' title="bbox 401 151 500 171"><span class='ocrx_word'>tree</span></span>
<span class='ocr_line' title="bbox 405 175 505 195"><span class='ocrx_word'>(also: tree trunk)</span></span>
</div></body></html>`

// line-level hOCR (words without boxes), as older programs wrote
func TestWordListFromHOCR(t *testing.T) {
	rects := parseHOCR(sampleHOCR)
	if len(rects) != 5 || rects[1].text != "house, home" || rects[0].w != 100 {
		t.Fatalf("lines %+v", rects)
	}
	items := wordList(rects)
	if len(items) != 2 {
		t.Fatalf("items %+v", items)
	}
	if items[0].Questions[0] != "huis" || len(items[0].Answers) != 2 || items[0].Answers[1] != "home" {
		t.Errorf("first item %+v", items[0])
	}
	// a line in only one column continues the row above
	if items[1].Questions[0] != "boom" || items[1].Answers[0] != "tree (also: tree trunk)" {
		t.Errorf("second item %+v", items[1])
	}
}

func TestLoadWordListNeedsTesseract(t *testing.T) {
	img := filepath.Join(t.TempDir(), "list.png")
	os.WriteFile(img, []byte("not really a picture"), 0o644)
	_, err := LoadWordList(img)
	if errors.Is(err, ErrNoTesseract) {
		t.Skip("Tesseract is not installed")
	}
	if err == nil {
		t.Error("a broken picture should give an error")
	}
}

// Tesseract 5: both columns of a row in one line, with word boxes.
const tesseract5HOCR = `<html><body>
<span class='ocr_line' title="bbox 61 21 600 49"><span class='ocrx_word' title='bbox 61 21 140 49'>huis</span>
 <span class='ocrx_word' title='bbox 420 21 520 49'>house</span></span>
<span class='ocr_line' title="bbox 57 91 600 119"><span class='ocrx_word' title='bbox 57 91 110 119'>ijs</span>
 <span class='ocrx_word' title='bbox 420 91 470 119'>ice</span> <span class='ocrx_word' title='bbox 482 91 590 119'>cream</span></span>
</body></html>`

func TestWordListFromWordBoxes(t *testing.T) {
	items := wordList(parseHOCR(tesseract5HOCR))
	if len(items) != 2 || items[0].Questions[0] != "huis" || items[0].Answers[0] != "house" ||
		items[1].Answers[0] != "ice cream" {
		t.Errorf("items %+v", items)
	}
}

// A real run of Tesseract on a picture of a printed list.
func TestRecognisePicture(t *testing.T) {
	items, err := LoadWordList(filepath.Join("testdata", "list.png"))
	if errors.Is(err, ErrNoTesseract) {
		t.Skip("Tesseract is not installed")
	}
	if err != nil {
		t.Fatal(err)
	}
	want := [][2]string{{"huis", "house"}, {"boom", "tree"}, {"fiets", "bicycle"}, {"kaas", "cheese"}, {"ijs", "ice cream"}}
	if len(items) != len(want) {
		t.Fatalf("%d items: %+v", len(items), items)
	}
	for i, w := range want {
		if items[i].Questions[0] != w[0] || items[i].Answers[0] != w[1] {
			t.Errorf("item %d: %q -> %q, want %q -> %q", i, items[i].Questions, items[i].Answers, w[0], w[1])
		}
	}
}
