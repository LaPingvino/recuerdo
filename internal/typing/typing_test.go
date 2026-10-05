package typing

import (
	"errors"
	"math/rand"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestLetterExercises(t *testing.T) {
	q := LayoutByID("qwerty")
	ex := LetterExercises(q)
	if len(ex) != 37 || Levels(q) != 57 {
		t.Fatalf("%d letter exercises, %d levels", len(ex), Levels(q))
	}
	want := [][]string{{"f", "j"}, {"d", "k"}, {"s", "l"}, {"a", ";"}, {"a", "s", "d", "f"}, {"j", "k", "l", ";"},
		{"g", "h"}, {"f", "g", "h", "j"}, {"a", "s", "d", "f", "g", "h", "j", "k", "l", ";"}}
	if !reflect.DeepEqual(ex[:9], want) {
		t.Errorf("home row %q", ex[:9])
	}
	// the bottom row starts after the ISO key; "everything" is right too
	if !reflect.DeepEqual(ex[18], []string{"v", "m"}) { // under f and j
		t.Errorf("bottom row %q", ex[18])
	}
	all := ex[36]
	if slices.Contains(all, "\\") || !slices.Contains(all, "/") || !slices.Contains(all, "z") || len(all) != 40 {
		t.Errorf("everything %q", all)
	}
	// AZERTY has its w (OpenTeacher had x twice)
	for _, id := range []string{"azerty-be", "azerty-fr"} {
		if _, ok := LayoutByID(id).KeyFor("w"); !ok {
			t.Errorf("%s has no w", id)
		}
	}
	// every layout: 5 rows with the geometry's keys, a finger for each key
	for _, l := range Layouts {
		for r, row := range l.Rows {
			if len(row) != len(geometry[r]) {
				t.Errorf("%s row %d: %d keys", l.ID, r, len(row))
			}
		}
		if k, _ := l.KeyFor(" "); k.Finger != 5 {
			t.Errorf("%s space", l.ID)
		}
	}
	if k, _ := q.KeyFor("f"); k.Finger != 4 {
		t.Errorf("f is typed by finger %d", k.Finger)
	}
}

func TestExercise(t *testing.T) {
	q := LayoutByID("qwerty")
	rng := rand.New(rand.NewSource(1))
	e := Exercise(q, 0, nil, rng)
	if len(e) > 59 || strings.Trim(e, "fj ") != "" || !strings.Contains(e, " ") {
		t.Errorf("level 0: %q", e)
	}
	for _, g := range strings.Fields(e)[:len(strings.Fields(e))-1] {
		if len(g) != 5 {
			t.Errorf("group %q", g)
		}
	}
	// word levels: only words the layout can type
	words := Exercise(q, 40, []string{"rivière", "jardin", "école"}, rng)
	if strings.ContainsAny(words, "èé") || len(strings.Fields(words)) != 8 {
		t.Errorf("words %q", words)
	}
	if fr := Exercise(LayoutByID("azerty-fr"), 40, []string{"école"}, rng); !strings.Contains(fr, "école") {
		t.Errorf("french %q", fr)
	}
	if len(Words("nl")) < 100 || len(Words("xx")) != len(Words("en")) || Words("pt_BR")[0] != Words("en")[0] {
		t.Error("word lists")
	}
}

func TestSpeed(t *testing.T) {
	if WordsPerMinute(strings.Repeat("x", 100), 60) != 20 || WordsPerMinute("abc", 0) != 0 {
		t.Error("words per minute")
	}
	q := LayoutByID("qwerty")
	if TargetSpeed(q, 0) != 20 || TargetSpeed(q, 36) != 20 || TargetSpeed(q, 37) != 20 || TargetSpeed(q, 56) != 80 {
		t.Errorf("targets %d %d %d", TargetSpeed(q, 37), TargetSpeed(q, 46), TargetSpeed(q, 56))
	}
}

// OpenTeacher's scripted session (typingTutorModelTest), with the values.
func TestSession(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	ps, _ := Load(filepath.Join(t.TempDir(), "typing.json"))
	if _, err := ps.Add(" ", "qwerty", "en", rng); !errors.Is(err, ErrNameEmpty) {
		t.Error("empty name")
	}
	p, err := ps.Add("Anna", "qwerty", "en", rng)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ps.Add("anna", "dvorak", "en", rng); !errors.Is(err, ErrNameTaken) {
		t.Error("name taken")
	}
	if _, ok := p.Last(); ok || p.Status != Start || !strings.Contains(p.Instruction(rng), "'f'") {
		t.Error("start")
	}
	fast := func() float64 { return float64(len(p.Current)) / 5 / 25 * 60 } // 25 wpm
	p.Finish(fast(), 0, rng)
	if p.Level != 1 || p.Status != Next || !strings.Contains(p.Instruction(rng), "first exercise") {
		t.Errorf("after a good first exercise: level %d %s", p.Level, p.Status)
	}
	p.Finish(fast(), 3, rng)
	if p.Level != 1 || p.Status != Mistakes || !strings.Contains(p.Instruction(rng), "3") {
		t.Errorf("mistakes: %d %s %q", p.Level, p.Status, p.Instruction(rng))
	}
	p.Finish(fast()*4, 0, rng) // about 6 wpm
	if p.Level != 1 || p.Status != Slow {
		t.Errorf("slow: %d %s", p.Level, p.Status)
	}
	for p.Level < 9 {
		p.Finish(fast(), 0, rng)
	}
	if !strings.Contains(p.Instruction(rng), "home row") {
		t.Errorf("leaving the home row: %q", p.Instruction(rng))
	}
	for p.Status != Done {
		target := TargetSpeed(p.KeyboardLayout(), p.Level)
		p.Finish(float64(len(p.Current))/5/float64(target+1)*60, 0, rng)
	}
	if p.Level != 56 || !strings.Contains(p.Instruction(rng), "finished this typing course") {
		t.Errorf("done at %d", p.Level)
	}
	if err := ps.Save(); err != nil {
		t.Fatal(err)
	}
	again, err := Load(ps.path)
	if err != nil || again.Get("ANNA") == nil || again.Get("anna").Level != 56 || len(again.Get("anna").Results) != len(p.Results) {
		t.Fatalf("reloaded: %v", err)
	}
	again.Remove("Anna")
	if again.Get("Anna") != nil {
		t.Error("removed")
	}
}
