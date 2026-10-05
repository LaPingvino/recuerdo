package typingcourse

import (
	"strings"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LaPingvino/recuerdo/internal/typing"
)

// fingerColors are fixed (OpenTeacher picked random ones every time):
// the left hand warm, the right hand cool, the thumbs grey.
var fingerColors = map[int][3]int{
	1: {244, 163, 163}, 2: {246, 195, 143}, 3: {243, 224, 138}, 4: {200, 230, 160},
	7: {166, 220, 216}, 8: {169, 200, 240}, 9: {195, 181, 239}, 10: {231, 179, 224},
	5: {215, 215, 215},
}

// Keyboard draws a layout's keys, each in its finger's colour; the key
// to type next black, a wrongly typed key red.
type Keyboard struct {
	*qt.QWidget
	layout  typing.Layout
	current string // the character to type next
	wrong   string // the character typed wrongly
}

// NewKeyboard makes an on-screen keyboard.
func NewKeyboard(l typing.Layout, parent *qt.QWidget) *Keyboard {
	k := &Keyboard{QWidget: qt.NewQWidget(parent), layout: l}
	k.SetMinimumSize2(450, 150)
	k.SetSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Expanding) // the keyboard takes the room
	k.OnPaintEvent(func(_ func(*qt.QPaintEvent), _ *qt.QPaintEvent) { k.paint() })
	return k
}

// SetLayout shows another layout.
func (k *Keyboard) SetLayout(l typing.Layout) {
	k.layout = l
	k.Update()
}

// Show marks the next and the wrong character.
func (k *Keyboard) Mark(current, wrong string) {
	k.current, k.wrong = strings.ToLower(current), strings.ToLower(wrong)
	k.Update()
}

var keyText = map[string]string{
	typing.Backspace: "⌫", typing.Tab: "Tab", typing.Enter: "Enter", typing.CapsLock: "Caps",
	typing.Shift: "Shift", typing.Space: "",
}

func (k *Keyboard) paint() {
	p := qt.NewQPainter2(k.QPaintDevice)
	defer p.End()
	p.SetRenderHint(qt.QPainter__Antialiasing)
	cell := min(float64(k.Width())/15, float64(k.Height())/5)
	font := p.Font()
	font.SetPointSizeF(max(7, cell/3.4))
	p.SetFont(font)
	for r, row := range k.layout.Rows {
		for _, key := range row {
			rect := qt.NewQRectF4(key.X*cell+1, float64(r)*cell+1, key.Width*cell-2, cell-2)
			c := fingerColors[key.Finger]
			bg, fg := qt.NewQColor3(c[0], c[1], c[2]), qt.NewQColor3(40, 40, 40)
			switch {
			case k.wrong != "" && k.matches(key, k.wrong):
				bg, fg = qt.NewQColor3(194, 69, 61), qt.NewQColor3(255, 255, 255)
			case k.current != "" && k.matches(key, k.current):
				bg, fg = qt.NewQColor3(30, 30, 30), qt.NewQColor3(255, 255, 255)
			}
			p.SetBrush(qt.NewQBrush3(bg))
			p.SetPen(qt.NewQColor3(150, 150, 150))
			p.DrawRoundedRect(rect, cell/8, cell/8)
			label := key.Label
			if t, ok := keyText[label]; ok {
				label = t
			}
			p.SetPen(fg)
			p.DrawText5(rect, int(qt.AlignCenter), label)
		}
	}
}

func (k *Keyboard) matches(key typing.Key, c string) bool {
	if c == " " {
		return key.Label == typing.Space
	}
	return key.Label == c
}
