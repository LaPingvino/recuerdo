package formulapad

import (
	"fmt"
	"os"
	"testing"

	qt "github.com/mappu/miqt/qt6"
)

var padErr error

func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"formulapad-test"})
	padErr = checkPad()
	os.Exit(m.Run())
}

func TestPad(t *testing.T) {
	if padErr != nil {
		t.Fatal(padErr)
	}
}

func checkPad() error {
	win := qt.NewQWidget2()
	defer func() {
		win.Close()
		win.DeleteLater()
		qt.QCoreApplication_SendPostedEvents2(nil, int(qt.QEvent__DeferredDelete))
	}()
	layout := qt.NewQVBoxLayout(win)
	// a formula answer: the field is TeX
	edit := qt.NewQLineEdit(win)
	layout.AddWidget(edit.QWidget)
	pad := New(edit, false, win)
	layout.AddWidget(pad.QWidget)
	edit.SetText("x = ")
	edit.SetCursorPosition(4)
	pad.Press("frac")
	edit.Insert("-b")
	pad.Press("pm")
	pad.Press("sqrt")
	edit.Insert("b^2-4ac")
	if got := edit.Text(); got != `x = \frac{-b\pm \sqrt{b^2-4ac}}{}` {
		return fmt.Errorf("built %q", got)
	}
	edit.SetText("2")
	edit.SelectAll()
	pad.Press("sqrt")
	if got, pos := edit.Text(), edit.CursorPosition(); got != `\sqrt{2}` || pos != 8 {
		return fmt.Errorf("selection: %q at %d", got, pos)
	}
	if got := pad.Preview.Text(); got != "√2" {
		return fmt.Errorf("preview %q", got)
	}
	// words with formulas: a button outside one starts one
	words := qt.NewQLineEdit(win)
	layout.AddWidget(words.QWidget)
	wp := New(words, true, win)
	layout.AddWidget(wp.QWidget)
	words.SetText("area: ")
	words.SetCursorPosition(6)
	wp.Press("pi")
	if got, pos := words.Text(), words.CursorPosition(); got != `area: $\pi $` || pos != 11 {
		return fmt.Errorf("wrap: %q at %d", got, pos)
	}
	wp.Press("sup") // inside the formula now: no new $
	if got := words.Text(); got != `area: $\pi ^{}$` {
		return fmt.Errorf("inside: %q", got)
	}
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		edit.SetText(`x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`)
		words.SetText(`the area: $\pi r^2$`)
		win.Resize(760, 520)
		win.Show()
		qt.QCoreApplication_ProcessEvents()
		win.Grab().Save(dir + "/formulapad.png")
	}
	return nil
}

func TestInsideMath(t *testing.T) {
	for _, c := range []struct {
		s    string
		pos  int
		want bool
	}{
		{"a $x$ b", 3, true}, {"a $x$ b", 6, false}, {`a \$ $x`, 7, true}, {"é $x", 3, true}, {"", 0, false},
	} {
		if got := insideMath(c.s, c.pos); got != c.want {
			t.Errorf("insideMath(%q, %d) = %v", c.s, c.pos, got)
		}
	}
}
