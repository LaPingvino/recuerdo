package plaintextwords

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/mappu/miqt/qt6"
)

var problem error

// Qt on the main thread: the dialog is driven in TestMain.
func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	os.Setenv("RECUERDO_DATA", filepath.Join("..", "..", "..", "..", "..", ".."))
	qt.NewQApplication([]string{"plaintext-test"})
	problem = drive()
	os.Exit(m.Run())
}

func drive() error {
	d := New(nil)
	d.Show()
	// a line without "=" or tab: the dialog says which and stays open
	d.SetText("hond = dog\nkat cat\n")
	d.Check()
	if d.Problem() == "" || d.Lesson() != nil || d.Result() == int(qt.QDialog__Accepted) {
		return fmt.Errorf("bad line accepted: problem %q", d.Problem())
	}
	if want := "line 2: missing equals sign or tab"; d.Problem()[:len(want)] != want {
		return fmt.Errorf("problem %q", d.Problem())
	}
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		qt.QCoreApplication_ProcessEvents()
		d.Grab().Save(filepath.Join(dir, "type-a-list.png"))
	}
	// nothing typed
	d.SetText("  \n")
	d.Check()
	if d.Problem() != "Type at least one word." {
		return fmt.Errorf("empty: %q", d.Problem())
	}
	// a good list: tabs and "=", several answers
	d.SetText("hond = dog\nkat\tcat\nhuis = house, home\n")
	d.Check()
	l := d.Lesson()
	if d.Problem() != "" || l == nil || len(l.Data.List.Items) != 3 || l.Data.List.Items[1].Answers[0] != "cat" ||
		len(l.Data.List.Items[2].Answers) != 2 || !l.Data.Changed || l.DataType != "words" {
		return fmt.Errorf("lesson %+v, problem %q", l, d.Problem())
	}
	return nil
}

func TestTypeAList(t *testing.T) {
	if problem != nil {
		t.Fatal(problem)
	}
}
