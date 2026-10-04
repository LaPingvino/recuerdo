package results

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LaPingvino/recuerdo/internal/teaching"
	qt "github.com/mappu/miqt/qt6"
)

func init() { runtime.LockOSThread() }

// Qt widgets need a QApplication on the main thread, which is where
// TestMain runs; the check is done here rather than in a Test function.
func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"results-test"})

	dialogErr = checkDialog()
	os.Exit(m.Run())
}

var dialogErr error

// TestResultsDialog reports the check TestMain ran on the main thread.
func TestResultsDialog(t *testing.T) {
	if dialogErr != nil {
		t.Fatal(dialogErr)
	}
}

func checkDialog() error {
	report := teaching.Report{
		Rows: []teaching.Row{
			{Question: "een", Answer: "one", Given: "uno", Right: false},
			{Question: "een", Answer: "one", Given: "one", Right: true},
		},
		MostDoneWrong: "een <b>",
		Finished:      true,
		ThinkingTime:  12 * time.Second,
	}
	d := Show(nil, report, "5,5", "Dutch")
	// close it before the program ends, and let Qt delete it (it deletes
	// itself on close): Qt 6 on Windows crashed tearing down a dialog
	// still shown at exit
	defer func() {
		d.Close()
		qt.QCoreApplication_SendPostedEvents2(nil, int(qt.QEvent__DeferredDelete))
	}()
	if d.WindowTitle() != "Results" || !d.IsVisible() {
		return fmt.Errorf("title %q, visible %v", d.WindowTitle(), d.IsVisible())
	}

	var texts []string
	var tables []*qt.QTableWidget
	var walk func(o *qt.QObject)
	walk = func(o *qt.QObject) {
		for _, c := range o.Children() {
			switch {
			case c.Inherits("QLabel"):
				texts = append(texts, qt.UnsafeNewQLabel(c.UnsafePointer()).Text())
			case c.Inherits("QTableWidget"):
				tables = append(tables, qt.UnsafeNewQTableWidget(c.UnsafePointer()))
			}
			walk(c)
		}
	}
	walk(d.QObject)
	all := strings.Join(texts, "\n")
	for _, want := range []string{"Total thinking time: 12 seconds", "Note (Dutch):", "5,5", "een &lt;b&gt;", "Completed: yes"} {
		if !strings.Contains(all, want) {
			return fmt.Errorf("missing %q in labels:\n%s", want, all)
		}
	}

	if len(tables) != 1 {
		return fmt.Errorf("%d tables", len(tables))
	}
	table := tables[0]
	if table.RowCount() != 2 || table.Item(0, 2).Text() != "uno" ||
		table.Item(0, 3).CheckState() != qt.Unchecked || table.Item(1, 3).CheckState() != qt.Checked {
		return fmt.Errorf("table rows %d", table.RowCount())
	}
	d.Close()
	return nil
}
