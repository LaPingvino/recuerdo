package webapi

import (
	"encoding/base64"
	"errors"
	"math"
	"net/http"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// Topography lessons in the web version: a map (a picture) with places on
// it, in the picture's pixels, as the desktop has them (lessons/topo).

// isTopo reports whether the open lesson is a map with places.
func (a *App) isTopo() bool {
	if a.data == nil {
		return false
	}
	_, ok := a.data.Resources[lesson.MapImageResource].([]byte)
	return ok
}

// mapURL is the lesson's map as a data URL.
func (a *App) mapURL() string {
	img, ok := a.data.Resources[lesson.MapImageResource].([]byte)
	if !ok {
		return ""
	}
	return "data:" + http.DetectContentType(img) + ";base64," + base64.StdEncoding.EncodeToString(img)
}

// NewTopo starts a map lesson on a picture, without places yet.
func (a *App) NewTopo(title string, image []byte) (Lesson, error) {
	if len(image) == 0 {
		return Lesson{}, errors.New("choose a picture of a map")
	}
	d := lesson.NewLessonData()
	d.List.Title = title
	d.Resources[lesson.MapImageResource] = image
	a.data, a.name, a.session = d, title+".ottp", nil
	return a.Lesson()
}

// AddPlace puts a place named name at x, y on the map.
func (a *App) AddPlace(name string, x, y int) (Item, error) {
	if !a.isTopo() {
		return Item{}, errors.New("this lesson has no map")
	}
	id := 1
	for _, it := range a.data.List.Items {
		id = max(id, it.ID+1)
	}
	it := lesson.WordItem{ID: id, Name: name, Questions: []string{name}, Answers: []string{name}, X: &x, Y: &y}
	a.data.List.Items = append(a.data.List.Items, it)
	return itemOf(it), nil
}

// SetPlace renames a place and moves it to x, y.
func (a *App) SetPlace(id int, name string, x, y int) error {
	if a.data == nil {
		return ErrNoLesson
	}
	for i := range a.data.List.Items {
		if it := &a.data.List.Items[i]; it.ID == id {
			it.Name, it.Questions, it.Answers = name, []string{name}, []string{name}
			it.X, it.Y = &x, &y
			return nil
		}
	}
	return errors.New("no such place")
}

// NearestPlace is the name of the place nearest x, y within radius
// pixels ("" if none): clicking the map answers with it.
func (a *App) NearestPlace(x, y int, radius float64) string {
	if a.data == nil {
		return ""
	}
	best, name := radius, ""
	for _, it := range a.data.List.Items {
		if it.X == nil || it.Y == nil {
			continue
		}
		if d := math.Hypot(float64(*it.X-x), float64(*it.Y-y)); d <= best {
			best, name = d, it.Name
		}
	}
	return name
}

// itemOf is a lesson item as the page has it.
func itemOf(it lesson.WordItem) Item {
	q, ans := compose(it.Questions), compose(it.Answers)
	return Item{ID: it.ID, Question: q, Answer: ans, Comment: it.Comment, X: it.X, Y: it.Y}
}
