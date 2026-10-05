package typingcourse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LaPingvino/recuerdo/internal/typing"
)

var flowErr error

func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"typing-test"})
	flowErr = flow()
	os.Exit(m.Run())
}

func TestCourse(t *testing.T) {
	if flowErr != nil {
		t.Fatal(flowErr)
	}
}

type failure string

func (f failure) Error() string { return string(f) }

func shot(w *Window, name string) {
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		qt.QCoreApplication_ProcessEvents()
		w.Grab().Save(filepath.Join(dir, name))
	}
}

func flow() error {
	dir, _ := os.MkdirTemp("", "typing")
	defer os.RemoveAll(dir)
	ProfilesPath = func() string { return filepath.Join(dir, "typing.json") }
	w := Show(nil)
	defer func() {
		w.Close()
		w.DeleteLater()
		qt.QCoreApplication_SendPostedEvents2(nil, int(qt.QEvent__DeferredDelete))
	}()
	shot(w, "1-new-profile.png")
	p, err := w.profiles.Add("Anna", "qwerty", "en", w.rng)
	if err != nil {
		return err
	}
	w.profile = p
	w.showInstructions()
	shot(w, "2-instructions.png")
	w.startExercise()
	text := p.Current
	if !strings.ContainsAny(text, "fj") {
		return failure("first exercise " + text)
	}
	// a fake clock: the exercise typed at 30 words per minute
	clock := time.Unix(1000, 0)
	w.now = func() time.Time { return clock }
	w.Type(string(text[0]))
	w.Type("q") // wrong
	if w.mistakes != 1 || w.pos != 1 || !strings.Contains(w.status.Text(), "1") {
		return failure("a wrong key: " + w.status.Text())
	}
	shot(w, "3-exercise.png")
	w.mistakes = 0 // as if typed without the mistake
	clock = clock.Add(time.Duration(float64(len(text))/5/30*60) * time.Second)
	w.Type(text[1:])
	if p.Level != 1 || p.Status != typing.Next {
		return failure("not passed: " + p.Status)
	}
	shot(w, "4-next.png")
	saved, _ := typing.Load(ProfilesPath())
	if saved.Get("Anna") == nil || saved.Get("Anna").Level != 1 {
		return failure("not saved")
	}
	w.showProfiles()
	shot(w, "5-profiles.png")
	return nil
}
