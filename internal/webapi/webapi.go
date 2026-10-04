// Package webapi is what Recuerdo's web version (cmd/recuerdo-web, Go
// compiled to WebAssembly) offers its page: open a lesson from a file's
// bytes, practise it with the same session as the desktop, and save it
// again. Everything goes in and out as JSON-friendly values, so it is
// tested here without a browser.
package webapi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/LaPingvino/recuerdo/internal/lesson"
	wordsreverser "github.com/LaPingvino/recuerdo/internal/modules/logic/reversers/words"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/checker"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/wordsString/composer"
	"github.com/LaPingvino/recuerdo/internal/richtext"
	"github.com/LaPingvino/recuerdo/internal/teaching"
)

// Item is a word pair as the page shows it.
type Item struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Comment  string `json:"comment,omitempty"`
	// the same as safe HTML, for words with markup (richtext)
	QuestionHTML string `json:"questionHtml"`
	AnswerHTML   string `json:"answerHtml"`
}

// Lesson is an opened lesson.
type Lesson struct {
	Title            string `json:"title"`
	QuestionLanguage string `json:"questionLanguage"`
	AnswerLanguage   string `json:"answerLanguage"`
	Items            []Item `json:"items"`
	Sessions         int    `json:"sessions"`
}

// Options start a practice session (empty values: the defaults).
type Options struct {
	LessonType string `json:"lessonType"`
	Order      string `json:"order"`
	Words      string `json:"words"`
	AskAnswers bool   `json:"askAnswers"`
}

// State is where a practice session is.
type State struct {
	Active   bool   `json:"active"`
	Done     bool   `json:"done"`
	Question string `json:"question,omitempty"`
	Asked    int    `json:"asked"`
	Total    int    `json:"total"`
	Right    int    `json:"right"`
	Answered int    `json:"answered"`
	// Answer is the right answer to the current question (Repeat answer
	// shows it first, In mind after "View answer"); Shuffle is Shuffle
	// answer's hint: its letters in another order.
	Answer  string `json:"answer,omitempty"`
	Shuffle string `json:"shuffle,omitempty"`
	// QuestionHTML and AnswerHTML show them with their markup (safe HTML)
	QuestionHTML string `json:"questionHtml,omitempty"`
	AnswerHTML   string `json:"answerHtml,omitempty"`
}

// Result is the outcome of an answer.
type Result struct {
	Right       bool   `json:"right"`
	Correct     string `json:"correct"`
	CorrectHTML string `json:"correctHtml"`
}

// Row is one answer of a finished session.
type Row struct {
	Question     string `json:"question"`
	Answer       string `json:"answer"`
	Given        string `json:"given"`
	Right        bool   `json:"right"`
	QuestionHTML string `json:"questionHtml"`
	AnswerHTML   string `json:"answerHtml"`
}

// Choices are the values the page offers for Options, with their labels
// in the current language.
type Choices struct {
	LessonTypes [][2]string `json:"lessonTypes"`
	Orders      [][2]string `json:"orders"`
	Words       [][2]string `json:"words"`
}

// App is the web version's state: one open lesson and its session.
type App struct {
	data    *lesson.LessonData
	name    string
	session *teaching.Session
	opts    Options
	// mathAnswer marks the items whose answer has a formula: typed answers
	// to them are compared without spaces
	mathAnswer map[int]bool
}

// ErrNoLesson is returned when there is no lesson to work on.
var ErrNoLesson = errors.New("open a lesson first")

// workDir is where files are put to load and save them (an in-memory file
// system in the browser).
func workDir() (string, error) { return os.MkdirTemp("", "recuerdo-web-") }

// Open opens a lesson file from its name (for the format) and contents.
func (a *App) Open(name string, data []byte) (Lesson, error) {
	dir, err := workDir()
	if err != nil {
		return Lesson{}, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, filepath.Base(name))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return Lesson{}, err
	}
	d, err := lesson.NewFileLoader().LoadFile(path)
	if err != nil {
		return Lesson{}, err
	}
	if len(d.List.Items) == 0 {
		return Lesson{}, fmt.Errorf("%s has no words", name)
	}
	a.data, a.name, a.session = d, name, nil
	return a.Lesson()
}

// OpenText makes a lesson of "question = answer" lines.
func (a *App) OpenText(title, text string) (Lesson, error) {
	items, err := lesson.ParseWordList(text, false)
	if err != nil {
		return Lesson{}, err
	}
	if len(items) == 0 {
		return Lesson{}, errors.New("type at least one word")
	}
	d := lesson.NewLessonData()
	d.List.Title, d.List.Items = title, items
	a.data, a.name, a.session = d, title+".otwd", nil
	return a.Lesson()
}

// Lesson is the open lesson.
func (a *App) Lesson() (Lesson, error) {
	if a.data == nil {
		return Lesson{}, ErrNoLesson
	}
	l := a.data.List
	out := Lesson{Title: l.Title, QuestionLanguage: l.QuestionLanguage, AnswerLanguage: l.AnswerLanguage,
		Items: []Item{}, Sessions: len(l.Tests)}
	for _, it := range l.Items {
		q, ans := compose(it.Questions), compose(it.Answers)
		out.Items = append(out.Items, Item{ID: it.ID, Question: q, Answer: ans, Comment: it.Comment,
			QuestionHTML: richtext.Sanitize(q), AnswerHTML: richtext.Sanitize(ans)})
	}
	return out, nil
}

func compose(words []string) string { return composer.Compose(checker.StoredAnswers(words)) }

// Choices are the options a session can have.
func (a *App) Choices() Choices {
	pairs := func(values []string) [][2]string {
		var out [][2]string
		for _, v := range values {
			out = append(out, [2]string{v, i18n.T(v)})
		}
		return out
	}
	return Choices{pairs(teaching.LessonTypes), pairs(teaching.Orders), pairs(teaching.WordChoices)}
}

// Start starts practising the open lesson.
func (a *App) Start(o Options) (State, error) {
	if a.data == nil {
		return State{}, ErrNoLesson
	}
	a.opts = o
	a.mathAnswer = map[int]bool{}
	for i, it := range a.data.List.Items {
		words := it.Answers
		if o.AskAnswers {
			words = it.Questions
		}
		for _, w := range words {
			a.mathAnswer[i] = a.mathAnswer[i] || richtext.HasMath(w)
		}
	}
	// answers are checked against the words' plain text (H<sub>2</sub>O
	// is typed H2O); the page shows the words with their markup
	a.session = teaching.New(plainList(a.data.List), teaching.Options{
		LessonType: o.LessonType, Order: o.Order, Words: o.Words, AskAnswers: o.AskAnswers,
	})
	a.session.Start()
	return a.State(), nil
}

// State is where the session is.
func (a *App) State() State {
	s := a.session
	if s == nil {
		return State{}
	}
	st := State{Active: !s.Done(), Done: s.Done()}
	st.Asked, st.Total = s.Progress()
	st.Right, st.Answered = s.Score()
	if item, index, ok := s.Current(); ok {
		st.Question = compose(item.Questions)
		st.Answer = s.CurrentAnswer()
		st.Shuffle = strings.TrimPrefix(teaching.ShuffleHint(st.Answer, nil), "Hint: ")
		st.QuestionHTML, st.AnswerHTML = a.shown(index)
	}
	return st
}

// Answer checks an answer to the current question and moves on.
func (a *App) Answer(text string) (Result, error) {
	if a.session == nil || a.session.Done() {
		return Result{}, errors.New("no question is being asked")
	}
	if strings.TrimSpace(text) == "" {
		return Result{}, errors.New("type an answer")
	}
	item, index, _ := a.session.Current()
	_, correctHTML := a.shown(index)
	if a.mathAnswer[index] {
		// a formula: its TeX, typed without regard to spaces, not OpenTeacher's
		// notation (whose commas would split f(x, y))
		correct := a.session.CurrentAnswer()
		right := false
		for _, w := range item.Answers {
			right = right || richtext.NormalizeAnswer(w) == richtext.NormalizeAnswer(text)
		}
		a.session.Record(right, text)
		a.finishIfDone()
		return Result{Right: right, Correct: correct, CorrectHTML: correctHTML}, nil
	}
	r := a.session.Answer(text)
	a.session.Next()
	a.finishIfDone()
	return Result{Right: r.Right, Correct: r.Correct, CorrectHTML: correctHTML}, nil
}

// Stop ends the session; what was answered is kept as a test.
func (a *App) Stop() State {
	if a.session != nil && !a.session.Done() {
		a.keep()
	}
	st := a.State()
	st.Active = false
	return st
}

func (a *App) finishIfDone() {
	if a.session.Done() {
		a.keep()
	}
}

// keep stores the session's answers in the lesson, as the desktop does.
func (a *App) keep() {
	if t := a.session.LessonTest(); len(t.Results) > 0 {
		a.data.List.Tests = append(a.data.List.Tests, t)
	}
}

// Report lists the answers of the session.
func (a *App) Report() []Row {
	if a.session == nil {
		return nil
	}
	// the original words (with their markup), the way round they were asked
	list := a.data.List
	list.Items = append([]lesson.WordItem(nil), list.Items...)
	if a.opts.AskAnswers {
		wordsreverser.Reverse(&list)
	}
	var rows []Row
	for _, r := range teaching.NewReport(list, a.session.Test()).Rows {
		rows = append(rows, Row{Question: richtext.Plain(r.Question), Answer: richtext.Plain(r.Answer), Given: r.Given,
			Right: r.Right, QuestionHTML: richtext.Sanitize(r.Question), AnswerHTML: richtext.Sanitize(r.Answer)})
	}
	return rows
}

// Save writes the lesson (with its sessions) in the format of name's
// extension, as the file's bytes.
func (a *App) Save(name string) ([]byte, error) {
	if a.data == nil {
		return nil, ErrNoLesson
	}
	dir, err := workDir()
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, filepath.Base(name))
	if err := lesson.NewFileSaver().SaveFile(a.data, path); err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// SetLanguage translates the choices and messages into lang.
func SetLanguage(dir, lang string) error { return i18n.Use(dir, lang) }

// ---- editing ----

// item makes an item of a question and an answer as typed ("hond",
// "dog, puppy"), with OpenTeacher's notation for several words.
func item(id int, question, answer string) (lesson.WordItem, error) {
	esc := strings.NewReplacer("\\", "\\\\", "=", "\\=", "\t", " ", "\n", " ")
	items, err := lesson.ParseWordList(esc.Replace(question)+" = "+esc.Replace(answer), false)
	if err != nil || len(items) != 1 {
		return lesson.WordItem{}, fmt.Errorf("%q = %q is not a word pair", question, answer)
	}
	it := items[0]
	it.ID = id
	return it, nil
}

// SetTitle renames the lesson.
func (a *App) SetTitle(title string) error {
	if a.data == nil {
		return ErrNoLesson
	}
	a.data.List.Title = strings.TrimSpace(title)
	return nil
}

// AddItem adds a word pair at the end.
func (a *App) AddItem(question, answer string) (Item, error) {
	if a.data == nil {
		return Item{}, ErrNoLesson
	}
	id := 0
	for _, it := range a.data.List.Items {
		id = max(id, it.ID+1)
	}
	it, err := item(id, question, answer)
	if err != nil {
		return Item{}, err
	}
	a.data.List.Items = append(a.data.List.Items, it)
	return Item{ID: id, Question: compose(it.Questions), Answer: compose(it.Answers)}, nil
}

// UpdateItem changes the question and answer of the item with id.
func (a *App) UpdateItem(id int, question, answer string) error {
	if a.data == nil {
		return ErrNoLesson
	}
	for i, old := range a.data.List.Items {
		if old.ID == id {
			it, err := item(id, question, answer)
			if err != nil {
				return err
			}
			it.Comment = old.Comment
			a.data.List.Items[i] = it
			return nil
		}
	}
	return fmt.Errorf("no word with id %d", id)
}

// RemoveItem removes the item with id.
func (a *App) RemoveItem(id int) error {
	if a.data == nil {
		return ErrNoLesson
	}
	items := a.data.List.Items
	for i, it := range items {
		if it.ID == id {
			a.data.List.Items = append(items[:i:i], items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no word with id %d", id)
}

// ---- the other practice modes ----

func (a *App) asking() error {
	if a.session == nil || a.session.Done() {
		return errors.New("no question is being asked")
	}
	return nil
}

// ViewAnswer is In mind's "View answer": the right answer (the thinking
// time ends now).
func (a *App) ViewAnswer() (string, error) {
	if err := a.asking(); err != nil {
		return "", err
	}
	return a.session.ViewAnswer(), nil
}

// Judge is In mind's "I was right" / "I was wrong"; it moves on.
func (a *App) Judge(right bool) error {
	if err := a.asking(); err != nil {
		return err
	}
	a.session.Judge(right)
	a.finishIfDone()
	return nil
}

// Skip asks the current question again later.
func (a *App) Skip() error {
	if err := a.asking(); err != nil {
		return err
	}
	a.session.Skip()
	return nil
}

// CorrectLast counts the last answer as right after all ("Correct
// anyway", for a typo).
func (a *App) CorrectLast() error {
	if a.session == nil {
		return errors.New("nothing was answered")
	}
	a.session.CorrectLast()
	return nil
}

// plainList is list with the words' markup taken out, for checking.
func plainList(list lesson.WordList) lesson.WordList {
	out := list
	out.Items = make([]lesson.WordItem, len(list.Items))
	plain := func(words []string) []string {
		out := make([]string, len(words))
		for i, w := range words {
			out[i] = richtext.Plain(w)
		}
		return out
	}
	for i, it := range list.Items {
		it.Questions, it.Answers = plain(it.Questions), plain(it.Answers)
		out.Items[i] = it
	}
	return out
}

// shown is the question and answer of the item at index as safe HTML,
// the way round the session asks them.
func (a *App) shown(index int) (question, answer string) {
	if index < 0 || index >= len(a.data.List.Items) {
		return "", ""
	}
	it := a.data.List.Items[index]
	q, ans := compose(it.Questions), compose(it.Answers)
	if a.opts.AskAnswers {
		q, ans = ans, q
	}
	return richtext.Sanitize(q), richtext.Sanitize(ans)
}
