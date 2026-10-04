package gui

import (
	"github.com/LaPingvino/recuerdo/internal/lesson"
	"os"
	"path/filepath"
	"strings"
	"time"

	qt "github.com/mappu/miqt/qt6"
)

// ScreenshotsEnv names a directory; when set, Recuerdo saves its main
// window at each step of a short tour to PNG files there, then quits.
// For design reviews, e.g. with QT_QPA_PLATFORM=offscreen:
//
//	RECUERDO_SCREENSHOTS=/tmp/shots QT_QPA_PLATFORM=offscreen recuerdo
const ScreenshotsEnv = "RECUERDO_SCREENSHOTS"

// sampleLesson is the word list the tour opens.
const sampleLesson = `huis,house
boom,tree
water,water
fiets,bicycle
kaas,cheese
vrienden,friends
morgen,tomorrow
gezellig,cosy
`

// startScreenshots runs the tour if ScreenshotsEnv is set.
func (mod *GuiModule) startScreenshots() {
	dir := os.Getenv(ScreenshotsEnv)
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		mod.logger.Error("screenshots: %v", err)
		return
	}
	mod.mainWindow.Resize(1000, 700)

	lessonFile := filepath.Join(dir, "sample-words.otwd")
	shoot := func(name string) func() {
		return func() {
			path := filepath.Join(dir, name+".png")
			if !mod.mainWindow.QWidget.Grab().Save(path) {
				mod.logger.Error("screenshots: could not save %s", path)
			}
		}
	}
	lessonTab := func(i int) func() {
		return func() {
			if mod.lastWords != nil {
				mod.lastWords.SetCurrentTab(i)
			}
		}
	}
	steps := []func(){
		shoot("01-start"),
		func() {
			if err := writeSample(lessonFile); err == nil {
				mod.loadSelectedFile(lessonFile)
			} else {
				mod.logger.Error("screenshots: sample lesson: %v", err)
			}
		},
		shoot("02-enter"),
		func() {
			// an edit, then saving as OpenTeaching Words and as PDF
			if mod.lastWords != nil {
				mod.lastWords.SetTitle("Sample words")
			}
			for _, ext := range []string{".otwd", ".pdf"} {
				if err := mod.SaveCurrentLessonTo(filepath.Join(dir, "saved"+ext)); err != nil {
					mod.logger.Error("screenshots: saving %s: %v", ext, err)
				}
			}
		},
		lessonTab(1),
		shoot("03-teach"),
		func() {
			if mod.lastWords != nil {
				mod.lastWords.StartTeaching()
			}
		},
		shoot("04-teaching"),
		lessonTab(2),
		shoot("05-results"),
		func() {
			if d, err := mod.gettingStartedDialog(); err == nil {
				d.Show()
				d.Grab().Save(filepath.Join(dir, "07-getting-started.png"))
				d.Close()
			}
		},
		func() {
			// the about dialog is modal: a separate timer, running in its
			// event loop, saves and closes it
			closer := qt.NewQTimer()
			closer.SetSingleShot(true)
			closer.OnTimeout(func() {
				if w := qt.QApplication_ActiveModalWidget(); w != nil {
					w.Grab().Save(filepath.Join(dir, "06-about.png"))
					w.Close()
				}
			})
			closer.Start(400)
			mod.showAboutDialog()
		},
		func() {
			os.Remove(lessonFile)
			qt.QCoreApplication_Quit()
		},
	}

	timer := qt.NewQTimer()
	next := 0
	timer.OnTimeout(func() {
		if next >= len(steps) {
			timer.Stop()
			return
		}
		// advance first: a step may open a modal dialog, whose event loop
		// keeps this timer running to the step that closes it
		step := steps[next]
		next++
		step()
	})
	timer.Start(400)
}

// writeSample saves the tour's lesson: sampleLesson's words with three
// earlier sessions (the last with answer times), for the results charts.
func writeSample(path string) error {
	items, err := lesson.ParseWordList(strings.ReplaceAll(sampleLesson, ",", " = "), true)
	if err != nil {
		return err
	}
	data := lesson.NewLessonData()
	data.List.Title = "sample-words"
	data.List.Items = items
	start := time.Date(2026, 10, 1, 19, 0, 0, 0, time.UTC)
	for s, rightUpTo := range []int{3, 5, 7} {
		t := lesson.Test{}
		day := start.AddDate(0, 0, s)
		t.Date = &day
		at := day
		for i := range items {
			at = at.Add(time.Duration(2+i%3*3) * time.Second)
			when := at
			r := lesson.TestResult{ItemID: items[i].ID, Result: "wrong", Time: &when}
			if i < rightUpTo {
				r.Result = "right"
			}
			t.Results = append(t.Results, r)
		}
		data.List.Tests = append(data.List.Tests, t)
	}
	return lesson.NewFileSaver().SaveFile(data, path)
}
