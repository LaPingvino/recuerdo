package lesson

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Loaders for formats OpenTeacher reads, ported from its
// logic/loaders modules.

// loadOTWDFile loads OpenTeaching Words (.otwd): a zip with the word list
// as list.json, as OpenTeacher 3 writes it.
func (fl *FileLoader) loadOTWDFile(path string) (*LessonData, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	f, err := zr.Open("list.json")
	if err != nil {
		return nil, fmt.Errorf("no list.json in %s", path)
	}
	defer f.Close()

	var list struct {
		Title            string `json:"title"`
		QuestionLanguage string `json:"questionLanguage"`
		AnswerLanguage   string `json:"answerLanguage"`
		Items            []struct {
			ID                    int        `json:"id"`
			Questions             [][]string `json:"questions"`
			Answers               [][]string `json:"answers"`
			Comment               string     `json:"comment"`
			CommentAfterAnswering string     `json:"commentAfterAnswering"`
		} `json:"items"`
		Tests []struct {
			Results []struct {
				ItemID int    `json:"itemId"`
				Result string `json:"result"`
				Active struct {
					Start string `json:"start"`
					End   string `json:"end"`
				} `json:"active"`
			} `json:"results"`
		} `json:"tests"`
	}
	if err := json.NewDecoder(f).Decode(&list); err != nil {
		return nil, fmt.Errorf("reading list.json: %w", err)
	}

	data := NewLessonData()
	data.List.Title = list.Title
	if data.List.Title == "" {
		data.List.Title = titleFromPath(path)
	}
	data.List.QuestionLanguage = list.QuestionLanguage
	data.List.AnswerLanguage = list.AnswerLanguage
	for _, it := range list.Items {
		data.List.Items = append(data.List.Items, WordItem{
			ID:        it.ID,
			Questions: flatten(it.Questions),
			Answers:   flatten(it.Answers),
			Comment:   strings.TrimSpace(it.Comment + " " + it.CommentAfterAnswering),
		})
	}
	for _, t := range list.Tests {
		var test Test
		for i, r := range t.Results {
			res := TestResult{ItemID: r.ItemID, Result: r.Result}
			if end, err := time.Parse(otTime, r.Active.End); err == nil {
				res.Time = &end
			}
			if start, err := time.Parse(otTime, r.Active.Start); err == nil && i == 0 {
				test.Date = &start
			}
			test.Results = append(test.Results, res)
		}
		data.List.Tests = append(data.List.Tests, test)
	}
	return data, nil
}

// flatten turns OpenTeacher's groups of words into one list.
func flatten(groups [][]string) []string {
	var out []string
	for _, g := range groups {
		for _, w := range g {
			if w = strings.TrimSpace(w); w != "" {
				out = append(out, w)
			}
		}
	}
	return out
}

// readText reads a file as UTF-8, decoding UTF-16 when it starts with a
// byte order mark (as ABBYY Lingvo Tutor writes its XML).
func readText(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if bytes.HasPrefix(raw, []byte{0xff, 0xfe}) || bytes.HasPrefix(raw, []byte{0xfe, 0xff}) {
		dec := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
		return io.ReadAll(transform.NewReader(bytes.NewReader(raw), dec))
	}
	return bytes.TrimPrefix(raw, []byte{0xef, 0xbb, 0xbf}), nil // UTF-8 byte order mark
}

// isAbbyy reports whether XML text is an ABBYY Lingvo Tutor dictionary.
func isAbbyy(text []byte) bool {
	return bytes.Contains(text, []byte("<dictionary")) && bytes.Contains(text, []byte("<card>"))
}

// loadAbbyy loads an ABBYY Lingvo Tutor dictionary: each card's word is
// the question, its translations the answers and its examples a comment.
func (fl *FileLoader) loadAbbyy(path string, text []byte) (*LessonData, error) {
	var dict struct {
		Title string `xml:"title,attr"`
		Cards []struct {
			Word     string `xml:"word"`
			Meanings []struct {
				Translations []string `xml:"translations>word"`
				Examples     []string `xml:"examples>example"`
			} `xml:"meanings>meaning"`
		} `xml:"card"`
	}
	dec := xml.NewDecoder(bytes.NewReader(text))
	// the text is UTF-8 now, whatever the declaration says
	dec.CharsetReader = func(_ string, r io.Reader) (io.Reader, error) { return r, nil }
	if err := dec.Decode(&dict); err != nil {
		return nil, fmt.Errorf("reading ABBYY dictionary: %w", err)
	}
	data := NewLessonData()
	data.List.Title = dict.Title
	if data.List.Title == "" {
		data.List.Title = titleFromPath(path)
	}
	for i, c := range dict.Cards {
		var answers, examples []string
		for _, m := range c.Meanings {
			answers = append(answers, m.Translations...)
			examples = append(examples, m.Examples...)
		}
		data.List.Items = append(data.List.Items, WordItem{
			ID:        i,
			Questions: fl.parseWordString(strings.TrimSpace(c.Word)),
			Answers:   answers,
			Comment:   strings.Join(examples, ", "),
		})
	}
	return data, nil
}

// isCSVHeader reports whether a spreadsheet's first row is a header, as
// OpenTeacher's CSV saver writes one (the languages or "Questions" and
// "Answers", then "Comment" columns) rather than a word pair.
func isCSVHeader(row []string) bool {
	if len(row) < 2 {
		return false
	}
	if len(row) > 2 && strings.EqualFold(strings.TrimSpace(row[2]), "Comment") {
		return true
	}
	q, a := strings.ToLower(strings.TrimSpace(row[0])), strings.ToLower(strings.TrimSpace(row[1]))
	// "Questions"/"Answers", or Teach2000's "Vraag"/"Antwoord (betekenis)"
	if headerWords[q] && headerWords[withoutNote(a)] {
		return true
	}
	return languageNames[q] && languageNames[a]
}

// headerWords are column names for questions and answers.
var headerWords = map[string]bool{
	"question": true, "questions": true, "answer": true, "answers": true,
	"vraag": true, "vragen": true, "antwoord": true, "antwoorden": true,
}

// withoutNote drops a note in brackets: "antwoord (betekenis)" -> "antwoord".
func withoutNote(s string) string {
	if i := strings.Index(s, " ("); i > 0 {
		return s[:i]
	}
	return s
}

// languageNames are language names in English and in the language itself
// (and Dutch, OpenTeacher's other main language), lower case.
var languageNames = func() map[string]bool {
	m := map[string]bool{}
	for _, n := range strings.Fields(`english engels englisch anglais inglés
		dutch nederlands niederländisch néerlandais holandés
		german duits deutsch allemand alemán
		french frans französisch français francés
		spanish spaans spanisch espagnol español
		italian italiaans italienisch italien italiano
		portuguese portugees portugiesisch portugais português
		latin latijn latein latín latina
		greek grieks griechisch grec griego ελληνικά
		russian russisch russe ruso русский
		polish pools polnisch polonais polski
		swedish zweeds schwedisch suédois svenska
		danish deens dänisch danois dansk
		norwegian noors norwegisch norvégien norsk
		finnish fins finnisch finnois suomi
		turkish turks türkisch turc türkçe
		arabic arabisch arabe árabe العربية
		chinese chinees chinesisch chinois chino 中文
		japanese japans japanisch japonais japonés 日本語
		korean koreaans koreanisch coréen 한국어
		hebrew hebreeuws hebräisch hébreu עברית
		esperanto frisian fries frysk afrikaans indonesian indonesisch hindi persian perzisch
		czech tsjechisch tschechisch čeština hungarian hongaars ungarisch magyar
		romanian roemeens rumänisch română croatian kroatisch hrvatski`) {
		m[n] = true
	}
	return m
}()

// loadCSVRecords turns spreadsheet rows into a lesson: questions and
// answers in the first two columns, a comment in the third, and a header
// row (see isCSVHeader) giving the languages.
func (fl *FileLoader) loadCSVRecords(path string, records [][]string) *LessonData {
	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	if len(records) > 0 && isCSVHeader(records[0]) {
		h := records[0]
		if q, a := strings.TrimSpace(h[0]), strings.TrimSpace(h[1]); languageNames[strings.ToLower(q)] && languageNames[strings.ToLower(a)] {
			data.List.QuestionLanguage, data.List.AnswerLanguage = q, a
		}
		records = records[1:]
	}
	id := 0
	for _, rec := range records {
		if len(rec) < 2 {
			continue
		}
		q := fl.parseWordString(strings.TrimSpace(rec[0]))
		a := fl.parseWordString(strings.TrimSpace(rec[1]))
		if len(q) == 0 || len(a) == 0 {
			continue
		}
		item := WordItem{ID: id, Questions: q, Answers: a}
		if len(rec) > 2 {
			item.Comment = strings.TrimSpace(rec[2])
		}
		data.List.Items = append(data.List.Items, item)
		id++
	}
	return data
}

// readCSV reads all rows of a CSV (or, for .tsv, tab separated) file.
func readCSV(path string, comma rune) ([][]string, error) {
	text, err := readText(path)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(bytes.NewReader(text))
	r.Comma = comma
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	return r.ReadAll()
}

// loadLines loads files with one word pair per line in the word list
// string format ("question = answer" or question<tab>answer): Backpack as
// WRTS exports it, with old Mac line ends. Lines without a pair are
// skipped.
func (fl *FileLoader) loadLines(path string) (*LessonData, error) {
	text, err := readText(path)
	if err != nil {
		return nil, err
	}
	norm := strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(string(text))
	items, err := ParseWordList(norm, true)
	if err != nil {
		return nil, err
	}
	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	data.List.Items = items
	return data, nil
}

// isVTrain reports whether text is a VTrain export: "front=back|" pairs.
func isVTrain(text []byte) bool {
	bars := bytes.Count(text, []byte("|"))
	return bars > 0 && bars == bytes.Count(text, []byte("="))
}

// loadVTrain loads VTrain's text export, ISO-8859-1 "front=back|" pairs.
func (fl *FileLoader) loadVTrain(path string, raw []byte) (*LessonData, error) {
	text, err := charmap.ISO8859_1.NewDecoder().Bytes(raw)
	if err != nil {
		return nil, err
	}
	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	for _, entry := range strings.Split(string(text), "|") {
		q, a, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		qs, as := fl.parseWordString(strings.TrimSpace(q)), fl.parseWordString(strings.TrimSpace(a))
		if len(qs) > 0 && len(as) > 0 {
			data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: qs, Answers: as})
		}
	}
	return data, nil
}
