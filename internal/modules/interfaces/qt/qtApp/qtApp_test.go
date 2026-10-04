package qtapp

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/version"
	qt "github.com/mappu/miqt/qt6"
)

var appErr error

// The module makes the QApplication, which must happen on the main
// thread: it is enabled in TestMain.
func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	version.Version = "9.9.9-test"
	mod := NewQtAppModule()
	if err := mod.Enable(context.Background()); err != nil {
		appErr = err
	} else {
		appErr = checkApp(mod)
	}
	os.Exit(m.Run())
}

func checkApp(mod *QtAppModule) error {
	if mod.GetApplication() == nil || qt.QCoreApplication_Instance() == nil {
		return fmt.Errorf("no QApplication")
	}
	for got, want := range map[string]string{
		qt.QCoreApplication_ApplicationName():    "Recuerdo",
		qt.QCoreApplication_ApplicationVersion(): "9.9.9-test",
		qt.QGuiApplication_DesktopFileName():     "eu.kiefte.Recuerdo", // Wayland docks match the .desktop file by it
		qt.QCoreApplication_OrganizationDomain(): "kiefte.eu",
	} {
		if got != want {
			return fmt.Errorf("got %q, want %q", got, want)
		}
	}
	if qt.QGuiApplication_WindowIcon().IsNull() {
		return fmt.Errorf("no window icon")
	}
	// enabling again keeps the one application
	app := mod.GetApplication()
	mod.Enable(context.Background())
	if mod.GetApplication() != app {
		return fmt.Errorf("a second QApplication was made")
	}
	return nil
}

func TestQtApp(t *testing.T) {
	if appErr != nil {
		t.Fatal(appErr)
	}
}
