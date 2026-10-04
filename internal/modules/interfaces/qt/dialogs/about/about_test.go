package about

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/version"
	"github.com/mappu/miqt/qt"
)

var aboutErr error

// Qt on the main thread: the dialog is built in TestMain.
func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"about-test"})
	aboutErr = checkAbout()
	os.Exit(m.Run())
}

// texts are the texts of the labels and buttons under o.
func texts(o *qt.QObject) []string {
	var out []string
	for _, c := range o.Children() {
		switch c.MetaObject().ClassName() {
		case "QLabel":
			out = append(out, qt.UnsafeNewQLabel(c.UnsafePointer()).Text())
		case "QPushButton":
			out = append(out, qt.UnsafeNewQPushButton(c.UnsafePointer()).Text())
		}
		out = append(out, texts(c)...)
	}
	return out
}

func checkAbout() error {
	version.Version = "9.9.9-test"
	mod := NewAboutDialogModule()
	mod.createDialog(nil)
	all := strings.Join(texts(mod.dialog.QObject), "\n")
	for _, want := range []string{"Recuerdo", "Version 9.9.9-test", "GNU General Public License", "Credits"} {
		if !strings.Contains(all, want) {
			return fmt.Errorf("the About dialog lacks %q:\n%s", want, all)
		}
	}
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		mod.dialog.Show()
		qt.QCoreApplication_ProcessEvents()
		mod.dialog.Grab().Save(dir + "/about.png")
	}
	return nil
}

func TestAboutDialog(t *testing.T) {
	if aboutErr != nil {
		t.Fatal(aboutErr)
	}
}

func TestCreditsText(t *testing.T) {
	c := CreditsText()
	if !strings.HasPrefix(c, "Recuerdo: Joop Kiefte") || !strings.Contains(c, "Based on OpenTeacher by:") ||
		!strings.Contains(c, "Marten de Vries") || !strings.Contains(c, "Milan Boers") {
		t.Errorf("credits:\n%s", c)
	}
}
