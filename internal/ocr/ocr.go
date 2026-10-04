// Package ocr reads a word list from a picture (a scan or photo of a
// printed list): Tesseract recognises the text lines and their positions
// (hOCR), and lines next to each other become a question and its answer.
// Port of OpenTeacher's logic/ocr/{tesseractRecognizer,wordListLoader}.
package ocr

import (
	"encoding/xml"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// ErrNoTesseract is returned when the tesseract program is not installed.
var ErrNoTesseract = errors.New("reading word lists from pictures needs Tesseract, which was not found")

// rect is a recognised text line and where it is.
type rect struct {
	x, y, w, h int
	text       string
}

// LoadWordList recognises the word list in the picture at path.
func LoadWordList(path string) ([]lesson.WordItem, error) {
	hocr, err := toHOCR(path)
	if err != nil {
		return nil, err
	}
	return wordList(parseHOCR(hocr)), nil
}

// toHOCR runs Tesseract on a picture and returns its hOCR output.
func toHOCR(path string) (string, error) {
	tesseract, err := exec.LookPath("tesseract")
	if err != nil {
		return "", ErrNoTesseract
	}
	dir, err := os.MkdirTemp("", "recuerdo-ocr-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	out := filepath.Join(dir, "page")
	if b, err := exec.Command(tesseract, path, out, "hocr").CombinedOutput(); err != nil {
		return "", errors.New("tesseract: " + strings.TrimSpace(string(b)))
	}
	// newer Tesseract versions write .hocr, older ones .html
	for _, ext := range []string{".hocr", ".html"} {
		if b, err := os.ReadFile(out + ext); err == nil {
			return string(b), nil
		}
	}
	return "", errors.New("tesseract wrote no hOCR output")
}

// parseHOCR returns the words of an hOCR document (spans of class
// ocrx_word with their bounding box), or its text lines (ocr_line) when
// the words have no boxes, as older OCR programs wrote.
func parseHOCR(hocr string) []rect {
	if words := parseSpans(hocr, "ocrx_word"); len(words) > 0 {
		return words
	}
	return parseSpans(hocr, "ocr_line")
}

// parseSpans returns the spans of a class with a bounding box, and their
// text (that of nested spans included).
func parseSpans(hocr, class string) []rect {
	dec := xml.NewDecoder(strings.NewReader(hocr))
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	dec.Entity = xml.HTMLEntity
	var rects []rect
	depth := -1 // span depth inside the current span of the class
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local != "span" {
				continue
			}
			if depth < 0 && attr(t, "class") == class {
				if r, ok := bbox(attr(t, "title")); ok {
					rects = append(rects, r)
					depth = 0
				}
			} else if depth >= 0 {
				depth++
			}
		case xml.EndElement:
			if t.Name.Local == "span" && depth >= 0 {
				depth--
			}
		case xml.CharData:
			if depth >= 0 && len(rects) > 0 {
				rects[len(rects)-1].text += string(t)
			}
		}
	}
	var out []rect
	for _, r := range rects {
		if r.text = strings.TrimSpace(r.text); r.text != "" {
			out = append(out, r)
		}
	}
	return out
}

func attr(e xml.StartElement, name string) string {
	for _, a := range e.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// bbox reads "bbox x0 y0 x1 y1; ..." from an hOCR title.
func bbox(title string) (rect, bool) {
	f := strings.Fields(strings.SplitN(title, ";", 2)[0])
	if len(f) < 5 || f[0] != "bbox" {
		return rect{}, false
	}
	var n [4]int
	for i := range n {
		v, err := strconv.Atoi(f[i+1])
		if err != nil {
			return rect{}, false
		}
		n[i] = v
	}
	return rect{x: n[0], y: n[1], w: n[2] - n[0], h: n[3] - n[1]}, true
}

// wordList turns recognised words (or lines) into word pairs, after
// OpenTeacher: boxes at about the same height form a row; in a row, a box
// close to the previous one (a gap under two text heights) belongs to the
// same cell, a wider gap starts the next column; a row with one cell
// continues the row above; in a row with more, the first cell is the
// question and the last the answer. (OpenTeacher grouped whole lines,
// but Tesseract 5 puts both columns of a row in one line, so words are
// grouped here.)
func wordList(rects []rect) []lesson.WordItem {
	if len(rects) == 0 {
		return nil
	}
	heights := make([]int, len(rects))
	for i, r := range rects {
		heights[i] = r.h
	}
	sort.Ints(heights)
	h := float64(heights[len(heights)/2])

	var pairs [][2][]rect
	for _, row := range detectRows(rects, h*0.5) {
		cells := detectCells(row, h*2)
		switch {
		case len(cells) == 1 && len(pairs) > 0:
			last := &pairs[len(pairs)-1]
			last[1] = append(last[1], cells[0]...)
		case len(cells) >= 2:
			pairs = append(pairs, [2][]rect{cells[0], cells[len(cells)-1]})
		}
	}
	var items []lesson.WordItem
	for _, p := range pairs {
		q, a := joinText(p[0]), joinText(p[1])
		items = append(items, lesson.WordItem{ID: len(items), Questions: words(q), Answers: words(a)})
	}
	return items
}

// detectRows groups boxes whose tops are within margin of the previous.
func detectRows(rects []rect, margin float64) [][]rect {
	sorted := append([]rect{}, rects...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].y < sorted[j].y })
	var rows [][]rect
	for i, r := range sorted {
		if i > 0 && float64(r.y-sorted[i-1].y) < margin {
			rows[len(rows)-1] = append(rows[len(rows)-1], r)
		} else {
			rows = append(rows, []rect{r})
		}
	}
	return rows
}

// detectCells splits a row, left to right, where the gap between boxes
// is at least maxGap.
func detectCells(row []rect, maxGap float64) [][]rect {
	row = append([]rect{}, row...)
	sort.SliceStable(row, func(i, j int) bool { return row[i].x < row[j].x })
	var cells [][]rect
	for i, r := range row {
		if i > 0 && float64(r.x-(row[i-1].x+row[i-1].w)) < maxGap {
			cells[len(cells)-1] = append(cells[len(cells)-1], r)
		} else {
			cells = append(cells, []rect{r})
		}
	}
	return cells
}

func joinText(col []rect) string {
	var parts []string
	for _, r := range col {
		parts = append(parts, r.text)
	}
	return strings.Join(strings.Fields(strings.Join(parts, " ")), " ")
}

// words splits alternatives like OpenTeacher's words string parser.
func words(s string) []string {
	var out []string
	for _, w := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' }) {
		if w = strings.TrimSpace(w); w != "" {
			out = append(out, w)
		}
	}
	return out
}
