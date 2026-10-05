package testserver

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient(t *testing.T) {
	s := newStore(t)
	s.CreateUser("jansen", Teacher, "t")
	s.CreateUser("anna", Student, "s")
	cert, key, fp, err := Certificate(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	pair, _ := tls.LoadX509KeyPair(cert, key)
	srv := httptest.NewUnstartedServer((&API{Store: s, Version: "test", Secure: true}).Handler(nil))
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{pair}}
	srv.StartTLS()
	defer srv.Close()
	addr := strings.TrimPrefix(srv.URL, "https://")

	// first contact: the fingerprint to compare (trust on first use)
	_, err = Connect(addr, "")
	var unknown *UnknownCertificate
	if !errors.As(err, &unknown) || unknown.Fingerprint != fp {
		t.Fatalf("first connect: %v (want fingerprint %s)", err, fp)
	}
	if _, err := Connect(srv.URL, strings.Replace(fp, fp[:2], "00", 1)); !errors.As(err, &unknown) {
		t.Fatalf("a different certificate: %v", err)
	}
	teacher, err := Connect(addr, strings.ToLower(fp))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := teacher.Login("jansen", "wrong"); err == nil {
		t.Error("wrong password")
	}
	if u, err := teacher.Login("jansen", "t"); err != nil || u.Role != Teacher {
		t.Fatalf("login %+v %v", u, err)
	}
	test, err := teacher.CreateTest(animals)
	if err != nil {
		t.Fatal(err)
	}
	users, _ := teacher.Users()
	if len(users) != 1 || teacher.AssignStudent(test.ID, users[0].ID) != nil {
		t.Fatalf("students %+v", users)
	}

	student, _ := Connect(srv.URL, fp)
	student.Login("anna", "s")
	st, err := student.StudentTest(test.ID)
	if err != nil || len(st.Items) != 4 || !st.Items[2].Math {
		t.Fatalf("student's test %+v %v", st, err)
	}
	if err := student.HandIn(test.ID, map[int]string{0: "dog", 1: "cat", 2: "\\pi r^2", 3: "H2O"}); err != nil {
		t.Fatal(err)
	}
	var apiErr *APIError
	if err := student.HandIn(test.ID, map[int]string{}); !errors.As(err, &apiErr) || apiErr.Status != http.StatusConflict {
		t.Errorf("handing in twice: %v", err)
	}
	results, _ := teacher.Results(test.ID)
	if len(results) != 1 || results[0].Note != 100 {
		t.Fatalf("results %+v", results)
	}
	if _, err := student.Result(test.ID, users[0].ID); err == nil {
		t.Error("a result before publishing")
	}
	teacher.Publish(test.ID, 0)
	if r, err := student.Result(test.ID, users[0].ID); err != nil || r.Note != 100 {
		t.Errorf("published result %+v %v", r, err)
	}
	student.Logout()
	if _, err := student.Tests(); err == nil {
		t.Error("calls after logging out")
	}
}

func TestClientPlainHTTP(t *testing.T) {
	s := newStore(t)
	srv := httptest.NewServer((&API{Store: s}).Handler(nil))
	defer srv.Close()
	c, err := Connect(strings.TrimPrefix(srv.URL, "http://"), "")
	if err != nil || !strings.HasPrefix(c.Address(), "http://") {
		t.Fatalf("plain http: %v", err)
	}
	other := httptest.NewServer(http.NotFoundHandler())
	defer other.Close()
	if _, err := Connect(other.URL, ""); err == nil {
		t.Error("not a test server, but connected")
	}
}
