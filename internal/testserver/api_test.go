package testserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
)

type client struct {
	t     *testing.T
	base  string
	http  *http.Client
	token string // bearer token (the desktop); otherwise the cookie (the web)
}

func (c *client) do(method, path string, body any) (int, map[string]any, []any) {
	c.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.base+path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var obj map[string]any
	var list []any
	if json.Unmarshal(data, &obj) != nil {
		json.Unmarshal(data, &list)
	}
	return resp.StatusCode, obj, list
}

func (c *client) want(status int, method, path string, body any) (map[string]any, []any) {
	c.t.Helper()
	got, obj, list := c.do(method, path, body)
	if got != status {
		c.t.Fatalf("%s %s: %d %v, want %d", method, path, got, obj, status)
	}
	return obj, list
}

func login(t *testing.T, base, name, password string, bearer bool) *client {
	jar, _ := cookiejar.New(nil)
	c := &client{t: t, base: base, http: &http.Client{Jar: jar}}
	obj, _ := c.want(200, "POST", "/api/login", map[string]string{"name": name, "password": password})
	if bearer {
		c.token = obj["token"].(string)
		c.http.Jar = nil
	}
	return c
}

func id(obj map[string]any) int64 { return int64(obj["id"].(float64)) }

func TestAPI(t *testing.T) {
	s := newStore(t)
	s.CreateUser("admin", Admin, "a")
	srv := httptest.NewServer((&API{Store: s, Version: "test"}).Handler(http.NotFoundHandler()))
	defer srv.Close()

	anon := &client{t: t, base: srv.URL, http: http.DefaultClient}
	if obj, _ := anon.want(200, "GET", "/api/info", nil); obj["name"] != "Recuerdo test server" {
		t.Errorf("info %v", obj)
	}
	anon.want(401, "GET", "/api/me", nil)
	anon.want(401, "GET", "/api/users", nil)
	anon.want(401, "POST", "/api/login", map[string]string{"name": "admin", "password": "wrong"})

	// the admin makes accounts and a class (the web client: a cookie)
	admin := login(t, srv.URL, "admin", "a", false)
	if obj, _ := admin.want(200, "GET", "/api/me", nil); obj["role"] != Admin {
		t.Errorf("me %v", obj)
	}
	teacher, _ := admin.want(201, "POST", "/api/users", map[string]string{"name": "jansen", "role": Teacher, "password": "t"})
	other, _ := admin.want(201, "POST", "/api/users", map[string]string{"name": "pietersen", "role": Teacher, "password": "p"})
	anna, _ := admin.want(201, "POST", "/api/users", map[string]string{"name": "anna", "role": Student, "password": "s"})
	bram, _ := admin.want(201, "POST", "/api/users", map[string]string{"name": "bram", "role": Student, "password": "s"})
	admin.want(409, "POST", "/api/users", map[string]string{"name": "anna", "role": Student, "password": "x"})
	class, _ := admin.want(201, "POST", "/api/groups", map[string]string{"name": "2B"})
	admin.want(200, "POST", fmt.Sprintf("/api/groups/%d/members", id(class)), map[string]int64{"userId": id(anna)})
	admin.want(409, "DELETE", fmt.Sprintf("/api/users/%d", 1), nil) // not yourself
	_ = other

	// the teacher (the desktop: a bearer token) makes and assigns a test
	tc := login(t, srv.URL, "jansen", "t", true)
	tc.want(403, "POST", "/api/users", map[string]string{"name": "x", "role": Student, "password": "x"})
	if _, list := tc.want(200, "GET", "/api/users", nil); len(list) != 2 {
		t.Errorf("a teacher sees the students only: %v", list)
	}
	test, _ := tc.want(201, "POST", "/api/tests", map[string]any{"list": animals})
	tid := id(test)
	tc.want(200, "POST", fmt.Sprintf("/api/tests/%d/groups", tid), map[string]int64{"groupId": id(class)})
	tc.want(200, "POST", fmt.Sprintf("/api/tests/%d/students", tid), map[string]int64{"userId": id(bram)})
	tc.want(400, "POST", fmt.Sprintf("/api/tests/%d/students", tid), map[string]int64{"userId": id(teacher)})

	// another teacher cannot touch it
	pc := login(t, srv.URL, "pietersen", "p", true)
	pc.want(403, "GET", fmt.Sprintf("/api/tests/%d/results", tid), nil)
	pc.want(403, "DELETE", fmt.Sprintf("/api/tests/%d", tid), nil)
	pc.want(404, "GET", fmt.Sprintf("/api/tests/%d", tid), nil)

	// a student sees the questions, not the answers, and hands in once
	ac := login(t, srv.URL, "anna", "s", false)
	ac.want(403, "POST", "/api/tests", map[string]any{"list": animals})
	ac.want(403, "GET", "/api/users", nil)
	if _, list := ac.want(200, "GET", "/api/tests", nil); len(list) != 1 {
		t.Errorf("anna's tests %v", list)
	}
	got, _ := ac.want(200, "GET", fmt.Sprintf("/api/tests/%d", tid), nil)
	if strings.Contains(fmt.Sprint(got["list"]), "dog") {
		t.Errorf("a student got the answers: %v", got["list"])
	}
	ac.want(201, "POST", fmt.Sprintf("/api/tests/%d/answers", tid), map[string]any{"answers": map[string]string{"0": "dog", "1": "cta"}})
	ac.want(409, "POST", fmt.Sprintf("/api/tests/%d/answers", tid), map[string]any{"answers": map[string]string{"0": "dog"}})
	ac.want(404, "GET", fmt.Sprintf("/api/tests/%d/results/%d", tid, id(anna)), nil) // not published
	ac.want(404, "GET", fmt.Sprintf("/api/tests/%d/results/%d", tid, id(bram)), nil) // not hers

	// the teacher checks, counts a typo as right, publishes
	_, results := tc.want(200, "GET", fmt.Sprintf("/api/tests/%d/results", tid), nil)
	if len(results) != 1 || results[0].(map[string]any)["note"].(float64) != 25 {
		t.Errorf("results %v", results)
	}
	r, _ := tc.want(200, "PATCH", fmt.Sprintf("/api/tests/%d/results/%d", tid, id(anna)), map[string]any{"itemId": 1, "right": true})
	if r["note"].(float64) != 50 {
		t.Errorf("override %v", r)
	}
	tc.want(200, "POST", fmt.Sprintf("/api/tests/%d/publish", tid), map[string]bool{"published": true})
	if r, _ := ac.want(200, "GET", fmt.Sprintf("/api/tests/%d/results/%d", tid, id(anna)), nil); r["note"].(float64) != 50 {
		t.Errorf("anna's result %v", r)
	}
	tc.want(200, "PATCH", fmt.Sprintf("/api/tests/%d", tid), map[string]bool{"open": false})
	bc := login(t, srv.URL, "bram", "s", true)
	bc.want(409, "POST", fmt.Sprintf("/api/tests/%d/answers", tid), map[string]any{"answers": map[string]string{"0": "dog"}})

	// a form from another site cannot use the cookie: bodies must be JSON
	req, _ := http.NewRequest("POST", srv.URL+fmt.Sprintf("/api/tests/%d/publish", tid), strings.NewReader("published=false"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := ac.http.Do(req)
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("form post: %d", resp.StatusCode)
	}
	resp.Body.Close()

	// changing your own password; logging out ends the session
	ac.want(403, "PATCH", fmt.Sprintf("/api/users/%d", id(bram)), map[string]string{"password": "x"})
	ac.want(200, "PATCH", fmt.Sprintf("/api/users/%d", id(anna)), map[string]string{"password": "new"})
	ac.want(401, "GET", "/api/me", nil) // a new password ends the sessions
	ac = login(t, srv.URL, "anna", "new", false)
	ac.want(200, "POST", "/api/logout", nil)
	ac.want(401, "GET", "/api/me", nil)
	anon.want(404, "GET", "/api/nothing", nil)
}
