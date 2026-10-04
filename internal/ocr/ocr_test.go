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
<span class='ocr_line' title="bbox 100 100 200 120; baseline 0 0"><span class='ocrx_word' title='bbox 100 100 200 120'>huis</span></span>
<span class='ocr_line' title="bbox 400 102 520 122"><span class='ocrx_word'>house,</span> <span class='ocrx_word'>home</span></span>
<span class='ocr_line' title="bbox 100 150 200 170"><span class='ocrx_word'>boom</span></span>
<span class='ocr_line' title="bbox 401 151 500 171"><span class='ocrx_word'>tree</span></span>
<span class='ocr_line' title="bbox 405 175 505 195"><span class='ocrx_word'>(also: tree trunk)</span></span>
</div></body></html>`

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
