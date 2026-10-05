package testserver

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// UnknownCertificate is returned when a server's certificate is not the
// one trusted for it (or none is trusted yet): its fingerprint, for the
// user to compare with the one the server printed (trust on first use).
type UnknownCertificate struct{ Fingerprint string }

func (e *UnknownCertificate) Error() string {
	return "the server's certificate is not known: fingerprint " + e.Fingerprint
}

// Client talks to a test server (the desktop's test mode).
type Client struct {
	base  *url.URL
	http  *http.Client
	token string
	// Fingerprint is the server's certificate fingerprint (https).
	Fingerprint string
}

// Connect opens a client for a test server address ("host:port",
// "http://..." or "https://..."; without a scheme https is tried first,
// then http). For https, pinned is the certificate fingerprint trusted
// for this server; an empty or different one gives an
// *UnknownCertificate error.
func Connect(address, pinned string) (*Client, error) {
	address = strings.TrimSpace(address)
	if !strings.Contains(address, "://") {
		c, err := Connect("https://"+address, pinned)
		var unknown *UnknownCertificate
		if err == nil || errors.As(err, &unknown) {
			return c, err
		}
		if c, err2 := Connect("http://"+address, ""); err2 == nil {
			return c, nil
		}
		return nil, err
	}
	u, err := url.Parse(strings.TrimRight(address, "/"))
	if err != nil {
		return nil, err
	}
	c := &Client{base: u}
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, TLSHandshakeTimeout: 10 * time.Second}
	if u.Scheme == "https" {
		// the server's own certificate is checked by its fingerprint, not
		// by a certificate authority (it signs itself)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12,
			VerifyConnection: func(cs tls.ConnectionState) error {
				if len(cs.PeerCertificates) == 0 {
					return errors.New("no certificate")
				}
				c.Fingerprint = Fingerprint(cs.PeerCertificates[0].Raw)
				if !strings.EqualFold(c.Fingerprint, pinned) {
					return &UnknownCertificate{Fingerprint: c.Fingerprint}
				}
				return nil
			}}
	}
	c.http = &http.Client{Transport: transport, Timeout: 30 * time.Second}
	var info map[string]string
	if err := c.call("GET", "info", nil, &info); err != nil {
		var unknown *UnknownCertificate
		if errors.As(err, &unknown) {
			return nil, unknown
		}
		return nil, err
	}
	if info["name"] != "Recuerdo test server" {
		return nil, fmt.Errorf("%s is not a Recuerdo test server", address)
	}
	return c, nil
}

// Address is the server's address.
func (c *Client) Address() string { return c.base.String() }

// APIError is an error the server answered with.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string { return e.Message }

func (c *Client) call(method, path string, body, out any) error {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.base.String()+"/api/"+path, rd)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		var unknown *UnknownCertificate
		if errors.As(err, &unknown) {
			return unknown
		}
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var e struct{ Error string }
		json.Unmarshal(data, &e)
		if e.Error == "" {
			e.Error = resp.Status
		}
		return &APIError{Status: resp.StatusCode, Message: e.Error}
	}
	if out != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

// Login logs in; later calls are made as this user.
func (c *Client) Login(name, password string) (User, error) {
	var r struct {
		Token string
		User  User
	}
	if err := c.call("POST", "login", map[string]string{"name": name, "password": password}, &r); err != nil {
		return User{}, err
	}
	c.token = r.Token
	return r.User, nil
}

// Logout ends the session.
func (c *Client) Logout() error {
	err := c.call("POST", "logout", nil, nil)
	c.token = ""
	return err
}

// Tests lists the user's tests.
func (c *Client) Tests() ([]TestSummary, error) {
	var t []TestSummary
	return t, c.call("GET", "tests", nil, &t)
}

// StudentTest is a test as an assigned student gets it.
func (c *Client) StudentTest(id int64) (StudentTest, error) {
	var t StudentTest
	return t, c.call("GET", fmt.Sprint("tests/", id), nil, &t)
}

// Test is a test as its teacher gets it (with the answers).
func (c *Client) Test(id int64) (Test, error) {
	var t Test
	return t, c.call("GET", fmt.Sprint("tests/", id), nil, &t)
}

// HandIn hands in a student's answers (item ID to the typed answer).
func (c *Client) HandIn(id int64, answers map[int]string) error {
	return c.call("POST", fmt.Sprint("tests/", id, "/answers"), map[string]any{"answers": answers}, nil)
}

// Result is a student's result (their own, once published; any for the
// test's teacher).
func (c *Client) Result(testID, studentID int64) (Result, error) {
	var r Result
	return r, c.call("GET", fmt.Sprint("tests/", testID, "/results/", studentID), nil, &r)
}

// CreateTest makes a test of a word list (teachers).
func (c *Client) CreateTest(list lesson.WordList) (Test, error) {
	var t Test
	return t, c.call("POST", "tests", map[string]any{"list": list}, &t)
}

// Results are a test's hand-ins (its teacher).
func (c *Client) Results(testID int64) ([]Result, error) {
	var r []Result
	return r, c.call("GET", fmt.Sprint("tests/", testID, "/results"), nil, &r)
}

// Override counts an answer as right or wrong after all.
func (c *Client) Override(testID, studentID int64, itemID int, right bool) (Result, error) {
	var r Result
	return r, c.call("PATCH", fmt.Sprint("tests/", testID, "/results/", studentID),
		map[string]any{"itemId": itemID, "right": right}, &r)
}

// Publish shows results to the students: one (studentID) or all (0).
func (c *Client) Publish(testID, studentID int64) error {
	return c.call("POST", fmt.Sprint("tests/", testID, "/publish"), map[string]any{"studentId": studentID, "published": true}, nil)
}

// SetOpen opens or closes a test for handing in.
func (c *Client) SetOpen(testID int64, open bool) error {
	return c.call("PATCH", fmt.Sprint("tests/", testID), map[string]bool{"open": open}, nil)
}

// Users are the accounts a teacher (students) or admin (all) sees.
func (c *Client) Users() ([]User, error) {
	var u []User
	return u, c.call("GET", "users", nil, &u)
}

// Groups are the groups with their members.
func (c *Client) Groups() ([]Group, error) {
	var g []Group
	return g, c.call("GET", "groups", nil, &g)
}

// AssignStudent gives a test to a student; AssignGroup to a group.
func (c *Client) AssignStudent(testID, userID int64) error {
	return c.call("POST", fmt.Sprint("tests/", testID, "/students"), map[string]int64{"userId": userID}, nil)
}

// AssignGroup gives a test to everyone in a group.
func (c *Client) AssignGroup(testID, groupID int64) error {
	return c.call("POST", fmt.Sprint("tests/", testID, "/groups"), map[string]int64{"groupId": groupID}, nil)
}
