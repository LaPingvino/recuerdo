// Package testmode is the desktop's test mode (File > Test mode): connect
// to a test server (`recuerdo testserver`), log in, and as a student take
// the tests a teacher gave you and see your results; as a teacher make
// tests of a lesson, give them to students and groups, check and publish
// the results.
package testmode

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/formulapad"
	"github.com/LaPingvino/recuerdo/internal/richtext"
	"github.com/LaPingvino/recuerdo/internal/testserver"
)

// ---- what is remembered: the last server and name, trusted certificates ----

// StateFile is where test mode remembers things (replaceable in tests).
var StateFile = func() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "recuerdo", "testmode.json")
}

type state struct {
	Address string            `json:"address"`
	Name    string            `json:"name"`
	Pins    map[string]string `json:"pins"` // server address: certificate fingerprint
}

func loadState() state {
	var s state
	if data, err := os.ReadFile(StateFile()); err == nil {
		json.Unmarshal(data, &s)
	}
	if s.Pins == nil {
		s.Pins = map[string]string{}
	}
	return s
}

func (s state) save() error {
	data, _ := json.MarshalIndent(s, "", "  ")
	if err := os.MkdirAll(filepath.Dir(StateFile()), 0o755); err != nil {
		return err
	}
	return os.WriteFile(StateFile(), data, 0o600)
}

// ConfirmFingerprint asks whether a server's certificate is the right
// one (replaceable in tests).
var ConfirmFingerprint = func(parent *qt.QWidget, address, fingerprint string) bool {
	text := i18n.Tf("Is this the test server's certificate? Compare its fingerprint with the one the server shows (ask your teacher):\n\n%s\n\n%s", address, fingerprint)
	return qt.QMessageBox_Question5(parent, i18n.T("New test server"), text,
		qt.QMessageBox__Yes|qt.QMessageBox__No) == qt.QMessageBox__Yes
}

// ---- the window ----

// Window is test mode's window.
type Window struct {
	*qt.QDialog
	stack   *qt.QStackedWidget
	page    *qt.QWidget
	status  *qt.QLabel
	client  *testserver.Client
	user    testserver.User
	state   state
	lesson  func() *lesson.WordList // the open lesson, for teachers
	answers map[int]*qt.QLineEdit   // the test being taken
	pads    []*formulapad.Pad       // its formula builders
}

// Show opens test mode. lesson gives the open lesson's word list (nil
// when there is none).
func Show(parent *qt.QWidget, lesson func() *lesson.WordList) *Window {
	w := &Window{QDialog: qt.NewQDialog(parent), state: loadState(), lesson: lesson}
	w.SetWindowTitle(i18n.T("Test mode"))
	w.Resize(760, 600)
	layout := qt.NewQVBoxLayout(w.QWidget)
	w.stack = qt.NewQStackedWidget(w.QWidget)
	layout.AddWidget(w.stack.QWidget)
	w.status = qt.NewQLabel(w.QWidget)
	w.status.SetWordWrap(true)
	layout.AddWidget(w.status.QWidget)
	w.showLogin()
	w.Show()
	return w
}

// setPage replaces the shown page.
func (w *Window) setPage(page *qt.QWidget) {
	if w.page != nil {
		w.stack.RemoveWidget(w.page)
		w.page.DeleteLater()
	}
	w.page = page
	w.stack.AddWidget(page)
	w.stack.SetCurrentWidget(page)
}

func (w *Window) message(text string, ok bool) {
	color := "#c2453d"
	if ok {
		color = "#3b8f4f"
	}
	w.status.SetStyleSheet("color: " + color + "; font-weight: bold;")
	w.status.SetText(text)
}

func (w *Window) fail(err error) bool {
	if err == nil {
		return false
	}
	w.message(err.Error(), false)
	return true
}

func heading(text string) *qt.QLabel {
	l := qt.NewQLabel3(text)
	l.SetStyleSheet("font-size: 16px; font-weight: bold;")
	return l
}

func button(text string, onClick func()) *qt.QPushButton {
	b := qt.NewQPushButton3(text)
	b.OnClicked(onClick)
	return b
}

// ---- connecting and logging in ----

func (w *Window) showLogin() {
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout(page)
	layout.AddWidget(heading(i18n.T("Log in to a test server")).QWidget)
	form := qt.NewQFormLayout2()
	address := qt.NewQLineEdit3(w.state.Address)
	address.SetPlaceholderText("school.example:8770")
	name := qt.NewQLineEdit3(w.state.Name)
	password := qt.NewQLineEdit2()
	password.SetEchoMode(qt.QLineEdit__Password)
	form.AddRow3(i18n.T("Server"), address.QWidget)
	form.AddRow3(i18n.T("Name"), name.QWidget)
	form.AddRow3(i18n.T("Password"), password.QWidget)
	layout.AddLayout(form.QLayout)
	login := button(i18n.T("Log in"), func() {
		w.connect(strings.TrimSpace(address.Text()), strings.TrimSpace(name.Text()), password.Text())
	})
	login.SetDefault(true)
	password.OnReturnPressed(login.Click)
	layout.AddWidget(login.QWidget)
	hint := qt.NewQLabel3(i18n.T("Your teacher tells you the server's address. Students can also take tests in a web browser, at the same address."))
	hint.SetWordWrap(true)
	hint.SetStyleSheet("color: gray;")
	layout.AddWidget(hint.QWidget)
	layout.AddStretch()
	w.setPage(page)
	if address.Text() == "" {
		address.SetFocus()
	} else if name.Text() == "" {
		name.SetFocus()
	} else {
		password.SetFocus()
	}
}

func (w *Window) connect(address, name, password string) {
	if address == "" || name == "" {
		w.message(i18n.T("Fill in the server, your name and your password."), false)
		return
	}
	client, err := testserver.Connect(address, w.state.Pins[address])
	var unknown *testserver.UnknownCertificate
	if errors.As(err, &unknown) {
		if !ConfirmFingerprint(w.QWidget, address, unknown.Fingerprint) {
			w.message(i18n.T("Not connected: the certificate was not confirmed."), false)
			return
		}
		w.state.Pins[address] = unknown.Fingerprint
		client, err = testserver.Connect(address, unknown.Fingerprint)
	}
	if w.fail(err) {
		return
	}
	user, err := client.Login(name, password)
	if w.fail(err) {
		return
	}
	w.client, w.user = client, user
	w.state.Address, w.state.Name = address, name
	w.state.save()
	w.message("", true)
	w.showHome()
}

func (w *Window) showHome() {
	if w.user.Role == testserver.Student {
		w.showStudentTests()
		return
	}
	w.showTeacherTests()
}

// logoutRow is the top row of pages after logging in.
func (w *Window) logoutRow(title string) *qt.QHBoxLayout {
	row := qt.NewQHBoxLayout2()
	row.AddWidget(heading(title).QWidget)
	row.AddStretch()
	who := qt.NewQLabel3(w.user.Name)
	who.SetStyleSheet("color: gray;")
	row.AddWidget(who.QWidget)
	row.AddWidget(button(i18n.T("Log out"), func() {
		w.client.Logout()
		w.client = nil
		w.message("", true)
		w.showLogin()
	}).QWidget)
	return row
}

func table(headers ...string) *qt.QTableWidget {
	t := qt.NewQTableWidget2()
	t.SetColumnCount(len(headers))
	t.SetHorizontalHeaderLabels(headers)
	t.SetEditTriggers(qt.QAbstractItemView__NoEditTriggers)
	t.SetSelectionBehavior(qt.QAbstractItemView__SelectRows)
	t.SetSelectionMode(qt.QAbstractItemView__SingleSelection)
	t.VerticalHeader().Hide()
	t.HorizontalHeader().SetStretchLastSection(true)
	return t
}

func cell(t *qt.QTableWidget, row, col int, text string) {
	t.SetItem(row, col, qt.NewQTableWidgetItem2(text))
}

// ---- students ----

func (w *Window) showStudentTests() {
	tests, err := w.client.Tests()
	if w.fail(err) {
		return
	}
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout(page)
	layout.AddLayout(w.logoutRow(i18n.T("Your tests")).QLayout)
	if len(tests) == 0 {
		layout.AddWidget(qt.NewQLabel3(i18n.T("No tests for you yet.")).QWidget)
		layout.AddStretch()
		w.setPage(page)
		return
	}
	list := table(i18n.T("Test"), i18n.T("Words"), i18n.T("Status"))
	list.SetRowCount(len(tests))
	for i, t := range tests {
		status := i18n.T("To do")
		switch {
		case t.Published:
			status = i18n.T("Result ready")
		case t.HandedIn:
			status = i18n.T("Handed in")
		case !t.Open:
			status = i18n.T("Closed")
		}
		cell(list, i, 0, t.Title)
		cell(list, i, 1, fmt.Sprint(t.Words))
		cell(list, i, 2, status)
	}
	list.ResizeColumnsToContents()
	layout.AddWidget(list.QWidget)
	take := button(i18n.T("Take the test"), nil)
	result := button(i18n.T("See the result"), nil)
	update := func() {
		r := list.CurrentRow()
		take.SetEnabled(r >= 0 && tests[r].Open && !tests[r].HandedIn)
		result.SetEnabled(r >= 0 && tests[r].Published)
	}
	list.OnCurrentCellChanged(func(int, int, int, int) { update() })
	take.OnClicked(func() { w.takeTest(tests[list.CurrentRow()].ID) })
	result.OnClicked(func() { w.showStudentResult(tests[list.CurrentRow()].ID) })
	row := qt.NewQHBoxLayout2()
	row.AddWidget(take.QWidget)
	row.AddWidget(result.QWidget)
	row.AddWidget(button(i18n.T("Refresh"), w.showStudentTests).QWidget)
	row.AddStretch()
	layout.AddLayout(row.QLayout)
	update()
	w.setPage(page)
}

func (w *Window) takeTest(id int64) {
	test, err := w.client.StudentTest(id)
	if w.fail(err) {
		return
	}
	w.message("", true)
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout(page)
	layout.AddLayout(w.logoutRow(test.Title).QLayout)
	scroll := qt.NewQScrollArea(page)
	scroll.SetWidgetResizable(true)
	inner := qt.NewQWidget2()
	items := qt.NewQVBoxLayout(inner)
	w.answers, w.pads = map[int]*qt.QLineEdit{}, nil
	for n, it := range test.Items {
		q := qt.NewQLabel3(fmt.Sprintf("<b>%d.</b> %s", n+1, richtext.RichWithMath(it.QuestionHTML)))
		q.SetTextFormat(qt.RichText)
		q.SetStyleSheet("font-size: 15px;")
		items.AddWidget(q.QWidget)
		edit := qt.NewQLineEdit2()
		if it.Math {
			edit.SetPlaceholderText(i18n.T("Formula, e.g. x^2 + 1"))
		} else {
			edit.SetPlaceholderText(i18n.T("Your answer"))
		}
		items.AddWidget(edit.QWidget)
		if it.Comment != "" {
			c := qt.NewQLabel3(it.Comment)
			c.SetStyleSheet("color: gray;")
			items.AddWidget(c.QWidget)
		}
		if it.Math {
			pad := formulapad.New(edit, false, inner)
			items.AddWidget(pad.QWidget)
			w.pads = append(w.pads, pad)
		}
		w.answers[it.ID] = edit
	}
	items.AddStretch()
	scroll.SetWidget(inner)
	layout.AddWidget(scroll.QWidget)
	row := qt.NewQHBoxLayout2()
	row.AddWidget(button(i18n.T("Back"), w.showStudentTests).QWidget)
	row.AddStretch()
	row.AddWidget(button(i18n.T("Hand in"), func() { w.handIn(id, true) }).QWidget)
	layout.AddLayout(row.QLayout)
	w.setPage(page)
}

// HandInConfirm asks before handing in (replaceable in tests).
var HandInConfirm = func(parent *qt.QWidget, text string) bool {
	return qt.QMessageBox_Question5(parent, i18n.T("Hand in"), text, qt.QMessageBox__Yes|qt.QMessageBox__No) == qt.QMessageBox__Yes
}

func (w *Window) handIn(id int64, ask bool) {
	answers := map[int]string{}
	empty := 0
	for itemID, edit := range w.answers {
		answers[itemID] = edit.Text()
		if strings.TrimSpace(edit.Text()) == "" {
			empty++
		}
	}
	text := i18n.T("Hand in? You cannot change your answers afterwards.")
	if empty > 0 {
		text = i18n.Tf("%d questions have no answer. Hand in anyway? You cannot change your answers afterwards.", empty)
	}
	if ask && !HandInConfirm(w.QWidget, text) {
		return
	}
	if w.fail(w.client.HandIn(id, answers)) {
		return
	}
	w.message(i18n.T("Handed in. Your teacher will check your answers."), true)
	w.showStudentTests()
}

func (w *Window) showStudentResult(id int64) {
	w.message("", true)
	test, err := w.client.StudentTest(id)
	if w.fail(err) {
		return
	}
	res, err := w.client.Result(id, w.user.ID)
	if w.fail(err) {
		return
	}
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout(page)
	layout.AddLayout(w.logoutRow(test.Title).QLayout)
	right := 0
	for _, it := range res.Items {
		if it.Right {
			right++
		}
	}
	score := heading(i18n.Tf("%d of %d right (%d%%)", right, len(res.Items), res.Note))
	layout.AddWidget(score.QWidget)
	questions := map[int]string{}
	for _, it := range test.Items {
		questions[it.ID] = it.Question
	}
	list := table(i18n.T("Question"), i18n.T("Your answer"), "")
	list.SetRowCount(len(res.Items))
	for i, it := range res.Items {
		cell(list, i, 0, questions[it.ItemID])
		given := it.Given
		if given == "" {
			given = "—"
		}
		cell(list, i, 1, given)
		mark := i18n.T("Wrong")
		if it.Right && it.Overridden {
			mark = i18n.T("Right (your teacher's decision)")
		} else if it.Right {
			mark = i18n.T("Right")
		}
		cell(list, i, 2, mark)
		color := qt.NewQColor3(194, 69, 61)
		if it.Right {
			color = qt.NewQColor3(59, 143, 79)
		}
		list.Item(i, 2).SetForeground(qt.NewQBrush3(color))
	}
	list.ResizeColumnsToContents()
	layout.AddWidget(list.QWidget)
	row := qt.NewQHBoxLayout2()
	row.AddWidget(button(i18n.T("Back"), w.showStudentTests).QWidget)
	row.AddStretch()
	layout.AddLayout(row.QLayout)
	w.setPage(page)
}

// ---- teachers ----

func (w *Window) showTeacherTests() {
	w.message("", true)
	tests, err := w.client.Tests()
	if w.fail(err) {
		return
	}
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout(page)
	layout.AddLayout(w.logoutRow(i18n.T("Your tests")).QLayout)
	if w.user.Role == testserver.Admin {
		hint := qt.NewQLabel3(i18n.T("Accounts and groups are managed on the server's web page."))
		hint.SetStyleSheet("color: gray;")
		layout.AddWidget(hint.QWidget)
	}
	list := table(i18n.T("Test"), i18n.T("Words"), i18n.T("Status"))
	list.SetRowCount(len(tests))
	for i, t := range tests {
		status := i18n.T("Open")
		if !t.Open {
			status = i18n.T("Closed")
		}
		cell(list, i, 0, t.Title)
		cell(list, i, 1, fmt.Sprint(t.Words))
		cell(list, i, 2, status)
	}
	list.ResizeColumnsToContents()
	layout.AddWidget(list.QWidget)
	open := button(i18n.T("Open the test"), func() { w.showTeacherTest(tests[list.CurrentRow()].ID) })
	open.SetEnabled(false)
	list.OnCurrentCellChanged(func(row, _, _, _ int) { open.SetEnabled(row >= 0) })
	list.OnCellDoubleClicked(func(row, _ int) { w.showTeacherTest(tests[row].ID) })
	row := qt.NewQHBoxLayout2()
	row.AddWidget(open.QWidget)
	if w.user.Role == testserver.Teacher {
		make := button(i18n.T("New test of the open lesson"), w.makeTest)
		make.SetEnabled(w.lesson != nil && w.lesson() != nil && len(w.lesson().Items) > 0)
		if !make.IsEnabled() {
			make.SetToolTip(i18n.T("Open a word lesson in Recuerdo first."))
		}
		row.AddWidget(make.QWidget)
	}
	row.AddWidget(button(i18n.T("Refresh"), w.showTeacherTests).QWidget)
	row.AddStretch()
	layout.AddLayout(row.QLayout)
	w.setPage(page)
}

func (w *Window) makeTest() {
	list := w.lesson()
	if list == nil {
		return
	}
	test, err := w.client.CreateTest(*list)
	if w.fail(err) {
		return
	}
	w.message(i18n.Tf("Test made: %s. Now give it to students or groups.", test.Title), true)
	w.showTeacherTest(test.ID)
}

func (w *Window) showTeacherTest(id int64) {
	test, err := w.client.Test(id)
	if w.fail(err) {
		return
	}
	results, err := w.client.Results(id)
	if w.fail(err) {
		return
	}
	students, _ := w.client.Users()
	groups, _ := w.client.Groups()
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout(page)
	layout.AddLayout(w.logoutRow(test.Title).QLayout)
	reload := func() { w.showTeacherTest(id) }

	top := qt.NewQHBoxLayout2()
	openText := i18n.T("Close for handing in")
	if !test.Open {
		openText = i18n.T("Open for handing in")
	}
	top.AddWidget(button(openText, func() {
		if !w.fail(w.client.SetOpen(id, !test.Open)) {
			reload()
		}
	}).QWidget)
	top.AddStretch()
	layout.AddLayout(top.QLayout)

	// who takes it
	var given []string
	for _, s := range test.Students {
		given = append(given, s.Name)
	}
	for _, g := range test.Groups {
		given = append(given, i18n.Tf("Group %s", g.Name))
	}
	if len(given) == 0 {
		given = []string{i18n.T("No one yet.")}
	}
	who := qt.NewQLabel3("<b>" + i18n.T("Given to") + ":</b> " + html.EscapeString(strings.Join(given, ", ")))
	who.SetWordWrap(true)
	layout.AddWidget(who.QWidget)
	assign := qt.NewQHBoxLayout2()
	student := qt.NewQComboBox2()
	student.AddItem(i18n.T("A student…"))
	for _, s := range students {
		student.AddItem3(s.Name, qt.NewQVariant6(s.ID))
	}
	group := qt.NewQComboBox2()
	group.AddItem(i18n.T("A group…"))
	for _, g := range groups {
		group.AddItem3(g.Name, qt.NewQVariant6(g.ID))
	}
	assign.AddWidget(student.QWidget)
	assign.AddWidget(button(i18n.T("Add"), func() {
		if student.CurrentIndex() > 0 && !w.fail(w.client.AssignStudent(id, student.CurrentData().ToLongLong())) {
			reload()
		}
	}).QWidget)
	assign.AddWidget(group.QWidget)
	assign.AddWidget(button(i18n.T("Add"), func() {
		if group.CurrentIndex() > 0 && !w.fail(w.client.AssignGroup(id, group.CurrentData().ToLongLong())) {
			reload()
		}
	}).QWidget)
	assign.AddStretch()
	layout.AddLayout(assign.QLayout)

	// hand-ins
	layout.AddWidget(heading(i18n.T("Handed in")).QWidget)
	list := table(i18n.T("Student"), i18n.T("Score"), i18n.T("Published"))
	list.SetRowCount(len(results))
	handedIn := map[int64]bool{}
	for i, r := range results {
		handedIn[r.Student.ID] = true
		cell(list, i, 0, r.Student.Name)
		cell(list, i, 1, fmt.Sprintf("%d/%d (%d%%)", rightCount(r), len(r.Items), r.Note))
		published := "–"
		if r.Published {
			published = "✓"
		}
		cell(list, i, 2, published)
	}
	list.ResizeColumnsToContents()
	layout.AddWidget(list.QWidget)
	expected := map[int64]string{}
	for _, s := range test.Students {
		expected[s.ID] = s.Name
	}
	for _, g := range test.Groups {
		for _, full := range groups {
			if full.ID == g.ID {
				for _, m := range full.Members {
					expected[m.ID] = m.Name
				}
			}
		}
	}
	var missing []string
	for sid, name := range expected {
		if !handedIn[sid] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		l := qt.NewQLabel3(i18n.Tf("Not handed in yet: %s", strings.Join(missing, ", ")))
		l.SetWordWrap(true)
		l.SetStyleSheet("color: gray;")
		layout.AddWidget(l.QWidget)
	}
	row := qt.NewQHBoxLayout2()
	row.AddWidget(button(i18n.T("Back"), w.showTeacherTests).QWidget)
	check := button(i18n.T("Check"), func() { w.showTeacherResult(test, results[list.CurrentRow()].Student.ID) })
	check.SetEnabled(false)
	list.OnCurrentCellChanged(func(r, _, _, _ int) { check.SetEnabled(r >= 0) })
	list.OnCellDoubleClicked(func(r, _ int) { w.showTeacherResult(test, results[r].Student.ID) })
	row.AddWidget(check.QWidget)
	row.AddStretch()
	if len(results) > 0 {
		row.AddWidget(button(i18n.T("Publish all results"), func() {
			if !w.fail(w.client.Publish(id, 0)) {
				w.message(i18n.T("Published: the students can see their results."), true)
				reload()
			}
		}).QWidget)
	}
	layout.AddLayout(row.QLayout)
	w.setPage(page)
}

func rightCount(r testserver.Result) int {
	n := 0
	for _, it := range r.Items {
		if it.Right {
			n++
		}
	}
	return n
}

func (w *Window) showTeacherResult(test testserver.Test, studentID int64) {
	w.message("", true)
	r, err := w.client.Result(test.ID, studentID)
	if w.fail(err) {
		return
	}
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout(page)
	layout.AddLayout(w.logoutRow(test.Title + ": " + r.Student.Name).QLayout)
	layout.AddWidget(heading(i18n.Tf("%d of %d right (%d%%)", rightCount(r), len(r.Items), r.Note)).QWidget)
	words := map[int]lesson.WordItem{}
	for _, it := range test.List.Items {
		words[it.ID] = it
	}
	list := table(i18n.T("Question"), i18n.T("Right answers"), i18n.T("Given"), "")
	list.SetRowCount(len(r.Items))
	for i, it := range r.Items {
		word := words[it.ItemID]
		for col, text := range []string{strings.Join(word.Questions, ", "), strings.Join(word.Answers, ", ")} {
			// the words with their markup and formulas, as rich text
			l := qt.NewQLabel3(richtext.RichWithMath(text))
			l.SetTextFormat(qt.RichText)
			l.SetContentsMargins(4, 0, 4, 0)
			list.SetCellWidget(i, col, l.QWidget)
		}
		given := it.Given
		if given == "" {
			given = "—"
		}
		cell(list, i, 2, given)
		mark, color := i18n.T("Wrong"), qt.NewQColor3(194, 69, 61)
		if it.Right {
			mark, color = i18n.T("Right"), qt.NewQColor3(59, 143, 79)
		}
		cell(list, i, 3, mark)
		list.Item(i, 3).SetForeground(qt.NewQBrush3(color))
	}
	list.ResizeColumnsToContents()
	layout.AddWidget(list.QWidget)
	row := qt.NewQHBoxLayout2()
	row.AddWidget(button(i18n.T("Back to the test"), func() { w.showTeacherTest(test.ID) }).QWidget)
	flip := button(i18n.T("Count as right"), nil)
	flip.SetEnabled(false)
	list.OnCurrentCellChanged(func(row, _, _, _ int) {
		flip.SetEnabled(row >= 0)
		if row >= 0 && r.Items[row].Right {
			flip.SetText(i18n.T("Count as wrong"))
		} else {
			flip.SetText(i18n.T("Count as right"))
		}
	})
	flip.OnClicked(func() {
		it := r.Items[list.CurrentRow()]
		if _, err := w.client.Override(test.ID, studentID, it.ItemID, !it.Right); !w.fail(err) {
			w.showTeacherResult(test, studentID)
		}
	})
	row.AddWidget(flip.QWidget)
	row.AddStretch()
	publish := button(i18n.T("Publish this result"), func() {
		if !w.fail(w.client.Publish(test.ID, studentID)) {
			w.message(i18n.T("Published: the students can see their results."), true)
			w.showTeacherResult(test, studentID)
		}
	})
	if r.Published {
		publish.SetText(i18n.T("Published"))
		publish.SetEnabled(false)
	}
	row.AddWidget(publish.QWidget)
	layout.AddLayout(row.QLayout)
	w.setPage(page)
}
