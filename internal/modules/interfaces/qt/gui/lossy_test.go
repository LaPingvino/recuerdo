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
