package testserver

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/teaching"
)

// Test is a word list a teacher gives to students.
type Test struct {
	ID        int64           `json:"id"`
	TeacherID int64           `json:"teacherId"`
	Title     string          `json:"title"`
	Open      bool            `json:"open"`
	Created   time.Time       `json:"created"`
	List      lesson.WordList `json:"list"`
	Students  []User          `json:"students"`
	Groups    []Group         `json:"groups"`
}

// CreateTest makes a test of a word list, owned by teacher.
func (s *Store) CreateTest(teacher User, list lesson.WordList) (Test, error) {
	if teacher.Role != Teacher {
		return Test{}, ErrForbidden
	}
	if len(list.Items) == 0 {
		return Test{}, errors.New("the list has no words")
	}
	list.Tests = nil // the teacher's own practice results stay home
	data, err := json.Marshal(list)
	if err != nil {
		return Test{}, err
	}
	title := strings.TrimSpace(list.Title)
	if title == "" {
		title = "Test"
	}
	now := s.now()
	res, err := s.db.Exec(`INSERT INTO tests (teacher_id, title, list, open, created) VALUES (?, ?, ?, 1, ?)`,
		teacher.ID, title, string(data), now.Unix())
	if err != nil {
		return Test{}, err
	}
	id, _ := res.LastInsertId()
	return Test{ID: id, TeacherID: teacher.ID, Title: title, Open: true, Created: time.Unix(now.Unix(), 0), List: list}, nil
}

// Test is a test with its list and who it is assigned to.
func (s *Store) Test(id int64) (Test, error) {
	var t Test
	var list string
	var open int
	var created int64
	err := s.db.QueryRow(`SELECT id, teacher_id, title, list, open, created FROM tests WHERE id = ?`, id).
		Scan(&t.ID, &t.TeacherID, &t.Title, &list, &open, &created)
	if err != nil {
		return Test{}, notFound(err)
	}
	t.Open, t.Created = open == 1, time.Unix(created, 0)
	if err := json.Unmarshal([]byte(list), &t.List); err != nil {
		return Test{}, err
	}
	if t.Students, err = queryUsers(s.db, `SELECT u.id, u.name, u.role FROM assigned_students a
		JOIN users u ON u.id = a.user_id WHERE a.test_id = ? ORDER BY u.name`, id); err != nil {
		return Test{}, err
	}
	rows, err := s.db.Query(`SELECT g.id, g.name FROM assigned_groups a JOIN groups g ON g.id = a.group_id
		WHERE a.test_id = ? ORDER BY g.name`, id)
	if err != nil {
		return Test{}, err
	}
	defer rows.Close()
	t.Groups = []Group{}
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return Test{}, err
		}
		t.Groups = append(t.Groups, g)
	}
	return t, rows.Err()
}

// ForStudent is the test as a student gets it: the questions, not the
// answers.
func (t Test) ForStudent() Test {
	items := make([]lesson.WordItem, len(t.List.Items))
	for i, it := range t.List.Items {
		items[i] = lesson.WordItem{ID: it.ID, Questions: it.Questions, Comment: it.Comment}
	}
	t.List = lesson.WordList{Title: t.List.Title, QuestionLanguage: t.List.QuestionLanguage,
		AnswerLanguage: t.List.AnswerLanguage, Items: items}
	t.Students, t.Groups = nil, nil
	return t
}

// TestSummary is a test in a list of tests.
type TestSummary struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	TeacherID int64     `json:"teacherId"`
	Open      bool      `json:"open"`
	Created   time.Time `json:"created"`
	Words     int       `json:"words"`
	HandedIn  bool      `json:"handedIn"`  // students: whether they handed it in
	Published bool      `json:"published"` // students: whether their result is out
}

// TestsFor lists the tests a user sees: a teacher their own, a student
// those assigned to them (directly or through a group), an admin all.
func (s *Store) TestsFor(u User) ([]TestSummary, error) {
	query := `SELECT t.id, t.title, t.teacher_id, t.open, t.created, t.list,
		a.student_id IS NOT NULL, COALESCE(a.published, 0) FROM tests t
		LEFT JOIN answers a ON a.test_id = t.id AND a.student_id = ?1`
	switch u.Role {
	case Teacher:
		query += ` WHERE t.teacher_id = ?1`
	case Student:
		query += ` WHERE t.id IN (SELECT test_id FROM assigned_students WHERE user_id = ?1
			UNION SELECT a.test_id FROM assigned_groups a JOIN members m ON m.group_id = a.group_id WHERE m.user_id = ?1)`
	}
	rows, err := s.db.Query(query+` ORDER BY t.created DESC, t.id DESC`, u.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tests := []TestSummary{}
	for rows.Next() {
		var t TestSummary
		var open, handedIn, published int
		var created int64
		var list string
		if err := rows.Scan(&t.ID, &t.Title, &t.TeacherID, &open, &created, &list, &handedIn, &published); err != nil {
			return nil, err
		}
		var l lesson.WordList
		json.Unmarshal([]byte(list), &l)
		t.Open, t.Created, t.Words = open == 1, time.Unix(created, 0), len(l.Items)
		t.HandedIn, t.Published = handedIn == 1, published == 1
		tests = append(tests, t)
	}
	return tests, rows.Err()
}

// Assigned reports whether a student may take a test.
func (s *Store) Assigned(testID, studentID int64) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM (SELECT 1 FROM assigned_students WHERE test_id = ?1 AND user_id = ?2
		UNION ALL SELECT 1 FROM assigned_groups a JOIN members m ON m.group_id = a.group_id
		WHERE a.test_id = ?1 AND m.user_id = ?2)`, testID, studentID).Scan(&n)
	return n > 0, err
}

// AssignStudent gives a test to a student.
func (s *Store) AssignStudent(testID, studentID int64) error {
	u, err := s.UserByID(studentID)
	if err != nil {
		return err
	}
	if u.Role != Student {
		return errors.New(u.Name + " is not a student")
	}
	return s.insertRef(`INSERT OR IGNORE INTO assigned_students (test_id, user_id) VALUES (?, ?)`, testID, studentID)
}

// UnassignStudent takes a test away from a student.
func (s *Store) UnassignStudent(testID, studentID int64) error {
	return s.exec1(`DELETE FROM assigned_students WHERE test_id = ? AND user_id = ?`, testID, studentID)
}

// AssignGroup gives a test to everyone in a group.
func (s *Store) AssignGroup(testID, groupID int64) error {
	return s.insertRef(`INSERT OR IGNORE INTO assigned_groups (test_id, group_id) VALUES (?, ?)`, testID, groupID)
}

// UnassignGroup takes a test away from a group.
func (s *Store) UnassignGroup(testID, groupID int64) error {
	return s.exec1(`DELETE FROM assigned_groups WHERE test_id = ? AND group_id = ?`, testID, groupID)
}

// SetOpen opens or closes a test for handing in.
func (s *Store) SetOpen(testID int64, open bool) error {
	v := 0
	if open {
		v = 1
	}
	return s.exec1(`UPDATE tests SET open = ? WHERE id = ?`, v, testID)
}

// DeleteTest removes a test with its answers.
func (s *Store) DeleteTest(id int64) error { return s.exec1(`DELETE FROM tests WHERE id = ?`, id) }

// ---- answers and results ----

// CheckedItem is one answer of a student, checked.
type CheckedItem struct {
	ItemID     int    `json:"itemId"`
	Given      string `json:"given"`
	Right      bool   `json:"right"`
	Overridden bool   `json:"overridden,omitempty"` // counted as right by the teacher
}

// Result is a student's handed-in test.
type Result struct {
	TestID    int64         `json:"testId"`
	Student   User          `json:"student"`
	Items     []CheckedItem `json:"items"`
	Note      int           `json:"note"` // percentage right
	Published bool          `json:"published"`
	HandedIn  time.Time     `json:"handedIn"`
}

// HandIn stores a student's answers (item ID to the typed answer) and
// checks them with Recuerdo's rule (OpenTeacher's notation, formulas,
// markup). Unanswered items count as wrong. A test is handed in once.
func (s *Store) HandIn(testID int64, student User, given map[int]string) (Result, error) {
	if student.Role != Student {
		return Result{}, ErrForbidden
	}
	t, err := s.Test(testID)
	if err != nil {
		return Result{}, err
	}
	if ok, err := s.Assigned(testID, student.ID); err != nil || !ok {
		if err == nil {
			err = ErrForbidden
		}
		return Result{}, err
	}
	if !t.Open {
		return Result{}, ErrClosed
	}
	r := Result{TestID: testID, Student: student, HandedIn: time.Unix(s.now().Unix(), 0)}
	for _, it := range t.List.Items {
		g := strings.TrimSpace(given[it.ID])
		r.Items = append(r.Items, CheckedItem{ItemID: it.ID, Given: g,
			Right: g != "" && teaching.Correct(g, it.Answers, false)})
	}
	r.Note = note(r.Items)
	givenJSON, _ := json.Marshal(given)
	checked, _ := json.Marshal(r.Items)
	_, err = s.db.Exec(`INSERT INTO answers (test_id, student_id, given, checked, note, published, handed_in)
		VALUES (?, ?, ?, ?, ?, 0, ?)`, testID, student.ID, string(givenJSON), string(checked), r.Note, r.HandedIn.Unix())
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		return Result{}, ErrHandedIn
	}
	return r, err
}

func note(items []CheckedItem) int {
	if len(items) == 0 {
		return 0
	}
	right := 0
	for _, it := range items {
		if it.Right {
			right++
		}
	}
	return right * 100 / len(items)
}

// Results are all hand-ins of a test, by student name.
func (s *Store) Results(testID int64) ([]Result, error) {
	rows, err := s.db.Query(`SELECT a.student_id, u.name, u.role, a.checked, a.note, a.published, a.handed_in
		FROM answers a JOIN users u ON u.id = a.student_id WHERE a.test_id = ? ORDER BY u.name`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := []Result{}
	for rows.Next() {
		r, err := scanResult(rows, testID)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// ResultOf is one student's hand-in of a test.
func (s *Store) ResultOf(testID, studentID int64) (Result, error) {
	row := s.db.QueryRow(`SELECT a.student_id, u.name, u.role, a.checked, a.note, a.published, a.handed_in
		FROM answers a JOIN users u ON u.id = a.student_id WHERE a.test_id = ? AND a.student_id = ?`, testID, studentID)
	r, err := scanResult(row, testID)
	return r, notFound(err)
}

type scanner interface{ Scan(...any) error }

func scanResult(row scanner, testID int64) (Result, error) {
	r := Result{TestID: testID}
	var checked string
	var published int
	var handedIn int64
	if err := row.Scan(&r.Student.ID, &r.Student.Name, &r.Student.Role, &checked, &r.Note, &published, &handedIn); err != nil {
		return Result{}, err
	}
	r.Published, r.HandedIn = published == 1, time.Unix(handedIn, 0)
	return r, json.Unmarshal([]byte(checked), &r.Items)
}

// Override counts an answer as right (or not) after all, as the teacher
// decides, and recomputes the note.
func (s *Store) Override(testID, studentID int64, itemID int, right bool) (Result, error) {
	r, err := s.ResultOf(testID, studentID)
	if err != nil {
		return Result{}, err
	}
	found := false
	for i := range r.Items {
		if r.Items[i].ItemID == itemID {
			r.Items[i].Right, r.Items[i].Overridden, found = right, true, true
		}
	}
	if !found {
		return Result{}, ErrNotFound
	}
	r.Note = note(r.Items)
	checked, _ := json.Marshal(r.Items)
	return r, s.exec1(`UPDATE answers SET checked = ?, note = ? WHERE test_id = ? AND student_id = ?`,
		string(checked), r.Note, testID, studentID)
}

// Publish makes results visible to the students: one student's, or
// everyone's who handed in (studentID 0).
func (s *Store) Publish(testID, studentID int64, published bool) error {
	v := 0
	if published {
		v = 1
	}
	if studentID == 0 {
		_, err := s.db.Exec(`UPDATE answers SET published = ? WHERE test_id = ?`, v, testID)
		return err
	}
	return s.exec1(`UPDATE answers SET published = ? WHERE test_id = ? AND student_id = ?`, v, testID, studentID)
}

func (s *Store) insertRef(query string, args ...any) error {
	_, err := s.db.Exec(query, args...)
	if err != nil && strings.Contains(err.Error(), "FOREIGN KEY") {
		return ErrNotFound
	}
	return err
}
