package gui

import (
	"os"
	"testing"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LaPingvino/recuerdo/internal/logging"
)

// Qt runs on the main thread: TestMain does the Qt work, the tests check
// what it recorded.
var tabs struct {
	closable, movable bool
	afterFirst        int
	welcomeAfterLast  bool
	keptModified      bool
	discarded         bool
}

func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"gui-test"})
	driveTabs()
	os.Exit(m.Run())
}

func driveTabs() {
	mod := &GuiModule{logger: logging.NewLogger("gui-test"), mainWindow: qt.NewQMainWindow(nil)}
	mod.showWelcome()
	open := func(title string) {
		if mod.tabWidget == nil {
			mod.tabWidget = qt.NewQTabWidget(nil)
			mod.tabWidget.SetTabsClosable(true)
			mod.tabWidget.SetMovable(true)
			mod.tabWidget.OnTabCloseRequested(mod.closeTab)
			mod.mainWindow.SetCentralWidget(mod.tabWidget.QWidget)
		}
		mod.tabWidget.AddTab(qt.NewQWidget(nil), title)
	}
	open("Dieren")
	open("Landen")
	tabs.closable, tabs.movable = mod.tabWidget.TabsClosable(), mod.tabWidget.IsMovable()
	mod.closeTab(0)
	tabs.afterFirst = mod.tabWidget.Count()
	mod.closeTab(0)
	tabs.welcomeAfterLast = mod.tabWidget == nil && mod.mainWindow.CentralWidget() != nil
	// a lesson with changes: Cancel keeps it, Discard closes it
	open("*Verbs")
	mod.askSave = func(string) qt.QMessageBox__StandardButton { return qt.QMessageBox__Cancel }
	mod.closeTab(0)
	tabs.keptModified = mod.tabWidget != nil && mod.tabWidget.Count() == 1
	mod.askSave = func(string) qt.QMessageBox__StandardButton { return qt.QMessageBox__Discard }
	mod.closeTab(0)
	tabs.discarded = mod.tabWidget == nil
}

func TestClosableTabs(t *testing.T) {
	if !tabs.closable || !tabs.movable {
		t.Errorf("tabs closable %v, movable %v", tabs.closable, tabs.movable)
	}
	if tabs.afterFirst != 1 {
		t.Errorf("after closing one of two: %d tabs", tabs.afterFirst)
	}
	if !tabs.welcomeAfterLast {
		t.Error("no start screen after the last lesson closed")
	}
	if !tabs.keptModified {
		t.Error("Cancel closed a lesson with changes")
	}
	if !tabs.discarded {
		t.Error("Discard did not close the lesson")
	}
}
