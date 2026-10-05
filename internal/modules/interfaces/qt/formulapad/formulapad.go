// Package formulapad is the desktop's formula builder: the buttons of
// richtext.Palette (shared with the web version) for a QLineEdit, with a
// live preview of the formula.
package formulapad

import (
	"unicode/utf16"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/LaPingvino/recuerdo/internal/richtext"
)

// Pad is a formula builder for one line edit.
type Pad struct {
	*qt.QWidget
	edit    *qt.QLineEdit
	wrap    bool
	Preview *qt.QLabel
}

// New makes a formula builder for edit. wrap: the field holds words with
// formulas in $...$ (the Enter tab), so a button outside a formula starts
// one; otherwise the whole field is TeX (a formula answer).
func New(edit *qt.QLineEdit, wrap bool, parent *qt.QWidget) *Pad {
	p := &Pad{QWidget: qt.NewQWidget(parent), edit: edit, wrap: wrap}
	p.SetObjectName(*qt.NewQAnyStringView3("formulaPad"))
	// formulas are written left to right, also in Arabic or Urdu
	p.SetLayoutDirection(qt.LeftToRight)
	layout := qt.NewQVBoxLayout(p.QWidget)
	layout.SetContentsMargins(4, 4, 4, 4)
	layout.SetSpacing(2)
	for _, g := range richtext.Palette {
		row := qt.NewQHBoxLayout2()
		row.SetSpacing(2)
		name := qt.NewQLabel3(i18n.T(g.Name))
		name.SetMinimumWidth(80)
		name.SetStyleSheet("color: gray; font-size: 11px;")
		row.AddWidget(name.QWidget)
		for _, it := range g.Items {
			b := qt.NewQToolButton(p.QWidget)
			b.SetText(it.Label)
			if it.Tip != "" {
				b.SetToolTip(i18n.T(it.Tip))
			} else {
				b.SetToolTip(it.Template)
			}
			b.SetMinimumWidth(30)
			// the line edit keeps the focus and its selection
			b.SetFocusPolicy(qt.NoFocus)
			b.OnClicked(func() { p.Press(it.ID) })
			row.AddWidget(b.QWidget)
		}
		row.AddStretch()
		layout.AddLayout(row.QLayout)
	}
	p.Preview = qt.NewQLabel(p.QWidget)
	p.Preview.SetTextFormat(qt.RichText)
	p.Preview.SetAlignment(qt.AlignCenter)
	p.Preview.SetMinimumHeight(30)
	p.Preview.SetStyleSheet("font-size: 18px; border-top: 1px dashed gray; padding-top: 4px;")
	layout.AddWidget(p.Preview.QWidget)
	edit.OnTextChanged(func(string) { p.update() })
	p.update()
	return p
}

// Press inserts a button's TeX at the cursor, around the selection.
func (p *Pad) Press(id string) {
	it, ok := richtext.PaletteItemByID(id)
	if !ok {
		return
	}
	start := p.edit.CursorPosition()
	if p.edit.HasSelectedText() {
		start = p.edit.SelectionStart()
	}
	text, cursor := richtext.Expand(it, p.edit.SelectedText())
	if p.wrap && !insideMath(p.edit.Text(), start) {
		text, cursor = "$"+text+"$", cursor+1
	}
	p.edit.Insert(text) // replaces the selection
	p.edit.SetCursorPosition(start + cursor)
	p.edit.SetFocus()
}

func (p *Pad) update() {
	v := p.edit.Text()
	if p.wrap {
		p.Preview.SetText(richtext.RichWithMath(v))
	} else {
		p.Preview.SetText(richtext.TeXToRich(v))
	}
}

// insideMath reports whether position pos (in UTF-16 units, as Qt
// counts) of s is inside a $...$ formula.
func insideMath(s string, pos int) bool {
	u := utf16.Encode([]rune(s))
	if pos > len(u) {
		pos = len(u)
	}
	before := u[:pos]
	n := 0
	for i := 0; i < len(before); i++ {
		switch before[i] {
		case '\\':
			i++
		case '$':
			n++
		}
	}
	return n%2 == 1
}

// Dialog edits text with the formula builder (for table cells: the Enter
// tab). It returns the new text and whether OK was pressed.
func Dialog(parent *qt.QWidget, text string) (string, bool) {
	d := qt.NewQDialog(parent)
	d.SetWindowTitle(i18n.T("Formula builder"))
	layout := qt.NewQVBoxLayout(d.QWidget)
	edit := qt.NewQLineEdit(d.QWidget)
	edit.SetText(text)
	edit.SetMinimumWidth(520)
	layout.AddWidget(edit.QWidget)
	layout.AddWidget(New(edit, true, d.QWidget).QWidget)
	buttons := qt.NewQDialogButtonBox4(qt.QDialogButtonBox__Ok | qt.QDialogButtonBox__Cancel)
	buttons.OnAccepted(d.Accept)
	buttons.OnRejected(d.Reject)
	layout.AddWidget(buttons.QWidget)
	edit.SetFocus()
	ok := d.Exec() == int(qt.QDialog__Accepted)
	result := edit.Text()
	d.DeleteLater()
	return result, ok
}
