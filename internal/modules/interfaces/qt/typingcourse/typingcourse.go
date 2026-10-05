// Package typingcourse is the desktop's touch typing course (File >
// Typing Course): profiles, the tutor's instructions between exercises,
// and the exercises with an on-screen keyboard (internal/typing).
package typingcourse

import (
	"fmt"
	"html"
	"math/rand"
	"strings"
	"time"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/LaPingvino/recuerdo/internal/langcode"
	"github.com/LaPingvino/recuerdo/internal/typing"
)

// ProfilesPath is where the profiles are kept (replaceable in tests).
var ProfilesPath = typing.DefaultPath

// Window is the typing course.
type Window struct {
	*qt.QDialog
	stack    *qt.QStackedWidget
	page     *qt.QWidget
	status   *qt.QLabel
	profiles *typing.Profiles
	profile  *typing.Profile
	rng      *rand.Rand

	// the exercise being typed
	text     []rune
	pos      int
	mistakes int
	started  time.Time
	shown    *qt.QLabel
	keyboard *Keyboard
	now      func() time.Time
}

// Show opens the typing course.
func Show(parent *qt.QWidget) *Window {
	w := &Window{QDialog: qt.NewQDialog(parent), rng: rand.New(rand.NewSource(time.Now().UnixNano())), now: time.Now}
	w.SetWindowTitle(i18n.T("Typing course"))
	w.Resize(820, 620)
	layout := qt.NewQVBoxLayout(w.QWidget)
	w.stack = qt.NewQStackedWidget(w.QWidget)
	layout.AddWidget(w.stack.QWidget)
	w.status = qt.NewQLabel(w.QWidget)
	w.status.SetWordWrap(true)
	layout.AddWidget(w.status.QWidget)
	var err error
	if w.profiles, err = typing.Load(ProfilesPath()); err != nil {
		w.status.SetText(err.Error())
		w.profiles, _ = typing.Load("")
	}
	w.showProfiles()
	w.Show()
	return w
}

func (w *Window) setPage(page *qt.QWidget) {
	if w.page != nil {
		w.stack.RemoveWidget(w.page)
		w.page.DeleteLater()
	}
	w.page = page
	w.stack.AddWidget(page)
	w.stack.SetCurrentWidget(page)
}

func heading(text string) *qt.QLabel {
	l := qt.NewQLabel3(text)
	l.SetStyleSheet("font-size: 17px; font-weight: bold;")
	return l
}

func button(text string, onClick func()) *qt.QPushButton {
	b := qt.NewQPushButton3(text)
	b.OnClicked(onClick)
	return b
}

// ---- profiles ----

func (w *Window) showProfiles() {
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout(page)
	layout.AddWidget(heading(i18n.T("Typing course")).QWidget)
	intro := qt.NewQLabel3(i18n.T("Learn to type with all ten fingers, without looking at the keyboard: first the letters, row by row, then words, faster and faster."))
	intro.SetWordWrap(true)
	layout.AddWidget(intro.QWidget)

	if len(w.profiles.List) > 0 {
		list := qt.NewQListWidget2()
		for _, p := range w.profiles.List {
			list.AddItem(fmt.Sprintf("%s — %s", p.Name, i18n.Tf("level %d of %d", p.Level+1, typing.Levels(p.KeyboardLayout()))))
		}
		list.SetCurrentRow(0)
		list.SetMaximumHeight(130)
		cont := func() {
			if r := list.CurrentRow(); r >= 0 {
				w.profile = w.profiles.List[r]
				w.showInstructions()
			}
		}
		list.OnItemDoubleClicked(func(*qt.QListWidgetItem) { cont() })
		layout.AddWidget(qt.NewQLabel3(i18n.T("Who is practising?")).QWidget)
		layout.AddWidget(list.QWidget)
		row := qt.NewQHBoxLayout2()
		c := button(i18n.T("Continue"), cont)
		c.SetDefault(true)
		row.AddWidget(c.QWidget)
		row.AddStretch()
		layout.AddLayout(row.QLayout)
	}

	// a new profile: name, keyboard (with a picture), language of the words
	layout.AddWidget(heading(i18n.T("New profile")).QWidget)
	form := qt.NewQFormLayout2()
	name := qt.NewQLineEdit2()
	kbd := qt.NewQComboBox2()
	for _, l := range typing.Layouts {
		kbd.AddItem3(i18n.T(l.Name), qt.NewQVariant11(l.ID))
	}
	lang := qt.NewQComboBox2()
	for _, code := range typing.WordLanguages() {
		lang.AddItem3(i18n.T(langcode.Name(code)), qt.NewQVariant11(code))
	}
	// the words in the interface language, or English
	for _, code := range []string{"en", strings.SplitN(i18n.Current(), "_", 2)[0]} {
		if i := lang.FindData(qt.NewQVariant11(code)); i >= 0 && code != "" {
			lang.SetCurrentIndex(i)
		}
	}
	form.AddRow3(i18n.T("Name"), name.QWidget)
	form.AddRow3(i18n.T("Keyboard"), kbd.QWidget)
	form.AddRow3(i18n.T("Words in"), lang.QWidget)
	layout.AddLayout(form.QLayout)
	preview := NewKeyboard(typing.Layouts[0], page)
	kbd.OnCurrentIndexChanged(func(i int) { preview.SetLayout(typing.LayoutByID(kbd.CurrentData().ToString())) })
	layout.AddWidget(preview.QWidget)
	row := qt.NewQHBoxLayout2()
	row.AddWidget(button(i18n.T("Start"), func() {
		p, err := w.profiles.Add(name.Text(), kbd.CurrentData().ToString(), lang.CurrentData().ToString(), w.rng)
		if err != nil {
			msg := i18n.T("A name is needed.")
			if err == typing.ErrNameTaken {
				msg = i18n.T("That name is taken: choose it in the list above.")
			}
			w.message(msg, false)
			return
		}
		w.save()
		w.profile = p
		w.showInstructions()
	}).QWidget)
	row.AddStretch()
	layout.AddLayout(row.QLayout)
	layout.AddStretch()
	w.setPage(page)
}

func (w *Window) save() {
	if err := w.profiles.Save(); err != nil {
		w.message(err.Error(), false)
	}
}

func (w *Window) message(text string, ok bool) {
	color := "#c2453d"
	if ok {
		color = "#3b8f4f"
	}
	w.status.SetStyleSheet("color: " + color + ";")
	w.status.SetText(text)
}

// ---- instructions ----

func (w *Window) showInstructions() {
	w.message("", true)
	p, l := w.profile, w.profile.KeyboardLayout()
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout(page)
	top := qt.NewQHBoxLayout2()
	top.AddWidget(heading(p.Name).QWidget)
	top.AddStretch()
	top.AddWidget(button(i18n.T("Other profile"), w.showProfiles).QWidget)
	layout.AddLayout(top.QLayout)
	text := qt.NewQLabel3(p.Instruction(w.rng))
	text.SetWordWrap(true)
	text.SetStyleSheet("font-size: 14px;")
	layout.AddWidget(text.QWidget)
	form := qt.NewQFormLayout2()
	form.AddRow3(i18n.T("Level"), qt.NewQLabel3(fmt.Sprintf("%d / %d", p.Level+1, typing.Levels(l))).QWidget)
	target := typing.TargetSpeed(l, p.Level)
	if last, ok := p.Last(); ok {
		form.AddRow3(i18n.T("Speed"), qt.NewQLabel3(i18n.Tf("%d words per minute (needed: %d)", last.Speed(), target)).QWidget)
		form.AddRow3(i18n.T("Mistakes"), qt.NewQLabel3(fmt.Sprint(last.Mistakes)).QWidget)
	} else {
		form.AddRow3(i18n.T("Speed needed"), qt.NewQLabel3(i18n.Tf("%d words per minute", target)).QWidget)
	}
	layout.AddLayout(form.QLayout)
	start := button(i18n.T("Start the exercise"), w.startExercise)
	start.SetDefault(true)
	row := qt.NewQHBoxLayout2()
	row.AddWidget(start.QWidget)
	row.AddStretch()
	layout.AddLayout(row.QLayout)
	layout.AddStretch()
	w.setPage(page)
	start.SetFocus()
}

// ---- the exercise ----

func (w *Window) startExercise() {
	w.text, w.pos, w.mistakes, w.started = []rune(w.profile.Current), 0, 0, time.Time{}
	page := qt.NewQWidget2()
	page.SetFocusPolicy(qt.StrongFocus)
	layout := qt.NewQVBoxLayout(page)
	hint := qt.NewQLabel3(i18n.T("Type the text below. The clock starts at your first key; a wrong key is shown in red and must be typed again."))
	hint.SetWordWrap(true)
	hint.SetStyleSheet("color: gray;")
	hint.SetSizePolicy2(qt.QSizePolicy__Preferred, qt.QSizePolicy__Maximum)
	layout.AddWidget(hint.QWidget)
	w.shown = qt.NewQLabel(page)
	w.shown.SetTextFormat(qt.RichText)
	w.shown.SetWordWrap(true)
	w.shown.SetStyleSheet("font-family: monospace; font-size: 22px; padding: 12px;")
	w.shown.SetSizePolicy2(qt.QSizePolicy__Preferred, qt.QSizePolicy__Maximum)
	layout.AddWidget(w.shown.QWidget)
	w.keyboard = NewKeyboard(w.profile.KeyboardLayout(), page)
	layout.AddWidget(w.keyboard.QWidget)
	row := qt.NewQHBoxLayout2()
	stop := button(i18n.T("Stop"), w.showInstructions)
	stop.SetFocusPolicy(qt.NoFocus)
	row.AddWidget(stop.QWidget)
	row.AddStretch()
	layout.AddLayout(row.QLayout)
	page.OnKeyPressEvent(func(super func(*qt.QKeyEvent), e *qt.QKeyEvent) {
		if e.Text() == "" {
			super(e)
			return
		}
		w.Type(e.Text())
	})
	w.setPage(page)
	w.render("")
	page.SetFocus()
}

// Type takes typed characters, as from the keyboard.
func (w *Window) Type(typed string) {
	for _, r := range typed {
		if w.pos >= len(w.text) {
			return
		}
		if w.started.IsZero() {
			w.started = w.now()
		}
		if r != w.text[w.pos] {
			w.mistakes++
			w.render(string(r))
			continue
		}
		w.pos++
		if w.pos == len(w.text) {
			w.finish()
			return
		}
		w.render("")
	}
}

// render shows the text (typed grey, the next character underlined) and
// marks the keyboard.
func (w *Window) render(wrong string) {
	done := html.EscapeString(string(w.text[:w.pos]))
	next := html.EscapeString(string(w.text[w.pos : w.pos+1]))
	if next == " " {
		next = "&nbsp;"
	}
	rest := html.EscapeString(string(w.text[w.pos+1:]))
	w.shown.SetText(`<span style="color:#9aa4ae">` + done + `</span><span style="text-decoration: underline; font-weight: bold;">` +
		next + `</span><span style="color:#555">` + rest + `</span>`)
	w.keyboard.Mark(string(w.text[w.pos]), wrong)
	if wrong != "" {
		w.message(i18n.Tf("That's a mistake (mistakes: %d).", w.mistakes), false)
	} else {
		w.message("", true)
	}
}

func (w *Window) finish() {
	seconds := w.now().Sub(w.started).Seconds()
	w.profile.Finish(seconds, w.mistakes, w.rng)
	w.save()
	w.showInstructions()
}
