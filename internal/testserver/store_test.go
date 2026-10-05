package testserver

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// must is v, failing the test (by panicking) on an error
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func noErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

var animals = lesson.WordList{Title: "Animals", Items: []lesson.WordItem{
	{ID: 0, Questions: []string{"hond"}, Answers: []string{"dog"}},
	{ID: 1, Questions: []string{"kat"}, Answers: []string{"cat", "kitty"}},
	{ID: 2, Questions: []string{"area of a circle"}, Answers: []string{"$\\pi r^2$"}},
	{ID: 3, Questions: []string{"water"}, Answers: []string{"H<sub>2</sub>O"}},
}}

func TestUsersAndSessions(t *testing.T) {
	s := newStore(t)
	u := must(s.CreateUser("anna", Student, "secret"))
	if _, err := s.CreateUser("anna", Teacher, "x"); !errors.Is(err, ErrExists) {
		t.Errorf("same name twice: %v", err)
	}
	if _, err := s.CreateUser("bob", "janitor", "x"); err == nil {
		t.Error("unknown role accepted")
	}
	if _, _, err := s.Login("anna", "wrong"); !errors.Is(err, ErrLogin) {
		t.Errorf("wrong password: %v", err)
	}
	if _, _, err := s.Login("nobody", "secret"); !errors.Is(err, ErrLogin) {
		t.Errorf("unknown user: %v", err)
	}
	token, who := must2(t)(s.Login(" anna ", "secret"))
	if who.ID != u.ID || who.Role != Student {
		t.Errorf("logged in as %+v", who)
	}
	if got := must(s.Session(token)); got.Name != "anna" {
		t.Errorf("session %+v", got)
	}
	// sessions expire, and end with a new password or a logout
	s.now = func() time.Time { return time.Now().Add(SessionTime + time.Minute) }
	if _, err := s.Session(token); !errors.Is(err, ErrNotFound) {
		t.Errorf("expired session: %v", err)
	}
	s.now = time.Now
	token, _ = must2(t)(s.Login("anna", "secret"))
	if err := s.SetPassword(u.ID, "new"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Session(token); err == nil {
		t.Error("session survived a new password")
	}
	token, _ = must2(t)(s.Login("anna", "new"))
	s.Logout(token)
	if _, err := s.Session(token); err == nil {
		t.Error("session survived a logout")
	}
}

func must2(t *testing.T) func(string, User, error) (string, User) {
	return func(a string, b User, err error) (string, User) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return a, b
	}
}

func TestClassroom(t *testing.T) {
	s := newStore(t)
	teacher := must(s.CreateUser("ms. jansen", Teacher, "t"))
	anna := must(s.CreateUser("anna", Student, "a"))
	bram := must(s.CreateUser("bram", Student, "b"))
	carla := must(s.CreateUser("carla", Student, "c"))
	class := must(s.CreateGroup("2B"))
	if err := s.AddMember(class.ID, teacher.ID); err == nil {
		t.Error("a teacher became a group member")
	}
	for _, u := range []User{anna, bram} {
		if err := s.AddMember(class.ID, u.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateTest(anna, animals); !errors.Is(err, ErrForbidden) {
		t.Errorf("a student made a test: %v", err)
	}
	test := must(s.CreateTest(teacher, animals))
	noErr(t, s.AssignGroup(test.ID, class.ID))
	noErr(t, s.AssignStudent(test.ID, carla.ID))

	// what students see: their tests, without the answers
	if list := must(s.TestsFor(anna)); len(list) != 1 || list[0].Words != 4 || list[0].HandedIn {
		t.Errorf("anna's tests %+v", list)
	}
	if list := must(s.TestsFor(teacher)); len(list) != 1 {
		t.Errorf("teacher's tests %+v", list)
	}
	full := must(s.Test(test.ID))
	if len(full.Students) != 1 || len(full.Groups) != 1 {
		t.Errorf("assigned to %+v %+v", full.Students, full.Groups)
	}
	for _, it := range full.ForStudent().List.Items {
		if len(it.Answers) != 0 {
			t.Errorf("a student sees answers: %+v", it)
		}
	}

	// hand in: checked with OpenTeacher's notation, formulas and markup
	r := must(s.HandIn(test.ID, anna, map[int]string{0: "dog", 1: "kitty", 2: "\\pi r^2", 3: "H2O"}))
	if r.Note != 100 {
		t.Errorf("all right: %+v", r)
	}
	r = must(s.HandIn(test.ID, bram, map[int]string{0: "dgo", 1: "cat"}))
	if r.Note != 25 || r.Items[0].Right || !r.Items[1].Right || r.Items[2].Right {
		t.Errorf("bram %+v", r)
	}
	if _, err := s.HandIn(test.ID, bram, map[int]string{0: "dog"}); !errors.Is(err, ErrHandedIn) {
		t.Errorf("handed in twice: %v", err)
	}
	outsider := must(s.CreateUser("dirk", Student, "d"))
	if _, err := s.HandIn(test.ID, outsider, nil); !errors.Is(err, ErrForbidden) {
		t.Errorf("not assigned: %v", err)
	}
	noErr(t, s.SetOpen(test.ID, false))
	if _, err := s.HandIn(test.ID, carla, map[int]string{0: "dog"}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed test: %v", err)
	}

	// the teacher counts a typo as right, and publishes
	r = must(s.Override(test.ID, bram.ID, 0, true))
	if r.Note != 50 || !r.Items[0].Overridden {
		t.Errorf("override %+v", r)
	}
	if _, err := s.Override(test.ID, bram.ID, 99, true); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown item: %v", err)
	}
	if got := must(s.ResultOf(test.ID, bram.ID)); got.Published || got.Note != 50 {
		t.Errorf("before publishing %+v", got)
	}
	noErr(t, s.Publish(test.ID, 0, true))
	results := must(s.Results(test.ID))
	if len(results) != 2 || !results[0].Published || results[0].Student.Name != "anna" {
		t.Errorf("results %+v", results)
	}
	if list := must(s.TestsFor(bram)); !list[0].HandedIn || !list[0].Published {
		t.Errorf("bram's list %+v", list)
	}

	// removing a student removes their answers
	noErr(t, s.DeleteUser(bram.ID))
	if results := must(s.Results(test.ID)); len(results) != 1 {
		t.Errorf("after deleting bram %+v", results)
	}
	if _, err := s.ResultOf(test.ID, carla.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("carla did not hand in: %v", err)
	}
}

func TestReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s := must(Open(path))
	teacher := must(s.CreateUser("t", Teacher, "t"))
	must(s.CreateTest(teacher, animals))
	s.Close()
	s = must(Open(path)) // the schema is not made twice
	defer s.Close()
	if list := must(s.TestsFor(teacher)); len(list) != 1 || list[0].Title != "Animals" {
		t.Errorf("after reopening %+v", list)
	}
}
