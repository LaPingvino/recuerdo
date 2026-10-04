// Package plaintextwords lets a word list be typed or pasted as text,
// one word per line as "question = answer" (or with a tab), into a new
// lesson. Port of OpenTeacher's interfaces/qt/enterers/plainTextWords.
package plaintextwords

import (
	"github.com/LaPingvino/recuerdo/internal/i18n"
	"path/filepath"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessons/words"
	"github.com/LaPingvino/recuerdo/internal/resources"
	"github.com/mappu/miqt/qt"
)

// Dialog is where the list is typed.
type Dialog struct {
	*qt.QDialog
	edit    *qt.QPlainTextEdit
	problem *qt.QLabel
	items   []lesson.WordItem
}

// New creates the dialog.
func New(parent *qt.QWidget) *Dialog {
	d := &Dialog{QDialog: qt.NewQDialog(parent)}
	d.SetWindowTitle(i18n.T("Type a List"))
	d.Resize(520, 420)
	layout := qt.NewQVBoxLayout(d.QWidget)
	hint := qt.NewQLabel3(i18n.T("Type or paste the words: one per line, with an equals sign (=) or a tab between ") +
		"the question and the answer, like \"hond = dog\". Several words on one side are separated by commas.")
	hint.SetWordWrap(true)
	layout.AddWidget(hint.QWidget)
	d.edit = qt.NewQPlainTextEdit(d.QWidget)
	d.edit.SetPlaceholderText(i18n.T("hond = dog\nkat = cat"))
	layout.AddWidget2(d.edit.QWidget, 1)

	picker := words.NewIntegratedUnicodePicker(filepath.Join(resources.Dir(), "data", "character_sets.json"), d.QWidget)
	picker.OnCharacterSelected(func(ch string) {
		d.edit.InsertPlainText(ch)
		d.edit.SetFocus()
	})
	picker.Hide()
	layout.AddWidget(picker.QWidget)

	d.problem = qt.NewQLabel3("")
	d.problem.SetWordWrap(true)
	d.problem.SetStyleSheet("color: #c62828;")
	d.problem.SetVisible(false)
	layout.AddWidget(d.problem.QWidget)

	row := qt.NewQHBoxLayout2()
	chars := qt.NewQPushButton3("ä é ß…")
	chars.SetToolTip(i18n.T("Show or hide special characters"))
	chars.SetCheckable(true)
	chars.OnToggled(func(on bool) { picker.SetVisible(on) })
	row.AddWidget(chars.QWidget)
	row.AddStretch()
	buttons := qt.NewQDialogButtonBox(d.QWidget)
	buttons.SetStandardButtons(qt.QDialogButtonBox__Ok | qt.QDialogButtonBox__Cancel)
	buttons.Button(qt.QDialogButtonBox__Ok).SetText(i18n.T("Create Lesson"))
	buttons.OnAccepted(d.Check)
	buttons.OnRejected(func() { d.Reject() })
	row.AddWidget(buttons.QWidget)
	layout.AddLayout(row.QLayout)
	return d
}

// SetText puts text in the dialog.
func (d *Dialog) SetText(text string) { d.edit.SetPlainText(text) }

// Check reads the text: a list closes the dialog, a mistake is shown and
// the dialog stays open to correct it.
func (d *Dialog) Check() {
	items, err := lesson.ParseWordList(d.edit.ToPlainText(), false)
	switch {
	case err != nil:
		d.problem.SetText(err.Error() + ": put \"=\" or a tab between the question and the answer.")
	case len(items) == 0:
		d.problem.SetText(i18n.T("Type at least one word."))
	default:
		d.items = items
		d.problem.SetVisible(false)
		d.Accept()
		return
	}
	d.problem.SetVisible(true)
}

// Problem is the mistake shown ("" when none).
func (d *Dialog) Problem() string {
	if d.problem.IsHidden() {
		return ""
	}
	return d.problem.Text()
}

// Lesson is the typed list as a new word lesson (after the dialog was
// accepted), or nil.
func (d *Dialog) Lesson() *lesson.Lesson {
	if len(d.items) == 0 {
		return nil
	}
	l := lesson.NewLesson("words")
	l.Data.List.Title = "Typed list"
	l.Data.List.Items = d.items
	l.Data.Changed = true
	return l
}
