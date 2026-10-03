package lessontypes_test

import (
	"testing"
	"time"

	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
	allonce "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/allOnce"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/interval"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes/smart"
)

// Ported from OpenTeacher's logic/lessonTypes/test, run for every lesson type.
var lessonTypes = map[string]func([]int) lessontypes.LessonType{
	"allOnce": func(i []int) lessontypes.LessonType { return allonce.New(i) },
	"smart":   func(i []int) lessontypes.LessonType { return smart.New(i) },
	"interval": func(i []int) lessontypes.LessonType {
		return interval.New(i, interval.DefaultSettings)
	},
}

const items = 2

func TestEmptyIndexes(t *testing.T) {
	for name, create := range lessonTypes {
		l := create(nil)
		l.OnNewItem(func(int) { t.Errorf("%s: newItem called for an empty lesson", name) })
		done := false
		l.OnLessonDone(func() { done = true })
		l.Start()
		if !done {
			t.Errorf("%s: lessonDone not called", name)
		}
		if l.Test() != nil {
			t.Errorf("%s: a test was recorded without results", name)
		}
	}
}

func TestAllItemsAskedAndLessonDone(t *testing.T) {
	for name, create := range lessonTypes {
		l := create([]int{0, 1})
		var asked []int
		l.OnNewItem(func(i int) {
			if i < 0 || i >= items {
				t.Errorf("%s: asked index %d", name, i)
			}
			asked = append(asked, i)
			l.SetResult(lessontypes.Result{ItemID: i, Right: true})
		})
		done := false
		l.OnLessonDone(func() {
			done = true
			if l.Test() == nil || !l.Test().Finished {
				t.Errorf("%s: test not finished at lessonDone", name)
			}
		})
		l.Start()
		if !done || len(asked) < items {
			t.Errorf("%s: done=%v asked=%v", name, done, asked)
		}
		if l.AskedItems() != items || l.TotalItems() != items {
			t.Errorf("%s: asked %d of %d", name, l.AskedItems(), l.TotalItems())
		}
	}
}

func TestSkipAndPause(t *testing.T) {
	for name, create := range lessonTypes {
		l := create([]int{0, 1})
		l.Start()
		l.Skip()
		l.AddPause(lessontypes.Pause{Start: time.Now(), End: time.Now()})
		l.SetResult(lessontypes.Result{Right: true})
		if got := len(l.Test().Pauses); got != 1 {
			t.Errorf("%s: %d pauses", name, got)
		}
	}
}

// run answers every question with answer(index) and returns the order asked.
func run(l lessontypes.LessonType, answer func(int) bool) []int {
	var asked []int
	l.OnNewItem(func(i int) {
		asked = append(asked, i)
		if len(asked) > 50 {
			panic("lesson does not end")
		}
		l.SetResult(lessontypes.Result{ItemID: i, Right: answer(i)})
	})
	l.Start()
	return asked
}

func TestAllOnceSkipMovesToEnd(t *testing.T) {
	l := allonce.New([]int{0, 1, 2})
	var asked []int
	skipped := false
	l.OnNewItem(func(i int) {
		asked = append(asked, i)
		if i == 0 && !skipped {
			skipped = true
			l.Skip()
			return
		}
		l.SetResult(lessontypes.Result{ItemID: i, Right: true})
	})
	l.Start()
	if want := []int{0, 1, 2, 0}; !equal(asked, want) {
		t.Errorf("asked %v, want %v", asked, want)
	}
}

func TestSmartRepeatsWrongItems(t *testing.T) {
	wrongOnce := map[int]bool{1: true}
	asked := run(smart.New([]int{0, 1, 2, 3, 4}), func(i int) bool {
		if wrongOnce[i] {
			delete(wrongOnce, i)
			return false
		}
		return true
	})
	// 1 is wrong: asked again two items later and once more at the end
	if want := []int{0, 1, 2, 3, 1, 4, 1}; !equal(asked, want) {
		t.Errorf("asked %v, want %v", asked, want)
	}
}

func TestSmartCorrectLastAnswerTakesBackRepetitions(t *testing.T) {
	l := smart.New([]int{0, 1, 2, 3, 4})
	var asked []int
	l.OnNewItem(func(i int) {
		asked = append(asked, i)
		if i == 1 && len(asked) == 2 {
			l.SetResult(lessontypes.Result{ItemID: 1, Right: false})
			return
		}
		if i == 2 && len(asked) == 3 {
			// "I was right" about item 1, while 2 is on screen
			l.CorrectLastAnswer(lessontypes.Result{ItemID: 1, Right: true})
		}
		l.SetResult(lessontypes.Result{ItemID: i, Right: true})
	})
	l.Start()
	if want := []int{0, 1, 2, 3, 4}; !equal(asked, want) {
		t.Errorf("asked %v, want %v", asked, want)
	}
	if r := l.Test().Results[1]; !r.Right || r.ItemID != 1 {
		t.Errorf("corrected result = %+v", r)
	}
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Reference orders from running OpenTeacher's Python interval lesson type
// with the same choice of positions ("low" always picks the nearest
// allowed position, "high" the farthest).
func TestIntervalMatchesOpenTeacher(t *testing.T) {
	low := func(n int) int { return 0 }
	high := func(n int) int { return n - 1 }
	wrongFirstTime := func() func(int) bool {
		seen := map[int]bool{}
		return func(i int) bool {
			if i == 1 && !seen[i] {
				seen[i] = true
				return false
			}
			return true
		}
	}
	allRight := func(int) bool { return true }
	cases := []struct {
		name    string
		indexes []int
		intn    func(int) int
		answer  func(int) bool
		want    []int
	}{
		{"all right, low", []int{0, 1, 2}, low, allRight, []int{0, 1, 0, 1, 2, 2}},
		{"all right, high", []int{0, 1, 2, 3, 4}, high, allRight, []int{0, 1, 2, 3, 0, 1, 2, 3, 4, 4}},
		{"one wrong, high", []int{0, 1, 2, 3}, high, wrongFirstTime(), []int{0, 1, 2, 3, 0, 1, 2, 3, 1, 1, 1}},
	}
	for _, c := range cases {
		l := interval.New(c.indexes, interval.DefaultSettings)
		l.Intn = c.intn
		if got := run(l, c.answer); !equal(got, c.want) {
			t.Errorf("%s: asked %v, want %v", c.name, got, c.want)
		}
	}
}

func TestIntervalCounts(t *testing.T) {
	l := interval.New([]int{0, 1, 2}, interval.DefaultSettings)
	l.Intn = func(int) int { return 0 }
	run(l, func(int) bool { return true })
	if l.AskedItems() != 3 || l.TotalItems() != 3 || !l.Test().Finished {
		t.Errorf("asked %d of %d, finished %v", l.AskedItems(), l.TotalItems(), l.Test().Finished)
	}
}

func TestModuleNamesAreIdentifiers(t *testing.T) {
	type named interface {
		Name() string
		DisplayName() string
		Type() string
	}
	for _, m := range []named{allonce.NewAllOnceModule(), smart.NewSmartModule(), interval.NewIntervalModule()} {
		if m.Type() != "lessonType" || m.Name() == m.DisplayName() {
			t.Errorf("%T: type %q, name %q, display name %q", m, m.Type(), m.Name(), m.DisplayName())
		}
	}
}
