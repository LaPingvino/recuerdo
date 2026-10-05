package testmode

import (
	"crypto/tls"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/testserver"
)

var flowErr, teacherErr error

func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"testmode-test"})
	flowErr = studentFlow()
	teacherErr = teacherFlow()
	os.Exit(m.Run())
}

func TestTeacherFlow(t *testing.T) {
	if teacherErr != nil {
		t.Fatal(teacherErr)
	}
}

func TestStudentFlow(t *testing.T) {
	if flowErr != nil {
		t.Fatal(flowErr)
	}
}

type failure string

func (f failure) Error() string { return string(f) }

func shot(w *Window, name string) {
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		qt.QCoreApplication_ProcessEvents()
		w.Grab().Save(filepath.Join(dir, name))
	}
}

func studentFlow() error {
	dir, _ := os.MkdirTemp("", "testmode")
	defer os.RemoveAll(dir)
	StateFile = func() string { return filepath.Join(dir, "testmode.json") }
	asked := 0
	ConfirmFingerprint = func(_ *qt.QWidget, _, fp string) bool { asked++; return true }
	HandInConfirm = func(*qt.QWidget, string) bool { return true }

	store, err := testserver.Open(filepath.Join(dir, "t.db"))
	if err != nil {
		return err
	}
	defer store.Close()
	teacher, _ := store.CreateUser("jansen", testserver.Teacher, "t")
	anna, _ := store.CreateUser("anna", testserver.Student, "s")
	test, err := store.CreateTest(teacher, lesson.WordList{Title: "Chemistry and maths", Items: []lesson.WordItem{
		{ID: 0, Questions: []string{"water"}, Answers: []string{"H<sub>2</sub>O"}},
		{ID: 1, Questions: []string{"the area of a circle"}, Answers: []string{"$\\pi r^2$"}},
		{ID: 2, Questions: []string{"<ruby>水<rt>みず</rt></ruby>"}, Answers: []string{"water"}},
	}})
	if err != nil {
		return err
	}
	store.AssignStudent(test.ID, anna.ID)
	cert, key, _, err := testserver.Certificate(dir)
	if err != nil {
		return err
	}
	pair, _ := tls.LoadX509KeyPair(cert, key)
	srv := httptest.NewUnstartedServer((&testserver.API{Store: store, Secure: true}).Handler(nil))
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{pair}}
	srv.StartTLS()
	defer srv.Close()
	addr := strings.TrimPrefix(srv.URL, "https://")

	w := Show(nil, func() *lesson.WordList { return nil })
	defer func() {
		w.Close()
		w.DeleteLater()
		qt.QCoreApplication_SendPostedEvents2(nil, int(qt.QEvent__DeferredDelete))
	}()
	shot(w, "1-login.png")
	w.connect(addr, "anna", "wrong")
	if w.client != nil || asked != 1 {
		return failure("logged in with a wrong password, or not asked about the certificate")
	}
	w.connect(addr, "anna", "s")
	if w.client == nil || asked != 1 { // the fingerprint is remembered
		return failure("not logged in: " + w.status.Text())
	}
	if loadState().Pins[addr] == "" || loadState().Name != "anna" {
		return failure("not remembered")
	}
	shot(w, "2-tests.png")
	w.takeTest(test.ID)
	if len(w.answers) != 3 {
		return failure("answer fields")
	}
	w.answers[0].SetText("H2O")
	w.answers[1].SetText("\\pi r")
	w.answers[1].SetCursorPosition(len("\\pi r"))
	if len(w.pads) != 1 {
		return failure("one formula builder expected")
	}
	w.pads[0].Press("sup") // x² button: \pi r^{}
	w.answers[1].Insert("2")
	if w.answers[1].Text() != "\\pi r^{2}" {
		return failure("formula: " + w.answers[1].Text())
	}
	shot(w, "3-take.png")
	w.handIn(test.ID, true)
	if !strings.Contains(w.status.Text(), "Handed in") {
		return failure("hand in: " + w.status.Text())
	}
	r, err := store.ResultOf(test.ID, anna.ID)
	if err != nil || r.Note != 66 {
		return failure("checked on the server: " + strings.TrimSpace(w.status.Text()))
	}
	store.Publish(test.ID, 0, true)
	w.showStudentTests()
	w.showStudentResult(test.ID)
	shot(w, "4-result.png")
	return nil
}

func teacherFlow() error {
	dir, _ := os.MkdirTemp("", "testmode")
	defer os.RemoveAll(dir)
	StateFile = func() string { return filepath.Join(dir, "testmode.json") }
	ConfirmFingerprint = func(*qt.QWidget, string, string) bool { return true }
	store, err := testserver.Open(filepath.Join(dir, "t.db"))
	if err != nil {
		return err
	}
	defer store.Close()
	store.CreateUser("jansen", testserver.Teacher, "t")
	anna, _ := store.CreateUser("anna", testserver.Student, "s")
	bram, _ := store.CreateUser("bram", testserver.Student, "s")
	carla, _ := store.CreateUser("carla", testserver.Student, "s")
	class, _ := store.CreateGroup("2B")
	store.AddMember(class.ID, anna.ID)
	store.AddMember(class.ID, bram.ID)
	srv := httptest.NewServer((&testserver.API{Store: store}).Handler(nil))
	defer srv.Close()

	open := &lesson.WordList{Title: "Chemistry", Items: []lesson.WordItem{
		{ID: 0, Questions: []string{"water"}, Answers: []string{"H<sub>2</sub>O"}},
		{ID: 1, Questions: []string{"the area of a circle"}, Answers: []string{"$\\pi r^2$"}},
		{ID: 2, Questions: []string{"salt"}, Answers: []string{"NaCl"}},
	}}
	w := Show(nil, func() *lesson.WordList { return open })
	defer func() {
		w.Close()
		w.DeleteLater()
		qt.QCoreApplication_SendPostedEvents2(nil, int(qt.QEvent__DeferredDelete))
	}()
	w.Resize(800, 640)
	w.connect(srv.URL, "jansen", "t")
	if w.client == nil {
		return failure("teacher not logged in: " + w.status.Text())
	}
	w.makeTest()
	tests, _ := store.TestsFor(testserver.User{ID: 1, Role: testserver.Teacher})
	if len(tests) != 1 || tests[0].Title != "Chemistry" {
		return failure("test not made")
	}
	id := tests[0].ID
	store.AssignGroup(id, class.ID)
	store.AssignStudent(id, carla.ID)
	store.HandIn(id, anna, map[int]string{0: "H2O", 1: "\\pi r^2", 2: "NaCl"})
	store.HandIn(id, bram, map[int]string{0: "HO2", 1: "pi r^2", 2: "NaCl"})
	w.showTeacherTests()
	shot(w, "5-teacher-tests.png")
	w.showTeacherTest(id)
	shot(w, "6-teacher-test.png")
	test, _ := w.client.Test(id)
	w.showTeacherResult(test, bram.ID)
	shot(w, "7-teacher-check.png")
	if _, err := w.client.Override(id, bram.ID, 1, true); err != nil {
		return err
	}
	if r, _ := store.ResultOf(id, bram.ID); r.Note != 66 {
		return failure("override not stored")
	}
	if err := w.client.Publish(id, 0); err != nil {
		return err
	}
	if r, _ := store.ResultOf(id, anna.ID); !r.Published {
		return failure("not published")
	}
	return nil
}
