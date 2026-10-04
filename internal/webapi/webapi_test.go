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
