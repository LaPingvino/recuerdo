package media

import (
	"fmt"
	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/valuecombo"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/charts"
	"github.com/LaPingvino/recuerdo/internal/teaching"
	qt "github.com/mappu/miqt/qt6"
)

// Columns of the Enter tab's table.
const (
	colName = iota
	colQuestion
	colAnswer
	colFile
)

// MediaLessonWidget shows a media lesson: pictures, sounds, videos, texts
// and websites with a question and an answer each, to enter, practise and
// see the results of.
type MediaLessonWidget struct {
	*qt.QWidget
	lesson     *lesson.Lesson
	onModified func()
	tabs       *qt.QTabWidget

	// Enter
	table        *qt.QTableWidget
	enterPreview *Preview
	removeBtn    *qt.QPushButton
	filling      bool

	// Teach
	typeCombo, orderCombo      *qt.QComboBox
	startBtn                   *qt.QPushButton
	teachPreview               *Preview
	question, feedback, status *qt.QLabel
	answerRow                  *qt.QWidget
	answerEdit                 *qt.QLineEdit
	session                    *teaching.Session

	// Results
	summary  *qt.QLabel
	grades   *charts.GradesChart
	timeline *charts.TimelineChart
}

// NewMediaLessonWidget creates the widget for a media lesson.
func NewMediaLessonWidget(l *lesson.Lesson, parent *qt.QWidget) *MediaLessonWidget {
	w := &MediaLessonWidget{QWidget: qt.NewQWidget(parent), lesson: l}
	layout := qt.NewQVBoxLayout(w.QWidget)
	layout.SetContentsMargins(0, 0, 0, 0)
	w.tabs = qt.NewQTabWidget(w.QWidget)
	layout.AddWidget(w.tabs.QWidget)
	w.tabs.AddTab(w.enterTab(), i18n.T("Enter"))
	w.tabs.AddTab(w.teachTab(), i18n.T("Teach"))
	w.tabs.AddTab(w.resultsTab(), i18n.T("Results"))
	w.tabs.OnCurrentChanged(func(int) { w.refresh() })
	w.refresh()
	return w
}

// SetOnModified calls f after every change to the lesson.
func (w *MediaLessonWidget) SetOnModified(f func()) { w.onModified = f }

func (w *MediaLessonWidget) modified() {
	w.lesson.Data.Changed = true
	if w.onModified != nil {
		w.onModified()
	}
}

func (w *MediaLessonWidget) files() map[string][]byte { return lesson.MediaFiles(&w.lesson.Data) }

// ---- Enter ----

func (w *MediaLessonWidget) enterTab() *qt.QWidget {
	tab := qt.NewQWidget(nil)
	layout := qt.NewQVBoxLayout(tab)
	buttons := qt.NewQHBoxLayout2()
	addFiles := qt.NewQPushButton3(i18n.T("Add files…"))
	addFiles.SetToolTip(i18n.T("Pictures, sounds, videos or texts, kept inside the lesson"))
	addFiles.OnClicked(func() {
		paths := qt.QFileDialog_GetOpenFileNames4(w.QWidget, "Add files", "",
			"Media (*.png *.jpg *.jpeg *.gif *.bmp *.webp *.tif *.tiff *.txt *.mp3 *.wav *.ogg *.flac *.m4a *.mid *.mp4 *.webm *.mkv *.avi *.mov *.mpg);;All files (*)")
		for _, p := range paths {
			if err := w.AddFile(p); err != nil {
				qt.QMessageBox_Warning(w.QWidget, "Add files", err.Error())
			}
		}
	})
	buttons.AddWidget(addFiles.QWidget)
	addURL := qt.NewQPushButton3(i18n.T("Add web address…"))
	addURL.SetToolTip(i18n.T("A website or a YouTube, Vimeo or Dailymotion video"))
	addURL.OnClicked(func() {
		if u := qt.QInputDialog_GetText(w.QWidget, "Add web address", "Web address (https://…):"); u != "" {
			if err := w.AddAddress(u); err != nil {
				qt.QMessageBox_Warning(w.QWidget, "Add web address", err.Error())
			}
		}
	})
	buttons.AddWidget(addURL.QWidget)
	w.removeBtn = qt.NewQPushButton3(i18n.T("Remove"))
	w.removeBtn.SetEnabled(false)
	w.removeBtn.OnClicked(func() { w.Remove(w.table.CurrentRow()) })
	buttons.AddWidget(w.removeBtn.QWidget)
	buttons.AddStretch()
	layout.AddLayout(buttons.QLayout)

	body := qt.NewQHBoxLayout2()
	w.table = qt.NewQTableWidget(tab)
	w.table.SetColumnCount(4)
	w.table.SetHorizontalHeaderLabels([]string{i18n.T("Name"), i18n.T("Question"), i18n.T("Answer"), i18n.T("File")})
	w.table.HorizontalHeader().SetSectionResizeMode(qt.QHeaderView__Stretch)
	w.table.VerticalHeader().SetVisible(false)
	w.table.SetSelectionBehavior(qt.QAbstractItemView__SelectRows)
	w.table.OnCellChanged(func(row, col int) {
		if !w.filling {
			w.Edit(row, col, w.table.Item(row, col).Text())
		}
	})
	w.table.OnCurrentCellChanged(func(row, _, _, _ int) {
		w.removeBtn.SetEnabled(row >= 0)
		w.showEnterPreview(row)
	})
	body.AddWidget2(w.table.QWidget, 3)
	w.enterPreview = NewPreview(tab)
	body.AddWidget2(w.enterPreview.QWidget, 2)
	layout.AddLayout2(body.QLayout, 1)
	return tab
}

func (w *MediaLessonWidget) showEnterPreview(row int) {
	items := w.lesson.Data.List.Items
	if row < 0 || row >= len(items) {
		w.enterPreview.Show(nil, nil)
		return
	}
	w.enterPreview.Show(&items[row], w.files())
}

// AddFile embeds the file at path in the lesson as a new item named after
// it.
func (w *MediaLessonWidget) AddFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	files := w.files()
	name := ResourceName(path, files)
	files[name] = b
	base := filepath.Base(path)
	w.add(lesson.MediaItem(NextID(w.lesson.Data.List.Items), strings.TrimSuffix(base, filepath.Ext(base)), name, false, "", ""))
	return nil
}

// AddAddress adds a web address (a website or an online video) as an item.
func (w *MediaLessonWidget) AddAddress(address string) error {
	address = strings.TrimSpace(address)
	if !AddressOK(address) {
		return fmt.Errorf("%q is not a web address: it should start with https:// or http://", address)
	}
	w.add(lesson.MediaItem(NextID(w.lesson.Data.List.Items), address, address, true, "", ""))
	return nil
}

func (w *MediaLessonWidget) add(item lesson.WordItem) {
	w.lesson.Data.List.Items = append(w.lesson.Data.List.Items, item)
	w.modified()
	w.refresh()
	row := len(w.lesson.Data.List.Items) - 1
	w.table.SetCurrentCell(row, colAnswer)
}

// Edit changes the name, question or answer of the item in row.
func (w *MediaLessonWidget) Edit(row, col int, text string) {
	items := w.lesson.Data.List.Items
	if row < 0 || row >= len(items) {
		return
	}
	it := &items[row]
	text = strings.TrimSpace(text)
	words := func() []string {
		if text == "" {
			return nil
		}
		return []string{text}
	}
	switch col {
	case colName:
		it.Name = text
	case colQuestion:
		it.Questions = words()
	case colAnswer:
		it.Answers = words()
	default:
		return
	}
	w.modified()
}

// Remove removes the item in row (and its file, if no other item uses it).
func (w *MediaLessonWidget) Remove(row int) {
	items := w.lesson.Data.List.Items
	if row < 0 || row >= len(items) {
		return
	}
	gone := items[row]
	w.lesson.Data.List.Items = append(items[:row:row], items[row+1:]...)
	if gone.Filename != nil {
		used := false
		for _, it := range w.lesson.Data.List.Items {
			used = used || (it.Filename != nil && *it.Filename == *gone.Filename)
		}
		if !used {
			delete(w.files(), *gone.Filename)
		}
	}
	w.modified()
	w.refresh()
}

// ---- Teach ----

func (w *MediaLessonWidget) teachTab() *qt.QWidget {
	tab := qt.NewQWidget(nil)
	layout := qt.NewQVBoxLayout(tab)
	options := qt.NewQHBoxLayout2()
	combo := func(label string, items []string) *qt.QComboBox {
		options.AddWidget(qt.NewQLabel3(i18n.T(label)).QWidget)
		c := qt.NewQComboBox(tab)
		valuecombo.Fill(c, items)
		options.AddWidget(c.QWidget)
		return c
	}
	w.typeCombo = combo("Lesson type:", teaching.LessonTypes)
	w.orderCombo = combo("Order:", teaching.Orders)
	options.AddStretch()
	w.startBtn = qt.NewQPushButton3(i18n.T("Start"))
	w.startBtn.OnClicked(func() {
		if w.session != nil {
			w.Stop()
		} else {
			w.Start()
		}
	})
	options.AddWidget(w.startBtn.QWidget)
	layout.AddLayout(options.QLayout)

	w.teachPreview = NewPreview(tab)
	layout.AddWidget2(w.teachPreview.QWidget, 1)
	w.question = qt.NewQLabel3("")
	font := w.question.Font()
	font.SetPointSizeF(font.PointSizeF() * 1.5)
	w.question.SetFont(font)
	w.question.SetAlignment(qt.AlignCenter)
	w.question.SetWordWrap(true)
	layout.AddWidget(w.question.QWidget)

	w.answerRow = qt.NewQWidget(tab)
	row := qt.NewQHBoxLayout(w.answerRow)
	row.SetContentsMargins(0, 0, 0, 0)
	w.answerEdit = qt.NewQLineEdit(w.answerRow)
	w.answerEdit.SetPlaceholderText(i18n.T("Your answer"))
	w.answerEdit.OnReturnPressed(func() { w.Answer(w.answerEdit.Text()) })
	row.AddWidget(w.answerEdit.QWidget)
	check := qt.NewQPushButton3(i18n.T("Check"))
	check.OnClicked(func() { w.Answer(w.answerEdit.Text()) })
	row.AddWidget(check.QWidget)
	w.answerRow.SetVisible(false)
	layout.AddWidget(w.answerRow)

	status := qt.NewQHBoxLayout2()
	w.feedback = qt.NewQLabel3("")
	status.AddWidget(w.feedback.QWidget)
	status.AddStretch()
	w.status = qt.NewQLabel3("")
	status.AddWidget(w.status.QWidget)
	layout.AddLayout(status.QLayout)
	return tab
}

// Start starts practising the items that have an answer.
func (w *MediaLessonWidget) Start() {
	list := Teachable(w.lesson.Data.List)
	if len(list.Items) == 0 {
		w.question.SetText(i18n.T("Give the items an answer on the Enter tab first."))
		return
	}
	w.session = teaching.New(list, teaching.Options{
		LessonType: valuecombo.Value(w.typeCombo), Order: valuecombo.Value(w.orderCombo),
	})
	w.session.Start()
	w.startBtn.SetText(i18n.T("Stop"))
	w.feedback.SetText("")
	w.answerRow.SetVisible(true)
	w.ask()
}

// Stop ends the practice; answered questions are kept as a test.
func (w *MediaLessonWidget) Stop() {
	if w.session == nil {
		return
	}
	if t := w.session.LessonTest(); len(t.Results) > 0 {
		w.lesson.Data.List.Tests = append(w.lesson.Data.List.Tests, t)
		w.modified()
	}
	right, answered := w.session.Score()
	w.session = nil
	w.startBtn.SetText(i18n.T("Start"))
	w.answerRow.SetVisible(false)
	w.teachPreview.Show(nil, nil)
	w.question.SetText(fmt.Sprintf(i18n.T("Done: %d of %d right."), right, answered))
	w.status.SetText("")
	w.refresh()
}

func (w *MediaLessonWidget) ask() {
	item, _, ok := w.session.Current()
	if !ok {
		w.Stop()
		return
	}
	asked, total := w.session.Progress()
	w.status.SetText(fmt.Sprintf(i18n.T("%d of %d"), asked, total))
	w.teachPreview.Show(&item, w.files())
	q := "What is this?"
	if len(item.Questions) > 0 {
		q = strings.Join(item.Questions, ", ")
	}
	w.question.SetText(q)
	w.answerEdit.Clear()
	w.answerEdit.SetFocus()
}

// Answer checks an answer to the current item.
func (w *MediaLessonWidget) Answer(text string) {
	if w.session == nil || strings.TrimSpace(text) == "" {
		return
	}
	a := w.session.Answer(text)
	if a.Right {
		w.feedback.SetText(i18n.Tf("✔ Right: %s", a.Correct))
	} else {
		w.feedback.SetText(i18n.Tf("✘ The answer was: %s", a.Correct))
	}
	w.session.Next()
	w.ask()
}

// ---- Results ----

func (w *MediaLessonWidget) resultsTab() *qt.QWidget {
	tab := qt.NewQWidget(nil)
	layout := qt.NewQVBoxLayout(tab)
	w.summary = qt.NewQLabel3("")
	w.summary.SetWordWrap(true)
	layout.AddWidget(w.summary.QWidget)
	layout.AddWidget(qt.NewQLabel3(i18n.T("Last session:")).QWidget)
	w.timeline = charts.NewTimelineChart(tab)
	layout.AddWidget(w.timeline.QWidget)
	layout.AddWidget(qt.NewQLabel3(i18n.T("Each session:")).QWidget)
	w.grades = charts.NewGradesChart(tab)
	layout.AddWidget(w.grades.QWidget)
	layout.AddStretch()
	return tab
}

// refresh shows the lesson's current items and results.
func (w *MediaLessonWidget) refresh() {
	items := w.lesson.Data.List.Items
	w.filling = true
	row := w.table.CurrentRow()
	w.table.SetRowCount(len(items))
	for i, it := range items {
		cell := func(col int, text string, editable bool) {
			c := qt.NewQTableWidgetItem2(text)
			if !editable {
				c.SetFlags(c.Flags() &^ qt.ItemIsEditable)
			}
			w.table.SetItem(i, col, c)
		}
		cell(colName, it.Name, true)
		cell(colQuestion, strings.Join(it.Questions, ", "), true)
		cell(colAnswer, strings.Join(it.Answers, ", "), true)
		file, kind := "", Unknown
		if it.Filename != nil {
			file = *it.Filename
			kind = Kind(file, it.Remote != nil && *it.Remote)
			if it.Remote == nil || !*it.Remote {
				file = filepath.Base(file)
			}
		}
		cell(colFile, kind+": "+file, false)
	}
	w.filling = false
	if row >= len(items) {
		row = len(items) - 1
	}
	w.showEnterPreview(row)
	if w.session == nil && w.question.Text() == "" {
		n := len(Teachable(w.lesson.Data.List).Items)
		w.question.SetText(fmt.Sprintf(i18n.T("%d items to practise. Press Start."), n))
	}

	tests := w.lesson.Data.List.Tests
	if pct := charts.Percentages(tests); len(pct) == 0 {
		w.summary.SetText(i18n.T("No results yet: practise on the Teach tab."))
	} else {
		w.summary.SetText(fmt.Sprintf(i18n.T("%d sessions; the last one %d%% right."), len(pct), pct[len(pct)-1]))
	}
	if len(tests) > 0 {
		w.timeline.SetTest(tests[len(tests)-1])
	}
	w.grades.SetTests(tests)
}
