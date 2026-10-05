// Package testserver is Recuerdo's test mode: classroom tests from a
// server. A teacher makes a test from a word list and assigns it to
// students or groups; students hand in their answers, which the server
// checks with the same rule as practice (teaching.Correct); the teacher
// can count answers as right and publishes the results, which the
// students can then see.
//
// OpenTeacher had a test server (Django and SQLite) that never shipped;
// this keeps its model (users with roles, groups, tests, hand-ins,
// checked results) and stores it in SQLite.
package testserver

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"golang.org/x/crypto/bcrypt"
)

// Roles of users.
const (
	Admin   = "admin"   // manages users and groups
	Teacher = "teacher" // makes tests, checks and publishes results
	Student = "student" // takes tests
)

// Errors the API turns into status codes.
var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("not allowed")
	ErrExists    = errors.New("already exists")
	ErrLogin     = errors.New("wrong name or password")
	ErrClosed    = errors.New("the test is closed")
	ErrHandedIn  = errors.New("the answers were already handed in")
)

// SessionTime is how long a login lasts.
var SessionTime = 12 * time.Hour

// Store is the test server's database.
type Store struct {
	db  *sql.DB
	now func() time.Time
}

var schema = []string{
	`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE, role TEXT NOT NULL,
		password BLOB NOT NULL, created INTEGER NOT NULL)`,
	`CREATE TABLE sessions (token TEXT PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expires INTEGER NOT NULL)`,
	`CREATE TABLE groups (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE)`,
	`CREATE TABLE members (group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE, PRIMARY KEY (group_id, user_id))`,
	`CREATE TABLE tests (id INTEGER PRIMARY KEY, teacher_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title TEXT NOT NULL, list TEXT NOT NULL, open INTEGER NOT NULL DEFAULT 1, created INTEGER NOT NULL)`,
	`CREATE TABLE assigned_students (test_id INTEGER NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE, PRIMARY KEY (test_id, user_id))`,
	`CREATE TABLE assigned_groups (test_id INTEGER NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
		group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE, PRIMARY KEY (test_id, group_id))`,
	`CREATE TABLE answers (test_id INTEGER NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
		student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE, given TEXT NOT NULL,
		checked TEXT NOT NULL, note INTEGER NOT NULL, published INTEGER NOT NULL DEFAULT 0,
		handed_in INTEGER NOT NULL, PRIMARY KEY (test_id, student_id))`,
}

// Open opens (or makes) the database at path (":memory:" for tests).
func Open(path string) (*Store, error) {
	dsn := "file:" + path + "?_foreign_keys=on&_busy_timeout=5000"
	if path == ":memory:" {
		dsn = "file::memory:?_foreign_keys=on"
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if path == ":memory:" {
		db.SetMaxOpenConns(1) // one shared in-memory database
	}
	s := &Store{db: db, now: time.Now}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	var version int
	if err := s.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	for i := version; i < len(schema); i++ {
		if _, err := s.db.Exec(schema[i]); err != nil {
			return fmt.Errorf("database schema %d: %w", i+1, err)
		}
		if _, err := s.db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, i+1)); err != nil {
			return err
		}
	}
	return nil
}

// ---- users and sessions ----

// User is an account.
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// CreateUser adds an account.
func (s *Store) CreateUser(name, role, password string) (User, error) {
	name = strings.TrimSpace(name)
	if name == "" || password == "" {
		return User{}, errors.New("a name and a password are needed")
	}
	if role != Admin && role != Teacher && role != Student {
		return User{}, fmt.Errorf("unknown role %q", role)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	res, err := s.db.Exec(`INSERT INTO users (name, role, password, created) VALUES (?, ?, ?, ?)`,
		name, role, hash, s.now().Unix())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return User{}, ErrExists
		}
		return User{}, err
	}
	id, _ := res.LastInsertId()
	return User{ID: id, Name: name, Role: role}, nil
}

// SetPassword changes a user's password (and ends their sessions).
func (s *Store) SetPassword(userID int64, password string) error {
	if password == "" {
		return errors.New("a password is needed")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.exec1(`UPDATE users SET password = ? WHERE id = ?`, hash, userID); err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

// DeleteUser removes an account with everything that is theirs.
func (s *Store) DeleteUser(id int64) error { return s.exec1(`DELETE FROM users WHERE id = ?`, id) }

// Users lists the accounts, by name.
func (s *Store) Users() ([]User, error) {
	return queryUsers(s.db, `SELECT id, name, role FROM users ORDER BY name`)
}

// UserByID finds an account.
func (s *Store) UserByID(id int64) (User, error) {
	var u User
	err := s.db.QueryRow(`SELECT id, name, role FROM users WHERE id = ?`, id).Scan(&u.ID, &u.Name, &u.Role)
	return u, notFound(err)
}

// Login checks a name and password and starts a session: its token.
func (s *Store) Login(name, password string) (string, User, error) {
	var u User
	var hash []byte
	err := s.db.QueryRow(`SELECT id, name, role, password FROM users WHERE name = ?`, strings.TrimSpace(name)).
		Scan(&u.ID, &u.Name, &u.Role, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		// as slow as a wrong password, so names cannot be found out
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return "", User{}, ErrLogin
	}
	if err != nil {
		return "", User{}, err
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return "", User{}, ErrLogin
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", User{}, err
	}
	token := hex.EncodeToString(b)
	_, err = s.db.Exec(`INSERT INTO sessions (token, user_id, expires) VALUES (?, ?, ?)`,
		token, u.ID, s.now().Add(SessionTime).Unix())
	return token, u, err
}

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("recuerdo"), bcrypt.DefaultCost)

// Session is the user of a session token that has not expired.
func (s *Store) Session(token string) (User, error) {
	var u User
	err := s.db.QueryRow(`SELECT u.id, u.name, u.role FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token = ? AND s.expires > ?`, token, s.now().Unix()).Scan(&u.ID, &u.Name, &u.Role)
	return u, notFound(err)
}

// Logout ends a session.
func (s *Store) Logout(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

// ---- groups ----

// Group is a class: a named set of students.
type Group struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Members []User `json:"members,omitempty"`
}

// CreateGroup adds a group.
func (s *Store) CreateGroup(name string) (Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Group{}, errors.New("a name is needed")
	}
	res, err := s.db.Exec(`INSERT INTO groups (name) VALUES (?)`, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return Group{}, ErrExists
		}
		return Group{}, err
	}
	id, _ := res.LastInsertId()
	return Group{ID: id, Name: name}, nil
}

// DeleteGroup removes a group (not its students).
func (s *Store) DeleteGroup(id int64) error { return s.exec1(`DELETE FROM groups WHERE id = ?`, id) }

// Groups lists the groups with their members.
func (s *Store) Groups() ([]Group, error) {
	rows, err := s.db.Query(`SELECT id, name FROM groups ORDER BY name`)
	if err != nil {
		return nil, err
	}
	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			rows.Close()
			return nil, err
		}
		groups = append(groups, g)
	}
	rows.Close()
	for i := range groups {
		if groups[i].Members, err = queryUsers(s.db, `SELECT u.id, u.name, u.role FROM members m
			JOIN users u ON u.id = m.user_id WHERE m.group_id = ? ORDER BY u.name`, groups[i].ID); err != nil {
			return nil, err
		}
	}
	return groups, rows.Err()
}

// AddMember puts a student in a group.
func (s *Store) AddMember(groupID, userID int64) error {
	u, err := s.UserByID(userID)
	if err != nil {
		return err
	}
	if u.Role != Student {
		return fmt.Errorf("%s is not a student", u.Name)
	}
	_, err = s.db.Exec(`INSERT OR IGNORE INTO members (group_id, user_id) VALUES (?, ?)`, groupID, userID)
	if err != nil && strings.Contains(err.Error(), "FOREIGN KEY") {
		return ErrNotFound
	}
	return err
}

// RemoveMember takes a student out of a group.
func (s *Store) RemoveMember(groupID, userID int64) error {
	return s.exec1(`DELETE FROM members WHERE group_id = ? AND user_id = ?`, groupID, userID)
}

// ---- helpers ----

func (s *Store) exec1(query string, args ...any) error {
	res, err := s.db.Exec(query, args...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func queryUsers(db *sql.DB, query string, args ...any) ([]User, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
