package lesson

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// MapImageResource is the key in LessonData.Resources of a topography
// lesson's map picture (the image's bytes).
const MapImageResource = "mapImage"

type ottpItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// loadOpenTeachingTopoFile reads OpenTeaching Topography (.ottp): a zip
// with the places and results in list.json and the map as map.image.
func (fl *FileLoader) loadOpenTeachingTopoFile(path string) (*LessonData, error) {
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
		Title string     `json:"title"`
		Items []ottpItem `json:"items"`
		Tests []otTest   `json:"tests"`
	}
	if err := json.NewDecoder(f).Decode(&list); err != nil {
		return nil, fmt.Errorf("reading list.json: %w", err)
	}
	data := NewLessonData()
	data.List.Title = list.Title
	if data.List.Title == "" {
		data.List.Title = titleFromPath(path)
	}
	for _, it := range list.Items {
		x, y := it.X, it.Y
		data.List.Items = append(data.List.Items, WordItem{
			ID: it.ID, Name: it.Name, Questions: []string{it.Name}, Answers: []string{it.Name}, X: &x, Y: &y,
		})
	}
	data.List.Tests = fromOTTests(list.Tests)
	if m, err := zr.Open("map.image"); err == nil {
		img, err := io.ReadAll(m)
		m.Close()
		if err != nil {
			return nil, fmt.Errorf("reading the map: %w", err)
		}
		data.Resources[MapImageResource] = img
	}
	return data, nil
}

// saveOpenTeachingTopoFile writes OpenTeaching Topography (.ottp), as
// OpenTeacher 3 does (read back by loadOpenTeachingTopoFile).
func (fs *FileSaver) saveOpenTeachingTopoFile(data *LessonData, path string) error {
	list := struct {
		FileFormatVersion string     `json:"file-format-version"`
		Title             string     `json:"title,omitempty"`
		Items             []ottpItem `json:"items"`
		Tests             []otTest   `json:"tests"`
	}{FileFormatVersion: "3.1", Title: data.List.Title, Items: []ottpItem{}, Tests: toOTTests(data.List.Tests)}
	for _, it := range data.List.Items {
		x, y, ok := it.GetTopoCoordinates()
		if !ok {
			continue
		}
		name := it.Name
		if name == "" && len(it.Questions) > 0 {
			name = it.Questions[0]
		}
		list.Items = append(list.Items, ottpItem{ID: it.ID, Name: name, X: x, Y: y})
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
	if img, ok := data.Resources[MapImageResource].([]byte); ok && err == nil {
		if w, err = zw.Create("map.image"); err == nil {
			_, err = w.Write(img)
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
