package lesson

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"
)

// Savers for formats OpenTeacher writes, ported from its logic/savers.

// composeWords joins a word's alternatives as OpenTeacher writes them.
func composeWords(words []string) string { return strings.Join(words, ", ") }

// saveOTWDFile writes OpenTeaching Words (.otwd): a zip with the list as
// list.json, as OpenTeacher 3 does (read back by loadOTWDFile).
func (fs *FileSaver) saveOTWDFile(data *LessonData, path string) error {
	type item struct {
		ID        int        `json:"id"`
		Questions [][]string `json:"questions"`
		Answers   [][]string `json:"answers"`
		Comment   string     `json:"comment,omitempty"`
	}
	list := struct {
		FileFormatVersion string   `json:"file-format-version"`
		Title             string   `json:"title,omitempty"`
		QuestionLanguage  string   `json:"questionLanguage,omitempty"`
		AnswerLanguage    string   `json:"answerLanguage,omitempty"`
		Items             []item   `json:"items"`
		Tests             []otTest `json:"tests"`
	}{FileFormatVersion: "3.1", Title: data.List.Title, QuestionLanguage: data.List.QuestionLanguage, AnswerLanguage: data.List.AnswerLanguage, Items: []item{}}
	group := func(words []string) [][]string {
		if len(words) == 0 {
			return [][]string{}
		}
		return [][]string{words}
	}
	for _, it := range data.List.Items {
		list.Items = append(list.Items, item{ID: it.ID, Questions: group(it.Questions), Answers: group(it.Answers), Comment: it.Comment})
	}
	list.Tests = toOTTests(data.List.Tests)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("list.json")
	if err == nil {
		err = json.NewEncoder(w).Encode(list)
	}
	if cerr := zw.Close(); err == nil {
		err = cerr
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// saveWRTSFile writes a WRTS list (XML) like OpenTeacher's template.
func (fs *FileSaver) saveWRTSFile(data *LessonData, path string) error {
	type woord struct {
		A string `xml:"a"`
		B string `xml:"b"`
	}
	now := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700")
	doc := struct {
		XMLName xml.Name `xml:"wrts"`
		Lijst   struct {
			ID         string  `xml:"id,attr"`
			Titel      string  `xml:"titel"`
			Datum      string  `xml:"datum"`
			Downloaded string  `xml:"downloaded"`
			Created    string  `xml:"created"`
			Updated    string  `xml:"updated"`
			Auteur     string  `xml:"auteur"`
			TaalA      string  `xml:"taal>a"`
			TaalB      string  `xml:"taal>b"`
			Woorden    []woord `xml:"woord"`
		} `xml:"lijst"`
	}{}
	l := &doc.Lijst
	l.ID, l.Titel = "0000000", data.List.Title
	l.Downloaded, l.Created, l.Updated = now, now, now
	l.Auteur = "Created by: Recuerdo"
	l.TaalA, l.TaalB = data.List.QuestionLanguage, data.List.AnswerLanguage
	for _, it := range data.List.Items {
		l.Woorden = append(l.Woorden, woord{A: composeWords(it.Questions), B: composeWords(it.Answers)})
	}
	out, err := xml.MarshalIndent(doc, "", "\t")
	if err != nil {
		return err
	}
	content := xml.Header + "<!-- Wrts, http://www.wrts.nl/ -->\n" + string(out) + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

// saveSYLKFile writes a SYLK spreadsheet (.slk), which spreadsheet
// programs open: the title, a header row and a row per word.
func (fs *FileSaver) saveSYLKFile(data *LessonData, path string) error {
	var b strings.Builder
	b.WriteString("ID;PRecuerdo\n")
	y := 0
	row := func(cells ...string) {
		y++
		for x, c := range cells {
			// in SYLK a ; in text is written ;; and the text is quoted
			c = strings.ReplaceAll(strings.ReplaceAll(c, ";", ";;"), "\n", " ")
			fmt.Fprintf(&b, "C;X%d;Y%d;K\"%s\"\n", x+1, y, c)
		}
	}
	if data.List.Title != "" {
		row(data.List.Title)
		row()
	}
	row("Questions", "Answers", "Comment")
	for _, it := range data.List.Items {
		row(composeWords(it.Questions), composeWords(it.Answers), it.Comment)
	}
	b.WriteString("E\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
