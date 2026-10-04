package valuecombo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/mappu/miqt/qt"
)

var shown, value, afterSet string

// Qt on the main thread: the combo box is used in TestMain.
func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"valuecombo-test"})
	dir := t()
	os.WriteFile(filepath.Join(dir, "recuerdo-nl.po"), []byte("msgid \"All once\"\nmsgstr \"Alles één keer\"\n"), 0o644)
	i18n.Use(dir, "nl")
	c := qt.NewQComboBox(nil)
	Fill(c, []string{"All once", "Smart"})
	shown, value = c.CurrentText(), Value(c)
	Set(c, "Smart")
	afterSet = Value(c)
	i18n.Use(dir, "")
	os.RemoveAll(dir)
	os.Exit(m.Run())
}

func t() string { d, _ := os.MkdirTemp("", "valuecombo"); return d }

func TestValueCombo(t *testing.T) {
	if shown != "Alles één keer" || value != "All once" || afterSet != "Smart" {
		t.Errorf("shown %q, value %q, after Set %q", shown, value, afterSet)
	}
}
