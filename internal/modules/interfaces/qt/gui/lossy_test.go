package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/logging"
)

var lossy struct {
	asked     []string
	unchanged bool
}

// driveLossy saves (Ctrl+S) a .csv lesson with test results and
// declines the question: the file stays as it was.
func driveLossy() {
	dir, _ := os.MkdirTemp("", "lossy")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "words.csv")
	os.WriteFile(path, []byte("hond,dog\n"), 0o644)
	mod := &GuiModule{logger: logging.NewLogger("gui-test"), mainWindow: qt.NewQMainWindow(nil)}
	mod.tabWidget = qt.NewQTabWidget(nil)
	tab := qt.NewQWidget(nil)
	mod.tabWidget.AddTab(tab, "words")
	l := &lesson.Lesson{Path: path, Data: *lesson.NewLessonData()}
	l.Data.List.Items = []lesson.WordItem{{Questions: []string{"hond"}, Answers: []string{"dog"}, Comment: "een dier"}}
	l.Data.List.Tests = []lesson.Test{{}, {}}
	mod.rememberLesson(tab, l)
	mod.askLossy = func(p string, losses []string) bool {
		lossy.asked = losses
		return false
	}
	mod.saveCurrentLesson(false)
	b, _ := os.ReadFile(path)
	lossy.unchanged = string(b) == "hond,dog\n"
}

func TestSaveAsksBeforeLosing(t *testing.T) {
	if got := strings.Join(lossy.asked, "; "); got != "the results of 2 tests" {
		t.Errorf("asked about %q", got)
	}
	if !lossy.unchanged {
		t.Error("declining still saved")
	}
}

// fakePractice is a lesson widget with a practice running.
type fakePractice struct{ end func() }

func (p fakePractice) KeepPractice() { p.end() }

var kept struct {
	ended, asked, stillOpen bool
}

// driveKeepPractice closes a tab whose practice is running: the practice
// ends (its answers kept, the lesson changed), so closing asks first.
func driveKeepPractice() {
	mod := &GuiModule{logger: logging.NewLogger("gui-test"), mainWindow: qt.NewQMainWindow(nil)}
	mod.tabWidget = qt.NewQTabWidget(nil)
	tab := qt.NewQWidget(nil)
	mod.tabWidget.AddTab(tab, "Dieren")
	mod.rememberPractice(tab, fakePractice{end: func() {
		kept.ended = true
		mod.markModified(tab)
	}})
	mod.askSave = func(string) qt.QMessageBox__StandardButton {
		kept.asked = true
		return qt.QMessageBox__Cancel
	}
	mod.closeTab(0)
	kept.stillOpen = mod.tabWidget.Count() == 1
}

func TestClosingKeepsThePractice(t *testing.T) {
	if !kept.ended || !kept.asked || !kept.stillOpen {
		t.Errorf("practice ended %v, asked %v, still open after Cancel %v", kept.ended, kept.asked, kept.stillOpen)
	}
}
