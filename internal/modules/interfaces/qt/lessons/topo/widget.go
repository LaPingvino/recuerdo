package topo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/charts"
	"github.com/LaPingvino/recuerdo/internal/resources"
	"github.com/LaPingvino/recuerdo/internal/teaching"
	"github.com/mappu/miqt/qt"
)

// The two ways of practising, as in OpenTeacher.
const (
	PlaceName = "Place – Name" // a place is shown: type its name
	NamePlace = "Name – Place" // a name is shown: click the place
)

const defaultHint = "Click the map to add a place there, or type the name of a place the map knows."

// TopoLessonWidget shows a topography lesson: places on a map to enter,
// practise and see the results of.
type TopoLessonWidget struct {
	*qt.QWidget
	lesson     *lesson.Lesson
	maps       []Map
	known      *Map // the bundled map shown, for its known places
	onModified func()
	tabs       *qt.QTabWidget

	// Enter
	mapCombo  *qt.QComboBox
	enterMap  *MapView
	nameEdit  *qt.QLineEdit
	placeList *qt.QListWidget
	enterHint *qt.QLabel
	filling   bool // the list is being filled: ignore its edits
	removeBtn *qt.QPushButton
	addAllBtn *qt.QPushButton

	// Teach
	orderCombo, typeCombo, sequenceCombo *qt.QComboBox
	startBtn                             *qt.QPushButton
	prompt, feedback, progress           *qt.QLabel
	answerRow                            *qt.QWidget
	answerEdit                           *qt.QLineEdit
	teachMap                             *MapView
	session                              *teaching.Session
	order                                string

	// Results
	summary  *qt.QLabel
	grades   *charts.GradesChart
	timeline *charts.TimelineChart
}

// NewTopoLessonWidget creates the widget for a topography lesson.
func NewTopoLessonWidget(l *lesson.Lesson, parent *qt.QWidget) *TopoLessonWidget {
	w := &TopoLessonWidget{QWidget: qt.NewQWidget(parent), lesson: l,
		maps: BundledMaps(filepath.Join(resources.Dir(), "data", "maps"))}
	layout := qt.NewQVBoxLayout(w.QWidget)
	layout.SetContentsMargins(0, 0, 0, 0)
	w.tabs = qt.NewQTabWidget(w.QWidget)
	layout.AddWidget(w.tabs.QWidget)
	w.tabs.AddTab(w.enterTab(), "Enter")
	w.tabs.AddTab(w.teachTab(), "Teach")
	w.tabs.AddTab(w.resultsTab(), "Results")
	w.tabs.OnCurrentChanged(func(int) { w.refresh() })
	w.loadMap()
	w.refresh()
	return w
}

// SetOnModified calls f after every change to the lesson.
func (w *TopoLessonWidget) SetOnModified(f func()) { w.onModified = f }

func (w *TopoLessonWidget) modified() {
	w.lesson.Data.Changed = true
	if w.onModified != nil {
		w.onModified()
	}
}

func (w *TopoLessonWidget) places() []Place { return Places(w.lesson.Data.List.Items) }

// ---- Enter ----

func (w *TopoLessonWidget) enterTab() *qt.QWidget {
	tab := qt.NewQWidget(nil)
	layout := qt.NewQVBoxLayout(tab)

	top := qt.NewQHBoxLayout2()
	mapLabel := qt.NewQLabel3("Map:")
	top.AddWidget(mapLabel.QWidget)
	w.mapCombo = qt.NewQComboBox(tab)
	w.mapCombo.AddItem("(choose a map)")
	for _, m := range w.maps {
		w.mapCombo.AddItem(m.Name)
	}
	w.mapCombo.OnActivated(func(i int) {
		if i > 0 {
			w.useBundledMap(w.maps[i-1])
		}
	})
	top.AddWidget(w.mapCombo.QWidget)
	other := qt.NewQPushButton3("Other picture…")
	other.SetToolTip("Use a picture of your own as the map")
	other.OnClicked(func() {
		path := qt.QFileDialog_GetOpenFileName4(w.QWidget, "Choose a map", "",
			"Pictures (*.png *.jpg *.jpeg *.gif *.bmp *.webp);;All files (*)")
		if path != "" {
			if err := w.UseMapPicture(path); err != nil {
				qt.QMessageBox_Warning(w.QWidget, "Choose a map", err.Error())
			}
		}
	})
	top.AddWidget(other.QWidget)
	w.addAllBtn = qt.NewQPushButton3("Add all places")
	w.addAllBtn.SetToolTip("Add every place this map knows")
	w.addAllBtn.OnClicked(w.AddAllKnown)
	top.AddWidget(w.addAllBtn.QWidget)
	top.AddStretch()
	layout.AddLayout(top.QLayout)

	body := qt.NewQHBoxLayout2()
	w.enterMap = NewMapView(tab)
	w.enterMap.OnClick(w.ClickEnter)
	body.AddWidget2(w.enterMap.QWidget, 3)

	side := qt.NewQVBoxLayout2()
	addRow := qt.NewQHBoxLayout2()
	w.nameEdit = qt.NewQLineEdit(tab)
	w.nameEdit.SetPlaceholderText("Name of a place")
	w.nameEdit.OnReturnPressed(func() { w.AddByName(w.nameEdit.Text()) })
	addRow.AddWidget(w.nameEdit.QWidget)
	add := qt.NewQPushButton3("Add")
	add.OnClicked(func() { w.AddByName(w.nameEdit.Text()) })
	addRow.AddWidget(add.QWidget)
	side.AddLayout(addRow.QLayout)
	w.enterHint = qt.NewQLabel3(defaultHint)
	w.enterHint.SetWordWrap(true)
	side.AddWidget(w.enterHint.QWidget)
	w.placeList = qt.NewQListWidget(tab)
	w.placeList.OnCurrentRowChanged(func(row int) {
		if ps := w.places(); row >= 0 && row < len(ps) {
			w.enterMap.Select(ps[row].ID)
		}
		w.removeBtn.SetEnabled(row >= 0)
	})
	w.placeList.OnItemChanged(func(item *qt.QListWidgetItem) {
		if !w.filling {
			w.Rename(w.placeList.Row(item), item.Text())
		}
	})
	side.AddWidget(w.placeList.QWidget)
	w.removeBtn = qt.NewQPushButton3("Remove")
	w.removeBtn.SetEnabled(false)
	w.removeBtn.OnClicked(func() { w.Remove(w.placeList.CurrentRow()) })
	side.AddWidget(w.removeBtn.QWidget)
	sideW := qt.NewQWidget(tab)
	sideW.SetLayout(side.QLayout)
	sideW.SetMaximumWidth(280)
	body.AddWidget2(sideW, 1)
	layout.AddLayout2(body.QLayout, 1)
	return tab
}

// loadMap shows the lesson's map and finds whether it is a bundled one.
func (w *TopoLessonWidget) loadMap() {
	img, _ := w.lesson.Data.Resources[lesson.MapImageResource].([]byte)
	w.known = nil
	w.mapCombo.SetCurrentIndex(0)
	if len(img) == 0 {
		w.enterMap.SetImage(nil)
		w.teachMap.SetImage(nil)
		return
	}
	for i, m := range w.maps {
		if b, err := os.ReadFile(m.Image); err == nil && bytes.Equal(b, img) {
			w.known = &w.maps[i]
			w.mapCombo.SetCurrentIndex(i + 1)
		}
	}
	q := qt.QImage_FromDataWithData(img)
	w.enterMap.SetImage(q)
	w.teachMap.SetImage(q)
}

func (w *TopoLessonWidget) useBundledMap(m Map) {
	if err := w.UseMapPicture(m.Image); err != nil {
		qt.QMessageBox_Warning(w.QWidget, "Choose a map", err.Error())
	}
}

// UseMapPicture makes the picture at path the lesson's map.
func (w *TopoLessonWidget) UseMapPicture(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if qt.QImage_FromDataWithData(b).IsNull() {
		return fmt.Errorf("%s is not a picture Recuerdo can show", filepath.Base(path))
	}
	w.lesson.Data.Resources[lesson.MapImageResource] = b
	w.loadMap()
	w.modified()
	w.refresh()
	return nil
}

// AddByName adds a place: where the map knows it, or where the map is
// clicked next.
func (w *TopoLessonWidget) AddByName(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	if w.known != nil {
		if p, ok := w.known.Find(name); ok {
			w.add(p.Name, p.X, p.Y)
			w.nameEdit.Clear()
			return
		}
	}
	w.enterHint.SetText(fmt.Sprintf("Now click where %s is on the map.", name))
}

// ClickEnter handles a click on the map at picture coordinates: it selects
// a place there, or adds one (named after the name typed, if any).
func (w *TopoLessonWidget) ClickEnter(x, y int) {
	if p, ok := w.enterMap.Near(x, y); ok {
		for i, q := range w.places() {
			if q.ID == p.ID {
				w.placeList.SetCurrentRow(i)
			}
		}
		return
	}
	name := strings.TrimSpace(w.nameEdit.Text())
	if name == "" {
		name = fmt.Sprintf("Place %d", len(w.places())+1)
	}
	w.add(name, x, y)
	w.nameEdit.Clear()
	if strings.HasPrefix(name, "Place ") {
		w.placeList.EditItem(w.placeList.Item(w.placeList.Count() - 1))
	}
}

func (w *TopoLessonWidget) add(name string, x, y int) {
	items := &w.lesson.Data.List.Items
	*items = append(*items, Item(Place{ID: NextID(*items), Name: name, X: x, Y: y}))
	w.enterHint.SetText(defaultHint)
	w.modified()
	w.refresh()
	w.placeList.SetCurrentRow(w.placeList.Count() - 1)
}

// AddAllKnown adds the places the bundled map knows that the lesson lacks.
func (w *TopoLessonWidget) AddAllKnown() {
	if w.known == nil {
		return
	}
	have := map[string]bool{}
	for _, p := range w.places() {
		have[strings.ToLower(p.Name)] = true
	}
	items := &w.lesson.Data.List.Items
	added := 0
	for _, p := range w.known.Places {
		if !have[strings.ToLower(p.Name)] {
			*items = append(*items, Item(Place{ID: NextID(*items), Name: p.Name, X: p.X, Y: p.Y}))
			added++
		}
	}
	if added > 0 {
		w.modified()
		w.refresh()
	}
}

// Rename renames the place in row of the list.
func (w *TopoLessonWidget) Rename(row int, name string) {
	name = strings.TrimSpace(name)
	ps := w.places()
	if row < 0 || row >= len(ps) || name == "" || name == ps[row].Name {
		return
	}
	for i, it := range w.lesson.Data.List.Items {
		if it.ID == ps[row].ID {
			p := ps[row]
			p.Name = name
			w.lesson.Data.List.Items[i] = Item(p)
		}
	}
	w.modified()
	w.refresh()
}

// Remove removes the place in row of the list.
func (w *TopoLessonWidget) Remove(row int) {
	ps := w.places()
	if row < 0 || row >= len(ps) {
		return
	}
	items := w.lesson.Data.List.Items[:0]
	for _, it := range w.lesson.Data.List.Items {
		if it.ID != ps[row].ID {
			items = append(items, it)
		}
	}
	w.lesson.Data.List.Items = items
	w.modified()
	w.refresh()
}

// ---- Teach ----

func (w *TopoLessonWidget) teachTab() *qt.QWidget {
	tab := qt.NewQWidget(nil)
	layout := qt.NewQVBoxLayout(tab)
	options := qt.NewQHBoxLayout2()
	combo := func(label string, items []string) *qt.QComboBox {
		options.AddWidget(qt.NewQLabel3(label).QWidget)
		c := qt.NewQComboBox(tab)
		c.AddItems(items)
		options.AddWidget(c.QWidget)
		return c
	}
	w.orderCombo = combo("Ask:", []string{PlaceName, NamePlace})
	w.orderCombo.SetToolTip("Place – Name: a place is marked, type its name.\nName – Place: a name is given, click the place.")
	w.typeCombo = combo("Lesson type:", teaching.LessonTypes)
	w.sequenceCombo = combo("Order:", teaching.Orders)
	options.AddStretch()
	w.startBtn = qt.NewQPushButton3("Start")
	w.startBtn.OnClicked(func() {
		if w.session != nil {
			w.Stop()
		} else {
			w.Start(w.orderCombo.CurrentText())
		}
	})
	options.AddWidget(w.startBtn.QWidget)
	layout.AddLayout(options.QLayout)

	w.prompt = qt.NewQLabel3("")
	font := w.prompt.Font()
	font.SetPointSizeF(font.PointSizeF() * 1.6)
	w.prompt.SetFont(font)
	w.prompt.SetAlignment(qt.AlignCenter)
	layout.AddWidget(w.prompt.QWidget)

	w.answerRow = qt.NewQWidget(tab)
	row := qt.NewQHBoxLayout(w.answerRow)
	row.SetContentsMargins(0, 0, 0, 0)
	w.answerEdit = qt.NewQLineEdit(w.answerRow)
	w.answerEdit.SetPlaceholderText("Name of the marked place")
	w.answerEdit.OnReturnPressed(func() { w.AnswerName(w.answerEdit.Text()) })
	row.AddWidget(w.answerEdit.QWidget)
	check := qt.NewQPushButton3("Check")
	check.OnClicked(func() { w.AnswerName(w.answerEdit.Text()) })
	row.AddWidget(check.QWidget)
	layout.AddWidget(w.answerRow)

	w.teachMap = NewMapView(tab)
	w.teachMap.OnClick(func(x, y int) { w.AnswerClick(x, y) })
	layout.AddWidget2(w.teachMap.QWidget, 1)

	status := qt.NewQHBoxLayout2()
	w.feedback = qt.NewQLabel3("")
	status.AddWidget(w.feedback.QWidget)
	status.AddStretch()
	w.progress = qt.NewQLabel3("")
	status.AddWidget(w.progress.QWidget)
	layout.AddLayout(status.QLayout)
	return tab
}

// Start starts practising in order (PlaceName or NamePlace).
func (w *TopoLessonWidget) Start(order string) {
	if len(w.places()) == 0 {
		w.prompt.SetText("Enter some places first.")
		return
	}
	w.order = order
	w.orderCombo.SetCurrentText(order)
	w.session = teaching.New(w.lesson.Data.List, teaching.Options{
		LessonType: w.typeCombo.CurrentText(), Order: w.sequenceCombo.CurrentText(),
	})
	w.session.Start()
	w.startBtn.SetText("Stop")
	w.orderCombo.SetEnabled(false)
	w.feedback.SetText("")
	w.ask()
}

// Stop ends the practice; answered questions are kept as a test.
func (w *TopoLessonWidget) Stop() {
	if w.session == nil {
		return
	}
	if t := w.session.LessonTest(); len(t.Results) > 0 {
		w.lesson.Data.List.Tests = append(w.lesson.Data.List.Tests, t)
		w.modified()
	}
	right, answered := w.session.Score()
	w.session = nil
	w.startBtn.SetText("Start")
	w.orderCombo.SetEnabled(true)
	w.answerRow.SetVisible(false)
	w.teachMap.Select(-1)
	w.prompt.SetText(fmt.Sprintf("Done: %d of %d right.", right, answered))
	w.progress.SetText("")
	w.refresh()
}

func (w *TopoLessonWidget) ask() {
	item, _, ok := w.session.Current()
	if !ok {
		w.Stop()
		return
	}
	asked, total := w.session.Progress()
	w.progress.SetText(fmt.Sprintf("%d of %d", asked, total))
	w.teachMap.SetPlaces(w.places(), false)
	if w.order == NamePlace {
		w.prompt.SetText("Where is " + item.Name + "?")
		w.answerRow.SetVisible(false)
		w.teachMap.Select(-1)
		return
	}
	w.prompt.SetText("Which place is marked?")
	w.answerRow.SetVisible(true)
	w.answerEdit.Clear()
	w.answerEdit.SetFocus()
	w.teachMap.Select(item.ID)
}

// AnswerName answers the marked place's name (Place – Name).
func (w *TopoLessonWidget) AnswerName(text string) {
	if w.session == nil || w.order != PlaceName || strings.TrimSpace(text) == "" {
		return
	}
	a := w.session.Answer(text)
	w.answered(a.Right, a.Correct)
}

// AnswerClick answers by clicking the map at picture coordinates
// (Name – Place); the place nearest the click is the answer.
func (w *TopoLessonWidget) AnswerClick(x, y int) {
	if w.session == nil || w.order != NamePlace {
		return
	}
	item, _, ok := w.session.Current()
	if !ok {
		return
	}
	p, near := Nearest(w.places(), x, y, 1e9)
	right := near && p.ID == item.ID
	w.session.Record(right, p.Name)
	w.answered(right, item.Name)
}

func (w *TopoLessonWidget) answered(right bool, correct string) {
	if right {
		w.feedback.SetText("✔ Right: " + correct)
	} else {
		w.feedback.SetText("✘ That was " + correct)
	}
	w.session.Next()
	w.ask()
}

// ---- Results ----

func (w *TopoLessonWidget) resultsTab() *qt.QWidget {
	tab := qt.NewQWidget(nil)
	layout := qt.NewQVBoxLayout(tab)
	w.summary = qt.NewQLabel3("")
	w.summary.SetWordWrap(true)
	layout.AddWidget(w.summary.QWidget)
	layout.AddWidget(qt.NewQLabel3("Last session:").QWidget)
	w.timeline = charts.NewTimelineChart(tab)
	layout.AddWidget(w.timeline.QWidget)
	layout.AddWidget(qt.NewQLabel3("Each session:").QWidget)
	w.grades = charts.NewGradesChart(tab)
	layout.AddWidget(w.grades.QWidget)
	layout.AddStretch()
	return tab
}

// refresh shows the lesson's current places and results.
func (w *TopoLessonWidget) refresh() {
	places := w.places()
	w.filling = true
	row := w.placeList.CurrentRow()
	w.placeList.Clear()
	for _, p := range places {
		item := qt.NewQListWidgetItem7(p.Name, w.placeList)
		item.SetFlags(item.Flags() | qt.ItemIsEditable)
	}
	if row >= 0 && row < len(places) {
		w.placeList.SetCurrentRow(row)
	}
	w.filling = false
	w.enterMap.SetPlaces(places, true)
	w.addAllBtn.SetEnabled(w.known != nil)
	if w.session == nil {
		w.teachMap.SetPlaces(places, true)
		w.answerRow.SetVisible(false)
		if w.prompt.Text() == "" || len(places) == 0 {
			w.prompt.SetText(fmt.Sprintf("%d places. Choose how to practise and press Start.", len(places)))
		}
	}

	tests := w.lesson.Data.List.Tests
	pct := charts.Percentages(tests)
	switch {
	case len(pct) == 0:
		w.summary.SetText("No results yet: practise on the Teach tab.")
	default:
		w.summary.SetText(fmt.Sprintf("%d sessions; the last one %d%% right.", len(pct), pct[len(pct)-1]))
	}
	if len(tests) > 0 {
		w.timeline.SetTest(tests[len(tests)-1])
	}
	w.grades.SetTests(tests)
}
