// Package results shows the results of a test: every answer, the grade,
// and a few facts about the test. Port of OpenTeacher's
// interfaces/qt/dialogs/results with the words test viewer it shows.
package results

import (
	"context"
	"fmt"
	"html"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/teaching"
	"github.com/mappu/miqt/qt"
)

// Show opens a (non-modal) results window for a report; grade and
// notation may be empty.
func Show(parent *qt.QWidget, report teaching.Report, grade, notation string) *qt.QDialog {
	dialog := qt.NewQDialog(parent)
	dialog.SetWindowTitle("Results")
	dialog.SetAttribute2(qt.WA_DeleteOnClose, true)
	dialog.Resize(720, 420)

	layout := qt.NewQHBoxLayout(dialog.QWidget)

	// left: thinking time and every answer
	left := qt.NewQVBoxLayout2()
	timeLabel := qt.NewQLabel(dialog.QWidget)
	timeLabel.SetText(report.ThinkingTimeText())
	left.AddWidget(timeLabel.QWidget)

	table := qt.NewQTableWidget2()
	table.SetColumnCount(4)
	table.SetHorizontalHeaderLabels([]string{"Question", "Answer", "Given answer", "Correct"})
	table.SetRowCount(len(report.Rows))
	for i, row := range report.Rows {
		table.SetItem(i, 0, qt.NewQTableWidgetItem2(row.Question))
		table.SetItem(i, 1, qt.NewQTableWidgetItem2(row.Answer))
		table.SetItem(i, 2, qt.NewQTableWidgetItem2(row.Given))
		correct := qt.NewQTableWidgetItem2("")
		if row.Right {
			correct.SetCheckState(qt.Checked)
		} else {
			correct.SetCheckState(qt.Unchecked)
		}
		correct.SetFlags(qt.ItemIsEnabled)
		table.SetItem(i, 3, correct)
	}
	table.SetEditTriggers(qt.QAbstractItemView__NoEditTriggers)
	table.HorizontalHeader().SetStretchLastSection(true)
	table.ResizeColumnsToContents()
	left.AddWidget(table.QWidget)
	layout.AddLayout2(left.QLayout, 3)

	// right: grade and facts
	facts := qt.NewQVBoxLayout2()
	if grade != "" {
		gradeLabel := qt.NewQLabel(dialog.QWidget)
		gradeLabel.SetText(fmt.Sprintf(`Note (%s):<br /><span style="font-size: 40px">%s</span>`, notation, grade))
		facts.AddWidget(gradeLabel.QWidget)
	}
	mostWrong := report.MostDoneWrong
	if mostWrong == "" {
		mostWrong = "-"
	}
	wrongLabel := qt.NewQLabel(dialog.QWidget)
	wrongLabel.SetText(fmt.Sprintf(`Word most done wrong:<br /><span style="font-size: 14px">%s</span>`, html.EscapeString(mostWrong)))
	facts.AddWidget(wrongLabel.QWidget)
	completed := "no"
	if report.Finished {
		completed = "yes"
	}
	completedLabel := qt.NewQLabel(dialog.QWidget)
	completedLabel.SetText("Completed: " + completed)
	facts.AddWidget(completedLabel.QWidget)
	facts.AddStretch()

	closeButton := qt.NewQPushButton(dialog.QWidget)
	closeButton.SetText("Close")
	closeButton.OnClicked(func() { dialog.Close() })
	facts.AddWidget(closeButton.QWidget)
	layout.AddLayout2(facts.QLayout, 1)

	dialog.Show()
	return dialog
}

// ResultsDialogModule offers Show as an OpenTeacher "resultsDialog" module.
type ResultsDialogModule struct {
	*core.BaseModule
	manager *core.Manager
}

// NewResultsDialogModule creates the module.
func NewResultsDialogModule() *ResultsDialogModule {
	return &ResultsDialogModule{BaseModule: core.NewBaseModule("resultsDialog", "results-dialog")}
}

// ShowResults opens the results of a test.
func (mod *ResultsDialogModule) ShowResults(parent *qt.QWidget, report teaching.Report, grade, notation string) {
	Show(parent, report, grade, notation)
}

func (mod *ResultsDialogModule) Enable(ctx context.Context) error { return mod.BaseModule.Enable(ctx) }
func (mod *ResultsDialogModule) Disable(ctx context.Context) error {
	return mod.BaseModule.Disable(ctx)
}

// SetManager sets the module manager reference.
func (mod *ResultsDialogModule) SetManager(manager *core.Manager) { mod.manager = manager }

// InitResultsDialogModule creates and returns the module.
func InitResultsDialogModule() core.Module { return NewResultsDialogModule() }
