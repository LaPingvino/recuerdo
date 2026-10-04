package webapi

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPractiseAndSave(t *testing.T) {
	var a App
	if _, err := a.Start(Options{}); err != ErrNoLesson {
		t.Errorf("start without a lesson: %v", err)
	}
	l, err := a.OpenText("Dieren", "hond = dog\nkat = cat, kitty\n")
	if err != nil || len(l.Items) != 2 || l.Items[1].Answer != "cat, kitty" {
		t.Fatalf("%+v %v", l, err)
	}
	st, _ := a.Start(Options{})
	if !st.Active || st.Question != "hond" || st.Total != 2 {
		t.Fatalf("start: %+v", st)
	}
	if r, _ := a.Answer("dog"); !r.Right {
		t.Error("dog should be right")
	}
	if r, _ := a.Answer("mouse"); r.Right || r.Correct != "cat, kitty" {
		t.Errorf("mouse: %+v", r)
	}
	st = a.State()
	if !st.Done || st.Right != 1 || st.Answered != 2 {
		t.Errorf("end: %+v", st)
	}
	if rows := a.Report(); len(rows) != 2 || rows[1].Given != "mouse" || rows[1].Right {
		t.Errorf("report %+v", rows)
	}
	// the session is kept in the lesson, and saved with it
	b, err := a.Save("dieren.otwd")
	if err != nil || !bytes.HasPrefix(b, []byte("PK")) {
		t.Fatalf("save: %d bytes %v", len(b), err)
	}
	var again App
	l, err = again.Open("dieren.otwd", b)
	if err != nil || l.Sessions != 1 || len(l.Items) != 2 {
		t.Errorf("reopened: %+v %v", l, err)
	}
}

func TestOpenSampleFiles(t *testing.T) {
	dir := filepath.Join("..", "..", "testdata", "legacy_files")
	for _, name := range []string{"application_x-openteachingwords.openteacher3x.otwd", "text_csv.openteacher3x.csv"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		var a App
		if l, err := a.Open(name, b); err != nil || len(l.Items) == 0 {
			t.Errorf("%s: %+v %v", name, l, err)
		}
	}
	var a App
	if _, err := a.Open("empty.csv", []byte("")); err == nil {
		t.Error("an empty file opened")
	}
}

func TestChoices(t *testing.T) {
	c := (&App{}).Choices()
	if len(c.LessonTypes) != 3 || c.LessonTypes[0][0] != "All once" || len(c.Orders) == 0 {
		t.Errorf("%+v", c)
	}
}

func TestEditing(t *testing.T) {
	var a App
	a.OpenText("Dieren", "hond = dog\n")
	it, err := a.AddItem("kat", "cat, kitty")
	if err != nil || it.ID != 1 || it.Answer != "cat, kitty" {
		t.Fatalf("add: %+v %v", it, err)
	}
	if _, err := a.AddItem("a = b", "c"); err != nil {
		t.Errorf("an equals sign in a word: %v", err)
	}
	if err := a.UpdateItem(0, "hond", "dog, puppy"); err != nil {
		t.Fatal(err)
	}
	if err := a.RemoveItem(1); err != nil || a.RemoveItem(1) == nil {
		t.Errorf("remove: %v", err)
	}
	a.SetTitle(" Huisdieren ")
	l, _ := a.Lesson()
	if l.Title != "Huisdieren" || len(l.Items) != 2 || l.Items[0].Answer != "dog, puppy" || l.Items[1].Question != "a = b" {
		t.Errorf("lesson %+v", l)
	}
	if a.UpdateItem(9, "x", "y") == nil {
		t.Error("updated a word that is not there")
	}
}

func TestModes(t *testing.T) {
	var a App
	a.OpenText("Dieren", "hond = dog\nkat = cat\nmuis = mouse\n")
	st, _ := a.Start(Options{})
	if st.Answer != "dog" || len(st.Shuffle) != 3 {
		t.Errorf("state: %+v", st)
	}
	// In mind: view the answer, judge
	if ans, _ := a.ViewAnswer(); ans != "dog" {
		t.Errorf("view answer %q", ans)
	}
	a.Judge(true)
	// skip kat: it comes back later
	a.Skip()
	if q := a.State().Question; q != "muis" {
		t.Errorf("after skipping: %q", q)
	}
	a.Answer("mice")
	a.CorrectLast() // a typo, counted as right after all
	st = a.State()
	if st.Right != 2 || st.Question != "kat" {
		t.Errorf("after correcting: %+v", st)
	}
	a.Answer("cat")
	if st = a.State(); !st.Done || st.Right != 3 {
		t.Errorf("end: %+v", st)
	}
	if a.Judge(true) == nil || a.Skip() == nil {
		t.Error("judging after the end")
	}
}

func TestRichWords(t *testing.T) {
	var a App
	a.OpenText("Chemie", "water = H<sub>2</sub>O\n<ruby>漢<rt>かん</rt></ruby> = kan\n")
	l, _ := a.Lesson()
	if l.Items[0].AnswerHTML != "H<sub>2</sub>O" || l.Items[1].QuestionHTML != "<ruby>漢<rt>かん</rt></ruby>" {
		t.Errorf("html: %+v", l.Items)
	}
	st, _ := a.Start(Options{})
	if st.QuestionHTML != "water" || st.AnswerHTML != "H<sub>2</sub>O" {
		t.Errorf("state: %+v", st)
	}
	if r, _ := a.Answer("H2O"); !r.Right || r.CorrectHTML != "H<sub>2</sub>O" {
		t.Errorf("H2O: %+v", r)
	}
	if st = a.State(); st.Question != "漢" || st.QuestionHTML != "<ruby>漢<rt>かん</rt></ruby>" {
		t.Errorf("furigana question: %+v", st)
	}
	// asked the other way round
	a.Start(Options{AskAnswers: true})
	if st = a.State(); st.QuestionHTML != "H<sub>2</sub>O" {
		t.Errorf("reversed: %+v", st)
	}
	if r, _ := a.Answer("water"); !r.Right {
		t.Error("water should be right")
	}
}

func TestFormulas(t *testing.T) {
	var a App
	a.OpenText("Wiskunde", "the area of a circle = $\\pi r^2$\nx squared plus one = $x^2 + 1$\n")
	l, _ := a.Lesson()
	if l.Items[0].AnswerHTML != "$\\pi r^2$" {
		t.Errorf("formulas stay as TeX for KaTeX: %q", l.Items[0].AnswerHTML)
	}
	a.Start(Options{})
	if r, _ := a.Answer("\\pi r^2"); !r.Right {
		t.Errorf("\\pi r^2: %+v", r)
	}
	if r, _ := a.Answer("x^2+1"); !r.Right {
		t.Errorf("x^2+1 (without spaces): %+v", r)
	}
	if rows := a.Report(); len(rows) != 2 || rows[0].AnswerHTML != "$\\pi r^2$" || rows[0].Answer != "\\pir^2" {
		t.Errorf("report: %+v", rows)
	}
	if st := a.State(); !st.Done || st.Right != 2 {
		t.Errorf("both right: %+v", st)
	}
	a.OpenText("f", "a function of two variables = $f(x, y)$\n")
	a.Start(Options{})
	if r, _ := a.Answer("f(x,y)"); !r.Right || r.CorrectHTML != "$f(x, y)$" {
		t.Errorf("f(x,y): %+v", r)
	}
	a.OpenText("f", "a function of two variables = $f(x, y)$\n")
	a.Start(Options{})
	if r, _ := a.Answer("f(y,x)"); r.Right {
		t.Error("f(y,x) is not f(x, y)")
	}
}

func TestFormulaBuilder(t *testing.T) {
	var a App
	a.OpenText("Wiskunde", "area of a circle = $\\pi r^2$\nhond = dog\n")
	if _, err := a.Start(Options{}); err != nil {
		t.Fatal(err)
	}
	if st := a.State(); !st.AnswerIsMath {
		t.Errorf("a formula answer: AnswerIsMath false (%+v)", st)
	}
	a.Answer("\\pi r^2")
	if st := a.State(); st.AnswerIsMath {
		t.Errorf("a word answer: AnswerIsMath true (%+v)", st)
	}
	if len(a.Palette()) == 0 {
		t.Error("no palette")
	}
	if ins, err := a.Expand("frac", "a"); err != nil || ins.Text != `\frac{a}{}` || ins.Cursor != 9 {
		t.Errorf("Expand = %+v, %v", ins, err)
	}
	if _, err := a.Expand("nope", ""); err == nil {
		t.Error("an unknown button: no error")
	}
}
