package lesson

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// MediaFilesResource is the key in LessonData.Resources of a media
// lesson's embedded files: a map[string][]byte from the item's filename
// ("resources/picture.png") to the file's bytes.
const MediaFilesResource = "mediaFiles"

type otmdItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Remote   bool   `json:"remote"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// MediaFiles are data's embedded media files (made when missing).
func MediaFiles(data *LessonData) map[string][]byte {
	if data.Resources == nil {
		data.Resources = map[string]interface{}{}
	}
	files, ok := data.Resources[MediaFilesResource].(map[string][]byte)
	if !ok {
		files = map[string][]byte{}
		data.Resources[MediaFilesResource] = files
	}
	return files
}

// MediaItem is a media lesson item: a file (embedded, or remote: a URL)
// with a question and an answer.
func MediaItem(id int, name, filename string, remote bool, question, answer string) WordItem {
	r := remote
	f := filename
	it := WordItem{ID: id, Name: name, Filename: &f, Remote: &r}
	if question = strings.TrimSpace(question); question != "" {
		it.Questions = []string{question}
	}
	if answer = strings.TrimSpace(answer); answer != "" {
		it.Answers = []string{answer}
	}
	return it
}

// loadOpenTeachingMediaFile reads OpenTeaching Media (.otmd): a zip with
// the items and results in list.json and the embedded files under
// resources/.
func (fl *FileLoader) loadOpenTeachingMediaFile(path string) (*LessonData, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	f, err := zr.Open("list.json")
	if err != nil {
		return nil, fmt.Errorf("no list.json in %s", path)
	}
	var list struct {
		Title string     `json:"title"`
		Items []otmdItem `json:"items"`
		Tests []otTest   `json:"tests"`
	}
	err = json.NewDecoder(f).Decode(&list)
	f.Close()
	if err != nil {
		return nil, fmt.Errorf("reading list.json: %w", err)
	}
	data := NewLessonData()
	data.List.Title = list.Title
	if data.List.Title == "" {
		data.List.Title = titleFromPath(path)
	}
	files := MediaFiles(data)
	for _, it := range list.Items {
		data.List.Items = append(data.List.Items, MediaItem(it.ID, it.Name, it.Filename, it.Remote, it.Question, it.Answer))
		if it.Remote || it.Filename == "" {
			continue
		}
		m, err := zr.Open(it.Filename)
		if err != nil {
			continue // OpenTeacher too shows nothing for a missing file
		}
		b, err := io.ReadAll(m)
		m.Close()
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", it.Filename, err)
		}
		files[it.Filename] = b
	}
	data.List.Tests = fromOTTests(list.Tests)
	return data, nil
}

// saveOpenTeachingMediaFile writes OpenTeaching Media (.otmd), as
// OpenTeacher 3 does (read back by loadOpenTeachingMediaFile).
func (fs *FileSaver) saveOpenTeachingMediaFile(data *LessonData, path string) error {
	list := struct {
		FileFormatVersion string     `json:"file-format-version"`
		Title             string     `json:"title,omitempty"`
		Items             []otmdItem `json:"items"`
		Tests             []otTest   `json:"tests"`
	}{FileFormatVersion: "3.1", Title: data.List.Title, Items: []otmdItem{}, Tests: toOTTests(data.List.Tests)}
	files := MediaFiles(data)
	used := map[string]bool{}
	for _, it := range data.List.Items {
		o := otmdItem{ID: it.ID, Name: it.Name}
		if it.Filename != nil {
			o.Filename = *it.Filename
		}
		if it.Remote != nil {
			o.Remote = *it.Remote
		}
		if len(it.Questions) > 0 {
			o.Question = strings.Join(it.Questions, ", ")
		}
		if len(it.Answers) > 0 {
			o.Answer = strings.Join(it.Answers, ", ")
		}
		if !o.Remote && files[o.Filename] != nil {
			used[o.Filename] = true
		}
		list.Items = append(list.Items, o)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("list.json")
	if err == nil {
		err = json.NewEncoder(w).Encode(list)
	}
	for name := range used {
		if err != nil {
			break
		}
		if w, err = zw.Create(name); err == nil {
			_, err = w.Write(files[name])
		}
	}
	if cerr := zw.Close(); err == nil {
		err = cerr
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}
