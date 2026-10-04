package topo

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/mappu/miqt/qt"
)

// Qt runs on the main thread: the widget is driven in TestMain and the
// tests check what happened.
var (
	checks   = map[string]string{} // name -> problem ("" ok)
	modified int
)

func check(name string, ok bool, format string, args ...any) {
	if !ok {
		checks[name] = fmt.Sprintf(format, args...)
	} else if _, seen := checks[name]; !seen {
		checks[name] = ""
	}
}

func shot(w *TopoLessonWidget, tab int, name string) {
	dir := os.Getenv("SHOT_DIR")
	if dir == "" {
		return
	}
	w.tabs.SetCurrentIndex(tab)
	qt.QCoreApplication_ProcessEvents()
	w.Grab().Save(filepath.Join(dir, name))
}

func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	os.Setenv("RECUERDO_DATA", filepath.Join("..", "..", "..", "..", "..", ".."))
	qt.NewQApplication([]string{"topo-test"})
	drive()
	os.Exit(m.Run())
}

func drive() {
	l := lesson.NewLesson("topo")
	w := NewTopoLessonWidget(l, nil)
	w.SetOnModified(func() { modified++ })
	w.Resize(1000, 680)
	w.Show()
	qt.QCoreApplication_ProcessEvents()
	shot(w, 0, "01-empty.png")

	// choose Europe: the map is the lesson's, and its places are known
	var europe Map
	for _, m := range w.maps {
		if m.Name == "Europe" {
			europe = m
		}
	}
	err := w.UseMapPicture(europe.Image)
	img, _ := l.Data.Resources[lesson.MapImageResource].([]byte)
	check("map", err == nil && len(img) > 1000 && w.known != nil && w.known.Name == "Europe",
		"map not used: %v, %d bytes, known %v", err, len(img), w.known)

	w.AddByName("amsterdam")
	w.AddByName("Brussels")
	w.nameEdit.SetText("Atlantis")
	w.AddByName("Atlantis") // unknown: placed by the next click
	w.ClickEnter(100, 900)
	w.ClickEnter(500, 950) // no name typed: "Place 4"
	ps := w.places()
	check("enter", len(ps) == 4 && ps[0].Name == "Amsterdam" && ps[0].X == 335 && ps[1].Name == "Brussels" && ps[2].Name == "Atlantis" &&
		ps[2].X == 100 && ps[3].Name == "Place 4" && w.placeList.Count() == 4,
		"places %+v", ps)
	w.Rename(3, "Ys")
	w.Remove(2)
	ps = w.places()
	check("edit", len(ps) == 3 && ps[2].Name == "Ys" && w.placeList.Item(2).Text() == "Ys", "after edit: %+v", ps)
	w.ClickEnter(337, 560) // near Amsterdam: selects it, adds nothing
	check("select", len(w.places()) == 3 && w.placeList.CurrentRow() == 0, "click near a place: %d places, row %d",
		len(w.places()), w.placeList.CurrentRow())
	shot(w, 0, "02-enter.png")

	// Place – Name: type the marked place's name
	w.Start(PlaceName)
	first, _, _ := w.session.Current()
	shot(w, 1, "03-place-name.png")
	w.AnswerName(first.Name)
	check("feedback", w.feedback.Text() == "✔ Right: "+first.Name, "feedback %q", w.feedback.Text())
	w.AnswerName("Atlantis")
	if w.session != nil {
		second, _, _ := w.session.Current()
		w.AnswerName(second.Name)
	}
	check("place-name", w.session == nil && len(l.Data.List.Tests) == 1, "session should be done and stored")
	if len(l.Data.List.Tests) == 1 {
		r := l.Data.List.Tests[0].Results
		check("results", len(r) == 3 && r[0].Result == "right" && r[1].Result == "wrong" && r[0].ItemID == first.ID,
			"results %+v", r)
	}

	// Name – Place: click the place
	w.Start(NamePlace)
	cur, _, _ := w.session.Current()
	shot(w, 1, "04-name-place.png")
	w.AnswerClick(*cur.X+3, *cur.Y-2)
	cur2, _, _ := w.session.Current()
	other := w.places()[0]
	if other.ID == cur2.ID {
		other = w.places()[1]
	}
	w.AnswerClick(other.X, other.Y)
	w.Stop()
	check("name-place", len(l.Data.List.Tests) == 2 && len(l.Data.List.Tests[1].Results) == 2 &&
		l.Data.List.Tests[1].Results[0].Result == "right" && l.Data.List.Tests[1].Results[1].Result == "wrong",
		"tests %+v", l.Data.List.Tests)
	check("modified", modified > 5 && l.Data.Changed, "modified %d times", modified)
	shot(w, 2, "05-results.png")

	// a lesson file with its own map (OpenTeacher's sample)
	data, err := lesson.NewFileLoader().LoadFile(filepath.Join("..", "..", "..", "..", "..", "..", "testdata",
		"legacy_files", "application_x-openteachingtopography.openteacher3x.ottp"))
	if err == nil {
		sample := lesson.NewLesson("topo")
		sample.Data = *data
		s := NewTopoLessonWidget(sample, nil)
		// its map is OpenTeacher's Africa map: recognised, so its places are known
		check("sample", s.enterMap.HasImage() && s.known != nil && s.known.Name == "Africa" &&
			s.mapCombo.CurrentText() == "Africa" && s.placeList.Count() == 1, "sample map not shown or not recognised")
		s.Resize(800, 560)
		s.Show()
		shot(s, 0, "06-sample.png")
	} else {
		check("sample", false, "%v", err)
	}
}

func TestTopoWidget(t *testing.T) {
	for name, problem := range checks {
		if problem != "" {
			t.Errorf("%s: %s", name, problem)
		}
	}
	if len(checks) < 10 {
		t.Errorf("only %d checks ran", len(checks))
	}
}
