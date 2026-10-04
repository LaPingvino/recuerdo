package lessonDialogs

import (
	"fmt"
	"os"
	"testing"

	qt "github.com/mappu/miqt/qt6"
)

var propertiesErr error

// Qt on the main thread (macOS needs it): the dialog is driven in TestMain.
func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"lessondialogs-test"})
	propertiesErr = checkProperties()
	os.Exit(m.Run())
}

func checkProperties() error {
	mod := NewLessonDialogsModule()
	mod.createPropertiesDialog(nil)
	mod.loadPropertiesData(map[string]interface{}{
		"name": "Dieren", "questionLanguage": "Dutch", "answerLanguage": "English",
		"hasLanguages": true, "itemCount": 12, "sessionCount": 3,
	})
	if mod.itemCountLabel.Text() != "12" || mod.sessionsLabel.Text() != "3" || mod.propQLangEdit.IsHidden() {
		return fmt.Errorf("shown: %q items, %q sessions, languages hidden %v", mod.itemCountLabel.Text(),
			mod.sessionsLabel.Text(), mod.propQLangEdit.IsHidden())
	}
	mod.propNameEdit.SetText("  Huisdieren ")
	got := mod.getPropertiesData()
	if got["name"] != "Huisdieren" || got["questionLanguage"] != "Dutch" || got["answerLanguage"] != "English" {
		return fmt.Errorf("read back: %v", got)
	}
	// a topography lesson has no languages
	mod.loadPropertiesData(map[string]interface{}{"name": "Europe", "hasLanguages": false, "itemCount": 3})
	if !mod.propQLangEdit.IsHidden() || !mod.propALangEdit.IsHidden() {
		return fmt.Errorf("language fields shown for a lesson without languages")
	}
	return nil
}

func TestPropertiesDialog(t *testing.T) {
	if propertiesErr != nil {
		t.Fatal(propertiesErr)
	}
}
