package testserver

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// CookieName is the session cookie of the web client.
const CookieName = "recuerdo_session"

// maxBody limits request bodies (a word list with its pictures as data
// URLs can be large).
const maxBody = 8 << 20

// API is the test server's HTTP interface (JSON under /api/).
type API struct {
	Store   *Store
	Version string
	// Secure marks the session cookie Secure (when served over TLS).
	Secure bool
}

// Handler serves the API; other paths go to next (the web version), or
// are not found.
func (a *API) Handler(next http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/info", a.info)
	mux.HandleFunc("POST /api/login", a.login)
	mux.HandleFunc("POST /api/logout", a.logout)
	mux.HandleFunc("GET /api/me", a.auth(a.me))

	mux.HandleFunc("GET /api/users", a.auth(a.users, Admin, Teacher))
	mux.HandleFunc("POST /api/users", a.auth(a.createUser, Admin))
	mux.HandleFunc("PATCH /api/users/{id}", a.auth(a.updateUser))
	mux.HandleFunc("DELETE /api/users/{id}", a.auth(a.deleteUser, Admin))
	mux.HandleFunc("GET /api/groups", a.auth(a.groups, Admin, Teacher))
	mux.HandleFunc("POST /api/groups", a.auth(a.createGroup, Admin))
	mux.HandleFunc("DELETE /api/groups/{id}", a.auth(a.deleteGroup, Admin))
	mux.HandleFunc("POST /api/groups/{id}/members", a.auth(a.addMember, Admin))
	mux.HandleFunc("DELETE /api/groups/{id}/members/{user}", a.auth(a.removeMember, Admin))

	mux.HandleFunc("GET /api/tests", a.auth(a.tests))
	mux.HandleFunc("POST /api/tests", a.auth(a.createTest, Teacher))
	mux.HandleFunc("GET /api/tests/{id}", a.auth(a.test))
	mux.HandleFunc("PATCH /api/tests/{id}", a.auth(a.owned(a.updateTest)))
	mux.HandleFunc("DELETE /api/tests/{id}", a.auth(a.owned(a.deleteTest)))
	mux.HandleFunc("POST /api/tests/{id}/students", a.auth(a.owned(a.assignStudent)))
	mux.HandleFunc("DELETE /api/tests/{id}/students/{user}", a.auth(a.owned(a.unassignStudent)))
	mux.HandleFunc("POST /api/tests/{id}/groups", a.auth(a.owned(a.assignGroup)))
	mux.HandleFunc("DELETE /api/tests/{id}/groups/{group}", a.auth(a.owned(a.unassignGroup)))
	mux.HandleFunc("POST /api/tests/{id}/answers", a.auth(a.handIn, Student))
	mux.HandleFunc("GET /api/tests/{id}/results", a.auth(a.owned(a.results)))
	mux.HandleFunc("GET /api/tests/{id}/results/{user}", a.auth(a.result))
	mux.HandleFunc("PATCH /api/tests/{id}/results/{user}", a.auth(a.owned(a.override)))
	mux.HandleFunc("POST /api/tests/{id}/publish", a.auth(a.owned(a.publish)))
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) { fail(w, http.StatusNotFound, "no such call") })
	if next != nil {
		mux.Handle("/", next)
	}
	return mux
}

// ---- plumbing ----

type handler func(w http.ResponseWriter, r *http.Request, u User)

// auth lets a request through with a valid session (cookie or bearer
// token) and, when roles are given, one of those roles. A request with
// a body must be JSON: a form on another site cannot send that with the
// session cookie.
func (a *API) auth(h handler, roles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength != 0 && r.Method != http.MethodGet {
			if mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mt != "application/json" {
				fail(w, http.StatusUnsupportedMediaType, "send JSON")
				return
			}
		}
		u, err := a.Store.Session(token(r))
		if err != nil {
			fail(w, http.StatusUnauthorized, "log in first")
			return
		}
		if len(roles) > 0 && !contains(roles, u.Role) {
			fail(w, http.StatusForbidden, "not allowed for a "+u.Role)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		h(w, r, u)
	}
}

// owned lets through the teacher who made the test {id}, and admins.
func (a *API) owned(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request, u User) {
		t, err := a.Store.Test(pathID(r, "id"))
		if err != nil {
			a.error(w, err)
			return
		}
		if u.Role != Admin && !(u.Role == Teacher && t.TeacherID == u.ID) {
			fail(w, http.StatusForbidden, "not your test")
			return
		}
		h(w, r, u)
	}
}

func token(r *http.Request) string {
	if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	if c, err := r.Cookie(CookieName); err == nil {
		return c.Value
	}
	return ""
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func pathID(r *http.Request, name string) int64 {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil {
		return -1
	}
	return id
}

func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, msg string) {
	reply(w, status, map[string]string{"error": msg})
}

func (a *API) error(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		fail(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrForbidden):
		fail(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrExists), errors.Is(err, ErrHandedIn), errors.Is(err, ErrClosed):
		fail(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrLogin):
		fail(w, http.StatusUnauthorized, err.Error())
	default:
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			fail(w, http.StatusRequestEntityTooLarge, "too large")
			return
		}
		fail(w, http.StatusBadRequest, err.Error())
	}
}

func decode(r *http.Request, v any) error {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	return d.Decode(v)
}

// ---- sessions ----

func (a *API) info(w http.ResponseWriter, r *http.Request) {
	reply(w, http.StatusOK, map[string]string{"name": "Recuerdo test server", "version": a.Version})
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	if mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mt != "application/json" {
		fail(w, http.StatusUnsupportedMediaType, "send JSON")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	var in struct{ Name, Password string }
	if err := decode(r, &in); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	tok, u, err := a.Store.Login(in.Name, in.Password)
	if err != nil {
		a.error(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: tok, Path: "/", HttpOnly: true, Secure: a.Secure,
		SameSite: http.SameSiteStrictMode, Expires: time.Now().Add(SessionTime)})
	reply(w, http.StatusOK, map[string]any{"token": tok, "user": u})
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	a.Store.Logout(token(r))
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: a.Secure})
	reply(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) me(w http.ResponseWriter, r *http.Request, u User) { reply(w, http.StatusOK, u) }

// ---- users and groups ----

func (a *API) users(w http.ResponseWriter, r *http.Request, u User) {
	users, err := a.Store.Users()
	if err != nil {
		a.error(w, err)
		return
	}
	if u.Role == Teacher { // teachers see the students, to give them tests
		students := []User{}
		for _, x := range users {
			if x.Role == Student {
				students = append(students, x)
			}
		}
		users = students
	}
	reply(w, http.StatusOK, users)
}

func (a *API) createUser(w http.ResponseWriter, r *http.Request, _ User) {
	var in struct{ Name, Role, Password string }
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	u, err := a.Store.CreateUser(in.Name, in.Role, in.Password)
	if err != nil {
		a.error(w, err)
		return
	}
	reply(w, http.StatusCreated, u)
}

// updateUser changes a password: an admin anyone's, others their own.
func (a *API) updateUser(w http.ResponseWriter, r *http.Request, u User) {
	id := pathID(r, "id")
	if u.Role != Admin && id != u.ID {
		fail(w, http.StatusForbidden, "not your account")
		return
	}
	var in struct{ Password string }
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	if err := a.Store.SetPassword(id, in.Password); err != nil {
		a.error(w, err)
		return
	}
	reply(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) deleteUser(w http.ResponseWriter, r *http.Request, u User) {
	if pathID(r, "id") == u.ID {
		fail(w, http.StatusConflict, "you cannot remove yourself")
		return
	}
	a.done(w, a.Store.DeleteUser(pathID(r, "id")))
}

func (a *API) groups(w http.ResponseWriter, r *http.Request, _ User) {
	groups, err := a.Store.Groups()
	if err != nil {
		a.error(w, err)
		return
	}
	if groups == nil {
		groups = []Group{}
	}
	reply(w, http.StatusOK, groups)
}

func (a *API) createGroup(w http.ResponseWriter, r *http.Request, _ User) {
	var in struct{ Name string }
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	g, err := a.Store.CreateGroup(in.Name)
	if err != nil {
		a.error(w, err)
		return
	}
	reply(w, http.StatusCreated, g)
}

func (a *API) deleteGroup(w http.ResponseWriter, r *http.Request, _ User) {
	a.done(w, a.Store.DeleteGroup(pathID(r, "id")))
}

func (a *API) addMember(w http.ResponseWriter, r *http.Request, _ User) {
	var in struct{ UserID int64 }
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	a.done(w, a.Store.AddMember(pathID(r, "id"), in.UserID))
}

func (a *API) removeMember(w http.ResponseWriter, r *http.Request, _ User) {
	a.done(w, a.Store.RemoveMember(pathID(r, "id"), pathID(r, "user")))
}

func (a *API) done(w http.ResponseWriter, err error) {
	if err != nil {
		a.error(w, err)
		return
	}
	reply(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---- tests ----

func (a *API) tests(w http.ResponseWriter, r *http.Request, u User) {
	tests, err := a.Store.TestsFor(u)
	if err != nil {
		a.error(w, err)
		return
	}
	reply(w, http.StatusOK, tests)
}

func (a *API) createTest(w http.ResponseWriter, r *http.Request, u User) {
	var in struct{ List lesson.WordList }
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	t, err := a.Store.CreateTest(u, in.List)
	if err != nil {
		a.error(w, err)
		return
	}
	reply(w, http.StatusCreated, t)
}

// test gives the teacher (and admins) the whole test, an assigned
// student only the questions.
func (a *API) test(w http.ResponseWriter, r *http.Request, u User) {
	t, err := a.Store.Test(pathID(r, "id"))
	if err != nil {
		a.error(w, err)
		return
	}
	switch {
	case u.Role == Admin || (u.Role == Teacher && t.TeacherID == u.ID):
		reply(w, http.StatusOK, t)
	case u.Role == Student:
		if ok, _ := a.Store.Assigned(t.ID, u.ID); ok {
			reply(w, http.StatusOK, a.Store.StudentView(t, u.ID))
			return
		}
		fallthrough
	default:
		fail(w, http.StatusNotFound, ErrNotFound.Error())
	}
}

func (a *API) updateTest(w http.ResponseWriter, r *http.Request, _ User) {
	var in struct{ Open *bool }
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	if in.Open == nil {
		fail(w, http.StatusBadRequest, "nothing to change")
		return
	}
	a.done(w, a.Store.SetOpen(pathID(r, "id"), *in.Open))
}

func (a *API) deleteTest(w http.ResponseWriter, r *http.Request, _ User) {
	a.done(w, a.Store.DeleteTest(pathID(r, "id")))
}

func (a *API) assignStudent(w http.ResponseWriter, r *http.Request, _ User) {
	var in struct{ UserID int64 }
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	a.done(w, a.Store.AssignStudent(pathID(r, "id"), in.UserID))
}

func (a *API) unassignStudent(w http.ResponseWriter, r *http.Request, _ User) {
	a.done(w, a.Store.UnassignStudent(pathID(r, "id"), pathID(r, "user")))
}

func (a *API) assignGroup(w http.ResponseWriter, r *http.Request, _ User) {
	var in struct{ GroupID int64 }
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	a.done(w, a.Store.AssignGroup(pathID(r, "id"), in.GroupID))
}

func (a *API) unassignGroup(w http.ResponseWriter, r *http.Request, _ User) {
	a.done(w, a.Store.UnassignGroup(pathID(r, "id"), pathID(r, "group")))
}

// handIn takes a student's answers ({"answers": {"0": "dog", ...}}, by
// item ID). The result is only shown once the teacher publishes it.
func (a *API) handIn(w http.ResponseWriter, r *http.Request, u User) {
	var in struct{ Answers map[int]string }
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	res, err := a.Store.HandIn(pathID(r, "id"), u, in.Answers)
	if err != nil {
		if errors.Is(err, ErrForbidden) { // not assigned: as if there is no such test
			err = ErrNotFound
		}
		a.error(w, err)
		return
	}
	reply(w, http.StatusCreated, map[string]any{"handedIn": res.HandedIn})
}

func (a *API) results(w http.ResponseWriter, r *http.Request, _ User) {
	results, err := a.Store.Results(pathID(r, "id"))
	if err != nil {
		a.error(w, err)
		return
	}
	reply(w, http.StatusOK, results)
}

// result is one student's result: for the test's teacher (and admins),
// or for the student once it is published.
func (a *API) result(w http.ResponseWriter, r *http.Request, u User) {
	testID, studentID := pathID(r, "id"), pathID(r, "user")
	t, err := a.Store.Test(testID)
	if err != nil {
		a.error(w, err)
		return
	}
	teacher := u.Role == Admin || (u.Role == Teacher && t.TeacherID == u.ID)
	if !teacher && u.ID != studentID {
		fail(w, http.StatusNotFound, ErrNotFound.Error())
		return
	}
	res, err := a.Store.ResultOf(testID, studentID)
	if err != nil {
		a.error(w, err)
		return
	}
	if !teacher && !res.Published {
		fail(w, http.StatusNotFound, "not published yet")
		return
	}
	reply(w, http.StatusOK, res)
}

func (a *API) override(w http.ResponseWriter, r *http.Request, _ User) {
	var in struct {
		ItemID int
		Right  bool
	}
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	res, err := a.Store.Override(pathID(r, "id"), pathID(r, "user"), in.ItemID, in.Right)
	if err != nil {
		a.error(w, err)
		return
	}
	reply(w, http.StatusOK, res)
}

// publish shows results to the students: {"published": true} for all,
// with "studentId" for one.
func (a *API) publish(w http.ResponseWriter, r *http.Request, _ User) {
	var in struct {
		StudentID int64
		Published *bool
	}
	if err := decode(r, &in); err != nil {
		a.error(w, err)
		return
	}
	on := in.Published == nil || *in.Published
	a.done(w, a.Store.Publish(pathID(r, "id"), in.StudentID, on))
}
