package gui

import (
	"os"
	"path/filepath"

	"github.com/mappu/miqt/qt"
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

	lessonFile := filepath.Join(dir, "sample-words.csv")
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
			if err := os.WriteFile(lessonFile, []byte(sampleLesson), 0o644); err == nil {
				mod.loadSelectedFile(lessonFile)
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
