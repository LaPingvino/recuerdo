package lesson

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	mimicry "github.com/LaPingvino/recuerdo/internal/modules/logic/mimicryTypefaceConverter"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

// Loaders for older Dutch and other programs' formats, ported from
// OpenTeacher's (which are based on inspecting files, not on
// documentation).

// readLegacyText reads a file in a legacy 8-bit encoding as lines, with
// any line ends.
func readLegacyText(path string, enc encoding.Encoding) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text, err := enc.NewDecoder().Bytes(raw)
	if err != nil {
		return nil, err
	}
	norm := strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(string(text))
	return strings.Split(norm, "\n"), nil
}

var overhoorFont = regexp.MustCompile(`\[FONT:([^,\]]*)`)

// loadOverhoor loads Overhoor voor Windows lists (.oh, .ohw: code page
// 850; .oh4: ISO-8859-1): "question = answer" lines, after an optional
// line naming the fonts of both sides, which may be Greek mimicry fonts.
func (fl *FileLoader) loadOverhoor(path string) (*LessonData, error) {
	var enc encoding.Encoding = charmap.CodePage850
	if strings.EqualFold(filepath.Ext(path), ".oh4") {
		enc = charmap.ISO8859_1
	}
	lines, err := readLegacyText(path, enc)
	if err != nil {
		return nil, err
	}
	qFont, aFont := "", ""
	if len(lines) > 0 && strings.HasPrefix(lines[0], "[FONT") {
		if fonts := overhoorFont.FindAllStringSubmatch(lines[0], -1); len(fonts) == 2 {
			qFont, aFont = fonts[0][1], fonts[1][1]
		}
		lines = lines[1:]
	}
	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	for _, line := range lines {
		q, a, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		qs := fl.parseWordString(strings.TrimSpace(mimicry.Convert(qFont, q)))
		as := fl.parseWordString(strings.TrimSpace(mimicry.Convert(aFont, a)))
		if len(qs) > 0 && len(as) > 0 {
			data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: qs, Answers: as})
		}
	}
	return data, nil
}

// loadOVR loads Overhoringsprogramma Talen lists (ISO-8859-1): the
// question and answer languages, four lines of settings, then for each
// word its question and its answers, ended by "-" and "0".
func (fl *FileLoader) loadOVR(path string) (*LessonData, error) {
	lines, err := readLegacyText(path, charmap.ISO8859_1)
	if err != nil {
		return nil, err
	}
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	if len(lines) < 6 {
		return nil, fmt.Errorf("not an Overhoringsprogramma Talen list")
	}
	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	data.List.QuestionLanguage, data.List.AnswerLanguage = lines[0], lines[1]
	rest := lines[6:]
	for len(rest) > 0 {
		question := rest[0]
		rest = rest[1:]
		var answers []string
		for len(rest) >= 2 && !(rest[0] == "-" && rest[1] == "0") {
			answers = append(answers, rest[0])
			rest = rest[1:]
		}
		if len(rest) < 2 {
			break // a word without its end marker: the file ends here
		}
		rest = rest[2:]
		if q := fl.parseWordString(question); len(q) > 0 && len(answers) > 0 {
			data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: q, Answers: answers})
		}
	}
	return data, nil
}

// loadGranule loads a Granule deck (.dkf): XML cards with a front and a
// back.
func (fl *FileLoader) loadGranule(path string) (*LessonData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var deck struct {
		Cards []struct {
			Front string `xml:"front"`
			Back  string `xml:"back"`
		} `xml:"card"`
	}
	if err := decodeXML(f, &deck); err != nil {
		return nil, fmt.Errorf("reading Granule deck: %w", err)
	}
	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	for _, c := range deck.Cards {
		q, a := fl.parseWordString(strings.TrimSpace(c.Front)), fl.parseWordString(strings.TrimSpace(c.Back))
		if len(q) > 0 || len(a) > 0 {
			data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: q, Answers: a})
		}
	}
	return data, nil
}

// loadDomingo loads Domingo lists (.voc, UTF-8): a question line and an
// answer line for each word, up to the first empty line.
func (fl *FileLoader) loadDomingo(path string) (*LessonData, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	norm := strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(string(raw))
	var lines []string
	for _, l := range strings.Split(norm, "\n") {
		if l = strings.TrimSpace(l); l == "" {
			break
		}
		lines = append(lines, l)
	}
	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	for i := 0; i+1 < len(lines); i += 2 {
		q, a := fl.parseWordString(lines[i]), fl.parseWordString(lines[i+1])
		data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: q, Answers: a})
	}
	return data, nil
}

// loadFMD loads a Fresh Memory dictionary (.fmd): the field names give
// the languages; each entry's first fields are the question, the answer
// and a comment.
func (fl *FileLoader) loadFMD(path string) (*LessonData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var dict struct {
		Fields  []string `xml:"fields>field"`
		Entries []struct {
			F []string `xml:"f"`
		} `xml:"entries>e"`
	}
	if err := decodeXML(f, &dict); err != nil {
		return nil, fmt.Errorf("reading Fresh Memory dictionary: %w", err)
	}
	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	if len(dict.Fields) >= 2 && !strings.EqualFold(dict.Fields[0], "Question") {
		data.List.QuestionLanguage, data.List.AnswerLanguage = dict.Fields[0], dict.Fields[1]
	}
	field := func(fs []string, i int) string {
		if i < len(fs) {
			return strings.TrimSpace(fs[i])
		}
		return ""
	}
	for _, e := range dict.Entries {
		q, a := fl.parseWordString(field(e.F, 0)), fl.parseWordString(field(e.F, 1))
		if len(q) > 0 || len(a) > 0 {
			data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: q, Answers: a, Comment: field(e.F, 2)})
		}
	}
	return data, nil
}

// isUTF16 reports whether a file starts with a UTF-16 byte order mark.
func isUTF16(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	head := make([]byte, 2)
	n, _ := f.Read(head)
	return n == 2 && (bytes.Equal(head, []byte{0xff, 0xfe}) || bytes.Equal(head, []byte{0xfe, 0xff}))
}

// loadVocabularium loads Vocabularium lists (.voc, UTF-16): a version
// line, "question-language answer-language", a "!" line used as the
// title, then "question<tab>answer" lines.
func (fl *FileLoader) loadVocabularium(path string) (*LessonData, error) {
	text, err := readText(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(string(text)), "\n")
	data := NewLessonData()
	for i, line := range lines {
		switch {
		case i == 0:
		case i == 1:
			q, a, _ := strings.Cut(strings.TrimSpace(line), " ")
			data.List.QuestionLanguage, data.List.AnswerLanguage = strings.TrimSpace(q), strings.TrimSpace(a)
		case strings.HasPrefix(line, "!") && data.List.Title == "":
			data.List.Title = strings.TrimSpace(line[1:])
		default:
			q, a, ok := strings.Cut(line, "\t")
			if !ok {
				continue
			}
			qs, as := fl.parseWordString(strings.TrimSpace(q)), fl.parseWordString(strings.TrimSpace(a))
			if len(qs) > 0 || len(as) > 0 {
				data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: qs, Answers: as})
			}
		}
	}
	if data.List.Title == "" {
		data.List.Title = titleFromPath(path)
	}
	return data, nil
}

// loadVokabelTrainer loads VokabelTrainer lists (.vtl3, UTF-16 XML): the
// words in <Vokabeln>, their translations in <Uebersetzungen> and comments
// in <Kommentare>. (OpenTeacher's loader took the words as answers too.)
func (fl *FileLoader) loadVokabelTrainer(path string) (*LessonData, error) {
	text, err := readText(path)
	if err != nil {
		return nil, err
	}
	var root struct {
		Items []struct {
			Words        []string `xml:"Vokabeln>string"`
			Translations []string `xml:"Uebersetzungen>string"`
			Comments     []string `xml:"Kommentare>string"`
		} `xml:"Vokabeldatensatz>Datensatz"`
	}
	if err := decodeXML(bytes.NewReader(text), &root); err != nil {
		return nil, fmt.Errorf("reading VokabelTrainer list: %w", err)
	}
	clean := func(ss []string) []string {
		var out []string
		for _, s := range ss {
			if s = strings.TrimSpace(s); s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	for _, it := range root.Items {
		q, a := clean(it.Words), clean(it.Translations)
		if len(q) > 0 || len(a) > 0 {
			data.List.Items = append(data.List.Items, WordItem{
				ID: len(data.List.Items), Questions: q, Answers: a,
				Comment: strings.Join(clean(it.Comments), "; "),
			})
		}
	}
	return data, nil
}
