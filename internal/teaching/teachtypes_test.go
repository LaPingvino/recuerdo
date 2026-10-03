package teaching

import (
	"testing"
	"time"
)

// seq returns the given numbers in a loop, as a fixed "random" source.
func seq(vals ...float64) func() float64 {
	i := 0
	return func() float64 { v := vals[i%len(vals)]; i++; return v }
}

// Reference hints from OpenTeacher's Python algorithm with the same numbers.
func TestShuffleHint(t *testing.T) {
	cases := []struct {
		answer string
		random []float64
		want   string
	}{
		{"three", []float64{0}, "Hint: hreet"},
		{"three", []float64{0.99}, "Hint: ethre"},
		{"three", []float64{0.5, 0.2}, "Hint: ereht"},
		{"être", []float64{0.3}, "Hint: treê"},
		{"ab", []float64{0}, "Hint: .."},
		{"aaa", []float64{0}, "Hint: ..."},
		{"", []float64{0}, "Hint: "},
	}
	for _, c := range cases {
		if got := ShuffleHint(c.answer, seq(c.random...)); got != c.want {
			t.Errorf("ShuffleHint(%q, %v) = %q, want %q", c.answer, c.random, got, c.want)
		}
	}
	if got := ShuffleHint("random", nil); len([]rune(got)) != len("Hint: random") {
		t.Errorf("with math/rand: %q", got)
	}
}

func TestCurrentAnswer(t *testing.T) {
	s := New(list, Options{})
	if s.CurrentAnswer() != "" {
		t.Error("answer before start")
	}
	s.Start()
	if got := s.CurrentAnswer(); got != "two" {
		t.Errorf("CurrentAnswer = %q", got)
	}
}

func TestInMind(t *testing.T) {
	clock := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	now := func() time.Time { clock = clock.Add(2 * time.Second); return clock }
	s := New(list, Options{LessonType: Smart, Now: now})
	s.Start()
	if got := s.ViewAnswer(); got != "two" {
		t.Errorf("ViewAnswer = %q", got)
	}
	s.Judge(true)
	s.ViewAnswer()
	s.Judge(false) // een: wrong, smart asks it again
	for !s.Done() {
		s.ViewAnswer()
		s.Judge(true)
	}
	res := s.Test().Results
	if len(res) != 4 || res[1].Right || res[1].GivenAnswer != "" {
		t.Fatalf("results %+v", res)
	}
	// thinking time ends when the answer is viewed, not when judged
	if d := res[0].End.Sub(res[0].Start); d != 2*time.Second {
		t.Errorf("thinking time %v", d)
	}
	if right, answered := s.Score(); right != 3 || answered != 4 {
		t.Errorf("score %d/%d", right, answered)
	}
}
