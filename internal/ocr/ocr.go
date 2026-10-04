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

// parseHOCR returns the text lines (spans of class ocr_line, with their
// bounding box in the title) of an hOCR document.
func parseHOCR(hocr string) []rect {
	dec := xml.NewDecoder(strings.NewReader(hocr))
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	dec.Entity = xml.HTMLEntity
	var rects []rect
	depth := -1 // span depth inside the current line, -1 outside lines
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
			if depth < 0 && attr(t, "class") == "ocr_line" {
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
	return rects
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

// wordList turns text lines into word pairs, as OpenTeacher does: lines
// at about the same height form a row, and in each row the lines at
// about the same place form a column; a row with one column continues
// the row above; rows with two columns are a question and its answers.
func wordList(rects []rect) []lesson.WordItem {
	if len(rects) == 0 {
		return nil
	}
	margin := float64(rects[0].h) * 0.5
	rows := detectRows(rects, margin)
	columns := detectColumns(rows, margin)
	var filtered [][]rect
	for _, row := range columns {
		if len(row) == 1 {
			if len(filtered) > 0 {
				filtered[len(filtered)-1] = append(filtered[len(filtered)-1], row[0]...)
			}
			continue
		}
		filtered = append(filtered, append(append([]rect{}, row[0]...), row[len(row)-1]...))
	}
	var items []lesson.WordItem
	for _, row := range detectColumns(filtered, margin) {
		if len(row) != 2 {
			continue
		}
		q, a := joinText(row[0]), joinText(row[1])
		items = append(items, lesson.WordItem{ID: len(items), Questions: words(q), Answers: words(a)})
	}
	return items
}

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

func detectColumns(rows [][]rect, margin float64) [][][]rect {
	margin *= 4
	var table [][][]rect
	for _, row := range rows {
		row = append([]rect{}, row...)
		sort.SliceStable(row, func(i, j int) bool { return row[i].x < row[j].x })
		var cols [][]rect
		for i, r := range row {
			if i > 0 && float64(r.x-row[i-1].x) < margin {
				cols[len(cols)-1] = append(cols[len(cols)-1], r)
			} else {
				cols = append(cols, []rect{r})
			}
		}
		table = append(table, cols)
	}
	return table
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
