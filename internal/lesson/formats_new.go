package lesson

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

// More of OpenTeacher's loaders, ported from its logic/loaders modules.

// loadAPKG loads an Anki package: a zip with the collection as an Anki 2
// database (collection.anki2).
func (fl *FileLoader) loadAPKG(path string) (*LessonData, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	src, err := zr.Open("collection.anki2")
	if err != nil {
		return nil, fmt.Errorf("no collection.anki2 in %s", path)
	}
	defer src.Close()
	tmp, err := os.CreateTemp("", "recuerdo-*.anki2")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	_, err = io.Copy(tmp, src)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return nil, err
	}
	data, err := fl.loadSQLiteFile(tmp.Name())
	if err != nil {
		return nil, err
	}
	if data.List.Title == titleFromPath(tmp.Name()) {
		data.List.Title = titleFromPath(path)
	}
	return data, nil
}

// decodeXML decodes XML that may declare an encoding other than UTF-8
// (the files here are UTF-8 or ASCII in practice).
func decodeXML(r io.Reader, v any) error {
	dec := xml.NewDecoder(r)
	dec.CharsetReader = func(_ string, r io.Reader) (io.Reader, error) { return r, nil }
	return dec.Decode(v)
}

// loadWRTS loads a WRTS list (the Dutch website's XML export).
func (fl *FileLoader) loadWRTS(path string) (*LessonData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var root struct {
		List struct {
			Title string `xml:"titel"`
			LangA string `xml:"taal>a"`
			LangB string `xml:"taal>b"`
			Words []struct {
				A string `xml:"a"`
				B string `xml:"b"`
			} `xml:"woord"`
		} `xml:"lijst"`
	}
	if err := decodeXML(f, &root); err != nil {
		return nil, fmt.Errorf("reading WRTS list: %w", err)
	}
	data := NewLessonData()
	data.List.Title = strings.TrimSpace(root.List.Title)
	if data.List.Title == "" {
		data.List.Title = titleFromPath(path)
	}
	data.List.QuestionLanguage = strings.TrimSpace(root.List.LangA)
	data.List.AnswerLanguage = strings.TrimSpace(root.List.LangB)
	for _, w := range root.List.Words {
		q, a := fl.parseWordString(strings.TrimSpace(w.A)), fl.parseWordString(strings.TrimSpace(w.B))
		if len(q) > 0 || len(a) > 0 {
			data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: q, Answers: a})
		}
	}
	return data, nil
}

// paukerSide is a side of a Pauker card: its text directly or in <Text>.
type paukerSide struct {
	Inner string `xml:",chardata"`
	Text  string `xml:"Text"`
}

func (s paukerSide) text() string {
	if t := strings.TrimSpace(s.Text); t != "" {
		return t
	}
	return strings.TrimSpace(s.Inner)
}

// loadPauker loads a Pauker lesson (.pau XML, or .pau.gz gzipped).
func (fl *FileLoader) loadPauker(path string) (*LessonData, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if bytes.HasPrefix(raw, []byte{0x1f, 0x8b}) {
		zr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		if raw, err = io.ReadAll(zr); err != nil {
			return nil, err
		}
	}
	var lesson struct {
		Description string `xml:"Description"`
		Cards       []struct {
			Front   paukerSide `xml:"FrontSide"`
			Back    paukerSide `xml:"BackSide"`
			Reverse paukerSide `xml:"ReverseSide"`
		} `xml:"Batch>Card"`
	}
	if err := decodeXML(bytes.NewReader(raw), &lesson); err != nil {
		return nil, fmt.Errorf("reading Pauker lesson: %w", err)
	}
	data := NewLessonData()
	data.List.Title, _, _ = strings.Cut(strings.TrimSpace(lesson.Description), "\n")
	if data.List.Title = strings.TrimSpace(data.List.Title); data.List.Title == "" {
		data.List.Title = titleFromPath(strings.TrimSuffix(path, ".gz"))
	}
	for _, c := range lesson.Cards {
		back := c.Back.text()
		if back == "" {
			back = c.Reverse.text()
		}
		q, a := fl.parseWordString(c.Front.text()), fl.parseWordString(back)
		if len(q) > 0 || len(a) > 0 {
			data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: q, Answers: a})
		}
	}
	return data, nil
}

// loadJML loads a jMemorize lesson: lesson.xml in a zip, or that XML on
// its own. Cards are put in the order they were created.
func (fl *FileLoader) loadJML(path string) (*LessonData, error) {
	var r io.Reader
	if zr, err := zip.OpenReader(path); err == nil {
		defer zr.Close()
		f, err := zr.Open("lesson.xml")
		if err != nil {
			return nil, fmt.Errorf("no lesson.xml in %s", path)
		}
		defer f.Close()
		r = f
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}
	var lesson struct {
		Cards []struct {
			Front   string `xml:"Frontside,attr"`
			Back    string `xml:"Backside,attr"`
			Created string `xml:"DateCreated,attr"`
		} `xml:"Category>Deck>Card"`
	}
	if err := decodeXML(r, &lesson); err != nil {
		return nil, fmt.Errorf("reading jMemorize lesson: %w", err)
	}
	cards := lesson.Cards
	created := func(i int) time.Time {
		t, _ := time.Parse("02-Jan-2006 15:04:05", cards[i].Created)
		return t
	}
	sort.SliceStable(cards, func(i, j int) bool { return created(i).Before(created(j)) })

	data := NewLessonData()
	data.List.Title = titleFromPath(path)
	for _, c := range cards {
		q, a := fl.parseWordString(strings.TrimSpace(c.Front)), fl.parseWordString(strings.TrimSpace(c.Back))
		if len(q) > 0 || len(a) > 0 {
			data.List.Items = append(data.List.Items, WordItem{ID: len(data.List.Items), Questions: q, Answers: a})
		}
	}
	return data, nil
}
