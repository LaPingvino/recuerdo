package media

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
	checks   = map[string]string{}
	modified int
)

func check(name string, ok bool, format string, args ...any) {
	if !ok {
		checks[name] = fmt.Sprintf(format, args...)
	} else if _, seen := checks[name]; !seen {
		checks[name] = ""
	}
}

func shot(w *MediaLessonWidget, tab int, name string) {
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		w.tabs.SetCurrentIndex(tab)
		qt.QCoreApplication_ProcessEvents()
		w.Grab().Save(filepath.Join(dir, name))
	}
}

func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"media-test"})
	drive()
	os.Exit(m.Run())
}

func drive() {
	root := filepath.Join("..", "..", "..", "..", "..", "..")
	data, err := lesson.NewFileLoader().LoadFile(filepath.Join(root, "testdata", "legacy_files",
		"application_x-openteachingmedia.openteacher3x.otmd"))
	if err != nil {
		check("load", false, "%v", err)
		return
	}
	l := lesson.NewLesson("media")
	l.Data = *data
	w := NewMediaLessonWidget(l, nil)
	w.SetOnModified(func() { modified++ })
	w.Resize(1000, 640)
	w.Show()
	check("table", w.table.RowCount() == 2 && w.table.Item(0, colFile).Text() == "website: http://openteacher.org/" &&
		w.table.Item(1, colFile).Text() == "picture: openteacher-icon.png", "table rows %d, %q",
		w.table.RowCount(), w.table.Item(0, colFile).Text())

	// add a picture and a text file from disk, and a video address
	dir, _ := os.MkdirTemp("", "media-test")
	defer os.RemoveAll(dir)
	pic := filepath.Join(dir, "dog.png")
	icon := lesson.MediaFiles(&l.Data)["resources/openteacher-icon.png"]
	os.WriteFile(pic, icon, 0o644)
	txt := filepath.Join(dir, "poem.txt")
	os.WriteFile(txt, []byte("Roses are red"), 0o644)
	check("add file", w.AddFile(pic) == nil && w.AddFile(txt) == nil, "adding files failed")
	check("bad address", w.AddAddress("openteacher") != nil, "accepted a non-address")
	check("add address", w.AddAddress("https://www.youtube.com/watch?v=abc") == nil, "address refused")
	items := l.Data.List.Items
	check("added", len(items) == 5 && items[2].Name == "dog" && *items[2].Filename == "resources/dog.png" &&
		items[2].ID == 2 && *items[4].Remote && lesson.MediaFiles(&l.Data)["resources/poem.txt"] != nil,
		"items %+v", items)

	// answers, as typed in the table
	w.table.Item(1, colAnswer).SetText("logo")
	w.table.Item(2, colQuestion).SetText("Which animal?")
	w.table.Item(2, colAnswer).SetText("dog")
	w.Edit(3, colAnswer, "roses")
	check("edit", len(items) == 5 && l.Data.List.Items[1].Answers[0] == "logo" &&
		l.Data.List.Items[2].Questions[0] == "Which animal?", "edits not kept: %+v", l.Data.List.Items)
	w.Remove(4)
	check("remove", len(l.Data.List.Items) == 4, "%d items", len(l.Data.List.Items))

	// previews: a picture in place, a website with an Open button
	w.table.SetCurrentCell(2, colName)
	check("picture preview", w.enterPreview.picture.IsVisibleTo(w.QWidget) && !w.enterPreview.open.IsVisibleTo(w.QWidget),
		"picture not shown")
	shot(w, 0, "01-enter.png")
	opened := ""
	w.enterPreview.Opened = func(t string) { opened = t }
	w.table.SetCurrentCell(0, colName)
	w.enterPreview.Open()
	check("open website", opened == "http://openteacher.org/" && w.enterPreview.open.IsVisibleTo(w.QWidget),
		"opened %q", opened)

	// practise the items with an answer (the sample website has "b")
	w.Start()
	asked := 0
	for w.session != nil && asked < 10 {
		item, _, _ := w.session.Current()
		if asked == 1 {
			shot(w, 1, "02-teach.png")
		}
		answer := item.Answers[0]
		if asked == 0 {
			answer = "wrong"
		}
		w.Answer(answer)
		asked++
	}
	tests := l.Data.List.Tests
	check("teach", asked == 4 && len(tests) == 1 && len(tests[0].Results) == 4 && tests[0].Results[0].Result == "wrong" &&
		tests[0].Results[1].Result == "right", "asked %d, tests %+v", asked, tests)
	check("modified", modified >= 8 && l.Data.Changed, "modified %d", modified)
	shot(w, 2, "03-results.png")

	// the lesson saves with its files
	out := filepath.Join(dir, "saved.otmd")
	if err := lesson.NewFileSaver().SaveFile(&l.Data, out); err != nil {
		check("save", false, "%v", err)
	} else if again, err := lesson.NewFileLoader().LoadFile(out); err != nil {
		check("save", false, "%v", err)
	} else {
		check("save", len(again.List.Items) == 4 && string(lesson.MediaFiles(again)["resources/poem.txt"]) == "Roses are red",
			"saved lesson %+v", again.List.Items)
	}
}

func TestMediaWidget(t *testing.T) {
	for name, problem := range checks {
		if problem != "" {
			t.Errorf("%s: %s", name, problem)
		}
	}
	if len(checks) < 12 {
		t.Errorf("only %d checks ran", len(checks))
	}
}
